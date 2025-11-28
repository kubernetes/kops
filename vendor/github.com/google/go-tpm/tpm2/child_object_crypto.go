package tpm2

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"errors"
	"fmt"
)

var (
	labelStorage   = "STORAGE"
	labelIntegrity = "INTEGRITY"

	ErrUnsupportedCipher = errors.New("unsupported block cipher")
	ErrUnsupportedMode   = errors.New("unsupported block cipher mode of operation")
)

// deriveAndEncrypt derives a symmetric key and uses it to encrypt the plaintext.
func deriveAndEncrypt(pub LabeledEncapsulationKey, seed []byte, context []byte, plaintext []byte) ([]byte, error) {
	// Only AES is supported.
	if pub.SymmetricParameters().Algorithm != TPMAlgAES {
		return nil, fmt.Errorf("%w %v", ErrUnsupportedCipher, pub.SymmetricParameters().Algorithm)
	}
	mode, err := pub.SymmetricParameters().Mode.AES()
	if err != nil {
		return nil, err
	}
	if *mode != TPMAlgCFB {
		return nil, fmt.Errorf("%w %v", ErrUnsupportedMode, *mode)
	}
	bits, err := pub.SymmetricParameters().KeyBits.AES()
	if err != nil {
		return nil, err
	}

	hash, err := pub.NameAlg().Hash()
	if err != nil {
		return nil, err
	}
	key, err := aes.NewCipher(KDFa(hash, seed, labelStorage, context, nil, int(*bits)))
	if err != nil {
		return nil, err
	}
	ciphertext := make([]byte, len(plaintext))
	cipher.NewCFBEncrypter(key, make([]byte, key.BlockSize())).XORKeyStream(ciphertext, plaintext)
	return ciphertext, nil
}

// deriveAndHMAC derives an HMAC key and uses it to HMAC the data, which can be provided in multiple chunks.
func deriveAndHMAC(pub LabeledEncapsulationKey, seed []byte, data ...[]byte) ([]byte, error) {
	hash, err := pub.NameAlg().Hash()
	if err != nil {
		return nil, err
	}
	key := KDFa(hash, seed, labelIntegrity, nil, nil, hash.Size()*8)
	hmac := hmac.New(hash.New, key)
	for _, data := range data {
		hmac.Write(data)
	}
	return hmac.Sum(nil), nil
}
