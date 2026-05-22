package akahu

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func TestAPIError_StandardErrorBody(t *testing.T) {
	srv, _ := newTestServer(t, http.StatusBadRequest, `{"success":false,"message":"bad request: missing field"}`)
	c := newTestClient(t, srv)

	_, err := c.Users.Get(context.Background(), testUserToken)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error is not *APIError: %T", err)
	}
	if apiErr.StatusCode != http.StatusBadRequest {
		t.Errorf("StatusCode = %d", apiErr.StatusCode)
	}
	if apiErr.Message != "bad request: missing field" {
		t.Errorf("Message = %q", apiErr.Message)
	}
	if !IsAkahuError(err) {
		t.Error("IsAkahuError = false, want true")
	}
}

func TestAPIError_OAuthErrorWithDescription(t *testing.T) {
	srv, _ := newTestServer(t, http.StatusBadRequest, `{"error":"invalid_grant","error_description":"the auth code has expired"}`)
	c := newTestClient(t, srv)

	_, err := c.Auth.Exchange(context.Background(), "code", "https://x")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("not *APIError: %T", err)
	}
	if apiErr.ErrorCode != "invalid_grant" {
		t.Errorf("ErrorCode = %q", apiErr.ErrorCode)
	}
	if apiErr.Message != "the auth code has expired" {
		t.Errorf("Message = %q", apiErr.Message)
	}
}

func TestAPIError_OAuthErrorMappedMessage(t *testing.T) {
	srv, _ := newTestServer(t, http.StatusBadRequest, `{"error":"invalid_scope"}`)
	c := newTestClient(t, srv)

	_, err := c.Auth.Exchange(context.Background(), "code", "https://x")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("not *APIError")
	}
	if apiErr.Message != "Unknown or invalid scope." {
		t.Errorf("Message = %q (want mapped message)", apiErr.Message)
	}
}

func TestAPIError_PlainHTTPErrorNoBody(t *testing.T) {
	srv, _ := newTestServer(t, http.StatusInternalServerError, ``)
	c := newTestClient(t, srv)

	_, err := c.Users.Get(context.Background(), testUserToken)
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("not *APIError")
	}
	if apiErr.StatusCode != 500 {
		t.Errorf("StatusCode = %d", apiErr.StatusCode)
	}
}
