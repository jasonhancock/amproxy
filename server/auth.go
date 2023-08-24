package server

import "errors"

// ErrCredentialsNotFound is the error returned by an AuthProvider if credentials
// corresponding to a given AccessKey cannot be located
var ErrCredentialsNotFound = errors.New("credentials not found")

// AuthProvider represents the interface for an authentication provider
type AuthProvider interface {
	// CredsForKey given an access key, returns the set of corresponding credentials.
	// If no corresponding credentials can be found, an ErrCredentialsNotFound will
	// be returned.
	CredsForKey(string) (*Creds, error)
}

// Creds represents an api key set and the metrics they are allowed to access
type Creds struct {
	AccessKey string              `yaml:"access_key"`
	SecretKey string              `yaml:"secret_key"`
	Metrics   map[string]struct{} `yaml:"metrics"`
}

// AllowMetric returns true if a given metric is allowed for this set of credentials
func (c *Creds) AllowMetric(name string) bool {
	_, ok := c.Metrics[name]
	return ok
}

type mockAuthProvider struct {
	CredsForKeyFn func(string) (*Creds, error)
}

func (m *mockAuthProvider) CredsForKey(key string) (*Creds, error) {
	if m.CredsForKeyFn != nil {
		return m.CredsForKeyFn(key)
	}
	panic("not implemented")
}
