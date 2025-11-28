// Package transport implements types for physically talking to TPMs.
package transport

import (
	"io"

	"github.com/google/go-tpm/tpmutil"
)

// TPM represents a logical connection to a TPM.
type TPM interface {
	Send(input []byte) ([]byte, error)
}

// TPMCloser represents a logical connection to a TPM and you can close it.
type TPMCloser interface {
	TPM
	io.Closer
}

// wrappedRW represents a struct that wraps an io.ReadWriter
// to a transport.TPM to be compatible with tpmdirect.
type wrappedRW struct {
	transport io.ReadWriter
}

// wrappedRWC represents a struct that wraps an io.ReadWriteCloser
// to a transport.TPM to be compatible with tpmdirect.
type wrappedRWC struct {
	transport io.ReadWriteCloser
}

// wrappedTPM represents a struct that wraps a transport.TPM's underlying
// transport to use with legacy code that expects an io.ReadWriter.
type wrappedTPM struct {
	response []byte
	tpm      TPM
}

// FromReadWriter takes in a io.ReadWriter and returns a
// transport.TPM wrapping the io.ReadWriter.
func FromReadWriter(rw io.ReadWriter) TPM {
	return &wrappedRW{transport: rw}
}

// FromReadWriteCloser takes in a io.ReadWriteCloser and returns a
// transport.TPMCloser wrapping the io.ReadWriteCloser.
func FromReadWriteCloser(rwc io.ReadWriteCloser) TPMCloser {
	return &wrappedRWC{transport: rwc}
}

// ToReadWriter takes in a transport TPM and returns an
// io.ReadWriter wrapping the transport TPM.
func ToReadWriter(tpm TPM) io.ReadWriter {
	return &wrappedTPM{tpm: tpm}
}

// Read copies t.response into the p buffer and return the appropriate length.
func (t *wrappedTPM) Read(p []byte) (int, error) {
	n := copy(p, t.response)
	t.response = t.response[n:]
	if len(t.response) == 0 {
		return n, io.EOF
	}
	return n, nil
}

// Write implements the io.ReadWriter interface.
func (t *wrappedTPM) Write(p []byte) (n int, err error) {
	t.response, err = t.tpm.Send(p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

// Send implements the TPM interface.
func (t *wrappedRW) Send(input []byte) ([]byte, error) {
	return tpmutil.RunCommandRaw(t.transport, input)
}

// Send implements the TPM interface.
func (t *wrappedRWC) Send(input []byte) ([]byte, error) {
	return tpmutil.RunCommandRaw(t.transport, input)
}

// Close implements the TPM interface.
func (t *wrappedRWC) Close() error {
	return t.transport.Close()
}
