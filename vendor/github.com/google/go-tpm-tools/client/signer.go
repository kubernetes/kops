package client

import (
	"crypto"
	"crypto/rsa"
	"encoding/asn1"
	"fmt"
	"io"
	"math/big"
	"sync"

	"github.com/google/go-tpm-tools/internal"
	"github.com/google/go-tpm/legacy/tpm2"
)

// Global mutex to protect against concurrent TPM access.
var signerMutex sync.Mutex

type tpmSigner struct {
	Key  *Key
	Hash crypto.Hash
}

// Public returns the tpmSigners public key.
func (signer *tpmSigner) Public() crypto.PublicKey {
	return signer.Key.PublicKey()
}

// Sign uses the TPM key to sign the digest.
// The digest must be hashed from the same hash algorithm as the keys scheme.
// The opts hash function must also match the keys scheme (or be nil).
// Concurrent use of Sign is thread safe, but it is not safe to access the TPM
// from other sources while Sign is executing.
// For RSAPSS signatures, you cannot specify custom salt lengths. The salt
// length will be (keyBits/8) - digestSize - 2, unless that is less than the
// digestSize in which case, saltLen will be digestSize. The only normal case
// where saltLen is not digestSize is when using 1024 keyBits with SHA512.
func (signer *tpmSigner) Sign(_ io.Reader, digest []byte, opts crypto.SignerOpts) (signature []byte, err error) {
	if pssOpts, ok := opts.(*rsa.PSSOptions); ok {
		if signer.Key.pubArea.RSAParameters == nil {
			return nil, fmt.Errorf("invalid options: PSSOptions can only be used with RSA keys")
		}
		if signer.Key.pubArea.RSAParameters.Sign.Alg != tpm2.AlgRSAPSS {
			return nil, fmt.Errorf("invalid options: PSSOptions cannot be used with signing alg: %v", signer.Key.pubArea.RSAParameters.Sign.Alg)
		}
		if pssOpts.SaltLength != rsa.PSSSaltLengthAuto {
			return nil, fmt.Errorf("salt length must be rsa.PSSSaltLengthAuto")
		}
	}
	if opts != nil && opts.HashFunc() != signer.Hash {
		return nil, fmt.Errorf("hash algorithm: got %v, want %v", opts.HashFunc(), signer.Hash)
	}
	if len(digest) != signer.Hash.Size() {
		return nil, fmt.Errorf("digest length: got %d, want %d", digest, signer.Hash.Size())
	}

	signerMutex.Lock()
	defer signerMutex.Unlock()

	auth, err := signer.Key.session.Auth()
	if err != nil {
		return nil, err
	}

	sig, err := tpm2.SignWithSession(signer.Key.rw, auth.Session, signer.Key.handle, "", digest, nil, nil)
	if err != nil {
		return nil, err
	}
	return getSignature(sig)
}

// GetSigner returns a crypto.Signer wrapping the loaded TPM Key.
// Concurrent use of one or more Signers is thread safe, but it is not safe to
// access the TPM from other sources while using a Signer.
// The returned Signer lasts the lifetime of the Key, and will no longer work
// once the Key has been closed.
func (k *Key) GetSigner() (crypto.Signer, error) {
	if k.hasAttribute(tpm2.FlagRestricted) {
		return nil, fmt.Errorf("restricted keys are not supported")
	}
	hashAlg, err := internal.GetSigningHashAlg(k.pubArea)
	if err != nil {
		return nil, err
	}
	// For crypto.Signer, Go does the hashing. Make sure the hash is supported.
	hash, err := hashAlg.Hash()
	if err != nil {
		return nil, err
	}
	return &tpmSigner{k, hash}, nil
}

// SignData signs a data buffer with a TPM loaded key. Unlike GetSigner, this
// method works with restricted and unrestricted keys. If this method is called
// on a restriced key, the TPM itself will hash the provided data, failing the
// signing operation if the data begins with TPM_GENERATED_VALUE.
func (k *Key) SignData(data []byte) ([]byte, error) {
	hashAlg, err := internal.GetSigningHashAlg(k.pubArea)
	if err != nil {
		return nil, err
	}

	var digest []byte
	var ticket *tpm2.Ticket
	if k.hasAttribute(tpm2.FlagRestricted) {
		// Restricted keys can only sign data hashed by the TPM. We use the
		// owner hierarchy for the Ticket, but any non-Null hierarchy would do.
		digest, ticket, err = tpm2.Hash(k.rw, hashAlg, data, tpm2.HandleOwner)
		if err != nil {
			return nil, err
		}
	} else {
		// Unrestricted keys can sign any digest, no need for TPM hashing.
		hash, err := hashAlg.Hash()
		if err != nil {
			return nil, err
		}
		hasher := hash.New()
		hasher.Write(data)
		digest = hasher.Sum(nil)
	}

	auth, err := k.session.Auth()
	if err != nil {
		return nil, err
	}
	sig, err := tpm2.SignWithSession(k.rw, auth.Session, k.handle, "", digest, ticket, nil)
	if err != nil {
		return nil, err
	}
	return getSignature(sig)
}

func getSignature(sig *tpm2.Signature) ([]byte, error) {
	switch sig.Alg {
	case tpm2.AlgRSASSA:
		return sig.RSA.Signature, nil
	case tpm2.AlgRSAPSS:
		return sig.RSA.Signature, nil
	case tpm2.AlgECDSA:
		sigStruct := struct{ R, S *big.Int }{sig.ECC.R, sig.ECC.S}
		return asn1.Marshal(sigStruct)
	default:
		return nil, fmt.Errorf("unsupported signing algorithm: %v", sig.Alg)
	}
}
