package delivery

import (
	"chatly/internal/config"
	"chatly/internal/middleware"
	"chatly/internal/websocket"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "chatly/docs"
)

func NewRouter(cfg *config.Config, rdb *redis.Client, authHandler *AuthHandler, msgHandler *MessageHandler, wsHandler *websocket.Handler) *gin.Engine {
	r := gin.Default()

	r.Static("/static", "./static")
	r.StaticFile("/", "./static/index.html")
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api")
	api.Use(middleware.RateLimiter(rdb, 20, time.Minute))
	{
		api.POST("/register", authHandler.Register)
		api.POST("/login", authHandler.Login)
		api.POST("/refresh", authHandler.Refresh)

		// Public history
		api.GET("/messages", msgHandler.GetHistory)

		// Protected routes
		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware(cfg))
		{
			protected.GET("/me", authHandler.Me)
			protected.GET("/ws", wsHandler.HandleWS(cfg))
		}
	}

	return r
}
