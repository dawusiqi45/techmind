package middleware

import (
	"time"

	"techmind/internal/monitor"

	"github.com/gin-gonic/gin"
)

// Metrics 记录 HTTP 请求数量、状态码和耗时。
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		monitor.ObserveHTTPRequest(c.Request.Method, c.FullPath(), c.Writer.Status(), time.Since(start))
	}
}
