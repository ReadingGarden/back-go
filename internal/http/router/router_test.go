package router_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ReadingGarden/back-go/internal/config"
	"github.com/ReadingGarden/back-go/internal/http/router"
)

func TestNew(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		method     string
		target     string
		wantStatus int
	}{
		{
			name:       "health check",
			method:     http.MethodGet,
			target:     "/healthz",
			wantStatus: http.StatusOK,
		},
		{
			name:       "unknown api route",
			method:     http.MethodGet,
			target:     "/api/v1/book/unknown",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "swagger ui",
			method:     http.MethodGet,
			target:     "/swagger/index.html",
			wantStatus: http.StatusOK,
		},
	}

	engine, err := router.New(config.Config{
		AppEnv:  "test",
		GinMode: "test",
		Port:    "8080",
		Swagger: config.SwaggerConfig{
			Enabled:  true,
			BasePath: "/api/v1",
		},
	})
	if err != nil {
		t.Fatalf("router.New() error = %v", err)
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(tt.method, tt.target, nil)
			rec := httptest.NewRecorder()

			engine.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}
