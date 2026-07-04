package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const HeaderRequestID = "X-Request-ID"

// RequestID 中间件：若请求头已有 X-Request-ID 则复用，否则生成 UUID。
// 同时将 request_id 写入响应头和 gin.Context，方便日志和错误事件关联。
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader(HeaderRequestID)
		if rid == "" {
			rid = uuid.New().String()
		}
		c.Set("request_id", rid)
		c.Header(HeaderRequestID, rid)
		c.Next()
	}
}

// GetRequestID 从 Context 取 request_id
func GetRequestID(c *gin.Context) string {
	rid, _ := c.Get("request_id")
	if s, ok := rid.(string); ok {
		return s
	}
	return ""
}
