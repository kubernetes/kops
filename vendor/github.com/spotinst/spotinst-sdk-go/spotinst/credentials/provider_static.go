package credentials

import (
	"errors"
)

// StaticCredentialsProviderName provides a name of Static provider.
const StaticCredentialsProviderName = "StaticProvider"

// ErrStaticCredentialsEmpty is emitted when static credentials are empty.
var ErrStaticCredentialsEmpty = errors.New("spotinst: static credentials are empty")

// A StaticProvider is a set of credentials which are set programmatically.
type StaticProvider struct {
	Value
}

// NewStaticCredentials returns a pointer to a new Credentials object
// wrapping a static credentials value provider.
func NewStaticCredentials(token, account string) *Credentials {
	return NewCredentials(&StaticProvider{Value: Value{
		Token:   token,
		Account: account,
	}})
}

// Retrieve returns the credentials or error if the credentials are invalid.
func (s *StaticProvider) Retrieve() (Value, error) {
	if s.Token == "" {
		return Value{ProviderName: StaticCredentialsProviderName},
			ErrStaticCredentialsEmpty
	}
	if len(s.Value.ProviderName) == 0 {
		s.Value.ProviderName = StaticCredentialsProviderName
	}
	return s.Value, nil
}

func (s *StaticProvider) String() string {
	return StaticCredentialsProviderName
}
