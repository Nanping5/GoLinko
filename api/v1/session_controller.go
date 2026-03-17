package v1

import (
	gormss "GoLinko/internal/service/gorms"

	"github.com/gin-gonic/gin"
)

// OpenSession 开启会话
func OpenSession(c *gin.Context) {
	send_id := c.GetString("userId")

	if send_id == "" {
		c.JSON(200, gin.H{
			"code":    400,
			"message": "用户未登录",
		})
		return
	}
	receive_id := c.Query("receive_id")
	if receive_id == "" {
		c.JSON(200, gin.H{
			"code":    400,
			"message": "接收用户ID不能为空",
		})
		return
	}
	msg, session_id, ret := gormss.SessionService.OpenSession(c.Request.Context(), send_id, receive_id)

	JsonBack(c, msg, ret, gin.H{
		"session_id": session_id,
	})
}

// GetSessionList 获取会话列表
func GetSessionList(c *gin.Context) {
	send_id := c.GetString("userId")

	if send_id == "" {
		c.JSON(200, gin.H{
			"code":    400,
			"message": "用户未登录",
		})
		return
	}

	msg, resp, ret := gormss.SessionService.GetSessionList(c.Request.Context(), send_id)

	JsonBack(c, msg, ret, resp)
}

// GetGroupSessionList 获取群会话列表
func GetGroupSessionList(c *gin.Context) {
	send_id := c.GetString("userId")

	if send_id == "" {
		c.JSON(200, gin.H{
			"code":    400,
			"message": "用户未登录",
		})
		return
	}

	msg, resp, ret := gormss.SessionService.GetGroupSessionList(c.Request.Context(), send_id)

	JsonBack(c, msg, ret, resp)
}

// DeleteSession 删除会话
func DeleteSession(c *gin.Context) {
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

	msg, ret := gormss.SessionService.DeleteSession(c.Request.Context(), send_id, session_id)

	JsonBack(c, msg, ret, nil)

}

// CheckOpenSessionAllowed 检查是否允许开启会话（双方是否是好友关系）
func CheckOpenSessionAllowed(c *gin.Context) {
	send_id := c.GetString("userId")

	if send_id == "" {
		c.JSON(200, gin.H{
			"code":    400,
			"message": "用户未登录",
		})
		return
	}
	receive_id := c.Query("receive_id")
	if receive_id == "" {
		c.JSON(200, gin.H{
			"code":    400,
			"message": "接收用户ID不能为空",
		})
		return
	}

	msg, resp, ret := gormss.SessionService.CheckOpenSessionAllowed(c.Request.Context(), send_id, receive_id)

	JsonBack(c, msg, ret, resp)
}
