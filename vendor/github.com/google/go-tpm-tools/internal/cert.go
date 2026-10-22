package internal

import (
	"bytes"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
)

const (
	maxIssuingCertificateURLs = 3
	maxCertChainLength        = 4
)

var knownRootKeyIDs = [][]byte{
	// tpm_ek_root_1.cer
	{0x65, 0xF4, 0xE4, 0xE6, 0xAA, 0xF6, 0xFD, 0x5A, 0xD2, 0x88, 0x9C, 0xA8, 0x53, 0x55, 0xF7, 0x00, 0x8E, 0x08, 0xF7, 0xA5},
	// gcp_ek_ak_ca_root.crt (Cloud CAS)
	{0x49, 0xE7, 0x4A, 0x5B, 0x56, 0x29, 0xF5, 0x9D, 0x79, 0xB7, 0xA6, 0x30, 0x3C, 0x03, 0xB2, 0x8F, 0xE7, 0x14, 0xDD, 0x4C},
}

// isIssuedByKnownRoot checks if the certificate was signed by a known Google Root CA
// using the AuthorityKeyId extension.
func isIssuedByKnownRoot(cert *x509.Certificate) bool {
	if cert == nil || len(cert.AuthorityKeyId) == 0 {
		return false
	}
	for _, rootKeyID := range knownRootKeyIDs {
		if bytes.Equal(cert.AuthorityKeyId, rootKeyID) {
			return true
		}
	}
	return false
}

// GetAKIntermediateCerts constructs the AK's intermediate certificate chain.
// If an error is encountered in the process, return a nil slice and the error.
// Stops before fetching the root certificate to avoid depending on the Root CA's
// region being available in the CS Client (since GCA separately gets the root
// certificate for verification).
func GetAKIntermediateCerts(cert *x509.Certificate, client *http.Client) ([][]byte, error) {
	var certs [][]byte
	currentCert := cert
	for len(certs) <= maxCertChainLength {
		if isIssuedByKnownRoot(currentCert) {
			return certs, nil
		}
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
			resp.Body.Close()
			continue
		}
		certBytes, err := io.ReadAll(resp.Body)
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
