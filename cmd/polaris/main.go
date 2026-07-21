package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opencode-ai/polaris/internal/api"
	"github.com/opencode-ai/polaris/internal/api/middleware"
	"github.com/opencode-ai/polaris/internal/db"
	"github.com/opencode-ai/polaris/internal/worker"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
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

	docsDir := os.Getenv("DOCS_DIR")
	if docsDir == "" {
		docsDir = "./docs/book"
	}

	logger.Info().Str("addr", addr).Msg("connecting to postgres")
	pg, err := db.NewPostgres(dsn)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to connect to postgres")
	}
	defer pg.Close()
	logger.Info().Msg("postgres connected")

	if err := db.RunMigrations(pg, "internal/db/migrations"); err != nil {
		logger.Fatal().Err(err).Msg("failed to run migrations")
	}
	logger.Info().Msg("migrations complete")

	logger.Info().Str("addr", redisAddr).Msg("connecting to redis")
	rdb, err := db.NewRedis(redisAddr)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to connect to redis")
	}
	defer rdb.Close()
	logger.Info().Msg("redis connected")

	gin.SetMode(gin.ReleaseMode)
	router := api.NewRouter(pg, rdb, jwtSecret)

	sched := worker.NewScheduler(pg, rdb)
	schedCtx, schedCancel := context.WithCancel(context.Background())
	defer schedCancel()
	go sched.Start(schedCtx)
	router.Use(middleware.Logging(logger))

	// Seed admin user if ADMIN_EMAIL and ADMIN_PASSWORD are set
	adminEmail := os.Getenv("ADMIN_EMAIL")
	adminPass := os.Getenv("ADMIN_PASSWORD")
	if adminEmail != "" && adminPass != "" {
		err := seedAdmin(pg, adminEmail, adminPass)
		if err != nil {
			logger.Warn().Err(err).Msg("failed to seed admin user")
		} else {
			logger.Info().Str("email", adminEmail).Msg("admin user ready")
		}
	}

	if info, err := os.Stat(docsDir); err == nil && info.IsDir() {
		router.Static("/docs", docsDir)
		logger.Info().Str("dir", docsDir).Msg("serving docs at /docs")
	} else {
		logger.Warn().Str("dir", docsDir).Msg("docs directory not found, /docs not served")
	}

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

func seedAdmin(pg *pgxpool.Pool, email, password string) error {
	ctx := context.Background()
	var count int
	pg.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE is_admin = true`).Scan(&count)
	if count > 0 {
		return nil
	}
	var userID int64
	err := pg.QueryRow(ctx, `SELECT id FROM users WHERE email = $1`, email).Scan(&userID)
	if err == nil {
		_, err := pg.Exec(ctx, `UPDATE users SET is_admin = true, trust_score = 10.0 WHERE id = $1`, userID)
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = pg.Exec(ctx,
		`INSERT INTO users (username, email, password_hash, is_admin, trust_score, onboarding_stage)
		 VALUES ($1, $2, $3, true, 10.0, 4)`,
		"admin", email, string(hash))
	return err
}
