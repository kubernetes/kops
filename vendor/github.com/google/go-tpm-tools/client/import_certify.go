package client

import (
	"fmt"

	tpb "github.com/google/go-tpm-tools/proto/tpm"
	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpm2/transport"
)

// This file aims to implement the attester side of https://trustedcomputinggroup.org/wp-content/uploads/EK-Based-Key-Attestation-with-TPM-Firmware-Version-V1-RC1_9July2025.pdf#page=8
// For reference: https://github.com/TrustedComputingGroup/tpm-fw-attestation-reference-code

func ekResponse(tpm transport.TPM) (*tpm2.CreatePrimaryResponse, error) {
	// SVSM currently only supports attesting an RSA EK.
	// We may parameterize this later for more options.
	return tpm2.CreatePrimary{
		PrimaryHandle: tpm2.TPMRHEndorsement,
		InPublic:      tpm2.New2B(tpm2.RSAEKTemplate),
	}.Execute(tpm)
}

func makeAK(tpm transport.TPM, keyAlgo tpm2.TPMAlgID) (*tpm2.CreatePrimaryResponse, error) {
	var public []byte
	var err error
	switch keyAlgo {
	case tpm2.TPMAlgECC:
		public, err = AKTemplateECC().Encode()
	case tpm2.TPMAlgRSA:
		public, err = AKTemplateRSA().Encode()
	default:
		return nil, fmt.Errorf("unsupported keyAlgo %v", keyAlgo)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create AK: %w", err)
	}
	cp, err := tpm2.CreatePrimary{
		PrimaryHandle: tpm2.TPMRHOwner,
		InPublic:      tpm2.BytesAs2B[tpm2.TPMTPublic](public),
	}.Execute(tpm)
	if err != nil {
		return nil, err
	}
	return cp, nil
}

// CreateCertifiedAKBlob creates an AK and certifies it, thus solving the TPM registration challenge.
func CreateCertifiedAKBlob(tpm transport.TPM, req *tpb.ImportBlob, keyAlgo tpm2.TPMAlgID) (*tpb.CertifiedBlob, error) {
	ek, err := ekResponse(tpm)
	if err != nil {
		return nil, fmt.Errorf("failed to create RSA EK: %w", err)
	}

	// Import the restricted HMAC key.
	imported, err := tpm2.Import{
		ParentHandle: tpm2.AuthHandle{
			Handle: ek.ObjectHandle,
			Name:   ek.Name,
			Auth:   tpm2.Policy(tpm2.TPMAlgSHA256, 32, ekPolicy),
		},
		ObjectPublic: tpm2.BytesAs2B[tpm2.TPMTPublic](req.GetPublicArea()),
		Duplicate:    tpm2.TPM2BPrivate{Buffer: req.GetDuplicate()},
		InSymSeed:    tpm2.TPM2BEncryptedSecret{Buffer: req.GetEncryptedSeed()},
	}.Execute(tpm)
	if err != nil {
		tpm2.FlushContext{
			FlushHandle: ek.ObjectHandle,
		}.Execute(tpm)
		return nil, fmt.Errorf("failed to import blob: %w", err)
	}

	// Load the imported HMAC key.
	loaded, err := tpm2.Load{
		ParentHandle: tpm2.AuthHandle{
			Handle: ek.ObjectHandle,
			Name:   ek.Name,
			Auth:   tpm2.Policy(tpm2.TPMAlgSHA256, 32, ekPolicy),
		},
		InPublic:  tpm2.BytesAs2B[tpm2.TPMTPublic](req.GetPublicArea()),
		InPrivate: imported.OutPrivate,
	}.Execute(tpm)
	// Flush before checking error and potentially early returning since we need to flush in both situations.
	tpm2.FlushContext{
		FlushHandle: ek.ObjectHandle,
	}.Execute(tpm)
	if err != nil {
		return nil, fmt.Errorf("failed to load HMAC: %w", err)
	}

	defer tpm2.FlushContext{
		FlushHandle: loaded.ObjectHandle,
	}.Execute(tpm)

	ak, err := makeAK(tpm, keyAlgo)
	if err != nil {
		return nil, err
	}
	defer tpm2.FlushContext{
		FlushHandle: ak.ObjectHandle,
	}.Execute(tpm)

	// Certify a newly created AK.
	certified, err := tpm2.Certify{
		ObjectHandle: tpm2.NamedHandle{
			Handle: ak.ObjectHandle,
			Name:   ak.Name,
		},
		SignHandle: tpm2.NamedHandle{
			Handle: loaded.ObjectHandle,
			Name:   loaded.Name,
		},
	}.Execute(tpm)
	if err != nil {
		return nil, fmt.Errorf("failed to certify blob: %w", err)
	}

	return &tpb.CertifiedBlob{
		PubArea:     ak.OutPublic.Bytes(),
		CertifyInfo: certified.CertifyInfo.Bytes(),
		RawSig:      tpm2.Marshal(certified.Signature),
	}, nil
}

func ekPolicy(t transport.TPM, handle tpm2.TPMISHPolicy, nonceTPM tpm2.TPM2BNonce) error {
	cmd := tpm2.PolicySecret{
		AuthHandle:    tpm2.TPMRHEndorsement,
		PolicySession: handle,
		NonceTPM:      nonceTPM,
	}
	_, err := cmd.Execute(t)
	return err
}
