package controller

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"techmind/internal/logic"
	"techmind/internal/middleware"
	"techmind/internal/pkg/response"
	"techmind/internal/pkg/snowflake"

	"github.com/gin-gonic/gin"
)

// RegisterReq 注册请求体
type RegisterReq struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required,min=6,max=72"`
	Email    string `json:"email"    binding:"required,email"`
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
		if errors.Is(err, logic.ErrUserNotExist) || errors.Is(err, logic.ErrWrongPassword) || errors.Is(err, logic.ErrUserDisabled) {
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

type UpdateProfileReq struct {
	Username string `json:"username" binding:"omitempty,min=3,max=32"`
	Email    string `json:"email"    binding:"omitempty,email"`
}

func UpdateProfile(c *gin.Context) {
	uid, ok := middleware.GetCurrentUserID(c)
	if !ok {
		response.Fail(c, response.CodeUnauthorized)
		return
	}

	var req UpdateProfileReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(c, response.CodeInvalidParam, err.Error())
		return
	}

	err := logic.UpdateProfile(uid, &logic.UpdateProfileInput{
		Username: req.Username,
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

	profile, _ := logic.GetProfile(uid)
	response.OK(c, profile)
}

func UploadAvatar(c *gin.Context) {
	uid, ok := middleware.GetCurrentUserID(c)
	if !ok {
		response.Fail(c, response.CodeUnauthorized)
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 3*1024*1024)
	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		response.FailWithMsg(c, response.CodeInvalidParam, "请选择头像文件")
		return
	}
	defer file.Close()

	if header.Size > 2*1024*1024 {
		response.FailWithMsg(c, response.CodeInvalidParam, "头像文件不能超过 2MB")
		return
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
		response.FailWithMsg(c, response.CodeInvalidParam, "仅支持 jpg/png/webp 格式")
		return
	}

	headerBytes := make([]byte, 512)
	n, err := io.ReadFull(file, headerBytes)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		response.FailWithMsg(c, response.CodeInvalidParam, "无法读取头像文件")
		return
	}
	headerBytes = headerBytes[:n]
	contentType := http.DetectContentType(headerBytes)
	if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/webp" {
		response.FailWithMsg(c, response.CodeInvalidParam, "头像内容不是有效的 jpg/png/webp 图片")
		return
	}

	filename := fmt.Sprintf("%d%s", snowflake.GenID(), ext)
	savePath := filepath.Join("uploads", "avatars", filename)

	if err := os.MkdirAll(filepath.Dir(savePath), 0755); err != nil {
		response.Fail(c, response.CodeServerError)
		return
	}

	dst, err := os.Create(savePath)
	if err != nil {
		response.Fail(c, response.CodeServerError)
		return
	}
	defer dst.Close()

	if _, err := dst.Write(headerBytes); err != nil {
		_ = os.Remove(savePath)
		response.Fail(c, response.CodeServerError)
		return
	}
	if _, err := io.Copy(dst, file); err != nil {
		_ = os.Remove(savePath)
		response.Fail(c, response.CodeServerError)
		return
	}

	avatarURL := "/uploads/avatars/" + filename
	if err := logic.UpdateAvatar(uid, avatarURL); err != nil {
		_ = os.Remove(savePath)
		response.Fail(c, response.CodeServerError)
		return
	}

	response.OK(c, gin.H{"avatar": avatarURL})
}

func ListUserFavorites(c *gin.Context) {
	uid, ok := middleware.GetCurrentUserID(c)
	if !ok {
		response.Fail(c, response.CodeUnauthorized)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	list, total, err := logic.ListUserFavorites(uid, page, pageSize)
	if err != nil {
		response.Fail(c, response.CodeServerError)
		return
	}
	response.OK(c, gin.H{"list": list, "total": total})
}

func ListUserArticles(c *gin.Context) {
	uid, ok := middleware.GetCurrentUserID(c)
	if !ok {
		response.Fail(c, response.CodeUnauthorized)
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	list, total, err := logic.ListUserArticles(uid, page, pageSize)
	if err != nil {
		response.Fail(c, response.CodeServerError)
		return
	}
	response.OK(c, gin.H{"list": list, "total": total})
}

func ListUserLikes(c *gin.Context) {
	uid, ok := middleware.GetCurrentUserID(c)
	if !ok {
		response.Fail(c, response.CodeUnauthorized)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	list, total, err := logic.ListUserLikes(uid, page, pageSize)
	if err != nil {
		response.Fail(c, response.CodeServerError)
		return
	}
	response.OK(c, gin.H{"list": list, "total": total})
}
