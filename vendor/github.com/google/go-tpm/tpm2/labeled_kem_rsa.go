package tpm2

import (
	"crypto"
	"crypto/rsa"
	"errors"
	"fmt"
	"io"
	"strings"
)

var (
	// The source of randomness used for encapsulation ran out of data.
	ErrInsufficientRandom = errors.New("random source did not provide enough data")
)

// An rsaKey is an RSA-OAEP-based Labeled Encapsulation key.
type rsaKey struct {
	// The actual public key.
	rsaPub rsa.PublicKey
	// The scheme hash algorithm to use for the OAEP-based encapsulation.
	hash crypto.Hash
	// The name algorithm of the key.
	nameAlg TPMIAlgHash
	// The symmetric parameters of the key.
	symParms *TPMTSymDefObject
}

func importRSAEncapsulationKey(pub *TPMTPublic) (*rsaKey, error) {
	rsaParms, err := pub.Parameters.RSADetail()
	if err != nil {
		return nil, err
	}
	rsaPub, err := pub.Unique.RSA()
	if err != nil {
		return nil, err
	}
	rsa, err := RSAPub(rsaParms, rsaPub)
	if err != nil {
		return nil, err
	}

	// Decide what hash algorithm to use for OAEP.
	// It's the scheme hash algorithm if not null, otherwise it's the name algorithm.
	hashAlgID := pub.NameAlg
	if rsaParms.Scheme.Scheme == TPMAlgOAEP {
		oaep, err := rsaParms.Scheme.Details.OAEP()
		if err != nil {
			return nil, err
		}
		if oaep.HashAlg != TPMAlgNull {
			hashAlgID = oaep.HashAlg
		}
	}
	hashAlg, err := hashAlgID.Hash()
	if err != nil {
		return nil, err
	}

	return &rsaKey{
		rsaPub:   *rsa,
		hash:     hashAlg,
		nameAlg:  pub.NameAlg,
		symParms: &rsaParms.Symmetric,
	}, nil
}

// Encapsulate performs the OAEP-based RSA Labeled Encapsulation.
func (pub *rsaKey) Encapsulate(random io.Reader, label string) (secret []byte, ciphertext []byte, err error) {
	secret = make([]byte, pub.hash.Size())
	n, err := random.Read(secret)
	if err != nil {
		return nil, nil, err
	}
	if n != len(secret) {
		return nil, nil, fmt.Errorf("%w: only read %d bytes but %d were needed", ErrInsufficientRandom, n, len(secret))
	}

	ciphertext, err = pub.encapsulateDerandomized(random, secret, label)
	if err != nil {
		return nil, nil, err
	}
	return secret, ciphertext, err
}

// encapsulateDerandomized is a derandomized internal version of Encapsulate for testing.
func (pub *rsaKey) encapsulateDerandomized(oaepSaltReader io.Reader, secret []byte, label string) (ciphertext []byte, err error) {
	// Ensure label is null-terminated.
	if !strings.HasSuffix(label, "\x00") {
		label = label + "\x00"
	}

	if len(secret) != pub.hash.Size() {
		return nil, fmt.Errorf("%w: secret was only %d bytes but %d were needed", ErrInsufficientRandom, len(secret), pub.hash.Size())
	}

	ciphertext, err = rsa.EncryptOAEP(pub.hash.New(), oaepSaltReader, &pub.rsaPub, secret, []byte(label))
	if err != nil {
		return nil, err
	}
	return ciphertext, err
}

// NameAlg implements LabeledEncapsulationKey.
func (pub *rsaKey) NameAlg() TPMAlgID {
	return pub.nameAlg
}

// SymmetricParameters implements LabeledEncapsulationkey.
func (pub *rsaKey) SymmetricParameters() *TPMTSymDefObject {
	return pub.symParms
}
