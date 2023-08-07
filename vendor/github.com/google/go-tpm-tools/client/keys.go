// Package client contains some high-level TPM 2.0 functions.
package client

import (
	"bytes"
	"crypto"
	"crypto/subtle"
	"crypto/x509"
	"errors"
	"fmt"
	"io"

	"github.com/google/go-tpm-tools/internal"
	pb "github.com/google/go-tpm-tools/proto/tpm"
	"github.com/google/go-tpm/legacy/tpm2"
	"github.com/google/go-tpm/tpmutil"
)

// Key wraps an active asymmetric TPM2 key. This can either be a signing key or
// an encryption key. Users of Key should be sure to call Close() when the Key
// is no longer needed, so that the underlying TPM handle can be freed.
// Concurrent accesses on Key are not safe, with the exception of the
// Sign method called on the crypto.Signer returned by Key.GetSigner.
type Key struct {
	rw      io.ReadWriter
	handle  tpmutil.Handle
	pubArea tpm2.Public
	pubKey  crypto.PublicKey
	name    tpm2.Name
	session Session
	cert    *x509.Certificate
}

// EndorsementKeyRSA generates and loads a key from DefaultEKTemplateRSA.
func EndorsementKeyRSA(rw io.ReadWriter) (*Key, error) {
	ekRsa, err := NewCachedKey(rw, tpm2.HandleEndorsement, DefaultEKTemplateRSA(), EKReservedHandle)
	if err != nil {
		return nil, err
	}
	if err := ekRsa.trySetCertificateFromNvram(EKCertNVIndexRSA); err != nil {
		ekRsa.Close()
		return nil, err
	}
	return ekRsa, nil
}

// EndorsementKeyECC generates and loads a key from DefaultEKTemplateECC.
func EndorsementKeyECC(rw io.ReadWriter) (*Key, error) {
	ekEcc, err := NewCachedKey(rw, tpm2.HandleEndorsement, DefaultEKTemplateECC(), EKECCReservedHandle)
	if err != nil {
		return nil, err
	}
	if err := ekEcc.trySetCertificateFromNvram(EKCertNVIndexECC); err != nil {
		ekEcc.Close()
		return nil, err
	}
	return ekEcc, nil
}

// StorageRootKeyRSA generates and loads a key from SRKTemplateRSA.
func StorageRootKeyRSA(rw io.ReadWriter) (*Key, error) {
	return NewCachedKey(rw, tpm2.HandleOwner, SRKTemplateRSA(), SRKReservedHandle)
}

// StorageRootKeyECC generates and loads a key from SRKTemplateECC.
func StorageRootKeyECC(rw io.ReadWriter) (*Key, error) {
	return NewCachedKey(rw, tpm2.HandleOwner, SRKTemplateECC(), SRKECCReservedHandle)
}

// AttestationKeyRSA generates and loads a key from AKTemplateRSA in the Owner hierarchy.
func AttestationKeyRSA(rw io.ReadWriter) (*Key, error) {
	return NewCachedKey(rw, tpm2.HandleOwner, AKTemplateRSA(), DefaultAKRSAHandle)
}

// AttestationKeyECC generates and loads a key from AKTemplateECC in the Owner hierarchy.
func AttestationKeyECC(rw io.ReadWriter) (*Key, error) {
	return NewCachedKey(rw, tpm2.HandleOwner, AKTemplateECC(), DefaultAKECCHandle)
}

// EndorsementKeyFromNvIndex generates and loads an endorsement key using the
// template stored at the provided nvdata index. This is useful for TPMs which
// have a preinstalled AK template.
func EndorsementKeyFromNvIndex(rw io.ReadWriter, idx uint32) (*Key, error) {
	return KeyFromNvIndex(rw, tpm2.HandleEndorsement, idx)
}

// GceAttestationKeyRSA generates and loads the GCE RSA AK. Note that this
// function will only work on a GCE VM. Unlike AttestationKeyRSA, this key uses
// the Endorsement Hierarchy and its template loaded from GceAKTemplateNVIndexRSA.
func GceAttestationKeyRSA(rw io.ReadWriter) (*Key, error) {
	akRsa, err := EndorsementKeyFromNvIndex(rw, GceAKTemplateNVIndexRSA)
	if err != nil {
		return nil, err
	}
	if err := akRsa.trySetCertificateFromNvram(GceAKCertNVIndexRSA); err != nil {
		akRsa.Close()
		return nil, err
	}
	return akRsa, nil
}

// GceAttestationKeyECC generates and loads the GCE ECC AK. Note that this
// function will only work on a GCE VM. Unlike AttestationKeyECC, this key uses
// the Endorsement Hierarchy and its template loaded from GceAKTemplateNVIndexECC.
func GceAttestationKeyECC(rw io.ReadWriter) (*Key, error) {
	akEcc, err := EndorsementKeyFromNvIndex(rw, GceAKTemplateNVIndexECC)
	if err != nil {
		return nil, err
	}
	if err := akEcc.trySetCertificateFromNvram(GceAKCertNVIndexECC); err != nil {
		akEcc.Close()
		return nil, err
	}
	return akEcc, nil
}

// LoadCachedKey loads a key from cachedHandle.
// If the key is not found, an error is returned.
// This function will not overwrite an existing key, unlike NewCachedKey.
func LoadCachedKey(rw io.ReadWriter, cachedHandle tpmutil.Handle, keySession Session) (k *Key, err error) {
	cachedPub, _, _, err := tpm2.ReadPublic(rw, cachedHandle)
	if err != nil {
		return nil, fmt.Errorf("failed to read public area of cached key: %w", err)
	}

	k = &Key{rw: rw, handle: cachedHandle, pubArea: cachedPub, session: keySession}
	return k, k.finish()
}

// KeyFromNvIndex generates and loads a key under the provided parent
// (possibly a hierarchy root tpm2.Handle{Owner|Endorsement|Platform|Null})
// using the template stored at the provided nvdata index.
func KeyFromNvIndex(rw io.ReadWriter, parent tpmutil.Handle, idx uint32) (*Key, error) {
	data, err := tpm2.NVReadEx(rw, tpmutil.Handle(idx), tpm2.HandleOwner, "", 0)
	if err != nil {
		return nil, fmt.Errorf("read error at index %d: %w", idx, err)
	}
	template, err := tpm2.DecodePublic(data)
	if err != nil {
		return nil, fmt.Errorf("index %d data was not a TPM key template: %w", idx, err)
	}
	return NewKey(rw, parent, template)
}

// NewCachedKey is almost identical to NewKey, except that it initially tries to
// see if the a key matching the provided template is at cachedHandle. If so,
// that key is returned. If not, the key is created as in NewKey, and that key
// is persisted to the cachedHandle, overwriting any existing key there.
func NewCachedKey(rw io.ReadWriter, parent tpmutil.Handle, template tpm2.Public, cachedHandle tpmutil.Handle) (k *Key, err error) {
	owner := tpm2.HandleOwner
	if parent == tpm2.HandlePlatform {
		owner = tpm2.HandlePlatform
	} else if parent == tpm2.HandleNull {
		return nil, fmt.Errorf("cannot cache objects in the null hierarchy")
	}

	cachedPub, _, _, err := tpm2.ReadPublic(rw, cachedHandle)
	if err == nil {
		if cachedPub.MatchesTemplate(template) {
			k = &Key{rw: rw, handle: cachedHandle, pubArea: cachedPub}
			return k, k.finish()
		}
		// Kick out old cached key if it does not match
		if err = tpm2.EvictControl(rw, "", owner, cachedHandle, cachedHandle); err != nil {
			return nil, err
		}
	}

	k, err = NewKey(rw, parent, template)
	if err != nil {
		return nil, err
	}
	defer tpm2.FlushContext(rw, k.handle)

	if err = tpm2.EvictControl(rw, "", owner, k.handle, cachedHandle); err != nil {
		return nil, err
	}
	k.handle = cachedHandle
	return k, nil
}

// NewKey generates a key from the template and loads that key into the TPM
// under the specified parent. NewKey can call many different TPM commands:
//   - If parent is tpm2.Handle{Owner|Endorsement|Platform|Null} a primary key
//     is created in the specified hierarchy (using CreatePrimary).
//   - If parent is a valid key handle, a normal key object is created under
//     that parent (using Create and Load). NOTE: Not yet supported.
//
// This function also assumes that the desired key:
//   - Does not have its usage locked to specific PCR values
//   - Usable with empty authorization sessions (i.e. doesn't need a password)
func NewKey(rw io.ReadWriter, parent tpmutil.Handle, template tpm2.Public) (k *Key, err error) {
	if !isHierarchy(parent) {
		// TODO add support for normal objects with Create() and Load()
		return nil, fmt.Errorf("unsupported parent handle: %x", parent)
	}

	handle, pubArea, _, _, _, _, err := tpm2.CreatePrimaryEx(rw, parent, tpm2.PCRSelection{}, "", "", template)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			tpm2.FlushContext(rw, handle)
		}
	}()

	k = &Key{rw: rw, handle: handle}
	if k.pubArea, err = tpm2.DecodePublic(pubArea); err != nil {
		return
	}
	return k, k.finish()
}

func (k *Key) finish() error {
	var err error
	if k.pubKey, err = k.pubArea.Key(); err != nil {
		return err
	}
	if k.name, err = k.pubArea.Name(); err != nil {
		return err
	}
	// We determine the right type of session based on the auth policy
	if k.session == nil {
		if bytes.Equal(k.pubArea.AuthPolicy, defaultEKAuthPolicy()) {
			if k.session, err = NewEKSession(k.rw); err != nil {
				return err
			}
		} else if len(k.pubArea.AuthPolicy) == 0 {
			k.session = NullSession{}
		} else {
			return fmt.Errorf("unknown auth policy when creating key")
		}
	}
	return nil
}

// Handle allows this key to be used directly with other go-tpm commands.
func (k *Key) Handle() tpmutil.Handle {
	return k.handle
}

// Name is hash of this key's public area. Only the Digest field will ever be
// populated. It is useful for various TPM commands related to authorization.
// This is equivalent to k.PublicArea.Name(), except that is cannot fail.
func (k *Key) Name() tpm2.Name {
	return k.name
}

// PublicArea exposes the key's entire public area. This is useful for
// determining additional properties of the underlying TPM key.
func (k *Key) PublicArea() tpm2.Public {
	return k.pubArea
}

// PublicKey provides a go interface to the loaded key's public area.
func (k *Key) PublicKey() crypto.PublicKey {
	return k.pubKey
}

// Close should be called when the key is no longer needed. This is important to
// do as most TPMs can only have a small number of key simultaneously loaded.
func (k *Key) Close() {
	if k.session != nil {
		k.session.Close()
	}
	tpm2.FlushContext(k.rw, k.handle)
}

// Seal seals the sensitive byte buffer to a key. This key must be an SRK (we
// currently do not support sealing to EKs). Optionally, the SealOpts struct can
// be modified to provide sealed-to PCRs. In this case, the sensitive data can
// only be unsealed if the seal-time PCRs are in the SealOpts-specified state.
// There must not be overlap in PCRs between SealOpts' Current and Target.
// During the sealing process, certification data will be created allowing
// Unseal() to validate the state of the TPM during the sealing process.
func (k *Key) Seal(sensitive []byte, opts SealOpts) (*pb.SealedBytes, error) {
	var pcrs *pb.PCRs
	var err error
	var auth []byte

	pcrs, err = mergePCRSelAndProto(k.rw, opts.Current, opts.Target)
	if err != nil {
		return nil, fmt.Errorf("invalid SealOpts: %v", err)
	}
	if len(pcrs.GetPcrs()) > 0 {
		auth = internal.PCRSessionAuth(pcrs, SessionHashAlg)
	}
	certifySel := FullPcrSel(CertifyHashAlgTpm)
	sb, err := sealHelper(k.rw, k.Handle(), auth, sensitive, certifySel)
	if err != nil {
		return nil, err
	}

	for pcrNum := range pcrs.GetPcrs() {
		sb.Pcrs = append(sb.Pcrs, pcrNum)
	}
	sb.Hash = pcrs.GetHash()
	sb.Srk = pb.ObjectType(k.pubArea.Type)
	return sb, nil
}

func sealHelper(rw io.ReadWriter, parentHandle tpmutil.Handle, auth []byte, sensitive []byte, certifyPCRsSel tpm2.PCRSelection) (*pb.SealedBytes, error) {
	inPublic := tpm2.Public{
		Type:       tpm2.AlgKeyedHash,
		NameAlg:    SessionHashAlgTpm,
		Attributes: tpm2.FlagFixedTPM | tpm2.FlagFixedParent,
		AuthPolicy: auth,
	}
	if auth == nil {
		inPublic.Attributes |= tpm2.FlagUserWithAuth
	} else {
		inPublic.Attributes |= tpm2.FlagAdminWithPolicy
	}

	priv, pub, creationData, _, ticket, err := tpm2.CreateKeyWithSensitive(rw, parentHandle, certifyPCRsSel, "", "", inPublic, sensitive)
	if err != nil {
		return nil, fmt.Errorf("failed to create key: %w", err)
	}
	certifiedPcr, err := ReadPCRs(rw, certifyPCRsSel)
	if err != nil {
		return nil, fmt.Errorf("failed to read PCRs: %w", err)
	}
	computedDigest := internal.PCRDigest(certifiedPcr, SessionHashAlg)

	decodedCreationData, err := tpm2.DecodeCreationData(creationData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode creation data: %w", err)
	}

	// make sure PCRs haven't being altered after sealing
	if subtle.ConstantTimeCompare(computedDigest, decodedCreationData.PCRDigest) == 0 {
		return nil, fmt.Errorf("PCRs have been modified after sealing")
	}

	sb := &pb.SealedBytes{}
	sb.CertifiedPcrs = certifiedPcr
	sb.Priv = priv
	sb.Pub = pub
	sb.CreationData = creationData
	if sb.Ticket, err = tpmutil.Pack(ticket); err != nil {
		return nil, err
	}
	return sb, nil
}

// Unseal attempts to reverse the process of Seal(), using the PCRs, public, and
// private data in proto.SealedBytes. Optionally, the UnsealOpts parameter can
// be used to verify the state of the TPM when the data was sealed. The
// zero-value UnsealOpts can be passed to skip certification.
func (k *Key) Unseal(in *pb.SealedBytes, opts UnsealOpts) ([]byte, error) {
	if in.Srk != pb.ObjectType(k.pubArea.Type) {
		return nil, fmt.Errorf("expected key of type %v, got %v", in.Srk, k.pubArea.Type)
	}
	sealed, _, err := tpm2.Load(
		k.rw,
		k.Handle(),
		/*parentPassword=*/ "",
		in.GetPub(),
		in.GetPriv())
	if err != nil {
		return nil, fmt.Errorf("failed to load sealed object: %w", err)
	}
	defer tpm2.FlushContext(k.rw, sealed)

	pcrs, err := mergePCRSelAndProto(k.rw, opts.CertifyCurrent, opts.CertifyExpected)
	if err != nil {
		return nil, fmt.Errorf("invalid UnsealOpts: %v", err)
	}
	if len(pcrs.GetPcrs()) > 0 {
		if err := internal.CheckSubset(pcrs, in.GetCertifiedPcrs()); err != nil {
			return nil, fmt.Errorf("failed to certify PCRs: %w", err)
		}

		var ticket tpm2.Ticket
		if _, err = tpmutil.Unpack(in.GetTicket(), &ticket); err != nil {
			return nil, fmt.Errorf("ticket unpack failed: %w", err)
		}
		creationHash := SessionHashAlg.New()
		creationHash.Write(in.GetCreationData())

		_, _, certErr := tpm2.CertifyCreation(k.rw, "", sealed, tpm2.HandleNull, nil, creationHash.Sum(nil), tpm2.SigScheme{}, ticket)
		// There is a bug in some older TPMs, where they are unable to
		// CertifyCreation when using a Null signing handle (despite this
		// being allowed by all versions of the TPM spec). To work around
		// this bug, we use a temporary signing key and ignore the signed
		// result. To reduce the cost of this workaround, we use a cached
		// ECC signing key.
		// We can detect this bug, as it triggers a RCInsufficient
		// Unmarshaling error.
		if paramErr, ok := certErr.(tpm2.ParameterError); ok && paramErr.Code == tpm2.RCInsufficient {
			signer, err := AttestationKeyECC(k.rw)
			if err != nil {
				return nil, fmt.Errorf("failed to create fallback signing key: %w", err)
			}
			defer signer.Close()
			_, _, certErr = tpm2.CertifyCreation(k.rw, "", sealed, signer.Handle(), nil, creationHash.Sum(nil), tpm2.SigScheme{}, ticket)
		}
		if certErr != nil {
			return nil, fmt.Errorf("failed to certify creation: %w", certErr)
		}

		// verify certify PCRs haven't been modified
		decodedCreationData, err := tpm2.DecodeCreationData(in.GetCreationData())
		if err != nil {
			return nil, fmt.Errorf("failed to decode creation data: %w", err)
		}
		if !internal.SamePCRSelection(in.GetCertifiedPcrs(), decodedCreationData.PCRSelection) {
			return nil, fmt.Errorf("certify PCRs does not match the PCR selection in the creation data")
		}
		expectedDigest := internal.PCRDigest(in.GetCertifiedPcrs(), SessionHashAlg)
		if subtle.ConstantTimeCompare(decodedCreationData.PCRDigest, expectedDigest) == 0 {
			return nil, fmt.Errorf("certify PCRs digest does not match the digest in the creation data")
		}
	}

	sel := tpm2.PCRSelection{Hash: tpm2.Algorithm(in.GetHash())}
	for _, pcr := range in.GetPcrs() {
		sel.PCRs = append(sel.PCRs, int(pcr))
	}

	session, err := NewPCRSession(k.rw, sel)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	auth, err := session.Auth()
	if err != nil {
		return nil, err
	}
	return tpm2.UnsealWithSession(k.rw, auth.Session, sealed, "")
}

// Quote will tell TPM to compute a hash of a set of given PCR selection, together with
// some extra data (typically a nonce), sign it with the given signing key, and return
// the signature and the attestation data. This function will return an error if
// the key is not a restricted signing key.
func (k *Key) Quote(selpcr tpm2.PCRSelection, extraData []byte) (*pb.Quote, error) {
	// Make sure that we have a valid signing key before trying quote
	var err error
	if _, err = internal.GetSigningHashAlg(k.pubArea); err != nil {
		return nil, err
	}
	if !k.hasAttribute(tpm2.FlagRestricted) {
		return nil, fmt.Errorf("unrestricted keys are insecure to use with Quote")
	}

	quote := &pb.Quote{}
	quote.Quote, quote.RawSig, err = tpm2.QuoteRaw(k.rw, k.Handle(), "", "", extraData, selpcr, tpm2.AlgNull)
	if err != nil {
		return nil, fmt.Errorf("failed to quote: %w", err)
	}
	quote.Pcrs, err = ReadPCRs(k.rw, selpcr)
	if err != nil {
		return nil, fmt.Errorf("failed to read PCRs: %w", err)
	}
	// Verify the quote client-side to make sure we didn't mess things up.
	// NOTE: the quote still must be verified server-side as well.
	if err := internal.VerifyQuote(quote, k.PublicKey(), extraData); err != nil {
		return nil, fmt.Errorf("failed to verify quote: %w", err)
	}
	return quote, nil
}

// Reseal is a shortcut to call Unseal() followed by Seal().
// CertifyOpt(nillable) will be used in Unseal(), and SealOpt(nillable)
// will be used in Seal()
func (k *Key) Reseal(in *pb.SealedBytes, uOpts UnsealOpts, sOpts SealOpts) (*pb.SealedBytes, error) {
	sensitive, err := k.Unseal(in, uOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to unseal: %w", err)
	}
	return k.Seal(sensitive, sOpts)
}

func (k *Key) hasAttribute(attr tpm2.KeyProp) bool {
	return k.pubArea.Attributes&attr != 0
}

// Cert returns the parsed certificate (or nil) for the given key.
func (k *Key) Cert() *x509.Certificate {
	return k.cert
}

// CertDERBytes provides the ASN.1 DER content of the key's certificate. If the
// key does not have a certficate, returns nil.
func (k *Key) CertDERBytes() []byte {
	if k.cert == nil {
		return nil
	}
	return k.cert.Raw
}

// SetCert assigns the provided certificate to the key after verifying it matches the key.
func (k *Key) SetCert(cert *x509.Certificate) error {
	certPubKey := cert.PublicKey.(crypto.PublicKey) // This cast cannot fail
	if !internal.PubKeysEqual(certPubKey, k.pubKey) {
		return errors.New("certificate does not match key")
	}

	k.cert = cert
	return nil
}

// Attempt to fetch a key's certificate from NVRAM. If the certificate is simply
// missing, this function succeeds (and no certificate is set). This is to allow
// for AKs and EKs that simply don't have a certificate. However, if the
// certificate read from NVRAM is either malformed or does not match the key, we
// return an error.
func (k *Key) trySetCertificateFromNvram(index uint32) error {
	certASN1, err := tpm2.NVReadEx(k.rw, tpmutil.Handle(index), tpm2.HandleOwner, "", 0)
	if err != nil {
		// Either the cert data is missing, or we are not allowed to read it
		return nil
	}
	x509Cert, err := x509.ParseCertificate(certASN1)
	if err != nil {
		return fmt.Errorf("failed to parse certificate from NV memory: %w", err)
	}
	return k.SetCert(x509Cert)
}
