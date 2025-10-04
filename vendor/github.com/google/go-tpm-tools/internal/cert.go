package internal

import (
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
)

const (
	maxIssuingCertificateURLs = 3
	maxCertChainLength        = 4
)

// GetCertificateChain constructs the certificate chain for the key's certificate.
// If an error is encountered in the process, return what has been constructed so far.
func GetCertificateChain(cert *x509.Certificate, client *http.Client) ([][]byte, error) {
	var certs [][]byte
	currentCert := cert
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
