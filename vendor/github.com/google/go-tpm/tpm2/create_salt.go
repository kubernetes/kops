package tpm2

import "io"

var (
	labelSecret = "SECRET"
)

// CreateEncryptedSalt encrypts a random salt for secure session establishment.
func CreateEncryptedSalt(rand io.Reader, pub LabeledEncapsulationKey) (salt []byte, encSecret []byte, err error) {
	// The salt value is directly used from the Labeled Key Encapsulation operation.
	// See Part 1, "Salted and Bound Session Key Generation"
	return pub.Encapsulate(rand, labelSecret)
}
