package service

import (
	"testing"
	"time"

	"github.com/ReadingGarden/back-go/internal/auth/entity"
	"github.com/ReadingGarden/back-go/internal/config"
)

func TestTokenServiceAccessTokenRoundTrip(t *testing.T) {
	t.Parallel()

	tokenService := NewTokenService(nil, config.AuthConfig{
		JWTSecretKey:    "secret",
		JWTAlgorithm:    "HS256",
		AccessTokenTTL:  24 * time.Hour,
		RefreshTokenTTL: 60 * 7 * 24 * time.Hour,
	})

	token, err := tokenService.GenerateAccessToken(entity.User{
		UserNo:   10,
		UserNick: "테스트유저",
	})
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	claims, err := tokenService.VerifyAccessToken(token)
	if err != nil {
		t.Fatalf("VerifyAccessToken() error = %v", err)
	}

	if claims.UserNo != 10 {
		t.Fatalf("claims.UserNo = %d, want 10", claims.UserNo)
	}
	if claims.UserNick != "테스트유저" {
		t.Fatalf("claims.UserNick = %q, want 테스트유저", claims.UserNick)
	}
	if claims.Type != 0 {
		t.Fatalf("claims.Type = %d, want 0", claims.Type)
	}
	if claims.Exp-claims.Iat != int64((24 * time.Hour).Seconds()) {
		t.Fatalf("ttl = %d, want %d", claims.Exp-claims.Iat, int64((24 * time.Hour).Seconds()))
	}
}

func TestTokenServiceVerifyAccessTokenExpired(t *testing.T) {
	t.Parallel()

	tokenService := NewTokenService(nil, config.AuthConfig{
		JWTSecretKey:    "secret",
		JWTAlgorithm:    "HS256",
		AccessTokenTTL:  -1 * time.Second,
		RefreshTokenTTL: 60 * 7 * 24 * time.Hour,
	})

	token, err := tokenService.GenerateAccessToken(entity.User{
		UserNo:   10,
		UserNick: "테스트유저",
	})
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	if _, err := tokenService.VerifyAccessToken(token); err == nil {
		t.Fatal("VerifyAccessToken() error = nil, want expired error")
	}
}
