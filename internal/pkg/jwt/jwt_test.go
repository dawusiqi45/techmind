package jwt

import (
	"errors"
	"testing"

	"techmind/internal/pkg/settings"
)

func TestTokenTypesAreNotInterchangeable(t *testing.T) {
	Init(&settings.JWTSetting{Secret: "test-secret-at-least-thirty-two-characters", AccessExpireMin: 5, RefreshExpireH: 1})

	access, err := GenAccessToken(42)
	if err != nil {
		t.Fatal(err)
	}
	refresh, err := GenRefreshToken(42)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := ParseAccessToken(access); err != nil {
		t.Fatalf("parse access token: %v", err)
	}
	if _, err := ParseRefreshToken(refresh); err != nil {
		t.Fatalf("parse refresh token: %v", err)
	}
	if _, err := ParseAccessToken(refresh); !errors.Is(err, ErrTokenInvalid) {
		t.Fatalf("refresh token accepted as access token: %v", err)
	}
	if _, err := ParseRefreshToken(access); !errors.Is(err, ErrTokenInvalid) {
		t.Fatalf("access token accepted as refresh token: %v", err)
	}
}
