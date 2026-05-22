package akahu

import "context"

// User represents an Akahu user account, as returned by GET /me.
type User struct {
	// ID is Akahu's unique identifier for this user.
	ID string `json:"_id"`
	// Email is the address the user registered with Akahu. Always present
	// when the app has the AKAHU scope.
	Email string `json:"email,omitempty"`
	// PreferredName is the user's preferred name, if they have provided one.
	PreferredName string `json:"preferred_name,omitempty"`
	// AccessGrantedAt is the ISO 8601 timestamp at which the user first
	// granted the calling app access to their accounts.
	AccessGrantedAt string `json:"access_granted_at"`

	// FirstName, LastName: only present on legacy users. Prefer Parties data.
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

// UsersService provides access to /me.
type UsersService struct{ baseService }

// Get returns the user associated with the given user access token.
//
// API: GET /me
func (s *UsersService) Get(ctx context.Context, token string) (*User, error) {
	var out User
	err := s.c.doRequest(ctx, apiRequest{
		method: "GET",
		path:   "/me",
		auth:   tokenAuth{token: token},
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}
