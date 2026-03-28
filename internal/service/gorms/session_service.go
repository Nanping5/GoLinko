package gormss

import (
	"GoLinko/internal/dao"
	"GoLinko/internal/dto/request"
	"GoLinko/internal/dto/response"
	"GoLinko/internal/model"
	"strings"

	myredis "GoLinko/internal/service/redis"
	constants "GoLinko/pkg/const"
	"GoLinko/pkg/enum/contact/contact_status_enum"
	"GoLinko/pkg/enum/contact/contact_type_enum"
	"GoLinko/pkg/enum/group_info/group_status_enum"
	"GoLinko/pkg/enum/user_info/user_status_enum"
	"GoLinko/pkg/utils"
	"GoLinko/pkg/zlog"
	"context"
	"encoding/json"
	"errors"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type sessionService struct {
}

var SessionService = &sessionService{}

func (s *sessionService) OpenSession(ctx context.Context, send_id, receive_id string) (string, string, int) {
	msg, Allowed, _ := s.CheckOpenSessionAllowed(ctx, send_id, receive_id)
	if !Allowed.Allowed {
		return msg, "", -2
	}
	db := dao.NewDbClient(ctx)

	// 获取标准化 PairKey (单聊采用 sort(id1, id2))
	pairKey := s.getPairKey(send_id, receive_id)

	session := model.Session{}
	err := db.Where("pair_key = ?", pairKey).First(&session).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			zlog.GetLogger().Info("会话不存在，创建新会话", zap.String("pair_key", pairKey))
			req := request.CreateNewSessionRequest{
				SendId:    send_id,
				ReceiveId: receive_id,
			}
			return s.CreateNewSession(ctx, req)
		} else {
			zlog.GetLogger().Error("查询会话失败", zap.String("pair_key", pairKey), zap.Error(err))
			return constants.SYSTEM_ERROR, "", -1
		}
	}
	// 重新打开会话时，取消当前用户的隐藏状态
	if err := db.Unscoped().Where("user_id = ? AND session_id = ?", send_id, session.Uuid).Delete(&model.UserSessionHide{}).Error; err != nil {
		zlog.GetLogger().Error("清理会话隐藏状态失败", zap.String("user_id", send_id), zap.String("session_id", session.Uuid), zap.Error(err))
	}
	return "会话创建成功", session.Uuid, 0
}

// getPairKey 生成单聊唯一键 (对 uid1, uid2 排序后组合)
func (s *sessionService) getPairKey(id1, id2 string) string {
	if id1 < id2 {
		return id1 + "_" + id2
	}
	return id2 + "_" + id1
}

// CreateNewSession 创建新会话
func (s *sessionService) CreateNewSession(ctx context.Context, req request.CreateNewSessionRequest) (string, string, int) {
	db := dao.NewDbClient(ctx)

	session := model.Session{
		Uuid:      utils.GenerateSessionID(),
		SendId:    req.SendId,
		ReceiveId: req.ReceiveId,
	}

	// 极简类型判断：根据 ID 首字母直接决定逻辑分支 (U: 用户, G: 群组)
	if len(req.ReceiveId) == 0 {
		return "接收者ID不能为空", "", -2
	}

	if req.ReceiveId[0] == 'G' {
		var groupInfo model.GroupInfo
		if err := db.Where("uuid = ?", req.ReceiveId).First(&groupInfo).Error; err != nil {
			return "群聊不存在", "", -2
		}
		if groupInfo.Status == group_status_enum.DISABLE {
			zlog.GetLogger().Info("群聊已被禁用", zap.String("group_id", req.ReceiveId))
			return "群聊已被禁用", "", -2
		}
		session.ReceiveName = groupInfo.Name
		session.Avatar = groupInfo.Avatar
		session.PairKey = req.ReceiveId
	} else if req.ReceiveId[0] == 'U' {
		var userInfo model.UserInfo
		if err := db.Where("uuid = ?", req.ReceiveId).First(&userInfo).Error; err != nil {
			return "用户不存在", "", -2
		}
		if userInfo.Status == user_status_enum.DISABLE {
			zlog.GetLogger().Info("用户已被禁用", zap.String("user_id", req.ReceiveId))
			return "用户已被禁用", "", -2
		}
		session.ReceiveName = userInfo.Nickname
		session.Avatar = userInfo.Avatar
		session.PairKey = s.getPairKey(req.SendId, req.ReceiveId)
	} else {
		return "非法的ID格式", "", -2
	}

	if err := db.Create(&session).Error; err != nil {
		// 物理唯一约束保证不再产生重复记录
		var existing model.Session
		if e2 := db.Where("pair_key = ?", session.PairKey).First(&existing).Error; e2 == nil {
			return "会话创建成功", existing.Uuid, 0
		}
		zlog.GetLogger().Error("创建会话失败", zap.Error(err))
		return constants.SYSTEM_ERROR, "", -1
	}
	// 创建会话成功后，删除相关缓存
	if err := myredis.DelKeyWithPatternIfExist("group_session_list_" + req.SendId); err != nil {
		zlog.GetLogger().Error("删除缓存失败", zap.String("key_pattern", "group_session_list_"+req.SendId), zap.Error(err))
	}
	if err := myredis.DelKeyWithPatternIfExist("session_list_" + req.SendId); err != nil {
		zlog.GetLogger().Error("删除缓存失败", zap.String("key_pattern", "session_list_"+req.SendId), zap.Error(err))
	}
	// 群聊场景下，清理缓存
	if err := myredis.DelKeyWithPatternIfExist("group_session_list_" + req.ReceiveId); err != nil {
		zlog.GetLogger().Error("删除接收者缓存失败", zap.String("key_pattern", "group_session_list_"+req.ReceiveId), zap.Error(err))
	}

	return "会话创建成功", session.Uuid, 0
}

// GetSessionList 获取用户会话列表（仅单聊）
func (s *sessionService) GetSessionList(ctx context.Context, send_id string) (string, []response.GetSessionListResponse, int) {

	db := dao.NewDbClient(ctx)
	respString, err := myredis.GetKeyNilIsError("session_list_" + send_id)
	if err != nil || respString == "" {

		type sessionRow struct {
			Uuid          string
			ContactId     string
			ContactName   string
			ContactAvatar string
		}
		var rows []sessionRow
		// 通过 pair_key 解析出另一方，并关联 user_info 获取即时头像/昵称
		if err := db.Raw(`
			SELECT s.uuid,
				res.contact_id,
				ui.nickname AS contact_name,
				ui.avatar AS contact_avatar
			FROM session s
			LEFT JOIN user_session_hide ush
				ON ush.session_id = s.uuid
				AND ush.user_id = ?
				AND ush.deleted_at IS NULL
			INNER JOIN (
				/* 核心逻辑：从 pair_key 中解析出非自己的那个 ID 作为 contact_id */
				SELECT uuid, 
					CASE 
						WHEN SUBSTRING_INDEX(pair_key, '_', 1) = ? THEN SUBSTRING_INDEX(pair_key, '_', -1)
						ELSE SUBSTRING_INDEX(pair_key, '_', 1)
					END AS contact_id
				FROM session 
				WHERE pair_key LIKE CONCAT('%', ?, '%') AND pair_key LIKE '%_%'
			) res ON s.uuid = res.uuid
			INNER JOIN user_contact uc 
				ON uc.user_id = ? 
				AND uc.contact_id = res.contact_id 
				AND uc.contact_type = ?
			LEFT JOIN user_info ui ON ui.uuid = res.contact_id AND ui.deleted_at IS NULL
			WHERE s.deleted_at IS NULL AND ush.id IS NULL
			ORDER BY s.updated_at DESC
		`, send_id, send_id, send_id, send_id, contact_type_enum.USER).Scan(&rows).Error; err != nil {
			zlog.GetLogger().Error("查询会话列表失败", zap.String("send_id", send_id), zap.Error(err))
			return constants.SYSTEM_ERROR, nil, -1
		}

		var resp []response.GetSessionListResponse
		for _, row := range rows {
			resp = append(resp, response.GetSessionListResponse{
				SessionId:   row.Uuid,
				ReceiveId:   row.ContactId,
				ReceiveName: row.ContactName,
				Avatar:      row.ContactAvatar,
			})
		}
		respString, err := json.Marshal(resp)
		if err != nil {
			zlog.GetLogger().Error("序列化会话列表失败", zap.String("send_id", send_id), zap.Error(err))
		}
		if err := myredis.SetKeyEx("session_list_"+send_id, string(respString), 60*constants.REDIS_TIMEOUT); err != nil {
			zlog.GetLogger().Error(err.Error())
		}
		return "查询会话列表成功", resp, 0
	}
	resp := []response.GetSessionListResponse{}
	if err := json.Unmarshal([]byte(respString), &resp); err != nil {
		zlog.GetLogger().Error("反序列化会话列表失败", zap.String("send_id", send_id), zap.Error(err))
		return constants.SYSTEM_ERROR, nil, -1
	}
	return "查询会话列表成功", resp, 0
}

// GetGroupSessionList 获取群聊会话列表
func (s *sessionService) GetGroupSessionList(ctx context.Context, send_id string) (string, []response.GetGroupSessionListResponse, int) {

	db := dao.NewDbClient(ctx)
	respString, err := myredis.GetKeyNilIsError("group_session_list_" + send_id)

	if err != nil || respString == "" {

		var sessions []model.Session
		// 通过 JOIN user_contact 表过滤出该用户加入的群聊会话
		if err := db.Table("session as s").
			Joins("INNER JOIN user_contact uc ON s.pair_key = uc.contact_id").
			Joins("LEFT JOIN user_session_hide ush ON ush.session_id = s.uuid AND ush.user_id = ? AND ush.deleted_at IS NULL", send_id).
			Where("uc.user_id = ? AND uc.contact_type = ? AND uc.status NOT IN (?, ?) AND ush.id IS NULL", send_id, contact_type_enum.GROUP, contact_status_enum.QUIT_GROUP, contact_status_enum.KICK_OUT_GROUP).
			Order("s.updated_at DESC").
			Find(&sessions).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return "群聊会话列表为空", []response.GetGroupSessionListResponse{}, 0
			}
			zlog.GetLogger().Error("查询群聊会话列表失败", zap.String("send_id", send_id), zap.Error(err))
			return constants.SYSTEM_ERROR, nil, -1
		}
		var resp []response.GetGroupSessionListResponse
		for _, session := range sessions {
			resp = append(resp, response.GetGroupSessionListResponse{
				SessionId: session.Uuid,
				GroupId:   session.ReceiveId,
				GroupName: session.ReceiveName,
				Avatar:    session.Avatar,
			})
		}
		respString, err := json.Marshal(resp)
		if err != nil {
			zlog.GetLogger().Error("序列化群聊会话列表失败", zap.String("send_id", send_id), zap.Error(err))
		}
		if err := myredis.SetKeyEx("group_session_list_"+send_id, string(respString), 60*constants.REDIS_TIMEOUT); err != nil {
			zlog.GetLogger().Error(err.Error())
		}
		return "查询群聊会话列表成功", resp, 0
	}
	resp := []response.GetGroupSessionListResponse{}
	if err := json.Unmarshal([]byte(respString), &resp); err != nil {
		zlog.GetLogger().Error("反序列化群聊会话列表失败", zap.String("send_id", send_id), zap.Error(err))
		return constants.SYSTEM_ERROR, nil, -1
	}
	return "查询群聊会话列表成功", resp, 0
}

// DeleteSession 删除会话
func (s *sessionService) DeleteSession(ctx context.Context, send_id, session_id string) (string, int) {

	db := dao.NewDbClient(ctx)
	var session model.Session
	if err := db.Where("uuid = ?", session_id).First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "会话不存在", -2
		}
		zlog.GetLogger().Error("查询会话失败", zap.String("session_id", session_id), zap.String("send_id", send_id), zap.Error(err))
		return constants.SYSTEM_ERROR, -1
	}

	if !s.canAccessSession(ctx, send_id, session) {
		return "无权操作该会话", -2
	}

	hide := model.UserSessionHide{}
	err := db.Unscoped().Where("user_id = ? AND session_id = ?", send_id, session_id).First(&hide).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		hide = model.UserSessionHide{UserId: send_id, SessionId: session_id}
		if err := db.Create(&hide).Error; err != nil {
			zlog.GetLogger().Error("隐藏会话失败", zap.String("session_id", session_id), zap.String("send_id", send_id), zap.Error(err))
			return constants.SYSTEM_ERROR, -1
		}
	} else if err != nil {
		zlog.GetLogger().Error("查询会话隐藏状态失败", zap.String("session_id", session_id), zap.String("send_id", send_id), zap.Error(err))
		return constants.SYSTEM_ERROR, -1
	} else if hide.Model != nil && hide.DeletedAt.Valid {
		if err := db.Unscoped().Model(&hide).Update("deleted_at", nil).Error; err != nil {
			zlog.GetLogger().Error("恢复会话隐藏状态失败", zap.String("session_id", session_id), zap.String("send_id", send_id), zap.Error(err))
			return constants.SYSTEM_ERROR, -1
		}
	}

	// 删除会话成功后，删除相关缓存
	if err := myredis.DelKeyWithPatternIfExist("group_session_list_" + send_id); err != nil {
		zlog.GetLogger().Error("删除缓存失败", zap.String("key_pattern", "group_session_list_"+send_id), zap.Error(err))
	}
	if err := myredis.DelKeyWithPatternIfExist("session_list_" + send_id); err != nil {
		zlog.GetLogger().Error("删除缓存失败", zap.String("key_pattern", "session_list_"+send_id), zap.Error(err))
	}
	return "删除会话成功", 0
}

func (s *sessionService) canAccessSession(ctx context.Context, userID string, session model.Session) bool {
	if session.PairKey != "" && session.PairKey[0] == 'G' {
		contact := model.UserContact{}
		db := dao.NewDbClient(ctx)
		if err := db.Where("user_id = ? AND contact_id = ? AND contact_type = ? AND status NOT IN (?, ?)",
			userID, session.PairKey, contact_type_enum.GROUP, contact_status_enum.QUIT_GROUP, contact_status_enum.KICK_OUT_GROUP).
			First(&contact).Error; err != nil {
			return false
		}
		return true
	}

	if len(session.PairKey) == 0 {
		return false
	}

	parts := strings.Split(session.PairKey, "_")
	if len(parts) != 2 {
		return false
	}
	return parts[0] == userID || parts[1] == userID
}

func (s *sessionService) CheckOpenSessionAllowed(ctx context.Context, send_id, receive_id string) (string, *response.CheckOpenSessionAllowedResponse, int) {
	contact := model.UserContact{}
	db := dao.NewDbClient(ctx)
	if err := db.Where("user_id=? and contact_id=?", send_id, receive_id).First(&contact).Error; err != nil {
		zlog.GetLogger().Error("查询用户关系失败", zap.String("send_id", send_id), zap.String("receive_id", receive_id), zap.Error(err))
		return constants.SYSTEM_ERROR, nil, -1
	}
	if contact.Status == contact_status_enum.BE_BLACK {
		return "你已被对方拉黑，无法发起会话", &response.CheckOpenSessionAllowedResponse{Allowed: false}, 0
	}
	if contact.Status == contact_status_enum.BLACK {
		return "你已将对方拉黑，无法发起会话", &response.CheckOpenSessionAllowedResponse{Allowed: false}, 0
	}
	// 检查好友删除状态
	if contact.Status == contact_status_enum.BE_DELETE || contact.Status == contact_status_enum.DELETE {
		return "好友关系已删除，无法发起会话", &response.CheckOpenSessionAllowedResponse{Allowed: false}, 0
	}
	// 检查群聊退出/踢出状态
	if contact.Status == contact_status_enum.QUIT_GROUP || contact.Status == contact_status_enum.KICK_OUT_GROUP {
		return "已不在群聊中，无法发起会话", &response.CheckOpenSessionAllowedResponse{Allowed: false}, 0
	}

	// 根据首字母快速判断类型并检查状态
	if len(receive_id) > 0 && receive_id[0] == 'G' {
		group := model.GroupInfo{}
		if err := db.Where("uuid=?", receive_id).First(&group).Error; err != nil {
			zlog.GetLogger().Error("群组不存在", zap.String("group_id", receive_id))
			return "群组不存在", &response.CheckOpenSessionAllowedResponse{Allowed: false}, -1
		}
		if group.Status == group_status_enum.DISABLE {
			return "群聊已被禁用", &response.CheckOpenSessionAllowedResponse{Allowed: false}, -2
		}
	} else {
		user := model.UserInfo{}
		if err := db.Where("uuid=?", receive_id).First(&user).Error; err != nil {
			zlog.GetLogger().Error("用户不存在", zap.String("user_id", receive_id))
			return "用户不存在", &response.CheckOpenSessionAllowedResponse{Allowed: false}, -1
		}
		if user.Status == user_status_enum.DISABLE {
			return "对方已被禁用，无法发起会话", &response.CheckOpenSessionAllowedResponse{Allowed: false}, -2
		}
	}

	return "可以发起会话", &response.CheckOpenSessionAllowedResponse{Allowed: true}, 0
}
