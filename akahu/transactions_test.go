package akahu

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestTransactionsListCursorEnd(t *testing.T) {
	body := `{"success":true,"items":[],"cursor":{"next":null}}`
	srv, _ := newTestServer(t, http.StatusOK, body)
	c := newTestClient(t, srv)

	page, err := c.Transactions.List(context.Background(), testUserToken, nil)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if page.Cursor.Next != nil {
		t.Errorf("cursor.next should be nil at end, got %v", *page.Cursor.Next)
	}
}

func TestTransactionsGetMany_PostsBody(t *testing.T) {
	body := `{"success":true,"items":[{"_id":"tx1","_user":"u","_account":"a","_connection":"c","created_at":"","updated_at":"","date":"","description":"","amount":1,"type":"DEBIT"}]}`
	srv, rec := newTestServer(t, http.StatusOK, body)
	c := newTestClient(t, srv)

	out, err := c.Transactions.GetMany(context.Background(), testUserToken, []string{"tx1", "tx2"})
	if err != nil {
		t.Fatalf("GetMany: %v", err)
	}
	if rec.Method != "POST" || rec.Path != "/transactions/ids" {
		t.Errorf("got %s %s", rec.Method, rec.Path)
	}
	var ids []string
	if err := json.Unmarshal(rec.Body, &ids); err != nil {
		t.Fatalf("body decode: %v", err)
	}
	if len(ids) != 2 || ids[0] != "tx1" {
		t.Errorf("body ids = %v", ids)
	}
	if len(out) != 1 {
		t.Errorf("out len = %d", len(out))
	}
}

func TestNullCursorRejected(t *testing.T) {
	srv, _ := newTestServer(t, http.StatusOK, `{"success":true,"items":[],"cursor":{"next":null}}`)
	c := newTestClient(t, srv)

	_, err := c.Transactions.List(context.Background(), testUserToken, &TransactionQuery{Cursor: "null"})
	if err == nil {
		t.Fatal("expected error for cursor=\"null\", got nil")
	}
}
