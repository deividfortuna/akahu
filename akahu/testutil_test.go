package akahu

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	testAppToken  = "app_token_test"
	testAppSecret = "app_secret_test"
	testUserToken = "user_token_test"
)

// recordedRequest captures everything the server saw on a single request.
type recordedRequest struct {
	Method string
	Path   string
	Query  string
	Header http.Header
	Body   []byte
}

// newTestServer spins up an httptest.Server that records the incoming request
// and replies with the provided status code and body.
func newTestServer(t *testing.T, status int, respBody string) (*httptest.Server, *recordedRequest) {
	t.Helper()
	rec := &recordedRequest{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec.Method = r.Method
		rec.Path = r.URL.Path
		rec.Query = r.URL.RawQuery
		rec.Header = r.Header.Clone()
		buf := make([]byte, r.ContentLength)
		if r.ContentLength > 0 {
			_, _ = r.Body.Read(buf)
		}
		rec.Body = buf
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(respBody))
	}))
	t.Cleanup(srv.Close)
	return srv, rec
}

// newTestClient constructs a Client that talks to srv.
func newTestClient(t *testing.T, srv *httptest.Server, opts ...ClientOption) *Client {
	t.Helper()
	all := []ClientOption{
		WithBaseURL(srv.URL),
		WithAppSecret(testAppSecret),
	}
	all = append(all, opts...)
	c, err := New(testAppToken, all...)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c
}
