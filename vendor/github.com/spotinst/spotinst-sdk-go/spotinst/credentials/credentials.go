package credentials

import (
	"fmt"
	"sync"
)

// A Value is the Spotinst credentials value for individual credential fields.
type Value struct {
	// Spotinst API token.
	Token string `json:"token"`

	// Spotinst account ID.
	Account string `json:"account"`

	// Provider used to get credentials.
	ProviderName string `json:"-"`
}

// A Provider is the interface for any component which will provide credentials
// Value.
//
// The Provider should not need to implement its own mutexes, because
// that will be managed by Credentials.
type Provider interface {
	fmt.Stringer

	// Refresh returns nil if it successfully retrieved the value.
	// Error is returned if the value were not obtainable, or empty.
	Retrieve() (Value, error)
}

// A Credentials provides synchronous safe retrieval of Spotinst credentials.
// Credentials will cache the credentials value.
//
// Credentials is safe to use across multiple goroutines and will manage the
// synchronous state so the Providers do not need to implement their own
// synchronization.
//
// The first Credentials.Get() will always call Provider.Retrieve() to get the
// first instance of the credentials Value. All calls to Get() after that
// will return the cached credentials Value.
type Credentials struct {
	provider     Provider
	mu           sync.Mutex
	forceRefresh bool
	creds        Value
}

// NewCredentials returns a pointer to a new Credentials with the provider set.
func NewCredentials(provider Provider) *Credentials {
	return &Credentials{
		provider:     provider,
		forceRefresh: true,
	}
}

// Get returns the credentials value, or error if the credentials Value failed
// to be retrieved.
//
// Will return the cached credentials Value. If the credentials Value is empty
// the Provider's Retrieve() will be called to refresh the credentials.
func (c *Credentials) Get() (Value, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.creds.Token == "" || c.forceRefresh {
		creds, err := c.provider.Retrieve()
		if err != nil {
			return Value{}, err
		}
		c.creds = creds
		c.forceRefresh = false
	}

	return c.creds, nil
}

// Refresh refreshes the credentials and forces them to be retrieved on the
// next call to Get().
func (c *Credentials) Refresh() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.forceRefresh = true
}
