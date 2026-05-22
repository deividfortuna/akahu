package akahu

import (
	"errors"
	"fmt"
)

// APIError is returned when the Akahu API responds with a non-2xx status or
// a payload of {"success": false, ...}.
type APIError struct {
	// StatusCode is the HTTP status code from the response.
	StatusCode int
	// Message is a human-readable error message extracted from the response
	// body (the `message` field for normal errors, or `error_description` /
	// a mapped value for OAuth errors).
	Message string
	// ErrorCode is the OAuth `error` field when present (e.g. "invalid_grant").
	// Empty for non-OAuth errors.
	ErrorCode string
	// Body is the raw response body, useful for debugging.
	Body []byte
}

func (e *APIError) Error() string {
	if e.ErrorCode != "" {
		return fmt.Sprintf("akahu: %s (%s, status %d)", e.Message, e.ErrorCode, e.StatusCode)
	}
	return fmt.Sprintf("akahu: %s (status %d)", e.Message, e.StatusCode)
}

// WebhookValidationError is returned when validation of a webhook signature
// fails, either because the signature does not match or because the requested
// signing key has been superseded.
type WebhookValidationError struct {
	Reason string
	Inner  error
}

func (e *WebhookValidationError) Error() string {
	if e.Inner != nil {
		return fmt.Sprintf("akahu webhook validation: %s: %v", e.Reason, e.Inner)
	}
	return "akahu webhook validation: " + e.Reason
}

func (e *WebhookValidationError) Unwrap() error { return e.Inner }

// IsAkahuError reports whether err originates from this SDK (either an
// *APIError or a *WebhookValidationError).
func IsAkahuError(err error) bool {
	var a *APIError
	var w *WebhookValidationError
	return errors.As(err, &a) || errors.As(err, &w)
}

// oAuthErrorMessages maps standard OAuth2 error codes to human-readable
// messages, matching the JS SDK behaviour.
var oAuthErrorMessages = map[string]string{
	"invalid_request":           "Invalid OAuth request.",
	"unauthorized_client":       "This application is not authorized to make this request.",
	"unsupported_response_type": "Unsupported OAuth response type.",
	"invalid_scope":             "Unknown or invalid scope.",
	"server_error":              "Unknown server error.",
	"temporarily_unavailable":   "The authorization server is temporarily unavailable.",
	"invalid_grant":             "Invalid OAuth request.",
}
