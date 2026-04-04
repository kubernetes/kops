/*
Copyright 2026 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package azure

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/smallstep/pkcs7"
	expirationcache "k8s.io/client-go/tools/cache"
)

// roundTripperFunc adapts a function into an http.RoundTripper for tests.
type roundTripperFunc func(*http.Request) (*http.Response, error)

// RoundTrip implements http.RoundTripper.
func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// newTestIntermediateCertCaches returns fresh positive and negative TTL
// caches so tests run independently from the package-level caches.
func newTestIntermediateCertCaches(positiveTTL, negativeTTL time.Duration) (expirationcache.Store, expirationcache.Store) {
	return expirationcache.NewTTLStore(intermediateCertCacheEntryKeyFunc, positiveTTL),
		expirationcache.NewTTLStore(intermediateCertCacheEntryKeyFunc, negativeTTL)
}

// testPKI generates a CA certificate, a leaf certificate, and their keys for test PKCS7 signing.
func testPKI(tb testing.TB) (*x509.Certificate, *rsa.PrivateKey, *x509.Certificate, *rsa.PrivateKey) {
	tb.Helper()

	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		tb.Fatalf("generating CA key: %v", err)
	}
	caTemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Test Root CA"},
		SubjectKeyId:          []byte("test-ca-ski"),
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	caCertDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		tb.Fatalf("creating CA cert: %v", err)
	}
	caCert, err := x509.ParseCertificate(caCertDER)
	if err != nil {
		tb.Fatalf("parsing CA cert: %v", err)
	}

	leafKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		tb.Fatalf("generating leaf key: %v", err)
	}
	leafTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: "Test Signer"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		DNSNames:     []string{"metadata.azure.com"},
	}
	leafCertDER, err := x509.CreateCertificate(rand.Reader, leafTemplate, caCert, &leafKey.PublicKey, caKey)
	if err != nil {
		tb.Fatalf("creating leaf cert: %v", err)
	}
	leafCert, err := x509.ParseCertificate(leafCertDER)
	if err != nil {
		tb.Fatalf("parsing leaf cert: %v", err)
	}

	return caCert, caKey, leafCert, leafKey
}

// testPKIChain generates a root CA, an intermediate CA, and a leaf signer for
// tests that need to distinguish embedded-chain verification from fetch fallback.
func testPKIChain(tb testing.TB) (*x509.Certificate, *x509.Certificate, *x509.Certificate, *rsa.PrivateKey) {
	tb.Helper()

	rootKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		tb.Fatalf("generating root key: %v", err)
	}
	rootTemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(10),
		Subject:               pkix.Name{CommonName: "Test Root CA"},
		SubjectKeyId:          []byte("test-root-ski"),
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	rootCertDER, err := x509.CreateCertificate(rand.Reader, rootTemplate, rootTemplate, &rootKey.PublicKey, rootKey)
	if err != nil {
		tb.Fatalf("creating root cert: %v", err)
	}
	rootCert, err := x509.ParseCertificate(rootCertDER)
	if err != nil {
		tb.Fatalf("parsing root cert: %v", err)
	}

	intermediateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		tb.Fatalf("generating intermediate key: %v", err)
	}
	intermediateTemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(11),
		Subject:               pkix.Name{CommonName: "Test Intermediate CA"},
		SubjectKeyId:          []byte("test-intermediate-ski"),
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	intermediateCertDER, err := x509.CreateCertificate(rand.Reader, intermediateTemplate, rootCert, &intermediateKey.PublicKey, rootKey)
	if err != nil {
		tb.Fatalf("creating intermediate cert: %v", err)
	}
	intermediateCert, err := x509.ParseCertificate(intermediateCertDER)
	if err != nil {
		tb.Fatalf("parsing intermediate cert: %v", err)
	}

	leafKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		tb.Fatalf("generating leaf key: %v", err)
	}
	leafTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(12),
		Subject:      pkix.Name{CommonName: "Test Signer"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		DNSNames:     []string{"metadata.azure.com"},
	}
	leafCertDER, err := x509.CreateCertificate(rand.Reader, leafTemplate, intermediateCert, &leafKey.PublicKey, intermediateKey)
	if err != nil {
		tb.Fatalf("creating leaf cert: %v", err)
	}
	leafCert, err := x509.ParseCertificate(leafCertDER)
	if err != nil {
		tb.Fatalf("parsing leaf cert: %v", err)
	}

	return rootCert, intermediateCert, leafCert, leafKey
}

// createTestPKCS7 creates a PKCS7 SignedData containing the given content, signed by the leaf cert.
func createTestPKCS7(tb testing.TB, content []byte, leafCert *x509.Certificate, leafKey *rsa.PrivateKey, parents ...*x509.Certificate) []byte {
	tb.Helper()

	sd, err := pkcs7.NewSignedData(content)
	if err != nil {
		tb.Fatalf("creating signed data: %v", err)
	}
	if len(parents) == 0 {
		if err := sd.AddSigner(leafCert, leafKey, pkcs7.SignerInfoConfig{}); err != nil {
			tb.Fatalf("adding signer: %v", err)
		}
	} else {
		if err := sd.AddSignerChain(leafCert, leafKey, parents, pkcs7.SignerInfoConfig{}); err != nil {
			tb.Fatalf("adding signer chain: %v", err)
		}
	}
	derBytes, err := sd.Finish()
	if err != nil {
		tb.Fatalf("finishing signed data: %v", err)
	}
	return derBytes
}

// testSignature signs attested data and returns the PKCS7 as base64 text.
func testSignature(tb testing.TB, data attestedData, leafCert *x509.Certificate, leafKey *rsa.PrivateKey, caCert *x509.Certificate) string {
	tb.Helper()
	content, _ := json.Marshal(data)
	pkcs7DER := createTestPKCS7(tb, content, leafCert, leafKey, caCert)
	return base64.StdEncoding.EncodeToString(pkcs7DER)
}

// formatMicrosoftAttestedTime formats a time the way Azure IMDS actually
// emits it: with a "-0000" UTC offset instead of Go's default "+0000".
func formatMicrosoftAttestedTime(ts time.Time) string {
	return strings.TrimSuffix(ts.UTC().Format(attestedDocumentTimeFormat), "+0000") + "-0000"
}

// TestNonceForBody verifies that the nonce derivation is deterministic,
// produces the expected hex length, and varies with input.
func TestNonceForBody(t *testing.T) {
	testCases := []struct {
		name  string
		body  []byte
		other []byte
	}{
		{"non-nil body", []byte("test-body"), []byte("other-body")},
		{"nil body", nil, []byte("different")},
		{"empty body", []byte(""), nil},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			nonce := nonceForBody(tc.body)
			if len(nonce) != attestedDocumentNonceLength {
				t.Fatalf("nonce length: got %d, want %d", len(nonce), attestedDocumentNonceLength)
			}
			if nonce != nonceForBody(tc.body) {
				t.Fatal("nonce is not deterministic")
			}
			if tc.other != nil && nonce == nonceForBody(tc.other) {
				t.Fatal("different bodies produced the same nonce")
			}
		})
	}
}

// TestMicrosoftIntermediateCandidateURLs verifies that only normalized
// Microsoft PKI AIA URLs are returned.
func TestMicrosoftIntermediateCandidateURLs(t *testing.T) {
	signer := &x509.Certificate{
		IssuingCertificateURL: []string{
			"http://www.microsoft.com/pkiops/certs/Completely%20New%20Azure%20Metadata%20Issuing%20CA%2042%20-%20xsign.crt",
			"https://www.microsoft.com/pkiops/certs/Completely%20New%20Azure%20Metadata%20Issuing%20CA%2042.crt?ignored=1",
			"https://www.microsoft.com.evil.test/pkiops/certs/not-allowed.crt",
		},
	}

	got, err := microsoftIntermediateCandidateURLs(microsoftIntermediateCertBaseURL, signer)
	if err != nil {
		t.Fatalf("collecting candidate URLs: %v", err)
	}

	want := []string{
		"https://www.microsoft.com/pkiops/certs/Completely%20New%20Azure%20Metadata%20Issuing%20CA%2042%20-%20xsign.crt",
	}
	if !slices.Equal(got, want) {
		t.Fatalf("candidate URLs mismatch: got %v, want %v", got, want)
	}
}

// TestMicrosoftIntermediateCandidateURLs_RejectsNonMicrosoftPKIURLs verifies
// that out-of-scope AIA URLs are rejected.
func TestMicrosoftIntermediateCandidateURLs_RejectsNonMicrosoftPKIURLs(t *testing.T) {
	signer := &x509.Certificate{
		IssuingCertificateURL: []string{
			"https://www.microsoft.com.evil.test/pkiops/certs/not-allowed.crt",
			"https://www.microsoft.com@evil.test/pkiops/certs/not-allowed.crt",
			"https://www.microsoft.com/pkiops/other/not-allowed.crt",
		},
	}

	_, err := microsoftIntermediateCandidateURLs(microsoftIntermediateCertBaseURL, signer)
	if err == nil {
		t.Fatal("expected error for non-Microsoft PKI AIA URLs")
	}
}

// TestFetchIntermediateCertsFromBaseURL_UsesValidatedSignerAIA verifies that
// fetching uses only normalized, validated AIA URLs.
func TestFetchIntermediateCertsFromBaseURL_UsesValidatedSignerAIA(t *testing.T) {
	caCert, _, _, _ := testPKI(t)
	signer := &x509.Certificate{
		RawIssuer:      caCert.RawSubject,
		AuthorityKeyId: caCert.SubjectKeyId,
		IssuingCertificateURL: []string{
			"http://example.test/pkiops/certs/allowed.crt",
			"http://127.0.0.1:1/not-used.crt",
		},
	}

	var gotURLs []string
	client := &http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			gotURLs = append(gotURLs, req.URL.String())
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(caCert.Raw)),
				Header:     make(http.Header),
			}, nil
		}),
	}

	baseURL := "https://example.test/pkiops/certs"
	pool, err := fetchIntermediateCertsFromBaseURL(client, baseURL, signer)
	if err != nil {
		t.Fatalf("fetching intermediate certificates: %v", err)
	}
	if pool == nil {
		t.Fatal("expected intermediate certificate pool")
	}

	wantURLs := []string{
		"https://example.test/pkiops/certs/allowed.crt",
	}
	if !slices.Equal(gotURLs, wantURLs) {
		t.Fatalf("requested URLs mismatch: got %v, want %v", gotURLs, wantURLs)
	}
}

// TestFetchIntermediateCertsFromBaseURL_RejectsNonMatchingIntermediate verifies
// that fetched certificates are ignored unless they match the signer's issuer identity.
func TestFetchIntermediateCertsFromBaseURL_RejectsNonMatchingIntermediate(t *testing.T) {
	expectedIssuer, _, _, _ := testPKI(t)
	_, otherIssuer, _, _ := testPKIChain(t)

	signer := &x509.Certificate{
		RawIssuer:      expectedIssuer.RawSubject,
		AuthorityKeyId: expectedIssuer.SubjectKeyId,
		IssuingCertificateURL: []string{
			"https://example.test/pkiops/certs/not-the-issuer.crt",
		},
	}

	client := &http.Client{
		Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(otherIssuer.Raw)),
				Header:     make(http.Header),
			}, nil
		}),
	}

	_, err := fetchIntermediateCertsFromBaseURL(client, "https://example.test/pkiops/certs", signer)
	if err == nil {
		t.Fatal("expected error for non-matching fetched intermediate")
	}
}

// TestFetchIntermediateCertsFromBaseURL_KeepsOnlyMatchingIntermediate verifies
// that when multiple AIA URLs return different certs, only the one that matches
// the signer's issuer identity ends up in the pool.
func TestFetchIntermediateCertsFromBaseURL_KeepsOnlyMatchingIntermediate(t *testing.T) {
	matchingCA, _, leafCert, _ := testPKI(t)
	_, otherIssuer, _, _ := testPKIChain(t)

	signer := &x509.Certificate{
		RawIssuer:      matchingCA.RawSubject,
		AuthorityKeyId: matchingCA.SubjectKeyId,
		IssuingCertificateURL: []string{
			"https://example.test/pkiops/certs/first.crt",
			"https://example.test/pkiops/certs/second.crt",
		},
	}

	calls := 0
	client := &http.Client{
		Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
			calls++
			var body []byte
			if calls == 1 {
				body = otherIssuer.Raw // first URL returns wrong cert
			} else {
				body = matchingCA.Raw // second URL returns the right cert
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(body)),
				Header:     make(http.Header),
			}, nil
		}),
	}

	pool, err := fetchIntermediateCertsFromBaseURL(client, "https://example.test/pkiops/certs", signer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 2 {
		t.Fatalf("expected both URLs to be fetched, got %d calls", calls)
	}
	// Pool should contain matchingCA (not otherIssuer). Verify by using the pool
	// as roots and verifying the leaf signed by matchingCA — succeeds iff the
	// matching CA is in the pool.
	if _, err := leafCert.Verify(x509.VerifyOptions{
		Roots:     pool,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	}); err != nil {
		t.Errorf("expected pool to contain matching CA: %v", err)
	}
}

// TestValidateFetchedIntermediateForSigner exercises each per-branch
// reject/accept rule for the fetched intermediate identity check.
func TestValidateFetchedIntermediateForSigner(t *testing.T) {
	caCert, _, _, _ := testPKI(t)

	signerMatching := &x509.Certificate{
		RawIssuer:      caCert.RawSubject,
		AuthorityKeyId: caCert.SubjectKeyId,
	}

	t.Run("accepts matching CA", func(t *testing.T) {
		if err := validateFetchedIntermediateForSigner(signerMatching, caCert); err != nil {
			t.Fatalf("unexpected rejection: %v", err)
		}
	})

	t.Run("rejects non-CA cert", func(t *testing.T) {
		nonCA := &x509.Certificate{
			RawSubject:   caCert.RawSubject,
			SubjectKeyId: caCert.SubjectKeyId,
			IsCA:         false,
		}
		if err := validateFetchedIntermediateForSigner(signerMatching, nonCA); err == nil {
			t.Fatal("expected rejection of non-CA certificate")
		}
	})

	t.Run("rejects subject mismatch", func(t *testing.T) {
		signer := &x509.Certificate{
			RawIssuer:      []byte("different-issuer"),
			AuthorityKeyId: caCert.SubjectKeyId,
		}
		if err := validateFetchedIntermediateForSigner(signer, caCert); err == nil {
			t.Fatal("expected rejection when cert subject does not match signer issuer")
		}
	})

	t.Run("rejects SKI/AKI mismatch", func(t *testing.T) {
		signer := &x509.Certificate{
			RawIssuer:      caCert.RawSubject,
			AuthorityKeyId: []byte("different-akid"),
		}
		if err := validateFetchedIntermediateForSigner(signer, caCert); err == nil {
			t.Fatal("expected rejection when cert SKI does not match signer AKI")
		}
	})

	t.Run("rejects nil signer", func(t *testing.T) {
		if err := validateFetchedIntermediateForSigner(nil, caCert); err == nil {
			t.Fatal("expected error for nil signer")
		}
	})

	t.Run("rejects nil cert", func(t *testing.T) {
		if err := validateFetchedIntermediateForSigner(signerMatching, nil); err == nil {
			t.Fatal("expected error for nil cert")
		}
	})
}

// TestFetchCertificate_RejectsBadResponse exercises the rejection paths of
// fetchCertificate: oversized body, non-200 status, and unparseable DER.
func TestFetchCertificate_RejectsBadResponse(t *testing.T) {
	testCases := []struct {
		name   string
		status int
		body   []byte
	}{
		{"oversized body", http.StatusOK, bytes.Repeat([]byte{0x30}, intermediateCertMaxResponseBytes+1)},
		{"non-200 status", http.StatusNotFound, nil},
		{"invalid DER", http.StatusOK, []byte("not a certificate")},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := &http.Client{
				Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: tc.status,
						Body:       io.NopCloser(bytes.NewReader(tc.body)),
						Header:     make(http.Header),
					}, nil
				}),
			}
			if _, err := fetchCertificate(client, "https://example.test/cert.crt"); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

// TestIntermediateCertPoolWithCaches_ReusesSameSignerCacheEntry verifies
// cache reuse for repeated lookups of the same signer issuer.
func TestIntermediateCertPoolWithCaches_ReusesSameSignerCacheEntry(t *testing.T) {
	positive, negative := newTestIntermediateCertCaches(24*time.Hour, 5*time.Minute)

	signer := &x509.Certificate{
		RawIssuer:      []byte("issuer-1"),
		AuthorityKeyId: []byte("akid-1"),
	}

	expectedPool := x509.NewCertPool()
	fetchCalls := 0
	fetch := func(gotSigner *x509.Certificate) (*x509.CertPool, error) {
		fetchCalls++
		if gotSigner != signer {
			t.Fatalf("unexpected signer: got %p, want %p", gotSigner, signer)
		}
		return expectedPool, nil
	}

	first, err := intermediateCertPoolWithCaches(signer, fetch, positive, negative)
	if err != nil {
		t.Fatalf("first cache fetch: %v", err)
	}
	second, err := intermediateCertPoolWithCaches(signer, fetch, positive, negative)
	if err != nil {
		t.Fatalf("second cache fetch: %v", err)
	}

	if fetchCalls != 1 {
		t.Fatalf("fetch call count mismatch: got %d, want 1", fetchCalls)
	}
	if first != expectedPool || second != expectedPool {
		t.Fatal("expected cached pool to be reused for the same signer")
	}
}

// TestIntermediateCertPoolWithCaches_SeparatesDifferentSignerCacheEntries
// verifies that different issuer keys do not share one cache entry.
func TestIntermediateCertPoolWithCaches_SeparatesDifferentSignerCacheEntries(t *testing.T) {
	positive, negative := newTestIntermediateCertCaches(24*time.Hour, 5*time.Minute)

	signer1 := &x509.Certificate{
		RawIssuer:      []byte("issuer-1"),
		AuthorityKeyId: []byte("akid-1"),
	}
	signer2 := &x509.Certificate{
		RawIssuer:      []byte("issuer-1"),
		AuthorityKeyId: []byte("akid-2"),
	}

	pool1 := x509.NewCertPool()
	pool2 := x509.NewCertPool()
	fetchCalls := 0
	fetch := func(gotSigner *x509.Certificate) (*x509.CertPool, error) {
		fetchCalls++
		switch gotSigner {
		case signer1:
			return pool1, nil
		case signer2:
			return pool2, nil
		default:
			t.Fatalf("unexpected signer: %p", gotSigner)
			return nil, nil
		}
	}

	first, err := intermediateCertPoolWithCaches(signer1, fetch, positive, negative)
	if err != nil {
		t.Fatalf("first signer cache fetch: %v", err)
	}
	second, err := intermediateCertPoolWithCaches(signer2, fetch, positive, negative)
	if err != nil {
		t.Fatalf("second signer cache fetch: %v", err)
	}
	firstAgain, err := intermediateCertPoolWithCaches(signer1, fetch, positive, negative)
	if err != nil {
		t.Fatalf("first signer cache reuse: %v", err)
	}

	if fetchCalls != 2 {
		t.Fatalf("fetch call count mismatch: got %d, want 2", fetchCalls)
	}
	if first != pool1 || firstAgain != pool1 {
		t.Fatal("expected signer1 to reuse its own cached pool")
	}
	if second != pool2 {
		t.Fatal("expected signer2 to use a distinct cached pool")
	}
}

// TestIntermediateCertPoolWithCaches_CachesFetchFailure verifies that a
// failed fetch is remembered so repeated attempts for the same issuer cannot
// amplify into repeated network calls.
func TestIntermediateCertPoolWithCaches_CachesFetchFailure(t *testing.T) {
	positive, negative := newTestIntermediateCertCaches(24*time.Hour, 5*time.Minute)

	signer := &x509.Certificate{
		RawIssuer:      []byte("issuer-1"),
		AuthorityKeyId: []byte("akid-1"),
	}

	fetchCalls := 0
	fetch := func(*x509.Certificate) (*x509.CertPool, error) {
		fetchCalls++
		return nil, fmt.Errorf("simulated fetch failure")
	}

	if _, err := intermediateCertPoolWithCaches(signer, fetch, positive, negative); err == nil {
		t.Fatal("expected error from first fetch")
	}
	if _, err := intermediateCertPoolWithCaches(signer, fetch, positive, negative); err == nil {
		t.Fatal("expected error from cached negative entry")
	}

	if fetchCalls != 1 {
		t.Fatalf("fetch call count mismatch: got %d, want 1 (negative cache should absorb the second call)", fetchCalls)
	}
}

// TestIntermediateCertPoolWithCaches_NegativeCacheExpires verifies that
// once the negative cache TTL passes, the fetcher is invoked again.
func TestIntermediateCertPoolWithCaches_NegativeCacheExpires(t *testing.T) {
	// Short negative TTL so the test doesn't have to sleep long.
	positive, negative := newTestIntermediateCertCaches(24*time.Hour, 50*time.Millisecond)

	signer := &x509.Certificate{
		RawIssuer:      []byte("issuer-1"),
		AuthorityKeyId: []byte("akid-1"),
	}

	fetchCalls := 0
	fetch := func(*x509.Certificate) (*x509.CertPool, error) {
		fetchCalls++
		return nil, fmt.Errorf("simulated fetch failure")
	}

	if _, err := intermediateCertPoolWithCaches(signer, fetch, positive, negative); err == nil {
		t.Fatal("expected error from first fetch")
	}

	time.Sleep(100 * time.Millisecond)

	if _, err := intermediateCertPoolWithCaches(signer, fetch, positive, negative); err == nil {
		t.Fatal("expected error from re-fetch after negative TTL")
	}

	if fetchCalls != 2 {
		t.Fatalf("fetch call count mismatch: got %d, want 2 (negative cache should have expired)", fetchCalls)
	}
}

// TestIntermediateCertPoolWithCaches_PositiveOverridesStaleNegative verifies
// that a positive cache entry wins over a still-live negative entry for the
// same issuer, per the positive-first lookup order documented in the code.
func TestIntermediateCertPoolWithCaches_PositiveOverridesStaleNegative(t *testing.T) {
	positive, negative := newTestIntermediateCertCaches(24*time.Hour, 5*time.Minute)

	signer := &x509.Certificate{
		RawIssuer:      []byte("issuer-1"),
		AuthorityKeyId: []byte("akid-1"),
	}

	expectedPool := x509.NewCertPool()
	callCount := 0
	fetch := func(*x509.Certificate) (*x509.CertPool, error) {
		callCount++
		if callCount == 1 {
			return nil, fmt.Errorf("simulated fetch failure")
		}
		return expectedPool, nil
	}

	// First call fails and is cached negatively.
	if _, err := intermediateCertPoolWithCaches(signer, fetch, positive, negative); err == nil {
		t.Fatal("expected error from first fetch")
	}

	// Inject a positive entry under the same key, simulating a later success
	// that bypassed the negative-cache check (e.g., via a second TTLStore
	// instance or concurrent write).
	entry := &intermediateCertCacheEntry{
		key:  intermediateCacheKey{rawIssuer: "issuer-1", authorityKeyID: "akid-1"},
		pool: expectedPool,
	}
	if err := positive.Add(entry); err != nil {
		t.Fatalf("adding positive entry: %v", err)
	}

	// Next lookup must return the positive pool and must not call fetch.
	got, err := intermediateCertPoolWithCaches(signer, fetch, positive, negative)
	if err != nil {
		t.Fatalf("expected cache hit, got error: %v", err)
	}
	if got != expectedPool {
		t.Fatal("expected positive cached pool to be returned over the negative entry")
	}
	if callCount != 1 {
		t.Fatalf("fetch call count mismatch: got %d, want 1 (positive hit should not re-fetch)", callCount)
	}
}

// TestValidateAzureMetadataSignerSAN verifies that signer certificates are
// accepted iff their SAN identifies an Azure metadata endpoint.
func TestValidateAzureMetadataSignerSAN(t *testing.T) {
	testCases := []struct {
		name     string
		dnsNames []string
		wantErr  bool
	}{
		{"exact metadata.azure.com", []string{"metadata.azure.com"}, false},
		{"regional subdomain", []string{"northeurope.metadata.azure.com"}, false},
		{"mixed with other names", []string{"other.example.com", "metadata.azure.com"}, false},
		{"non-Azure metadata cert", []string{"metadata.example.com"}, true},
		{"suffix attack", []string{"metadata.azure.com.evil.test"}, true},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			signer := &x509.Certificate{DNSNames: tc.dnsNames}
			err := validateAzureMetadataSignerSAN(signer)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected SAN validation error")
				}
				return
			}
			if err != nil {
				t.Errorf("expected SAN validation to pass, got: %v", err)
			}
		})
	}
}

// TestVerifyAttestedDocumentWithRootAndFetcher exercises chain-trust, replay
// protection, expiration, and the happy path in one table. The test PKCS7
// already embeds the issuing root, so no intermediate fetch should be needed.
func TestVerifyAttestedDocumentWithRootAndFetcher(t *testing.T) {
	caCert, _, leafCert, leafKey := testPKI(t)
	body := []byte("test-body")
	now := time.Now().UTC()

	trustedPool := x509.NewCertPool()
	trustedPool.AddCert(caCert)

	validTimestamps := attestedTimeStamp{
		CreatedOn: now.Add(-time.Minute).Format(attestedDocumentTimeFormat),
		ExpiresOn: now.Add(time.Hour).Format(attestedDocumentTimeFormat),
	}
	ageWithinSkewTimestamps := attestedTimeStamp{
		CreatedOn: now.Add(-attestedDocumentMaxAge - time.Minute).Format(attestedDocumentTimeFormat),
		ExpiresOn: now.Add(time.Hour).Format(attestedDocumentTimeFormat),
	}
	expiredTimestamps := attestedTimeStamp{
		CreatedOn: now.Add(-(2*time.Minute + 30*time.Second)).Format(attestedDocumentTimeFormat),
		ExpiresOn: now.Add(-(attestedDocumentMaxClockSkew + 10*time.Second)).Format(attestedDocumentTimeFormat),
	}
	expiryWithinSkewTimestamps := attestedTimeStamp{
		CreatedOn: now.Add(-2 * time.Minute).Format(attestedDocumentTimeFormat),
		ExpiresOn: now.Add(-time.Minute).Format(attestedDocumentTimeFormat),
	}
	staleTimestamps := attestedTimeStamp{
		CreatedOn: now.Add(-(attestedDocumentMaxAge + attestedDocumentMaxClockSkew + time.Minute)).Format(attestedDocumentTimeFormat),
		ExpiresOn: now.Add(time.Hour).Format(attestedDocumentTimeFormat),
	}
	futureTimestamps := attestedTimeStamp{
		CreatedOn: now.Add(attestedDocumentMaxClockSkew + time.Minute).Format(attestedDocumentTimeFormat),
		ExpiresOn: now.Add(time.Hour).Format(attestedDocumentTimeFormat),
	}

	testCases := []struct {
		name          string
		data          attestedData
		trustStore    *x509.CertPool
		wantErrSubstr string
		wantVMId      string
	}{
		{
			name:          "untrusted signature",
			data:          attestedData{VMId: "test-vm-id", SubscriptionId: "test-subscription", Nonce: nonceForBody(body), TimeStamp: validTimestamps},
			trustStore:    x509.NewCertPool(), // empty — test CA is not trusted
			wantErrSubstr: "verifying PKCS7 certificate chain",
		},
		{
			name:          "missing vmId",
			data:          attestedData{SubscriptionId: "test-subscription", Nonce: nonceForBody(body), TimeStamp: validTimestamps},
			trustStore:    trustedPool,
			wantErrSubstr: "vmId is required",
		},
		{
			name:          "missing subscriptionId",
			data:          attestedData{VMId: "test-vm-id", Nonce: nonceForBody(body), TimeStamp: validTimestamps},
			trustStore:    trustedPool,
			wantErrSubstr: "subscriptionId is required",
		},
		{
			name:          "nonce mismatch",
			data:          attestedData{VMId: "test-vm-id", SubscriptionId: "test-subscription", Nonce: "wrong-nonce", TimeStamp: validTimestamps},
			trustStore:    trustedPool,
			wantErrSubstr: "nonce mismatch",
		},
		{
			name:          "expired document",
			data:          attestedData{VMId: "test-vm-id", SubscriptionId: "test-subscription", Nonce: nonceForBody(body), TimeStamp: expiredTimestamps},
			trustStore:    trustedPool,
			wantErrSubstr: "expired at",
		},
		{
			name:       "createdOn older than maxAge but within skew",
			data:       attestedData{VMId: "test-vm-id", SubscriptionId: "test-subscription", Nonce: nonceForBody(body), TimeStamp: ageWithinSkewTimestamps},
			trustStore: trustedPool,
			wantVMId:   "test-vm-id",
		},
		{
			name:       "expiresOn slightly in past but within skew",
			data:       attestedData{VMId: "test-vm-id", SubscriptionId: "test-subscription", Nonce: nonceForBody(body), TimeStamp: expiryWithinSkewTimestamps},
			trustStore: trustedPool,
			wantVMId:   "test-vm-id",
		},
		{
			name:          "stale createdOn",
			data:          attestedData{VMId: "test-vm-id", SubscriptionId: "test-subscription", Nonce: nonceForBody(body), TimeStamp: staleTimestamps},
			trustStore:    trustedPool,
			wantErrSubstr: "older than",
		},
		{
			name: "future createdOn",
			data: attestedData{
				VMId:           "test-vm-id",
				SubscriptionId: "test-subscription",
				Nonce:          nonceForBody(body),
				TimeStamp:      futureTimestamps,
			},
			trustStore:    trustedPool,
			wantErrSubstr: "too far in the future",
		},
		{
			name: "missing createdOn",
			data: attestedData{
				VMId:           "test-vm-id",
				SubscriptionId: "test-subscription",
				Nonce:          nonceForBody(body),
				TimeStamp:      attestedTimeStamp{ExpiresOn: now.Add(time.Hour).Format(attestedDocumentTimeFormat)},
			},
			trustStore:    trustedPool,
			wantErrSubstr: "createdOn is required",
		},
		{
			name: "malformed createdOn",
			data: attestedData{
				VMId:           "test-vm-id",
				SubscriptionId: "test-subscription",
				Nonce:          nonceForBody(body),
				TimeStamp:      attestedTimeStamp{CreatedOn: "not-a-time", ExpiresOn: now.Add(time.Hour).Format(attestedDocumentTimeFormat)},
			},
			trustStore:    trustedPool,
			wantErrSubstr: "parsing attested document creation",
		},
		{
			name: "success",
			data: attestedData{
				VMId:           "02aab8a4-74ef-476e-8182-f6d2ba4166a6",
				SubscriptionId: "8d10da13-8125-4ba9-a717-bf7490507b3d",
				Nonce:          nonceForBody(body),
				TimeStamp:      validTimestamps,
			},
			trustStore: trustedPool,
			wantVMId:   "02aab8a4-74ef-476e-8182-f6d2ba4166a6",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sig := testSignature(t, tc.data, leafCert, leafKey, caCert)
			fetchCalls := 0
			result, err := verifyAttestedDocumentWithRootAndFetcher(sig, body, tc.trustStore, func(*x509.Certificate) (*x509.CertPool, error) {
				fetchCalls++
				return x509.NewCertPool(), nil
			})
			if fetchCalls != 0 {
				t.Fatalf("fetch call count mismatch: got %d, want 0", fetchCalls)
			}
			if tc.wantErrSubstr != "" {
				if err == nil {
					t.Fatal("expected error")
				}
				if !strings.Contains(err.Error(), tc.wantErrSubstr) {
					t.Fatalf("error mismatch: got %q, want substring %q", err, tc.wantErrSubstr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantVMId != "" && result.VMId != tc.wantVMId {
				t.Errorf("vmId: got %q, want %q", result.VMId, tc.wantVMId)
			}
		})
	}
}

// TestVerifyAttestedDocumentWithRootAndFetcher_ChainResolution exercises
// the chain-resolution and fetch-fallback logic in a single table: whether
// the embedded PKCS7 chain suffices, whether the fetcher is consulted, and
// whether content or fetch errors short-circuit correctly.
func TestVerifyAttestedDocumentWithRootAndFetcher_ChainResolution(t *testing.T) {
	rootCert, intermediateCert, leafCert, leafKey := testPKIChain(t)
	body := []byte("test-body")

	fetchFailure := fmt.Errorf("simulated fetch failure")

	testCases := []struct {
		name              string
		nonce             string
		embedIntermediate bool
		fetchErr          error
		wantErrSubstr     string
		wantFetchCalls    int
	}{
		{
			name:              "embedded chain suffices",
			nonce:             nonceForBody(body),
			embedIntermediate: true,
			wantFetchCalls:    0,
		},
		{
			name:              "fetches missing intermediate",
			nonce:             nonceForBody(body),
			embedIntermediate: false,
			wantFetchCalls:    1,
		},
		{
			name:              "invalid content exits before fetch",
			nonce:             "wrong-nonce",
			embedIntermediate: false,
			wantErrSubstr:     "nonce mismatch",
			wantFetchCalls:    0,
		},
		{
			name:              "propagates fetch error",
			nonce:             nonceForBody(body),
			embedIntermediate: false,
			fetchErr:          fetchFailure,
			wantErrSubstr:     "fetch failure",
			wantFetchCalls:    1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			content, err := json.Marshal(attestedData{
				VMId:           "test-vm-id",
				SubscriptionId: "test-subscription",
				Nonce:          tc.nonce,
				TimeStamp: attestedTimeStamp{
					CreatedOn: time.Now().UTC().Format(attestedDocumentTimeFormat),
					ExpiresOn: time.Now().Add(time.Hour).UTC().Format(attestedDocumentTimeFormat),
				},
			})
			if err != nil {
				t.Fatalf("marshalling attested data: %v", err)
			}

			var sig string
			if tc.embedIntermediate {
				sig = base64.StdEncoding.EncodeToString(createTestPKCS7(t, content, leafCert, leafKey, intermediateCert))
			} else {
				sig = base64.StdEncoding.EncodeToString(createTestPKCS7(t, content, leafCert, leafKey))
			}

			rootPool := x509.NewCertPool()
			rootPool.AddCert(rootCert)

			fetchCalls := 0
			result, err := verifyAttestedDocumentWithRootAndFetcher(sig, body, rootPool, func(*x509.Certificate) (*x509.CertPool, error) {
				fetchCalls++
				if tc.fetchErr != nil {
					return nil, tc.fetchErr
				}
				pool := x509.NewCertPool()
				pool.AddCert(intermediateCert)
				return pool, nil
			})

			if fetchCalls != tc.wantFetchCalls {
				t.Fatalf("fetch call count: got %d, want %d", fetchCalls, tc.wantFetchCalls)
			}
			if tc.wantErrSubstr != "" {
				if err == nil {
					t.Fatal("expected error")
				}
				if !strings.Contains(err.Error(), tc.wantErrSubstr) {
					t.Fatalf("error mismatch: got %q, want substring %q", err, tc.wantErrSubstr)
				}
				if tc.fetchErr != nil && !errors.Is(err, tc.fetchErr) {
					t.Errorf("expected wrapped fetchErr in error tree, got: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.VMId != "test-vm-id" {
				t.Fatalf("vmId: got %q, want %q", result.VMId, "test-vm-id")
			}
		})
	}
}

// TestParseAndValidateAttestedDocumentContent_AcceptsMicrosoftFormat verifies
// that the verifier decodes the JSON field names and timestamp format that
// Azure IMDS actually emits — including the -0000 timezone suffix that Go's
// default time formatter does not produce. This locks in compatibility with
// real Microsoft IMDS responses without depending on a signed fixture whose
// certificates expire.
func TestParseAndValidateAttestedDocumentContent_AcceptsMicrosoftFormat(t *testing.T) {
	body := []byte("test-body")
	createdOn := formatMicrosoftAttestedTime(time.Now().Add(-time.Minute))
	expiresOn := formatMicrosoftAttestedTime(time.Now().Add(time.Hour))
	content := fmt.Sprintf(`{
		"vmId": "3ceb0a9e-ff74-4e17-924a-f2acd3b31310",
		"subscriptionId": "46678f10-4bbb-447e-98e8-d2829589f2d8",
		"nonce": %q,
		"timeStamp": {
			"createdOn": %q,
			"expiresOn": %q
		}
	}`, nonceForBody(body), createdOn, expiresOn)

	data, err := parseAndValidateAttestedDocumentContent([]byte(content), body)
	if err != nil {
		t.Fatalf("parsing Microsoft-format content: %v", err)
	}
	if data.VMId != "3ceb0a9e-ff74-4e17-924a-f2acd3b31310" {
		t.Errorf("vmId decode mismatch: got %q", data.VMId)
	}
	if data.SubscriptionId != "46678f10-4bbb-447e-98e8-d2829589f2d8" {
		t.Errorf("subscriptionId decode mismatch: got %q", data.SubscriptionId)
	}
	if data.TimeStamp.CreatedOn != createdOn {
		t.Errorf("createdOn decode mismatch: got %q", data.TimeStamp.CreatedOn)
	}
	if data.TimeStamp.ExpiresOn != expiresOn {
		t.Errorf("expiresOn decode mismatch: got %q", data.TimeStamp.ExpiresOn)
	}
}
