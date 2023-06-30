package tpm2

import (
	"crypto"

	legacy "github.com/google/go-tpm/legacy/tpm2"
)

// KDFa implements TPM 2.0's default key derivation function, as defined in
// section 11.4.9.2 of the TPM revision 2 specification part 1.
// See: https://trustedcomputinggroup.org/resource/tpm-library-specification/
// The key & label parameters must not be zero length.
// The label parameter is a non-null-terminated string.
// The contextU & contextV parameters are optional.
func KDFa(h crypto.Hash, key []byte, label string, contextU, contextV []byte, bits int) []byte {
	return legacy.KDFaHash(h, key, label, contextU, contextV, bits)
}

// KDFe implements TPM 2.0's ECDH key derivation function, as defined in
// section 11.4.9.3 of the TPM revision 2 specification part 1.
// See: https://trustedcomputinggroup.org/resource/tpm-library-specification/
// The z parameter is the x coordinate of one party's private ECC key multiplied
// by the other party's public ECC point.
// The use parameter is a non-null-terminated string.
// The partyUInfo and partyVInfo are the x coordinates of the initiator's and
// the responder's ECC points, respectively.
func KDFe(h crypto.Hash, z []byte, use string, partyUInfo, partyVInfo []byte, bits int) []byte {
	return legacy.KDFeHash(h, z, use, partyUInfo, partyVInfo, bits)
}
