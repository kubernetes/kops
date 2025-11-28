package tpm2

import (
	"errors"
	"fmt"
	"io"
)

var (
	ErrUnsupportedType = errors.New("unsupported key type")
)

// An LabeledEncapsulationKey represents a public key used in a TPM labeled-encapsulation scheme.
type LabeledEncapsulationKey interface {
	// Encapsulate performs the labeled key encapsulation.
	Encapsulate(random io.Reader, label string) (secret []byte, ciphertext []byte, err error)
	// NameAlg fetches the Name hash algorithm of the encapsulation key.
	NameAlg() TPMAlgID
	// SymmetricParameters fetches the symmetric parameters for protection.
	SymmetricParameters() *TPMTSymDefObject
}

// ImportEncapsulationKey imports the TPM-form public key as a LabeledEncapsulationkey.
func ImportEncapsulationKey(pub *TPMTPublic) (LabeledEncapsulationKey, error) {
	switch pub.Type {
	case TPMAlgRSA:
		return importRSAEncapsulationKey(pub)
	case TPMAlgECC:
		return importECCEncapsulationKey(pub)
	default:
		return nil, fmt.Errorf("%w %v", ErrUnsupportedType, pub.Type)
	}
}
