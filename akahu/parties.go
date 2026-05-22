package akahu

import "context"

// PartyType describes the legal nature of a party.
type PartyType string

const (
	PartyTypeIndividual PartyType = "INDIVIDUAL"
	PartyTypeJoint      PartyType = "JOINT"
	PartyTypeTrust      PartyType = "TRUST"
	PartyTypeLLC        PartyType = "LLC"
)

// PartyName is the user's name as sourced from the connected institution.
type PartyName struct {
	Value string `json:"value"`
}

// PartyDob is the user's date of birth (YYYY-MM-DD) as sourced from the
// connected institution.
type PartyDob struct {
	Value string `json:"value"`
}

// PartyTaxNumber is the user's IRD number (XXX-XXX-XXX).
type PartyTaxNumber struct {
	Value string `json:"value"`
}

// PartyPhoneNumber is a phone number from the connected institution.
type PartyPhoneNumber struct {
	Subtype  string `json:"subtype"` // MOBILE, HOME, WORK
	Verified bool   `json:"verified"`
	Value    string `json:"value"`
}

// PartyEmail is an email address from the connected institution.
type PartyEmail struct {
	Subtype  string `json:"subtype"` // PRIMARY
	Verified bool   `json:"verified"`
	Value    string `json:"value"`
}

// PartyAddressComponents are the normalised parts of a PartyAddress.
type PartyAddressComponents struct {
	Street     string `json:"street"`
	Suburb     string `json:"suburb"`
	City       string `json:"city"`
	Region     string `json:"region"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

// PartyAddress is a postal/residential address from the connected institution.
type PartyAddress struct {
	Subtype           string                 `json:"subtype"` // RESIDENTIAL, POSTAL
	Value             string                 `json:"value"`
	Formatted         string                 `json:"formatted"`
	Components        PartyAddressComponents `json:"components"`
	GoogleMapsPlaceID string                 `json:"google_maps_place_id"`
}

// Party is profile data sourced from the connected financial institution.
type Party struct {
	ID            string             `json:"_id"`
	Authorisation string             `json:"_authorisation"`
	Connection    string             `json:"_connection"`
	User          string             `json:"_user"`
	Type          PartyType          `json:"type"`
	Name          *PartyName         `json:"name,omitempty"`
	Dob           *PartyDob          `json:"dob,omitempty"`
	TaxNumber     *PartyTaxNumber    `json:"tax_number,omitempty"`
	PhoneNumbers  []PartyPhoneNumber `json:"phone_numbers,omitempty"`
	EmailAddresses []PartyEmail      `json:"email_addresses,omitempty"`
	Addresses     []PartyAddress     `json:"addresses,omitempty"`
}

// PartiesService provides access to /parties.
type PartiesService struct{ baseService }

// List returns parties (institution profile data) related to accounts the user
// has shared with the app.
//
// API: GET /parties
func (s *PartiesService) List(ctx context.Context, token string) ([]Party, error) {
	var out []Party
	err := s.c.doRequest(ctx, apiRequest{
		method: "GET",
		path:   "/parties",
		auth:   tokenAuth{token: token},
	}, &out)
	return out, err
}
