package tpm2

import (
	"bytes"
	"fmt"
	"reflect"
)

// CommandAudit represents an audit session for attesting the execution of a
// series of commands in the TPM. It is useful for both command and session
// auditing.
type CommandAudit struct {
	hash   TPMIAlgHash
	digest []byte
}

// NewAudit initializes a new CommandAudit with the specified hash algorithm.
func NewAudit(hash TPMIAlgHash) (*CommandAudit, error) {
	h, err := hash.Hash()
	if err != nil {
		return nil, err
	}
	return &CommandAudit{
		hash:   hash,
		digest: make([]byte, h.Size()),
	}, nil
}

// AuditCommand extends the audit digest with the given command and response.
// Go Generics do not allow type parameters on methods, otherwise this would be
// a method on CommandAudit.
// See https://github.com/golang/go/issues/49085 for more information.
func AuditCommand[C Command[R, *R], R any](a *CommandAudit, cmd C, rsp *R) error {
	cc := cmd.Command()
	cpHash, err := auditCPHash[R](cc, a.hash, cmd)
	if err != nil {
		return err
	}
	rpHash, err := auditRPHash(cc, a.hash, rsp)
	if err != nil {
		return err
	}
	ha, err := a.hash.Hash()
	if err != nil {
		return err
	}
	h := ha.New()
	h.Write(a.digest)
	h.Write(cpHash)
	h.Write(rpHash)
	a.digest = h.Sum(nil)
	return nil
}

// Digest returns the current digest of the audit.
func (a *CommandAudit) Digest() []byte {
	return a.digest
}

// auditCPHash calculates the command parameter hash for a given command with
// the given hash algorithm. The command is assumed to not have any decrypt
// sessions.
func auditCPHash[R any](cc TPMCC, h TPMIAlgHash, c Command[R, *R]) ([]byte, error) {
	names, err := cmdNames(c)
	if err != nil {
		return nil, err
	}
	parms, err := cmdParameters(c, nil)
	if err != nil {
		return nil, err
	}
	return cpHash(h, cc, names, parms)
}

// auditRPHash calculates the response parameter hash for a given response with
// the given hash algorithm. The command is assumed to be successful and to not
// have any encrypt sessions.
func auditRPHash(cc TPMCC, h TPMIAlgHash, r any) ([]byte, error) {
	var parms bytes.Buffer
	parameters := taggedMembers(reflect.ValueOf(r).Elem(), "handle", true)
	for i, parameter := range parameters {
		if err := marshal(&parms, parameter); err != nil {
			return nil, fmt.Errorf("marshalling parameter %v: %w", i+1, err)
		}
	}
	return rpHash(h, TPMRCSuccess, cc, parms.Bytes())
}
