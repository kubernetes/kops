package client

import (
	"crypto/sha256"

	"github.com/google/go-tpm/legacy/tpm2"
	"github.com/google/go-tpm/tpmutil"
)

// Calculations from Credential_Profile_EK_V2.0, section 2.1.5.3 - authPolicy
func defaultEKAuthPolicy() []byte {
	buf, err := tpmutil.Pack(tpm2.CmdPolicySecret, tpm2.HandleEndorsement)
	if err != nil {
		panic(err)
	}
	digest1 := sha256.Sum256(append(make([]byte, 32), buf...))
	// We would normally append the policy buffer to digest1, but the
	// policy buffer is empty for the default Auth Policy.
	digest2 := sha256.Sum256(digest1[:])
	return digest2[:]
}

func defaultEKAttributes() tpm2.KeyProp {
	// The EK is a storage key that must use session-based authorization.
	return (tpm2.FlagStorageDefault | tpm2.FlagAdminWithPolicy) & ^tpm2.FlagUserWithAuth
}

func defaultSRKAttributes() tpm2.KeyProp {
	// FlagNoDA doesn't do anything (as the AuthPolicy is nil). However, this is
	// what Windows does, and we don't want to conflict.
	return tpm2.FlagStorageDefault | tpm2.FlagNoDA
}

func defaultSymScheme() *tpm2.SymScheme {
	return &tpm2.SymScheme{
		Alg:     tpm2.AlgAES,
		KeyBits: 128,
		Mode:    tpm2.AlgCFB,
	}
}

func defaultRSAParams() *tpm2.RSAParams {
	return &tpm2.RSAParams{
		Symmetric:  defaultSymScheme(),
		KeyBits:    2048,
		ModulusRaw: make([]byte, 256), // public.unique must be all zeros
	}
}

func defaultECCParams() *tpm2.ECCParams {
	return &tpm2.ECCParams{
		Symmetric: defaultSymScheme(),
		CurveID:   tpm2.CurveNISTP256,
		Point: tpm2.ECPoint{
			XRaw: make([]byte, 32),
			YRaw: make([]byte, 32),
		},
	}
}

// DefaultEKTemplateRSA returns the default Endorsement Key (EK) template as
// specified in Credential_Profile_EK_V2.0, section 2.1.5.1 - authPolicy.
// https://trustedcomputinggroup.org/wp-content/uploads/Credential_Profile_EK_V2.0_R14_published.pdf
func DefaultEKTemplateRSA() tpm2.Public {
	return tpm2.Public{
		Type:          tpm2.AlgRSA,
		NameAlg:       tpm2.AlgSHA256,
		Attributes:    defaultEKAttributes(),
		AuthPolicy:    defaultEKAuthPolicy(),
		RSAParameters: defaultRSAParams(),
	}
}

// DefaultEKTemplateECC returns the default Endorsement Key (EK) template as
// specified in Credential_Profile_EK_V2.0, section 2.1.5.2 - authPolicy.
// https://trustedcomputinggroup.org/wp-content/uploads/Credential_Profile_EK_V2.0_R14_published.pdf
func DefaultEKTemplateECC() tpm2.Public {
	return tpm2.Public{
		Type:          tpm2.AlgECC,
		NameAlg:       tpm2.AlgSHA256,
		Attributes:    defaultEKAttributes(),
		AuthPolicy:    defaultEKAuthPolicy(),
		ECCParameters: defaultECCParams(),
	}
}

// AKTemplateRSA returns a potential Attestation Key (AK) template.
// This is very similar to DefaultEKTemplateRSA, except that this will be a
// signing key instead of an encrypting key.
func AKTemplateRSA() tpm2.Public {
	return tpm2.Public{
		Type:       tpm2.AlgRSA,
		NameAlg:    tpm2.AlgSHA256,
		Attributes: tpm2.FlagSignerDefault,
		RSAParameters: &tpm2.RSAParams{
			Sign: &tpm2.SigScheme{
				Alg:  tpm2.AlgRSASSA,
				Hash: tpm2.AlgSHA256,
			},
			KeyBits: 2048,
		},
	}
}

// AKTemplateECC returns a potential Attestation Key (AK) template.
// This is very similar to DefaultEKTemplateECC, except that this will be a
// signing key instead of an encrypting key.
func AKTemplateECC() tpm2.Public {
	params := defaultECCParams()
	params.Symmetric = nil
	params.Sign = &tpm2.SigScheme{
		Alg:  tpm2.AlgECDSA,
		Hash: tpm2.AlgSHA256,
	}
	return tpm2.Public{
		Type:          tpm2.AlgECC,
		NameAlg:       tpm2.AlgSHA256,
		Attributes:    tpm2.FlagSignerDefault,
		ECCParameters: params,
	}
}

// SRKTemplateRSA returns a standard Storage Root Key (SRK) template.
// This is based upon the advice in the TCG's TPM v2.0 Provisioning Guidance.
func SRKTemplateRSA() tpm2.Public {
	return tpm2.Public{
		Type:          tpm2.AlgRSA,
		NameAlg:       tpm2.AlgSHA256,
		Attributes:    defaultSRKAttributes(),
		RSAParameters: defaultRSAParams(),
	}
}

// SRKTemplateECC returns a standard Storage Root Key (SRK) template.
// This is based upon the advice in the TCG's TPM v2.0 Provisioning Guidance.
func SRKTemplateECC() tpm2.Public {
	return tpm2.Public{
		Type:          tpm2.AlgECC,
		NameAlg:       tpm2.AlgSHA256,
		Attributes:    defaultSRKAttributes(),
		ECCParameters: defaultECCParams(),
	}
}
