package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/ReadingGarden/back-go/internal/auth/entity"
	authservice "github.com/ReadingGarden/back-go/internal/auth/service"
)

type Service interface {
	CreateUser(c *gin.Context, payload entity.CreateUserPayload) (int, interface{}, error)
	Login(c *gin.Context, payload entity.LoginPayload) (int, interface{}, error)
	Logout(c *gin.Context, authorization string) (int, interface{}, error)
	Refresh(c *gin.Context, payload entity.RefreshTokenPayload) (int, interface{}, error)
	DeleteUser(c *gin.Context, authorization string) (int, interface{}, error)
	FindPassword(c *gin.Context, payload entity.UserEmailPayload) (int, interface{}, error)
	AuthCheck(c *gin.Context, payload entity.UserPasswordAuthPayload) (int, interface{}, error)
	UpdatePasswordNoToken(c *gin.Context, payload entity.UpdateUserPasswordPayload) (int, interface{}, error)
	UpdatePassword(c *gin.Context, authorization string, payload entity.UpdateUserPasswordPayload) (int, interface{}, error)
	GetUser(c *gin.Context, authorization string) (int, interface{}, error)
	UpdateUser(c *gin.Context, authorization string, payload entity.UpdateUserPayload) (int, interface{}, error)
}

type ginServiceAdapter struct {
	svc *authservice.Service
}

func (a ginServiceAdapter) CreateUser(c *gin.Context, payload entity.CreateUserPayload) (int, interface{}, error) {
	return a.svc.CreateUser(c.Request.Context(), payload)
}

func (a ginServiceAdapter) Login(c *gin.Context, payload entity.LoginPayload) (int, interface{}, error) {
	return a.svc.Login(c.Request.Context(), payload)
}

func (a ginServiceAdapter) Logout(c *gin.Context, authorization string) (int, interface{}, error) {
	return a.svc.Logout(c.Request.Context(), authorization)
}

func (a ginServiceAdapter) Refresh(c *gin.Context, payload entity.RefreshTokenPayload) (int, interface{}, error) {
	return a.svc.Refresh(c.Request.Context(), payload)
}

func (a ginServiceAdapter) DeleteUser(c *gin.Context, authorization string) (int, interface{}, error) {
	return a.svc.DeleteUser(c.Request.Context(), authorization)
}

func (a ginServiceAdapter) FindPassword(c *gin.Context, payload entity.UserEmailPayload) (int, interface{}, error) {
	return a.svc.FindPassword(c.Request.Context(), payload)
}

func (a ginServiceAdapter) AuthCheck(c *gin.Context, payload entity.UserPasswordAuthPayload) (int, interface{}, error) {
	return a.svc.AuthCheck(c.Request.Context(), payload)
}

func (a ginServiceAdapter) UpdatePasswordNoToken(c *gin.Context, payload entity.UpdateUserPasswordPayload) (int, interface{}, error) {
	return a.svc.UpdatePasswordNoToken(c.Request.Context(), payload)
}

func (a ginServiceAdapter) UpdatePassword(c *gin.Context, authorization string, payload entity.UpdateUserPasswordPayload) (int, interface{}, error) {
	return a.svc.UpdatePassword(c.Request.Context(), authorization, payload)
}

func (a ginServiceAdapter) GetUser(c *gin.Context, authorization string) (int, interface{}, error) {
	return a.svc.GetUser(c.Request.Context(), authorization)
}

func (a ginServiceAdapter) UpdateUser(c *gin.Context, authorization string, payload entity.UpdateUserPayload) (int, interface{}, error) {
	return a.svc.UpdateUser(c.Request.Context(), authorization, payload)
}

func RegisterRoutes(group *gin.RouterGroup, service *authservice.Service) {
	register(group, ginServiceAdapter{svc: service})
}

func register(group *gin.RouterGroup, service Service) {
	group.POST("", func(c *gin.Context) {
		payload, ok := decodeCreateUserPayload(c)
		if !ok {
			return
		}
		status, body, err := service.CreateUser(c, payload)
		respond(c, status, body, err)
	})

	group.POST("/login", func(c *gin.Context) {
		payload, ok := decodeLoginPayload(c)
		if !ok {
			return
		}
		status, body, err := service.Login(c, payload)
		respond(c, status, body, err)
	})

	protected := group.Group("")
	protected.Use(requireBearerHeader())

	protected.POST("/logout", func(c *gin.Context) {
		status, body, err := service.Logout(c, c.GetHeader("Authorization"))
		respond(c, status, body, err)
	})
	protected.DELETE("", func(c *gin.Context) {
		status, body, err := service.DeleteUser(c, c.GetHeader("Authorization"))
		respond(c, status, body, err)
	})
	protected.PUT("/update-password", func(c *gin.Context) {
		payload, ok := decodeUpdateUserPasswordPayload(c)
		if !ok {
			return
		}
		status, body, err := service.UpdatePassword(c, c.GetHeader("Authorization"), payload)
		respond(c, status, body, err)
	})
	protected.GET("", func(c *gin.Context) {
		status, body, err := service.GetUser(c, c.GetHeader("Authorization"))
		respond(c, status, body, err)
	})
	protected.PUT("", func(c *gin.Context) {
		payload, ok := decodeUpdateUserPayload(c)
		if !ok {
			return
		}
		status, body, err := service.UpdateUser(c, c.GetHeader("Authorization"), payload)
		respond(c, status, body, err)
	})

	group.POST("/refresh", func(c *gin.Context) {
		payload, ok := decodeRefreshTokenPayload(c)
		if !ok {
			return
		}
		status, body, err := service.Refresh(c, payload)
		respond(c, status, body, err)
	})
	group.POST("/find-password", func(c *gin.Context) {
		payload, ok := decodeUserEmailPayload(c)
		if !ok {
			return
		}
		status, body, err := service.FindPassword(c, payload)
		respond(c, status, body, err)
	})
	group.POST("/find-password/check", func(c *gin.Context) {
		payload, ok := decodeUserPasswordAuthPayload(c)
		if !ok {
			return
		}
		status, body, err := service.AuthCheck(c, payload)
		respond(c, status, body, err)
	})
	group.PUT("/find-password/update-password", func(c *gin.Context) {
		payload, ok := decodeUpdateUserPasswordPayload(c)
		if !ok {
			return
		}
		status, body, err := service.UpdatePasswordNoToken(c, payload)
		respond(c, status, body, err)
	})
}

func requireBearerHeader() gin.HandlerFunc {
	return func(c *gin.Context) {
		authorization := c.GetHeader("Authorization")
		if !strings.HasPrefix(authorization, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"detail": "Unauthorized"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func respond(c *gin.Context, status int, body interface{}, err error) {
	if err != nil {
		if errors.Is(err, authservice.ErrDecodeError) || errors.Is(err, authservice.ErrExpiredSignature) || errors.Is(err, authservice.ErrInvalidToken) {
			c.JSON(http.StatusUnauthorized, entity.HTTPResp{RespCode: 401, RespMsg: err.Error()})
			return
		}
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(status, body)
}
