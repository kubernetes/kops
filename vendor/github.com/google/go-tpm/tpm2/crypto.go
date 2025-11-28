package tpm2

import (
	"crypto"
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"fmt"
	"math/big"
)

// Priv converts a TPM private key into one recognized by the crypto package.
func Priv(public TPMTPublic, sensitive TPMTSensitive) (crypto.PrivateKey, error) {

	var privateKey crypto.PrivateKey

	publicKey, err := Pub(public)
	if err != nil {
		return nil, err
	}

	switch public.Type {
	case TPMAlgRSA:
		publicKey := publicKey.(*rsa.PublicKey)

		if sensitive.SensitiveType != TPMAlgRSA {
			return nil, fmt.Errorf("sensitive type is not equal to public type")
		}

		prime, err := sensitive.Sensitive.RSA()
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve the RSA prime number")
		}

		P := new(big.Int).SetBytes(prime.Buffer)
		Q := new(big.Int).Div(publicKey.N, P)
		phiN := new(big.Int).Mul(new(big.Int).Sub(P, big.NewInt(1)), new(big.Int).Sub(Q, big.NewInt(1)))
		D := new(big.Int).ModInverse(big.NewInt(int64(publicKey.E)), phiN)

		rsaKey := &rsa.PrivateKey{
			PublicKey: *publicKey,
			D:         D,
			Primes:    []*big.Int{P, Q},
		}
		rsaKey.Precompute()

		privateKey = rsaKey
	case TPMAlgECC:
		publicKey := publicKey.(*ecdsa.PublicKey)

		if sensitive.SensitiveType != TPMAlgECC {
			return nil, fmt.Errorf("sensitive type is not equal to public type")
		}

		d, err := sensitive.Sensitive.ECC()
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve the ECC")
		}

		D := new(big.Int).SetBytes(d.Buffer)

		ecdsaKey := &ecdsa.PrivateKey{
			PublicKey: *publicKey,
			D:         D,
		}

		privateKey = ecdsaKey
	default:
		return nil, fmt.Errorf("unsupported public key type: %v", public.Type)
	}

	return privateKey, nil
}

// Pub converts a TPM public key into one recognized by the crypto package.
func Pub(public TPMTPublic) (crypto.PublicKey, error) {
	var publicKey crypto.PublicKey

	switch public.Type {
	case TPMAlgRSA:
		parameters, err := public.Parameters.RSADetail()
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve the RSA parameters")
		}

		n, err := public.Unique.RSA()
		if err != nil {
			return nil, fmt.Errorf("failed to parse and retrieve the RSA modulus")
		}

		publicKey, err = RSAPub(parameters, n)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve the RSA public key")
		}
	case TPMAlgECC:
		parameters, err := public.Parameters.ECCDetail()
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve the ECC parameters")
		}

		pub, err := public.Unique.ECC()
		if err != nil {
			return nil, fmt.Errorf("failed to parse and retrieve the ECC point")
		}

		publicKey, err = ECDSAPub(parameters, pub)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve the ECC public key")
		}
	default:
		return nil, fmt.Errorf("unsupported public key type: %v", public.Type)
	}

	return publicKey, nil
}

// RSAPub converts a TPM RSA public key into one recognized by the rsa package.
func RSAPub(parms *TPMSRSAParms, pub *TPM2BPublicKeyRSA) (*rsa.PublicKey, error) {
	result := rsa.PublicKey{
		N: big.NewInt(0).SetBytes(pub.Buffer),
		E: int(parms.Exponent),
	}
	// TPM considers 65537 to be the default RSA public exponent, and 0 in
	// the parms
	// indicates so.
	if result.E == 0 {
		result.E = 65537
	}
	return &result, nil
}

// ECDSAPub converts a TPM ECC public key into one recognized by the ecdh package
func ECDSAPub(parms *TPMSECCParms, pub *TPMSECCPoint) (*ecdsa.PublicKey, error) {

	var c elliptic.Curve
	switch parms.CurveID {
	case TPMECCNistP256:
		c = elliptic.P256()
	case TPMECCNistP384:
		c = elliptic.P384()
	case TPMECCNistP521:
		c = elliptic.P521()
	default:
		return nil, fmt.Errorf("unknown curve: %v", parms.CurveID)
	}

	pubKey := ecdsa.PublicKey{
		Curve: c,
		X:     big.NewInt(0).SetBytes(pub.X.Buffer),
		Y:     big.NewInt(0).SetBytes(pub.Y.Buffer),
	}

	return &pubKey, nil
}

// ECDHPub converts a TPM ECC public key into one recognized by the ecdh package
func ECDHPub(parms *TPMSECCParms, pub *TPMSECCPoint) (*ecdh.PublicKey, error) {

	pubKey, err := ECDSAPub(parms, pub)
	if err != nil {
		return nil, err
	}

	return pubKey.ECDH()
}

// ECCPoint returns an uncompressed ECC Point
func ECCPoint(pubKey *ecdh.PublicKey) (*big.Int, *big.Int, error) {
	b := pubKey.Bytes()
	size, err := elementLength(pubKey.Curve())
	if err != nil {
		return nil, nil, fmt.Errorf("ECCPoint: %w", err)
	}
	return big.NewInt(0).SetBytes(b[1 : size+1]),
		big.NewInt(0).SetBytes(b[size+1:]), nil
}

func elementLength(c ecdh.Curve) (int, error) {
	switch c {
	case ecdh.P256():
		// crypto/internal/nistec/fiat.p256ElementLen
		return 32, nil
	case ecdh.P384():
		// crypto/internal/nistec/fiat.p384ElementLen
		return 48, nil
	case ecdh.P521():
		// crypto/internal/nistec/fiat.p521ElementLen
		return 66, nil
	default:
		return 0, fmt.Errorf("unknown element length for curve: %v", c)
	}
}
