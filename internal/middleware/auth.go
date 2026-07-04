package middleware

import (
	"errors"

	"techmind/internal/pkg/jwt"
	"techmind/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

const ctxKeyUserID = "user_id"

// JWT 中间件：校验 Authorization: Bearer <token>，将 user_id 写入 Context
func JWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := c.GetHeader("Authorization")
		if len(tokenStr) > 7 && tokenStr[:7] == "Bearer " {
			tokenStr = tokenStr[7:]
		} else {
			response.AbortWithUnauthorized(c)
			return
		}

		claims, err := jwt.ParseToken(tokenStr)
		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) {
				response.AbortWithUnauthorized(c)
				return
			}
			response.AbortWithUnauthorized(c)
			return
		}

		c.Set(ctxKeyUserID, claims.UserID)
		c.Next()
	}
}

// GetCurrentUserID 从 Context 获取当前登录用户 ID
func GetCurrentUserID(c *gin.Context) (int64, bool) {
	val, exists := c.Get(ctxKeyUserID)
	if !exists {
		return 0, false
	}
	uid, ok := val.(int64)
	return uid, ok
}
