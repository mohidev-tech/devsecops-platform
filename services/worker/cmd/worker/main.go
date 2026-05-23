package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mohidev-tech/devsecops-platform/services/worker/internal/drain"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		logger.Error("DATABASE_URL is required")
		os.Exit(1)
	}
	batch := envInt("BATCH_SIZE", 10)
	interval := time.Duration(envInt("POLL_SECONDS", 2)) * time.Second

	bootCtx, cancelBoot := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelBoot()
	pool, err := pgxpool.New(bootCtx, dsn)
	if err != nil {
		logger.Error("db connect failed", "err", err)
		os.Exit(1)
	}
	defer pool.Close()
	if err := pool.Ping(bootCtx); err != nil {
		logger.Error("db ping failed", "err", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	go func() { <-stop; cancel() }()

	logger.Info("worker starting", "batch", batch, "interval", interval)
	tick := time.NewTicker(interval)
	defer tick.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("worker stopped")
			return
		case <-tick.C:
			n, err := drain.Once(ctx, pool, batch)
			if err != nil {
				logger.Error("drain error", "err", err)
				continue
			}
			if n > 0 {
				logger.Info("drained", "n", n)
			}
		}
	}
}

func envInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
