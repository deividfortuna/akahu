// webhook_server validates and dispatches incoming Akahu webhook events.
//
// Usage:
//
//	AKAHU_APP_TOKEN=app_token_... AKAHU_APP_SECRET=app_secret_... \
//	    go run ./examples/webhook_server
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/deividfortuna/akahu/akahu"
)

func main() {
	appToken := os.Getenv("AKAHU_APP_TOKEN")
	appSecret := os.Getenv("AKAHU_APP_SECRET")
	if appToken == "" || appSecret == "" {
		log.Fatal("set AKAHU_APP_TOKEN and AKAHU_APP_SECRET")
	}

	c, err := akahu.New(appToken, akahu.WithAppSecret(appSecret))
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		keyID, err := strconv.Atoi(r.Header.Get("X-Akahu-Signing-Key"))
		if err != nil {
			http.Error(w, "missing X-Akahu-Signing-Key", http.StatusBadRequest)
			return
		}
		signature := r.Header.Get("X-Akahu-Signature")

		payload, err := c.Webhooks.ValidateWebhook(context.Background(), keyID, signature, body, nil)
		if err != nil {
			log.Printf("webhook validation failed: %v", err)
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}

		switch p := payload.(type) {
		case *akahu.AccountUpdatePayload:
			fmt.Printf("account %s updated: fields=%v\n", p.ItemID, p.UpdatedFields)
		case *akahu.TransactionUpdatePayload:
			fmt.Printf("account %s: %d new transactions\n", p.ItemID, p.NewTransactions)
		case *akahu.PaymentUpdatePayload:
			fmt.Printf("payment %s -> %s\n", p.ItemID, p.Status)
		default:
			fmt.Printf("event: type=%s code=%s\n", p.WebhookType(), p.WebhookCode())
		}
		w.WriteHeader(http.StatusNoContent)
	})

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
