package akahu

import (
	"net/http"
	"time"
)

// ClientOption configures a Client during New.
type ClientOption func(*clientConfig) error

type clientConfig struct {
	appToken      string
	appSecret     string
	httpClient    *http.Client
	baseURL       string
	oauthBaseURL  string
	userAgent     string
	retries       int
	requestHeader http.Header
}

// WithAppSecret sets the Akahu app secret. This is required for endpoints that
// authenticate as the app itself (e.g. Connections.List, Categories, Webhook
// keys, OAuth code exchange). Do not use this option in client/browser code.
func WithAppSecret(secret string) ClientOption {
	return func(c *clientConfig) error {
		c.appSecret = secret
		return nil
	}
}

// WithHTTPClient overrides the underlying *http.Client. Use this to set a
// custom timeout, transport, or proxy:
//
//	akahu.WithHTTPClient(&http.Client{Timeout: 15 * time.Second})
func WithHTTPClient(hc *http.Client) ClientOption {
	return func(c *clientConfig) error {
		c.httpClient = hc
		return nil
	}
}

// WithBaseURL overrides the API base URL. The default is
// https://api.akahu.io/v1. Useful for staging/testing environments.
func WithBaseURL(url string) ClientOption {
	return func(c *clientConfig) error {
		c.baseURL = url
		return nil
	}
}

// WithOAuthBaseURL overrides the OAuth authorization endpoint used by
// Auth.BuildAuthorizationURL and Identities.BuildAuthorizationURL. The default
// is https://oauth.akahu.nz.
func WithOAuthBaseURL(url string) ClientOption {
	return func(c *clientConfig) error {
		c.oauthBaseURL = url
		return nil
	}
}

// WithUserAgent overrides the User-Agent header sent on every request. The
// default is "akahu-sdk-go/<version>".
func WithUserAgent(ua string) ClientOption {
	return func(c *clientConfig) error {
		c.userAgent = ua
		return nil
	}
}

// WithRetries configures the maximum number of retries on transient network
// failures. Retries are only performed for idempotent requests (GET, PUT,
// DELETE, or POST with an Idempotency-Key). Default 0.
func WithRetries(n int) ClientOption {
	return func(c *clientConfig) error {
		if n < 0 {
			n = 0
		}
		c.retries = n
		return nil
	}
}

// WithRequestHeaders adds extra headers that will be merged into every request.
func WithRequestHeaders(h http.Header) ClientOption {
	return func(c *clientConfig) error {
		c.requestHeader = h.Clone()
		return nil
	}
}

// defaultHTTPClient returns the http.Client used when WithHTTPClient is not
// provided.
func defaultHTTPClient() *http.Client {
	return &http.Client{Timeout: 30 * time.Second}
}
