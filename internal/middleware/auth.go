package middleware

import (
	"errors"

	mysqlDAO "techmind/internal/dao/mysql"
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

		claims, err := jwt.ParseAccessToken(tokenStr)
		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) {
				response.AbortWithUnauthorized(c)
				return
			}
			response.AbortWithUnauthorized(c)
			return
		}

		user, err := mysqlDAO.GetUserByID(claims.UserID)
		if err != nil || user == nil || user.Status != 1 {
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

// RequireAdmin 在服务端校验当前用户的管理员角色。前端路由守卫仅用于体验，不能作为安全边界。
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, ok := GetCurrentUserID(c)
		if !ok {
			response.AbortWithUnauthorized(c)
			return
		}

		user, err := mysqlDAO.GetUserByID(uid)
		if err != nil || user == nil || user.Role != 1 || user.Status != 1 {
			c.AbortWithStatusJSON(403, response.Response{
				Code: response.CodeForbidden,
				Msg:  response.CodeForbidden.Msg(),
			})
			return
		}
		c.Next()
	}
}
