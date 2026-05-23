package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mohidev-tech/devsecops-platform/services/api/internal/handlers"
	"github.com/mohidev-tech/devsecops-platform/services/api/internal/jobs"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		logger.Error("DATABASE_URL is required")
		os.Exit(1)
	}

	bootCtx, cancelBoot := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelBoot()
	repo, err := jobs.NewPostgresRepo(bootCtx, dsn)
	if err != nil {
		logger.Error("db init failed", "err", err)
		os.Exit(1)
	}
	defer repo.Close()

	addr := envOr("LISTEN_ADDR", ":8080")
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", handlers.Live)
	mux.Handle("/readyz", handlers.Ready(repoPinger{repo: repo}))
	mux.Handle("/metrics", handlers.Metrics(repo))
	mux.Handle("/api/v1/jobs", handlers.Jobs(logger, repo))

	srv := &http.Server{
		Addr:              addr,
		Handler:           handlers.WithRequestID(handlers.WithLogging(logger, mux)),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("api starting", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	logger.Info("api stopped")
}

type repoPinger struct{ repo *jobs.PostgresRepo }

func (p repoPinger) PingCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return p.repo.Pool().Ping(ctx)
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
