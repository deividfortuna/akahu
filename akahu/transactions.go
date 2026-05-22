package akahu

import (
	"context"
	"net/url"
)

// TransactionType describes the kind of transaction reported by Akahu.
type TransactionType string

const (
	TransactionTypeCredit        TransactionType = "CREDIT"
	TransactionTypeDebit         TransactionType = "DEBIT"
	TransactionTypePayment       TransactionType = "PAYMENT"
	TransactionTypeTransfer      TransactionType = "TRANSFER"
	TransactionTypeStandingOrder TransactionType = "STANDING ORDER"
	TransactionTypeEFTPOS        TransactionType = "EFTPOS"
	TransactionTypeInterest      TransactionType = "INTEREST"
	TransactionTypeFee           TransactionType = "FEE"
	TransactionTypeCreditCard    TransactionType = "CREDIT CARD"
	TransactionTypeTax           TransactionType = "TAX"
	TransactionTypeDirectDebit   TransactionType = "DIRECT DEBIT"
	TransactionTypeDirectCredit  TransactionType = "DIRECT CREDIT"
	TransactionTypeATM           TransactionType = "ATM"
	TransactionTypeLoan          TransactionType = "LOAN"
)

// CurrencyConversion describes a transaction's foreign-currency conversion.
type CurrencyConversion struct {
	Rate     float64 `json:"rate"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

// TransactionMerchant is the enriched merchant associated with a transaction.
type TransactionMerchant struct {
	ID      string `json:"_id"`
	Name    string `json:"name"`
	Website string `json:"website,omitempty"`
}

// TransactionCategoryGroup is a single grouping applied to an enriched
// transaction's category.
type TransactionCategoryGroup struct {
	ID   string `json:"_id"`
	Name string `json:"name"`
}

// TransactionCategory is the enriched category nested on an enriched
// transaction.
type TransactionCategory struct {
	ID     string                              `json:"_id"`
	Name   string                              `json:"name"`
	Groups map[string]TransactionCategoryGroup `json:"groups"`
}

// TransactionMeta is enrichment metadata on a transaction.
type TransactionMeta struct {
	Particulars  string              `json:"particulars,omitempty"`
	Code         string              `json:"code,omitempty"`
	Reference    string              `json:"reference,omitempty"`
	OtherAccount string              `json:"other_account,omitempty"`
	Conversion   *CurrencyConversion `json:"conversion,omitempty"`
	Logo         string              `json:"logo,omitempty"`
	CardSuffix   string              `json:"card_suffix,omitempty"`
}

// Transaction represents a single transaction. Both raw and enriched fields
// are present; enriched fields (Merchant, Category, Meta) will be nil for
// raw transactions.
type Transaction struct {
	ID                string          `json:"_id"`
	User              string          `json:"_user"`
	Account           string          `json:"_account"`
	Connection        string          `json:"_connection"`
	Migrated          string          `json:"_migrated,omitempty"`
	MigratedAccount   string          `json:"_migrated_account,omitempty"`
	CreatedAt         string          `json:"created_at"`
	UpdatedAt         string          `json:"updated_at"`
	Date              string          `json:"date"`
	Hash              string          `json:"hash,omitempty"` // deprecated, use ID
	Description       string          `json:"description"`
	Amount            float64         `json:"amount"`
	Balance           *float64        `json:"balance,omitempty"`
	Type              TransactionType `json:"type"`

	// Enriched fields (nil on raw transactions).
	Merchant *TransactionMerchant `json:"merchant,omitempty"`
	Category *TransactionCategory `json:"category,omitempty"`
	Meta     *TransactionMeta     `json:"meta,omitempty"`
}

// PendingTransaction represents a pending (not yet posted) transaction.
type PendingTransaction struct {
	User       string          `json:"_user"`
	Account    string          `json:"_account"`
	Connection string          `json:"_connection"`
	UpdatedAt  string          `json:"updated_at"`
	Date       string          `json:"date"`

	Description string          `json:"description"`
	Amount      float64         `json:"amount"`
	Type        TransactionType `json:"type"`

	// Enriched-only field.
	Meta *TransactionMeta `json:"meta,omitempty"`
}

// TransactionQuery filters a transaction list response.
type TransactionQuery struct {
	// Start is an ISO 8601 date string. Defaults to 30 days ago.
	Start string
	// End is an ISO 8601 date string. Defaults to today.
	End string
	// Cursor is the pagination cursor returned by a previous page.
	Cursor string
}

func (q *TransactionQuery) values() url.Values {
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
	if q.Cursor != "" {
		v.Set("cursor", q.Cursor)
	}
	if len(v) == 0 {
		return nil
	}
	return v
}

// TransactionsService provides access to /transactions.
type TransactionsService struct{ baseService }

// List returns a paginated list of transactions for all accounts linked by
// the user.
//
// API: GET /transactions
func (s *TransactionsService) List(ctx context.Context, token string, q *TransactionQuery) (*Page[Transaction], error) {
	var out Page[Transaction]
	err := s.c.doRequest(ctx, apiRequest{
		method: "GET",
		path:   "/transactions",
		auth:   tokenAuth{token: token},
		query:  q.values(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ListPending returns all pending transactions across the user's accounts.
//
// API: GET /transactions/pending
func (s *TransactionsService) ListPending(ctx context.Context, token string) ([]PendingTransaction, error) {
	var out []PendingTransaction
	err := s.c.doRequest(ctx, apiRequest{
		method: "GET",
		path:   "/transactions/pending",
		auth:   tokenAuth{token: token},
	}, &out)
	return out, err
}

// Get returns a single transaction by id.
//
// API: GET /transactions/{id}
func (s *TransactionsService) Get(ctx context.Context, token, transactionID string) (*Transaction, error) {
	var out Transaction
	err := s.c.doRequest(ctx, apiRequest{
		method: "GET",
		path:   "/transactions/" + transactionID,
		auth:   tokenAuth{token: token},
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// GetMany retrieves multiple transactions by id. All ids must belong to the
// user associated with token.
//
// API: POST /transactions/ids
func (s *TransactionsService) GetMany(ctx context.Context, token string, transactionIDs []string, opts ...RequestOption) ([]Transaction, error) {
	var out []Transaction
	err := s.c.doRequest(ctx, apiRequest{
		method:  "POST",
		path:    "/transactions/ids",
		auth:    tokenAuth{token: token},
		body:    transactionIDs,
		options: opts,
	}, &out)
	return out, err
}
