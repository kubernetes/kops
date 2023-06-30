//go:build windows

package transport

import (
	legacy "github.com/google/go-tpm/legacy/tpm2"
)

// Wrap the legacy OpenTPM function so callers don't have to import both the
// legacy and the new TPM 2.0 API.
// TODO: When we delete the legacy API, we can make this the only copy of
// OpenTPM.

// OpenTPM opens a channel to the TPM at the given path. If the file is a
// device, then it treats it like a normal TPM device, and if the file is a
// Unix domain socket, then it opens a connection to the socket.
//
// This function may also be invoked with no paths, as tpm2.OpenTPM(). In this
// case, the default paths on Linux (/dev/tpmrm0 then /dev/tpm0), will be used.
func OpenTPM() (TPMCloser, error) {
	rwc, err := legacy.OpenTPM()
	if err != nil {
		return nil, err
	}
	return &wrappedRWC{
		transport: rwc,
	}, nil
}
