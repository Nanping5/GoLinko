package gormss

import (
	"GoLinko/internal/dao"
	"GoLinko/internal/dto/response"
	"GoLinko/internal/model"
	"context"
	"time"

	"gorm.io/gorm"
)

type adminService struct {
}

var AdminService = &adminService{}

func (s *adminService) GetUserList(ctx context.Context, userId string) (string, []response.GetUserListResponse, int) {
	var users []model.UserInfo

	// 检查管理员权限（读操作，使用一致性读）
	db := dao.GetDBForRead(ctx)
	user := model.UserInfo{}
	if err := db.Where("uuid = ?", userId).First(&user).Error; err != nil {
		return "权限检查失败", nil, -2
	}
	if user.IsAdmin != 1 {
		return "权限不足", nil, -2
	}

	// 查询用户列表（纯读操作）
	db = dao.GetDBForRead(ctx)
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
	// 权限检查
	if !checkAdmin(ctx, userId) {
		return "权限不足", -2
	}

	// 写操作：禁用用户
	var err error
	ctx, err = dao.WriteAndMark(ctx, func(db *gorm.DB) error {
		targetUser := model.UserInfo{}
		if err := db.Where("uuid = ?", targetUserId).First(&targetUser).Error; err != nil {
			return err
		}
		if targetUser.Status == 1 {
			return nil // 已是禁用状态
		}
		targetUser.Status = 1
		return db.Save(&targetUser).Error
	})

	if err != nil {
		return "禁用用户失败", -1
	}
	return "用户禁用成功", 0
}

func (s *adminService) AbleUser(ctx context.Context, userId string, targetUserId string) (string, int) {
	// 权限检查
	if !checkAdmin(ctx, userId) {
		return "权限不足", -2
	}

	// 写操作：启用用户
	var err error
	ctx, err = dao.WriteAndMark(ctx, func(db *gorm.DB) error {
		targetUser := model.UserInfo{}
		if err := db.Where("uuid = ?", targetUserId).First(&targetUser).Error; err != nil {
			return err
		}
		if targetUser.Status == 0 {
			return nil // 已是启用状态
		}
		targetUser.Status = 0
		return db.Save(&targetUser).Error
	})

	if err != nil {
		return "启用用户失败", -1
	}
	return "用户启用成功", 0
}

func (s *adminService) DeleteUser(ctx context.Context, userId string, targetUserId string) (string, int) {
	// 权限检查
	if !checkAdmin(ctx, userId) {
		return "权限不足", -2
	}

	// 写操作：删除用户
	var err error
	ctx, err = dao.WriteAndMark(ctx, func(db *gorm.DB) error {
		targetUser := model.UserInfo{}
		if err := db.Where("uuid = ?", targetUserId).First(&targetUser).Error; err != nil {
			return err
		}
		return db.Delete(&targetUser).Error
	})

	if err != nil {
		return "删除用户失败", -1
	}
	return "用户删除成功", 0
}

func (s *adminService) SetAdmin(ctx context.Context, userId string, targetUserId string) (string, int) {
	// 权限检查
	if !checkAdmin(ctx, userId) {
		return "权限不足", -2
	}

	// 写操作：设置管理员
	var err error
	ctx, err = dao.WriteAndMark(ctx, func(db *gorm.DB) error {
		targetUser := model.UserInfo{}
		if err := db.Where("uuid = ?", targetUserId).First(&targetUser).Error; err != nil {
			return err
		}
		if targetUser.IsAdmin == 1 {
			return nil // 已是管理员
		}
		targetUser.IsAdmin = 1
		return db.Save(&targetUser).Error
	})

	if err != nil {
		return "设置管理员失败", -1
	}
	return "设置管理员成功", 0
}

// checkAdmin 检查用户是否为管理员（纯读操作）
func checkAdmin(ctx context.Context, userId string) bool {
	db := dao.GetDBForRead(ctx)
	user := model.UserInfo{}
	if err := db.Where("uuid = ?", userId).First(&user).Error; err != nil {
		return false
	}
	return user.IsAdmin == 1
}
