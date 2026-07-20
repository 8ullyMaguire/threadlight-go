package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type HealthHandler struct {
	pg  *pgxpool.Pool
	rdb *redis.Client
}

func NewHealthHandler(pg *pgxpool.Pool, rdb *redis.Client) *HealthHandler {
	return &HealthHandler{pg: pg, rdb: rdb}
}

func (h *HealthHandler) Health(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	status := http.StatusOK
	dbOk := true
	redisOk := true

	if err := h.pg.Ping(ctx); err != nil {
		dbOk = false
		status = http.StatusServiceUnavailable
	}

	if err := h.rdb.Ping(ctx).Err(); err != nil {
		redisOk = false
		status = http.StatusServiceUnavailable
	}

	c.JSON(status, gin.H{
		"status":  "ok",
		"postgres": dbOk,
		"redis":   redisOk,
	})
}
