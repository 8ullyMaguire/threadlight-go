package worker

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Worker struct {
	pg     *pgxpool.Pool
	rdb    *redis.Client
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

func New(pg *pgxpool.Pool, rdb *redis.Client) *Worker {
	ctx, cancel := context.WithCancel(context.Background())
	return &Worker{
		pg:     pg,
		rdb:    rdb,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (w *Worker) Start() {
	log.Println("worker: starting background workers")

	w.wg.Add(4)
	go w.runTrendingWorker()
	go w.runDailyRewardWorker()
	go w.runBountyExpiryWorker()
	go w.runArchivalWorker()
}

func (w *Worker) Stop() {
	log.Println("worker: stopping background workers")
	w.cancel()
	w.wg.Wait()
}

func (w *Worker) runPeriodic(name string, interval time.Duration, fn func(context.Context)) {
	defer w.wg.Done()
	log.Printf("worker: starting %s (every %v)", name, interval)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	fn(w.ctx)

	for {
		select {
		case <-w.ctx.Done():
			log.Printf("worker: %s stopped", name)
			return
		case <-ticker.C:
			fn(w.ctx)
		}
	}
}

func (w *Worker) runTrendingWorker() {
	w.runPeriodic("trending", 15*time.Minute, w.calculateTrending)
}

func (w *Worker) runDailyRewardWorker() {
	w.runPeriodic("daily-reward", 1*time.Hour, w.distributeDailyRewards)
}

func (w *Worker) runBountyExpiryWorker() {
	w.runPeriodic("bounty-expiry", 30*time.Minute, w.expireBounties)
}

func (w *Worker) runArchivalWorker() {
	w.runPeriodic("archival", 1*time.Hour, w.archiveOldContent)
}
