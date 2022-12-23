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
	"github.com/pborman/uuid"
	"golang.org/x/crypto/cryptobyte"
	"golang.org/x/crypto/cryptobyte/asn1"
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
	// https://developer.amd.com/wp-content/resources/56421.pdf

	// VcekGUID is the Versioned Chip Endorsement Key GUID
	VcekGUID = "63da758d-e664-4564-adc5-f4b93be8accd"
	// AskGUID is the AMD signing Key GUID
	AskGUID = "4ab7b379-bbac-4fe4-a02f-05aef327c782"
	// ArkGUID is the AMD Root Key GUID
	ArkGUID = "c0b406a4-a803-4952-9743-3fb6014cd0ae"

	// ExpectedReportVersion is set by the SNP API specification
	// https://www.amd.com/system/files/TechDocs/56860.pdf
	ExpectedReportVersion = 2
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
	validMask := uint64((1 << 21) - 1)
	if guestPolicy&^validMask != 0 {
		return result, fmt.Errorf("policy[63:21] are reserved mbz, got 0x%x", guestPolicy)
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
	var minimumSize = 0x40

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

	reservedAuthor := binary.LittleEndian.Uint32(data[0x48:0x4C])
	if reservedAuthor&0xfffffffe != 0 {
		return nil, fmt.Errorf("mbz bits at offset 0x48 not zero: 0x%08x", reservedAuthor&0xfffffffe)
	}
	r.AuthorKeyEn = reservedAuthor
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
	if err := mbz(data, 0x188, 0x1A0); err != nil {
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
	if version != ExpectedReportVersion {
		return fmt.Errorf("report version is: %d. Expected %d", version, ExpectedReportVersion)
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

	var reservedAuthor uint32
	if r.AuthorKeyEn == 1 {
		reservedAuthor |= 0x01
	}
	binary.LittleEndian.PutUint32(data[0x48:0x4C], reservedAuthor)
	copy(data[0x50:0x90], r.ReportData[:])
	copy(data[0x90:0xC0], r.Measurement[:])
	copy(data[0xC0:0xE0], r.HostData[:])
	copy(data[0xE0:0x110], r.IdKeyDigest[:])
	copy(data[0x110:0x140], r.AuthorKeyDigest[:])
	copy(data[0x140:0x160], r.ReportId[:])
	copy(data[0x160:0x180], r.ReportIdMa[:])
	binary.LittleEndian.PutUint64(data[0x180:0x188], r.ReportedTcb)
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
	h.GUID = clone(data[0:GUIDSize])
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
		next.GUID = make([]byte, GUIDSize)
		copy(next.GUID, entry.GUID)
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
	g := uuid.Parse(guid)
	if g == nil {
		return nil, fmt.Errorf("GUID string format is XXXXXXXX-XXXX-XXXX-XXXXXXXXXXXXXXXX, got %s", guid)
	}
	for _, entry := range c.Entries {
		if uuid.Equal(entry.GUID, g) {
			return entry.RawCert, nil
		}
	}
	return nil, fmt.Errorf("cert not found for GUID %s", guid)
}

// Proto returns the certificate chain represented in an extended guest request's
// data pages. The GHCB specification allows any number of entries in the pages,
// so missing certificates aren't an error. If certificates are missing, you can
// choose to fetch them yourself by calling verify.GetAttestationFromReport.
func (c *CertTable) Proto() *pb.CertificateChain {
	var vcek, ask, ark []byte
	var err error
	vcek, err = c.GetByGUIDString(VcekGUID)
	if err != nil {
		logger.Warningf("Warning: VCEK certificate not found in data pages: %v", err)
	}
	ask, err = c.GetByGUIDString(AskGUID)
	if err != nil {
		logger.Warningf("ASK certificate not found in data pages: %v", err)
	}
	ark, err = c.GetByGUIDString(ArkGUID)
	if err != nil {
		logger.Warningf("ARK certificate not found in data pages: %v", err)
	}
	return &pb.CertificateChain{
		VcekCert: vcek,
		AskCert:  ask,
		ArkCert:  ark,
	}
}
