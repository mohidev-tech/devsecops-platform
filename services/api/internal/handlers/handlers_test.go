package handlers

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mohidev-tech/devsecops-platform/services/api/internal/jobs"
)

func TestLiveOK(t *testing.T) {
	rr := httptest.NewRecorder()
	Live(rr, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("got %d", rr.Code)
	}
}

type okPinger struct{}

func (okPinger) PingCheck() error { return nil }

type badPinger struct{ err error }

func (b badPinger) PingCheck() error { return b.err }

func TestReady_OK_WhenPingerHealthy(t *testing.T) {
	rr := httptest.NewRecorder()
	Ready(okPinger{})(rr, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("got %d", rr.Code)
	}
}

func TestReady_503_WhenPingerFails(t *testing.T) {
	rr := httptest.NewRecorder()
	Ready(badPinger{err: io.EOF})(rr, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("got %d", rr.Code)
	}
}

func TestJobs_PostInsertsAndReturnsID(t *testing.T) {
	repo := jobs.NewFakeRepo()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/jobs", strings.NewReader(`{"x":1}`))
	Jobs(logger, repo)(rr, req)
	if rr.Code != http.StatusAccepted {
		t.Fatalf("got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), `"id"`) {
		t.Fatalf("missing id in %q", rr.Body.String())
	}
	if n, _ := repo.Count(req.Context(), jobs.StatusPending); n != 1 {
		t.Fatalf("expected 1 pending, got %d", n)
	}
}

func TestJobs_GetRejected(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	rr := httptest.NewRecorder()
	Jobs(logger, jobs.NewFakeRepo())(rr, httptest.NewRequest(http.MethodGet, "/api/v1/jobs", nil))
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("got %d", rr.Code)
	}
}

func TestMetrics_ExposesJobCounts(t *testing.T) {
	repo := jobs.NewFakeRepo()
	_, _ = repo.Enqueue(httptest.NewRequest(http.MethodPost, "/", nil).Context(), []byte("x"))
	rr := httptest.NewRecorder()
	Metrics(repo)(rr, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	body := rr.Body.String()
	for _, want := range []string{"api_requests_total", "api_jobs_pending 1", "api_jobs_done 0"} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q in %q", want, body)
		}
	}
}
