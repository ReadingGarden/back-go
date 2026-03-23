package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/ReadingGarden/back-go/internal/auth/entity"
)

type mockService struct {
	createUserPayload            entity.CreateUserPayload
	loginPayload                 entity.LoginPayload
	findPasswordPayload          entity.UserEmailPayload
	authCheckPayload             entity.UserPasswordAuthPayload
	updatePasswordNoTokenPayload entity.UpdateUserPasswordPayload
	updateUserPayload            entity.UpdateUserPayload
	refreshPayload               entity.RefreshTokenPayload
	authorization                string
}

func (m *mockService) CreateUser(_ *gin.Context, payload entity.CreateUserPayload) (int, interface{}, error) {
	m.createUserPayload = payload
	return http.StatusCreated, entity.DataResp{RespCode: 201, RespMsg: "회원가입 성공", Data: map[string]interface{}{}}, nil
}

func (m *mockService) Login(_ *gin.Context, payload entity.LoginPayload) (int, interface{}, error) {
	m.loginPayload = payload
	return http.StatusOK, entity.DataResp{RespCode: 200, RespMsg: "로그인 성공", Data: map[string]interface{}{}}, nil
}

func (m *mockService) Logout(_ *gin.Context, authorization string) (int, interface{}, error) {
	m.authorization = authorization
	return http.StatusOK, entity.DataResp{RespCode: 200, RespMsg: "로그아웃 성공", Data: map[string]interface{}{}}, nil
}

func (m *mockService) Refresh(_ *gin.Context, payload entity.RefreshTokenPayload) (int, interface{}, error) {
	m.refreshPayload = payload
	return http.StatusOK, entity.DataResp{RespCode: 200, RespMsg: "토큰 발급 성공", Data: "token"}, nil
}

func (m *mockService) DeleteUser(_ *gin.Context, authorization string) (int, interface{}, error) {
	m.authorization = authorization
	return http.StatusOK, entity.DataResp{RespCode: 200, RespMsg: "회원 탈퇴 성공", Data: map[string]interface{}{}}, nil
}

func (m *mockService) FindPassword(_ *gin.Context, payload entity.UserEmailPayload) (int, interface{}, error) {
	m.findPasswordPayload = payload
	return http.StatusOK, entity.DataResp{RespCode: 200, RespMsg: "메일이 발송되었습니다. 확인해주세요.", Data: map[string]interface{}{}}, nil
}

func (m *mockService) AuthCheck(_ *gin.Context, payload entity.UserPasswordAuthPayload) (int, interface{}, error) {
	m.authCheckPayload = payload
	return http.StatusOK, entity.DataResp{RespCode: 200, RespMsg: "인증 성공", Data: map[string]interface{}{}}, nil
}

func (m *mockService) UpdatePasswordNoToken(_ *gin.Context, payload entity.UpdateUserPasswordPayload) (int, interface{}, error) {
	m.updatePasswordNoTokenPayload = payload
	return http.StatusOK, entity.HTTPResp{RespCode: 200, RespMsg: "비밀번호 변경 성공"}, nil
}

func (m *mockService) UpdatePassword(_ *gin.Context, authorization string, payload entity.UpdateUserPasswordPayload) (int, interface{}, error) {
	m.authorization = authorization
	m.updatePasswordNoTokenPayload = payload
	return http.StatusOK, entity.HTTPResp{RespCode: 200, RespMsg: "비밀번호 변경 성공"}, nil
}

func (m *mockService) GetUser(_ *gin.Context, authorization string) (int, interface{}, error) {
	m.authorization = authorization
	return http.StatusOK, entity.DataResp{RespCode: 200, RespMsg: "조회 성공", Data: map[string]interface{}{}}, nil
}

func (m *mockService) UpdateUser(_ *gin.Context, authorization string, payload entity.UpdateUserPayload) (int, interface{}, error) {
	m.authorization = authorization
	m.updateUserPayload = payload
	return http.StatusOK, entity.DataResp{RespCode: 200, RespMsg: "프로필 변경 성공", Data: map[string]interface{}{}}, nil
}

func TestRegisterRoutes_AuthEndpoints(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	service := &mockService{}
	engine := gin.New()
	register(engine.Group("/api/v1/auth"), service)

	tests := []struct {
		name       string
		method     string
		target     string
		body       string
		auth       string
		wantStatus int
		verify     func(t *testing.T)
	}{
		{
			name:       "create user keeps defaults for omitted fields",
			method:     http.MethodPost,
			target:     "/api/v1/auth",
			body:       `{}`,
			wantStatus: http.StatusCreated,
			verify: func(t *testing.T) {
				if service.createUserPayload.UserEmail != "" || service.createUserPayload.UserPassword != "" || service.createUserPayload.UserFCM != "" {
					t.Fatalf("createUserPayload = %+v, want zero-value strings", service.createUserPayload)
				}
			},
		},
		{
			name:       "login validates required fields",
			method:     http.MethodPost,
			target:     "/api/v1/auth/login",
			body:       `{}`,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "logout requires bearer header",
			method:     http.MethodPost,
			target:     "/api/v1/auth/logout",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "logout passes authorization header",
			method:     http.MethodPost,
			target:     "/api/v1/auth/logout",
			auth:       "Bearer access-token",
			wantStatus: http.StatusOK,
			verify: func(t *testing.T) {
				if service.authorization != "Bearer access-token" {
					t.Fatalf("authorization = %q, want Bearer access-token", service.authorization)
				}
			},
		},
		{
			name:       "refresh validates required body",
			method:     http.MethodPost,
			target:     "/api/v1/auth/refresh",
			body:       `{}`,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "refresh forwards payload",
			method:     http.MethodPost,
			target:     "/api/v1/auth/refresh",
			body:       `{"refresh_token":"refresh-token"}`,
			wantStatus: http.StatusOK,
			verify: func(t *testing.T) {
				if service.refreshPayload.RefreshToken != "refresh-token" {
					t.Fatalf("refreshPayload = %+v", service.refreshPayload)
				}
			},
		},
		{
			name:       "find password forwards payload",
			method:     http.MethodPost,
			target:     "/api/v1/auth/find-password",
			body:       `{"user_email":"user@example.com"}`,
			wantStatus: http.StatusOK,
			verify: func(t *testing.T) {
				if service.findPasswordPayload.UserEmail != "user@example.com" {
					t.Fatalf("findPasswordPayload = %+v", service.findPasswordPayload)
				}
			},
		},
		{
			name:       "auth check forwards payload",
			method:     http.MethodPost,
			target:     "/api/v1/auth/find-password/check",
			body:       `{"user_email":"user@example.com","auth_number":"abcde"}`,
			wantStatus: http.StatusOK,
			verify: func(t *testing.T) {
				if service.authCheckPayload.UserEmail != "user@example.com" || service.authCheckPayload.AuthNumber != "abcde" {
					t.Fatalf("authCheckPayload = %+v", service.authCheckPayload)
				}
			},
		},
		{
			name:       "update password without token allows omitted email",
			method:     http.MethodPut,
			target:     "/api/v1/auth/find-password/update-password",
			body:       `{"user_password":"pw"}`,
			wantStatus: http.StatusOK,
			verify: func(t *testing.T) {
				if service.updatePasswordNoTokenPayload.UserEmail != nil {
					t.Fatalf("user email = %v, want nil", *service.updatePasswordNoTokenPayload.UserEmail)
				}
			},
		},
		{
			name:       "get user route exists",
			method:     http.MethodGet,
			target:     "/api/v1/auth",
			auth:       "Bearer access-token",
			wantStatus: http.StatusOK,
		},
		{
			name:       "update user forwards nil values",
			method:     http.MethodPut,
			target:     "/api/v1/auth",
			body:       `{}`,
			auth:       "Bearer access-token",
			wantStatus: http.StatusOK,
			verify: func(t *testing.T) {
				if service.updateUserPayload.UserNick != nil || service.updateUserPayload.UserImage != nil {
					t.Fatalf("updateUserPayload = %+v, want nil fields", service.updateUserPayload)
				}
			},
		},
		{
			name:       "delete user route exists",
			method:     http.MethodDelete,
			target:     "/api/v1/auth",
			auth:       "Bearer access-token",
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var body *bytes.Reader
			if tt.body == "" {
				body = bytes.NewReader(nil)
			} else {
				body = bytes.NewReader([]byte(tt.body))
			}

			req := httptest.NewRequest(tt.method, tt.target, body)
			if tt.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}
			if tt.auth != "" {
				req.Header.Set("Authorization", tt.auth)
			}
			rec := httptest.NewRecorder()

			engine.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d, body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if rec.Header().Get("Set-Cookie") != "" {
				t.Fatalf("Set-Cookie = %q, want empty", rec.Header().Get("Set-Cookie"))
			}
			if contentType := rec.Header().Get("Content-Type"); contentType == "" {
				t.Fatal("Content-Type is empty")
			}
			if tt.verify != nil {
				tt.verify(t)
			}
		})
	}
}

func TestRegisterRoutes_ProtectedRouteUnauthorizedShape(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	register(engine.Group("/api/v1/auth"), &mockService{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth", nil)
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if body["detail"] != "Unauthorized" {
		t.Fatalf("detail = %q, want %q", body["detail"], "Unauthorized")
	}
}
