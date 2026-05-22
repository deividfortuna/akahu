package akahu

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"sync"
	"time"
)

// WebhookType is the broad event category.
type WebhookType string

const (
	WebhookTypeToken       WebhookType = "TOKEN"
	WebhookTypeAccount     WebhookType = "ACCOUNT"
	WebhookTypeTransaction WebhookType = "TRANSACTION"
	WebhookTypeTransfer    WebhookType = "TRANSFER"
	WebhookTypePayment     WebhookType = "PAYMENT"
)

// WebhookStatus is the delivery status of a webhook event.
type WebhookStatus string

const (
	WebhookStatusSent   WebhookStatus = "SENT"
	WebhookStatusFailed WebhookStatus = "FAILED"
	WebhookStatusRetry  WebhookStatus = "RETRY"
)

// Webhook describes an active webhook subscription.
type Webhook struct {
	ID           string      `json:"_id"`
	Type         WebhookType `json:"type"`
	State        string      `json:"state"`
	URL          string      `json:"url"`
	CreatedAt    string      `json:"created_at"`
	UpdatedAt    string      `json:"updated_at"`
	LastCalledAt string      `json:"last_called_at"`
}

// WebhookCreateParams is the body for subscribing to a webhook.
type WebhookCreateParams struct {
	WebhookType WebhookType `json:"webhook_type"`
	State       string      `json:"state,omitempty"`
}

// WebhookEventQuery filters /webhook-events results.
type WebhookEventQuery struct {
	Status WebhookStatus // required
	Start  string
	End    string
}

func (q *WebhookEventQuery) values() url.Values {
	if q == nil {
		return nil
	}
	v := url.Values{}
	if q.Status != "" {
		v.Set("status", string(q.Status))
	}
	if q.Start != "" {
		v.Set("start", q.Start)
	}
	if q.End != "" {
		v.Set("end", q.End)
	}
	if len(v) == 0 {
		return nil
	}
	return v
}

// ----- Webhook payload dispatch ---------------------------------------------

// WebhookPayload is the parsed body of a webhook event. Use a type switch on
// the concrete type to extract event-specific fields.
type WebhookPayload interface {
	WebhookType() WebhookType
	WebhookCode() string
	State() string
	isWebhookPayload()
}

// BasePayload is embedded in every concrete webhook payload.
type BasePayload struct {
	Type      WebhookType `json:"webhook_type"`
	Code      string      `json:"webhook_code"`
	StateName string      `json:"state"`
}

func (b BasePayload) WebhookType() WebhookType { return b.Type }
func (b BasePayload) WebhookCode() string       { return b.Code }
func (b BasePayload) State() string             { return b.StateName }
func (BasePayload) isWebhookPayload()            {}

// CancelledPayload is sent when a webhook subscription is cancelled.
type CancelledPayload struct{ BasePayload }

// TokenDeletePayload is sent when a user token is deleted.
type TokenDeletePayload struct {
	BasePayload
	ItemID string `json:"item_id"`
}

// AccountCreatePayload is sent when an account is linked.
type AccountCreatePayload struct {
	BasePayload
	ItemID string `json:"item_id"`
}

// AccountDeletePayload is sent when an account is unlinked.
type AccountDeletePayload struct {
	BasePayload
	ItemID string `json:"item_id"`
}

// AccountUpdatePayload is sent when an account's metadata is updated.
type AccountUpdatePayload struct {
	BasePayload
	ItemID        string   `json:"item_id"`
	UpdatedFields []string `json:"updated_fields"`
}

// AccountMigratePayload is sent when an account is migrated to official open
// banking.
type AccountMigratePayload struct {
	BasePayload
	PreviousItemID string `json:"previous_item_id"`
	NewItemID      string `json:"new_item_id"`
}

// TransactionUpdatePayload is sent when new transactions are available
// (webhook_code is INITIAL_UPDATE or DEFAULT_UPDATE).
type TransactionUpdatePayload struct {
	BasePayload
	ItemID             string   `json:"item_id"`
	NewTransactions    int      `json:"new_transactions"`
	NewTransactionIDs  []string `json:"new_transaction_ids"`
}

// TransactionDeletePayload is sent when transactions are removed.
type TransactionDeletePayload struct {
	BasePayload
	ItemID              string   `json:"item_id"`
	RemovedTransactions []string `json:"removed_transactions"`
}

// TransferReceivedPayload is sent when a transfer is received.
type TransferReceivedPayload struct {
	BasePayload
	ItemID     string `json:"item_id"`
	ReceivedAt string `json:"received_at"`
}

// TransferUpdatePayload is sent when a transfer's status changes.
type TransferUpdatePayload struct {
	BasePayload
	ItemID     string         `json:"item_id"`
	Status     TransferStatus `json:"status"`
	StatusText string         `json:"status_text,omitempty"`
}

// PaymentReceivedPayload is sent when a payment is received.
type PaymentReceivedPayload struct {
	BasePayload
	ItemID     string `json:"item_id"`
	ReceivedAt string `json:"received_at"`
}

// PaymentUpdatePayload is sent when a payment's status changes.
type PaymentUpdatePayload struct {
	BasePayload
	ItemID     string            `json:"item_id"`
	Status     PaymentStatus     `json:"status"`
	StatusText string            `json:"status_text,omitempty"`
	StatusCode PaymentStatusCode `json:"status_code,omitempty"`
}

// ParseWebhookPayload decodes a webhook request body into the appropriate
// concrete WebhookPayload type.
func ParseWebhookPayload(body []byte) (WebhookPayload, error) {
	var head BasePayload
	if err := json.Unmarshal(body, &head); err != nil {
		return nil, fmt.Errorf("akahu: parse webhook envelope: %w", err)
	}

	// WEBHOOK_CANCELLED applies to any webhook_type.
	if head.Code == "WEBHOOK_CANCELLED" {
		var p CancelledPayload
		if err := json.Unmarshal(body, &p); err != nil {
			return nil, err
		}
		return &p, nil
	}

	switch head.Type {
	case WebhookTypeToken:
		switch head.Code {
		case "DELETE":
			return decodePayload[TokenDeletePayload](body)
		}
	case WebhookTypeAccount:
		switch head.Code {
		case "CREATE":
			return decodePayload[AccountCreatePayload](body)
		case "DELETE":
			return decodePayload[AccountDeletePayload](body)
		case "UPDATE":
			return decodePayload[AccountUpdatePayload](body)
		case "MIGRATE":
			return decodePayload[AccountMigratePayload](body)
		}
	case WebhookTypeTransaction:
		switch head.Code {
		case "INITIAL_UPDATE", "DEFAULT_UPDATE":
			return decodePayload[TransactionUpdatePayload](body)
		case "DELETE":
			return decodePayload[TransactionDeletePayload](body)
		}
	case WebhookTypeTransfer:
		switch head.Code {
		case "RECEIVED":
			return decodePayload[TransferReceivedPayload](body)
		case "UPDATE":
			return decodePayload[TransferUpdatePayload](body)
		}
	case WebhookTypePayment:
		switch head.Code {
		case "RECEIVED":
			return decodePayload[PaymentReceivedPayload](body)
		case "UPDATE":
			return decodePayload[PaymentUpdatePayload](body)
		}
	}

	return nil, fmt.Errorf("akahu: unknown webhook payload type=%q code=%q", head.Type, head.Code)
}

func decodePayload[T any](body []byte) (WebhookPayload, error) {
	var p T
	if err := json.Unmarshal(body, &p); err != nil {
		return nil, err
	}
	// All concrete payload types have a *T method set that satisfies
	// WebhookPayload via embedded BasePayload.
	if wp, ok := any(&p).(WebhookPayload); ok {
		return wp, nil
	}
	return nil, fmt.Errorf("akahu: %T does not implement WebhookPayload", p)
}

// ----- WebhookEvent (history) ----------------------------------------------

// WebhookEvent is a past webhook event as returned by /webhook-events.
type WebhookEvent struct {
	ID           string         `json:"_id"`
	Hook         string         `json:"_hook"`
	Status       WebhookStatus  `json:"status"`
	Payload      WebhookPayload `json:"-"`
	CreatedAt    string         `json:"created_at"`
	UpdatedAt    string         `json:"updated_at"`
	LastFailedAt string         `json:"last_failed_at,omitempty"`
}

// UnmarshalJSON decodes WebhookEvent and dispatches Payload to the appropriate
// concrete type.
func (e *WebhookEvent) UnmarshalJSON(data []byte) error {
	type alias WebhookEvent
	var aux struct {
		Payload json.RawMessage `json:"payload"`
		alias
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	*e = WebhookEvent(aux.alias)
	if len(aux.Payload) > 0 {
		p, err := ParseWebhookPayload(aux.Payload)
		if err != nil {
			return err
		}
		e.Payload = p
	}
	return nil
}

// ----- Signing key cache ---------------------------------------------------

// SigningKeyCache is the contract for caching webhook signing keys across
// processes/restarts. Implementations should be safe for concurrent use.
type SigningKeyCache interface {
	// Get returns the cached value for key, or "" if not present.
	Get(ctx context.Context, key string) (string, error)
	// Set stores value under key.
	Set(ctx context.Context, key, value string) error
}

// MemoryKeyCache is the default in-memory SigningKeyCache used when callers
// pass a nil CacheConfig to ValidateWebhook.
type MemoryKeyCache struct {
	mu sync.RWMutex
	m  map[string]string
}

// NewMemoryKeyCache returns a fresh in-memory cache.
func NewMemoryKeyCache() *MemoryKeyCache {
	return &MemoryKeyCache{m: make(map[string]string)}
}

func (c *MemoryKeyCache) Get(_ context.Context, key string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.m[key], nil
}

func (c *MemoryKeyCache) Set(_ context.Context, key, value string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.m == nil {
		c.m = make(map[string]string)
	}
	c.m[key] = value
	return nil
}

// CacheConfig controls signing-key caching for ValidateWebhook.
type CacheConfig struct {
	// Cache is the cache implementation. Defaults to a per-service
	// in-memory cache when nil.
	Cache SigningKeyCache
	// Key is the cache key. Default "akahu__webhook_key".
	Key string
	// MaxAge is the max age of a cached key before it is re-fetched.
	// Default 24h.
	MaxAge time.Duration
}

// cachedKeyData is the JSON-serialised cache entry shared with the JS SDK.
type cachedKeyData struct {
	ID            int    `json:"id"`
	Key           string `json:"key"`
	LastRefreshed string `json:"lastRefreshed"`
}

// ----- WebhooksService ------------------------------------------------------

// WebhooksService provides access to /webhooks, /webhook-events and /keys.
type WebhooksService struct {
	baseService
	defaultCacheOnce sync.Once
	defaultCache     SigningKeyCache
}

// List returns active webhook subscriptions for the given user token.
//
// API: GET /webhooks
func (s *WebhooksService) List(ctx context.Context, token string) ([]Webhook, error) {
	var out []Webhook
	err := s.c.doRequest(ctx, apiRequest{
		method: "GET",
		path:   "/webhooks",
		auth:   tokenAuth{token: token},
	}, &out)
	return out, err
}

// Subscribe creates a new webhook subscription. Returns the new webhook id.
//
// API: POST /webhooks
func (s *WebhooksService) Subscribe(ctx context.Context, token string, p WebhookCreateParams, opts ...RequestOption) (string, error) {
	var id string
	err := s.c.doRequest(ctx, apiRequest{
		method:  "POST",
		path:    "/webhooks",
		auth:    tokenAuth{token: token},
		body:    p,
		options: opts,
	}, &id)
	return id, err
}

// Unsubscribe cancels an active webhook subscription.
//
// API: DELETE /webhooks/{id}
func (s *WebhooksService) Unsubscribe(ctx context.Context, token, webhookID string) error {
	return s.c.doRequest(ctx, apiRequest{
		method: "DELETE",
		path:   "/webhooks/" + webhookID,
		auth:   tokenAuth{token: token},
	}, nil)
}

// ListEvents returns past webhook events filtered by status and date range.
//
// API: GET /webhook-events
func (s *WebhooksService) ListEvents(ctx context.Context, q WebhookEventQuery) ([]WebhookEvent, error) {
	var out []WebhookEvent
	err := s.c.doRequest(ctx, apiRequest{
		method: "GET",
		path:   "/webhook-events",
		auth:   basicAuth{},
		query:  q.values(),
	}, &out)
	return out, err
}

// GetPublicKey returns the PEM-encoded public key for the given key id.
//
// API: GET /keys/{id}
func (s *WebhooksService) GetPublicKey(ctx context.Context, keyID int) (string, error) {
	var pemKey string
	err := s.c.doRequest(ctx, apiRequest{
		method: "GET",
		path:   "/keys/" + strconv.Itoa(keyID),
		auth:   basicAuth{},
	}, &pemKey)
	return pemKey, err
}

// ValidateWebhook verifies the signature on a webhook request body and
// returns the parsed payload. The signing public key is fetched from /keys
// (and cached). Returns *WebhookValidationError on validation failure.
//
// keyID and signature come from the X-Akahu-Signing-Key and X-Akahu-Signature
// request headers. body is the raw request body — it must not be re-serialised
// before passing here.
func (s *WebhooksService) ValidateWebhook(ctx context.Context, keyID int, signature string, body []byte, cfg *CacheConfig) (WebhookPayload, error) {
	c := s.resolveCacheConfig(cfg)

	pubPEM, err := s.fetchPublicKey(ctx, keyID, c)
	if err != nil {
		return nil, err
	}

	if err := verifyWebhookSignature(pubPEM, signature, body); err != nil {
		return nil, &WebhookValidationError{Reason: "signature verification failed", Inner: err}
	}

	return ParseWebhookPayload(body)
}

func (s *WebhooksService) resolveCacheConfig(cfg *CacheConfig) CacheConfig {
	out := CacheConfig{}
	if cfg != nil {
		out = *cfg
	}
	if out.Cache == nil {
		s.defaultCacheOnce.Do(func() { s.defaultCache = NewMemoryKeyCache() })
		out.Cache = s.defaultCache
	}
	if out.Key == "" {
		out.Key = "akahu__webhook_key"
	}
	if out.MaxAge <= 0 {
		out.MaxAge = 24 * time.Hour
	}
	return out
}

// fetchPublicKey returns the public key for keyID, consulting cfg.Cache first
// and falling back to a /keys lookup. Implements the JS rotation logic.
func (s *WebhooksService) fetchPublicKey(ctx context.Context, keyID int, cfg CacheConfig) (string, error) {
	cached, ok, err := s.lookupCachedKey(ctx, cfg)
	if err != nil {
		return "", err
	}
	if ok {
		switch {
		case keyID == cached.ID:
			return cached.Key, nil
		case keyID < cached.ID:
			return "", &WebhookValidationError{
				Reason: fmt.Sprintf("webhook signing key (id %d) has expired", keyID),
			}
		}
		// keyID > cached.ID: fall through and refetch the newer key.
	}

	pem, err := s.GetPublicKey(ctx, keyID)
	if err != nil {
		return "", err
	}
	fresh := cachedKeyData{
		ID:            keyID,
		Key:           pem,
		LastRefreshed: time.Now().UTC().Format(time.RFC3339),
	}
	if err := s.storeCachedKey(ctx, cfg, fresh); err != nil {
		return "", err
	}
	return pem, nil
}

func (s *WebhooksService) lookupCachedKey(ctx context.Context, cfg CacheConfig) (cachedKeyData, bool, error) {
	raw, err := cfg.Cache.Get(ctx, cfg.Key)
	if err != nil {
		return cachedKeyData{}, false, err
	}
	if raw == "" {
		return cachedKeyData{}, false, nil
	}
	var d cachedKeyData
	if err := json.Unmarshal([]byte(raw), &d); err != nil {
		// Treat invalid cache data as a miss.
		return cachedKeyData{}, false, nil
	}
	last, err := time.Parse(time.RFC3339, d.LastRefreshed)
	if err != nil {
		return cachedKeyData{}, false, nil
	}
	if time.Since(last) > cfg.MaxAge {
		return cachedKeyData{}, false, nil
	}
	return d, true, nil
}

func (s *WebhooksService) storeCachedKey(ctx context.Context, cfg CacheConfig, d cachedKeyData) error {
	buf, err := json.Marshal(d)
	if err != nil {
		return err
	}
	return cfg.Cache.Set(ctx, cfg.Key, string(buf))
}

// verifyWebhookSignature validates an RSA-SHA256 webhook signature.
func verifyWebhookSignature(pemKey, signatureB64 string, body []byte) error {
	block, _ := pem.Decode([]byte(pemKey))
	if block == nil {
		return errors.New("invalid PEM-encoded public key")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		// Some keys may be in PKCS1 format.
		if rsaPub, perr := x509.ParsePKCS1PublicKey(block.Bytes); perr == nil {
			pub = rsaPub
		} else {
			return fmt.Errorf("parse public key: %w", err)
		}
	}
	rsaKey, ok := pub.(*rsa.PublicKey)
	if !ok {
		return errors.New("public key is not RSA")
	}
	sig, err := base64.StdEncoding.DecodeString(signatureB64)
	if err != nil {
		return fmt.Errorf("decode signature: %w", err)
	}
	digest := sha256.Sum256(body)
	if err := rsa.VerifyPKCS1v15(rsaKey, crypto.SHA256, digest[:], sig); err != nil {
		return err
	}
	return nil
}
