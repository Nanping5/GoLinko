package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// JsonBack 统一的 JSON 响应格式
// ret 参数说明：
// 0：成功，返回数据
// -1：系统错误，如序列化失败，数据库错误等
// -2：业务未成功，如登录失败，参数错误等
func JsonBack(c *gin.Context, message string, ret int, data any) {
	if ret == 0 {
		if data != nil {
			c.JSON(http.StatusOK, gin.H{
				"code":    200,
				"message": message,
				"data":    data,
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"code":    200,
				"message": message,
			})
		}
	} else if ret == -2 {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": message,
		})
	} else if ret == -1 {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": message,
		})
	}
}
