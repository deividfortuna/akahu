package akahu

// baseService is embedded in every resource service. It holds a pointer back
// to the owning Client so resources can share the apiCall machinery.
type baseService struct{ c *Client }
