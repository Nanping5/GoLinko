package middleware

import (
	"GoLinko/pkg/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

func JWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		token = strings.TrimPrefix(token, "Bearer ")
		// 如果 Header 中没有，尝试从 Query 参数中获取 (用于 WebSocket)
		if token == "" {
			token = c.Query("token")
		}
		if token == "" {
			c.JSON(200, gin.H{
				"code":    401,
				"message": "缺少token",
			})
			c.Abort()
			return
		}
		claims, err := utils.ParseJwtToken(token)
		if err != nil {
			c.JSON(200, gin.H{
				"code":    401,
				"message": "无效的token",
			})
			c.Abort()
			return
		}
		c.Set("userId", claims.UserId)
		c.Next()
	}

}
