package akahu

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

// generateTestKey returns a 2048-bit RSA private key plus its PEM-encoded
// public key.
func generateTestKey(t *testing.T) (*rsa.PrivateKey, string) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	pubBytes, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		t.Fatalf("MarshalPKIXPublicKey: %v", err)
	}
	pemKey := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})
	return key, string(pemKey)
}

func signBody(t *testing.T, key *rsa.PrivateKey, body []byte) string {
	t.Helper()
	digest := sha256.Sum256(body)
	sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	return base64.StdEncoding.EncodeToString(sig)
}

func TestParseWebhookPayload_AllVariants(t *testing.T) {
	cases := []struct {
		name string
		body string
		want WebhookPayload
	}{
		{"cancelled", `{"webhook_type":"ACCOUNT","webhook_code":"WEBHOOK_CANCELLED","state":"s"}`, &CancelledPayload{}},
		{"token-delete", `{"webhook_type":"TOKEN","webhook_code":"DELETE","state":"s","item_id":"i"}`, &TokenDeletePayload{}},
		{"acc-create", `{"webhook_type":"ACCOUNT","webhook_code":"CREATE","state":"s","item_id":"i"}`, &AccountCreatePayload{}},
		{"acc-delete", `{"webhook_type":"ACCOUNT","webhook_code":"DELETE","state":"s","item_id":"i"}`, &AccountDeletePayload{}},
		{"acc-update", `{"webhook_type":"ACCOUNT","webhook_code":"UPDATE","state":"s","item_id":"i","updated_fields":["balance"]}`, &AccountUpdatePayload{}},
		{"acc-migrate", `{"webhook_type":"ACCOUNT","webhook_code":"MIGRATE","state":"s","previous_item_id":"p","new_item_id":"n"}`, &AccountMigratePayload{}},
		{"tx-initial", `{"webhook_type":"TRANSACTION","webhook_code":"INITIAL_UPDATE","state":"s","item_id":"i","new_transactions":1,"new_transaction_ids":["x"]}`, &TransactionUpdatePayload{}},
		{"tx-default", `{"webhook_type":"TRANSACTION","webhook_code":"DEFAULT_UPDATE","state":"s","item_id":"i","new_transactions":1,"new_transaction_ids":["x"]}`, &TransactionUpdatePayload{}},
		{"tx-delete", `{"webhook_type":"TRANSACTION","webhook_code":"DELETE","state":"s","item_id":"i","removed_transactions":["x"]}`, &TransactionDeletePayload{}},
		{"trf-recv", `{"webhook_type":"TRANSFER","webhook_code":"RECEIVED","state":"s","item_id":"i","received_at":"t"}`, &TransferReceivedPayload{}},
		{"trf-update", `{"webhook_type":"TRANSFER","webhook_code":"UPDATE","state":"s","item_id":"i","status":"SENT"}`, &TransferUpdatePayload{}},
		{"pmt-recv", `{"webhook_type":"PAYMENT","webhook_code":"RECEIVED","state":"s","item_id":"i","received_at":"t"}`, &PaymentReceivedPayload{}},
		{"pmt-update", `{"webhook_type":"PAYMENT","webhook_code":"UPDATE","state":"s","item_id":"i","status":"SENT"}`, &PaymentUpdatePayload{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseWebhookPayload([]byte(tc.body))
			if err != nil {
				t.Fatalf("ParseWebhookPayload: %v", err)
			}
			gotType := typeNameOf(got)
			wantType := typeNameOf(tc.want)
			if gotType != wantType {
				t.Errorf("got %s, want %s", gotType, wantType)
			}
		})
	}
}

func typeNameOf(v any) string {
	return strings.TrimPrefix(stringType(v), "*akahu.")
}

func stringType(v any) string {
	if v == nil {
		return "nil"
	}
	t := []rune("")
	_ = t
	return formatType(v)
}

// formatType is a tiny helper that returns a stable representation of v's
// dynamic type. We avoid reflect to keep imports minimal in tests.
func formatType(v any) string {
	switch v.(type) {
	case *CancelledPayload:
		return "*akahu.CancelledPayload"
	case *TokenDeletePayload:
		return "*akahu.TokenDeletePayload"
	case *AccountCreatePayload:
		return "*akahu.AccountCreatePayload"
	case *AccountDeletePayload:
		return "*akahu.AccountDeletePayload"
	case *AccountUpdatePayload:
		return "*akahu.AccountUpdatePayload"
	case *AccountMigratePayload:
		return "*akahu.AccountMigratePayload"
	case *TransactionUpdatePayload:
		return "*akahu.TransactionUpdatePayload"
	case *TransactionDeletePayload:
		return "*akahu.TransactionDeletePayload"
	case *TransferReceivedPayload:
		return "*akahu.TransferReceivedPayload"
	case *TransferUpdatePayload:
		return "*akahu.TransferUpdatePayload"
	case *PaymentReceivedPayload:
		return "*akahu.PaymentReceivedPayload"
	case *PaymentUpdatePayload:
		return "*akahu.PaymentUpdatePayload"
	}
	return "unknown"
}

func TestValidateWebhook_HappyPath(t *testing.T) {
	priv, pemKey := generateTestKey(t)
	const keyID = 7
	body := []byte(`{"webhook_type":"ACCOUNT","webhook_code":"UPDATE","state":"s","item_id":"acc_1","updated_fields":["balance"]}`)
	signature := signBody(t, priv, body)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/keys/"+strconv.Itoa(keyID) {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		// /keys returns the PEM under "item_id" (string payload shape).
		_, _ = w.Write([]byte(`{"success":true,"item_id":` + jsonString(pemKey) + `}`))
	}))
	defer srv.Close()
	c := newTestClient(t, srv)

	payload, err := c.Webhooks.ValidateWebhook(context.Background(), keyID, signature, body, nil)
	if err != nil {
		t.Fatalf("ValidateWebhook: %v", err)
	}
	if payload.WebhookCode() != "UPDATE" {
		t.Errorf("code = %q", payload.WebhookCode())
	}
	upd, ok := payload.(*AccountUpdatePayload)
	if !ok {
		t.Fatalf("type = %T", payload)
	}
	if upd.ItemID != "acc_1" {
		t.Errorf("item_id = %q", upd.ItemID)
	}
}

func TestValidateWebhook_BadSignature(t *testing.T) {
	_, pemKey := generateTestKey(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"item_id":` + jsonString(pemKey) + `}`))
	}))
	defer srv.Close()
	c := newTestClient(t, srv)

	body := []byte(`{"webhook_type":"ACCOUNT","webhook_code":"UPDATE","state":"s","item_id":"acc_1","updated_fields":[]}`)
	badSig := base64.StdEncoding.EncodeToString([]byte("not a real signature"))

	_, err := c.Webhooks.ValidateWebhook(context.Background(), 1, badSig, body, nil)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if _, ok := err.(*WebhookValidationError); !ok {
		t.Errorf("err type = %T", err)
	}
}

func TestValidateWebhook_ExpiredKey(t *testing.T) {
	priv, pemKey := generateTestKey(t)
	body := []byte(`{"webhook_type":"ACCOUNT","webhook_code":"UPDATE","state":"s","item_id":"x","updated_fields":[]}`)
	sig := signBody(t, priv, body)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"item_id":` + jsonString(pemKey) + `}`))
	}))
	defer srv.Close()
	c := newTestClient(t, srv)

	// Pre-populate cache with id=10 (newer than requested 5).
	cache := NewMemoryKeyCache()
	entry := cachedKeyData{ID: 10, Key: pemKey, LastRefreshed: nowRFC3339()}
	buf, _ := json.Marshal(entry)
	_ = cache.Set(context.Background(), "akahu__webhook_key", string(buf))

	cfg := &CacheConfig{Cache: cache}
	_, err := c.Webhooks.ValidateWebhook(context.Background(), 5, sig, body, cfg)
	if err == nil {
		t.Fatal("expected expired-key error")
	}
	if _, ok := err.(*WebhookValidationError); !ok {
		t.Errorf("err type = %T", err)
	}
	if !strings.Contains(err.Error(), "expired") {
		t.Errorf("err = %v", err)
	}
}

func nowRFC3339() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func jsonString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
