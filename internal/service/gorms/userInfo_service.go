package gormss

import (
	"GoLinko/internal/dao"
	"GoLinko/internal/dto/request"
	"GoLinko/internal/dto/response"
	"GoLinko/internal/model"
	myredis "GoLinko/internal/service/redis"
	"GoLinko/internal/service/sms"
	constants "GoLinko/pkg/const"
	"GoLinko/pkg/utils"
	"GoLinko/pkg/zlog"
	"context"
	"errors"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type userInfoService struct {
}

var UserInfoService = new(userInfoService)

// Register 用户注册
func (u *userInfoService) Register(ctx context.Context, registerReq request.UserRegisterRequest) (string, *response.UserRegisterResponse, int) {
	msg, isValid := u.CheckCodeIsValid(registerReq.Email, registerReq.Code)
	if !isValid {
		return msg, nil, -2
	}

	msg, exist := u.CheckEmailExist(registerReq.Email)
	if exist {
		zlog.GetLogger().Error(msg, zap.String("email", registerReq.Email))
		return msg, nil, -2
	}
	pwd := model.GeneratePassword(registerReq.Password)

	user := model.UserInfo{
		Uuid:      utils.GenerateUserID(),
		Nickname:  registerReq.Nickname,
		Telephone: registerReq.Telephone,
		Email:     registerReq.Email,
		Password:  pwd,
	}
	db := dao.NewDbClient(ctx)
	if err := db.Create(&user).Error; err != nil {
		zlog.GetLogger().Error("创建用户失败", zap.Error(err))
		return constants.SYSTEM_ERROR, nil, -1
	}
	registerResp := &response.UserRegisterResponse{
		Uuid:      user.Uuid,
		Email:     user.Email,
		Nickname:  user.Nickname,
		Telephone: user.Telephone,
		IsAdmin:   user.IsAdmin,
		Status:    user.Status,
		Avatar:    user.Avatar,
		CreatedAt: user.CreatedAt.Format("2006-01-02"),
	}
	return "注册成功", registerResp, 0
}

// 登录
func (u *userInfoService) Login(ctx context.Context, loginReq request.UserLoginRequest) (msg string, resp *response.UserLoginResponse, ret int) {
	password := loginReq.Password
	email := loginReq.Email
	user := model.UserInfo{}
	db := dao.NewDbClient(ctx)
	msg, exist := u.CheckEmailExist(email)
	if !exist {
		msg = "用户不存在"
		ret = -2
		zlog.GetLogger().Error(msg, zap.String("email", email))
		return msg, nil, ret
	}

	if err := db.Where("email=?", email).First(&user).Error; err != nil {
		zlog.GetLogger().Error("查询用户信息失败", zap.Error(err), zap.String("email", email))
		return constants.SYSTEM_ERROR, nil, -1
	}

	if !model.CheckPassword(user.Password, password) {
		msg = "账户或密码错误"
		ret = -2
		zlog.GetLogger().Error(msg, zap.String("email", email))
		return msg, nil, ret
	}
	token, err := utils.GenerateJwtToken(user.Uuid)
	if err != nil {
		zlog.GetLogger().Error("生成JWT令牌失败", zap.Error(err), zap.String("user_id", user.Uuid))
		return constants.SYSTEM_ERROR, nil, -1
	}
	loginResp := &response.UserLoginResponse{
		Uuid:      user.Uuid,
		Email:     user.Email,
		Nickname:  user.Nickname,
		Avatar:    user.Avatar,
		Telephone: user.Telephone,
		Gender:    user.Gender,
		Birthday:  user.Birthday,
		Signature: user.Signature,
		IsAdmin:   user.IsAdmin,
		Status:    user.Status,
		Token:     token,
	}
	msg = "登录成功"
	ret = 0
	return msg, loginResp, ret

}

// LoginByCode 验证码登录
func (u *userInfoService) LoginByCode(ctx context.Context, req request.UserLoginByCodeRequest) (msg string, resp *response.UserLoginResponse, ret int) {
	email := req.Email

	msg, isValid := u.CheckCodeIsValid(email, req.Code)
	if !isValid {
		return msg, nil, -2
	}
	user := model.UserInfo{}
	db := dao.NewDbClient(ctx)
	msg, exist := u.CheckEmailExist(email)
	if !exist {
		msg = "用户不存在"
		ret = -2
		zlog.GetLogger().Error(msg, zap.String("email", email))
		return msg, nil, ret
	}
	if err := db.Where("email=?", email).First(&user).Error; err != nil {
		zlog.GetLogger().Error("查询用户信息失败", zap.Error(err), zap.String("email", email))
		return constants.SYSTEM_ERROR, nil, -1
	}
	//分发token
	token, err := utils.GenerateJwtToken(user.Uuid)
	if err != nil {
		zlog.GetLogger().Error("生成JWT令牌失败", zap.Error(err), zap.String("user_id", user.Uuid))
		return constants.SYSTEM_ERROR, nil, -1
	}
	loginResp := &response.UserLoginResponse{
		Uuid:      user.Uuid,
		Email:     user.Email,
		Nickname:  user.Nickname,
		Avatar:    user.Avatar,
		Telephone: user.Telephone,
		Gender:    user.Gender,
		Birthday:  user.Birthday,
		Signature: user.Signature,
		IsAdmin:   user.IsAdmin,
		Status:    user.Status,
		Token:     token,
	}
	msg = "登录成功"
	ret = 0
	return msg, loginResp, ret
}

// SendEmailCode 发送验证码
func (u *userInfoService) SendEmailCode(req request.SendEmailCodeRequest) (msg string, ret int) {
	email := req.Email
	// 这里应该调用邮件服务发送验证码
	return sms.VerifyEmailCode(email)
}

// 检查账户是否存在
func (u *userInfoService) CheckEmailExist(email string) (string, bool) {
	user := model.UserInfo{}
	db := dao.GetDB()
	res := db.First(&user, "email = ?", email)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return "用户不存在", false
		}
		zlog.GetLogger().Error("查询用户信息失败", zap.Error(res.Error))
		return "查询用户信息失败", false // 查询出错不应视为邮箱已存在
	}
	return "用户已存在", true
}

// CheckCodeIsValid 验证验证码是否有效
func (u *userInfoService) CheckCodeIsValid(email, code string) (string, bool) {
	key := "email_code:" + email
	cachedCode, err := myredis.GetKey(key)
	if err != nil {
		zlog.GetLogger().Error("获取验证码失败", zap.Error(err))
		return constants.SYSTEM_ERROR, false
	}
	if cachedCode != code {
		msg := "验证码错误或已过期"
		zlog.GetLogger().Error(msg, zap.String("email", email))
		return msg, false
	} else {
		// 验证码正确，删除验证码缓存
		if err := myredis.DelKeyIfExist(key); err != nil {
			zlog.GetLogger().Error("", zap.Error(err))
			return constants.SYSTEM_ERROR, false
		}
	}

	return "验证码有效", true
}

// GetUserById 根据用户id获取用户信息
func (u *userInfoService) GetUserById(ctx context.Context, uuid string) (string, *response.GetUserByIdResponse, int) {
	db := dao.NewDbClient(ctx)
	user := model.UserInfo{}
	if err := db.Where("uuid=?", uuid).First(&user).Error; err != nil {
		zlog.GetLogger().Error("查询用户信息失败", zap.Error(err), zap.String("uuid", uuid))
		return constants.SYSTEM_ERROR, nil, -1
	}
	resp := &response.GetUserByIdResponse{
		Uuid:      user.Uuid,
		Email:     user.Email,
		Nickname:  user.Nickname,
		Telephone: user.Telephone,
		Gender:    user.Gender,
		Avatar:    user.Avatar,
		Birthday:  user.Birthday,
		Signature: user.Signature,
		IsAdmin:   user.IsAdmin,
		Status:    user.Status,
		CreatedAt: user.CreatedAt.Format("2006-01-02-15:04:05"),
	}
	return "查询用户信息成功", resp, 0
}

// UpdateUserInfo 更新用户信息（不要求更新全部）
func (u *userInfoService) UpdateUserInfo(ctx context.Context, req request.UserUpdateInfoRequest) (string, int, *response.UserUpdateUserInfoResponse) {
	db := dao.NewDbClient(ctx)

	// 从数据库查询用户模型对象
	user := model.UserInfo{}
	if err := db.Where("uuid=?", req.Uuid).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			zlog.GetLogger().Error("用户不存在", zap.String("uuid", req.Uuid))
			return "用户不存在", -2, nil
		}
		zlog.GetLogger().Error("查询用户信息失败", zap.Error(err), zap.String("uuid", req.Uuid))
		return constants.SYSTEM_ERROR, -1, nil
	}

	// 更新用户信息（只更新提供的字段）
	if req.Nickname != nil {
		user.Nickname = *req.Nickname
	}
	if req.Avatar != nil {
		user.Avatar = *req.Avatar
	}
	if req.Birthday != nil {
		user.Birthday = *req.Birthday
	}
	if req.Gender != nil {
		user.Gender = *req.Gender
	}
	if req.Signature != nil {
		user.Signature = *req.Signature
	}
	if req.Telephone != nil {
		user.Telephone = *req.Telephone
	}

	// 保存更新
	if err := db.Save(&user).Error; err != nil {
		zlog.GetLogger().Error("更新用户信息失败", zap.Error(err), zap.String("uuid", req.Uuid))
		return constants.SYSTEM_ERROR, -1, nil
	}
	// 返回更新后的实际值
	resp := &response.UserUpdateUserInfoResponse{
		Uuid:      user.Uuid,
		Nickname:  &user.Nickname,
		Avatar:    &user.Avatar,
		Birthday:  &user.Birthday,
		Gender:    &user.Gender,
		Signature: &user.Signature,
		Telephone: &user.Telephone,
	}
	return "更新用户信息成功", 0, resp
}
