package akahu

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestPaymentsCreate(t *testing.T) {
	body := `{"success":true,"item":{"_id":"pmt_1","from":"acc_1","to":{"name":"Bob","account_number":"01-0000-0000000-00"},"amount":12.34,"meta":{"source":{},"destination":{}},"sid":"s","status":"READY","final":false,"timeline":[],"created_at":"","updated_at":""}}`
	srv, rec := newTestServer(t, http.StatusOK, body)
	c := newTestClient(t, srv)

	p := PaymentCreateParams{From: "acc_1", Amount: 12.34, To: PaymentToAccount{Name: "Bob", AccountNumber: "01-0000-0000000-00"}}
	out, err := c.Payments.Create(context.Background(), testUserToken, p)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if rec.Method != "POST" || rec.Path != "/payments" {
		t.Errorf("got %s %s", rec.Method, rec.Path)
	}
	var sent map[string]any
	_ = json.Unmarshal(rec.Body, &sent)
	if sent["from"] != "acc_1" {
		t.Errorf("body from = %v", sent["from"])
	}
	if out.Status != PaymentStatusReady {
		t.Errorf("status = %q", out.Status)
	}
}

func TestPaymentsCancel(t *testing.T) {
	srv, rec := newTestServer(t, http.StatusOK, `{"success":true}`)
	c := newTestClient(t, srv)

	if err := c.Payments.Cancel(context.Background(), testUserToken, "pmt_1"); err != nil {
		t.Fatalf("Cancel: %v", err)
	}
	if rec.Method != "PUT" || rec.Path != "/payments/pmt_1" {
		t.Errorf("got %s %s", rec.Method, rec.Path)
	}
}

func TestPaymentsCreateToIRD(t *testing.T) {
	body := `{"success":true,"item":{"_id":"pmt_2","from":"acc_1","to":{"name":"IRD","account_number":""},"amount":99,"meta":{"source":{},"destination":{}},"sid":"s","status":"READY","final":false,"timeline":[],"created_at":"","updated_at":""}}`
	srv, rec := newTestServer(t, http.StatusOK, body)
	c := newTestClient(t, srv)

	p := IRDPaymentCreateParams{From: "acc_1", Amount: 99, Meta: IRDPaymentMeta{TaxNumber: "123-456-789", TaxType: "INC"}}
	if _, err := c.Payments.CreateToIRD(context.Background(), testUserToken, p); err != nil {
		t.Fatalf("CreateToIRD: %v", err)
	}
	if rec.Path != "/payments/ird" {
		t.Errorf("path = %q", rec.Path)
	}
}
