package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/opencode-ai/polaris/internal/api"
	"github.com/opencode-ai/polaris/internal/db"
	"github.com/opencode-ai/polaris/internal/api/middleware"
	"github.com/rs/zerolog"
)

func main() {
	// Structured logging
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}).
		Level(zerolog.DebugLevel).
		With().
		Timestamp().
		Caller().
		Logger()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://polaris:polaris@localhost:5432/polaris?sslmode=disable"
		logger.Warn().Msg("DATABASE_URL not set, using default dev connection")
	}
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
		logger.Warn().Msg("REDIS_ADDR not set, using default localhost:6379")
	}
	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		logger.Warn().Msg("JWT_SECRET not set, using dev secret - DO NOT USE IN PRODUCTION")
		jwtSecret = "dev-secret-change-in-production"
	}

	logger.Info().Str("addr", addr).Msg("connecting to postgres")
	pg, err := db.NewPostgres(dsn)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to connect to postgres")
	}
	defer pg.Close()
	logger.Info().Msg("postgres connected")

	logger.Info().Str("addr", redisAddr).Msg("connecting to redis")
	rdb, err := db.NewRedis(redisAddr)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to connect to redis")
	}
	defer rdb.Close()
	logger.Info().Msg("redis connected")

	gin.SetMode(gin.ReleaseMode)
	router := api.NewRouter(pg, rdb, jwtSecret)
	// Add logging middleware at the top of the chain
	router.Use(middleware.Logging(logger))

	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		logger.Info().Str("addr", addr).Msg("server starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("server failed")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	logger.Info().Str("signal", sig.String()).Msg("shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal().Err(err).Msg("shutdown failed")
	}
	logger.Info().Msg("server stopped")
}
