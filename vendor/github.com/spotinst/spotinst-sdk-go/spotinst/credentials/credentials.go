package credentials

import (
	"sync"
)

// A Credentials provides synchronous safe retrieval of Spotinst credentials.
// Credentials will cache the credentials value.
//
// Credentials is safe to use across multiple goroutines and will manage the
// synchronous state so the Providers do not need to implement their own
// synchronization.
//
// The first Credentials.Get() will always call Provider.Retrieve() to get the
// first instance of the credentials Value. All calls to Get() after that will
// return the cached credentials Value.
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

// Refresh refreshes the credentials and forces them to be retrieved on the next
// call to Get().
func (c *Credentials) Refresh() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.forceRefresh = true
}
