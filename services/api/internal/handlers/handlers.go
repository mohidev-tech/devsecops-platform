package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/mohidev-tech/devsecops-platform/services/api/internal/jobs"
)

var requestsTotal atomic.Uint64

func Live(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"live"}`))
}

// Ready depends on repo health. Pass a Pinger that errors when DB is unreachable.
type Pinger interface {
	PingCheck() error
}

func Ready(p Pinger) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		if p != nil {
			if err := p.PingCheck(); err != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				_, _ = w.Write([]byte(`{"status":"not ready"}`))
				return
			}
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ready"}`))
	}
}

func Metrics(repo jobs.Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pending, _ := repo.Count(r.Context(), jobs.StatusPending)
		done, _ := repo.Count(r.Context(), jobs.StatusDone)
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		_, _ = w.Write([]byte(
			"# HELP api_requests_total Total HTTP requests handled.\n" +
				"# TYPE api_requests_total counter\n" +
				"api_requests_total " + itoa(requestsTotal.Load()) + "\n" +
				"# HELP api_jobs_pending Number of jobs awaiting processing.\n" +
				"# TYPE api_jobs_pending gauge\n" +
				"api_jobs_pending " + itoa(uint64(pending)) + "\n" +
				"# HELP api_jobs_done Number of jobs completed.\n" +
				"# TYPE api_jobs_done gauge\n" +
				"api_jobs_done " + itoa(uint64(done)) + "\n",
		))
	}
}

const maxJobBody = 64 * 1024

func Jobs(logger *slog.Logger, repo jobs.Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		body, err := io.ReadAll(io.LimitReader(r.Body, maxJobBody))
		if err != nil {
			http.Error(w, "read body", http.StatusBadRequest)
			return
		}
		id, err := repo.Enqueue(r.Context(), body)
		if err != nil {
			logger.Error("enqueue failed", "err", err, "request_id", r.Header.Get("X-Request-ID"))
			http.Error(w, "enqueue failed", http.StatusInternalServerError)
			return
		}
		logger.Info("job enqueued", "id", id, "request_id", r.Header.Get("X-Request-ID"))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(map[string]string{"id": id, "status": "accepted"})
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

func itoa(v uint64) string {
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
