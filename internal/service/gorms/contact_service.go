package gormss

import (
	"GoLinko/internal/dao"
	"GoLinko/internal/dto/request"
	"GoLinko/internal/dto/response"
	"GoLinko/internal/model"
	"GoLinko/internal/service/chat"
	myredis "GoLinko/internal/service/redis"
	constants "GoLinko/pkg/const"
	"GoLinko/pkg/enum/contact/contact_status_enum"
	"GoLinko/pkg/enum/contact/contact_type_enum"
	"GoLinko/pkg/enum/contact_apply/contact_apply_status_enum"
	"GoLinko/pkg/enum/group_info/group_status_enum"
	"GoLinko/pkg/enum/message/message_status_enum"
	"GoLinko/pkg/enum/message/message_type_enum"
	"GoLinko/pkg/enum/user_info/user_status_enum"
	"GoLinko/pkg/utils"
	"GoLinko/pkg/zlog"
	"context"
	"encoding/json"
	"errors"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type contactInfoService struct {
}

var ContactInfoService = &contactInfoService{}

// GetContactUserList 获取联系人列表
func (c *contactInfoService) GetContactUserList(ctx context.Context, userId string) (string, []response.UserListResponse, int) {
	var contactList []model.UserContact
	db := dao.NewDbClient(ctx)

	if err := db.Where("user_id = ? AND status NOT IN (?) AND contact_type = ?",
		userId, []int8{4, 3}, contact_type_enum.USER).Find(&contactList).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			msg := "目前没有联系人"
			zlog.GetLogger().Info(msg, zap.String("userId", userId))
			return msg, nil, 0
		} else {
			zlog.GetLogger().Error("查询联系人列表失败", zap.String("userId", userId), zap.Error(err))
			return "查询联系人列表失败", nil, -1
		}
	}

	if len(contactList) == 0 {
		return "目前没有联系人", nil, 0
	}

	// 提取所有联系人UUID，一次性联查，避免 N+1
	contactIds := make([]string, len(contactList))
	for i, contact := range contactList {
		contactIds[i] = contact.ContactId
	}

	var userInfos []model.UserInfo
	if err := db.Where("uuid IN ?", contactIds).Find(&userInfos).Error; err != nil {
		zlog.GetLogger().Error("联查联系人信息失败", zap.Error(err))
		return "查询联系人列表失败", nil, -1
	}

	resp := make([]response.UserListResponse, 0, len(userInfos))
	for _, userInfo := range userInfos {
		resp = append(resp, response.UserListResponse{
			UserId:   userInfo.Uuid,
			Nickname: userInfo.Nickname,
			Avatar:   userInfo.Avatar,
		})
	}

	return "查询联系人列表成功", resp, 0
}

// LoadMyJoinedGroups 获取用户加入的群组列表
func (c *contactInfoService) LoadMyJoinedGroups(ctx context.Context, userId string) (string, []response.LoadMyJoinedGroupsResponse, int) {
	// 查询用户加入的群组列表
	var contactList []model.UserContact
	db := dao.NewDbClient(ctx)
	// 过滤退群或被踢，且只查群聊类型
	if err := db.Order("created_at desc").Where("user_id = ? AND contact_type = ? AND status NOT IN (6, 7)",
		userId, contact_type_enum.GROUP).Find(&contactList).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			msg := "未加入任何群组"
			zlog.GetLogger().Info(msg, zap.String("userId", userId))
			return msg, nil, 0
		} else {
			zlog.GetLogger().Error("查询用户加入的群组列表失败", zap.String("userId", userId), zap.Error(err))
			return "查询用户加入的群组列表失败", nil, -1
		}
	}

	if len(contactList) == 0 {
		return "未加入任何群组", nil, 0
	}

	// 提取群组ID，一次性联查
	groupIds := make([]string, 0, len(contactList))
	for _, contact := range contactList {
		groupIds = append(groupIds, contact.ContactId)
	}

	var groupInfos []model.GroupInfo
	if err := db.Where("uuid IN ?", groupIds).Find(&groupInfos).Error; err != nil {
		zlog.GetLogger().Error("查询群信息列表失败", zap.Error(err))
		return "查询群信息列表失败", nil, -1
	}

	resp := make([]response.LoadMyJoinedGroupsResponse, 0, len(groupInfos))
	for _, groupInfo := range groupInfos {
		// 过滤掉自己是创建者的群组 (业务逻辑：如果是“加入”的群组，通常不包含自己创建的)
		if groupInfo.OwnerId != userId {
			resp = append(resp, response.LoadMyJoinedGroupsResponse{
				GroupId:   groupInfo.Uuid,
				GroupName: groupInfo.Name,
				Avatar:    groupInfo.Avatar,
			})
		}
	}
	return "获取已加入的群组成功", resp, 0
}

// GetContactInfo 获取联系人信息
func (c *contactInfoService) GetContactInfo(ctx context.Context, userId string, contactId string) (string, *response.GetContactInfoResponse, int) {
	db := dao.NewDbClient(ctx)
	var contact model.UserContact
	if err := db.Where("user_id = ? AND contact_id = ?", userId, contactId).First(&contact).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			msg := "联系人不存在"
			zlog.GetLogger().Info(msg, zap.String("userId", userId), zap.String("contactId", contactId))
			return msg, nil, 0
		} else {
			zlog.GetLogger().Error("查询联系人信息失败", zap.String("userId", userId), zap.String("contactId", contactId), zap.Error(err))
			return "查询联系人信息失败", nil, -1
		}
	}
	if contact.ContactType == 0 {
		// 用户
		var userInfo model.UserInfo
		if err := db.Where("uuid = ?", contactId).First(&userInfo).Error; err != nil {
			zlog.GetLogger().Error("查询联系人信息失败", zap.String("contactId", contactId), zap.Error(err))
			return "查询联系人信息失败", nil, -1
		}
		resp := &response.GetContactInfoResponse{
			ContactId:        userInfo.Uuid,
			ContactName:      userInfo.Nickname,
			ContactAvatar:    userInfo.Avatar,
			ContactPhone:     userInfo.Telephone,
			ContactEmail:     userInfo.Email,
			ContactGender:    userInfo.Gender,
			ContactSignature: userInfo.Signature,
			ContactBirthday:  userInfo.Birthday,
		}
		if userInfo.Status == user_status_enum.DISABLE {
			zlog.GetLogger().Info("查询联系人信息成功，但联系人已禁用", zap.String("contactId", contactId))
			resp.ContactName = resp.ContactName + "（已禁用）"
		}

		return "查询联系人信息成功", resp, 0

	} else {
		// 群聊
		var groupInfo model.GroupInfo
		if err := db.Where("uuid = ?", contactId).First(&groupInfo).Error; err != nil {
			zlog.GetLogger().Error("查询群信息失败", zap.String("contactId", contactId), zap.Error(err))
			return "查询群信息失败", nil, -1
		}
		resp := &response.GetContactInfoResponse{
			ContactId:        groupInfo.Uuid,
			ContactName:      groupInfo.Name,
			ContactAvatar:    groupInfo.Avatar,
			ContactNotice:    groupInfo.Notice,
			ContactAddMode:   groupInfo.AddMode,
			ContactOwnerId:   groupInfo.OwnerId,
			ContactMemberCnt: groupInfo.MemberCnt,
		}
		if groupInfo.Status == group_status_enum.DISABLE {
			zlog.GetLogger().Info("查询群信息成功，但群已禁用", zap.String("contactId", contactId))
			resp.ContactName = resp.ContactName + "（已禁用）"
		}
		return "查询群信息成功", resp, 0
	}
}

// DeleteContact 删除联系人（仅用户）
func (c *contactInfoService) DeleteContact(ctx context.Context, userId string, contactId string) (string, int) {
	db := dao.NewDbClient(ctx)

	// 事务执行删除操作，确保数据一致性
	err := db.Transaction(func(tx *gorm.DB) error {
		// 删除用户联系人记录 (物理删除 Unscoped 以防重新添加时唯一索引冲突)
		if err := tx.Unscoped().Where("user_id = ? AND contact_id = ?", userId, contactId).Delete(&model.UserContact{}).Error; err != nil {
			zlog.GetLogger().Error("删除联系人失败", zap.String("userId", userId), zap.String("contactId", contactId), zap.Error(err))
			return err
		}
		if err := tx.Unscoped().Where("user_id = ? AND contact_id = ?", contactId, userId).Delete(&model.UserContact{}).Error; err != nil {
			zlog.GetLogger().Error("删除联系人失败", zap.String("userId", userId), zap.String("contactId", contactId), zap.Error(err))
			return err
		}
		// 删除用户会话记录
		pairKey := SessionService.getPairKey(userId, contactId)
		if err := tx.Unscoped().Where("pair_key = ?", pairKey).Delete(&model.Session{}).Error; err != nil {
			zlog.GetLogger().Error("删除会话失败", zap.String("pairKey", pairKey), zap.Error(err))
			return err
		}
		// 删除好友申请记录
		if err := tx.Unscoped().Where("user_id = ? AND contact_id = ?", userId, contactId).Delete(&model.ContactApply{}).Error; err != nil {
			zlog.GetLogger().Error("删除申请记录失败", zap.String("userId", userId), zap.String("contactId", contactId), zap.Error(err))
			return err
		}
		if err := tx.Unscoped().Where("user_id = ? AND contact_id = ?", contactId, userId).Delete(&model.ContactApply{}).Error; err != nil {
			zlog.GetLogger().Error("删除申请记录失败", zap.String("userId", userId), zap.String("contactId", contactId), zap.Error(err))
			return err
		}
		return nil
	})

	if err != nil {
		zlog.GetLogger().Error("事务执行失败：删除联系人", zap.String("userId", userId), zap.String("contactId", contactId), zap.Error(err))
		return "删除联系人失败", -1
	}

	return "删除联系人成功", 0
}

// ApplyAddContact 申请添加联系人或群聊
func (c *contactInfoService) ApplyAddContact(ctx context.Context, userId string, req request.ApplyAddContactRequest) (string, int) {
	db := dao.NewDbClient(ctx)

	//检查是否给自己发送申请
	if userId == req.ContactId {
		return "不能给自己发送申请", -2
	}

	// 1. 根据首字母快速判断目标类型 (U: 用户, G: 群组)
	if len(req.ContactId) == 0 {
		return "目标ID不能为空", -2
	}

	var contactType int8
	if req.ContactId[0] == 'G' {
		var groupInfo model.GroupInfo
		if err := db.Where("uuid = ?", req.ContactId).First(&groupInfo).Error; err != nil {
			return "群聊不存在", -2
		}
		if groupInfo.Status == group_status_enum.DISABLE {
			return "群聊已禁用", -2
		}
		contactType = contact_type_enum.GROUP
	} else if req.ContactId[0] == 'U' {
		var userInfo model.UserInfo
		if err := db.Where("uuid = ?", req.ContactId).First(&userInfo).Error; err != nil {
			return "用户不存在", -2
		}
		if userInfo.Status == user_status_enum.DISABLE {
			return "用户已禁用", -2
		}
		contactType = contact_type_enum.USER
	} else {
		return "非法的ID格式", -2
	}

	// 2. 检查是否已经是好友或已在群中
	var existingContact model.UserContact
	if err := db.Where("user_id = ? AND contact_id = ?", userId, req.ContactId).First(&existingContact).Error; err == nil {
		// 状态逻辑处理 (比如被删可重新加，但正常、拉黑等状态不应重复加)
		if req.ContactId[0] == 'U' && existingContact.Status != 3 && existingContact.Status != 4 {
			return "对方已在你的联系人列表中", -2
		}
		if req.ContactId[0] == 'G' && existingContact.Status != 6 && existingContact.Status != 7 {
			return "你已在该群聊中", -2
		}
	}

	// 3. 处理申请记录
	var contactApply model.ContactApply
	err := db.Where("user_id = ? AND contact_id = ?", userId, req.ContactId).First(&contactApply).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 创建新申请
		newApply := model.ContactApply{
			Uuid:        utils.GenerateApplyID(),
			UserId:      userId,
			ContactId:   req.ContactId,
			Message:     req.Message,
			Status:      contact_apply_status_enum.PENDING,
			ContactType: contactType,
			LastApplyAt: time.Now(),
		}
		if err := db.Create(&newApply).Error; err != nil {
			zlog.GetLogger().Error("创建申请记录失败", zap.Error(err))
			return constants.SYSTEM_ERROR, -1
		}
	} else if err == nil {
		// 如果被拉黑，不允许申请
		if contactApply.Status == contact_apply_status_enum.BLACK {
			return "你已被对方拉黑，无法申请", -2
		}
		// 更新现有申请（覆盖消息，更新时间，重置为待处理）
		contactApply.Message = req.Message
		contactApply.Status = contact_apply_status_enum.PENDING
		contactApply.LastApplyAt = time.Now()
		if err := db.Save(&contactApply).Error; err != nil {
			zlog.GetLogger().Error("更新申请记录失败", zap.Error(err))
			return constants.SYSTEM_ERROR, -1
		}
	} else {
		zlog.GetLogger().Error("查询申请记录失败", zap.Error(err))
		return constants.SYSTEM_ERROR, -1
	}

	if req.ContactId[0] == 'G' {
		return "申请已提交，请等待管理员审核", 0
	}
	return "申请已提交，请等待对方审核", 0
}

// GetContactApplyList 获取联系人/群聊申请列表 (获取自己收到的好友申请和自己创建群组收到的入群申请)
func (c *contactInfoService) GetContactApplyList(ctx context.Context, userId string) (string, []response.ContactApplyListResponse, int) {
	db := dao.NewDbClient(ctx)

	// 1. 查找用户自己创建的所有群组 UUID
	var myGroups []model.GroupInfo
	db.Where("owner_id = ?", userId).Find(&myGroups)
	groupIds := make([]string, 0, len(myGroups))
	for _, g := range myGroups {
		groupIds = append(groupIds, g.Uuid)
	}

	// 2. 查找申请记录：
	// - 申请目标是用户自己
	// - 或者申请目标是用户拥有的群组
	// 并且状态为待处理 (PENDING)
	var applies []model.ContactApply
	query := db.Where("((contact_id = ? AND contact_type = ?) OR (contact_id IN ? AND contact_type = ?)) AND status = ?",
		userId, contact_type_enum.USER, groupIds, contact_type_enum.GROUP, contact_apply_status_enum.PENDING)

	if err := query.Order("last_apply_at desc").Find(&applies).Error; err != nil {
		zlog.GetLogger().Error("查询申请列表失败", zap.Error(err))
		return "查询申请列表失败", nil, -1
	}

	if len(applies) == 0 {
		return "获取申请列表成功", []response.ContactApplyListResponse{}, 0
	}

	// 3. 批量查询申请人信息
	applicantIds := make([]string, 0, len(applies))
	targetGroupIds := make([]string, 0)
	for _, apply := range applies {
		applicantIds = append(applicantIds, apply.UserId)
		if apply.ContactType == contact_type_enum.GROUP {
			targetGroupIds = append(targetGroupIds, apply.ContactId)
		}
	}

	var applicants []model.UserInfo
	if err := db.Where("uuid IN ?", applicantIds).Find(&applicants).Error; err != nil {
		zlog.GetLogger().Error("批量查询申请人信息失败", zap.Error(err))
		return constants.SYSTEM_ERROR, nil, -1
	}
	applicantMap := make(map[string]model.UserInfo, len(applicants))
	for _, a := range applicants {
		applicantMap[a.Uuid] = a
	}

	groupNameMap := make(map[string]string)
	if len(targetGroupIds) > 0 {
		var targetGroups []model.GroupInfo
		if err := db.Where("uuid IN ?", targetGroupIds).Find(&targetGroups).Error; err != nil {
			zlog.GetLogger().Error("批量查询目标群组信息失败", zap.Error(err))
		} else {
			for _, g := range targetGroups {
				groupNameMap[g.Uuid] = g.Name
			}
		}
	}

	// 4. 构建响应
	resp := make([]response.ContactApplyListResponse, 0, len(applies))
	for _, apply := range applies {
		applicant, ok := applicantMap[apply.UserId]
		if !ok {
			continue
		}
		targetName := "个人"
		if apply.ContactType == contact_type_enum.GROUP {
			if name, ok := groupNameMap[apply.ContactId]; ok {
				targetName = name
			} else {
				targetName = "未知群组"
			}
		}
		msg := apply.Message
		if msg == "" {
			msg = "对方没有留下申请消息"
		}
		resp = append(resp, response.ContactApplyListResponse{
			ApplyId:     apply.Uuid,
			ApplicantId: applicant.Uuid,
			Applicant:   applicant.Nickname,
			Avatar:      applicant.Avatar,
			Message:     "申请理由:" + msg,
			ContactType: apply.ContactType,
			TargetName:  targetName,
			CreatedAt:   apply.LastApplyAt.Format("2006-01-02 15:04:05"),
		})
	}
	return "获取申请列表成功", resp, 0
}

// AcceptAddContact 同意添加联系人或群组申请
func (c *contactInfoService) AcceptAddContact(ctx context.Context, userId string, applyId string) (string, int) {
	db := dao.NewDbClient(ctx)
	var notifyUserID string
	var notifyMessage []byte
	var notifyMessageID string

	// 1. 查找申请记录
	var apply model.ContactApply
	if err := db.Where("uuid = ?", applyId).First(&apply).Error; err != nil {
		return "申请记录不存在", -2
	}

	if apply.Status != contact_apply_status_enum.PENDING {
		return "申请已处理", -2
	}

	// 2.只有被申请人或群主才能同意
	if apply.ContactType == contact_type_enum.USER {
		if apply.ContactId != userId {
			return "无权进行此操作", -2
		}
	} else {
		var group model.GroupInfo
		if err := db.Where("uuid = ?", apply.ContactId).First(&group).Error; err != nil || group.OwnerId != userId {
			return "只有群主可以同意入群申请", -2
		}
	}

	// 3. 事务处理
	err := db.Transaction(func(tx *gorm.DB) error {
		// 更新申请状态
		if err := tx.Model(&apply).Update("status", contact_apply_status_enum.AGREE).Error; err != nil {
			return err
		}

		if apply.ContactType == contact_type_enum.USER {
			// 好友逻辑：创建双向联系人记录（若已存在则跳过，防止重复添加）
			contacts := []model.UserContact{
				{UserId: apply.UserId, ContactId: apply.ContactId, ContactType: int8(contact_type_enum.USER), Status: int8(contact_status_enum.NORMAL)},
				{UserId: apply.ContactId, ContactId: apply.UserId, ContactType: int8(contact_type_enum.USER), Status: int8(contact_status_enum.NORMAL)},
			}
			for _, contact := range contacts {
				var existing model.UserContact
				err := tx.Where("user_id = ? AND contact_id = ?", contact.UserId, contact.ContactId).First(&existing).Error
				if errors.Is(err, gorm.ErrRecordNotFound) {
					if err := tx.Create(&contact).Error; err != nil {
						return err
					}
				} else if err != nil {
					return err
				} else {
					// 已存在则恢复为正常状态
					if err := tx.Model(&existing).Update("status", contact_status_enum.NORMAL).Error; err != nil {
						return err
					}
				}
			}

			// 创建唯一会话
			var applicant, acceptor model.UserInfo
			tx.Where("uuid = ?", apply.UserId).First(&applicant)
			tx.Where("uuid = ?", apply.ContactId).First(&acceptor)

			// 采用标准化的 pair_key
			pairKey := SessionService.getPairKey(apply.UserId, apply.ContactId)

			var existingSession model.Session
			checkErr := tx.Unscoped().Where("pair_key = ?", pairKey).First(&existingSession).Error
			sessionUuid := ""
			if errors.Is(checkErr, gorm.ErrRecordNotFound) {
				sessionUuid = utils.GenerateSessionID()
				newSession := model.Session{
					Uuid:        sessionUuid,
					PairKey:     pairKey,
					SendId:      apply.UserId,
					ReceiveId:   apply.ContactId,
					ReceiveName: acceptor.Nickname,
					Avatar:      acceptor.Avatar,
				}
				if err := tx.Create(&newSession).Error; err != nil {
					return err
				}
			} else if checkErr != nil {
				return checkErr
			} else {
				sessionUuid = existingSession.Uuid
				// 如果是软删除的，恢复它
				if err := tx.Unscoped().Model(&existingSession).Updates(map[string]interface{}{
					"deleted_at": nil,
				}).Error; err != nil {
					return err
				}
			}

			// 添加自动打招呼消息
			greetMsg := model.Message{
				Uuid:       utils.GenerateMessageID(),
				SessionId:  sessionUuid,
				Type:       message_type_enum.Text,
				Content:    "我已添加你为好友，现在可以开始聊天了",
				SendId:     apply.ContactId, // 被申请人发送
				SendName:   acceptor.Nickname,
				SendAvatar: acceptor.Avatar,
				ReceiveId:  apply.UserId,
				Status:     message_status_enum.Sent,
			}
			if err := tx.Create(&greetMsg).Error; err != nil {
				return err
			}

			notifyUserID = apply.UserId
			notifyMessageID = greetMsg.Uuid
			createdAt := greetMsg.CreatedAt.Format("2006-01-02 15:04:05")
			if createdAt == "0001-01-01 00:00:00" {
				createdAt = time.Now().Format("2006-01-02 15:04:05")
			}
			notifyPayload := map[string]any{
				"session_id":  sessionUuid,
				"send_id":     apply.ContactId,
				"send_name":   acceptor.Nickname,
				"send_avatar": acceptor.Avatar,
				"receive_id":  apply.UserId,
				"type":        message_type_enum.Text,
				"content":     greetMsg.Content,
				"created_at":  createdAt,
			}
			b, _ := json.Marshal(notifyPayload)
			notifyMessage = b
			// 清理缓存
			_ = myredis.DelKeyWithPatternIfExist("message_list_*_" + sessionUuid)
			_ = myredis.DelKeyWithPatternIfExist("session_list_" + apply.UserId)
			_ = myredis.DelKeyWithPatternIfExist("session_list_" + apply.ContactId)

		} else {
			// 群组逻辑：创建群成员记录
			newMember := model.UserContact{
				UserId:      apply.UserId,
				ContactId:   apply.ContactId,
				ContactType: int8(contact_type_enum.GROUP),
				Status:      int8(contact_status_enum.NORMAL),
			}
			if err := tx.Create(&newMember).Error; err != nil {
				return err
			}

			// E. 更新群成员数量和成员列表
			var groupInfo model.GroupInfo
			if err := tx.Where("uuid = ?", apply.ContactId).First(&groupInfo).Error; err != nil {
				return err
			}

			var members []string
			if err := json.Unmarshal(groupInfo.Members, &members); err != nil {
				return err
			}
			members = append(members, apply.UserId)
			newMembersJson, _ := json.Marshal(members)

			if err := tx.Model(&model.GroupInfo{}).Where("uuid = ?", apply.ContactId).Updates(map[string]interface{}{
				"member_cnt": gorm.Expr("member_cnt + ?", 1),
				"members":    newMembersJson,
			}).Error; err != nil {
				return err
			}

			// 加入群聊后发送欢迎消息
			var user model.UserInfo
			if err := tx.Where("uuid = ?", apply.UserId).First(&user).Error; err == nil {
				welcomeMsg := model.Message{
					Uuid:      utils.GenerateMessageID(),
					SessionId: apply.ContactId, // 群聊 sessionId 通常就是 groupId
					Type:      message_type_enum.Text,
					Content:   "欢迎 " + user.Nickname + " 加入群聊",
					SendId:    "SYSTEM",
					SendName:  "系统",
					ReceiveId: apply.ContactId,
					Status:    message_status_enum.Sent,
				}
				if err := tx.Create(&welcomeMsg).Error; err != nil {
					return err
				}
				// 清理群聊消息缓存
				_ = myredis.DelKeyWithPatternIfExist("group_message_list_" + apply.ContactId)
			}
		}

		return nil
	})

	if err != nil {
		zlog.GetLogger().Error("处理申请事务失败", zap.Error(err))
		return constants.SYSTEM_ERROR, -1
	}

	if notifyUserID != "" && len(notifyMessage) > 0 {
		chat.ChatServer.PushToUser(notifyUserID, chat.MessageBack{Message: notifyMessage, Uuid: notifyMessageID})
	}

	return "已同意申请", 0
}

// RejectAddContact 拒绝添加联系人或群组申请
func (c *contactInfoService) RejectAddContact(ctx context.Context, userId string, applyId string) (string, int) {
	db := dao.NewDbClient(ctx)

	// 1. 查找申请记录
	var apply model.ContactApply
	if err := db.Where("uuid = ?", applyId).First(&apply).Error; err != nil {
		return "申请记录不存在", -2
	}

	if apply.Status != contact_apply_status_enum.PENDING {
		return "申请已处理", -2
	}

	// 2. 权限校验
	if apply.ContactType == contact_type_enum.USER {
		if apply.ContactId != userId {
			return "无权进行此操作", -2
		}
	} else {
		var group model.GroupInfo
		if err := db.Where("uuid = ?", apply.ContactId).First(&group).Error; err != nil || group.OwnerId != userId {
			return "只有群主可以拒绝入群申请", -2
		}
	}

	// 3. 更新状态为拒绝
	if err := db.Model(&apply).Update("status", contact_apply_status_enum.REFUSE).Error; err != nil {
		zlog.GetLogger().Error("拒绝申请失败", zap.Error(err))
		return constants.SYSTEM_ERROR, -1
	}

	return "已拒绝该申请", 0
}

// BlackContact 拉黑联系人 (仅限个人好友)
func (c *contactInfoService) BlackContact(ctx context.Context, userId string, contactId string) (string, int) {
	db := dao.NewDbClient(ctx)

	// 首先检查是否为好友关系
	var contact model.UserContact
	if err := db.Where("user_id = ? AND contact_id = ? AND contact_type = ?",
		userId, contactId, contact_type_enum.USER).First(&contact).Error; err != nil {
		return "联系人不存在，无法拉黑", -2
	}

	// 事务处理：拉黑是双向状态变更（主动拉黑 vs 被拉黑）
	err := db.Transaction(func(tx *gorm.DB) error {
		// 1. 更新自己的记录状态为 BLACK (拉黑)
		if err := tx.Model(&model.UserContact{}).
			Where("user_id = ? AND contact_id = ?", userId, contactId).
			Update("status", contact_status_enum.BLACK).Error; err != nil {
			return err
		}

		// 2. 更新对方的记录状态为 BE_BLACK (被拉黑)
		if err := tx.Model(&model.UserContact{}).
			Where("user_id = ? AND contact_id = ?", contactId, userId).
			Update("status", contact_status_enum.BE_BLACK).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		zlog.GetLogger().Error("拉黑联系人事务失败", zap.Error(err))
		return constants.SYSTEM_ERROR, -1
	}

	return "拉黑成功", 0
}

// UnblackContact 解除拉黑联系人
func (c *contactInfoService) UnblackContact(ctx context.Context, userId string, contactId string) (string, int) {
	db := dao.NewDbClient(ctx)

	var contact model.UserContact
	if err := db.Where("user_id = ? AND contact_id = ? AND status = ?",
		userId, contactId, contact_status_enum.BLACK).First(&contact).Error; err != nil {
		return "该联系人不在黑名单中", -2
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		// 1. 恢复两边的状态为 NORMAL
		if err := tx.Model(&model.UserContact{}).
			Where("user_id = ? AND contact_id = ?", userId, contactId).
			Update("status", contact_status_enum.NORMAL).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.UserContact{}).
			Where("user_id = ? AND contact_id = ?", contactId, userId).
			Update("status", contact_status_enum.NORMAL).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		zlog.GetLogger().Error("解除拉黑事务失败", zap.Error(err))
		return constants.SYSTEM_ERROR, -1
	}

	return "已解除拉黑", 0
}
