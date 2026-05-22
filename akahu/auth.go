package akahu

import (
	"context"
	"net/url"
	"strings"
)

// AuthorizationToken is the response from an OAuth code exchange.
type AuthorizationToken struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"` // always "bearer"
	Scope       string `json:"scope"`
}

// AuthURLParams configures BuildAuthorizationURL.
type AuthURLParams struct {
	// RedirectURI is the URL the user is redirected to after granting or
	// denying access. Required, must match one of the app's configured
	// Redirect URIs.
	RedirectURI string
	// ResponseType defaults to "code".
	ResponseType string
	// Scope defaults to "ENDURING_CONSENT" for the standard auth flow.
	Scope string
	// State is an opaque string returned to your app on redirect.
	State string
	// Email pre-fills the user's email on the Akahu login page.
	Email string
	// Connection scopes the flow to a specific connection (institution).
	Connection string
	// RedirectMode set to "deep_link" if RedirectURI activates a native
	// mobile app (Universal Links / App Links).
	RedirectMode string

	// BaseURL overrides the default https://oauth.akahu.nz on a per-call
	// basis. Most callers should leave this empty and use WithOAuthBaseURL
	// when building the client instead.
	BaseURL string
	// Path appended to the base URL. Default is empty (i.e. root).
	Path string
}

// AuthService provides OAuth helpers.
type AuthService struct{ baseService }

// BuildAuthorizationURL constructs the OAuth authorization URL the user must
// visit. The returned URL never errors — invalid input simply produces a URL
// the server will reject.
//
// API: see https://developers.akahu.nz/docs/authorizing-with-oauth2
func (s *AuthService) BuildAuthorizationURL(p AuthURLParams) string {
	return s.buildURLWithDefaultScope(p, "ENDURING_CONSENT")
}

func (s *AuthService) buildURLWithDefaultScope(p AuthURLParams, defaultScope string) string {
	base := p.BaseURL
	if base == "" {
		base = s.c.cfg.oauthBaseURL
	}
	responseType := p.ResponseType
	if responseType == "" {
		responseType = "code"
	}
	scope := p.Scope
	if scope == "" {
		scope = defaultScope
	}

	q := url.Values{}
	q.Set("response_type", responseType)
	q.Set("redirect_uri", p.RedirectURI)
	q.Set("scope", scope)
	q.Set("client_id", s.c.cfg.appToken)
	if p.Email != "" {
		q.Set("email", p.Email)
	}
	if p.Connection != "" {
		q.Set("connection", p.Connection)
	}
	if p.State != "" {
		q.Set("state", p.State)
	}
	if p.RedirectMode != "" {
		q.Set("redirect_mode", p.RedirectMode)
	}

	u := strings.TrimRight(base, "/")
	if p.Path != "" {
		if !strings.HasPrefix(p.Path, "/") {
			u += "/"
		}
		u += p.Path
	}
	return u + "?" + q.Encode()
}

// Exchange swaps an OAuth authorization code for a long-lived user access
// token. Requires WithAppSecret to have been provided to New (the call sends
// client_id and client_secret in the body, so no Authorization header is set).
//
// API: POST /token
func (s *AuthService) Exchange(ctx context.Context, code, redirectURI string) (*AuthorizationToken, error) {
	return s.exchange(ctx, code, redirectURI, "authorization_code")
}

// ExchangeWithGrantType is a variant of Exchange that lets the caller specify
// a non-default grant_type. Most callers should use Exchange.
func (s *AuthService) ExchangeWithGrantType(ctx context.Context, code, redirectURI, grantType string) (*AuthorizationToken, error) {
	return s.exchange(ctx, code, redirectURI, grantType)
}

func (s *AuthService) exchange(ctx context.Context, code, redirectURI, grantType string) (*AuthorizationToken, error) {
	if s.c.cfg.appSecret == "" {
		return nil, &APIError{Message: "OAuth code exchange requires an app secret; configure with akahu.WithAppSecret"}
	}
	body := map[string]string{
		"code":          code,
		"redirect_uri":  redirectURI,
		"grant_type":    grantType,
		"client_id":     s.c.cfg.appToken,
		"client_secret": s.c.cfg.appSecret,
	}
	var out AuthorizationToken
	err := s.c.doRequest(ctx, apiRequest{
		method: "POST",
		path:   "/token",
		body:   body,
		auth:   noAuth{},
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Revoke invalidates a user access token.
//
// API: DELETE /token
func (s *AuthService) Revoke(ctx context.Context, token string) error {
	return s.c.doRequest(ctx, apiRequest{
		method: "DELETE",
		path:   "/token",
		auth:   tokenAuth{token: token},
	}, nil)
}
