package akahu

import (
	"errors"
	"strings"
)

const (
	defaultBaseURL      = "https://api.akahu.io/v1"
	defaultOAuthBaseURL = "https://oauth.akahu.nz"
	appTokenPrefix      = "app_token_"
)

// Client is the entry point for all Akahu API access. Construct one with New
// and reuse it across goroutines — the underlying *http.Client is concurrency-
// safe.
type Client struct {
	cfg clientConfig

	// Resource services. Initialised in New.
	Auth           *AuthService
	Identities     *IdentitiesService
	Users          *UsersService
	Parties        *PartiesService
	Accounts       *AccountsService
	Authorisations *AuthorisationsService
	Connections    *ConnectionsService
	Categories     *CategoriesService
	Payments       *PaymentsService
	Transfers      *TransfersService
	Transactions   *TransactionsService
	Webhooks       *WebhooksService
}

// New constructs a Client. appToken is required and must begin with
// "app_token_". See the WithXxx ClientOption funcs for additional configuration.
func New(appToken string, opts ...ClientOption) (*Client, error) {
	if !strings.HasPrefix(appToken, appTokenPrefix) {
		return nil, errors.New(`akahu: appToken must be a string beginning with "app_token_"`)
	}

	cfg := clientConfig{
		appToken:     appToken,
		baseURL:      defaultBaseURL,
		oauthBaseURL: defaultOAuthBaseURL,
		userAgent:    userAgent,
		httpClient:   defaultHTTPClient(),
	}
	for _, opt := range opts {
		if err := opt(&cfg); err != nil {
			return nil, err
		}
	}
	if cfg.httpClient == nil {
		cfg.httpClient = defaultHTTPClient()
	}
	if cfg.userAgent == "" {
		cfg.userAgent = userAgent
	}
	if cfg.baseURL == "" {
		cfg.baseURL = defaultBaseURL
	}
	if cfg.oauthBaseURL == "" {
		cfg.oauthBaseURL = defaultOAuthBaseURL
	}

	c := &Client{cfg: cfg}
	base := baseService{c: c}
	c.Auth = &AuthService{baseService: base}
	c.Identities = &IdentitiesService{baseService: base}
	c.Users = &UsersService{baseService: base}
	c.Parties = &PartiesService{baseService: base}
	c.Accounts = &AccountsService{baseService: base}
	c.Authorisations = &AuthorisationsService{baseService: base}
	c.Connections = &ConnectionsService{baseService: base}
	c.Categories = &CategoriesService{baseService: base}
	c.Payments = &PaymentsService{baseService: base}
	c.Transfers = &TransfersService{baseService: base}
	c.Transactions = &TransactionsService{baseService: base}
	c.Webhooks = &WebhooksService{baseService: base}

	return c, nil
}
