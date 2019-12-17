package credentials

import (
	"errors"
	"fmt"
)

// ErrNoValidProvidersFoundInChain Is returned when there are no valid credentials
// providers in the ChainProvider.
var ErrNoValidProvidersFoundInChain = errors.New("spotinst: no valid credentials providers in chain")

// A ChainProvider will search for a provider which returns credentials and cache
// that provider until Retrieve is called again.
//
// The ChainProvider provides a way of chaining multiple providers together which
// will pick the first available using priority order of the Providers in the list.
//
// If none of the Providers retrieve valid credentials Value, ChainProvider's
// Retrieve() will return the error ErrNoValidProvidersFoundInChain.
//
// If a Provider is found which returns valid credentials Value ChainProvider
// will cache that Provider for all calls until Retrieve is called again.
//
// Example of ChainProvider to be used with an EnvCredentialsProvider and
// FileCredentialsProvider. In this example EnvProvider will first check if any
// credentials are available via the environment variables. If there are none
// ChainProvider will check the next Provider in the list, FileProvider in this
// case. If FileCredentialsProvider does not return any credentials ChainProvider
// will return the error ErrNoValidProvidersFoundInChain.
//
//	creds := credentials.NewChainCredentials(
//		new(credentials.EnvProvider),
//		new(credentials.FileProvider),
//	)
type ChainProvider struct {
	Providers []Provider
	active    Provider
}

// NewChainCredentials returns a pointer to a new Credentials object
// wrapping a chain of providers.
func NewChainCredentials(providers ...Provider) *Credentials {
	return NewCredentials(&ChainProvider{
		Providers: providers,
	})
}

// Retrieve returns the credentials value or error if no provider returned
// without error. If a provider is found it will be cached.
func (c *ChainProvider) Retrieve() (Value, error) {
	var errs errorList
	for _, p := range c.Providers {
		value, err := p.Retrieve()
		if err == nil {
			c.active = p
			return value, nil
		}
		errs = append(errs, err)
	}
	c.active = nil

	err := ErrNoValidProvidersFoundInChain
	if len(errs) > 0 {
		err = errs
	}

	return Value{}, err
}

func (c *ChainProvider) String() string {
	var out string
	for i, provider := range c.Providers {
		out += provider.String()
		if i < len(c.Providers)-1 {
			out += " "
		}
	}
	return out
}

// An error list that satisfies the error interface.
type errorList []error

// Error returns the string representation of the error.
//
// Satisfies the error interface.
func (e errorList) Error() string {
	msg := ""
	if size := len(e); size > 0 {
		for i := 0; i < size; i++ {
			msg += fmt.Sprintf("%s", e[i].Error())

			// Check the next index to see if it is within the slice. If it is,
			// append a newline. We do this, because unit tests could be broken
			// with the additional '\n'.
			if i+1 < size {
				msg += "\n"
			}
		}
	}
	return msg
}
