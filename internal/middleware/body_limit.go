package middleware

import (
	"net/http"

	"techmind/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// BodyLimit 在任何绑定或 multipart 解析前限制请求体大小。
func BodyLimit(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxBytes {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, response.Response{
				Code: response.CodeInvalidParam,
				Msg:  "request body too large",
			})
			return
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		c.Next()
	}
}
