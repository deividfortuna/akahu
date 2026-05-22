package akahu

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestTransactionsGet(t *testing.T) {
	srv, rec := newTestServer(t, http.StatusOK, `{"success":true,"item":{"_id":"tx1","_user":"u","_account":"a","_connection":"c","created_at":"","updated_at":"","date":"","description":"","amount":1,"type":"DEBIT"}}`)
	c := newTestClient(t, srv)
	tx, err := c.Transactions.Get(context.Background(), testUserToken, "tx1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if tx.ID != "tx1" || rec.Path != "/transactions/tx1" {
		t.Errorf("tx=%+v path=%s", tx, rec.Path)
	}
}

func TestTransactionsListPending(t *testing.T) {
	srv, rec := newTestServer(t, http.StatusOK, `{"success":true,"items":[]}`)
	c := newTestClient(t, srv)
	if _, err := c.Transactions.ListPending(context.Background(), testUserToken); err != nil {
		t.Fatalf("ListPending: %v", err)
	}
	if rec.Path != "/transactions/pending" {
		t.Errorf("path = %s", rec.Path)
	}
}

func TestTransfersGet(t *testing.T) {
	srv, _ := newTestServer(t, http.StatusOK, `{"success":true,"item":{"_id":"trf_1","from":"a","to":"b","amount":1,"sid":"","status":"SENT","final":true,"timeline":[],"created_at":"","updated_at":""}}`)
	c := newTestClient(t, srv)
	if _, err := c.Transfers.Get(context.Background(), testUserToken, "trf_1"); err != nil {
		t.Fatalf("Get: %v", err)
	}
}

func TestPaymentsGetAndList(t *testing.T) {
	srv, _ := newTestServer(t, http.StatusOK, `{"success":true,"item":{"_id":"pmt_1","from":"a","to":{"name":"x","account_number":"y"},"amount":1,"meta":{"source":{},"destination":{}},"sid":"","status":"SENT","final":true,"timeline":[],"created_at":"","updated_at":""}}`)
	c := newTestClient(t, srv)
	if _, err := c.Payments.Get(context.Background(), testUserToken, "pmt_1"); err != nil {
		t.Fatalf("Get: %v", err)
	}
}

func TestPaymentsList(t *testing.T) {
	srv, rec := newTestServer(t, http.StatusOK, `{"success":true,"items":[]}`)
	c := newTestClient(t, srv)
	if _, err := c.Payments.List(context.Background(), testUserToken, &PaymentQuery{Start: "2026-01-01"}); err != nil {
		t.Fatalf("List: %v", err)
	}
	if rec.Query == "" {
		t.Errorf("query empty")
	}
}

func TestAccountsListPendingAndRefreshAll(t *testing.T) {
	srv, rec := newTestServer(t, http.StatusOK, `{"success":true,"items":[]}`)
	c := newTestClient(t, srv)
	if _, err := c.Accounts.ListPendingTransactions(context.Background(), testUserToken, "acc_1"); err != nil {
		t.Fatalf("ListPendingTransactions: %v", err)
	}
	if rec.Path != "/accounts/acc_1/transactions/pending" {
		t.Errorf("path = %s", rec.Path)
	}

	srv2, rec2 := newTestServer(t, http.StatusOK, `{"success":true}`)
	c2 := newTestClient(t, srv2)
	if err := c2.Accounts.RefreshAll(context.Background(), testUserToken); err != nil {
		t.Fatalf("RefreshAll: %v", err)
	}
	if rec2.Path != "/refresh" {
		t.Errorf("path = %s", rec2.Path)
	}
}

func TestAccountsRevokeDeprecated(t *testing.T) {
	srv, rec := newTestServer(t, http.StatusOK, `{"success":true}`)
	c := newTestClient(t, srv)
	if err := c.Accounts.Revoke(context.Background(), testUserToken, "acc_1"); err != nil {
		t.Fatalf("Revoke: %v", err)
	}
	if rec.Method != "DELETE" {
		t.Errorf("method = %s", rec.Method)
	}
}

func TestConnectionsGet(t *testing.T) {
	srv, _ := newTestServer(t, http.StatusOK, `{"success":true,"item":{"_id":"conn_1","name":"ANZ","logo":"x","connection_type":"classic","new_connections_enabled":true}}`)
	c := newTestClient(t, srv)
	if _, err := c.Connections.Get(context.Background(), "conn_1"); err != nil {
		t.Fatalf("Get: %v", err)
	}
}

func TestCategoriesList(t *testing.T) {
	srv, _ := newTestServer(t, http.StatusOK, `{"success":true,"items":[]}`)
	c := newTestClient(t, srv)
	if _, err := c.Categories.List(context.Background()); err != nil {
		t.Fatalf("List: %v", err)
	}
}

func TestIdentitiesVerifyName(t *testing.T) {
	srv, rec := newTestServer(t, http.StatusOK, `{"success":true,"item":{"sources":[]}}`)
	c := newTestClient(t, srv)
	out, err := c.Identities.VerifyName(context.Background(), "code_1", IdentityVerifyNameQuery{FamilyName: "Doe"})
	if err != nil {
		t.Fatalf("VerifyName: %v", err)
	}
	if rec.Path != "/identity/code_1/verify/name" || rec.Method != "POST" {
		t.Errorf("got %s %s", rec.Method, rec.Path)
	}
	if out == nil {
		t.Errorf("nil result")
	}
}

func TestWebhooksUnsubscribe(t *testing.T) {
	srv, rec := newTestServer(t, http.StatusOK, `{"success":true}`)
	c := newTestClient(t, srv)
	if err := c.Webhooks.Unsubscribe(context.Background(), testUserToken, "wh_1"); err != nil {
		t.Fatalf("Unsubscribe: %v", err)
	}
	if rec.Method != "DELETE" || rec.Path != "/webhooks/wh_1" {
		t.Errorf("got %s %s", rec.Method, rec.Path)
	}
}

func TestWebhookEventUnmarshalJSON_DispatchesPayload(t *testing.T) {
	raw := `{"_id":"e1","_hook":"h1","status":"SENT","created_at":"","updated_at":"","payload":{"webhook_type":"ACCOUNT","webhook_code":"UPDATE","state":"s","item_id":"a","updated_fields":["x"]}}`
	var e WebhookEvent
	if err := json.Unmarshal([]byte(raw), &e); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if e.ID != "e1" || e.Status != WebhookStatusSent {
		t.Errorf("event header wrong: %+v", e)
	}
	if _, ok := e.Payload.(*AccountUpdatePayload); !ok {
		t.Errorf("payload type = %T", e.Payload)
	}
}

func TestWebhookPayloadInterface(t *testing.T) {
	p := &AccountUpdatePayload{BasePayload: BasePayload{Type: WebhookTypeAccount, Code: "UPDATE", StateName: "s"}}
	if p.WebhookType() != WebhookTypeAccount || p.WebhookCode() != "UPDATE" || p.State() != "s" {
		t.Errorf("interface methods wrong: %+v", p)
	}
	// Verify sealing: the unexported method must be present (non-callable
	// from outside, so we can't check directly — just compile-time satisfied).
	var _ WebhookPayload = p
}

func TestNameVerificationSource_HolderAndPartyMeta(t *testing.T) {
	src := NameVerificationSource{
		Type: NameVerificationSourceHolder,
		Meta: json.RawMessage(`{"name":"Joint","holder":"J Doe","account_number":"x","bank":"ANZ"}`),
	}
	hm, err := src.HolderMeta()
	if err != nil {
		t.Fatalf("HolderMeta: %v", err)
	}
	if hm.Holder != "J Doe" {
		t.Errorf("holder = %q", hm.Holder)
	}
	if pm, _ := src.PartyMeta(); pm != nil {
		t.Errorf("PartyMeta should be nil for HOLDER_NAME source")
	}

	src2 := NameVerificationSource{
		Type: NameVerificationSourceParty,
		Meta: json.RawMessage(`{"type":"INDIVIDUAL","family_name":"Doe","full_name":"J Doe"}`),
	}
	pm, err := src2.PartyMeta()
	if err != nil {
		t.Fatalf("PartyMeta: %v", err)
	}
	if pm.FamilyName != "Doe" {
		t.Errorf("family = %q", pm.FamilyName)
	}
}

func TestAccountMetaPreservesRaw(t *testing.T) {
	body := `{"holder":"Mr X","loan_details":null,"vendor_thing":42}`
	var m AccountMeta
	if err := json.Unmarshal([]byte(body), &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if m.Holder != "Mr X" {
		t.Errorf("holder = %q", m.Holder)
	}
	if len(m.Raw) == 0 {
		t.Error("raw not populated")
	}
}
