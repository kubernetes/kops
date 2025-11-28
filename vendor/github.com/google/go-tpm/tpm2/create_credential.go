package tpm2

import (
	"io"
)

var (
	labelIdentity = "IDENTITY"
)

// CreateCredential creates an encrypted secret that can be recovered using ActivateCredential as part of a key-attestation flow.
func CreateCredential(rand io.Reader, pub LabeledEncapsulationKey, name []byte, credentialValue []byte) (idObject []byte, encSecret []byte, err error) {
	secret, ciphertext, err := pub.Encapsulate(rand, labelIdentity)
	if err != nil {
		return nil, nil, err
	}

	// Marshal the credentialValue as a TPM2B_DIGEST before encrypting it.
	// See Part 1, "Credential Protection", and Part 2, "TPMS_ID_OBJECT".
	credential2B := Marshal(TPM2BDigest{Buffer: credentialValue})

	// Encrypt the credentialValue as encIdentity.
	encIdentity, err := deriveAndEncrypt(pub, secret, name, credential2B)
	if err != nil {
		return nil, nil, err
	}

	// Compute the HMAC of (encIdentity || name)
	identityHMAC, err := deriveAndHMAC(pub, secret, nil, encIdentity, name)
	if err != nil {
		return nil, nil, err
	}

	// Marshal the virtual TPMS_ID_OBJECT ourselves. We have to do this since encIdentity's size is encrypted.
	idObject = make([]byte, 0, 2+len(identityHMAC)+len(encIdentity))
	idObject = append(idObject, Marshal(TPM2BDigest{Buffer: identityHMAC})...)
	idObject = append(idObject, encIdentity...)

	return idObject, ciphertext, nil
}
