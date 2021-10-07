package client

import (
	"fmt"

	pb "github.com/google/go-tpm-tools/proto/attest"
)

// AttestOpts allows for customizing the functionality of Attest.
type AttestOpts struct {
	// A unique, application-specific nonce used to guarantee freshness of the
	// attestation. This must not be empty, and should generally be long enough
	// to make brute force attacks infeasible.
	//
	// For security reasons, applications should not allow for attesting with
	// arbitrary, externally-provided nonces. The nonce should be prefixed or
	// otherwise bound (i.e. via a KDF) to application-specific data. For more
	// information on why this is an issue, see this paper on robust remote
	// attestation protocols:
	// https://citeseerx.ist.psu.edu/viewdoc/download?doi=10.1.1.70.4562&rep=rep1&type=pdf
	Nonce []byte
}

// Attest generates an Attestation containing the TCG Event Log and a Quote over
// all PCR banks. The provided nonce can be used to guarantee freshness of the
// attestation. This function will return an error if the key is not a
// restricted signing key.
//
// AttestOpts is used for additional configuration of the Attestation process.
// This is primarily used to pass the attestation's nonce:
//
//   attestation, err := key.Attest(client.AttestOpts{Nonce: my_nonce})
func (k *Key) Attest(opts AttestOpts) (*pb.Attestation, error) {
	if len(opts.Nonce) == 0 {
		return nil, fmt.Errorf("provided nonce must not be empty")
	}
	sels, err := implementedPCRs(k.rw)
	if err != nil {
		return nil, err
	}

	attestation := pb.Attestation{}
	if attestation.AkPub, err = k.PublicArea().Encode(); err != nil {
		return nil, fmt.Errorf("failed to encode public area: %w", err)
	}
	for _, sel := range sels {
		quote, err := k.Quote(sel, opts.Nonce)
		if err != nil {
			return nil, err
		}
		attestation.Quotes = append(attestation.Quotes, quote)
	}
	if attestation.EventLog, err = GetEventLog(k.rw); err != nil {
		return nil, fmt.Errorf("failed to retrieve TCG Event Log: %w", err)
	}
	return &attestation, nil
}
