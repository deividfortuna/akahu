package akahu

import "context"

// AuthorisationsService provides access to authorisation management endpoints.
type AuthorisationsService struct{ baseService }

// Revoke removes a single authorisation from the given user token. After this
// call, the token will no longer have access to the authorisation's accounts
// or transactions.
//
// API: DELETE /authorisations/{id}
func (s *AuthorisationsService) Revoke(ctx context.Context, token, authorisationID string) error {
	return s.c.doRequest(ctx, apiRequest{
		method: "DELETE",
		path:   "/authorisations/" + authorisationID,
		auth:   tokenAuth{token: token},
	}, nil)
}
