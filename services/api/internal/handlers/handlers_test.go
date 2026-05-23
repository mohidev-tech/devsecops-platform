package handlers

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLiveReady(t *testing.T) {
	for _, tc := range []struct {
		name string
		fn   http.HandlerFunc
	}{{"live", Live}, {"ready", Ready}} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			tc.fn(rr, httptest.NewRequest(http.MethodGet, "/", nil))
			if rr.Code != http.StatusOK {
				t.Fatalf("got %d", rr.Code)
			}
		})
	}
}

func TestJobsPostAccepted(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	rr := httptest.NewRecorder()
	Jobs(logger)(rr, httptest.NewRequest(http.MethodPost, "/api/v1/jobs", strings.NewReader("{}")))
	if rr.Code != http.StatusAccepted {
		t.Fatalf("got %d", rr.Code)
	}
}

func TestJobsGetRejected(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	rr := httptest.NewRecorder()
	Jobs(logger)(rr, httptest.NewRequest(http.MethodGet, "/api/v1/jobs", nil))
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("got %d", rr.Code)
	}
}

func TestMetricsExposed(t *testing.T) {
	rr := httptest.NewRecorder()
	Metrics(rr, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if !strings.Contains(rr.Body.String(), "api_requests_total") {
		t.Fatal("missing metric")
	}
}
