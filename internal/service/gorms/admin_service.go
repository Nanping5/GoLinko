package gormss

import (
	"GoLinko/internal/dao"
	"GoLinko/internal/dto/response"
	"GoLinko/internal/model"
	"context"
	"time"
)

type adminService struct {
}

var AdminService = &adminService{}

func (s *adminService) GetUserList(ctx context.Context, userId string) (string, []response.GetUserListResponse, int) {
	var users []model.UserInfo
	db := dao.NewDbClient(ctx)
	if !checkAdmin(ctx, userId) {
		return "权限不足", nil, -2
	}
	if err := db.Order("created_at desc").Find(&users).Error; err != nil {
		return "查询用户列表失败", nil, -1
	}

	userList := make([]response.GetUserListResponse, 0, len(users))
	for _, item := range users {
		createdAt := ""
		if item.Model != nil && !item.CreatedAt.IsZero() {
			createdAt = item.CreatedAt.Format(time.DateTime)
		}
		userList = append(userList, response.GetUserListResponse{
			Uuid:      item.Uuid,
			Nickname:  item.Nickname,
			Telephone: item.Telephone,
			Email:     item.Email,
			IsAdmin:   item.IsAdmin == 1,
			Status:    item.Status,
			Avatar:    item.Avatar,
			CreatedAt: createdAt,
			Birthday:  item.Birthday,
			Gender:    item.Gender,
			Signature: item.Signature,
		})
	}

	return "查询用户列表成功", userList, 0
}

func (s *adminService) DisableUser(ctx context.Context, userId string, targetUserId string) (string, int) {
	db := dao.NewDbClient(ctx)
	if !checkAdmin(ctx, userId) {
		return "权限不足", -2
	}

	targetUser := model.UserInfo{}
	if err := db.Where("uuid = ?", targetUserId).First(&targetUser).Error; err != nil {
		return "查询目标用户失败", -1
	}

	if targetUser.Status == 1 {
		return "用户已被禁用", -2
	} else {
		targetUser.Status = 1
		if err := db.Save(&targetUser).Error; err != nil {
			return "禁用用户失败", -1
		}
	}
	return "用户禁用成功", 0
}

func (s *adminService) AbleUser(ctx context.Context, userId string, targetUserId string) (string, int) {
	db := dao.NewDbClient(ctx)
	if !checkAdmin(ctx, userId) {
		return "权限不足", -2
	}
	targetUser := model.UserInfo{}
	if err := db.Where("uuid = ?", targetUserId).First(&targetUser).Error; err != nil {
		return "查询目标用户失败", -1
	}

	if targetUser.Status == 0 {
		return "用户已被启用", -2
	} else {
		targetUser.Status = 0
		if err := db.Save(&targetUser).Error; err != nil {
			return "启用用户失败", -1
		}
	}
	return "用户启用成功", 0
}

func (s *adminService) DeleteUser(ctx context.Context, userId string, targetUserId string) (string, int) {
	db := dao.NewDbClient(ctx)
	if !checkAdmin(ctx, userId) {
		return "权限不足", -2
	}

	targetUser := model.UserInfo{}
	if err := db.Where("uuid = ?", targetUserId).First(&targetUser).Error; err != nil {
		return "查询目标用户失败", -1
	}

	if err := db.Delete(&targetUser).Error; err != nil {
		return "删除用户失败", -1
	}
	return "用户删除成功", 0
}

func (s *adminService) SetAdmin(ctx context.Context, userId string, targetUserId string) (string, int) {
	db := dao.NewDbClient(ctx)
	if !checkAdmin(ctx, userId) {
		return "权限不足", -2
	}

	targetUser := model.UserInfo{}
	if err := db.Where("uuid = ?", targetUserId).First(&targetUser).Error; err != nil {
		return "查询目标用户失败", -1
	}
	if targetUser.IsAdmin == 1 {
		return "用户已是管理员", -2
	} else {
		targetUser.IsAdmin = 1
		if err := db.Save(&targetUser).Error; err != nil {
			return "设置管理员失败", -1
		}
	}
	return "设置管理员成功", 0
}

func checkAdmin(ctx context.Context, userId string) bool {
	db := dao.NewDbClient(ctx)
	user := model.UserInfo{}
	if err := db.Where("uuid = ?", userId).First(&user).Error; err != nil {
		return false
	}
	return user.IsAdmin == 1
}
