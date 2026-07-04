package middleware

import (
	"fmt"
	"net/http"

	mysqlDAO "techmind/internal/dao/mysql"
	"techmind/internal/model"
	"techmind/internal/monitor"
	"techmind/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Recovery 捕获 panic，记录错误事件并返回统一响应。
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				msg := fmt.Sprintf("%v", r)
				zap.L().Error("panic recovered", zap.String("request_id", GetRequestID(c)), zap.String("panic", msg))
				recordErrorEvent(c, "panic", msg)
				c.AbortWithStatusJSON(http.StatusInternalServerError, response.Response{
					Code: response.CodeServerError,
					Msg:  response.CodeServerError.Msg(),
				})
			}
		}()
		c.Next()

		if c.Writer.Status() >= 500 {
			recordErrorEvent(c, "http", http.StatusText(c.Writer.Status()))
		}
	}
}

func recordErrorEvent(c *gin.Context, source, message string) {
	monitor.IncErrorEvent(source)
	event := &model.MonitorErrorEvent{
		Source:    source,
		Path:      c.FullPath(),
		RequestID: GetRequestID(c),
		Message:   message,
	}
	if err := mysqlDAO.CreateErrorEvent(event); err != nil {
		zap.L().Warn("record error event failed", zap.Error(err))
	}
}
