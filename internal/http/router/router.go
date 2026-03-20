package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/ReadingGarden/back-go/docs"
	authhandler "github.com/ReadingGarden/back-go/internal/auth/handler"
	bookhandler "github.com/ReadingGarden/back-go/internal/book/handler"
	"github.com/ReadingGarden/back-go/internal/config"
	gardenhandler "github.com/ReadingGarden/back-go/internal/garden/handler"
	memohandler "github.com/ReadingGarden/back-go/internal/memo/handler"
	pushhandler "github.com/ReadingGarden/back-go/internal/push/handler"
)

func New(cfg config.Config) (*gin.Engine, error) {
	gin.SetMode(cfg.GinMode)

	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())

	if err := engine.SetTrustedProxies(nil); err != nil {
		return nil, err
	}

	engine.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	if cfg.Swagger.Enabled {
		docs.SwaggerInfo.Title = "ReadingGarden API"
		docs.SwaggerInfo.Description = "Gin + sqlc migration backend for ReadingGarden."
		docs.SwaggerInfo.Version = "1.0"
		docs.SwaggerInfo.Host = cfg.Swagger.Host
		docs.SwaggerInfo.BasePath = cfg.Swagger.BasePath
		docs.SwaggerInfo.Schemes = []string{"http", "https"}

		engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	}

	apiV1 := engine.Group("/api/v1")
	authhandler.RegisterRoutes(apiV1.Group("/auth"))
	gardenhandler.RegisterRoutes(apiV1.Group("/garden"))
	bookhandler.RegisterRoutes(apiV1.Group("/book"))
	memohandler.RegisterRoutes(apiV1.Group("/memo"))
	pushhandler.RegisterRoutes(apiV1.Group("/push"))

	return engine, nil
}
