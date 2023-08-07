// Copyright (c) 2018, Google LLC All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tpm2

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"reflect"

	"github.com/google/go-tpm/tpmutil"
)

// NVPublic contains the public area of an NV index.
type NVPublic struct {
	NVIndex    tpmutil.Handle
	NameAlg    Algorithm
	Attributes NVAttr
	AuthPolicy tpmutil.U16Bytes
	DataSize   uint16
}

type tpmsSensitiveCreate struct {
	UserAuth tpmutil.U16Bytes
	Data     tpmutil.U16Bytes
}

// PCRSelection contains a slice of PCR indexes and a hash algorithm used in
// them.
type PCRSelection struct {
	Hash Algorithm
	PCRs []int
}

type tpmsPCRSelection struct {
	Hash Algorithm
	Size byte
	PCRs tpmutil.RawBytes
}

// Public contains the public area of an object.
type Public struct {
	Type       Algorithm
	NameAlg    Algorithm
	Attributes KeyProp
	AuthPolicy tpmutil.U16Bytes

	// Exactly one of the following fields should be set
	// When encoding/decoding, one will be picked based on Type.

	// RSAParameters contains both [rsa]parameters and [rsa]unique.
	RSAParameters *RSAParams
	// ECCParameters contains both [ecc]parameters and [ecc]unique.
	ECCParameters *ECCParams
	// SymCipherParameters contains both [sym]parameters and [sym]unique.
	SymCipherParameters *SymCipherParams
	// KeyedHashParameters contains both [keyedHash]parameters and [keyedHash]unique.
	KeyedHashParameters *KeyedHashParams
}

// Encode serializes a Public structure in TPM wire format.
func (p Public) Encode() ([]byte, error) {
	head, err := tpmutil.Pack(p.Type, p.NameAlg, p.Attributes, p.AuthPolicy)
	if err != nil {
		return nil, fmt.Errorf("encoding Type, NameAlg, Attributes, AuthPolicy: %v", err)
	}
	var params []byte
	switch p.Type {
	case AlgRSA:
		params, err = p.RSAParameters.encode()
	case AlgKeyedHash:
		params, err = p.KeyedHashParameters.encode()
	case AlgECC:
		params, err = p.ECCParameters.encode()
	case AlgSymCipher:
		params, err = p.SymCipherParameters.encode()
	default:
		err = fmt.Errorf("unsupported type in TPMT_PUBLIC: 0x%x", p.Type)
	}
	if err != nil {
		return nil, fmt.Errorf("encoding RSAParameters, ECCParameters, SymCipherParameters or KeyedHash: %v", err)
	}
	return concat(head, params)
}

// Key returns the (public) key from the public area of an object.
func (p Public) Key() (crypto.PublicKey, error) {
	var pubKey crypto.PublicKey
	switch p.Type {
	case AlgRSA:
		// Endianness of big.Int.Bytes/SetBytes and modulus in the TPM is the same
		// (big-endian).
		pubKey = &rsa.PublicKey{N: p.RSAParameters.Modulus(), E: int(p.RSAParameters.Exponent())}
	case AlgECC:
		curve, ok := toGoCurve[p.ECCParameters.CurveID]
		if !ok {
			return nil, fmt.Errorf("can't map TPM EC curve ID 0x%x to Go elliptic.Curve value", p.ECCParameters.CurveID)
		}
		pubKey = &ecdsa.PublicKey{
			X:     p.ECCParameters.Point.X(),
			Y:     p.ECCParameters.Point.Y(),
			Curve: curve,
		}
	default:
		return nil, fmt.Errorf("unsupported public key type 0x%x", p.Type)
	}
	return pubKey, nil
}

// Name computes the Digest-based Name from the public area of an object.
func (p Public) Name() (Name, error) {
	pubEncoded, err := p.Encode()
	if err != nil {
		return Name{}, err
	}
	hash, err := p.NameAlg.Hash()
	if err != nil {
		return Name{}, err
	}
	nameHash := hash.New()
	nameHash.Write(pubEncoded)
	return Name{
		Digest: &HashValue{
			Alg:   p.NameAlg,
			Value: nameHash.Sum(nil),
		},
	}, nil
}

// MatchesTemplate checks if the Public area has the same algorithms and
// parameters as the provided template. Note that this does not necessarily
// mean that the key was created from this template, as the Unique field is
// both provided in the template and overridden in the key creation process.
func (p Public) MatchesTemplate(template Public) bool {
	if p.Type != template.Type ||
		p.NameAlg != template.NameAlg ||
		p.Attributes != template.Attributes ||
		!bytes.Equal(p.AuthPolicy, template.AuthPolicy) {
		return false
	}
	switch p.Type {
	case AlgRSA:
		return p.RSAParameters.matchesTemplate(template.RSAParameters)
	case AlgECC:
		return p.ECCParameters.matchesTemplate(template.ECCParameters)
	case AlgSymCipher:
		return p.SymCipherParameters.matchesTemplate(template.SymCipherParameters)
	case AlgKeyedHash:
		return p.KeyedHashParameters.matchesTemplate(template.KeyedHashParameters)
	default:
		return true
	}
}

// DecodePublic decodes a TPMT_PUBLIC message. No error is returned if
// the input has extra trailing data.
func DecodePublic(buf []byte) (Public, error) {
	in := bytes.NewBuffer(buf)
	var pub Public
	var err error
	if err = tpmutil.UnpackBuf(in, &pub.Type, &pub.NameAlg, &pub.Attributes, &pub.AuthPolicy); err != nil {
		return pub, fmt.Errorf("decoding TPMT_PUBLIC: %v", err)
	}

	switch pub.Type {
	case AlgRSA:
		pub.RSAParameters, err = decodeRSAParams(in)
	case AlgECC:
		pub.ECCParameters, err = decodeECCParams(in)
	case AlgSymCipher:
		pub.SymCipherParameters, err = decodeSymCipherParams(in)
	case AlgKeyedHash:
		pub.KeyedHashParameters, err = decodeKeyedHashParams(in)
	default:
		err = fmt.Errorf("unsupported type in TPMT_PUBLIC: 0x%x", pub.Type)
	}
	return pub, err
}

// RSAParams represents parameters of an RSA key pair:
// both the TPMS_RSA_PARMS and the TPM2B_PUBLIC_KEY_RSA.
//
// Symmetric and Sign may be nil, depending on key Attributes in Public.
//
// ExponentRaw and ModulusRaw are the actual data encoded in the template, which
// is useful for templates that differ in zero-padding, for example.
type RSAParams struct {
	Symmetric   *SymScheme
	Sign        *SigScheme
	KeyBits     uint16
	ExponentRaw uint32
	ModulusRaw  tpmutil.U16Bytes
}

// Exponent returns the RSA exponent value represented by ExponentRaw, handling
// the fact that an exponent of 0 represents a value of 65537 (2^16 + 1).
func (p *RSAParams) Exponent() uint32 {
	if p.ExponentRaw == 0 {
		return defaultRSAExponent
	}
	return p.ExponentRaw
}

// Modulus returns the RSA modulus value represented by ModulusRaw, handling the
// fact that the same modulus value can have multiple different representations.
func (p *RSAParams) Modulus() *big.Int {
	return new(big.Int).SetBytes(p.ModulusRaw)
}

func (p *RSAParams) matchesTemplate(t *RSAParams) bool {
	return reflect.DeepEqual(p.Symmetric, t.Symmetric) &&
		reflect.DeepEqual(p.Sign, t.Sign) &&
		p.KeyBits == t.KeyBits && p.ExponentRaw == t.ExponentRaw
}

func (p *RSAParams) encode() ([]byte, error) {
	if p == nil {
		return nil, nil
	}
	sym, err := p.Symmetric.encode()
	if err != nil {
		return nil, fmt.Errorf("encoding Symmetric: %v", err)
	}
	sig, err := p.Sign.encode()
	if err != nil {
		return nil, fmt.Errorf("encoding Sign: %v", err)
	}
	rest, err := tpmutil.Pack(p.KeyBits, p.ExponentRaw, p.ModulusRaw)
	if err != nil {
		return nil, fmt.Errorf("encoding KeyBits, Exponent, Modulus: %v", err)
	}
	return concat(sym, sig, rest)
}

func decodeRSAParams(in *bytes.Buffer) (*RSAParams, error) {
	var params RSAParams
	var err error

	if params.Symmetric, err = decodeSymScheme(in); err != nil {
		return nil, fmt.Errorf("decoding Symmetric: %v", err)
	}
	if params.Sign, err = decodeSigScheme(in); err != nil {
		return nil, fmt.Errorf("decoding Sign: %v", err)
	}
	if err := tpmutil.UnpackBuf(in, &params.KeyBits, &params.ExponentRaw, &params.ModulusRaw); err != nil {
		return nil, fmt.Errorf("decoding KeyBits, Exponent, Modulus: %v", err)
	}
	return &params, nil
}

// ECCParams represents parameters of an ECC key pair:
// both the TPMS_ECC_PARMS and the TPMS_ECC_POINT.
//
// Symmetric, Sign and KDF may be nil, depending on key Attributes in Public.
type ECCParams struct {
	Symmetric *SymScheme
	Sign      *SigScheme
	CurveID   EllipticCurve
	KDF       *KDFScheme
	Point     ECPoint
}

// ECPoint represents a ECC coordinates for a point using byte buffers.
type ECPoint struct {
	XRaw, YRaw tpmutil.U16Bytes
}

// X returns the X Point value reprsented by XRaw.
func (p ECPoint) X() *big.Int {
	return new(big.Int).SetBytes(p.XRaw)
}

// Y returns the Y Point value reprsented by YRaw.
func (p ECPoint) Y() *big.Int {
	return new(big.Int).SetBytes(p.YRaw)
}

func (p *ECCParams) matchesTemplate(t *ECCParams) bool {
	return reflect.DeepEqual(p.Symmetric, t.Symmetric) &&
		reflect.DeepEqual(p.Sign, t.Sign) &&
		p.CurveID == t.CurveID && reflect.DeepEqual(p.KDF, t.KDF)
}

func (p *ECCParams) encode() ([]byte, error) {
	if p == nil {
		return nil, nil
	}
	sym, err := p.Symmetric.encode()
	if err != nil {
		return nil, fmt.Errorf("encoding Symmetric: %v", err)
	}
	sig, err := p.Sign.encode()
	if err != nil {
		return nil, fmt.Errorf("encoding Sign: %v", err)
	}
	curve, err := tpmutil.Pack(p.CurveID)
	if err != nil {
		return nil, fmt.Errorf("encoding CurveID: %v", err)
	}
	kdf, err := p.KDF.encode()
	if err != nil {
		return nil, fmt.Errorf("encoding KDF: %v", err)
	}
	point, err := tpmutil.Pack(p.Point.XRaw, p.Point.YRaw)
	if err != nil {
		return nil, fmt.Errorf("encoding Point: %v", err)
	}
	return concat(sym, sig, curve, kdf, point)
}

func decodeECCParams(in *bytes.Buffer) (*ECCParams, error) {
	var params ECCParams
	var err error

	if params.Symmetric, err = decodeSymScheme(in); err != nil {
		return nil, fmt.Errorf("decoding Symmetric: %v", err)
	}
	if params.Sign, err = decodeSigScheme(in); err != nil {
		return nil, fmt.Errorf("decoding Sign: %v", err)
	}
	if err := tpmutil.UnpackBuf(in, &params.CurveID); err != nil {
		return nil, fmt.Errorf("decoding CurveID: %v", err)
	}
	if params.KDF, err = decodeKDFScheme(in); err != nil {
		return nil, fmt.Errorf("decoding KDF: %v", err)
	}
	if err := tpmutil.UnpackBuf(in, &params.Point.XRaw, &params.Point.YRaw); err != nil {
		return nil, fmt.Errorf("decoding Point: %v", err)
	}
	return &params, nil
}

// SymCipherParams represents parameters of a symmetric block cipher TPM object:
// both the TPMS_SYMCIPHER_PARMS and the TPM2B_DIGEST (hash of the key).
type SymCipherParams struct {
	Symmetric *SymScheme
	Unique    tpmutil.U16Bytes
}

func (p *SymCipherParams) matchesTemplate(t *SymCipherParams) bool {
	return reflect.DeepEqual(p.Symmetric, t.Symmetric)
}

func (p *SymCipherParams) encode() ([]byte, error) {
	sym, err := p.Symmetric.encode()
	if err != nil {
		return nil, fmt.Errorf("encoding Symmetric: %v", err)
	}
	unique, err := tpmutil.Pack(p.Unique)
	if err != nil {
		return nil, fmt.Errorf("encoding Unique: %v", err)
	}
	return concat(sym, unique)
}

func decodeSymCipherParams(in *bytes.Buffer) (*SymCipherParams, error) {
	var params SymCipherParams
	var err error

	if params.Symmetric, err = decodeSymScheme(in); err != nil {
		return nil, fmt.Errorf("decoding Symmetric: %v", err)
	}
	if err := tpmutil.UnpackBuf(in, &params.Unique); err != nil {
		return nil, fmt.Errorf("decoding Unique: %v", err)
	}
	return &params, nil
}

// KeyedHashParams represents parameters of a keyed hash TPM object:
// both the TPMS_KEYEDHASH_PARMS and the TPM2B_DIGEST (hash of the key).
type KeyedHashParams struct {
	Alg    Algorithm
	Hash   Algorithm
	KDF    Algorithm
	Unique tpmutil.U16Bytes
}

func (p *KeyedHashParams) matchesTemplate(t *KeyedHashParams) bool {
	if p.Alg != t.Alg {
		return false
	}
	switch p.Alg {
	case AlgHMAC:
		return p.Hash == t.Hash
	case AlgXOR:
		return p.Hash == t.Hash && p.KDF == t.KDF
	default:
		return true
	}
}

func (p *KeyedHashParams) encode() ([]byte, error) {
	if p == nil {
		return tpmutil.Pack(AlgNull, tpmutil.U16Bytes(nil))
	}
	var params []byte
	var err error
	switch p.Alg {
	case AlgNull:
		params, err = tpmutil.Pack(p.Alg)
	case AlgHMAC:
		params, err = tpmutil.Pack(p.Alg, p.Hash)
	case AlgXOR:
		params, err = tpmutil.Pack(p.Alg, p.Hash, p.KDF)
	default:
		err = fmt.Errorf("unsupported KeyedHash Algorithm: 0x%x", p.Alg)
	}
	if err != nil {
		return nil, fmt.Errorf("encoding Alg Params: %v", err)
	}
	unique, err := tpmutil.Pack(p.Unique)
	if err != nil {
		return nil, fmt.Errorf("encoding Unique: %v", err)
	}
	return concat(params, unique)
}

func decodeKeyedHashParams(in *bytes.Buffer) (*KeyedHashParams, error) {
	var p KeyedHashParams
	var err error
	if err = tpmutil.UnpackBuf(in, &p.Alg); err != nil {
		return nil, fmt.Errorf("decoding Alg: %v", err)
	}
	switch p.Alg {
	case AlgNull:
		err = nil
	case AlgHMAC:
		err = tpmutil.UnpackBuf(in, &p.Hash)
	case AlgXOR:
		err = tpmutil.UnpackBuf(in, &p.Hash, &p.KDF)
	default:
		err = fmt.Errorf("unsupported KeyedHash Algorithm: 0x%x", p.Alg)
	}
	if err != nil {
		return nil, fmt.Errorf("decoding Alg Params: %v", err)
	}
	if err = tpmutil.UnpackBuf(in, &p.Unique); err != nil {
		return nil, fmt.Errorf("decoding Unique: %v", err)
	}
	return &p, nil
}

// SymScheme represents a symmetric encryption scheme.
// Known in the specification by TPMT_SYM_DEF_OBJECT.
type SymScheme struct {
	Alg     Algorithm
	KeyBits uint16
	Mode    Algorithm
}

func (s *SymScheme) encode() ([]byte, error) {
	if s == nil || s.Alg.IsNull() {
		return tpmutil.Pack(AlgNull)
	}
	return tpmutil.Pack(s.Alg, s.KeyBits, s.Mode)
}

func decodeSymScheme(in *bytes.Buffer) (*SymScheme, error) {
	var scheme SymScheme
	if err := tpmutil.UnpackBuf(in, &scheme.Alg); err != nil {
		return nil, fmt.Errorf("decoding Alg: %v", err)
	}
	if scheme.Alg == AlgNull {
		return nil, nil
	}
	if err := tpmutil.UnpackBuf(in, &scheme.KeyBits, &scheme.Mode); err != nil {
		return nil, fmt.Errorf("decoding KeyBits, Mode: %v", err)
	}
	return &scheme, nil
}

// AsymScheme represents am asymmetric encryption scheme.
type AsymScheme struct {
	Alg  Algorithm
	Hash Algorithm
}

func (s *AsymScheme) encode() ([]byte, error) {
	if s == nil || s.Alg.IsNull() {
		return tpmutil.Pack(AlgNull)
	}
	if s.Alg.UsesHash() {
		return tpmutil.Pack(s.Alg, s.Hash)
	}
	return tpmutil.Pack(s.Alg)
}

// SigScheme represents a signing scheme.
type SigScheme struct {
	Alg   Algorithm
	Hash  Algorithm
	Count uint32
}

func (s *SigScheme) encode() ([]byte, error) {
	if s == nil || s.Alg.IsNull() {
		return tpmutil.Pack(AlgNull)
	}
	if s.Alg.UsesCount() {
		return tpmutil.Pack(s.Alg, s.Hash, s.Count)
	}
	return tpmutil.Pack(s.Alg, s.Hash)
}

func decodeSigScheme(in *bytes.Buffer) (*SigScheme, error) {
	var scheme SigScheme
	if err := tpmutil.UnpackBuf(in, &scheme.Alg); err != nil {
		return nil, fmt.Errorf("decoding Alg: %v", err)
	}
	if scheme.Alg == AlgNull {
		return nil, nil
	}
	if err := tpmutil.UnpackBuf(in, &scheme.Hash); err != nil {
		return nil, fmt.Errorf("decoding Hash: %v", err)
	}
	if scheme.Alg.UsesCount() {
		if err := tpmutil.UnpackBuf(in, &scheme.Count); err != nil {
			return nil, fmt.Errorf("decoding Count: %v", err)
		}
	}
	return &scheme, nil
}

// KDFScheme represents a KDF (Key Derivation Function) scheme.
type KDFScheme struct {
	Alg  Algorithm
	Hash Algorithm
}

func (s *KDFScheme) encode() ([]byte, error) {
	if s == nil || s.Alg.IsNull() {
		return tpmutil.Pack(AlgNull)
	}
	return tpmutil.Pack(s.Alg, s.Hash)
}

func decodeKDFScheme(in *bytes.Buffer) (*KDFScheme, error) {
	var scheme KDFScheme
	if err := tpmutil.UnpackBuf(in, &scheme.Alg); err != nil {
		return nil, fmt.Errorf("decoding Alg: %v", err)
	}
	if scheme.Alg == AlgNull {
		return nil, nil
	}
	if err := tpmutil.UnpackBuf(in, &scheme.Hash); err != nil {
		return nil, fmt.Errorf("decoding Hash: %v", err)
	}
	return &scheme, nil
}

// Signature combines all possible signatures from RSA and ECC keys. Only one
// of RSA or ECC will be populated.
type Signature struct {
	Alg Algorithm
	RSA *SignatureRSA
	ECC *SignatureECC
}

// Encode serializes a Signature structure in TPM wire format.
func (s Signature) Encode() ([]byte, error) {
	head, err := tpmutil.Pack(s.Alg)
	if err != nil {
		return nil, fmt.Errorf("encoding Alg: %v", err)
	}
	var signature []byte
	switch s.Alg {
	case AlgRSASSA, AlgRSAPSS:
		if signature, err = tpmutil.Pack(s.RSA); err != nil {
			return nil, fmt.Errorf("encoding RSA: %v", err)
		}
	case AlgECDSA:
		signature, err = tpmutil.Pack(s.ECC.HashAlg, tpmutil.U16Bytes(s.ECC.R.Bytes()), tpmutil.U16Bytes(s.ECC.S.Bytes()))
		if err != nil {
			return nil, fmt.Errorf("encoding ECC: %v", err)
		}
	}
	return concat(head, signature)
}

// DecodeSignature decodes a serialized TPMT_SIGNATURE structure.
func DecodeSignature(in *bytes.Buffer) (*Signature, error) {
	var sig Signature
	if err := tpmutil.UnpackBuf(in, &sig.Alg); err != nil {
		return nil, fmt.Errorf("decoding Alg: %v", err)
	}
	switch sig.Alg {
	case AlgRSASSA, AlgRSAPSS:
		sig.RSA = new(SignatureRSA)
		if err := tpmutil.UnpackBuf(in, sig.RSA); err != nil {
			return nil, fmt.Errorf("decoding RSA: %v", err)
		}
	case AlgECDSA:
		sig.ECC = new(SignatureECC)
		var r, s tpmutil.U16Bytes
		if err := tpmutil.UnpackBuf(in, &sig.ECC.HashAlg, &r, &s); err != nil {
			return nil, fmt.Errorf("decoding ECC: %v", err)
		}
		sig.ECC.R = big.NewInt(0).SetBytes(r)
		sig.ECC.S = big.NewInt(0).SetBytes(s)
	default:
		return nil, fmt.Errorf("unsupported signature algorithm 0x%x", sig.Alg)
	}
	return &sig, nil
}

// SignatureRSA is an RSA-specific signature value.
type SignatureRSA struct {
	HashAlg   Algorithm
	Signature tpmutil.U16Bytes
}

// SignatureECC is an ECC-specific signature value.
type SignatureECC struct {
	HashAlg Algorithm
	R       *big.Int
	S       *big.Int
}

// Private contains private section of a TPM key.
type Private struct {
	Type      Algorithm
	AuthValue tpmutil.U16Bytes
	SeedValue tpmutil.U16Bytes
	Sensitive tpmutil.U16Bytes
}

// Encode serializes a Private structure in TPM wire format.
func (p Private) Encode() ([]byte, error) {
	if p.Type.IsNull() {
		return nil, nil
	}
	return tpmutil.Pack(p)
}

// AttestationData contains data attested by TPM commands (like Certify).
type AttestationData struct {
	Magic                uint32
	Type                 tpmutil.Tag
	QualifiedSigner      Name
	ExtraData            tpmutil.U16Bytes
	ClockInfo            ClockInfo
	FirmwareVersion      uint64
	AttestedCertifyInfo  *CertifyInfo
	AttestedQuoteInfo    *QuoteInfo
	AttestedCreationInfo *CreationInfo
}

// DecodeAttestationData decode a TPMS_ATTEST message. No error is returned if
// the input has extra trailing data.
func DecodeAttestationData(in []byte) (*AttestationData, error) {
	buf := bytes.NewBuffer(in)

	var ad AttestationData
	if err := tpmutil.UnpackBuf(buf, &ad.Magic, &ad.Type); err != nil {
		return nil, fmt.Errorf("decoding Magic/Type: %v", err)
	}
	// All attestation structures have the magic prefix
	// TPMS_GENERATED_VALUE to symbolize they were created by
	// the TPM when signed with an AK.
	if ad.Magic != 0xff544347 {
		return nil, fmt.Errorf("incorrect magic value: %x", ad.Magic)
	}

	n, err := DecodeName(buf)
	if err != nil {
		return nil, fmt.Errorf("decoding QualifiedSigner: %v", err)
	}
	ad.QualifiedSigner = *n
	if err := tpmutil.UnpackBuf(buf, &ad.ExtraData, &ad.ClockInfo, &ad.FirmwareVersion); err != nil {
		return nil, fmt.Errorf("decoding ExtraData/ClockInfo/FirmwareVersion: %v", err)
	}

	// The spec specifies several other types of attestation data. We only need
	// parsing of Certify & Creation attestation data for now. If you need
	// support for other attestation types, add them here.
	switch ad.Type {
	case TagAttestCertify:
		if ad.AttestedCertifyInfo, err = decodeCertifyInfo(buf); err != nil {
			return nil, fmt.Errorf("decoding AttestedCertifyInfo: %v", err)
		}
	case TagAttestCreation:
		if ad.AttestedCreationInfo, err = decodeCreationInfo(buf); err != nil {
			return nil, fmt.Errorf("decoding AttestedCreationInfo: %v", err)
		}
	case TagAttestQuote:
		if ad.AttestedQuoteInfo, err = decodeQuoteInfo(buf); err != nil {
			return nil, fmt.Errorf("decoding AttestedQuoteInfo: %v", err)
		}
	default:
		return nil, fmt.Errorf("only Quote, Certify & Creation attestation structures are supported, got type 0x%x", ad.Type)
	}

	return &ad, nil
}

// Encode serializes an AttestationData structure in TPM wire format.
func (ad AttestationData) Encode() ([]byte, error) {
	head, err := tpmutil.Pack(ad.Magic, ad.Type)
	if err != nil {
		return nil, fmt.Errorf("encoding Magic, Type: %v", err)
	}
	signer, err := ad.QualifiedSigner.Encode()
	if err != nil {
		return nil, fmt.Errorf("encoding QualifiedSigner: %v", err)
	}
	tail, err := tpmutil.Pack(ad.ExtraData, ad.ClockInfo, ad.FirmwareVersion)
	if err != nil {
		return nil, fmt.Errorf("encoding ExtraData, ClockInfo, FirmwareVersion: %v", err)
	}

	var info []byte
	switch ad.Type {
	case TagAttestCertify:
		if info, err = ad.AttestedCertifyInfo.encode(); err != nil {
			return nil, fmt.Errorf("encoding AttestedCertifyInfo: %v", err)
		}
	case TagAttestCreation:
		if info, err = ad.AttestedCreationInfo.encode(); err != nil {
			return nil, fmt.Errorf("encoding AttestedCreationInfo: %v", err)
		}
	case TagAttestQuote:
		if info, err = ad.AttestedQuoteInfo.encode(); err != nil {
			return nil, fmt.Errorf("encoding AttestedQuoteInfo: %v", err)
		}
	default:
		return nil, fmt.Errorf("only Quote, Certify & Creation attestation structures are supported, got type 0x%x", ad.Type)
	}

	return concat(head, signer, tail, info)
}

// CreationInfo contains Creation-specific data for TPMS_ATTEST.
type CreationInfo struct {
	Name Name
	// Most TPM2B_Digest structures contain a TPMU_HA structure
	// and get parsed to HashValue. This is never the case for the
	// digest in TPMS_CREATION_INFO.
	OpaqueDigest tpmutil.U16Bytes
}

func decodeCreationInfo(in *bytes.Buffer) (*CreationInfo, error) {
	var ci CreationInfo

	n, err := DecodeName(in)
	if err != nil {
		return nil, fmt.Errorf("decoding Name: %v", err)
	}
	ci.Name = *n

	if err := tpmutil.UnpackBuf(in, &ci.OpaqueDigest); err != nil {
		return nil, fmt.Errorf("decoding Digest: %v", err)
	}

	return &ci, nil
}

func (ci CreationInfo) encode() ([]byte, error) {
	n, err := ci.Name.Encode()
	if err != nil {
		return nil, fmt.Errorf("encoding Name: %v", err)
	}

	d, err := tpmutil.Pack(ci.OpaqueDigest)
	if err != nil {
		return nil, fmt.Errorf("encoding Digest: %v", err)
	}

	return concat(n, d)
}

// CertifyInfo contains Certify-specific data for TPMS_ATTEST.
type CertifyInfo struct {
	Name          Name
	QualifiedName Name
}

func decodeCertifyInfo(in *bytes.Buffer) (*CertifyInfo, error) {
	var ci CertifyInfo

	n, err := DecodeName(in)
	if err != nil {
		return nil, fmt.Errorf("decoding Name: %v", err)
	}
	ci.Name = *n

	n, err = DecodeName(in)
	if err != nil {
		return nil, fmt.Errorf("decoding QualifiedName: %v", err)
	}
	ci.QualifiedName = *n

	return &ci, nil
}

func (ci CertifyInfo) encode() ([]byte, error) {
	n, err := ci.Name.Encode()
	if err != nil {
		return nil, fmt.Errorf("encoding Name: %v", err)
	}
	qn, err := ci.QualifiedName.Encode()
	if err != nil {
		return nil, fmt.Errorf("encoding QualifiedName: %v", err)
	}
	return concat(n, qn)
}

// QuoteInfo represents a TPMS_QUOTE_INFO structure.
type QuoteInfo struct {
	PCRSelection PCRSelection
	PCRDigest    tpmutil.U16Bytes
}

func decodeQuoteInfo(in *bytes.Buffer) (*QuoteInfo, error) {
	var out QuoteInfo
	sel, err := decodeOneTPMLPCRSelection(in)
	if err != nil {
		return nil, fmt.Errorf("decoding PCRSelection: %v", err)
	}
	out.PCRSelection = sel

	if err := tpmutil.UnpackBuf(in, &out.PCRDigest); err != nil {
		return nil, fmt.Errorf("decoding PCRDigest: %v", err)
	}
	return &out, nil
}

func (qi QuoteInfo) encode() ([]byte, error) {
	sel, err := encodeTPMLPCRSelection(qi.PCRSelection)
	if err != nil {
		return nil, fmt.Errorf("encoding PCRSelection: %v", err)
	}

	digest, err := tpmutil.Pack(qi.PCRDigest)
	if err != nil {
		return nil, fmt.Errorf("encoding PCRDigest: %v", err)
	}

	return concat(sel, digest)
}

// IDObject represents an encrypted credential bound to a TPM object.
type IDObject struct {
	IntegrityHMAC tpmutil.U16Bytes
	// EncIdentity is packed raw, as the bytes representing the size
	// of the credential value are present within the encrypted blob.
	EncIdentity tpmutil.RawBytes
}

// CreationData describes the attributes and environment for an object created
// on the TPM. This structure encodes/decodes to/from TPMS_CREATION_DATA.
type CreationData struct {
	PCRSelection        PCRSelection
	PCRDigest           tpmutil.U16Bytes
	Locality            byte
	ParentNameAlg       Algorithm
	ParentName          Name
	ParentQualifiedName Name
	OutsideInfo         tpmutil.U16Bytes
}

// EncodeCreationData encodes byte array to TPMS_CREATION_DATA message.
func (cd *CreationData) EncodeCreationData() ([]byte, error) {
	sel, err := encodeTPMLPCRSelection(cd.PCRSelection)
	if err != nil {
		return nil, fmt.Errorf("encoding PCRSelection: %v", err)
	}
	d, err := tpmutil.Pack(cd.PCRDigest, cd.Locality, cd.ParentNameAlg)
	if err != nil {
		return nil, fmt.Errorf("encoding PCRDigest, Locality, ParentNameAlg: %v", err)
	}
	pn, err := cd.ParentName.Encode()
	if err != nil {
		return nil, fmt.Errorf("encoding ParentName: %v", err)
	}
	pqn, err := cd.ParentQualifiedName.Encode()
	if err != nil {
		return nil, fmt.Errorf("encoding ParentQualifiedName: %v", err)
	}
	o, err := tpmutil.Pack(cd.OutsideInfo)
	if err != nil {
		return nil, fmt.Errorf("encoding OutsideInfo: %v", err)
	}
	return concat(sel, d, pn, pqn, o)
}

// DecodeCreationData decodes a TPMS_CREATION_DATA message. No error is
// returned if the input has extra trailing data.
func DecodeCreationData(buf []byte) (*CreationData, error) {
	in := bytes.NewBuffer(buf)
	var out CreationData

	sel, err := decodeOneTPMLPCRSelection(in)
	if err != nil {
		return nil, fmt.Errorf("decodeOneTPMLPCRSelection returned error %v", err)
	}
	out.PCRSelection = sel

	if err := tpmutil.UnpackBuf(in, &out.PCRDigest, &out.Locality, &out.ParentNameAlg); err != nil {
		return nil, fmt.Errorf("decoding PCRDigest, Locality, ParentNameAlg: %v", err)
	}

	n, err := DecodeName(in)
	if err != nil {
		return nil, fmt.Errorf("decoding ParentName: %v", err)
	}
	out.ParentName = *n
	if n, err = DecodeName(in); err != nil {
		return nil, fmt.Errorf("decoding ParentQualifiedName: %v", err)
	}
	out.ParentQualifiedName = *n

	if err := tpmutil.UnpackBuf(in, &out.OutsideInfo); err != nil {
		return nil, fmt.Errorf("decoding OutsideInfo: %v", err)
	}

	return &out, nil
}

// Name represents a TPM2B_NAME, a name for TPM entities. Only one of
// Handle or Digest should be set.
type Name struct {
	Handle *tpmutil.Handle
	Digest *HashValue
}

// DecodeName deserializes a Name hash from the TPM wire format.
func DecodeName(in *bytes.Buffer) (*Name, error) {
	var nameBuf tpmutil.U16Bytes
	if err := tpmutil.UnpackBuf(in, &nameBuf); err != nil {
		return nil, err
	}

	name := new(Name)
	switch len(nameBuf) {
	case 0:
		// No name is present.
	case 4:
		name.Handle = new(tpmutil.Handle)
		if err := tpmutil.UnpackBuf(bytes.NewBuffer(nameBuf), name.Handle); err != nil {
			return nil, fmt.Errorf("decoding Handle: %v", err)
		}
	default:
		var err error
		name.Digest, err = decodeHashValue(bytes.NewBuffer(nameBuf))
		if err != nil {
			return nil, fmt.Errorf("decoding Digest: %v", err)
		}
	}
	return name, nil
}

// Encode serializes a Name hash into the TPM wire format.
func (n Name) Encode() ([]byte, error) {
	var buf []byte
	var err error
	switch {
	case n.Handle != nil:
		if buf, err = tpmutil.Pack(*n.Handle); err != nil {
			return nil, fmt.Errorf("encoding Handle: %v", err)
		}
	case n.Digest != nil:
		if buf, err = n.Digest.Encode(); err != nil {
			return nil, fmt.Errorf("encoding Digest: %v", err)
		}
	default:
		// Name is empty, which is valid.
	}
	return tpmutil.Pack(tpmutil.U16Bytes(buf))
}

// MatchesPublic compares Digest in Name against given Public structure. Note:
// this only works for regular Names, not Qualified Names.
func (n Name) MatchesPublic(p Public) (bool, error) {
	if n.Digest == nil {
		return false, errors.New("Name doesn't have a Digest, can't compare to Public")
	}
	expected, err := p.Name()
	if err != nil {
		return false, err
	}
	// No secrets, so no constant-time comparison
	return bytes.Equal(expected.Digest.Value, n.Digest.Value), nil
}

// HashValue is an algorithm-specific hash value.
type HashValue struct {
	Alg   Algorithm
	Value tpmutil.U16Bytes
}

func decodeHashValue(in *bytes.Buffer) (*HashValue, error) {
	var hv HashValue
	if err := tpmutil.UnpackBuf(in, &hv.Alg); err != nil {
		return nil, fmt.Errorf("decoding Alg: %v", err)
	}
	hfn, err := hv.Alg.Hash()
	if err != nil {
		return nil, err
	}
	hv.Value = make(tpmutil.U16Bytes, hfn.Size())
	if _, err := in.Read(hv.Value); err != nil {
		return nil, fmt.Errorf("decoding Value: %v", err)
	}
	return &hv, nil
}

// Encode represents the given hash value as a TPMT_HA structure.
func (hv HashValue) Encode() ([]byte, error) {
	return tpmutil.Pack(hv.Alg, tpmutil.RawBytes(hv.Value))
}

// ClockInfo contains TPM state info included in AttestationData.
type ClockInfo struct {
	Clock        uint64
	ResetCount   uint32
	RestartCount uint32
	Safe         byte
}

// AlgorithmAttributes represents a TPMA_ALGORITHM value.
type AlgorithmAttributes uint32

// AlgorithmDescription represents a TPMS_ALGORITHM_DESCRIPTION structure.
type AlgorithmDescription struct {
	ID         Algorithm
	Attributes AlgorithmAttributes
}

// TaggedProperty represents a TPMS_TAGGED_PROPERTY structure.
type TaggedProperty struct {
	Tag   TPMProp
	Value uint32
}

// Ticket represents evidence the TPM previously processed
// information.
type Ticket struct {
	Type      tpmutil.Tag
	Hierarchy tpmutil.Handle
	Digest    tpmutil.U16Bytes
}

// AuthCommand represents a TPMS_AUTH_COMMAND. This structure encapsulates parameters
// which authorize the use of a given handle or parameter.
type AuthCommand struct {
	Session    tpmutil.Handle
	Nonce      tpmutil.U16Bytes
	Attributes SessionAttributes
	Auth       tpmutil.U16Bytes
}

// TPMLDigest represents the TPML_Digest structure
// It is used to convey a list of digest values.
// This type is used in TPM2_PolicyOR() and in TPM2_PCR_Read()
type TPMLDigest struct {
	Digests []tpmutil.U16Bytes
}

// Encode converts the TPMLDigest structure into a byte slice
func (list *TPMLDigest) Encode() ([]byte, error) {
	res, err := tpmutil.Pack(uint32(len(list.Digests)))
	if err != nil {
		return nil, err
	}
	for _, item := range list.Digests {
		b, err := tpmutil.Pack(item)
		if err != nil {
			return nil, err
		}
		res = append(res, b...)

	}
	return res, nil
}

// DecodeTPMLDigest decodes a TPML_Digest part of a message.
func DecodeTPMLDigest(buf []byte) (*TPMLDigest, error) {
	in := bytes.NewBuffer(buf)
	var tpmld TPMLDigest
	var count uint32
	if err := binary.Read(in, binary.BigEndian, &count); err != nil {
		return nil, fmt.Errorf("decoding TPML_Digest: %v", err)
	}
	for in.Len() > 0 {
		var hash tpmutil.U16Bytes
		if err := hash.TPMUnmarshal(in); err != nil {
			return nil, err
		}
		tpmld.Digests = append(tpmld.Digests, hash)
	}
	if count != uint32(len(tpmld.Digests)) {
		return nil, fmt.Errorf("expected size and read size does not match")
	}
	return &tpmld, nil
}
