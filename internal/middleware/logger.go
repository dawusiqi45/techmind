package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logger 中间件：记录每个请求的方法、路径、状态码、耗时和 request_id
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		rid := GetRequestID(c)

		fields := []zap.Field{
			zap.String("rid", rid),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("ip", c.ClientIP()),
		}

		if status >= 500 {
			zap.L().Error("request", fields...)
		} else if status >= 400 {
			zap.L().Warn("request", fields...)
		} else {
			zap.L().Info("request", fields...)
		}
	}
}
