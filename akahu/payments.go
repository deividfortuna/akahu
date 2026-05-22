package akahu

import (
	"context"
	"net/url"
)

// PaymentPeriodFrequency is the frequency for a periodic payment limit.
type PaymentPeriodFrequency string

const (
	PaymentPeriodFrequencyDaily       PaymentPeriodFrequency = "DAILY"
	PaymentPeriodFrequencyWeekly      PaymentPeriodFrequency = "WEEKLY"
	PaymentPeriodFrequencyFortnightly PaymentPeriodFrequency = "FORTNIGHTLY"
	PaymentPeriodFrequencyMonthly     PaymentPeriodFrequency = "MONTHLY"
	PaymentPeriodFrequencyAnnually    PaymentPeriodFrequency = "ANNUALLY"
)

// PaymentConsentPeriodicLimit caps the total payment amount over a period.
type PaymentConsentPeriodicLimit struct {
	Amount    float64                `json:"amount"`
	Frequency PaymentPeriodFrequency `json:"frequency"`
}

// PaymentConsentPayee is a payee permitted by a payment consent.
type PaymentConsentPayee struct {
	Name          string `json:"name"`
	AccountNumber string `json:"account_number"`
}

// PaymentStatus is the high-level status of a Payment.
type PaymentStatus string

const (
	PaymentStatusReady           PaymentStatus = "READY"
	PaymentStatusPendingApproval PaymentStatus = "PENDING_APPROVAL"
	PaymentStatusPaused          PaymentStatus = "PAUSED"
	PaymentStatusSent            PaymentStatus = "SENT"
	PaymentStatusDeclined        PaymentStatus = "DECLINED"
	PaymentStatusError           PaymentStatus = "ERROR"
	PaymentStatusCancelled       PaymentStatus = "CANCELLED"
)

// PaymentStatusCode is the granular failure code on a Payment.
type PaymentStatusCode string

// PaymentApprovalType describes who approves a payment.
type PaymentApprovalType string

const (
	PaymentApprovalBank PaymentApprovalType = "BANK"
	PaymentApprovalUser PaymentApprovalType = "USER"
)

// PaymentToAccount is the payee's bank account on a payment.
type PaymentToAccount struct {
	Name          string `json:"name"`
	AccountNumber string `json:"account_number"`
}

// PaymentSourceMeta is metadata that appears on the payer's statement.
type PaymentSourceMeta struct {
	Code      string `json:"code,omitempty"`
	Reference string `json:"reference,omitempty"`
}

// PaymentDestinationMeta is metadata that appears on the payee's statement.
type PaymentDestinationMeta struct {
	Particulars string `json:"particulars,omitempty"`
	Code        string `json:"code,omitempty"`
	Reference   string `json:"reference,omitempty"`
}

// PaymentMeta groups source and destination statement metadata.
type PaymentMeta struct {
	Source      PaymentSourceMeta      `json:"source"`
	Destination PaymentDestinationMeta `json:"destination"`
}

// PaymentTimelineEntry is a single state transition in a Payment.Timeline.
type PaymentTimelineEntry struct {
	Status PaymentStatus `json:"status"`
	Time   string        `json:"time"`
	ETA    string        `json:"eta,omitempty"`
}

// Payment is the result of initiating a payment.
type Payment struct {
	ID            string                 `json:"_id"`
	From          string                 `json:"from"`
	To            PaymentToAccount       `json:"to"`
	Amount        float64                `json:"amount"`
	Meta          PaymentMeta            `json:"meta"`
	SID           string                 `json:"sid"`
	Status        PaymentStatus          `json:"status"`
	StatusCode    PaymentStatusCode      `json:"status_code,omitempty"`
	StatusText    string                 `json:"status_text,omitempty"`
	ApprovalType  PaymentApprovalType    `json:"approval_type,omitempty"`
	Final         bool                   `json:"final"`
	Timeline      []PaymentTimelineEntry `json:"timeline"`
	CreatedAt     string                 `json:"created_at"`
	UpdatedAt     string                 `json:"updated_at"`
	ReceivedAt    string                 `json:"received_at,omitempty"`
}

// PaymentCreateParams is the body for initiating a new bank account payment.
type PaymentCreateParams struct {
	From   string                  `json:"from"`
	Amount float64                 `json:"amount"`
	To     PaymentToAccount        `json:"to"`
	Meta   *PaymentCreateMeta      `json:"meta,omitempty"`
}

// PaymentCreateMeta is the meta payload sent with PaymentsService.Create.
type PaymentCreateMeta struct {
	Source      *PaymentSourceMeta      `json:"source,omitempty"`
	Destination *PaymentDestinationMeta `json:"destination,omitempty"`
}

// IRDPaymentMeta is the required tax metadata for an IRD payment.
type IRDPaymentMeta struct {
	TaxNumber string `json:"tax_number"`
	TaxType   string `json:"tax_type"`
	TaxPeriod string `json:"tax_period,omitempty"`
}

// IRDPaymentCreateParams is the body for initiating an IRD tax payment.
type IRDPaymentCreateParams struct {
	From   string         `json:"from"`
	Amount float64        `json:"amount"`
	Meta   IRDPaymentMeta `json:"meta"`
}

// PaymentQuery filters a payment list response by date range.
type PaymentQuery struct {
	Start string
	End   string
}

func (q *PaymentQuery) values() url.Values {
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

// PaymentsService provides access to /payments.
type PaymentsService struct{ baseService }

// Get returns a single payment by id.
//
// API: GET /payments/{id}
func (s *PaymentsService) Get(ctx context.Context, token, paymentID string) (*Payment, error) {
	var out Payment
	err := s.c.doRequest(ctx, apiRequest{
		method: "GET",
		path:   "/payments/" + paymentID,
		auth:   tokenAuth{token: token},
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns payments made within the given date range (defaults to last
// 30 days).
//
// API: GET /payments
func (s *PaymentsService) List(ctx context.Context, token string, q *PaymentQuery) ([]Payment, error) {
	var out []Payment
	err := s.c.doRequest(ctx, apiRequest{
		method: "GET",
		path:   "/payments",
		auth:   tokenAuth{token: token},
		query:  q.values(),
	}, &out)
	return out, err
}

// Create initiates a payment to an external bank account.
//
// API: POST /payments
func (s *PaymentsService) Create(ctx context.Context, token string, p PaymentCreateParams, opts ...RequestOption) (*Payment, error) {
	var out Payment
	err := s.c.doRequest(ctx, apiRequest{
		method:  "POST",
		path:    "/payments",
		auth:    tokenAuth{token: token},
		body:    p,
		options: opts,
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateToIRD initiates a tax payment to the Inland Revenue Department.
//
// API: POST /payments/ird
func (s *PaymentsService) CreateToIRD(ctx context.Context, token string, p IRDPaymentCreateParams, opts ...RequestOption) (*Payment, error) {
	var out Payment
	err := s.c.doRequest(ctx, apiRequest{
		method:  "POST",
		path:    "/payments/ird",
		auth:    tokenAuth{token: token},
		body:    p,
		options: opts,
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Cancel cancels a payment with status PENDING_APPROVAL.
//
// API: PUT /payments/{id}
func (s *PaymentsService) Cancel(ctx context.Context, token, paymentID string) error {
	return s.c.doRequest(ctx, apiRequest{
		method: "PUT",
		path:   "/payments/" + paymentID,
		auth:   tokenAuth{token: token},
	}, nil)
}
