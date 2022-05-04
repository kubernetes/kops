package client

import (
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"

	pb "github.com/google/go-tpm-tools/proto/attest"
)

const (
	maxIssuingCertificateURLs = 3
	maxCertChainLength        = 4
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
	// TCG Canonical Event Log to add to the attestation.
	// Currently, we only support PCR replay for PCRs orthogonal to those in the
	// firmware event log, where PCRs 0-9 and 14 are often measured. If the two
	// logs overlap, server-side verification using this library may fail.
	CanonicalEventLog []byte
	// If non-nil, will be used to fetch the AK certificate chain for validation.
	// Key.Attest() will construct the certificate chain by making GET requests to
	// the contents of Key.cert.IssuingCertificateURL using this client.
	CertChainFetcher *http.Client
}

// Given a certificate, iterates through its IssuingCertificateURLs and returns
// the certificate that signed it. If the certificate lacks an
// IssuingCertificateURL, return nil. If fetching the certificates fails or the
// cert chain is malformed, return an error.
func fetchIssuingCertificate(client *http.Client, cert *x509.Certificate) (*x509.Certificate, error) {
	// Check if we should event attempt fetching.
	if cert == nil || len(cert.IssuingCertificateURL) == 0 {
		return nil, nil
	}
	// For each URL, fetch and parse the certificate, then verify whether it signed cert.
	// If successful, return the parsed certificate. If any step in this process fails, try the next url.
	// If all the URLs fail, return the last error we got.
	// TODO(Issue #169): Return a multi-error here
	var lastErr error
	for i, url := range cert.IssuingCertificateURL {
		// Limit the number of attempts.
		if i >= maxIssuingCertificateURLs {
			break
		}
		resp, err := client.Get(url)
		if err != nil {
			lastErr = fmt.Errorf("failed to retrieve certificate at %v: %w", url, err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("certificate retrieval from %s returned non-OK status: %v", url, resp.StatusCode)
			continue
		}
		certBytes, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("failed to read response body from %s: %w", url, err)
			continue
		}

		parsedCert, err := x509.ParseCertificate(certBytes)
		if err != nil {
			lastErr = fmt.Errorf("failed to parse response from %s into a certificate: %w", url, err)
			continue
		}

		// Check if the parsed certificate signed the current one.
		if err = cert.CheckSignatureFrom(parsedCert); err != nil {
			lastErr = fmt.Errorf("parent certificate from %s did not sign child: %w", url, err)
			continue
		}
		return parsedCert, nil
	}
	return nil, lastErr
}

// Constructs the certificate chain for the key's certificate.
// If an error is encountered in the process, return what has been constructed so far.
func (k *Key) getCertificateChain(client *http.Client) ([][]byte, error) {
	var certs [][]byte
	currentCert := k.cert
	for len(certs) <= maxCertChainLength {
		issuingCert, err := fetchIssuingCertificate(client, currentCert)
		if err != nil {
			return nil, err
		}
		if issuingCert == nil {
			return certs, nil
		}
		certs = append(certs, issuingCert.Raw)
		currentCert = issuingCert
	}
	return nil, fmt.Errorf("max certificate chain length (%v) exceeded", maxCertChainLength)
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
	attestation.AkCert = k.CertDERBytes()
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
	if len(opts.CanonicalEventLog) != 0 {
		attestation.CanonicalEventLog = opts.CanonicalEventLog
	}

	// Attempt to construct certificate chain. fetchIssuingCertificate checks if
	// AK cert is present and contains intermediate cert URLs.
	if opts.CertChainFetcher != nil {
		attestation.IntermediateCerts, err = k.getCertificateChain(opts.CertChainFetcher)
		if err != nil {
			return nil, fmt.Errorf("fetching certificate chain: %w", err)
		}
	}

	return &attestation, nil
}
