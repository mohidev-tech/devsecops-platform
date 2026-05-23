package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"
)

var (
	requestsTotal atomic.Uint64
	jobsEnqueued  atomic.Uint64
)

func Live(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"live"}`))
}

func Ready(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ready"}`))
}

func Metrics(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	_, _ = w.Write([]byte(
		"# HELP api_requests_total Total HTTP requests handled.\n" +
			"# TYPE api_requests_total counter\n" +
			"api_requests_total " + u64(requestsTotal.Load()) + "\n" +
			"# HELP api_jobs_enqueued_total Total jobs enqueued.\n" +
			"# TYPE api_jobs_enqueued_total counter\n" +
			"api_jobs_enqueued_total " + u64(jobsEnqueued.Load()) + "\n",
	))
}

func Jobs(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		jobsEnqueued.Add(1)
		logger.Info("job enqueued", "request_id", r.Header.Get("X-Request-ID"))
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"status":"accepted"}`))
	}
}

func WithRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Request-ID") == "" {
			b := make([]byte, 8)
			_, _ = rand.Read(b)
			r.Header.Set("X-Request-ID", hex.EncodeToString(b))
		}
		w.Header().Set("X-Request-ID", r.Header.Get("X-Request-ID"))
		next.ServeHTTP(w, r)
	})
}

func WithLogging(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestsTotal.Add(1)
		next.ServeHTTP(w, r)
		logger.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"request_id", r.Header.Get("X-Request-ID"),
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

func u64(v uint64) string {
	if v == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	return string(buf[i:])
}
