package middleware

import (
	"time"

	mysqlDAO "techmind/internal/dao/mysql"
	"techmind/internal/model"
	"techmind/internal/monitor"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const defaultSlowRequestThreshold = 800 * time.Millisecond

// SlowRequest 记录超过阈值的 HTTP 请求。
func SlowRequest(threshold time.Duration) gin.HandlerFunc {
	if threshold <= 0 {
		threshold = defaultSlowRequestThreshold
	}
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		duration := time.Since(start)
		if duration < threshold {
			return
		}

		uid, _ := GetCurrentUserID(c)
		req := &model.MonitorSlowRequest{
			RequestID:  GetRequestID(c),
			Method:     c.Request.Method,
			Path:       c.FullPath(),
			StatusCode: c.Writer.Status(),
			DurationMs: int(duration.Milliseconds()),
			UserID:     uid,
		}
		monitor.IncSlowRequest(req.Method, req.Path)
		if err := mysqlDAO.CreateSlowRequest(req); err != nil {
			zap.L().Warn("record slow request failed", zap.Error(err))
		}
	}
}
