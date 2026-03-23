package service

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/ReadingGarden/back-go/internal/auth/entity"
	"github.com/ReadingGarden/back-go/internal/auth/repository"
)

const authMailSubject = "[독서가든] 인증번호 안내드립니다"

type AuthNumberResetScheduler interface {
	ScheduleReset(userNo int64, after time.Duration)
}

type AfterFuncResetScheduler struct {
	repo *repository.MySQLRepository
}

func NewAfterFuncResetScheduler(repo *repository.MySQLRepository) *AfterFuncResetScheduler {
	return &AfterFuncResetScheduler{repo: repo}
}

func (s *AfterFuncResetScheduler) ScheduleReset(userNo int64, after time.Duration) {
	time.AfterFunc(after, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = s.repo.UpdateUserAuthNumber(ctx, userNo, nil)
	})
}

type Service struct {
	repo           *repository.MySQLRepository
	tokenService   *TokenService
	mailer         Mailer
	resetScheduler AuthNumberResetScheduler
}

func New(repo *repository.MySQLRepository, tokenService *TokenService, mailer Mailer, resetScheduler AuthNumberResetScheduler) *Service {
	return &Service{
		repo:           repo,
		tokenService:   tokenService,
		mailer:         mailer,
		resetScheduler: resetScheduler,
	}
}

func (s *Service) CreateUser(ctx context.Context, payload entity.CreateUserPayload) (int, interface{}, error) {
	if payload.UserSocialID != "" {
		if _, err := s.repo.FindUserBySocial(ctx, payload.UserSocialID, payload.UserSocialType); err == nil {
			return http.StatusConflict, entity.HTTPResp{RespCode: 409, RespMsg: "소셜 아이디 중복"}, nil
		} else if !errors.Is(err, repository.ErrNotFound) {
			return 0, nil, err
		}
	} else {
		if _, err := s.repo.FindUserByEmail(ctx, payload.UserEmail); err == nil {
			return http.StatusConflict, entity.HTTPResp{RespCode: 409, RespMsg: "이메일 중복"}, nil
		} else if !errors.Is(err, repository.ErrNotFound) {
			return 0, nil, err
		}
	}

	if payload.UserPassword != "" {
		hashedPassword, err := hashPassword(payload.UserPassword)
		if err != nil {
			return 0, nil, err
		}
		payload.UserPassword = hashedPassword
	}

	user, err := s.repo.CreateUserGraph(ctx, payload, generateRandomNick())
	if err != nil {
		return 0, nil, err
	}

	tokenPair, err := s.tokenService.GeneratePairToken(ctx, user)
	if err != nil {
		return 0, nil, err
	}

	response := map[string]interface{}{
		"access_token":  tokenPair.AccessToken,
		"refresh_token": tokenPair.RefreshToken,
		"user_nick":     user.UserNick,
	}

	return http.StatusCreated, entity.DataResp{
		RespCode: 201,
		RespMsg:  "회원가입 성공",
		Data:     response,
	}, nil
}

func (s *Service) Login(ctx context.Context, payload entity.LoginPayload) (int, interface{}, error) {
	var (
		user entity.User
		err  error
	)

	if payload.UserSocialID != "" {
		user, err = s.repo.FindUserBySocial(ctx, payload.UserSocialID, payload.UserSocialType)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return http.StatusBadRequest, entity.HTTPResp{RespCode: 400, RespMsg: "등록되지 않은 소셜입니다"}, nil
			}
			return 0, nil, err
		}
	} else {
		user, err = s.repo.FindUserByEmail(ctx, payload.UserEmail)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return http.StatusBadRequest, entity.HTTPResp{RespCode: 400, RespMsg: "등록되지 않은 이메일 주소입니다."}, nil
			}
			return 0, nil, err
		}
		if !verifyPassword(payload.UserPassword, user.UserPassword) {
			return http.StatusBadRequest, entity.HTTPResp{RespCode: 400, RespMsg: "비밀번호가 일치하지 않습니다."}, nil
		}
	}

	tokenPair, err := s.tokenService.GeneratePairToken(ctx, user)
	if err != nil {
		return 0, nil, err
	}

	if _, err := s.repo.UpdateUserFCM(ctx, user.UserNo, payload.UserFCM); err != nil {
		return 0, nil, err
	}

	return http.StatusOK, entity.DataResp{
		RespCode: 200,
		RespMsg:  "로그인 성공",
		Data: map[string]interface{}{
			"access_token":  tokenPair.AccessToken,
			"refresh_token": tokenPair.RefreshToken,
		},
	}, nil
}

func (s *Service) Logout(ctx context.Context, authorization string) (int, interface{}, error) {
	user, status, body, err := s.authorize(ctx, authorization)
	if body != nil || err != nil {
		return status, body, err
	}

	refreshToken, err := s.repo.FindFirstRefreshTokenByUserNo(ctx, user.UserNo)
	if err != nil {
		return 0, nil, err
	}
	if err := s.repo.DeleteRefreshTokenByID(ctx, refreshToken.ID); err != nil {
		return 0, nil, err
	}
	if _, err := s.repo.ClearUserFCM(ctx, user.UserNo); err != nil {
		return 0, nil, err
	}

	return http.StatusOK, entity.DataResp{
		RespCode: 200,
		RespMsg:  "로그아웃 성공",
		Data:     map[string]interface{}{},
	}, nil
}

func (s *Service) DeleteUser(ctx context.Context, authorization string) (int, interface{}, error) {
	user, status, body, err := s.authorize(ctx, authorization)
	if body != nil || err != nil {
		return status, body, err
	}

	if err := s.repo.DeleteUserGraph(ctx, user); err != nil {
		return 0, nil, err
	}

	return http.StatusOK, entity.DataResp{
		RespCode: 200,
		RespMsg:  "회원 탈퇴 성공",
		Data:     map[string]interface{}{},
	}, nil
}

func (s *Service) FindPassword(ctx context.Context, payload entity.UserEmailPayload) (int, interface{}, error) {
	user, err := s.repo.FindUserByEmail(ctx, payload.UserEmail)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return http.StatusBadRequest, entity.HTTPResp{RespCode: 400, RespMsg: "등록되지 않은 이메일 주소입니다."}, nil
		}
		return 0, nil, err
	}

	authNumber := generateRandomString(5)
	if err := s.mailer.Send(payload.UserEmail, authMailSubject, authNumber); err != nil {
		return http.StatusInternalServerError, entity.HTTPResp{RespCode: 500, RespMsg: "메일 전송 실패"}, nil
	}

	if _, err := s.repo.UpdateUserAuthNumber(ctx, user.UserNo, &authNumber); err != nil {
		return 0, nil, err
	}
	if s.resetScheduler != nil {
		s.resetScheduler.ScheduleReset(user.UserNo, 5*time.Minute)
	}

	return http.StatusOK, entity.DataResp{
		RespCode: 200,
		RespMsg:  "메일이 발송되었습니다. 확인해주세요.",
		Data:     map[string]interface{}{},
	}, nil
}

func (s *Service) AuthCheck(ctx context.Context, payload entity.UserPasswordAuthPayload) (int, interface{}, error) {
	user, err := s.repo.FindUserByEmail(ctx, payload.UserEmail)
	if err != nil {
		return 0, nil, err
	}

	if user.UserAuthNumber != nil && *user.UserAuthNumber == payload.AuthNumber {
		return http.StatusOK, entity.DataResp{
			RespCode: 200,
			RespMsg:  "인증 성공",
			Data:     map[string]interface{}{},
		}, nil
	}

	return http.StatusBadRequest, entity.HTTPResp{RespCode: 400, RespMsg: "인증번호 불일치"}, nil
}

func (s *Service) UpdatePasswordNoToken(ctx context.Context, payload entity.UpdateUserPasswordPayload) (int, interface{}, error) {
	if payload.UserEmail == nil {
		return 0, nil, errors.New("user_email is nil")
	}

	user, err := s.repo.FindUserByEmail(ctx, *payload.UserEmail)
	if err != nil {
		return 0, nil, err
	}

	hashedPassword, err := hashPassword(payload.UserPassword)
	if err != nil {
		return 0, nil, err
	}

	if _, err := s.repo.UpdateUserPassword(ctx, user.UserNo, hashedPassword); err != nil {
		return 0, nil, err
	}

	return http.StatusOK, entity.HTTPResp{RespCode: 200, RespMsg: "비밀번호 변경 성공"}, nil
}

func (s *Service) UpdatePassword(ctx context.Context, authorization string, payload entity.UpdateUserPasswordPayload) (int, interface{}, error) {
	user, status, body, err := s.authorize(ctx, authorization)
	if body != nil || err != nil {
		return status, body, err
	}

	hashedPassword, err := hashPassword(payload.UserPassword)
	if err != nil {
		return 0, nil, err
	}
	if _, err := s.repo.UpdateUserPassword(ctx, user.UserNo, hashedPassword); err != nil {
		return 0, nil, err
	}

	return http.StatusOK, entity.HTTPResp{RespCode: 200, RespMsg: "비밀번호 변경 성공"}, nil
}

func (s *Service) GetUser(ctx context.Context, authorization string) (int, interface{}, error) {
	user, status, body, err := s.authorize(ctx, authorization)
	if body != nil || err != nil {
		return status, body, err
	}

	gardenCount, err := s.repo.CountGardenUsersByUser(ctx, user.UserNo)
	if err != nil {
		return 0, nil, err
	}
	readCount, err := s.repo.CountBooksByStatus(ctx, user.UserNo, 1)
	if err != nil {
		return 0, nil, err
	}
	likeCount, err := s.repo.CountBooksByStatus(ctx, user.UserNo, 2)
	if err != nil {
		return 0, nil, err
	}

	return http.StatusOK, entity.DataResp{
		RespCode: 200,
		RespMsg:  "조회 성공",
		Data: entity.GetUserData{
			UserNo:         user.UserNo,
			UserNick:       user.UserNick,
			UserEmail:      user.UserEmail,
			UserSocialType: user.UserSocialType,
			UserImage:      user.UserImage,
			UserCreatedAt:  user.UserCreatedAt,
			GardenCount:    gardenCount,
			ReadBookCount:  readCount,
			LikeBookCount:  likeCount,
		},
	}, nil
}

func (s *Service) UpdateUser(ctx context.Context, authorization string, payload entity.UpdateUserPayload) (int, interface{}, error) {
	user, status, body, err := s.authorize(ctx, authorization)
	if body != nil || err != nil {
		return status, body, err
	}

	updatedUser, err := s.repo.UpdateUserProfile(ctx, user.UserNo, payload.UserNick, payload.UserImage)
	if err != nil {
		return 0, nil, err
	}

	return http.StatusOK, entity.DataResp{
		RespCode: 200,
		RespMsg:  "프로필 변경 성공",
		Data:     entity.NewUserProfile(updatedUser),
	}, nil
}

func (s *Service) Refresh(ctx context.Context, payload entity.RefreshTokenPayload) (int, interface{}, error) {
	response, status, err := s.tokenService.Refresh(ctx, payload)
	if err != nil {
		switch {
		case errors.Is(err, ErrExpiredSignature), errors.Is(err, ErrInvalidToken), errors.Is(err, ErrDecodeError):
			return http.StatusUnauthorized, entity.HTTPResp{RespCode: 401, RespMsg: err.Error()}, nil
		default:
			return 0, nil, err
		}
	}

	return status, response, nil
}

func (s *Service) authorize(ctx context.Context, authorization string) (entity.User, int, interface{}, error) {
	token := extractBearerToken(authorization)
	if token == "" {
		return entity.User{}, http.StatusInternalServerError, entity.HTTPResp{
			RespCode: 500,
			RespMsg:  "유효하지 않은 토큰 값입니다.",
		}, nil
	}

	claims, err := s.tokenService.VerifyAccessToken(token)
	if err != nil {
		switch {
		case errors.Is(err, ErrExpiredSignature), errors.Is(err, ErrInvalidToken), errors.Is(err, ErrDecodeError):
			return entity.User{}, http.StatusUnauthorized, entity.HTTPResp{
				RespCode: 401,
				RespMsg:  err.Error(),
			}, nil
		default:
			return entity.User{}, 0, nil, err
		}
	}

	user, err := s.repo.FindUserByNo(ctx, claims.UserNo)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return entity.User{}, http.StatusBadRequest, entity.HTTPResp{
				RespCode: 400,
				RespMsg:  "일치하는 사용자 정보가 없습니다.",
			}, nil
		}
		return entity.User{}, 0, nil, err
	}

	return user, 0, nil, nil
}

func extractBearerToken(authorization string) string {
	if authorization == "" {
		return ""
	}

	parts := strings.Split(authorization, " ")
	if len(parts) < 2 {
		return ""
	}

	return parts[1]
}
