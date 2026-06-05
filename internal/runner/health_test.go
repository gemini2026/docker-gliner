package runner

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestWaitHealthyBecomesHealthy(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Fail the first two probes, then report healthy.
		if atomic.AddInt32(&hits, 1) < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := WaitHealthy(ctx, srv.URL, 10*time.Millisecond); err != nil {
		t.Fatalf("WaitHealthy: %v", err)
	}
	if atomic.LoadInt32(&hits) < 3 {
		t.Errorf("expected at least 3 probes, got %d", hits)
	}
}

func TestWaitHealthyTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	if err := WaitHealthy(ctx, srv.URL, 10*time.Millisecond); err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}
