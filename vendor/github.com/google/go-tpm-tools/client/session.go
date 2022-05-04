package client

import (
	"io"

	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpmutil"
)

type session interface {
	io.Closer
	Auth() (tpm2.AuthCommand, error)
}

func startAuthSession(rw io.ReadWriter) (session tpmutil.Handle, err error) {
	// This session assumes the bus is trusted, so we:
	// - use nil for tpmKey, encrypted salt, and symmetric
	// - use and all-zeros caller nonce, and ignore the returned nonce
	// As we are creating a plain TPM session, we:
	// - setup a policy session
	// - don't bind the session to any particular key
	session, _, err = tpm2.StartAuthSession(
		rw,
		/*tpmKey=*/ tpm2.HandleNull,
		/*bindKey=*/ tpm2.HandleNull,
		/*nonceCaller=*/ make([]byte, SessionHashAlg.Size()),
		/*encryptedSalt=*/ nil,
		/*sessionType=*/ tpm2.SessionPolicy,
		/*symmetric=*/ tpm2.AlgNull,
		/*authHash=*/ SessionHashAlgTpm)
	return
}

type pcrSession struct {
	rw      io.ReadWriter
	session tpmutil.Handle
	sel     tpm2.PCRSelection
}

func newPCRSession(rw io.ReadWriter, sel tpm2.PCRSelection) (session, error) {
	if len(sel.PCRs) == 0 {
		return nullSession{}, nil
	}
	session, err := startAuthSession(rw)
	return pcrSession{rw, session, sel}, err
}

func (p pcrSession) Auth() (auth tpm2.AuthCommand, err error) {
	if err = tpm2.PolicyPCR(p.rw, p.session, nil, p.sel); err != nil {
		return
	}
	return tpm2.AuthCommand{Session: p.session, Attributes: tpm2.AttrContinueSession}, nil
}

func (p pcrSession) Close() error {
	return tpm2.FlushContext(p.rw, p.session)
}

type ekSession struct {
	rw      io.ReadWriter
	session tpmutil.Handle
}

func newEKSession(rw io.ReadWriter) (session, error) {
	session, err := startAuthSession(rw)
	return ekSession{rw, session}, err
}

func (e ekSession) Auth() (auth tpm2.AuthCommand, err error) {
	nullAuth := tpm2.AuthCommand{Session: tpm2.HandlePasswordSession, Attributes: tpm2.AttrContinueSession}
	if _, _, err = tpm2.PolicySecret(e.rw, tpm2.HandleEndorsement, nullAuth, e.session, nil, nil, nil, 0); err != nil {
		return
	}
	return tpm2.AuthCommand{Session: e.session, Attributes: tpm2.AttrContinueSession}, nil
}

func (e ekSession) Close() error {
	return tpm2.FlushContext(e.rw, e.session)
}

type nullSession struct{}

func (n nullSession) Auth() (auth tpm2.AuthCommand, err error) {
	return tpm2.AuthCommand{Session: tpm2.HandlePasswordSession, Attributes: tpm2.AttrContinueSession}, nil
}

func (n nullSession) Close() error {
	return nil
}
