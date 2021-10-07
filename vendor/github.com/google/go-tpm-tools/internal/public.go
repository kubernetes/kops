package internal

import (
	"fmt"

	"github.com/google/go-tpm/tpm2"
)

// GetSigningHashAlg returns the hash algorithm used for a signing key. Returns
// an error if an algorithm isn't supported, or the key is not a signing key.
func GetSigningHashAlg(pubArea tpm2.Public) (tpm2.Algorithm, error) {
	if pubArea.Attributes&tpm2.FlagSign == 0 {
		return tpm2.AlgNull, fmt.Errorf("non-signing key used with signing operation")
	}

	var sigScheme *tpm2.SigScheme
	switch pubArea.Type {
	case tpm2.AlgRSA:
		sigScheme = pubArea.RSAParameters.Sign
	case tpm2.AlgECC:
		sigScheme = pubArea.ECCParameters.Sign
	default:
		return tpm2.AlgNull, fmt.Errorf("unsupported key type: %v", pubArea.Type)
	}

	if sigScheme == nil {
		return tpm2.AlgNull, fmt.Errorf("unsupported null signing scheme")
	}
	switch sigScheme.Alg {
	case tpm2.AlgRSAPSS, tpm2.AlgRSASSA, tpm2.AlgECDSA:
		return sigScheme.Hash, nil
	default:
		return tpm2.AlgNull, fmt.Errorf("unsupported signing algorithm: %v", sigScheme.Alg)
	}
}
