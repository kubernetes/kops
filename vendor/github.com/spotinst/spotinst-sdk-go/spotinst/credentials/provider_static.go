package credentials

import (
	"errors"
)

// StaticCredentialsProviderName specifies the name of the Static provider.
const StaticCredentialsProviderName = "StaticCredentialsProvider"

// ErrStaticCredentialsEmpty is returned when static credentials are empty.
var ErrStaticCredentialsEmpty = errors.New("spotinst: static credentials are empty")

// A StaticProvider is a set of credentials which are set programmatically.
type StaticProvider struct {
	Value
}

// NewStaticCredentials returns a pointer to a new Credentials object wrapping
// a static credentials value provider.
func NewStaticCredentials(token, account string) *Credentials {
	return NewCredentials(&StaticProvider{Value: Value{
		ProviderName: StaticCredentialsProviderName,
		Token:        token,
		Account:      account,
	}})
}

// Retrieve returns the credentials or error if the credentials are invalid.
func (s *StaticProvider) Retrieve() (Value, error) {
	if s.IsEmpty() {
		return s.Value, ErrStaticCredentialsEmpty
	}

	return s.Value, nil
}

// String returns the string representation of the provider.
func (s *StaticProvider) String() string { return StaticCredentialsProviderName }
