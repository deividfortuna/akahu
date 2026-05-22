package akahu

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

// TestRetry_RecoversFromTransientFailure exercises the retry path by closing
// the connection on the first request, then succeeding.
func TestRetry_RecoversFromTransientFailure(t *testing.T) {
	var attempts int32
	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n == 1 {
			// Hijack and close to simulate a transient network error.
			hj, ok := w.(http.Hijacker)
			if !ok {
				t.Fatal("hijacker not available")
			}
			conn, _, err := hj.Hijack()
			if err != nil {
				t.Fatal(err)
			}
			conn.Close()
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"item":{"_id":"u","access_granted_at":"x"}}`))
	}))
	srv.Start()
	t.Cleanup(srv.Close)

	c := newTestClient(t, srv, WithRetries(3))
	if _, err := c.Users.Get(context.Background(), testUserToken); err != nil {
		t.Fatalf("Users.Get: %v", err)
	}
	if got := atomic.LoadInt32(&attempts); got != 2 {
		t.Errorf("attempts = %d, want 2", got)
	}
}

func TestIsRetryable_DNSNotFound(t *testing.T) {
	err := &net.DNSError{IsNotFound: true}
	if isRetryable(err) {
		t.Error("DNS NotFound should not be retryable")
	}
}

func TestIsRetryable_Generic(t *testing.T) {
	if !isRetryable(net.ErrClosed) {
		t.Error("net.ErrClosed should be retryable")
	}
}

func TestIsRetryable_ContextCancelled(t *testing.T) {
	if isRetryable(context.Canceled) {
		t.Error("context.Canceled should not be retryable")
	}
}
