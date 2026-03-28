package gormss

import (
	"GoLinko/internal/dao"
	"GoLinko/internal/dto/request"
	"GoLinko/internal/dto/response"
	"GoLinko/internal/model"
	myredis "GoLinko/internal/service/redis"
	constants "GoLinko/pkg/const"
	"GoLinko/pkg/enum/contact/contact_status_enum"
	"GoLinko/pkg/enum/contact/contact_type_enum"
	"GoLinko/pkg/enum/group_info/group_status_enum"
	"GoLinko/pkg/enum/message/message_status_enum"
	"GoLinko/pkg/enum/message/message_type_enum"
	"GoLinko/pkg/utils"
	"GoLinko/pkg/zlog"
	"context"
	"encoding/json"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type groupInfoService struct {
}

var GroupInfoService = new(groupInfoService)

// CreateGroup 创建群组
func (g *groupInfoService) CreateGroup(ctx context.Context, req *request.CreateGroupRequest, ownerId string) (string, string, int) {
	db := dao.NewDbClient(ctx)
	if db == nil {
		return constants.SYSTEM_ERROR, "", -1
	}
	group := model.GroupInfo{
		Uuid:      utils.GenerateGroupID(),
		Name:      req.GroupName,
		Avatar:    req.Avatar,
		Notice:    req.Notice,
		AddMode:   *req.AddMode,
		OwnerId:   ownerId,
		MemberCnt: 1,
	}
	var menbers []string
	menbers = append(menbers, ownerId)
	var err error
	group.Members, err = json.Marshal(menbers)
	if err != nil {
		msg := "序列化群成员列表失败"
		zlog.GetLogger().Error(msg, zap.Error(err), zap.String("members", string(group.Members)))
		return constants.SYSTEM_ERROR, "", -1
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&group).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.UserContact{
			UserId:      group.OwnerId,
			ContactId:   group.Uuid,
			ContactType: contact_type_enum.GROUP,
			Status:      contact_status_enum.NORMAL,
		}).Error; err != nil {
			return err
		}

		// 创建群组后发送欢迎消息
		welcomeMsg := model.Message{
			Uuid:      utils.GenerateMessageID(),
			SessionId: group.Uuid,
			Type:      message_type_enum.Text,
			Content:   "群组创建成功，欢迎大家加入！",
			SendId:    "SYSTEM",
			SendName:  "系统",
			ReceiveId: group.Uuid,
			Status:    message_status_enum.Sent,
		}
		return tx.Create(&welcomeMsg).Error
	})
	if err != nil {
		zlog.GetLogger().Error("创建群组失败", zap.Error(err))
		return "创建群组失败", "", -1
	}
	return "创建群组成功", group.Uuid, 0
}

// LoadMyGroups 加载我的群组列表
func (g *groupInfoService) LoadMyGroups(ctx context.Context, userId string) (string, []response.LoadMyGroupsResponse, int) {
	db := dao.NewDbClient(ctx)
	if db == nil {
		return constants.SYSTEM_ERROR, nil, -1
	}

	// 先查询用户的群组联系人
	var userContacts []model.UserContact
	if err := db.Where("user_id=? AND contact_type=?", userId, contact_type_enum.GROUP).Find(&userContacts).Error; err != nil {
		zlog.GetLogger().Error("查询用户群组联系人失败", zap.Error(err), zap.String("user_id", userId))
		return constants.SYSTEM_ERROR, nil, -1
	}

	// 提取群组ID列表
	var groupIds []string
	for _, contact := range userContacts {
		groupIds = append(groupIds, contact.ContactId)
	}

	// 如果没有群组，直接返回空列表
	if len(groupIds) == 0 {
		return "查询群组成功", []response.LoadMyGroupsResponse{}, 0
	}

	// 批量查询群组信息
	var groupList []model.GroupInfo
	if err := db.Where("uuid IN (?) AND status=?", groupIds, 0).Find(&groupList).Error; err != nil {
		zlog.GetLogger().Error("查询群组信息失败", zap.Error(err), zap.String("user_id", userId))
		return constants.SYSTEM_ERROR, nil, -1
	}

	groupListResp := []response.LoadMyGroupsResponse{}
	for _, group := range groupList {
		groupListResp = append(groupListResp, response.LoadMyGroupsResponse{
			GroupId:   group.Uuid,
			GroupName: group.Name,
			Avatar:    group.Avatar,
			Notice:    group.Notice,
			AddMode:   group.AddMode,
			OwnerId:   group.OwnerId,
			CreatedAt: group.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return "查询群组成功", groupListResp, 0
}

// CheckGroupAddMode 检查加群方式
func (g *groupInfoService) CheckGroupAddMode(ctx context.Context, groupId string) (string, int8, int) {
	db := dao.NewDbClient(ctx)
	group := model.GroupInfo{}
	if err := db.Where("uuid = ?", groupId).First(&group).Error; err != nil {
		zlog.GetLogger().Error("查询群加群方式失败", zap.Error(err), zap.String("group_id", groupId))
		return constants.SYSTEM_ERROR, -1, -1
	}
	return "查询群加群方式成功", group.AddMode, 0
}

// 直接入群
func (g *groupInfoService) EnterGroupDirectly(ctx context.Context, groupId, userId string) (string, int) {
	group := model.GroupInfo{}
	db := dao.NewDbClient(ctx)

	if err := db.First(&group, "uuid=?", groupId).Error; err != nil {
		zlog.GetLogger().Error("查询群信息失败", zap.Error(err), zap.String("group_id", groupId))
		return constants.SYSTEM_ERROR, -1
	}

	if group.AddMode != 0 {
		return "该群不支持直接入群", -2
	}

	var members []string
	if err := json.Unmarshal(group.Members, &members); err != nil {
		zlog.GetLogger().Error("反序列化群成员列表失败", zap.Error(err), zap.String("members", string(group.Members)))
		return constants.SYSTEM_ERROR, -1
	}
	// 检查用户是否已经在群中
	for _, member := range members {
		if member == userId {
			return "用户已经在群中", -2
		}
	}
	members = append(members, userId)
	newMembersData, err := json.Marshal(members)
	if err != nil {
		zlog.GetLogger().Error("序列化群成员列表失败", zap.Error(err))
		return constants.SYSTEM_ERROR, -1
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.GroupInfo{}).Where("uuid = ?", groupId).Updates(map[string]interface{}{
			"member_cnt": gorm.Expr("member_cnt + ?", 1),
			"members":    newMembersData,
		}).Error; err != nil {
			return err
		}
		return tx.Create(&model.UserContact{
			UserId:      userId,
			ContactId:   groupId,
			ContactType: contact_type_enum.GROUP,
			Status:      contact_status_enum.NORMAL,
		}).Error
	})
	if err != nil {
		zlog.GetLogger().Error("入群事务失败", zap.Error(err), zap.String("group_id", groupId))
		return constants.SYSTEM_ERROR, -1
	}

	// 清理缓存
	if err := myredis.DelKeyWithPatternIfExist("group_info_" + groupId + "*"); err != nil {
		zlog.GetLogger().Error("删除群信息缓存失败", zap.Error(err), zap.String("group_id", groupId))
	}
	if err := myredis.DelKeyWithPatternIfExist("group_members_" + groupId + "*"); err != nil {
		zlog.GetLogger().Error("删除群成员列表缓存失败", zap.Error(err), zap.String("group_id", groupId))
	}

	return "入群成功", 0
}

// LeaveGroup 离开群
func (g *groupInfoService) LeaveGroup(ctx context.Context, groupId, userId string) (string, int) {
	group := model.GroupInfo{}
	db := dao.NewDbClient(ctx)
	if err := db.First(&group, "uuid=?", groupId).Error; err != nil {
		zlog.GetLogger().Error("查询群信息失败", zap.Error(err), zap.String("group_id", groupId))
		return constants.SYSTEM_ERROR, -1
	}
	var members []string
	if err := json.Unmarshal(group.Members, &members); err != nil {
		zlog.GetLogger().Error("反序列化群成员列表失败", zap.Error(err), zap.String("members", string(group.Members)))
		return constants.SYSTEM_ERROR, -1
	}
	// 群主不能离开群
	if group.OwnerId == userId {
		return "群主不能离开群", -2
	}

	found := false
	for i, member := range members {
		if member == userId {
			members = append(members[:i], members[i+1:]...)
			found = true
			break
		}
	}
	if !found {
		return "用户不在群中", -2
	}

	newMembersData, err := json.Marshal(members)
	if err != nil {
		zlog.GetLogger().Error("序列化群成员列表失败", zap.Error(err))
		return constants.SYSTEM_ERROR, -1
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.GroupInfo{}).Where("uuid = ?", groupId).Updates(map[string]interface{}{
			"member_cnt": gorm.Expr("member_cnt - ?", 1),
			"members":    newMembersData,
		}).Error; err != nil {
			return err
		}
		if err := tx.Where("pair_key = ?", groupId).Delete(&model.Session{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id=? AND contact_id=? AND contact_type=?", userId, group.Uuid, contact_type_enum.GROUP).Delete(&model.UserContact{}).Error; err != nil {
			return err
		}
		return tx.Where("user_id=? AND contact_id=?", userId, group.Uuid).Delete(&model.ContactApply{}).Error
	})
	if err != nil {
		zlog.GetLogger().Error("退群事务失败", zap.Error(err), zap.String("group_id", groupId))
		return constants.SYSTEM_ERROR, -1
	}

	// 清理缓存
	if err := myredis.DelKeyWithPatternIfExist("group_info_" + groupId + "*"); err != nil {
		zlog.GetLogger().Error("删除群信息缓存失败", zap.Error(err), zap.String("group_id", groupId))
	}
	if err := myredis.DelKeyWithPatternIfExist("group_members_" + groupId + "*"); err != nil {
		zlog.GetLogger().Error("删除群成员列表缓存失败", zap.Error(err), zap.String("group_id", groupId))
	}
	return "离开群成功", 0
}

// DismissGroup 解散群
func (g *groupInfoService) DismissGroup(ctx context.Context, groupId, userId string) (string, int) {
	group := model.GroupInfo{}
	db := dao.NewDbClient(ctx)

	if err := db.First(&group, "uuid=?", groupId).Error; err != nil {
		zlog.GetLogger().Error("查询群信息失败", zap.Error(err), zap.String("group_id", groupId))
		return constants.SYSTEM_ERROR, -1
	}
	if group.OwnerId != userId {
		return "只有群主才能解散群", -2
	}
	err := db.Transaction(func(tx *gorm.DB) error {
		//  更新群状态为禁用，并清空成员信息
		emptyMembers, _ := json.Marshal([]string{})
		if err := tx.Model(&model.GroupInfo{}).Where("uuid = ?", groupId).Updates(map[string]interface{}{
			"status":     group_status_enum.DISABLE,
			"member_cnt": 0,
			"members":    emptyMembers,
		}).Error; err != nil {
			return err
		}
		// 批量删除会话
		if err := tx.Where("receive_id = ?", group.Uuid).Delete(&model.Session{}).Error; err != nil {
			return err
		}
		// 批量删除群组联系人
		if err := tx.Where("contact_id = ? AND contact_type = ?", group.Uuid, contact_type_enum.GROUP).Delete(&model.UserContact{}).Error; err != nil {
			return err
		}
		// 批量删除申请记录
		return tx.Where("contact_id = ?", group.Uuid).Delete(&model.ContactApply{}).Error
	})
	if err != nil {
		zlog.GetLogger().Error("解散群事务失败", zap.Error(err), zap.String("group_id", groupId))
		return constants.SYSTEM_ERROR, -1
	}

	// 清理缓存
	if err := myredis.DelKeyWithPatternIfExist("group_info_" + groupId + "*"); err != nil {
		zlog.GetLogger().Error("删除群信息缓存失败", zap.Error(err), zap.String("group_id", groupId))
	}
	if err := myredis.DelKeyWithPatternIfExist("group_members_" + groupId + "*"); err != nil {
		zlog.GetLogger().Error("删除群成员列表缓存失败", zap.Error(err), zap.String("group_id", groupId))
	}
	return "解散群成功", 0
}

// GetGroupInfo 获取群信息
func (g *groupInfoService) GetGroupInfo(ctx context.Context, groupId string) (string, *response.GetGroupInfoResponse, int) {
	groupInfo := model.GroupInfo{}
	db := dao.NewDbClient(ctx)
	if err := db.Where("uuid = ?", groupId).First(&groupInfo).Error; err != nil {
		zlog.GetLogger().Error("查询群信息失败", zap.Error(err), zap.String("group_id", groupId))
		return constants.SYSTEM_ERROR, nil, -1
	}
	resp := &response.GetGroupInfoResponse{
		GroupId:   groupInfo.Uuid,
		GroupName: groupInfo.Name,
		Avatar:    groupInfo.Avatar,
		Notice:    groupInfo.Notice,
		AddMode:   groupInfo.AddMode,
		OwnerId:   groupInfo.OwnerId,
		MemberCnt: groupInfo.MemberCnt,
		CreatedAt: groupInfo.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: groupInfo.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	return "查询群信息成功", resp, 0
}

func (g *groupInfoService) UpdateGroupInfo(ctx context.Context, req request.UpdateGroupInfoRequest, operatorId string) (string, int) {
	groupInfo := model.GroupInfo{}
	db := dao.NewDbClient(ctx)
	if err := db.Where("uuid = ?", req.GroupId).First(&groupInfo).Error; err != nil {
		zlog.GetLogger().Error("查询群信息失败", zap.Error(err), zap.String("group_id", req.GroupId))
		return constants.SYSTEM_ERROR, -1
	}
	// 只有群主才能更新群信息
	if groupInfo.OwnerId != operatorId {
		return "只有群主才能更新群信息", -2
	}

	if req.GroupName != nil {
		groupInfo.Name = *req.GroupName
	}

	if req.Avatar != nil {
		groupInfo.Avatar = *req.Avatar
	}
	if req.Notice != nil {
		groupInfo.Notice = *req.Notice
	}
	if req.AddMode != nil {
		groupInfo.AddMode = *req.AddMode
	}

	if err := db.Save(&groupInfo).Error; err != nil {
		zlog.GetLogger().Error("更新群信息失败", zap.Error(err), zap.String("group_id", req.GroupId))
		return constants.SYSTEM_ERROR, -1
	}

	sessionList := []model.Session{}
	if err := db.Where("receive_id=?", groupInfo.Uuid).Find(&sessionList).Error; err != nil {
		zlog.GetLogger().Error("查询会话列表失败", zap.Error(err), zap.String("receive_id", groupInfo.Uuid))
		return constants.SYSTEM_ERROR, -1
	}
	for _, session := range sessionList {
		session.ReceiveName = groupInfo.Name
		session.Avatar = groupInfo.Avatar
		if err := db.Save(&session).Error; err != nil {
			zlog.GetLogger().Error("更新会话信息失败", zap.Error(err), zap.String("session_id", session.Uuid))
			return constants.SYSTEM_ERROR, -1
		}
	}

	// 清理缓存
	if err := myredis.DelKeyWithPatternIfExist("group_info_" + req.GroupId + "*"); err != nil {
		zlog.GetLogger().Error("删除群信息缓存失败", zap.Error(err), zap.String("group_id", req.GroupId))
		return constants.SYSTEM_ERROR, -1
	}

	return "更新群信息成功", 0
}

func (g *groupInfoService) GetGroupMembers(ctx context.Context, groupId string) (string, []response.GetGroupMembersResponse, int) {
	groupInfo := model.GroupInfo{}
	db := dao.NewDbClient(ctx)
	if err := db.Where("uuid = ?", groupId).First(&groupInfo).Error; err != nil {
		zlog.GetLogger().Error("查询群信息失败", zap.Error(err), zap.String("group_id", groupId))
		return constants.SYSTEM_ERROR, nil, -1
	}
	var members []string
	if err := json.Unmarshal(groupInfo.Members, &members); err != nil {
		zlog.GetLogger().Error("反序列化群成员列表失败", zap.Error(err), zap.String("members", string(groupInfo.Members)))
		return constants.SYSTEM_ERROR, nil, -1
	}
	if len(members) == 0 {
		return "查询群成员列表成功", []response.GetGroupMembersResponse{}, 0
	}
	// 批量查询，避免 N+1
	var userInfos []model.UserInfo
	if err := db.Where("uuid IN ?", members).Find(&userInfos).Error; err != nil {
		zlog.GetLogger().Error("批量查询群成员信息失败", zap.Error(err), zap.String("group_id", groupId))
		return constants.SYSTEM_ERROR, nil, -1
	}
	userMap := make(map[string]model.UserInfo, len(userInfos))
	for _, u := range userInfos {
		userMap[u.Uuid] = u
	}
	memberInfos := make([]response.GetGroupMembersResponse, 0, len(members))
	for _, memberId := range members {
		if u, ok := userMap[memberId]; ok {
			memberInfos = append(memberInfos, response.GetGroupMembersResponse{
				UserId:   u.Uuid,
				Nickname: u.Nickname,
				Avatar:   u.Avatar,
			})
		}
	}
	return "查询群成员列表成功", memberInfos, 0
}

func (g *groupInfoService) RemoveGroupMember(ctx context.Context, groupId, userId, operatorId string) (string, int) {
	groupInfo := model.GroupInfo{}
	db := dao.NewDbClient(ctx)
	if err := db.Where("uuid = ?", groupId).First(&groupInfo).Error; err != nil {
		zlog.GetLogger().Error("查询群信息失败", zap.Error(err), zap.String("group_id", groupId))
		return constants.SYSTEM_ERROR, -1
	}
	if groupInfo.OwnerId != operatorId {
		return "只有群主才能移除群成员", -2
	}
	if userId == groupInfo.OwnerId {
		return "群主不能被移除", -2
	}
	var members []string
	if err := json.Unmarshal(groupInfo.Members, &members); err != nil {
		zlog.GetLogger().Error("反序列化群成员列表失败", zap.Error(err), zap.String("members", string(groupInfo.Members)))
		return constants.SYSTEM_ERROR, -1
	}
	found := false
	for i, member := range members {
		if member == userId {
			members = append(members[:i], members[i+1:]...)
			found = true
			break
		}
	}
	if !found {
		return "用户不在群中", -2
	}
	newMembersData, err := json.Marshal(members)
	if err != nil {
		zlog.GetLogger().Error("序列化群成员列表失败", zap.Error(err))
		return constants.SYSTEM_ERROR, -1
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.GroupInfo{}).Where("uuid = ?", groupId).Updates(map[string]interface{}{
			"member_cnt": gorm.Expr("member_cnt - ?", 1),
			"members":    newMembersData,
		}).Error; err != nil {
			return err
		}
		if err := tx.Where("pair_key = ?", groupId).Delete(&model.Session{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id=? AND contact_id=? AND contact_type=?", userId, groupInfo.Uuid, contact_type_enum.GROUP).Delete(&model.UserContact{}).Error; err != nil {
			return err
		}
		return tx.Where("user_id=? AND contact_id=?", userId, groupInfo.Uuid).Delete(&model.ContactApply{}).Error
	})
	if err != nil {
		zlog.GetLogger().Error("移除群成员事务失败", zap.Error(err), zap.String("group_id", groupId))
		return constants.SYSTEM_ERROR, -1
	}

	// 清理缓存（非致命）
	if err := myredis.DelKeyWithPatternIfExist("group_info_" + groupId + "*"); err != nil {
		zlog.GetLogger().Error("删除群信息缓存失败", zap.Error(err), zap.String("group_id", groupId))
	}
	if err := myredis.DelKeyWithPatternIfExist("group_members_" + groupId + "*"); err != nil {
		zlog.GetLogger().Error("删除群成员列表缓存失败", zap.Error(err), zap.String("group_id", groupId))
	}
	return "移除群成员成功", 0
}
