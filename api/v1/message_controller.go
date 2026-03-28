package v1

import (
	gormss "GoLinko/internal/service/gorms"

	"github.com/gin-gonic/gin"
)

// GetMessageList 获取消息列表
func GetMessageList(c *gin.Context) {
	send_id := c.GetString("userId")

	if send_id == "" {
		c.JSON(200, gin.H{
			"code":    400,
			"message": "用户未登录",
		})
		return
	}
	session_id := c.Query("session_id")
	if session_id == "" {
		c.JSON(200, gin.H{
			"code":    400,
			"message": "会话ID不能为空",
		})
		return
	}
	msg, resp, ret := gormss.MessageService.GetMessageList(c.Request.Context(), send_id, session_id)

	JsonBack(c, msg, ret, resp)
}

// GetGroupMessageList 获取群消息列表
func GetGroupMessageList(c *gin.Context) {
	send_id := c.GetString("userId")

	if send_id == "" {
		c.JSON(200, gin.H{
			"code":    400,
			"message": "用户未登录",
		})
		return
	}
	group_id := c.Query("group_id")
	if group_id == "" {
		c.JSON(200, gin.H{
			"code":    400,
			"message": "群ID不能为空",
		})
		return
	}
	msg, resp, ret := gormss.MessageService.GetGroupMessageList(c.Request.Context(), send_id, group_id)

	JsonBack(c, msg, ret, resp)
}

// UploadAvatar 上传头像
func UploadAvatar(c *gin.Context) {
	msg, ret, avatarUrl := gormss.MessageService.UploadAvatar(c)
	var data interface{}
	if avatarUrl != "" {
		data = map[string]string{"url": avatarUrl}
	}
	JsonBack(c, msg, ret, data)
}

// UploadFile 上传文件
func UploadFile(c *gin.Context) {
	msg, ret, data := gormss.MessageService.UploadFile(c)
	JsonBack(c, msg, ret, data)
}
