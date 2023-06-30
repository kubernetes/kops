package client

import (
	"io"

	"github.com/google/go-tpm/legacy/tpm2"
	"github.com/google/go-tpm/tpmutil"
)

// Session is an interface for TPM sessions.
type Session interface {
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

// PCRSession is a TPM session that is bound to a set of PCRs.
type PCRSession struct {
	rw      io.ReadWriter
	session tpmutil.Handle
	sel     tpm2.PCRSelection
}

// NewPCRSession creates a new PCRSession.
func NewPCRSession(rw io.ReadWriter, sel tpm2.PCRSelection) (Session, error) {
	if len(sel.PCRs) == 0 {
		return NullSession{}, nil
	}
	session, err := startAuthSession(rw)
	return PCRSession{rw, session, sel}, err
}

// Auth returns the AuthCommand for the session.
func (p PCRSession) Auth() (auth tpm2.AuthCommand, err error) {
	if err = tpm2.PolicyPCR(p.rw, p.session, nil, p.sel); err != nil {
		return
	}
	return tpm2.AuthCommand{Session: p.session, Attributes: tpm2.AttrContinueSession}, nil
}

// Close closes the session.
func (p PCRSession) Close() error {
	return tpm2.FlushContext(p.rw, p.session)
}

// EKSession is a TPM session that is bound to the EK.
type EKSession struct {
	rw      io.ReadWriter
	session tpmutil.Handle
}

// NewEKSession creates a new EKSession.
func NewEKSession(rw io.ReadWriter) (Session, error) {
	session, err := startAuthSession(rw)
	return EKSession{rw, session}, err
}

// Auth returns the AuthCommand for the session.
func (e EKSession) Auth() (auth tpm2.AuthCommand, err error) {
	nullAuth := tpm2.AuthCommand{Session: tpm2.HandlePasswordSession, Attributes: tpm2.AttrContinueSession}
	if _, _, err = tpm2.PolicySecret(e.rw, tpm2.HandleEndorsement, nullAuth, e.session, nil, nil, nil, 0); err != nil {
		return
	}
	return tpm2.AuthCommand{Session: e.session, Attributes: tpm2.AttrContinueSession}, nil
}

// Close closes the session.
func (e EKSession) Close() error {
	return tpm2.FlushContext(e.rw, e.session)
}

// NullSession is a TPM session that is not bound to anything.
type NullSession struct{}

// Auth returns the AuthCommand for the session.
func (n NullSession) Auth() (auth tpm2.AuthCommand, err error) {
	return tpm2.AuthCommand{Session: tpm2.HandlePasswordSession, Attributes: tpm2.AttrContinueSession}, nil
}

// Close closes the session.
func (n NullSession) Close() error {
	return nil
}
