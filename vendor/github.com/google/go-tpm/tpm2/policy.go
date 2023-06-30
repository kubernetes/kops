package tpm2

import (
	"bytes"
	"crypto"
	"reflect"
)

// PolicyCalculator represents a TPM 2.0 policy that needs to be calculated
// synthetically (i.e., without a TPM).
type PolicyCalculator struct {
	alg   TPMIAlgHash
	hash  crypto.Hash
	state []byte
}

// NewPolicyCalculator creates a fresh policy using the given hash algorithm.
func NewPolicyCalculator(alg TPMIAlgHash) (*PolicyCalculator, error) {
	hash, err := alg.Hash()
	if err != nil {
		return nil, err
	}
	return &PolicyCalculator{
		alg:   alg,
		hash:  hash,
		state: make([]byte, hash.Size()),
	}, nil
}

// Reset resets the internal state of the policy hash to all 0x00.
func (p *PolicyCalculator) Reset() {
	p.state = make([]byte, p.hash.Size())
}

// Update updates the internal state of the policy hash by appending the
// current state with the given contents, and updating the new state to the
// hash of that.
func (p *PolicyCalculator) Update(data ...interface{}) error {
	hash := p.hash.New()
	hash.Write(p.state)
	var buf bytes.Buffer
	for _, d := range data {
		if err := marshal(&buf, reflect.ValueOf(d)); err != nil {
			return err
		}
	}
	hash.Write(buf.Bytes())
	p.state = hash.Sum(nil)
	return nil
}

// Hash returns the current state of the policy hash.
func (p *PolicyCalculator) Hash() *TPMTHA {
	result := TPMTHA{
		HashAlg: p.alg,
		Digest:  make([]byte, len(p.state)),
	}
	copy(result.Digest, p.state)
	return &result
}
