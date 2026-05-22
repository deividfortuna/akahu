package akahu

import "context"

// ConnectionType describes how Akahu integrates with the institution.
type ConnectionType string

const (
	ConnectionTypeClassic  ConnectionType = "classic"
	ConnectionTypeOfficial ConnectionType = "official"
)

// ConnectionMigrationMode controls how classic and official open-banking
// connections coexist for an app.
type ConnectionMigrationMode string

const (
	ConnectionMigrationModeStrict      ConnectionMigrationMode = "strict"
	ConnectionMigrationModeMigration   ConnectionMigrationMode = "migration"
	ConnectionMigrationModeSideBySide  ConnectionMigrationMode = "side_by_side"
	ConnectionMigrationModeDeveloper   ConnectionMigrationMode = "developer"
)

// Connection describes a financial institution that Akahu can connect to.
type Connection struct {
	ID                    string                   `json:"_id"`
	Classic               string                   `json:"_classic,omitempty"`
	Name                  string                   `json:"name"`
	Logo                  string                   `json:"logo"`
	ConnectionType        ConnectionType           `json:"connection_type"`
	NewConnectionsEnabled bool                     `json:"new_connections_enabled"`
	Mode                  *ConnectionMigrationMode `json:"mode,omitempty"`
	Deadline              string                   `json:"deadline,omitempty"`
}

// ConnectionInfo is the abbreviated connection metadata that appears nested
// inside other resources (e.g. Account.Connection).
type ConnectionInfo struct {
	ID             string         `json:"_id"`
	Name           string         `json:"name"`
	Logo           string         `json:"logo"`
	ConnectionType ConnectionType `json:"connection_type"`
}

// ConnectionsService provides access to /connections.
type ConnectionsService struct{ baseService }

// List returns all connections the app has access to.
//
// API: GET /connections
func (s *ConnectionsService) List(ctx context.Context) ([]Connection, error) {
	var out []Connection
	err := s.c.doRequest(ctx, apiRequest{
		method: "GET",
		path:   "/connections",
		auth:   basicAuth{},
	}, &out)
	return out, err
}

// Get returns a single connection by id.
//
// API: GET /connections/{id}
func (s *ConnectionsService) Get(ctx context.Context, connectionID string) (*Connection, error) {
	var out Connection
	err := s.c.doRequest(ctx, apiRequest{
		method: "GET",
		path:   "/connections/" + connectionID,
		auth:   basicAuth{},
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Refresh refreshes all accounts under the given connection that have been
// linked by the user associated with the given token.
//
// API: POST /refresh/{connection_id}
func (s *ConnectionsService) Refresh(ctx context.Context, token, connectionID string) error {
	return s.c.doRequest(ctx, apiRequest{
		method: "POST",
		path:   "/refresh/" + connectionID,
		auth:   tokenAuth{token: token},
	}, nil)
}
