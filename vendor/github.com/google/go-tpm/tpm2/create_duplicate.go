package tpm2

import (
	"io"
)

var (
	labelDuplicate = "DUPLICATE"
)

// CreateDuplicate encrypts an object so that it can be imported under a target Storage Key.
// An inner wrapper is not supported.
func CreateDuplicate(rand io.Reader, pub LabeledEncapsulationKey, name []byte, sensitive []byte) (duplicate []byte, encSecret []byte, err error) {
	secret, ciphertext, err := pub.Encapsulate(rand, labelDuplicate)
	if err != nil {
		return nil, nil, err
	}

	// Marshal the sensitive as a TPM2B_SENSITIVE before encrypting it.
	// See Part 1, "Outer Duplication Wrapper"
	sensitive2B := Marshal(TPM2BDigest{Buffer: sensitive})

	// Encrypt the sensitive2B as dupSensitive.
	dupSensitive, err := deriveAndEncrypt(pub, secret, name, sensitive2B)
	if err != nil {
		return nil, nil, err
	}

	// Compute the HMAC of (dupSensitive || name)
	outerHMAC, err := deriveAndHMAC(pub, secret, nil, dupSensitive, name)
	if err != nil {
		return nil, nil, err
	}

	// Marshal the virtual _PRIVATE ourselves. We have to do this since dupSensitive's size is encrypted.
	duplicate = make([]byte, 0, 2+len(outerHMAC)+len(dupSensitive))
	duplicate = append(duplicate, Marshal(TPM2BDigest{Buffer: outerHMAC})...)
	duplicate = append(duplicate, dupSensitive...)

	return duplicate, ciphertext, nil
}
