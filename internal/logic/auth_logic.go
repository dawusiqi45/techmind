package logic

import (
	"errors"
	"fmt"

	mysqlDAO "techmind/internal/dao/mysql"
	"techmind/internal/model"
	"techmind/internal/pkg/jwt"
	"techmind/internal/pkg/snowflake"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExist      = errors.New("username already exists")
	ErrUserNotExist   = errors.New("user not found")
	ErrWrongPassword  = errors.New("wrong password")
)

// RegisterInput 注册请求参数
type RegisterInput struct {
	Username string
	Password string
	Email    string
}

// LoginInput 登录请求参数
type LoginInput struct {
	Username string
	Password string
}

// TokenPair access + refresh token 对
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Register 注册新用户
func Register(in *RegisterInput) error {
	exists, err := mysqlDAO.ExistsUsername(in.Username)
	if err != nil {
		return fmt.Errorf("register: check username: %w", err)
	}
	if exists {
		return ErrUserExist
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("register: hash password: %w", err)
	}

	u := &model.User{
		ID:       snowflake.GenID(),
		Username: in.Username,
		Password: string(hash),
		Email:    in.Email,
		Role:     0,
		Status:   1,
	}
	return mysqlDAO.CreateUser(u)
}

// Login 登录，返回 token 对
func Login(in *LoginInput) (*TokenPair, error) {
	u, err := mysqlDAO.GetUserByUsername(in.Username)
	if err != nil {
		return nil, fmt.Errorf("login: query user: %w", err)
	}
	if u == nil {
		return nil, ErrUserNotExist
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(in.Password)); err != nil {
		return nil, ErrWrongPassword
	}

	accessToken, err := jwt.GenAccessToken(u.ID)
	if err != nil {
		return nil, fmt.Errorf("login: gen access token: %w", err)
	}
	refreshToken, err := jwt.GenRefreshToken(u.ID)
	if err != nil {
		return nil, fmt.Errorf("login: gen refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// RefreshToken 用 refresh token 换新 access token
func RefreshToken(refreshTokenStr string) (string, error) {
	claims, err := jwt.ParseToken(refreshTokenStr)
	if err != nil {
		return "", err
	}
	return jwt.GenAccessToken(claims.UserID)
}

// GetProfile 获取当前用户 profile
func GetProfile(userID int64) (*model.UserProfile, error) {
	u, err := mysqlDAO.GetUserByID(userID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrUserNotExist
	}
	return &model.UserProfile{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		Avatar:    u.Avatar,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
	}, nil
}

type UpdateProfileInput struct {
	Username string
	Email    string
}

func UpdateProfile(userID int64, in *UpdateProfileInput) error {
	if in.Username != "" {
		exists, err := mysqlDAO.ExistsUsername(in.Username)
		if err != nil {
			return fmt.Errorf("update profile: check username: %w", err)
		}
		u, err := mysqlDAO.GetUserByID(userID)
		if err != nil {
			return fmt.Errorf("update profile: get user: %w", err)
		}
		if u != nil && u.Username != in.Username && exists {
			return ErrUserExist
		}
	}
	return mysqlDAO.UpdateUserProfile(userID, in.Username, in.Email)
}

func UpdateAvatar(userID int64, avatar string) error {
	return mysqlDAO.UpdateUserAvatar(userID, avatar)
}
