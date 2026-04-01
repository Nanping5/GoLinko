package middleware

import (
	"GoLinko/internal/dao"
	"GoLinko/pkg/zlog"
	"context"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RequestConsistency 请求级读写一致性中间件
func RequestConsistency() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		c.Set("consistencyCtx", ctx)

		zlog.GetLogger().Debug("请求级一致性中间件已启用",
			zap.String("path", c.Request.URL.Path))

		c.Next()
	}
}

// GetConsistencyCtx 从 gin.Context 获取上下文
// 供 service 层使用
func GetConsistencyCtx(c *gin.Context) context.Context {
	if ctx, exists := c.Get("consistencyCtx"); exists {
		if consistencyCtx, ok := ctx.(context.Context); ok {
			return consistencyCtx
		}
	}
	// 回退到原始请求上下文
	return c.Request.Context()
}

// UpdateConsistencyCtx 更新一致性上下文（写操作后调用）
func UpdateConsistencyCtx(c *gin.Context, ctx context.Context) {
	c.Set("consistencyCtx", ctx)
}

// MarkRequestWritten 标记当前请求已发生写操作
// 供 controller 层在写操作后调用
func MarkRequestWritten(c *gin.Context) {
	ctx := GetConsistencyCtx(c)
	newCtx := dao.MarkWritten(ctx)
	UpdateConsistencyCtx(c, newCtx)

	zlog.GetLogger().Debug("请求已标记为写入状态",
		zap.String("path", c.Request.URL.Path))
}
