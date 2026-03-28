package v1

import (
	"GoLinko/internal/service/chat"

	"github.com/gin-gonic/gin"
)

func WsLogin(c *gin.Context) {
	// 从 JWT 中间件解析出的 userId 作为 WebSocket 标识，防止客户端伪造他人 ID
	clientId := c.GetString("userId")
	if clientId == "" {
		c.JSON(200, gin.H{
			"code":    401,
			"message": "用户未登录",
		})
		return
	}
	chat.NewClientInit(c, clientId)
}

func WsLogout(c *gin.Context) {
	user_id := c.GetString("userId")
	if user_id == "" {
		c.JSON(200, gin.H{
			"code":    400,
			"message": "用户未登录",
		})
		return
	}
	msg, ret := chat.WsLogout(user_id)
	JsonBack(c, msg, ret, nil)
}
