package api

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opencode-ai/polaris/internal/api/handlers"
	"github.com/opencode-ai/polaris/internal/api/middleware"
	"github.com/opencode-ai/polaris/internal/services"
	"github.com/redis/go-redis/v9"
)

func NewRouter(pg *pgxpool.Pool, rdb *redis.Client, jwtSecret string) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	authSvc := services.NewAuthService(pg, rdb, jwtSecret)

	authMw := middleware.NewAuthMiddleware(authSvc)

	authHandler := handlers.NewAuthHandler(authSvc)
	healthHandler := handlers.NewHealthHandler(pg, rdb)

	r.GET("/health", healthHandler.Health)

	apiV1 := r.Group("/api/v1")
	{
		auth := apiV1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		auth.POST("/auth/logout", authMw.RequireAuth(), authHandler.Logout)
		auth.GET("/auth/session", authMw.RequireAuth(), authHandler.Session)
	}

	return r
}
