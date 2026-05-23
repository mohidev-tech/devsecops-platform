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
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	addr := envOr("LISTEN_ADDR", ":8080")
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", handlers.Live)
	mux.HandleFunc("/readyz", handlers.Ready)
	mux.HandleFunc("/metrics", handlers.Metrics)
	mux.HandleFunc("/api/v1/jobs", handlers.Jobs(logger))

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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	logger.Info("api stopped")
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
