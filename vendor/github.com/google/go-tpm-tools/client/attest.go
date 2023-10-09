package client

import (
	"crypto/x509"
	"fmt"
	"io"
	"net/http"

	sabi "github.com/google/go-sev-guest/abi"
	sg "github.com/google/go-sev-guest/client"
	tg "github.com/google/go-tdx-guest/client"
	tabi "github.com/google/go-tdx-guest/client/linuxabi"
	pb "github.com/google/go-tpm-tools/proto/attest"
)

const (
	maxIssuingCertificateURLs = 3
	maxCertChainLength        = 4
)

// TEEDevice is an interface to add an attestation report from a TEE technology's
// attestation driver.
type TEEDevice interface {
	// AddAttestation uses the TEE device's attestation driver to collect an
	// attestation report, then adds it to the correct field of `attestation`.
	AddAttestation(attestation *pb.Attestation, options AttestOpts) error
	// Close finalizes any resources in use by the TEEDevice.
	Close() error
}

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
	// TCG Event Log to add to the attestation.
	// If not specified then it take Event Log by calling GetEventLog().
	TCGEventLog []byte
	// TCG Canonical Event Log to add to the attestation.
	// Currently, we only support PCR replay for PCRs orthogonal to those in the
	// firmware event log, where PCRs 0-9 and 14 are often measured. If the two
	// logs overlap, server-side verification using this library may fail.
	CanonicalEventLog []byte
	// If non-nil, will be used to fetch the AK certificate chain for validation.
	// Key.Attest() will construct the certificate chain by making GET requests to
	// the contents of Key.cert.IssuingCertificateURL using this client.
	CertChainFetcher *http.Client
	// TEEDevice implements the TEEDevice interface for collecting a Trusted execution
	// environment attestation. If nil, then Attest will try all known TEE devices,
	// and TEENonce must be nil. If not nil, Attest will not call Close() on the device.
	TEEDevice TEEDevice
	// TEENonce is the nonce that will be used in the TEE's attestation collection
	// mechanism. It is expected to be the size required by the technology. If nil,
	// then the nonce will be populated with Nonce, either truncated or zero-filled
	// depending on the technology's size. Leaving this nil is not recommended. If
	// nil, then TEEDevice must be nil.
	TEENonce []byte
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

// SevSnpDevice encapsulates the SEV-SNP attestation device to add its attestation report
// to a pb.Attestation.
type SevSnpDevice struct {
	Device sg.Device
}

// TdxDevice encapsulates the TDX attestation device to add its attestation quote
// to a pb.Attestation.
type TdxDevice struct {
	Device tg.Device
}

// CreateSevSnpDevice opens the SEV-SNP attestation driver and wraps it with behavior
// that allows it to add an attestation report to pb.Attestation.
func CreateSevSnpDevice() (*SevSnpDevice, error) {
	d, err := sg.OpenDevice()
	if err != nil {
		return nil, err
	}
	return &SevSnpDevice{Device: d}, nil
}

// AddAttestation will get the SEV-SNP attestation report given opts.TEENonce with
// associated certificates and add them to `attestation`. If opts.TEENonce is empty,
// then uses contents of opts.Nonce.
func (d *SevSnpDevice) AddAttestation(attestation *pb.Attestation, opts AttestOpts) error {
	var snpNonce [sabi.ReportDataSize]byte
	if len(opts.TEENonce) == 0 {
		copy(snpNonce[:], opts.Nonce)
	} else if len(opts.TEENonce) != sabi.ReportDataSize {
		return fmt.Errorf("the TEENonce size is %d. SEV-SNP device requires 64", len(opts.TEENonce))
	} else {
		copy(snpNonce[:], opts.TEENonce)
	}
	extReport, err := sg.GetExtendedReport(d.Device, snpNonce)
	if err != nil {
		return err
	}
	attestation.TeeAttestation = &pb.Attestation_SevSnpAttestation{
		SevSnpAttestation: extReport,
	}
	return nil
}

// Close will free the device handle held by the SevSnpDevice. Calling more
// than once has no effect.
func (d *SevSnpDevice) Close() error {
	if d.Device != nil {
		err := d.Device.Close()
		d.Device = nil
		return err
	}
	return nil
}

// CreateTdxDevice opens the TDX attestation driver and wraps it with behavior
// that allows it to add an attestation quote to pb.Attestation.
func CreateTdxDevice() (*TdxDevice, error) {
	d, err := tg.OpenDevice()
	if err != nil {
		return nil, err
	}
	return &TdxDevice{Device: d}, nil
}

// AddAttestation will get the TDX attestation quote given opts.TEENonce
// and add them to `attestation`. If opts.TEENonce is empty, then uses
// contents of opts.Nonce.
func (d *TdxDevice) AddAttestation(attestation *pb.Attestation, opts AttestOpts) error {
	var tdxNonce [tabi.TdReportDataSize]byte
	if len(opts.TEENonce) == 0 {
		copy(tdxNonce[:], opts.Nonce)
	} else if len(opts.TEENonce) != tabi.TdReportDataSize {
		return fmt.Errorf("the TEENonce size is %d. Intel TDX device requires %d", len(opts.TEENonce), tabi.TdReportDataSize)
	} else {
		copy(tdxNonce[:], opts.TEENonce)
	}
	quote, err := tg.GetQuote(d.Device, tdxNonce)
	if err != nil {
		return err
	}
	attestation.TeeAttestation = &pb.Attestation_TdxAttestation{
		TdxAttestation: quote,
	}
	return nil
}

// Close will free the device handle held by the TdxDevice. Calling more
// than once has no effect.
func (d *TdxDevice) Close() error {
	if d.Device != nil {
		err := d.Device.Close()
		d.Device = nil
		return err
	}
	return nil
}

// Does best effort to get a TEE hardware rooted attestation, but won't fail fatally
// unless the user provided a TEEDevice object.
func getTEEAttestationReport(attestation *pb.Attestation, opts AttestOpts) error {
	device := opts.TEEDevice
	if device != nil {
		return device.AddAttestation(attestation, opts)
	}

	// TEEDevice can't be nil while TEENonce is non-nil
	if opts.TEENonce != nil {
		return fmt.Errorf("got non-nil TEENonce when TEEDevice is nil: %v", opts.TEENonce)
	}

	// Try SEV-SNP.
	if device, err := CreateSevSnpDevice(); err == nil {
		// Don't return errors if the attestation collection fails, since
		// the user didn't specify a TEEDevice.
		device.AddAttestation(attestation, opts)
		device.Close()
		return nil
	}

	// Try TDX.
	if device, err := CreateTdxDevice(); err == nil {
		// Don't return errors if the attestation collection fails, since
		// the user didn't specify a TEEDevice.
		device.AddAttestation(attestation, opts)
		device.Close()
		return nil
	}
	// Add more devices here.
	return nil
}

// Attest generates an Attestation containing the TCG Event Log and a Quote over
// all PCR banks. The provided nonce can be used to guarantee freshness of the
// attestation. This function will return an error if the key is not a
// restricted signing key.
//
// AttestOpts is used for additional configuration of the Attestation process.
// This is primarily used to pass the attestation's nonce:
//
//	attestation, err := key.Attest(client.AttestOpts{Nonce: my_nonce})
func (k *Key) Attest(opts AttestOpts) (*pb.Attestation, error) {
	if len(opts.Nonce) == 0 {
		return nil, fmt.Errorf("provided nonce must not be empty")
	}
	sels, err := allocatedPCRs(k.rw)
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
	if opts.TCGEventLog == nil {
		if attestation.EventLog, err = GetEventLog(k.rw); err != nil {
			return nil, fmt.Errorf("failed to retrieve TCG Event Log: %w", err)
		}
	} else {
		attestation.EventLog = opts.TCGEventLog
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

	if err := getTEEAttestationReport(&attestation, opts); err != nil {
		return nil, fmt.Errorf("collecting TEE attestation report: %w", err)
	}

	return &attestation, nil
}
