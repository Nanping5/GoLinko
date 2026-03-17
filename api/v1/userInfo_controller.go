package v1

import (
	"GoLinko/internal/dto/request"
	gormss "GoLinko/internal/service/gorms"
	constants "GoLinko/pkg/const"
	"GoLinko/pkg/zlog"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Register 用户注册
func Register(c *gin.Context) {
	registerReq := request.UserRegisterRequest{}
	if err := c.ShouldBindJSON(&registerReq); err != nil {
		zlog.GetLogger().Error("注册参数绑定失败", zap.Error(err))
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": constants.SYSTEM_ERROR,
		})
		return
	}
	msg, userInfo, ret := gormss.UserInfoService.Register(c.Request.Context(), registerReq)
	JsonBack(c, msg, ret, userInfo)
}

// Login 用户登录
func Login(c *gin.Context) {
	var req request.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		zlog.GetLogger().Error("登录参数绑定失败", zap.Error(err))
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": constants.SYSTEM_ERROR,
		})
		return
	}

	// 这里应该调用 service 层的登录逻辑，验证用户身份并生成 token
	msg, userInfo, ret := gormss.UserInfoService.Login(c.Request.Context(), req)
	JsonBack(c, msg, ret, userInfo)
}

// 验证码登录
func LoginByCode(c *gin.Context) {
	var req request.UserLoginByCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		zlog.GetLogger().Error("验证码登录参数绑定失败", zap.Error(err))
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": constants.SYSTEM_ERROR,
		})
		return
	}
	msg, userInfo, ret := gormss.UserInfoService.LoginByCode(c.Request.Context(), req)
	JsonBack(c, msg, ret, userInfo)
}

// SendEmailCode 发送验证码
func SendEmailCode(c *gin.Context) {

	req := request.SendEmailCodeRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		zlog.GetLogger().Error("发送验证码参数绑定失败", zap.Error(err))
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": constants.SYSTEM_ERROR,
		})
		return
	}
	msg, ret := gormss.UserInfoService.SendEmailCode(req)

	JsonBack(c, msg, ret, nil)
}

//更新用户信息

func UpdateUserInfo(c *gin.Context) {

	req := request.UserUpdateInfoRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		zlog.GetLogger().Error("更新用户信息参数绑定失败", zap.Error(err))
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": constants.SYSTEM_ERROR,
		})
		return
	}
	// 从 Token 获取userId
	req.Uuid = c.GetString("userId")
	msg, ret, resp := gormss.UserInfoService.UpdateUserInfo(c.Request.Context(), req)

	JsonBack(c, msg, ret, resp)
}

// GetUserInfo 获取用户信息，支持通过 ?user_id= 查询他人资料
func GetUserInfo(c *gin.Context) {
	currentUserId := c.GetString("userId")
	if currentUserId == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    401,
			"message": "未登录或登录异常",
		})
		return
	}

	// 如果传入了 user_id 参数则查询指定用户，否则查询当前登录用户
	targetId := c.Query("user_id")
	if targetId == "" {
		targetId = currentUserId
	}

	msg, resp, ret := gormss.UserInfoService.GetUserById(c.Request.Context(), targetId)
	JsonBack(c, msg, ret, resp)
}
