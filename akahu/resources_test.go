package akahu

import (
	"context"
	"net/http"
	"testing"
)

func TestUsersGet(t *testing.T) {
	srv, rec := newTestServer(t, http.StatusOK, `{"success":true,"item":{"_id":"u","email":"a@b","access_granted_at":"2026-01-01T00:00:00Z"}}`)
	c := newTestClient(t, srv)
	u, err := c.Users.Get(context.Background(), testUserToken)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if u.Email != "a@b" || rec.Path != "/me" {
		t.Errorf("u=%+v path=%s", u, rec.Path)
	}
}

func TestPartiesList(t *testing.T) {
	srv, rec := newTestServer(t, http.StatusOK, `{"success":true,"items":[{"_id":"party_1","_authorisation":"a","_connection":"c","_user":"u","type":"INDIVIDUAL"}]}`)
	c := newTestClient(t, srv)
	out, err := c.Parties.List(context.Background(), testUserToken)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(out) != 1 || rec.Path != "/parties" {
		t.Errorf("out=%v path=%s", out, rec.Path)
	}
}

func TestAuthorisationsRevoke(t *testing.T) {
	srv, rec := newTestServer(t, http.StatusOK, `{"success":true}`)
	c := newTestClient(t, srv)
	if err := c.Authorisations.Revoke(context.Background(), testUserToken, "auth_1"); err != nil {
		t.Fatalf("Revoke: %v", err)
	}
	if rec.Method != "DELETE" || rec.Path != "/authorisations/auth_1" {
		t.Errorf("got %s %s", rec.Method, rec.Path)
	}
}

func TestCategoriesGet(t *testing.T) {
	srv, rec := newTestServer(t, http.StatusOK, `{"success":true,"item":{"_id":"cat_1","name":"Food","groups":{}}}`)
	c := newTestClient(t, srv)
	out, err := c.Categories.Get(context.Background(), "cat_1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if out.Name != "Food" || rec.Path != "/categories/cat_1" {
		t.Errorf("out=%+v path=%s", out, rec.Path)
	}
}

func TestConnectionsList(t *testing.T) {
	srv, rec := newTestServer(t, http.StatusOK, `{"success":true,"items":[{"_id":"conn_1","name":"ANZ","logo":"x","connection_type":"classic","new_connections_enabled":true}]}`)
	c := newTestClient(t, srv)
	out, err := c.Connections.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(out) != 1 || rec.Path != "/connections" {
		t.Errorf("out=%v path=%s", out, rec.Path)
	}
}

func TestConnectionsRefresh(t *testing.T) {
	srv, rec := newTestServer(t, http.StatusOK, `{"success":true}`)
	c := newTestClient(t, srv)
	if err := c.Connections.Refresh(context.Background(), testUserToken, "conn_1"); err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if rec.Method != "POST" || rec.Path != "/refresh/conn_1" {
		t.Errorf("got %s %s", rec.Method, rec.Path)
	}
}

func TestIdentitiesGet(t *testing.T) {
	srv, rec := newTestServer(t, http.StatusOK, `{"success":true,"item":{"_id":"i","status":"COMPLETE","created_at":"","updated_at":"","expires_at":""}}`)
	c := newTestClient(t, srv)
	out, err := c.Identities.Get(context.Background(), "code_1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if out.Status != IdentityStatusComplete || rec.Path != "/identity/code_1" {
		t.Errorf("out=%+v path=%s", out, rec.Path)
	}
}

func TestWebhooksSubscribe_ItemID(t *testing.T) {
	srv, rec := newTestServer(t, http.StatusOK, `{"success":true,"item_id":"hook_42"}`)
	c := newTestClient(t, srv)
	id, err := c.Webhooks.Subscribe(context.Background(), testUserToken, WebhookCreateParams{WebhookType: WebhookTypeAccount})
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	if id != "hook_42" {
		t.Errorf("id = %q", id)
	}
	if rec.Method != "POST" {
		t.Errorf("method = %s", rec.Method)
	}
}

func TestWebhooksList(t *testing.T) {
	srv, rec := newTestServer(t, http.StatusOK, `{"success":true,"items":[{"_id":"wh_1","type":"ACCOUNT","state":"","url":"u","created_at":"","updated_at":"","last_called_at":""}]}`)
	c := newTestClient(t, srv)
	out, err := c.Webhooks.List(context.Background(), testUserToken)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(out) != 1 || rec.Path != "/webhooks" {
		t.Errorf("out=%v path=%s", out, rec.Path)
	}
}

func TestWebhooksListEvents(t *testing.T) {
	srv, rec := newTestServer(t, http.StatusOK, `{"success":true,"items":[]}`)
	c := newTestClient(t, srv)
	if _, err := c.Webhooks.ListEvents(context.Background(), WebhookEventQuery{Status: WebhookStatusSent}); err != nil {
		t.Fatalf("ListEvents: %v", err)
	}
	if rec.Path != "/webhook-events" {
		t.Errorf("path = %s", rec.Path)
	}
	if rec.Query == "" {
		t.Errorf("query empty")
	}
}
