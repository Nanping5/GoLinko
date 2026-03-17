package v1

import (
	gormss "GoLinko/internal/service/gorms"

	"github.com/gin-gonic/gin"
)

func GetUserList(c *gin.Context) {
	userId := c.GetString("userId")
	if userId == "" {
		c.JSON(200, gin.H{
			"code":    400,
			"message": "用户未登录",
		})
		return
	}
	msg, userList, ret := gormss.AdminService.GetUserList(c.Request.Context(), userId)
	JsonBack(c, msg, ret, userList)
}

func AbleUser(c *gin.Context) {
	userId := c.GetString("userId")
	if userId == "" {
		c.JSON(200, gin.H{
			"code":    400,
			"message": "用户未登录",
		})
		return
	}
	targetUserId := c.Query("target_user_id")
	if targetUserId == "" {
		c.JSON(200, gin.H{
			"code":    400,
			"message": "目标用户ID不能为空",
		})
		return
	}
	msg, ret := gormss.AdminService.AbleUser(c.Request.Context(), userId, targetUserId)
	JsonBack(c, msg, ret, nil)
}

func DisableUser(c *gin.Context) {
	userId := c.GetString("userId")
	if userId == "" {
		c.JSON(200, gin.H{
			"code":    400,
			"message": "用户未登录",
		})
		return
	}
	targetUserId := c.Query("target_user_id")
	if targetUserId == "" {
		c.JSON(200, gin.H{
			"code":    400,
			"message": "目标用户ID不能为空",
		})
		return
	}
	msg, ret := gormss.AdminService.DisableUser(c.Request.Context(), userId, targetUserId)
	JsonBack(c, msg, ret, nil)
}

func DeleteUser(c *gin.Context) {
	userId := c.GetString("userId")
	if userId == "" {
		c.JSON(200, gin.H{
			"code":    400,
			"message": "用户未登录",
		})
		return
	}
	targetUserId := c.Query("target_user_id")
	if targetUserId == "" {
		c.JSON(200, gin.H{
			"code":    400,
			"message": "目标用户ID不能为空",
		})
		return
	}
	msg, ret := gormss.AdminService.DeleteUser(c.Request.Context(), userId, targetUserId)
	JsonBack(c, msg, ret, nil)
}

func SetAdmin(c *gin.Context) {
	userId := c.GetString("userId")
	if userId == "" {
		c.JSON(200, gin.H{
			"code":    400,
			"message": "用户未登录",
		})
		return
	}
	targetUserId := c.Query("target_user_id")
	if targetUserId == "" {
		c.JSON(200, gin.H{
			"code":    400,
			"message": "目标用户ID不能为空",
		})
		return
	}
	msg, ret := gormss.AdminService.SetAdmin(c.Request.Context(), userId, targetUserId)
	JsonBack(c, msg, ret, nil)
}
