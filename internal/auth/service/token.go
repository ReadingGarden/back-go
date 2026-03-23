package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ReadingGarden/back-go/internal/auth/entity"
	"github.com/ReadingGarden/back-go/internal/auth/repository"
	"github.com/ReadingGarden/back-go/internal/config"
)

var (
	ErrExpiredSignature = errors.New("Signature has expired")
	ErrInvalidToken     = errors.New("")
	ErrDecodeError      = errors.New("Not enough segments")
)

type TokenService struct {
	repo   *repository.MySQLRepository
	config config.AuthConfig
}

type tokenClaims struct {
	UserNo    int64  `json:"user_no"`
	UserNick  string `json:"user_nick"`
	Type      int    `json:"type"`
	Timestamp string `json:"timestamp"`
	Exp       int64  `json:"exp"`
	Iat       int64  `json:"iat"`
	Nbf       int64  `json:"nbf"`
}

type tokenHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

func NewTokenService(repo *repository.MySQLRepository, cfg config.AuthConfig) *TokenService {
	return &TokenService{
		repo:   repo,
		config: cfg,
	}
}

func (s *TokenService) GeneratePairToken(ctx context.Context, user entity.User) (entity.TokenPair, error) {
	accessToken, err := s.GenerateAccessToken(user)
	if err != nil {
		return entity.TokenPair{}, err
	}

	refreshToken, exp, err := s.generateRefreshToken(user)
	if err != nil {
		return entity.TokenPair{}, err
	}

	if err := s.repo.ReplaceRefreshToken(ctx, user.UserNo, refreshToken, exp); err != nil {
		return entity.TokenPair{}, err
	}

	return entity.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *TokenService) GenerateAccessToken(user entity.User) (string, error) {
	token, _, err := s.generateToken(user, 0, s.config.AccessTokenTTL)
	return token, err
}

func (s *TokenService) VerifyAccessToken(token string) (tokenClaims, error) {
	claims, err := s.decodeToken(token)
	if err != nil {
		return tokenClaims{}, err
	}
	if claims.Type != 0 {
		return tokenClaims{}, ErrInvalidToken
	}
	if claims.Exp < time.Now().UTC().Unix() {
		return tokenClaims{}, ErrExpiredSignature
	}

	return claims, nil
}

func (s *TokenService) Refresh(ctx context.Context, payload entity.RefreshTokenPayload) (entity.DataResp, int, error) {
	claims, err := s.decodeToken(payload.RefreshToken)
	if err != nil {
		return entity.DataResp{}, 0, err
	}
	if claims.Type != 1 {
		return entity.DataResp{}, 0, ErrInvalidToken
	}

	refreshToken, err := s.repo.FindRefreshToken(ctx, claims.UserNo, payload.RefreshToken)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return entity.DataResp{}, 0, ErrInvalidToken
		}
		return entity.DataResp{}, 0, err
	}

	if refreshToken.Exp.Before(time.Now().UTC()) {
		if deleteErr := s.repo.DeleteRefreshTokenByID(ctx, refreshToken.ID); deleteErr != nil {
			return entity.DataResp{}, 0, deleteErr
		}
		return entity.DataResp{}, 0, ErrExpiredSignature
	}

	user, err := s.repo.FindUserByNo(ctx, claims.UserNo)
	if err != nil {
		return entity.DataResp{}, 0, err
	}

	accessToken, err := s.GenerateAccessToken(user)
	if err != nil {
		return entity.DataResp{}, 0, err
	}

	return entity.DataResp{
		RespCode: 200,
		RespMsg:  "토큰 발급 성공",
		Data:     accessToken,
	}, 200, nil
}

func (s *TokenService) generateRefreshToken(user entity.User) (string, time.Time, error) {
	return s.generateToken(user, 1, s.config.RefreshTokenTTL)
}

func (s *TokenService) generateToken(user entity.User, tokenType int, ttl time.Duration) (string, time.Time, error) {
	if s.config.JWTAlgorithm != "" && s.config.JWTAlgorithm != "HS256" {
		return "", time.Time{}, fmt.Errorf("unsupported jwt algorithm: %s", s.config.JWTAlgorithm)
	}

	now := time.Now().UTC()
	exp := now.Add(ttl)

	headerPayload, err := json.Marshal(tokenHeader{Alg: "HS256", Typ: "JWT"})
	if err != nil {
		return "", time.Time{}, err
	}

	claimsPayload, err := json.Marshal(tokenClaims{
		UserNo:    user.UserNo,
		UserNick:  user.UserNick,
		Type:      tokenType,
		Timestamp: now.Format("2006-01-02 15:04:05.000000"),
		Exp:       exp.Unix(),
		Iat:       now.Unix(),
		Nbf:       now.Unix(),
	})
	if err != nil {
		return "", time.Time{}, err
	}

	encodedHeader := encodeBase64URL(headerPayload)
	encodedClaims := encodeBase64URL(claimsPayload)
	signingInput := encodedHeader + "." + encodedClaims

	mac := hmac.New(sha256.New, []byte(s.config.JWTSecretKey))
	if _, err := mac.Write([]byte(signingInput)); err != nil {
		return "", time.Time{}, err
	}

	signature := encodeBase64URL(mac.Sum(nil))

	return signingInput + "." + signature, exp, nil
}

func (s *TokenService) decodeToken(token string) (tokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return tokenClaims{}, ErrDecodeError
	}

	headerBytes, err := decodeBase64URL(parts[0])
	if err != nil {
		return tokenClaims{}, err
	}

	var header tokenHeader
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return tokenClaims{}, err
	}
	if header.Alg != "HS256" {
		return tokenClaims{}, ErrInvalidToken
	}

	signingInput := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, []byte(s.config.JWTSecretKey))
	if _, err := mac.Write([]byte(signingInput)); err != nil {
		return tokenClaims{}, err
	}
	expectedSignature := encodeBase64URL(mac.Sum(nil))
	if !hmac.Equal([]byte(expectedSignature), []byte(parts[2])) {
		return tokenClaims{}, errors.New("Signature verification failed")
	}

	payloadBytes, err := decodeBase64URL(parts[1])
	if err != nil {
		return tokenClaims{}, err
	}

	var claims tokenClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return tokenClaims{}, err
	}

	return claims, nil
}

func encodeBase64URL(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

func decodeBase64URL(data string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(data)
}
