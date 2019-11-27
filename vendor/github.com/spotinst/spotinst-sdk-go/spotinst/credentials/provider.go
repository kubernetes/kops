package credentials

import "fmt"

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
// The Provider should not need to implement its own mutexes, because that will
// be managed by Credentials.
type Provider interface {
	fmt.Stringer

	// Refresh returns nil if it successfully retrieved the value. Error is
	// returned if the value were not obtainable, or empty.
	Retrieve() (Value, error)
}
