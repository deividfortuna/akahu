package akahu

import (
	"context"
	"net/http"
	"testing"
)

func TestAccountsList(t *testing.T) {
	body := `{"success":true,"items":[{"_id":"acc_1","_authorisation":"auth_1","connection":{"_id":"conn","name":"ANZ","logo":"l","connection_type":"classic"},"name":"Cheque","status":"ACTIVE","type":"CHECKING","attributes":["TRANSACTIONS"]}]}`
	srv, rec := newTestServer(t, http.StatusOK, body)
	c := newTestClient(t, srv)

	accs, err := c.Accounts.List(context.Background(), testUserToken)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if rec.Method != "GET" || rec.Path != "/accounts" {
		t.Errorf("got %s %s", rec.Method, rec.Path)
	}
	if len(accs) != 1 {
		t.Fatalf("len(accs) = %d", len(accs))
	}
	if accs[0].ID != "acc_1" || accs[0].Type != AccountTypeChecking {
		t.Errorf("account fields wrong: %+v", accs[0])
	}
}

func TestAccountsGet(t *testing.T) {
	body := `{"success":true,"item":{"_id":"acc_1","_authorisation":"auth_1","connection":{"_id":"conn","name":"ANZ","logo":"l","connection_type":"classic"},"name":"Cheque","status":"ACTIVE","type":"CHECKING","attributes":["TRANSACTIONS"],"balance":{"currency":"NZD","current":100.5}}}`
	srv, rec := newTestServer(t, http.StatusOK, body)
	c := newTestClient(t, srv)

	acc, err := c.Accounts.Get(context.Background(), testUserToken, "acc_1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if rec.Path != "/accounts/acc_1" {
		t.Errorf("path = %q", rec.Path)
	}
	if acc.Balance == nil || acc.Balance.Current != 100.5 {
		t.Errorf("balance = %+v", acc.Balance)
	}
}

func TestAccountsRefresh(t *testing.T) {
	srv, rec := newTestServer(t, http.StatusOK, `{"success":true}`)
	c := newTestClient(t, srv)

	if err := c.Accounts.Refresh(context.Background(), testUserToken, "acc_1"); err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if rec.Method != "POST" || rec.Path != "/refresh/acc_1" {
		t.Errorf("got %s %s", rec.Method, rec.Path)
	}
}

func TestAccountsListTransactionsCursor(t *testing.T) {
	body := `{"success":true,"items":[{"_id":"tx1","_user":"u","_account":"a","_connection":"c","created_at":"","updated_at":"","date":"","description":"d","amount":1.0,"type":"DEBIT"}],"cursor":{"next":"abc123"}}`
	srv, _ := newTestServer(t, http.StatusOK, body)
	c := newTestClient(t, srv)

	page, err := c.Accounts.ListTransactions(context.Background(), testUserToken, "acc_1", &TransactionQuery{Start: "2026-01-01"})
	if err != nil {
		t.Fatalf("ListTransactions: %v", err)
	}
	if len(page.Items) != 1 {
		t.Fatalf("items = %d", len(page.Items))
	}
	if page.Cursor.Next == nil || *page.Cursor.Next != "abc123" {
		t.Errorf("cursor.next = %v", page.Cursor.Next)
	}
}
