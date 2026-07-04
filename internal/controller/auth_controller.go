package controller

import (
	"errors"

	"techmind/internal/logic"
	"techmind/internal/middleware"
	"techmind/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// RegisterReq 注册请求体
type RegisterReq struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required,min=6,max=72"`
	Email    string `json:"email"    binding:"omitempty,email"`
}

// LoginReq 登录请求体
type LoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RefreshReq 刷新 token 请求体
type RefreshReq struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Register POST /api/v1/auth/register
func Register(c *gin.Context) {
	var req RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	err := logic.Register(&logic.RegisterInput{
		Username: req.Username,
		Password: req.Password,
		Email:    req.Email,
	})
	if err != nil {
		if errors.Is(err, logic.ErrUserExist) {
			response.Fail(c, response.CodeUserExist)
			return
		}
		response.Fail(c, response.CodeServerError)
		return
	}
	response.OK(c, nil)
}

// Login POST /api/v1/auth/login
func Login(c *gin.Context) {
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	pair, err := logic.Login(&logic.LoginInput{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, logic.ErrUserNotExist) {
			response.Fail(c, response.CodeUserNotExist)
			return
		}
		if errors.Is(err, logic.ErrWrongPassword) {
			response.Fail(c, response.CodeWrongPassword)
			return
		}
		response.Fail(c, response.CodeServerError)
		return
	}
	response.OK(c, pair)
}

// RefreshToken POST /api/v1/auth/refresh
func RefreshToken(c *gin.Context) {
	var req RefreshReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	accessToken, err := logic.RefreshToken(req.RefreshToken)
	if err != nil {
		response.Fail(c, response.CodeUnauthorized)
		return
	}
	response.OK(c, gin.H{"access_token": accessToken})
}

// GetProfile GET /api/v1/user/profile
func GetProfile(c *gin.Context) {
	uid, ok := middleware.GetCurrentUserID(c)
	if !ok {
		response.Fail(c, response.CodeUnauthorized)
		return
	}

	profile, err := logic.GetProfile(uid)
	if err != nil {
		response.Fail(c, response.CodeServerError)
		return
	}
	response.OK(c, profile)
}
