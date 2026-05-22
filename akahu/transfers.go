package akahu

import (
	"context"
	"net/url"
)

// TransferStatus is the status of a Transfer.
type TransferStatus string

const (
	TransferStatusReady           TransferStatus = "READY"
	TransferStatusPendingApproval TransferStatus = "PENDING_APPROVAL"
	TransferStatusSent            TransferStatus = "SENT"
	TransferStatusDeclined        TransferStatus = "DECLINED"
	TransferStatusError           TransferStatus = "ERROR"
	TransferStatusPaused          TransferStatus = "PAUSED"
	TransferStatusCancelled       TransferStatus = "CANCELLED"
)

// TransferTimelineEntry is a state transition in Transfer.Timeline.
type TransferTimelineEntry struct {
	Status TransferStatus `json:"status"`
	Time   string         `json:"time"`
}

// Transfer is the result of initiating an inter-account transfer.
type Transfer struct {
	ID         string                  `json:"_id"`
	From       string                  `json:"from"`
	To         string                  `json:"to"`
	Amount     float64                 `json:"amount"`
	SID        string                  `json:"sid"`
	Status     TransferStatus          `json:"status"`
	StatusText string                  `json:"status_text,omitempty"`
	Final      bool                    `json:"final"`
	Timeline   []TransferTimelineEntry `json:"timeline"`
	CreatedAt  string                  `json:"created_at"`
	UpdatedAt  string                  `json:"updated_at"`
}

// TransferCreateParams is the body for initiating an inter-account transfer.
type TransferCreateParams struct {
	From   string  `json:"from"`
	To     string  `json:"to"`
	Amount float64 `json:"amount"`
}

// TransferQuery filters a transfer list response by date range.
type TransferQuery struct {
	Start string
	End   string
}

func (q *TransferQuery) values() url.Values {
	if q == nil {
		return nil
	}
	v := url.Values{}
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

// TransfersService provides access to /transfers.
type TransfersService struct{ baseService }

// Get returns a single transfer by id.
//
// API: GET /transfers/{id}
func (s *TransfersService) Get(ctx context.Context, token, transferID string) (*Transfer, error) {
	var out Transfer
	err := s.c.doRequest(ctx, apiRequest{
		method: "GET",
		path:   "/transfers/" + transferID,
		auth:   tokenAuth{token: token},
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns transfers made within the given date range.
//
// API: GET /transfers
func (s *TransfersService) List(ctx context.Context, token string, q *TransferQuery) ([]Transfer, error) {
	var out []Transfer
	err := s.c.doRequest(ctx, apiRequest{
		method: "GET",
		path:   "/transfers",
		auth:   tokenAuth{token: token},
		query:  q.values(),
	}, &out)
	return out, err
}

// Create initiates a transfer between two of the user's bank accounts.
//
// API: POST /transfers
func (s *TransfersService) Create(ctx context.Context, token string, t TransferCreateParams, opts ...RequestOption) (*Transfer, error) {
	var out Transfer
	err := s.c.doRequest(ctx, apiRequest{
		method:  "POST",
		path:    "/transfers",
		auth:    tokenAuth{token: token},
		body:    t,
		options: opts,
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}
