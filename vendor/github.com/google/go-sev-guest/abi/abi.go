// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package abi encapsulates types and status codes from the AMD-SP (AKA PSP) device.
package abi

import (
	"crypto/ecdsa"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"

	pb "github.com/google/go-sev-guest/proto/sevsnp"
	"github.com/google/logger"
	"github.com/google/uuid"
	"golang.org/x/crypto/cryptobyte"
	"golang.org/x/crypto/cryptobyte/asn1"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	// AeadAes256Gcm is the SNP API value for the AES-256-GCM encryption algorithm.
	AeadAes256Gcm = 1

	// SignEcdsaP384Sha384 is the SNP API value for the ECC+SHA signing algorithm.
	SignEcdsaP384Sha384 = 1

	// EccP384 is the SNP API value for the P-384 ECC curve identifier.
	EccP384 = 2

	// ReportSize is the ABI-specified byte size of an SEV-SNP attestation report.
	ReportSize = 0x4A0

	// FamilyIDSize is the field size of FAMILY_ID in an SEV-SNP attestation report.
	FamilyIDSize = 16
	// ImageIDSize is the field size of IMAGE_ID in an SEV-SNP attestation report.
	ImageIDSize = 16
	// ReportDataSize is the field size of REPORT_DATA in an SEV-SNP attestation report.
	ReportDataSize = 64
	// MeasurementSize is the field size of MEASUREMENT in an SEV-SNP attestation report.
	MeasurementSize = 48
	// HostDataSize is the field size of HOST_DATA in an SEV-SNP attestation report.
	HostDataSize = 32
	// IDKeyDigestSize is the field size of ID_KEY_DIGEST in an SEV-SNP attestation report.
	IDKeyDigestSize = 48
	// AuthorKeyDigestSize is the field size of AUTHOR_KEY_DIGEST in an SEV-SNP attestation report.
	AuthorKeyDigestSize = 48
	// ReportIDSize is the field size of REPORT_ID in an SEV-SNP attestation report.
	ReportIDSize = 32
	// ReportIDMASize is the field size of REPORT_ID_MA in an SEV-SNP attestation report.
	ReportIDMASize = 32
	// ChipIDSize is the field size of CHIP_ID in an SEV-SNP attestation report.
	ChipIDSize = 64
	// SignatureSize is the field size of SIGNATURE in an SEV-SNP attestation report.
	SignatureSize = 512

	policyOffset          = 0x08
	policySMTBit          = 16
	policyReserved1bit    = 17
	policyMigrateMABit    = 18
	policyDebugBit        = 19
	policySingleSocketBit = 20

	maxPlatformInfoBit = 1

	signatureOffset = 0x2A0
	ecdsaRSsize     = 72 // From the ECDSA-P384-SHA384 format in SEV SNP API specification.

	// From the ECDSA public key format in SEV SNP API specification.
	ecdsaQXoffset = 0x04
	ecdsaQYoffset = 0x4c
	ecdsaQYend    = 0x94
	// EcdsaP384Sha384SignatureSize is the length in bytes of the ECDSA-P384-SHA384 signature format.
	EcdsaP384Sha384SignatureSize = ecdsaRSsize + ecdsaRSsize
	// EcsdaPublicKeySize is the length in bytes of the Curve, QX, QY elliptic curve public key
	// representation in the AMD SEV ABI.
	EcsdaPublicKeySize = 0x404

	// CertTableEntrySize is the ABI size of the certificate table entry struct.
	CertTableEntrySize = 24

	// GUIDSize is the byte length of a GUID's binary representation.
	GUIDSize = 16

	// The following GUIDs are defined by the AMD Guest-host communication block specification
	// for MSG_REPORT_REQ:
	// https://www.amd.com/system/files/TechDocs/56421-guest-hypervisor-communication-block-standardization.pdf

	// VcekGUID is the Versioned Chip Endorsement Key GUID
	VcekGUID = "63da758d-e664-4564-adc5-f4b93be8accd"
	// VlekGUID is the Versioned Loaded Endorsement Key GUID
	VlekGUID = "a8074bc2-a25a-483e-aae6-39c045a0b8a1"
	// AskGUID is the AMD signing Key GUID. Used for the ASVK as well.
	AskGUID = "4ab7b379-bbac-4fe4-a02f-05aef327c782"
	// ArkGUID is the AMD Root Key GUID
	ArkGUID = "c0b406a4-a803-4952-9743-3fb6014cd0ae"
	// AsvkGUID may not be defined, but we'd like it to be, so that
	// a single machine can use both VCEK and VLEK report signing.
	AsvkGUID = "00000000-0000-0000-0000-000000000000"

	// ExtraPlatformInfoGUID represents more information about the machine collecting an attestation
	// report than just the report to help interpret the attestation report.
	ExtraPlatformInfoGUID = "ecae0c0f-9502-43b1-afa2-0ae2e0d565b6"
	// ExtraPlatformInfoV0Size is the minimum size for an ExtraPlatformInfo blob.
	ExtraPlatformInfoV0Size = 8

	// CpuidProductMask keeps only the SevProduct-relevant bits from the CPUID(1).EAX result.
	CpuidProductMask    = 0x0fff0fff
	extendedFamilyShift = 20
	extendedModelShift  = 16
	familyShift         = 8
	modelShift          = 4
	// Combined extended values
	zen3zen4Family = 0x19
	zen5Family     = 0x1A
	milanModel     = 0 | 1
	genoaModel     = (1 << 4) | 1
	turinModel     = 2

	// ReportVersion2 is set by the SNP API specification
	// https://web.archive.org/web/20231222054111if_/http://www.amd.com/content/dam/amd/en/documents/epyc-technical-docs/specifications/56860.pdf
	ReportVersion2 = 2

	// ReportVersion3 is set by the SNP API specification
	// https://www.amd.com/system/files/TechDocs/56860.pdf
	ReportVersion3 = 3
)

// CertTableHeaderEntry defines an entry of the beginning of an extended attestation report which
// points to a specific key's certificate.
type CertTableHeaderEntry struct {
	// GUID is one of VcekGUID, AskGUID, or ArkGUID to identify which key an offset/length corresponds
	// to.
	GUID uuid.UUID
	// Offset is the offset into the data pages passed to the extended get_report where the specified
	// key's certificate resides.
	Offset uint32
	// Length is the length of the certificate within the data pages.
	Length uint32
}

// CertTableEntry represents both the GUID and whole Certificate contents denoted by the
// CertTableHeaderEntry ABI struct.
type CertTableEntry struct {
	GUID    uuid.UUID
	RawCert []byte
}

// CertTable represents each (GUID, Blob) pair of certificates returned by an extended guest
// request.
type CertTable struct {
	Entries []CertTableEntry
}

// Appendix B.1 of the SEV API specification

// AskCert is the SEV format for AMD signing key certificates.
type AskCert struct {
	Version      uint32
	KeyID        uuid.UUID
	CertifyingID uuid.UUID // Equals KeyID if self-signed.
	KeyUsage     uint32    // Table 111: 00 == Root signing key, 0x13 == SEV signing key.
	PubExpSize   uint32    // Must be 2048 or 4096
	ModulusSize  uint32    // Must be 2048 or 4096
	PubExp       []byte
	Modulus      []byte
	Signature    []byte
}

// SnpPlatformInfo represents an interpretation of the PLATFORM_INFO field of an attestation report.
type SnpPlatformInfo struct {
	// SMTEnabled represents if the platform that produced the attestation report has SMT enabled.
	SMTEnabled bool
	// TSMEEnabled represents if the platform that produced the attestation report has transparent
	// secure memory encryption (TSME) enabled.
	TSMEEnabled bool
}

// SnpPolicy represents the bitmask guest policy that governs the VM's behavior from launch.
type SnpPolicy struct {
	// ABIMajor is the minimum SEV SNP ABI version needed to run the guest's minor version number.
	ABIMinor uint8
	// ABIMajor is the minimum SEV SNP ABI version needed to run the guest's major version number.
	ABIMajor uint8
	// SMT is true if symmetric multithreading is allowed.
	SMT bool
	// MigrateMA is true if the guest is allowed to have a migration agent.
	MigrateMA bool
	// Debug is true if the VM can be decrypted by the host for debugging purposes.
	Debug bool
	// SingleSocket is true if the guest may only be active on a single socket.
	SingleSocket bool
}

// ParseSnpPolicy interprets the SEV SNP API's guest policy bitmask into an SnpPolicy struct type.
func ParseSnpPolicy(guestPolicy uint64) (SnpPolicy, error) {
	result := SnpPolicy{}
	if guestPolicy&uint64(1<<policyReserved1bit) == 0 {
		return result, fmt.Errorf("policy[%d] is reserved, must be 1, got 0", policyReserved1bit)
	}
	if err := mbz64(guestPolicy, "policy", 63, 21); err != nil {
		return result, err
	}
	result.ABIMinor = uint8(guestPolicy & 0xff)
	result.ABIMajor = uint8((guestPolicy >> 8) & 0xff)
	result.SMT = (guestPolicy & (1 << policySMTBit)) != 0
	result.MigrateMA = (guestPolicy & (1 << policyMigrateMABit)) != 0
	result.Debug = (guestPolicy & (1 << policyDebugBit)) != 0
	result.SingleSocket = (guestPolicy & (1 << policySingleSocketBit)) != 0
	return result, nil
}

// SnpPolicyToBytes translates a structural representation of a valid SNP policy to its ABI format.
func SnpPolicyToBytes(policy SnpPolicy) uint64 {
	result := uint64(policy.ABIMinor) | uint64(policy.ABIMajor)<<8 | uint64(1<<policyReserved1bit)
	if policy.SMT {
		result |= uint64(1 << policySMTBit)
	}
	if policy.MigrateMA {
		result |= uint64(1 << policyMigrateMABit)
	}
	if policy.Debug {
		result |= uint64(1 << policyDebugBit)
	}
	if policy.SingleSocket {
		result |= uint64(1 << policySingleSocketBit)
	}
	return result
}

// ParseSnpPlatformInfo returns an interpretation of the given platform info, or an error for
// unrecognized bits.
func ParseSnpPlatformInfo(platformInfo uint64) (SnpPlatformInfo, error) {
	result := SnpPlatformInfo{
		SMTEnabled:  (platformInfo & (1 << 0)) != 0,
		TSMEEnabled: (platformInfo & (1 << 1)) != 0,
	}
	reserved := platformInfo & ^uint64((1<<(maxPlatformInfoBit+1))-1)
	if reserved != 0 {
		return result, fmt.Errorf("unrecognized platform info bit(s): 0x%x", platformInfo)
	}
	return result, nil
}

// ParseAskCert returns a struct representation of the AMD certificate format from a byte array.
func ParseAskCert(data []byte) (*AskCert, int, error) {
	var cert AskCert
	minimumSize := 0x40

	if len(data) < minimumSize {
		return nil, 0,
			fmt.Errorf("AMD signing key too small, %dB, need at least %dB for header",
				len(data), minimumSize)
	}
	cert.Version = binary.LittleEndian.Uint32(data[0:0x04])
	copy(cert.KeyID[:], data[0x04:0x14])
	copy(cert.CertifyingID[:], data[0x14:0x24])
	cert.KeyUsage = binary.LittleEndian.Uint32(data[0x24:0x28])
	// Check that the reserved region is zero.
	if err := mbz(data, 0x28, 0x38); err != nil {
		return nil, 0, err
	}
	cert.PubExpSize = binary.LittleEndian.Uint32(data[0x38:0x3C])
	if cert.PubExpSize != 2048 && cert.PubExpSize != 4096 {
		return nil, 0, fmt.Errorf("public exponent size %d is not 2048 or 4096", cert.PubExpSize)
	}
	cert.ModulusSize = binary.LittleEndian.Uint32(data[0x3C:0x40])
	if cert.ModulusSize != 2048 && cert.ModulusSize != 4096 {
		return nil, 0, fmt.Errorf("modulus size %d is not 2048 or 4096", cert.ModulusSize)
	}
	// Add byte size of the public exponent bit size and the byte size of the modulus size doubled to
	// include the signature size.
	minimumSize += int(cert.PubExpSize/8) + int(cert.ModulusSize/4)
	if len(data) < minimumSize {
		return nil, 0, fmt.Errorf("AMD signing key too small, %dB, need at least %dB for public exponent %d and modulus %d",
			len(data), minimumSize, cert.PubExpSize, cert.ModulusSize)
	}
	cert.PubExp = make([]byte, cert.PubExpSize/8)
	cert.Modulus = make([]byte, cert.ModulusSize/8)
	cert.Signature = make([]byte, cert.ModulusSize/8)
	pubExpEnd := (0x40 + cert.PubExpSize/8)
	copy(cert.PubExp[:], data[0x40:pubExpEnd])
	modulusEnd := pubExpEnd + (cert.ModulusSize / 8)
	copy(cert.Modulus[:], data[pubExpEnd:modulusEnd])
	signatureEnd := modulusEnd + (cert.ModulusSize / 8)
	copy(cert.Signature[:], data[modulusEnd:signatureEnd])

	// Return the offset of the next byte after the certificate as well as the certificate.
	return &cert, int(signatureEnd), nil
}

// findNonZero returns the first index which is not zero, otherwise the length of the slice.
func findNonZero(data []uint8, lo, hi int) int {
	for i := lo; i < hi; i++ {
		if data[i] != 0 {
			return i
		}
	}
	return hi
}

func mbz(data []uint8, lo, hi int) error {
	if findNonZero(data, lo, hi) != hi {
		return fmt.Errorf("mbz range [0x%x:0x%x] not all zero: %s", lo, hi, hex.EncodeToString(data[lo:hi]))
	}
	return nil
}

// Checks a must-be-zero range of a uint64 between bits hi down to lo inclusive.
func mbz64(data uint64, base string, hi, lo int) error {
	if (data>>lo)&((1<<(hi-lo+1))-1) != 0 {
		return fmt.Errorf("mbz range %s[0x%x:0x%x] not all zero: %x", base, lo, hi, data)
	}
	return nil
}

// ReportToSignatureDER returns the signature component of an attestation report in DER format for
// use in x509 verification.
func ReportToSignatureDER(report []byte) ([]byte, error) {
	if len(report) != ReportSize {
		return nil, fmt.Errorf("incorrect report size: %x, want %x", len(report), ReportSize)
	}
	algo := SignatureAlgo(report)
	if algo != SignEcdsaP384Sha384 {
		return nil, fmt.Errorf("unknown signature algorithm: %d", algo)
	}
	signature := report[signatureOffset:ReportSize]
	var b cryptobyte.Builder
	b.AddASN1(asn1.SEQUENCE, func(b *cryptobyte.Builder) {
		b.AddASN1BigInt(AmdBigInt(ecdsaGetR(signature)))
		b.AddASN1BigInt(AmdBigInt(ecdsaGetS(signature)))
	})
	return b.Bytes()
}

func ecdsaGetR(signature []byte) []byte {
	return signature[0x0:0x48]
}

func ecdsaGetS(signature []byte) []byte {
	return signature[0x48:0x90]
}

func clone(b []byte) []byte {
	result := make([]byte, len(b))
	copy(result, b)
	return result
}

func signatureAlgoSlice(report []byte) []byte {
	return report[0x34:0x38]
}

// SignatureAlgo returns the SignatureAlgo field of a raw SEV-SNP attestation report.
func SignatureAlgo(report []byte) uint32 {
	return binary.LittleEndian.Uint32(signatureAlgoSlice(report))
}

// ReportSigner represents which kind of key is expected to have signed the attestation report
type ReportSigner uint8

const (
	// VcekReportSigner is the SIGNING_KEY value for if the VCEK signed the attestation report.
	VcekReportSigner ReportSigner = iota
	// VlekReportSigner is the SIGNING_KEY value for if the VLEK signed the attestation report.
	VlekReportSigner
	endorseReserved2
	endorseReserved3
	endorseReserved4
	endorseReserved5
	endorseReserved6
	// NoneReportSigner is the SIGNING_KEY value for if the attestation report is not signed.
	NoneReportSigner
)

// SignerInfo represents information about the signing circumstances for the attestation report.
type SignerInfo struct {
	// SigningKey represents kind of key by which a report was signed.
	SigningKey ReportSigner
	// MaskChipKey is true if the host chose to enable CHIP_ID masking, to cause the report's CHIP_ID
	// to be all zeros.
	MaskChipKey bool
	// AuthorKeyEn is true if the VM is launched with an IDBLOCK that includes an author key.
	AuthorKeyEn bool
}

// String returns a ReportSigner string rendering.
func (k ReportSigner) String() string {
	switch k {
	case VcekReportSigner:
		return "VCEK"
	case VlekReportSigner:
		return "VLEK"
	case NoneReportSigner:
		return "None"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", byte(k))
	}
}

// ParseSignerInfo interprets report[0x48:0x4c] into its component pieces and errors
// on non-zero mbz fields.
func ParseSignerInfo(signerInfo uint32) (result SignerInfo, err error) {
	info64 := uint64(signerInfo)
	if err = mbz64(info64, "data[0x48:0x4C]", 31, 5); err != nil {
		return result, err
	}
	result.SigningKey = ReportSigner((signerInfo >> 2) & 7)
	if result.SigningKey > VlekReportSigner && result.SigningKey < NoneReportSigner {
		return result, fmt.Errorf("signing_key values 2-6 are reserved. Got %v", result.SigningKey)
	}
	result.MaskChipKey = (signerInfo & 2) != 0
	result.AuthorKeyEn = (signerInfo & 1) != 0
	return result, nil
}

// ComposeSignerInfo returns the uint32 value expected to populate the attestation report byte range
// 0x48:0x4C.
func ComposeSignerInfo(signerInfo SignerInfo) uint32 {
	var result uint32
	if signerInfo.AuthorKeyEn {
		result |= 1
	}
	if signerInfo.MaskChipKey {
		result |= 2
	}
	result |= uint32(signerInfo.SigningKey) << 2
	return result
}

// ReportSignerInfo returns the signer info component of a SEV-SNP raw report.
func ReportSignerInfo(data []byte) (uint32, error) {
	if len(data) < 0x4C {
		return 0, fmt.Errorf("report too small: %d", len(data))
	}
	return binary.LittleEndian.Uint32(data[0x48:0x4C]), nil
}

// ReportToProto creates a pb.Report from the little-endian AMD SEV-SNP attestation report byte
// array in SEV SNP ABI format for ATTESTATION_REPORT.
func ReportToProto(data []uint8) (*pb.Report, error) {
	if len(data) < ReportSize {
		return nil, fmt.Errorf("array size is 0x%x, an SEV-SNP attestation report size is 0x%x", len(data), ReportSize)
	}

	r := &pb.Report{}
	// r.Version should be 2, but that's left to validation step.
	r.Version = binary.LittleEndian.Uint32(data[0x00:0x04])
	r.GuestSvn = binary.LittleEndian.Uint32(data[0x04:0x08])
	r.Policy = binary.LittleEndian.Uint64(data[0x08:0x10])
	if _, err := ParseSnpPolicy(r.Policy); err != nil {
		return nil, fmt.Errorf("malformed guest policy: %v", err)
	}
	r.FamilyId = clone(data[0x10:0x20])
	r.ImageId = clone(data[0x20:0x30])
	r.Vmpl = binary.LittleEndian.Uint32(data[0x30:0x34])
	r.SignatureAlgo = SignatureAlgo(data)
	r.CurrentTcb = binary.LittleEndian.Uint64(data[0x38:0x40])
	r.PlatformInfo = binary.LittleEndian.Uint64(data[0x40:0x48])

	signerInfo, err := ParseSignerInfo(binary.LittleEndian.Uint32(data[0x48:0x4C]))
	if err != nil {
		return nil, err
	}
	r.SignerInfo = ComposeSignerInfo(signerInfo)
	if err := mbz(data, 0x4C, 0x50); err != nil {
		return nil, err
	}
	r.ReportData = clone(data[0x50:0x90])
	r.Measurement = clone(data[0x90:0xC0])
	r.HostData = clone(data[0xC0:0xE0])
	r.IdKeyDigest = clone(data[0xE0:0x110])
	r.AuthorKeyDigest = clone(data[0x110:0x140])
	r.ReportId = clone(data[0x140:0x160])
	r.ReportIdMa = clone(data[0x160:0x180])
	r.ReportedTcb = binary.LittleEndian.Uint64(data[0x180:0x188])

	mbzLo := 0x188
	if r.Version == ReportVersion3 {
		mbzLo = 0x18B
		r.Cpuid1EaxFms = FmsToCpuid1Eax(data[0x188], data[0x189], data[0x18A])
	}

	if err := mbz(data, mbzLo, 0x1A0); err != nil {
		return nil, err
	}
	r.ChipId = clone(data[0x1A0:0x1E0])
	r.CommittedTcb = binary.LittleEndian.Uint64(data[0x1E0:0x1E8])
	r.CurrentBuild = uint32(data[0x1E8])
	r.CurrentMinor = uint32(data[0x1E9])
	r.CurrentMajor = uint32(data[0x1EA])
	if err := mbz(data, 0x1EB, 0x1EC); err != nil {
		return nil, err
	}
	r.CommittedBuild = uint32(data[0x1EC])
	r.CommittedMinor = uint32(data[0x1ED])
	r.CommittedMajor = uint32(data[0x1EE])
	if err := mbz(data, 0x1EF, 0x1F0); err != nil {
		return nil, err
	}
	r.LaunchTcb = binary.LittleEndian.Uint64(data[0x1F0:0x1F8])
	if err := mbz(data, 0x1F8, signatureOffset); err != nil {
		return nil, err
	}
	if r.SignatureAlgo == SignEcdsaP384Sha384 {
		if err := mbz(data, signatureOffset+EcdsaP384Sha384SignatureSize, ReportSize); err != nil {
			return nil, err
		}
	}
	r.Signature = clone(data[signatureOffset:ReportSize])
	return r, nil
}

// ReportCertsToProto creates a pb.Attestation from the report and certificate table represented in
// data. The report is expected to take exactly abi.ReportSize bytes, followed by the certificate
// table.
func ReportCertsToProto(data []uint8) (*pb.Attestation, error) {
	var certs []uint8
	report := data
	if len(data) >= ReportSize {
		report = data[:ReportSize]
		certs = data[ReportSize:]
	}
	mreport, err := ReportToProto(report)
	if err != nil {
		return nil, err
	}
	table := new(CertTable)
	if err := table.Unmarshal(certs); err != nil {
		return nil, err
	}
	return &pb.Attestation{Report: mreport, CertificateChain: table.Proto()}, nil
}

func checkReportSizes(r *pb.Report) error {
	if len(r.FamilyId) != FamilyIDSize {
		return fmt.Errorf("report family_id length is %d, expect %d", len(r.FamilyId), FamilyIDSize)
	}
	if len(r.ImageId) != ImageIDSize {
		return fmt.Errorf("report image_id length is %d, expect %d", len(r.ImageId), ImageIDSize)
	}
	if len(r.ReportData) != ReportDataSize {
		return fmt.Errorf("report_data length is %d, expect %d", len(r.ReportData), ReportDataSize)
	}
	if len(r.Measurement) != MeasurementSize {
		return fmt.Errorf("measurement length is %d, expect %d", len(r.Measurement), MeasurementSize)
	}
	if len(r.HostData) != HostDataSize {
		return fmt.Errorf("host_data length is %d, expect %d", len(r.HostData), HostDataSize)
	}
	if len(r.IdKeyDigest) != IDKeyDigestSize {
		return fmt.Errorf("id_key_digest length is %d, expect %d", len(r.IdKeyDigest), IDKeyDigestSize)
	}
	if len(r.AuthorKeyDigest) != AuthorKeyDigestSize {
		return fmt.Errorf("author_key_digest length is %d, expect %d", len(r.AuthorKeyDigest), AuthorKeyDigestSize)
	}
	if len(r.ReportId) != ReportIDSize {
		return fmt.Errorf("report_id length is %d, expect %d", len(r.ReportId), ReportIDSize)
	}
	if len(r.ReportIdMa) != ReportIDMASize {
		return fmt.Errorf("report_id_ma length is %d, expect %d", len(r.ReportIdMa), ReportIDMASize)
	}
	if len(r.ChipId) != ChipIDSize {
		return fmt.Errorf("chip_id length is %d, expect %d", len(r.ChipId), ChipIDSize)
	}
	if len(r.Signature) != SignatureSize {
		return fmt.Errorf("signature length is %d, expect %d", len(r.Signature), SignatureSize)
	}
	return nil
}

// ValidateReportFormat returns an error if the provided buffer violates structural expectations of
// attestation report data.
func ValidateReportFormat(r []byte) error {
	if len(r) < ReportSize {
		return fmt.Errorf("report size is %d bytes. Expected %d bytes", len(r), ReportSize)
	}

	version := binary.LittleEndian.Uint32(r[0x00:0x04])
	if version != ReportVersion2 && version != ReportVersion3 {
		return fmt.Errorf("report version is: %d. Expected %d or %d", version, ReportVersion2, ReportVersion3)
	}

	policy := binary.LittleEndian.Uint64(r[0x08:0x10])
	if _, err := ParseSnpPolicy(policy); err != nil {
		return fmt.Errorf("malformed guest policy: %v", err)
	}
	return nil
}

// ReportToAbiBytes translates the report back into its little-endian ABI format.
func ReportToAbiBytes(r *pb.Report) ([]byte, error) {
	if r == nil {
		return nil, fmt.Errorf("report is nil")
	}
	if err := checkReportSizes(r); err != nil {
		return nil, err
	}
	// Zero-initialized array fills all the reserved fields with the required zeros.
	data := make([]byte, ReportSize)

	binary.LittleEndian.PutUint32(data[0x00:0x04], r.Version)
	binary.LittleEndian.PutUint32(data[0x04:0x08], r.GuestSvn)
	binary.LittleEndian.PutUint64(data[0x08:0x10], r.Policy)
	copy(data[0x10:0x20], r.FamilyId[:])
	copy(data[0x20:0x30], r.ImageId[:])
	binary.LittleEndian.PutUint32(data[0x30:0x34], r.Vmpl)
	binary.LittleEndian.PutUint32(signatureAlgoSlice(data), r.SignatureAlgo)
	binary.LittleEndian.PutUint64(data[0x38:0x40], r.CurrentTcb)
	binary.LittleEndian.PutUint64(data[0x40:0x48], r.PlatformInfo)

	if _, err := ParseSignerInfo(r.SignerInfo); err != nil {
		return nil, err
	}
	binary.LittleEndian.PutUint32(data[0x48:0x4C], r.SignerInfo)
	copy(data[0x50:0x90], r.ReportData[:])
	copy(data[0x90:0xC0], r.Measurement[:])
	copy(data[0xC0:0xE0], r.HostData[:])
	copy(data[0xE0:0x110], r.IdKeyDigest[:])
	copy(data[0x110:0x140], r.AuthorKeyDigest[:])
	copy(data[0x140:0x160], r.ReportId[:])
	copy(data[0x160:0x180], r.ReportIdMa[:])
	binary.LittleEndian.PutUint64(data[0x180:0x188], r.ReportedTcb)

	// Add CPUID information if this is a version 3 report.
	if r.Version == ReportVersion3 {
		family, model, stepping := FmsFromCpuid1Eax(r.Cpuid1EaxFms)
		data[0x188] = family
		data[0x189] = model
		data[0x18A] = stepping
	}

	copy(data[0x1A0:0x1E0], r.ChipId[:])
	binary.LittleEndian.PutUint64(data[0x1E0:0x1E8], r.CommittedTcb)
	if r.CurrentBuild >= (1 << 8) {
		return nil, fmt.Errorf("current_build field must fit in a byte, got %d", r.CurrentBuild)
	}
	if r.CurrentMinor >= (1 << 8) {
		return nil, fmt.Errorf("current_minor field must fit in a byte, got %d", r.CurrentMinor)
	}
	if r.CurrentMajor >= (1 << 8) {
		return nil, fmt.Errorf("current_major field must fit in a byte, got %d", r.CurrentMajor)
	}
	data[0x1E8] = byte(r.CurrentBuild)
	data[0x1E9] = byte(r.CurrentMinor)
	data[0x1EA] = byte(r.CurrentMajor)
	if r.CommittedBuild >= (1 << 8) {
		return nil, fmt.Errorf("committed_build field must fit in a byte, got %d", r.CommittedBuild)
	}
	if r.CommittedMinor >= (1 << 8) {
		return nil, fmt.Errorf("committed_minor field must fit in a byte, got %d", r.CommittedMinor)
	}
	if r.CommittedMajor >= (1 << 8) {
		return nil, fmt.Errorf("committed_major field must fit in a byte, got %d", r.CommittedMajor)
	}
	data[0x1EC] = byte(r.CommittedBuild)
	data[0x1ED] = byte(r.CommittedMinor)
	data[0x1EE] = byte(r.CommittedMajor)
	binary.LittleEndian.PutUint64(data[0x1F0:0x1F8], r.LaunchTcb)

	copy(data[signatureOffset:ReportSize], r.Signature[:])
	return data, nil
}

// SignedComponent returns the bytes of the SnpAttestationReport that are signed by the AMD-SP.
func SignedComponent(report []byte) []byte {
	// Table 21 of https://www.amd.com/system/files/TechDocs/56860.pdf shows the signature is over
	// all bytes prior to the signature in the report.
	return report[0:signatureOffset]
}

func reverse(d []byte) []byte {
	for i := 0; i < len(d)/2; i++ {
		swapIndex := len(d) - i - 1
		tmp := d[i]
		d[i] = d[swapIndex]
		d[swapIndex] = tmp
	}
	return d
}

func bigIntToAMDRS(b *big.Int) []byte {
	var result [ecdsaRSsize]byte
	b.FillBytes(result[:])
	return reverse(result[:])
}

// EcdsaPublicKeyToBytes returns the AMD SEV ABI format of the ECDSA P-384 curve public key.
func EcdsaPublicKeyToBytes(key *ecdsa.PublicKey) ([]byte, error) {
	result := make([]byte, EcsdaPublicKeySize)
	switch key.Curve.Params().Name {
	case "P-384":
		binary.LittleEndian.PutUint32(result[0:4], EccP384)
	default:
		return nil, fmt.Errorf("ecdsa public key is not on curve P-384")
	}
	copy(result[ecdsaQXoffset:ecdsaQYoffset], bigIntToAMDRS(key.X))
	copy(result[ecdsaQYoffset:ecdsaQYend], bigIntToAMDRS(key.Y))
	return result, nil
}

// AmdBigInt returns a given AMD format little endian big integer as a big.Int.
func AmdBigInt(b []byte) *big.Int {
	return new(big.Int).SetBytes(reverse(clone(b)))
}

// SetSignature sets the signature component the SnpAttestationReport with the specified
// representation of the R, S components of an ECDSA signature. Useful for testing.
func SetSignature(r, s *big.Int, report []byte) error {
	if len(report) != ReportSize {
		return fmt.Errorf("unexpected report size: %x, want %x", len(report), ReportSize)
	}
	signature := report[signatureOffset:ReportSize]
	copy(ecdsaGetR(signature), bigIntToAMDRS(r))
	copy(ecdsaGetS(signature), bigIntToAMDRS(s))
	return nil
}

// Unmarshal populates a CertTableHeaderEntry from its ABI representation.
func (h *CertTableHeaderEntry) Unmarshal(data []byte) error {
	if len(data) < CertTableEntrySize {
		return fmt.Errorf("data too small: %v, want %v", len(data), CertTableEntrySize)
	}
	copy(h.GUID[:], data[0:GUIDSize])
	uint32Size := 4
	h.Offset = binary.LittleEndian.Uint32(data[GUIDSize : GUIDSize+uint32Size])
	h.Length = binary.LittleEndian.Uint32(data[GUIDSize+uint32Size : CertTableEntrySize])
	return nil
}

// Write writes a CertTableHeaderEntry in its ABI representation to data.
func (h *CertTableHeaderEntry) Write(data []byte) error {
	if len(data) < CertTableEntrySize {
		return fmt.Errorf("data too small: %v, want %v", len(data), CertTableEntrySize)
	}
	copy(data[0:GUIDSize], h.GUID[:])
	uint32Size := 4
	binary.LittleEndian.PutUint32(data[GUIDSize:GUIDSize+uint32Size], h.Offset)
	binary.LittleEndian.PutUint32(data[GUIDSize+uint32Size:CertTableEntrySize], h.Length)
	return nil
}

// ParseSnpCertTableHeader interprets the data pages from an extended guest request for certificate
// information.
func ParseSnpCertTableHeader(certs []byte) ([]CertTableHeaderEntry, error) {
	var entries []CertTableHeaderEntry
	var index int
	slice := certs[:]
	// Allow an empty table without the zero terminator.
	if len(slice) == 0 {
		return nil, nil
	}
	for {
		var next CertTableHeaderEntry
		if err := next.Unmarshal(slice); err != nil {
			return nil, fmt.Errorf("cert table index %d entry unmarshalling error: %v", index, err)
		}

		slice = slice[CertTableEntrySize:]
		index += CertTableEntrySize

		// A whole zero entry found. We're done.
		if next.Offset == 0 && next.Length == 0 && findNonZero(next.GUID[:], 0, 16) == GUIDSize {
			break
		}

		entries = append(entries, next)
	}
	// Double-check that each offset is after the header.
	for i, entry := range entries {
		if entry.Offset < uint32(index) {
			return nil, fmt.Errorf("cert table entry %d has invalid offset into header (size %d): %d",
				i, entry.Offset, index)
		}
	}
	return entries, nil
}

// Unmarshal populates the certTable with the (GUID, Blob) pairs represented in the given bytes.
// The format of the bytes is specified by the SEV SNP API for extended guest requests.
func (c *CertTable) Unmarshal(certs []byte) error {
	certTableHeader, err := ParseSnpCertTableHeader(certs)
	if err != nil {
		return err
	}
	for i, entry := range certTableHeader {
		var next CertTableEntry
		copy(next.GUID[:], entry.GUID[:])
		if entry.Offset+entry.Length > uint32(len(certs)) {
			return fmt.Errorf("cert table entry %d specifies a byte range outside the certificate data block (size %d): offset=%d, length%d", i, len(certs), entry.Offset, entry.Length)
		}
		next.RawCert = make([]byte, entry.Length)
		copy(next.RawCert, certs[entry.Offset:entry.Offset+entry.Length])
		c.Entries = append(c.Entries, next)
	}
	return nil
}

// GetByGUIDString returns the raw bytes for a certificate that matches a key identified by the
// given GUID string.
func (c *CertTable) GetByGUIDString(guid string) ([]byte, error) {
	g, err := uuid.Parse(guid)
	if err != nil {
		return nil, err
	}
	for _, entry := range c.Entries {
		if entry.GUID == g {
			return entry.RawCert, nil
		}
	}
	return nil, fmt.Errorf("cert not found for GUID %s", guid)
}

// CertsFromProto returns the CertTable represented in the given certificate chain.
func CertsFromProto(chain *pb.CertificateChain) *CertTable {
	c := &CertTable{}
	if len(chain.GetArkCert()) != 0 {
		c.Entries = append(c.Entries,
			CertTableEntry{GUID: uuid.MustParse(ArkGUID), RawCert: chain.GetArkCert()})
	}
	if len(chain.GetAskCert()) != 0 {
		c.Entries = append(c.Entries,
			CertTableEntry{GUID: uuid.MustParse(AskGUID), RawCert: chain.GetAskCert()})
	}
	if len(chain.GetVcekCert()) != 0 {
		c.Entries = append(c.Entries,
			CertTableEntry{GUID: uuid.MustParse(VcekGUID), RawCert: chain.GetVcekCert()})
	}
	if len(chain.GetVlekCert()) != 0 {
		c.Entries = append(c.Entries,
			CertTableEntry{GUID: uuid.MustParse(VlekGUID), RawCert: chain.GetVlekCert()})
	}
	for guid, cert := range chain.GetExtras() {
		c.Entries = append(c.Entries,
			CertTableEntry{GUID: uuid.MustParse(guid), RawCert: cert})
	}
	return c
}

// Marshal returns the CertTable in its GUID table ABI format.
func (c *CertTable) Marshal() []byte {
	if len(c.Entries) == 0 {
		return nil
	}
	headerSize := uint32((len(c.Entries) + 1) * CertTableEntrySize)
	var dataSize uint32
	for _, entry := range c.Entries {
		dataSize += uint32(len(entry.RawCert))
	}
	output := make([]byte, dataSize+headerSize)
	cursor := headerSize
	for i, entry := range c.Entries {
		size := uint32(len(entry.RawCert))
		h := &CertTableHeaderEntry{GUID: entry.GUID, Offset: cursor, Length: size}
		copy(output[cursor:cursor+size], entry.RawCert)
		h.Write(output[i*CertTableEntrySize:])
		cursor += size
	}
	return output
}

// Proto returns the certificate chain represented in an extended guest request's
// data pages. The GHCB specification allows any number of entries in the pages,
// so missing certificates aren't an error. If certificates are missing, you can
// choose to fetch them yourself by calling verify.GetAttestationFromReport.
func (c *CertTable) Proto() *pb.CertificateChain {
	vcekGUID := uuid.MustParse(VcekGUID)
	vlekGUID := uuid.MustParse(VlekGUID)
	askGUID := uuid.MustParse(AskGUID)
	arkGUID := uuid.MustParse(ArkGUID)
	result := &pb.CertificateChain{Extras: make(map[string][]byte)}
	for _, entry := range c.Entries {
		switch {
		case entry.GUID == vcekGUID:
			result.VcekCert = entry.RawCert
		case entry.GUID == vlekGUID:
			result.VlekCert = entry.RawCert
		case entry.GUID == askGUID:
			result.AskCert = entry.RawCert
		case entry.GUID == arkGUID:
			result.ArkCert = entry.RawCert
		default:
			result.Extras[entry.GUID.String()] = entry.RawCert
		}
	}
	if len(result.VcekCert) == 0 && len(result.VlekCert) == 0 {
		logger.Warning("Warning: Neither VCEK nor VLEK certificate found in data pages")
	}
	return result
}

// cpuid returns the 4 register results of CPUID[EAX=op,ECX=0].
// See assembly implementations in cpuid_*.s
var cpuid func(op uint32) (eax, ebx, ecx, edx uint32)

// FmsToCpuid1Eax returns the masked CPUID_1_EAX value that represents the given
// family, model, stepping (FMS) values.
func FmsToCpuid1Eax(family, model, stepping byte) uint32 {
	var extendedFamily byte

	familyID := family
	if family >= 0xf {
		extendedFamily = family - 0xf
		familyID = 0xf
	}
	extendedModel := model >> 4
	modelID := model & 0xf
	return (uint32(extendedFamily) << extendedFamilyShift) |
		(uint32(extendedModel) << extendedModelShift) |
		(uint32(familyID) << familyShift) |
		(uint32(modelID) << modelShift) |
		(uint32(stepping & 0xf))
}

// FmsFromCpuid1Eax returns the family, model, stepping (FMS) values extracted from a
// CPUID_1_EAX value.
func FmsFromCpuid1Eax(eax uint32) (byte, byte, byte) {
	// 31:28 reserved
	// 27:20 Extended Family ID
	extendedFamily := byte((eax >> extendedFamilyShift) & 0xff)
	// 19:16 Extended Model ID
	extendedModel := byte((eax >> extendedModelShift) & 0xf)
	// 15:14 reserved
	// 11:8 Family ID
	familyID := byte((eax >> familyShift) & 0xf)
	// 7:4 Model
	modelID := byte((eax >> modelShift) & 0xf)
	// 3:0 Stepping
	family := extendedFamily + familyID
	model := (extendedModel << 4) | modelID
	stepping := byte(eax & 0xf)
	return family, model, stepping
}

// SevProductFromCpuid1Eax returns the SevProduct that is represented by cpuid(1).eax.
func SevProductFromCpuid1Eax(eax uint32) *pb.SevProduct {
	family, model, stepping := FmsFromCpuid1Eax(eax)
	// Ah, Fh, {0h,1h} values from the KDS specification,
	// section "Determining the Product Name".
	var productName pb.SevProduct_SevProductName
	unknown := func() {
		productName = pb.SevProduct_SEV_PRODUCT_UNKNOWN
		stepping = 0 // Reveal nothing.
	}
	// Product information specified by processor programming reference publications.
	switch family {
	case zen3zen4Family:
		switch model {
		case milanModel:
			productName = pb.SevProduct_SEV_PRODUCT_MILAN
		case genoaModel:
			productName = pb.SevProduct_SEV_PRODUCT_GENOA
		default:
			unknown()
		}
	case zen5Family:
		switch model {
		case turinModel:
			productName = pb.SevProduct_SEV_PRODUCT_TURIN
		default:
			unknown()
		}
	default:
		unknown()
	}
	return &pb.SevProduct{
		Name:            productName,
		MachineStepping: &wrapperspb.UInt32Value{Value: uint32(stepping)},
	}
}

// MaskedCpuid1EaxFromSevProduct returns the Cpuid1Eax value expected from the given product
// when masked with CpuidProductMask.
func MaskedCpuid1EaxFromSevProduct(product *pb.SevProduct) uint32 {
	if product == nil {
		return 0
	}
	var family, model, stepping byte
	if product.MachineStepping != nil {
		stepping = byte(product.MachineStepping.Value & 0xf)
	}
	switch product.Name {
	case pb.SevProduct_SEV_PRODUCT_MILAN:
		family = zen3zen4Family
		model = milanModel
	case pb.SevProduct_SEV_PRODUCT_GENOA:
		family = zen3zen4Family
		model = genoaModel
	case pb.SevProduct_SEV_PRODUCT_TURIN:
		family = zen5Family
		model = turinModel
	default:
		return 0
	}
	return FmsToCpuid1Eax(family, model, stepping)
}

// SevProduct returns the SEV product enum for the CPU that runs this
// function. Ought to be called from the client, not the verifier.
func SevProduct() *pb.SevProduct {
	// CPUID[EAX=1] is the processor info. The only bits we care about are in
	// the eax result.
	eax, _, _, _ := cpuid(1)
	return SevProductFromCpuid1Eax(eax & CpuidProductMask)
}

// MakeExtraPlatformInfo returns the representation of platform info needed on top of what an
// attestation report provides in order to interpret it with the help of the AMD KDS.
func MakeExtraPlatformInfo() *ExtraPlatformInfo {
	eax, _, _, _ := cpuid(1)
	return &ExtraPlatformInfo{
		Size:      ExtraPlatformInfoV0Size,
		Cpuid1Eax: eax & CpuidProductMask,
	}
}

// DefaultSevProduct returns the initial product version for a commercially available AMD SEV-SNP chip.
func DefaultSevProduct() *pb.SevProduct {
	return &pb.SevProduct{
		Name:            pb.SevProduct_SEV_PRODUCT_MILAN,
		MachineStepping: &wrapperspb.UInt32Value{Value: 1},
	}
}

// ExtraPlatformInfo represents environment information needed to interpret an attestation report when
// the VCEK certificate is not available in the auxblob.
type ExtraPlatformInfo struct {
	Size      uint32 // Size doubles as Version, following the Linux ABI expansion methodology.
	Cpuid1Eax uint32 // Provides product information
}

// ParseExtraPlatformInfo extracts an ExtraPlatformInfo from a blob if it matches expectations, or
// errors.
func ParseExtraPlatformInfo(data []byte) (*ExtraPlatformInfo, error) {
	if len(data) < ExtraPlatformInfoV0Size {
		return nil, fmt.Errorf("%d bytes is too small for ExtraPlatformInfoSize. Want >= %d bytes",
			len(data), ExtraPlatformInfoV0Size)
	}
	// Populate V0 data.
	result := &ExtraPlatformInfo{
		Size:      binary.LittleEndian.Uint32(data[0:0x04]),
		Cpuid1Eax: binary.LittleEndian.Uint32(data[0x04:0x08]),
	}
	if uint32(len(data)) != result.Size {
		return nil, fmt.Errorf("actual size %d bytes != reported size %d bytes", len(data), result.Size)
	}
	return result, nil
}

// Marshal returns ExtraPlatformInfo in its ABI format or errors.
func (i *ExtraPlatformInfo) Marshal() ([]byte, error) {
	if i.Size != ExtraPlatformInfoV0Size {
		return nil, fmt.Errorf("unsupported ExtraPlatformInfo size %d bytes", i.Size)
	}
	data := make([]byte, ExtraPlatformInfoV0Size)
	binary.LittleEndian.PutUint32(data[0:0x04], i.Size)
	binary.LittleEndian.PutUint32(data[0x04:0x08], i.Cpuid1Eax)
	return data, nil
}

// ExtendPlatformCertTable is a convenience function for parsing a CertTable, adding the
// ExtraPlatformInfoGUID entry, and returning the marshaled extended table.
func ExtendPlatformCertTable(data []byte, info *ExtraPlatformInfo) ([]byte, error) {
	certs := new(CertTable)
	if err := certs.Unmarshal(data); err != nil {
		return nil, err
	}
	// Don't extend the entries with unnecessary information about the platform
	// since the VCEK certificate already contains it in an extension.
	if _, err := certs.GetByGUIDString(VcekGUID); err == nil {
		return data, nil
	}
	// A directly constructed info cannot have a marshaling error.
	extra, err := info.Marshal()
	if err != nil {
		return nil, fmt.Errorf("could not marshal ExtraPlatformInfo: %v", err)
	}
	certs.Entries = append(certs.Entries, CertTableEntry{
		GUID:    uuid.MustParse(ExtraPlatformInfoGUID),
		RawCert: extra,
	})
	return certs.Marshal(), nil
}

// ExtendedPlatformCertTable is a convenience function for parsing a CertTable, adding the
// ExtraPlatformInfoGUID entry, and returning the marshaled extended table.
func ExtendedPlatformCertTable(data []byte) ([]byte, error) {
	return ExtendPlatformCertTable(data, MakeExtraPlatformInfo())
}
