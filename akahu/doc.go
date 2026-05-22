// Package akahu is a Go SDK for the Akahu open finance API.
//
// Akahu is New Zealand's open finance platform. This package is an unofficial
// Go port of the official JavaScript SDK at https://github.com/akahu-io/akahu-sdk-js.
//
// Quickstart:
//
//	c, err := akahu.New("app_token_...",
//	    akahu.WithAppSecret("app_secret_..."),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	accounts, err := c.Accounts.List(ctx, userToken)
//
// See https://developers.akahu.nz for full API documentation.
package akahu
