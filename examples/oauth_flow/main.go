// oauth_flow demonstrates the OAuth2 authorization-code exchange.
//
// On startup it prints the authorization URL the user should visit. After the
// user completes the flow, paste the `code` query parameter from the redirect
// URL back into the terminal, and the example will exchange it for a long-
// lived user access token.
package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/deividfortuna/akahu/akahu"
)

func main() {
	appToken := os.Getenv("AKAHU_APP_TOKEN")
	appSecret := os.Getenv("AKAHU_APP_SECRET")
	redirectURI := os.Getenv("AKAHU_REDIRECT_URI")
	if appToken == "" || appSecret == "" || redirectURI == "" {
		log.Fatal("set AKAHU_APP_TOKEN, AKAHU_APP_SECRET, AKAHU_REDIRECT_URI")
	}

	c, err := akahu.New(appToken, akahu.WithAppSecret(appSecret))
	if err != nil {
		log.Fatal(err)
	}

	authURL := c.Auth.BuildAuthorizationURL(akahu.AuthURLParams{
		RedirectURI: redirectURI,
		State:       "demo-state",
	})
	fmt.Println("1. Visit this URL in your browser:")
	fmt.Println("  ", authURL)
	fmt.Println("2. Complete the flow. After you're redirected, copy the `code` query parameter and paste it below.")
	fmt.Print("code: ")

	r := bufio.NewReader(os.Stdin)
	code, err := r.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	code = strings.TrimSpace(code)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tok, err := c.Auth.Exchange(ctx, code, redirectURI)
	if err != nil {
		log.Fatalf("Exchange: %v", err)
	}
	fmt.Printf("Access token: %s (scope %s)\n", tok.AccessToken, tok.Scope)
}
