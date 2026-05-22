# Akahu Go SDK

An unofficial Go SDK for the [Akahu](https://akahu.nz) open finance API,
ported from the official [JavaScript SDK](https://github.com/akahu-io/akahu-sdk-js).

> Akahu builds and maintains data integrations with banks and other financial
> institutions in New Zealand and bundles them into a simple API.

## Install

```sh
go get github.com/deividfortuna/akahu
```

Requires Go 1.22+.

## Quickstart

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/deividfortuna/akahu/akahu"
)

func main() {
    c, err := akahu.New("app_token_...",
        akahu.WithAppSecret("app_secret_..."), // server-side only
    )
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()
    accounts, err := c.Accounts.List(ctx, "user_token_...")
    if err != nil {
        log.Fatal(err)
    }
    for _, a := range accounts {
        fmt.Println(a.Name, a.FormattedAccount)
    }
}
```

## Authentication

Akahu uses two authentication modes:

- **User token** — passed as the first argument to most resource methods.
  Issued via the OAuth flow (`Auth.Exchange`) or copied from a [personal app](https://developers.akahu.nz/docs/personal-apps).
- **App basic auth** — used for endpoints that act on the app itself
  (`Connections.List`, `Categories.List`, `Webhooks.GetPublicKey`,
  `Auth.Exchange`). Configure with `akahu.WithAppSecret(...)`. **Do not embed
  your app secret in client-side or mobile code.**

## Resources

| Service | Methods |
|---|---|
| `c.Auth` | `BuildAuthorizationURL`, `Exchange`, `Revoke` |
| `c.Identities` | `BuildAuthorizationURL`, `Get`, `VerifyName` |
| `c.Users` | `Get` |
| `c.Parties` | `List` |
| `c.Accounts` | `List`, `Get`, `ListTransactions`, `ListPendingTransactions`, `Refresh`, `RefreshAll`, `Revoke` (deprecated) |
| `c.Authorisations` | `Revoke` |
| `c.Connections` | `List`, `Get`, `Refresh` |
| `c.Categories` | `List`, `Get` |
| `c.Payments` | `Get`, `List`, `Create`, `CreateToIRD`, `Cancel` |
| `c.Transfers` | `Get`, `List`, `Create` |
| `c.Transactions` | `List`, `ListPending`, `Get`, `GetMany` |
| `c.Webhooks` | `List`, `Subscribe`, `Unsubscribe`, `ListEvents`, `GetPublicKey`, `ValidateWebhook` |

## Pagination

`Transactions.List` and `Accounts.ListTransactions` are paginated. The returned
`*akahu.Page[T]` carries `Items` and a cursor:

```go
var all []akahu.Transaction
var cursor string
for {
    page, err := c.Transactions.List(ctx, userToken, &akahu.TransactionQuery{Cursor: cursor})
    if err != nil { return err }
    all = append(all, page.Items...)
    if page.Cursor.Next == nil { break }
    cursor = *page.Cursor.Next
}
```

## OAuth2

```go
url := c.Auth.BuildAuthorizationURL(akahu.AuthURLParams{
    RedirectURI: "https://myapp.example/callback",
    State:       "csrf-state",
})
// redirect the user to `url`...
// ...the user comes back with ?code=...
tok, err := c.Auth.Exchange(ctx, code, "https://myapp.example/callback")
```

## Webhooks

```go
http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
    body, _ := io.ReadAll(r.Body)
    keyID, _ := strconv.Atoi(r.Header.Get("X-Akahu-Signing-Key"))
    sig := r.Header.Get("X-Akahu-Signature")

    payload, err := c.Webhooks.ValidateWebhook(r.Context(), keyID, sig, body, nil)
    if err != nil { http.Error(w, err.Error(), 401); return }

    switch p := payload.(type) {
    case *akahu.AccountUpdatePayload:
        log.Println("account updated", p.ItemID, p.UpdatedFields)
    case *akahu.TransactionUpdatePayload:
        log.Println("new tx", p.NewTransactionIDs)
    }
    w.WriteHeader(204)
})
```

To share signing-key cache across processes (e.g. with Redis), implement
`akahu.SigningKeyCache` and pass it via `&akahu.CacheConfig{Cache: yourCache}`.

## Errors

Every API error is a `*akahu.APIError`:

```go
if err != nil {
    var apiErr *akahu.APIError
    if errors.As(err, &apiErr) {
        log.Printf("status %d: %s", apiErr.StatusCode, apiErr.Message)
    }
    return err
}
```

Use `akahu.IsAkahuError(err)` for a quick boolean check.

## Configuration

```go
c, err := akahu.New("app_token_...",
    akahu.WithAppSecret("app_secret_..."),
    akahu.WithHTTPClient(&http.Client{Timeout: 15 * time.Second}),
    akahu.WithRetries(3),
    akahu.WithBaseURL("https://staging.api.akahu.io/v1"),
    akahu.WithRequestHeaders(http.Header{"X-Trace-Id": {"..."}}),
)
```

Retries cover transient network errors only — never 4xx/5xx responses. POST
requests are retried only when an `Idempotency-Key` is set, which the SDK does
automatically.

## License

ISC. See [LICENSE](./LICENSE).
