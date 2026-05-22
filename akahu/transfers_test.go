package akahu

import (
	"context"
	"net/http"
	"testing"
)

func TestTransfersCreate(t *testing.T) {
	body := `{"success":true,"item":{"_id":"trf_1","from":"a","to":"b","amount":1,"sid":"s","status":"READY","final":false,"timeline":[],"created_at":"","updated_at":""}}`
	srv, rec := newTestServer(t, http.StatusOK, body)
	c := newTestClient(t, srv)

	out, err := c.Transfers.Create(context.Background(), testUserToken, TransferCreateParams{From: "a", To: "b", Amount: 1})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if rec.Method != "POST" || rec.Path != "/transfers" {
		t.Errorf("got %s %s", rec.Method, rec.Path)
	}
	if out.ID != "trf_1" {
		t.Errorf("ID = %q", out.ID)
	}
}

func TestTransfersList(t *testing.T) {
	body := `{"success":true,"items":[]}`
	srv, rec := newTestServer(t, http.StatusOK, body)
	c := newTestClient(t, srv)

	if _, err := c.Transfers.List(context.Background(), testUserToken, &TransferQuery{Start: "2026-01-01", End: "2026-04-01"}); err != nil {
		t.Fatalf("List: %v", err)
	}
	if rec.Query == "" || rec.Query == "start=&end=" {
		t.Errorf("query missing or empty: %q", rec.Query)
	}
}
