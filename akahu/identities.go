package akahu

import (
	"context"
	"encoding/json"
)

// IdentityStatus is the lifecycle status of an identity verification result.
type IdentityStatus string

const (
	IdentityStatusProcessing IdentityStatus = "PROCESSING"
	IdentityStatusComplete   IdentityStatus = "COMPLETE"
	IdentityStatusError      IdentityStatus = "ERROR"
)

// IdentityResult is the outcome of an Akahu identity verification request.
// Source/identities/addresses/accounts contain free-form vendor data — they
// are surfaced as json.RawMessage so callers can decode them as appropriate.
type IdentityResult struct {
	ID         string            `json:"_id"`
	Status     IdentityStatus    `json:"status"`
	CreatedAt  string            `json:"created_at"`
	UpdatedAt  string            `json:"updated_at"`
	ExpiresAt  string            `json:"expires_at"`
	Source     json.RawMessage   `json:"source,omitempty"`
	Errors     []string          `json:"errors,omitempty"`
	Identities []json.RawMessage `json:"identities,omitempty"`
	Addresses  []json.RawMessage `json:"addresses,omitempty"`
	Accounts   []json.RawMessage `json:"accounts,omitempty"`
}

// IdentityVerifyNameQuery is the body for VerifyName.
type IdentityVerifyNameQuery struct {
	GivenName  string `json:"given_name,omitempty"`
	MiddleName string `json:"middle_name,omitempty"`
	FamilyName string `json:"family_name"`
}

// NameVerificationFlags reports which name fields matched.
type NameVerificationFlags struct {
	GivenName    bool `json:"given_name"`
	GivenInitial bool `json:"given_initial"`
	MiddleName   bool `json:"middle_name"`
	MiddleInitial bool `json:"middle_initial"`
	FamilyName   bool `json:"family_name"`
}

// AccountHolderNameMeta is metadata for an account-holder name match.
type AccountHolderNameMeta struct {
	Name           string                  `json:"name"`
	Holder         string                  `json:"holder"`
	AccountNumber  string                  `json:"account_number"`
	Address        string                  `json:"address,omitempty"`
	Bank           string                  `json:"bank"`
	Branch         *AccountHolderBranch    `json:"branch,omitempty"`
}

// AccountHolderBranch is the branch detail nested in AccountHolderNameMeta.
type AccountHolderBranch struct {
	ID          string                       `json:"_id"`
	Description string                       `json:"description"`
	Phone       string                       `json:"phone"`
	Address     AccountHolderBranchAddress   `json:"address"`
}

// AccountHolderBranchAddress is the postal address of a branch.
type AccountHolderBranchAddress struct {
	Line1    string `json:"line1"`
	Line2    string `json:"line2,omitempty"`
	Line3    string `json:"line3,omitempty"`
	City     string `json:"city"`
	Country  string `json:"country"`
	Postcode string `json:"postcode"`
}

// PartyNameMeta is metadata for a party-name match.
type PartyNameMeta struct {
	Type       string   `json:"type"` // INDIVIDUAL, JOINT, TRUST, LLC
	Initials   []string `json:"initials,omitempty"`
	GivenName  string   `json:"given_name,omitempty"`
	MiddleName string   `json:"middle_name,omitempty"`
	FamilyName string   `json:"family_name"`
	FullName   string   `json:"full_name"`
	Prefix     string   `json:"prefix,omitempty"`
	Gender     string   `json:"gender,omitempty"`
}

// NameVerificationSourceType identifies which source provided a name match.
type NameVerificationSourceType string

const (
	NameVerificationSourceHolder NameVerificationSourceType = "HOLDER_NAME"
	NameVerificationSourceParty  NameVerificationSourceType = "PARTY_NAME"
)

// NameVerificationSource is one source-of-truth for a name verification. The
// concrete shape of Meta depends on Type:
//   - HOLDER_NAME -> use HolderMeta()
//   - PARTY_NAME  -> use PartyMeta()
type NameVerificationSource struct {
	Type         NameVerificationSourceType `json:"type"`
	MatchResult  string                     `json:"match_result"` // MATCH, PARTIAL_MATCH
	Meta         json.RawMessage            `json:"meta"`
	Verification NameVerificationFlags      `json:"verification"`
}

// HolderMeta decodes Meta as an AccountHolderNameMeta. Returns nil if Type is
// not HOLDER_NAME.
func (s *NameVerificationSource) HolderMeta() (*AccountHolderNameMeta, error) {
	if s.Type != NameVerificationSourceHolder {
		return nil, nil
	}
	var m AccountHolderNameMeta
	if err := json.Unmarshal(s.Meta, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// PartyMeta decodes Meta as a PartyNameMeta. Returns nil if Type is not
// PARTY_NAME.
func (s *NameVerificationSource) PartyMeta() (*PartyNameMeta, error) {
	if s.Type != NameVerificationSourceParty {
		return nil, nil
	}
	var m PartyNameMeta
	if err := json.Unmarshal(s.Meta, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// IdentityVerifyNameResult is the response from VerifyName.
type IdentityVerifyNameResult struct {
	Sources []NameVerificationSource `json:"sources"`
}

// IdentitiesService provides access to identity verification.
type IdentitiesService struct{ baseService }

// BuildAuthorizationURL constructs the OAuth URL for the identity-verification
// flow. Defaults to scope "ONEOFF".
func (s *IdentitiesService) BuildAuthorizationURL(p AuthURLParams) string {
	return s.c.Auth.buildURLWithDefaultScope(p, "ONEOFF")
}

// Get retrieves an identity result for the given OAuth code.
//
// API: GET /identity/{code}
func (s *IdentitiesService) Get(ctx context.Context, code string) (*IdentityResult, error) {
	var out IdentityResult
	err := s.c.doRequest(ctx, apiRequest{
		method: "GET",
		path:   "/identity/" + code,
		auth:   basicAuth{},
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// VerifyName matches a user-supplied name against an identity result (BETA).
//
// API: POST /identity/{code}/verify/name
func (s *IdentitiesService) VerifyName(ctx context.Context, code string, q IdentityVerifyNameQuery) (*IdentityVerifyNameResult, error) {
	var out IdentityVerifyNameResult
	err := s.c.doRequest(ctx, apiRequest{
		method: "POST",
		path:   "/identity/" + code + "/verify/name",
		body:   q,
		auth:   basicAuth{},
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}
