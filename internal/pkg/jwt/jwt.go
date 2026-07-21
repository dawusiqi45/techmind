package jwt

import (
	"errors"
	"fmt"
	"time"

	"techmind/internal/pkg/settings"

	gojwt "github.com/golang-jwt/jwt/v5"
)

// tokenType 区分 access token 和 refresh token，防止互换使用
type tokenType string

const (
	tokenTypeAccess  tokenType = "access"
	tokenTypeRefresh tokenType = "refresh"
)

// Claims 是 JWT payload
type Claims struct {
	UserID    int64     `json:"uid"`
	TokenType tokenType `json:"type"`
	gojwt.RegisteredClaims
}

var (
	ErrTokenExpired = errors.New("token expired")
	ErrTokenInvalid = errors.New("token invalid")
)

var cfg *settings.JWTSetting

// Init 保存 JWT 配置，应在 main 中调用一次
func Init(c *settings.JWTSetting) {
	cfg = c
}

// GenAccessToken 生成短期 access token
func GenAccessToken(userID int64) (string, error) {
	return genToken(userID, tokenTypeAccess, time.Duration(cfg.AccessExpireMin)*time.Minute)
}

// GenRefreshToken 生成长期 refresh token
func GenRefreshToken(userID int64) (string, error) {
	return genToken(userID, tokenTypeRefresh, time.Duration(cfg.RefreshExpireH)*time.Hour)
}

// ParseAccessToken 解析并验证 access token。
func ParseAccessToken(tokenStr string) (*Claims, error) {
	return parseToken(tokenStr, tokenTypeAccess)
}

// ParseRefreshToken 解析并验证 refresh token。
func ParseRefreshToken(tokenStr string) (*Claims, error) {
	return parseToken(tokenStr, tokenTypeRefresh)
}

func parseToken(tokenStr string, expectedType tokenType) (*Claims, error) {
	claims := &Claims{}
	token, err := gojwt.ParseWithClaims(tokenStr, claims, func(t *gojwt.Token) (interface{}, error) {
		if t.Method != gojwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(cfg.Secret), nil
	}, gojwt.WithValidMethods([]string{gojwt.SigningMethodHS256.Alg()}), gojwt.WithIssuer("techmind"))

	if err != nil {
		if errors.Is(err, gojwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}
	if !token.Valid || claims.UserID <= 0 || claims.TokenType != expectedType {
		return nil, ErrTokenInvalid
	}
	return claims, nil
}

func genToken(userID int64, tt tokenType, expire time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:    userID,
		TokenType: tt,
		RegisteredClaims: gojwt.RegisteredClaims{
			IssuedAt:  gojwt.NewNumericDate(now),
			ExpiresAt: gojwt.NewNumericDate(now.Add(expire)),
			Issuer:    "techmind",
		},
	}
	token := gojwt.NewWithClaims(gojwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}
