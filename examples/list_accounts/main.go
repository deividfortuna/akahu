// list_accounts prints the user's linked accounts and their balances.
//
// Usage:
//
//	AKAHU_APP_TOKEN=app_token_... AKAHU_USER_TOKEN=user_token_... \
//	    go run ./examples/list_accounts
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/deividfortuna/akahu/akahu"
)

func main() {
	appToken := os.Getenv("AKAHU_APP_TOKEN")
	userToken := os.Getenv("AKAHU_USER_TOKEN")
	if appToken == "" || userToken == "" {
		log.Fatal("AKAHU_APP_TOKEN and AKAHU_USER_TOKEN must be set")
	}

	c, err := akahu.New(appToken)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	user, err := c.Users.Get(ctx, userToken)
	if err != nil {
		log.Fatalf("Users.Get: %v", err)
	}
	accounts, err := c.Accounts.List(ctx, userToken)
	if err != nil {
		log.Fatalf("Accounts.List: %v", err)
	}

	fmt.Printf("%s has linked %d accounts:\n", user.Email, len(accounts))
	for _, a := range accounts {
		var bal string
		if a.Balance != nil && a.Balance.Available != nil {
			bal = fmt.Sprintf("$%.2f available", *a.Balance.Available)
		} else if a.Balance != nil {
			bal = fmt.Sprintf("$%.2f", a.Balance.Current)
		} else {
			bal = "(no balance)"
		}
		fmt.Printf("  %s account %q (%s) %s\n",
			a.Connection.Name, a.Name, a.FormattedAccount, bal)
	}
}
