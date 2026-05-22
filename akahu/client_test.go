package akahu

import (
	"context"
	"encoding/base64"
	"net/http"
	"strings"
	"testing"
)

func TestNew_RejectsInvalidAppToken(t *testing.T) {
	if _, err := New("nope"); err == nil {
		t.Fatal("expected error for invalid app token, got nil")
	}
}

func TestNew_AppliesDefaults(t *testing.T) {
	c, err := New("app_token_x")
	if err != nil {
		t.Fatal(err)
	}
	if c.cfg.baseURL != defaultBaseURL {
		t.Errorf("baseURL = %q, want %q", c.cfg.baseURL, defaultBaseURL)
	}
	if c.cfg.oauthBaseURL != defaultOAuthBaseURL {
		t.Errorf("oauthBaseURL = %q, want %q", c.cfg.oauthBaseURL, defaultOAuthBaseURL)
	}
	if c.cfg.userAgent != userAgent {
		t.Errorf("userAgent = %q, want %q", c.cfg.userAgent, userAgent)
	}
}

func TestStandardHeadersAlwaysSet(t *testing.T) {
	srv, rec := newTestServer(t, http.StatusOK, `{"success":true}`)
	c := newTestClient(t, srv)

	if err := c.Accounts.RefreshAll(context.Background(), testUserToken); err != nil {
		t.Fatalf("RefreshAll: %v", err)
	}

	if got := rec.Header.Get("X-Akahu-Sdk"); got != userAgent {
		t.Errorf("X-Akahu-Sdk = %q, want %q", got, userAgent)
	}
	if got := rec.Header.Get("X-Akahu-Id"); got != testAppToken {
		t.Errorf("X-Akahu-Id = %q, want %q", got, testAppToken)
	}
	if got := rec.Header.Get("User-Agent"); got != userAgent {
		t.Errorf("User-Agent = %q, want %q", got, userAgent)
	}
}

func TestTokenAuthHeader(t *testing.T) {
	srv, rec := newTestServer(t, http.StatusOK, `{"success":true,"item":{"_id":"u","access_granted_at":"2026-01-01T00:00:00Z"}}`)
	c := newTestClient(t, srv)

	if _, err := c.Users.Get(context.Background(), "user_token_abc"); err != nil {
		t.Fatalf("Users.Get: %v", err)
	}

	if got, want := rec.Header.Get("Authorization"), "Bearer user_token_abc"; got != want {
		t.Errorf("Authorization = %q, want %q", got, want)
	}
}

func TestBasicAuthHeader(t *testing.T) {
	srv, rec := newTestServer(t, http.StatusOK, `{"success":true,"items":[]}`)
	c := newTestClient(t, srv)

	if _, err := c.Categories.List(context.Background()); err != nil {
		t.Fatalf("Categories.List: %v", err)
	}

	want := "Basic " + base64.StdEncoding.EncodeToString([]byte(testAppToken+":"+testAppSecret))
	if got := rec.Header.Get("Authorization"); got != want {
		t.Errorf("Authorization = %q, want %q", got, want)
	}
}

func TestBasicAuthRequiresAppSecret(t *testing.T) {
	srv, _ := newTestServer(t, http.StatusOK, `{"success":true,"items":[]}`)
	c, err := New(testAppToken, WithBaseURL(srv.URL))
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.Categories.List(context.Background())
	if err == nil || !strings.Contains(err.Error(), "app secret") {
		t.Fatalf("expected 'app secret' error, got %v", err)
	}
}

func TestIdempotencyKeyOnPOST(t *testing.T) {
	srv, rec := newTestServer(t, http.StatusOK, `{"success":true}`)
	c := newTestClient(t, srv)

	if err := c.Accounts.Refresh(context.Background(), testUserToken, "acc_1"); err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if got := rec.Header.Get("Idempotency-Key"); got == "" {
		t.Errorf("Idempotency-Key should be set on POST, got empty")
	}
}

func TestIdempotencyKeyOverride(t *testing.T) {
	srv, rec := newTestServer(t, http.StatusOK, `{"success":true,"items":[]}`)
	c := newTestClient(t, srv)

	_, err := c.Transactions.GetMany(context.Background(), testUserToken, []string{"tx1"}, WithIdempotencyKey("custom-key-123"))
	if err != nil {
		t.Fatalf("GetMany: %v", err)
	}
	if got := rec.Header.Get("Idempotency-Key"); got != "custom-key-123" {
		t.Errorf("Idempotency-Key = %q, want %q", got, "custom-key-123")
	}
}

func TestIdempotencyKeyAbsentOnGET(t *testing.T) {
	srv, rec := newTestServer(t, http.StatusOK, `{"success":true,"items":[]}`)
	c := newTestClient(t, srv)

	if _, err := c.Accounts.List(context.Background(), testUserToken); err != nil {
		t.Fatalf("List: %v", err)
	}
	if got := rec.Header.Get("Idempotency-Key"); got != "" {
		t.Errorf("Idempotency-Key should be absent on GET, got %q", got)
	}
}
