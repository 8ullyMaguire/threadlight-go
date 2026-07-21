package worker

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Scheduler struct {
	pg     *pgxpool.Pool
	rdb    *redis.Client
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

func NewScheduler(pg *pgxpool.Pool, rdb *redis.Client) *Scheduler {
	return &Scheduler{
		pg:  pg,
		rdb: rdb,
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	s.ctx, s.cancel = context.WithCancel(ctx)
	log.Println("scheduler: starting background workers")

	workers := []struct {
		name     string
		interval time.Duration
		fn       func(context.Context)
	}{
		{"feed-rebuild", 5 * time.Minute, func(c context.Context) { RunFeedRebuild(c, s.pg) }},
		{"trust-decay", 1 * time.Hour, func(c context.Context) { RunTrustDecay(c, s.pg) }},
		{"trust-progression", 1 * time.Hour, func(c context.Context) {
			if err := RunTrustProgression(c, s.pg); err != nil {
				log.Printf("scheduler: trust-progression error: %v", err)
			}
		}},
		{"achievement-check", 30 * time.Minute, func(c context.Context) { RunAchievementCheck(c, s.pg) }},
		{"scheduled-post", 1 * time.Minute, func(c context.Context) { RunScheduledPostPublisher(c, s.pg) }},
		{"community-notes", 15 * time.Minute, func(c context.Context) { RunCommunityNotesProcessor(c, s.pg) }},
		{"active-stats", 5 * time.Minute, func(c context.Context) { RunActiveStatsUpdater(c, s.pg, s.rdb) }},
  {"prune", 1 * time.Hour, func(c context.Context) { RunPrune(c, s.pg) }},
  {"affinity-computation", 1 * time.Hour, func(c context.Context) {
  	if err := RunAffinityComputation(c, s.pg); err != nil {
  		log.Printf("scheduler: affinity-computation error: %v", err)
  	}
  }},
  {"algorithmic-list-refresh", 30 * time.Minute, func(c context.Context) {
  	if err := RunAlgorithmicListRefresh(c, s.pg); err != nil {
  		log.Printf("scheduler: algorithmic-list-refresh error: %v", err)
  	}
  }},
	}

	for _, w := range workers {
		s.wg.Add(1)
		go s.runLoop(w.name, w.interval, w.fn)
	}
}

func (s *Scheduler) Stop() {
	log.Println("scheduler: stopping background workers")
	s.cancel()
	s.wg.Wait()
}

func (s *Scheduler) runLoop(name string, interval time.Duration, fn func(context.Context)) {
	defer s.wg.Done()
	log.Printf("scheduler: starting %s (every %v)", name, interval)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	fn(s.ctx)

	for {
		select {
		case <-s.ctx.Done():
			log.Printf("scheduler: %s stopped", name)
			return
		case <-ticker.C:
			fn(s.ctx)
		}
	}
}
