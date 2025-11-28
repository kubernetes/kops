//go:build windows

package transport

import (
	legacy "github.com/google/go-tpm/legacy/tpm2"
)

// OpenTPM opens the local system TPM.
//
// Deprecated: Please use the individual transport packages (e.g.,
// go-tpm/tpm2/transport/windowstpm).
func OpenTPM() (TPMCloser, error) {
	rwc, err := legacy.OpenTPM()
	if err != nil {
		return nil, err
	}
	return &wrappedRWC{
		transport: rwc,
	}, nil
}
