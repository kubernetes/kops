package client

import (
	"fmt"
	"net/http"

	sabi "github.com/google/go-sev-guest/abi"
	sg "github.com/google/go-sev-guest/client"
	tg "github.com/google/go-tdx-guest/client"
	tabi "github.com/google/go-tdx-guest/client/linuxabi"
	tpb "github.com/google/go-tdx-guest/proto/tdx"
	"github.com/google/go-tpm-tools/internal"
	pb "github.com/google/go-tpm-tools/proto/attest"
)

// TEEDevice is an interface to add an attestation report from a TEE technology's
// attestation driver or quote provider.
type TEEDevice interface {
	// AddAttestation uses the TEE device's attestation driver or quote provider to collect an
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
	// Deprecated: Manually populate the pb.Attestation instead.
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

	// Setting this skips attaching the TEE attestation
	SkipTeeAttestation bool
}

// SevSnpQuoteProvider encapsulates the SEV-SNP attestation device to add its attestation report
// to a pb.Attestation.
type SevSnpQuoteProvider struct {
	QuoteProvider sg.QuoteProvider
}

// TdxDevice encapsulates the TDX attestation device to add its attestation quote
// to a pb.Attestation.
// Deprecated: TdxDevice is deprecated. It is recommended to use TdxQuoteProvider.
type TdxDevice struct {
	Device tg.Device
}

// TdxQuoteProvider encapsulates the TDX attestation device to add its attestation quote
// to a pb.Attestation.
type TdxQuoteProvider struct {
	QuoteProvider tg.QuoteProvider
}

// AddAttestation will get the SEV-SNP attestation report given opts.TEENonce with
// associated certificates and add them to `attestation`. If opts.TEENonce is empty,
// then uses contents of opts.Nonce.
func (d *SevSnpQuoteProvider) AddAttestation(attestation *pb.Attestation, opts AttestOpts) error {
	var snpNonce [sabi.ReportDataSize]byte
	if len(opts.TEENonce) == 0 {
		copy(snpNonce[:], opts.Nonce)
	} else if len(opts.TEENonce) != sabi.ReportDataSize {
		return fmt.Errorf("the TEENonce size is %d. SEV-SNP device requires 64", len(opts.TEENonce))
	} else {
		copy(snpNonce[:], opts.TEENonce)
	}
	raw, err := d.QuoteProvider.GetRawQuote(snpNonce)
	if err != nil {
		return err
	}
	extReport, err := sabi.ReportCertsToProto(raw)
	if err != nil {
		return err
	}
	attestation.TeeAttestation = &pb.Attestation_SevSnpAttestation{
		SevSnpAttestation: extReport,
	}
	return nil
}

// Close is a no-op.
func (d *SevSnpQuoteProvider) Close() error {
	return nil
}

// CreateSevSnpQuoteProvider creates the SEV-SNP quote provider and wraps it with behavior
// that allows it to add an attestation quote to pb.Attestation.
func CreateSevSnpQuoteProvider() (TEEDevice, error) {
	qp, err := sg.GetQuoteProvider()
	if err != nil {
		return nil, err
	}
	if !qp.IsSupported() {
		return nil, fmt.Errorf("sev-snp attestation reports not available")
	}
	return &SevSnpQuoteProvider{QuoteProvider: qp}, nil
}

// CreateTdxDevice opens the TDX attestation driver and wraps it with behavior
// that allows it to add an attestation quote to pb.Attestation.
// Deprecated: TdxDevice is deprecated, and use of CreateTdxQuoteProvider is
// recommended to create a TEEDevice.
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
	err := fillTdxNonce(opts, tdxNonce[:])
	if err != nil {
		return err
	}
	quote, err := tg.GetQuote(d.Device, tdxNonce)
	if err != nil {
		return err
	}
	return setTeeAttestationTdxQuote(quote, attestation)
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

// CreateTdxQuoteProvider creates the TDX quote provider and wraps it with behavior
// that allows it to add an attestation quote to pb.Attestation.
func CreateTdxQuoteProvider() (*TdxQuoteProvider, error) {
	qp, err := tg.GetQuoteProvider()
	if err != nil {
		return nil, err
	}
	if qp.IsSupported() != nil {
		// TDX quote provider has a fallback mechanism to fetch attestation quote
		// via device driver in case ConfigFS is not supported, so checking for TDX
		// device availability here. Once Device interface is fully removed from
		// subsequent go-tdx-guest versions, then below OpenDevice call should be
		// removed as well.
		d, err2 := tg.OpenDevice()
		if err2 != nil {
			return nil, fmt.Errorf("neither TDX device, nor quote provider is supported")
		}
		d.Close()
	}

	return &TdxQuoteProvider{QuoteProvider: qp}, nil
}

// AddAttestation will get the TDX attestation quote given opts.TEENonce
// and add them to `attestation`. If opts.TEENonce is empty, then uses
// contents of opts.Nonce.
func (qp *TdxQuoteProvider) AddAttestation(attestation *pb.Attestation, opts AttestOpts) error {
	var tdxNonce [tabi.TdReportDataSize]byte
	err := fillTdxNonce(opts, tdxNonce[:])
	if err != nil {
		return err
	}
	quote, err := tg.GetQuote(qp.QuoteProvider, tdxNonce)
	if err != nil {
		return err
	}
	return setTeeAttestationTdxQuote(quote, attestation)
}

// Close will free resources held by QuoteProvider.
func (qp *TdxQuoteProvider) Close() error {
	return nil
}

func fillTdxNonce(opts AttestOpts, tdxNonce []byte) error {
	if len(opts.TEENonce) == 0 {
		copy(tdxNonce[:], opts.Nonce)
	} else if len(opts.TEENonce) != tabi.TdReportDataSize {
		return fmt.Errorf("the TEENonce size is %d. Intel TDX device requires %d", len(opts.TEENonce), tabi.TdReportDataSize)
	} else {
		copy(tdxNonce[:], opts.TEENonce)
	}
	return nil
}

func setTeeAttestationTdxQuote(quote any, attestation *pb.Attestation) error {
	switch q := quote.(type) {
	case *tpb.QuoteV4:
		attestation.TeeAttestation = &pb.Attestation_TdxAttestation{
			TdxAttestation: q,
		}
	default:
		return fmt.Errorf("unsupported quote type: %T", quote)
	}
	return nil
}

// Does best effort to get a TEE hardware rooted attestation, but won't fail fatally
// unless the user provided a TEEDevice object.
func getTEEAttestationReport(attestation *pb.Attestation, opts AttestOpts) error {
	if opts.SkipTeeAttestation {
		return nil
	}
	device := opts.TEEDevice
	if device != nil {
		return device.AddAttestation(attestation, opts)
	}

	// TEEDevice can't be nil while TEENonce is non-nil
	if opts.TEENonce != nil {
		return fmt.Errorf("got non-nil TEENonce when TEEDevice is nil: %v", opts.TEENonce)
	}

	// Try SEV-SNP.
	if sevqp, err := CreateSevSnpQuoteProvider(); err == nil {
		// Don't return errors if the attestation collection fails, since
		// the user didn't specify a TEEDevice.
		sevqp.AddAttestation(attestation, opts)
		return nil
	}

	// Try TDX.
	if quoteProvider, err := CreateTdxQuoteProvider(); err == nil {
		// Don't return errors if the attestation collection fails, since
		// the user didn't specify a TEEDevice.
		quoteProvider.AddAttestation(attestation, opts)
		quoteProvider.Close()
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
	sels, err := AllocatedPCRs(k.rw)
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
		attestation.IntermediateCerts, err = internal.GetCertificateChain(k.cert, opts.CertChainFetcher)
		if err != nil {
			return nil, fmt.Errorf("fetching certificate chain: %w", err)
		}
	}

	// TODO: issues/504 this should be outside of this function, not related to TPM attestation
	if err := getTEEAttestationReport(&attestation, opts); err != nil {
		return nil, fmt.Errorf("collecting TEE attestation report: %w", err)
	}

	return &attestation, nil
}
