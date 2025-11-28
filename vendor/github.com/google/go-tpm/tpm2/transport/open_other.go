//go:build !windows

package transport

import (
	legacy "github.com/google/go-tpm/legacy/tpm2"
)

// OpenTPM opens the TPM at the given path. If no path is provided, it will
// attempt to use reasonable defaults.
//
// Deprecated: Please use the individual transport packages (e.g.,
// go-tpm/tpm2/transport/linuxtpm).
func OpenTPM(path ...string) (TPMCloser, error) {
	rwc, err := legacy.OpenTPM(path...)
	if err != nil {
		return nil, err
	}
	return &wrappedRWC{
		transport: rwc,
	}, nil
}
