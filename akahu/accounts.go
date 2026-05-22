package akahu

import (
	"context"
	"encoding/json"
)

// AccountType is the broad category of an account.
type AccountType string

const (
	AccountTypeChecking    AccountType = "CHECKING"
	AccountTypeSavings     AccountType = "SAVINGS"
	AccountTypeCreditCard  AccountType = "CREDITCARD"
	AccountTypeLoan        AccountType = "LOAN"
	AccountTypeKiwiSaver   AccountType = "KIWISAVER"
	AccountTypeInvestment  AccountType = "INVESTMENT"
	AccountTypeTermDeposit AccountType = "TERMDEPOSIT"
	AccountTypeForeign     AccountType = "FOREIGN"
	AccountTypeTax         AccountType = "TAX"
	AccountTypeRewards     AccountType = "REWARDS"
	AccountTypeWallet      AccountType = "WALLET"
)

// AccountAttribute describes a capability supported by an account.
type AccountAttribute string

const (
	AccountAttributePaymentTo    AccountAttribute = "PAYMENT_TO"
	AccountAttributePaymentFrom  AccountAttribute = "PAYMENT_FROM"
	AccountAttributeTransferTo   AccountAttribute = "TRANSFER_TO"
	AccountAttributeTransferFrom AccountAttribute = "TRANSFER_FROM"
	AccountAttributeTransactions AccountAttribute = "TRANSACTIONS"
)

// AccountStatus describes Akahu's current ability to refresh an account.
type AccountStatus string

const (
	AccountStatusActive   AccountStatus = "ACTIVE"
	AccountStatusInactive AccountStatus = "INACTIVE"
)

// AccountBalance represents the balance state of an account.
type AccountBalance struct {
	Currency  string   `json:"currency"`
	Current   float64  `json:"current"`
	Available *float64 `json:"available,omitempty"`
	Limit     *float64 `json:"limit,omitempty"`
	Overdrawn *bool    `json:"overdrawn,omitempty"`
}

// LoanInterest describes interest details on a loan account.
type LoanInterest struct {
	Rate      float64 `json:"rate"`
	Type      string  `json:"type"` // FIXED, FLOATING
	ExpiresAt string  `json:"expires_at,omitempty"`
}

// LoanTerm describes the term/duration of a loan.
type LoanTerm struct {
	Years  *int `json:"years,omitempty"`
	Months *int `json:"months,omitempty"`
}

// LoanRepayment describes a loan's repayment schedule.
type LoanRepayment struct {
	Frequency  string  `json:"frequency,omitempty"` // WEEKLY, FORTNIGHTLY, MONTHLY, QUARTERLY, BIANNUALLY, ANNUALLY
	NextDate   string  `json:"next_date,omitempty"`
	NextAmount float64 `json:"next_amount"`
}

// AccountLoanDetails captures loan-specific metadata for LOAN accounts.
type AccountLoanDetails struct {
	Purpose               string         `json:"purpose"`           // HOME, PERSONAL, BUSINESS, UNKNOWN
	Type                  string         `json:"type"`              // TABLE, REDUCING, REVOLVING, UNKNOWN
	Interest              LoanInterest   `json:"interest"`
	IsInterestOnly        bool           `json:"is_interest_only"`
	InterestOnlyExpiresAt string         `json:"interest_only_expires_at,omitempty"`
	Term                  *LoanTerm      `json:"term,omitempty"`
	MaturesAt             string         `json:"matures_at,omitempty"`
	InitialPrincipal      *float64       `json:"initial_principal,omitempty"`
	Repayment             *LoanRepayment `json:"repayment,omitempty"`
}

// AccountMeta is institution-supplied metadata. Known fields are surfaced
// directly; the full original JSON is preserved in Raw for callers that need
// vendor-specific extras.
type AccountMeta struct {
	Holder      string              `json:"holder,omitempty"`
	LoanDetails *AccountLoanDetails `json:"loan_details,omitempty"`
	// Raw is the original meta object as returned by the API.
	Raw json.RawMessage `json:"-"`
}

// UnmarshalJSON preserves the full meta payload in Raw while extracting known
// fields into the typed struct.
func (m *AccountMeta) UnmarshalJSON(data []byte) error {
	type alias AccountMeta
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*m = AccountMeta(a)
	m.Raw = append(json.RawMessage(nil), data...)
	return nil
}

// AccountRefreshState reports when each part of an account's data was last
// refreshed by Akahu.
type AccountRefreshState struct {
	Balance      string `json:"balance,omitempty"`
	Meta         string `json:"meta,omitempty"`
	Transactions string `json:"transactions,omitempty"`
	Party        string `json:"party,omitempty"`
}

// AccountPaymentConsent describes a consent that allows initiating payments
// from an account under defined limits and to a defined set of payees.
type AccountPaymentConsent struct {
	SingleLimit    float64                       `json:"single_limit"`
	PeriodicLimit  PaymentConsentPeriodicLimit   `json:"periodic_limit"`
	Payees         []PaymentConsentPayee         `json:"payees"`
}

// Account is a financial account linked to Akahu.
type Account struct {
	ID               string                  `json:"_id"`
	Migrated         string                  `json:"_migrated,omitempty"`
	Authorisation    string                  `json:"_authorisation"`
	Credentials      string                  `json:"_credentials,omitempty"` // deprecated, use Authorisation
	Connection       ConnectionInfo          `json:"connection"`
	Name             string                  `json:"name"`
	Status           AccountStatus           `json:"status"`
	FormattedAccount string                  `json:"formatted_account,omitempty"`
	Type             AccountType             `json:"type"`
	Attributes       []AccountAttribute      `json:"attributes"`
	Balance          *AccountBalance         `json:"balance,omitempty"`
	PaymentConsents  []AccountPaymentConsent `json:"payment_consents,omitempty"`
	Refreshed        *AccountRefreshState    `json:"refreshed,omitempty"`
	Meta             *AccountMeta            `json:"meta,omitempty"`
}

// AccountsService provides access to /accounts and /refresh.
type AccountsService struct{ baseService }

// List returns all accounts linked by the user associated with the given token.
//
// API: GET /accounts
func (s *AccountsService) List(ctx context.Context, token string) ([]Account, error) {
	var out []Account
	err := s.c.doRequest(ctx, apiRequest{
		method: "GET",
		path:   "/accounts",
		auth:   tokenAuth{token: token},
	}, &out)
	return out, err
}

// Get returns a single account by id.
//
// API: GET /accounts/{id}
func (s *AccountsService) Get(ctx context.Context, token, accountID string) (*Account, error) {
	var out Account
	err := s.c.doRequest(ctx, apiRequest{
		method: "GET",
		path:   "/accounts/" + accountID,
		auth:   tokenAuth{token: token},
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ListTransactions returns a paginated list of transactions for the given
// account.
//
// API: GET /accounts/{id}/transactions
func (s *AccountsService) ListTransactions(ctx context.Context, token, accountID string, q *TransactionQuery) (*Page[Transaction], error) {
	var out Page[Transaction]
	err := s.c.doRequest(ctx, apiRequest{
		method: "GET",
		path:   "/accounts/" + accountID + "/transactions",
		auth:   tokenAuth{token: token},
		query:  q.values(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ListPendingTransactions returns pending transactions for the given account.
//
// API: GET /accounts/{id}/transactions/pending
func (s *AccountsService) ListPendingTransactions(ctx context.Context, token, accountID string) ([]PendingTransaction, error) {
	var out []PendingTransaction
	err := s.c.doRequest(ctx, apiRequest{
		method: "GET",
		path:   "/accounts/" + accountID + "/transactions/pending",
		auth:   tokenAuth{token: token},
	}, &out)
	return out, err
}

// Revoke removes a single account from the given user token.
//
// API: DELETE /accounts/{id}
//
// Deprecated: use AuthorisationsService.Revoke instead.
func (s *AccountsService) Revoke(ctx context.Context, token, accountID string) error {
	return s.c.doRequest(ctx, apiRequest{
		method: "DELETE",
		path:   "/accounts/" + accountID,
		auth:   tokenAuth{token: token},
	}, nil)
}

// Refresh refreshes a single account.
//
// API: POST /refresh/{account_id}
func (s *AccountsService) Refresh(ctx context.Context, token, accountID string) error {
	return s.c.doRequest(ctx, apiRequest{
		method: "POST",
		path:   "/refresh/" + accountID,
		auth:   tokenAuth{token: token},
	}, nil)
}

// RefreshAll refreshes every account linked by the user.
//
// API: POST /refresh
func (s *AccountsService) RefreshAll(ctx context.Context, token string) error {
	return s.c.doRequest(ctx, apiRequest{
		method: "POST",
		path:   "/refresh",
		auth:   tokenAuth{token: token},
	}, nil)
}

