package akahu

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestBuildAuthorizationURL_Defaults(t *testing.T) {
	c, err := New(testAppToken)
	if err != nil {
		t.Fatal(err)
	}
	got := c.Auth.BuildAuthorizationURL(AuthURLParams{RedirectURI: "https://example.com/cb"})
	u, err := url.Parse(got)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if u.Host != "oauth.akahu.nz" {
		t.Errorf("host = %q", u.Host)
	}
	q := u.Query()
	if q.Get("response_type") != "code" {
		t.Errorf("response_type = %q", q.Get("response_type"))
	}
	if q.Get("scope") != "ENDURING_CONSENT" {
		t.Errorf("scope = %q", q.Get("scope"))
	}
	if q.Get("client_id") != testAppToken {
		t.Errorf("client_id = %q", q.Get("client_id"))
	}
	if q.Get("redirect_uri") != "https://example.com/cb" {
		t.Errorf("redirect_uri = %q", q.Get("redirect_uri"))
	}
}

func TestBuildAuthorizationURL_OptionalFields(t *testing.T) {
	c, _ := New(testAppToken, WithOAuthBaseURL("https://oauth.example/"))
	got := c.Auth.BuildAuthorizationURL(AuthURLParams{
		RedirectURI:  "https://example.com/cb",
		Scope:        "FOO",
		Email:        "a@b.com",
		Connection:   "conn_1",
		State:        "xyz",
		RedirectMode: "deep_link",
	})
	for _, want := range []string{"scope=FOO", "email=a%40b.com", "connection=conn_1", "state=xyz", "redirect_mode=deep_link"} {
		if !strings.Contains(got, want) {
			t.Errorf("URL missing %q: %s", want, got)
		}
	}
	if !strings.HasPrefix(got, "https://oauth.example") {
		t.Errorf("base URL not used: %s", got)
	}
}

func TestIdentitiesBuildAuthorizationURL_DefaultsToONEOFF(t *testing.T) {
	c, _ := New(testAppToken)
	got := c.Identities.BuildAuthorizationURL(AuthURLParams{RedirectURI: "https://example.com/cb"})
	if !strings.Contains(got, "scope=ONEOFF") {
		t.Errorf("expected scope=ONEOFF, got %s", got)
	}
}

func TestAuthExchange(t *testing.T) {
	body := `{"access_token":"user_token_xyz","token_type":"bearer","scope":"ENDURING_CONSENT"}`
	srv, rec := newTestServer(t, http.StatusOK, body)
	c := newTestClient(t, srv)

	tok, err := c.Auth.Exchange(context.Background(), "code123", "https://example.com/cb")
	if err != nil {
		t.Fatalf("Exchange: %v", err)
	}
	if rec.Method != "POST" || rec.Path != "/token" {
		t.Errorf("got %s %s", rec.Method, rec.Path)
	}
	if rec.Header.Get("Authorization") != "" {
		t.Errorf("Authorization header should be unset on /token, got %q", rec.Header.Get("Authorization"))
	}
	if tok.AccessToken != "user_token_xyz" {
		t.Errorf("AccessToken = %q", tok.AccessToken)
	}
}

func TestAuthExchange_RequiresAppSecret(t *testing.T) {
	srv, _ := newTestServer(t, http.StatusOK, `{}`)
	c, _ := New(testAppToken, WithBaseURL(srv.URL))
	_, err := c.Auth.Exchange(context.Background(), "c", "https://x")
	if err == nil {
		t.Fatal("expected error without app secret")
	}
}

func TestAuthRevoke(t *testing.T) {
	srv, rec := newTestServer(t, http.StatusOK, `{"success":true}`)
	c := newTestClient(t, srv)

	if err := c.Auth.Revoke(context.Background(), "user_token_x"); err != nil {
		t.Fatalf("Revoke: %v", err)
	}
	if rec.Method != "DELETE" || rec.Path != "/token" {
		t.Errorf("got %s %s", rec.Method, rec.Path)
	}
}
