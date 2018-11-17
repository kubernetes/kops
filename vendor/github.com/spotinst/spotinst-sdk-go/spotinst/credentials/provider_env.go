package credentials

import (
	"fmt"
	"os"
)

const (
	// EnvCredentialsProviderName provides a name of Env provider.
	EnvCredentialsProviderName = "EnvCredentialsProvider"

	// EnvCredentialsVarToken specifies the name of the environment variable
	// points to the Spotinst Token.
	EnvCredentialsVarToken = "SPOTINST_TOKEN"

	// EnvCredentialsVarAccount specifies the name of the environment variable
	// points to the Spotinst account ID.
	EnvCredentialsVarAccount = "SPOTINST_ACCOUNT"
)

// ErrEnvCredentialsTokenNotFound is returned when the Spotinst Token can't be
// found in the process's environment.
var ErrEnvCredentialsTokenNotFound = fmt.Errorf("spotinst: %s not found in environment", EnvCredentialsVarToken)

// A EnvProvider retrieves credentials from the environment variables of the
// running process.
//
// Environment variables used:
// * Token: SPOTINST_TOKEN
type EnvProvider struct {
	retrieved bool
}

// NewEnvCredentials returns a pointer to a new Credentials object
// wrapping the environment variable provider.
func NewEnvCredentials() *Credentials {
	return NewCredentials(&EnvProvider{})
}

// Retrieve retrieves the keys from the environment.
func (e *EnvProvider) Retrieve() (Value, error) {
	e.retrieved = false

	token := os.Getenv(EnvCredentialsVarToken)
	if token == "" {
		return Value{ProviderName: EnvCredentialsProviderName},
			ErrEnvCredentialsTokenNotFound
	}

	e.retrieved = true
	value := Value{
		Token:        token,
		Account:      os.Getenv(EnvCredentialsVarAccount),
		ProviderName: EnvCredentialsProviderName,
	}

	return value, nil
}

func (e *EnvProvider) String() string {
	return EnvCredentialsProviderName
}
