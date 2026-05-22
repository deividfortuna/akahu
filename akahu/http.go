package akahu

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"syscall"

	"github.com/google/uuid"
)

// authMode is the auth strategy for an individual request.
type authMode interface{ isAuthMode() }

type noAuth struct{}

func (noAuth) isAuthMode() {}

// tokenAuth uses Bearer token authentication for user-scoped endpoints.
type tokenAuth struct{ token string }

func (tokenAuth) isAuthMode() {}

// basicAuth uses HTTP Basic with appToken:appSecret for app-scoped endpoints.
type basicAuth struct{}

func (basicAuth) isAuthMode() {}

// requestOptions are accumulated by per-call RequestOption funcs.
type requestOptions struct {
	idempotencyKey string
}

// RequestOption customises an individual API call. Currently only
// WithIdempotencyKey is supported.
type RequestOption func(*requestOptions)

// WithIdempotencyKey overrides the auto-generated Idempotency-Key on POST
// requests. Use this when retrying a request you've previously issued, so the
// API can deduplicate.
func WithIdempotencyKey(key string) RequestOption {
	return func(o *requestOptions) { o.idempotencyKey = key }
}

// apiRequest is the internal description of a single API call.
type apiRequest struct {
	method  string     // GET, POST, PUT, DELETE
	path    string     // absolute path under baseURL, leading "/" required
	query   url.Values // optional
	body    any        // optional; marshalled to JSON
	auth    authMode
	options []RequestOption
}

// envelope is the standard response envelope returned by the Akahu API.
// Different endpoints populate different fields; the doRequest helper
// inspects the envelope to find the payload.
type envelope struct {
	Success bool             `json:"success"`
	Message string           `json:"message"`
	Cursor  *Cursor          `json:"cursor"`
	Item    *json.RawMessage `json:"item"`
	Items   *json.RawMessage `json:"items"`
	ItemID  *string          `json:"item_id"`

	// OAuth error fields. These appear at the top level when the /token
	// endpoint returns an error.
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// doRequest performs the API call described by req and decodes the response
// into out. out must be a non-nil pointer matching the expected payload shape:
//
//   - *T or **T for "item" responses,
//   - *[]T for "items" responses,
//   - *Page[T] for paginated responses,
//   - *string for "item_id" responses,
//   - nil for void responses.
//
// For OAuth-style responses (no "success" wrapper, fields at the top level),
// the entire body is unmarshalled directly into out.
func (c *Client) doRequest(ctx context.Context, req apiRequest, out any) error {
	// Apply per-request options.
	var opts requestOptions
	for _, fn := range req.options {
		fn(&opts)
	}

	// Build full URL (baseURL is always set; baseURL + path).
	u, err := joinURL(c.cfg.baseURL, req.path)
	if err != nil {
		return fmt.Errorf("akahu: build URL: %w", err)
	}
	if len(req.query) > 0 {
		u.RawQuery = req.query.Encode()
	}

	// Marshal body, if any.
	var bodyBytes []byte
	if req.body != nil {
		bodyBytes, err = json.Marshal(req.body)
		if err != nil {
			return fmt.Errorf("akahu: marshal body: %w", err)
		}
	}

	// Build base headers (cloned so we don't mutate the client config).
	header := http.Header{}
	if c.cfg.requestHeader != nil {
		for k, vv := range c.cfg.requestHeader {
			for _, v := range vv {
				header.Add(k, v)
			}
		}
	}
	header.Set("X-Akahu-Sdk", userAgent)
	header.Set("X-Akahu-Id", c.cfg.appToken)
	header.Set("User-Agent", c.cfg.userAgent)
	header.Set("Accept", "application/json")
	if bodyBytes != nil {
		header.Set("Content-Type", "application/json")
	}

	// Apply auth.
	switch a := req.auth.(type) {
	case tokenAuth:
		header.Set("Authorization", "Bearer "+a.token)
	case basicAuth:
		if c.cfg.appSecret == "" {
			return errors.New("akahu: this endpoint requires an app secret; configure with akahu.WithAppSecret")
		}
		creds := c.cfg.appToken + ":" + c.cfg.appSecret
		header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(creds)))
	case noAuth, nil:
		// No auth header set.
	}

	// Idempotency-Key for POST requests.
	method := strings.ToUpper(req.method)
	if method == "POST" {
		key := opts.idempotencyKey
		if key == "" {
			key = uuid.NewString()
		}
		header.Set("Idempotency-Key", key)
	}

	// Sanity check on null cursors (matches JS SDK).
	if cur := req.query.Get("cursor"); cur == "null" {
		return errors.New("akahu: pagination cursor cannot be \"null\"; a null next cursor in an API response indicates that the final page has been reached")
	}

	// Execute with retry.
	resp, body, err := c.send(ctx, method, u.String(), bodyBytes, header)
	if err != nil {
		return err
	}

	return decodeResponse(resp.StatusCode, body, out)
}

// send performs the HTTP request, applying retry policy on transient network
// errors only (4xx/5xx are returned to the caller for normal handling).
func (c *Client) send(ctx context.Context, method, urlStr string, bodyBytes []byte, header http.Header) (*http.Response, []byte, error) {
	hasIdempotency := header.Get("Idempotency-Key") != ""
	canRetry := method != "POST" || hasIdempotency

	var lastErr error
	attempts := c.cfg.retries + 1
	for i := 0; i < attempts; i++ {
		// New body.Reader each attempt — request body is single-read.
		var bodyReader io.Reader
		if bodyBytes != nil {
			bodyReader = bytes.NewReader(bodyBytes)
		}
		httpReq, err := http.NewRequestWithContext(ctx, method, urlStr, bodyReader)
		if err != nil {
			return nil, nil, fmt.Errorf("akahu: build request: %w", err)
		}
		httpReq.Header = header

		resp, err := c.cfg.httpClient.Do(httpReq)
		if err != nil {
			lastErr = err
			if canRetry && i+1 < attempts && isRetryable(err) {
				continue
			}
			return nil, nil, fmt.Errorf("akahu: request failed: %w", err)
		}

		respBody, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return nil, nil, fmt.Errorf("akahu: read response body: %w", readErr)
		}
		return resp, respBody, nil
	}
	// Unreachable in practice; loop returns or continues.
	return nil, nil, lastErr
}

// joinURL appends path to base, handling slashes carefully.
func joinURL(base, path string) (*url.URL, error) {
	u, err := url.Parse(base)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	u.Path = strings.TrimRight(u.Path, "/") + path
	return u, nil
}

// decodeResponse processes the HTTP response body according to the standard
// Akahu envelope and unmarshals the appropriate payload into out.
func decodeResponse(statusCode int, body []byte, out any) error {
	// Empty body but non-2xx — return generic error.
	isError := statusCode < 200 || statusCode >= 300

	// Try to parse the envelope, but don't fail hard on parse error — fall
	// back to returning a generic APIError.
	var env envelope
	parseErr := json.Unmarshal(body, &env)

	// OAuth error response: {"error": "invalid_grant", "error_description": "..."}.
	if env.Error != "" {
		msg := env.ErrorDescription
		if msg == "" {
			if mapped, ok := oAuthErrorMessages[env.Error]; ok {
				msg = mapped
			} else {
				msg = env.Error
			}
		}
		return &APIError{
			StatusCode: statusCode,
			Message:    msg,
			ErrorCode:  env.Error,
			Body:       body,
		}
	}

	// HTTP-level error or success: false.
	if isError || (parseErr == nil && len(body) > 0 && hasSuccessField(body) && !env.Success) {
		msg := env.Message
		if msg == "" {
			msg = http.StatusText(statusCode)
			if msg == "" {
				msg = "request failed"
			}
		}
		return &APIError{
			StatusCode: statusCode,
			Message:    msg,
			Body:       body,
		}
	}

	// Successful response with no body, or void return.
	if out == nil {
		return nil
	}

	// Paginated response: cursor present.
	if env.Cursor != nil {
		// Construct {items, cursor} object for unmarshalling into Page[T].
		// We stitch together a JSON object containing only those two fields.
		var raw json.RawMessage
		if env.Items != nil {
			raw = *env.Items
		} else {
			raw = json.RawMessage("[]")
		}
		page := struct {
			Items  json.RawMessage `json:"items"`
			Cursor *Cursor         `json:"cursor"`
		}{Items: raw, Cursor: env.Cursor}
		buf, err := json.Marshal(page)
		if err != nil {
			return fmt.Errorf("akahu: re-encode page: %w", err)
		}
		if err := json.Unmarshal(buf, out); err != nil {
			return fmt.Errorf("akahu: decode page: %w", err)
		}
		return nil
	}

	// Single-item response.
	if env.Item != nil {
		if err := json.Unmarshal(*env.Item, out); err != nil {
			return fmt.Errorf("akahu: decode item: %w", err)
		}
		return nil
	}

	// Item-id response.
	if env.ItemID != nil {
		strOut, ok := out.(*string)
		if !ok {
			return fmt.Errorf("akahu: decode item_id: out is %T, want *string", out)
		}
		*strOut = *env.ItemID
		return nil
	}

	// Items list response.
	if env.Items != nil {
		if err := json.Unmarshal(*env.Items, out); err != nil {
			return fmt.Errorf("akahu: decode items: %w", err)
		}
		return nil
	}

	// OAuth-style spread response: no envelope wrapper. Decode the full body
	// into out, but only if there are non-envelope fields present.
	if parseErr == nil && len(body) > 0 && !hasSuccessField(body) {
		if err := json.Unmarshal(body, out); err != nil {
			return fmt.Errorf("akahu: decode body: %w", err)
		}
		return nil
	}

	// Successful response with no payload (e.g. {"success": true}). Nothing
	// to decode; leave out at its zero value.
	return nil
}

// hasSuccessField reports whether the JSON body contains a top-level
// "success" key. Used to distinguish wrapped envelopes from spread (OAuth)
// responses.
func hasSuccessField(body []byte) bool {
	var probe map[string]json.RawMessage
	if err := json.Unmarshal(body, &probe); err != nil {
		return false
	}
	_, ok := probe["success"]
	return ok
}

// isRetryable reports whether a network-level error is worth retrying.
// HTTP-level errors (4xx/5xx) never reach this function — those go through
// decodeResponse instead.
func isRetryable(err error) bool {
	if err == nil {
		return false
	}
	// TLS / x509 issues are permanent.
	var tlsErr *tls.CertificateVerificationError
	if errors.As(err, &tlsErr) {
		return false
	}
	var unknownAuthErr x509.UnknownAuthorityError
	if errors.As(err, &unknownAuthErr) {
		return false
	}
	var hostnameErr x509.HostnameError
	if errors.As(err, &hostnameErr) {
		return false
	}
	var certInvalidErr x509.CertificateInvalidError
	if errors.As(err, &certInvalidErr) {
		return false
	}
	// DNS NotFound is permanent (cf. JS deny-list ENOTFOUND).
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) && dnsErr.IsNotFound {
		return false
	}
	// ENETUNREACH is permanent (matches JS deny-list).
	if errors.Is(err, syscall.ENETUNREACH) {
		return false
	}
	// Context cancellation: don't retry.
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	// Anything else (including timeouts at the http.Client level, EOF,
	// connection reset) is treated as transient.
	return true
}

