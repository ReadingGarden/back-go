package app

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/ReadingGarden/back-go/internal/config"
	httpRouter "github.com/ReadingGarden/back-go/internal/http/router"
)

type App struct {
	config config.Config
	router *gin.Engine
}

func New(cfg config.Config) (*App, error) {
	router, err := httpRouter.New(cfg)
	if err != nil {
		return nil, err
	}

	return &App{
		config: cfg,
		router: router,
	}, nil
}

func (a *App) Run() error {
	return a.router.Run(fmt.Sprintf(":%s", a.config.Port))
}
