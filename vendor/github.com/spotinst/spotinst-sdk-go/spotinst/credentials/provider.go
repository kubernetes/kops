package credentials

import "fmt"

// A Value is the Spotinst credentials value for individual credential fields.
type Value struct {
	// Spotinst API token.
	Token string `ini:"token" json:"token"`

	// Spotinst account ID.
	Account string `ini:"account" json:"account"`

	// Provider used to get credentials.
	ProviderName string `ini:"-" json:"-"`
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

// IsEmpty if all fields of a Value are empty.
func (v *Value) IsEmpty() bool { return v.Token == "" && v.Account == "" }

// IsComplete if all fields of a Value are set.
func (v *Value) IsComplete() bool { return v.Token != "" && v.Account != "" }

// Merge merges the passed in Value into the existing Value object.
func (v *Value) Merge(v2 Value) {
	if v.Token == "" {
		v.Token = v2.Token
	}
	if v.Account == "" {
		v.Account = v2.Account
	}
}
