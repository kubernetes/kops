// Copyright 2023 Google LLC
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

// Package abi provides Go representations and conversions for TDX attestation
// data structures
package abi

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"

	pb "github.com/google/go-tdx-guest/proto/tdx"
	"golang.org/x/crypto/cryptobyte"
	"golang.org/x/crypto/cryptobyte/asn1"
)

const (
	// QuoteMinSize is the minimum specified size of TDX generated quote
	QuoteMinSize = 0x3FC
	// QuoteVersion currently in support
	QuoteVersion = 4
	// AttestationKeyType supported value
	AttestationKeyType = 2 // (ECDSA-256-with-P-256 curve)
	// TeeTDX  for Attestation
	TeeTDX = 0x00000081
	// TeeTcbSvnSize is the size of TEE_TCB_SVN field in TdQuoteBody
	TeeTcbSvnSize = 0x10
	// MrSeamSize is the size  of MR_SEAM field in TdQuoteBody
	MrSeamSize = 0x30
	// TdAttributesSize is the size  of TD_ATTRIBUTES field in TdQuoteBody
	TdAttributesSize = 0x08
	// XfamSize is the size of XFAM field in TdQuoteBody
	XfamSize = 0x08
	// MrTdSize is the size of MR_TD field in TdQuoteBody
	MrTdSize = 0x30
	// MrConfigIDSize is the size of MR_CONFIG_ID field in TdQuoteBody
	MrConfigIDSize = 0x30
	// MrOwnerSize is the size of MR_OWNER field in TdQuoteBody
	MrOwnerSize = 0x30
	// MrOwnerConfigSize is the size of MR_OWNER_CONFIG field in TdQuoteBody
	MrOwnerConfigSize = 0x30
	// RtmrSize is the size of Runtime extendable measurement register
	RtmrSize = 0x30
	// ReportDataSize is the size of ReportData field in TdQuoteBody
	ReportDataSize = 0x40
	// QeVendorIDSize is the size of QeVendorID field in Header
	QeVendorIDSize = 0x10

	userDataSize                   = 0x14
	qeReportCertificationDataType  = 0x6
	pckReportCertificationDataType = 0x5
	qeReportSize                   = 0x180
	headerSize                     = 0x30
	tdQuoteBodySize                = 0x248
	qeSvnSize                      = 0x2
	pceSvnSize                     = 0x2

	mrSignerSeamSize                        = 0x30
	seamAttributesSize                      = 0x08
	cpuSvnSize                              = 0x10
	reserved1Size                           = 0x1C
	attributesSize                          = 0x10
	mrEnclaveSize                           = 0x20
	reserved2Size                           = 0x20
	mrSignerSize                            = 0x20
	reserved3Size                           = 0x60
	reserved4Size                           = 0x3C
	signatureSize                           = 0x40
	attestationKeySize                      = 0x40
	pckCertificateChainKnownSize            = 0x06
	qeAuthDataKnownSize                     = 0x02
	certificationDataKnownSize              = 0x06
	quoteV4AuthDataKnownSize                = 0x80
	quoteHeaderStart                        = 0x00
	quoteHeaderEnd                          = 0x30
	quoteBodyStart                          = 0x30
	quoteBodyEnd                            = 0x278
	quoteSignedDataSizeStart                = 0x278
	quoteSignedDataSizeEnd                  = 0x27C
	quoteSignedDataStart                    = quoteSignedDataSizeEnd
	headerVersionStart                      = 0x00
	headerVersionEnd                        = 0x02
	headerAttestationKeyTypeStart           = headerVersionEnd
	headerAttestationKeyTypeEnd             = 0x04
	headerTeeTypeStart                      = headerAttestationKeyTypeEnd
	headerTeeTypeEnd                        = 0x08
	headerPceSvnStart                       = headerTeeTypeEnd
	headerPceSvnEnd                         = 0xA
	headerQeSvnStart                        = headerPceSvnEnd
	headerQeSvnEnd                          = 0xC
	headerQeVendorIDStart                   = headerQeSvnEnd
	headerQeVendorIDEnd                     = 0x1C
	headerUserDataStart                     = headerQeVendorIDEnd
	headerUserDataEnd                       = 0x30
	intelQuoteV4Version                     = 4
	tdTeeTcbSvnStart                        = 0x00
	tdTeeTcbSvnEnd                          = 0x10
	tdMrSeamStart                           = tdTeeTcbSvnEnd
	tdMrSeamEnd                             = 0x40
	tdMrSignerSeamStart                     = tdMrSeamEnd
	tdMrSignerSeamEnd                       = 0x70
	tdSeamAttributesStart                   = tdMrSignerSeamEnd
	tdSeamAttributesEnd                     = 0x78
	tdAttributesStart                       = tdSeamAttributesEnd
	tdAttributesEnd                         = 0x80
	tdXfamStart                             = tdAttributesEnd
	tdXfamEnd                               = 0x88
	tdMrTdStart                             = tdXfamEnd
	tdMrTdEnd                               = 0xB8
	tdMrConfigIDStart                       = tdMrTdEnd
	tdMrConfigIDEnd                         = 0xE8
	tdMrOwnerStart                          = tdMrConfigIDEnd
	tdMrOwnerEnd                            = 0x118
	tdMrOwnerConfigStart                    = tdMrOwnerEnd
	tdMrOwnerConfigEnd                      = 0x148
	tdRtmrsStart                            = tdMrOwnerConfigEnd
	tdRtmrsEnd                              = 0x208
	tdReportDataStart                       = tdRtmrsEnd
	tdReportDataEnd                         = 0x248
	signedDataSignatureStart                = 0x00
	signedDataSignatureEnd                  = 0x40
	signedDataAttestationKeyStart           = signedDataSignatureEnd
	signedDataAttestationKeyEnd             = 0x80
	signedDataCertificationDataStart        = signedDataAttestationKeyEnd
	certificateDataTypeStart                = 0x00
	certificateDataTypeEnd                  = 0x02
	certificateSizeStart                    = certificateDataTypeEnd
	certificateSizeEnd                      = 0x06
	certificateDataStart                    = certificateSizeEnd
	enclaveReportStart                      = 0x00
	enclaveReportEnd                        = 0x180
	qeReportCertificationDataSignatureStart = enclaveReportEnd
	qeReportCertificationDataSignatureEnd   = 0x1C0
	qeReportCertificationDataAuthDataStart  = qeReportCertificationDataSignatureEnd
	qeCPUSvnStart                           = 0x00
	qeCPUSvnEnd                             = 0x10
	qeMiscSelectStart                       = qeCPUSvnEnd
	qeMiscSelectEnd                         = 0x14
	qeReserved1Start                        = qeMiscSelectEnd
	qeReserved1End                          = 0x30
	qeAttributesStart                       = qeReserved1End
	qeAttributesEnd                         = 0x40
	qeMrEnclaveStart                        = qeAttributesEnd
	qeMrEnclaveEnd                          = 0x60
	qeReserved2Start                        = qeMrEnclaveEnd
	qeReserved2End                          = 0x80
	qeMrSignerStart                         = qeReserved2End
	qeMrSignerEnd                           = 0xA0
	qeReserved3Start                        = qeMrSignerEnd
	qeReserved3End                          = 0x100
	qeIsvProdIDStart                        = qeReserved3End
	qeIsvProdIDEnd                          = 0x102
	qeIsvSvnStart                           = qeIsvProdIDEnd
	qeIsvSvnEnd                             = 0x104
	qeReserved4Start                        = qeIsvSvnEnd
	qeReserved4End                          = 0x140
	qeReportDataStart                       = qeReserved4End
	qeReportDataEnd                         = 0x180
	authDataParsedDataSizeStart             = 0x00
	authDataParsedDataSizeEnd               = 0x02
	authDataStart                           = authDataParsedDataSizeEnd
	pckCertChainCertificationDataTypeStart  = 0x00
	pckCertChainCertificationDataTypeEnd    = 0x02
	pckCertChainSizeStart                   = pckCertChainCertificationDataTypeEnd
	pckCertChainSizeEnd                     = 0x06
	pckCertChainDataStart                   = pckCertChainSizeEnd
	rtmrsCount                              = 4
)

var (
	// ErrQuoteNil error returned when Quote is nil
	ErrQuoteNil = errors.New("quote is nil")

	// ErrQuoteV4Nil error returned when QuoteV4 is nil
	ErrQuoteV4Nil = errors.New("QuoteV4 is nil")

	// ErrQuoteV4AuthDataNil error returned when QuoteV4 Auth Data is nil
	ErrQuoteV4AuthDataNil = errors.New("QuoteV4 authData is nil")

	// ErrCertificationDataNil error returned when Certification Data is nil
	ErrCertificationDataNil = errors.New("certification data is nil")

	// ErrQeReportCertificationDataNil error returned when QE report certification data is nil
	ErrQeReportCertificationDataNil = errors.New("QE Report certification data is nil")

	// ErrQeAuthDataNil error returned when QE Auth Data is nil
	ErrQeAuthDataNil = errors.New("QE AuthData is nil")

	// ErrQeReportNil error returned when QE Report is nil
	ErrQeReportNil = errors.New("QE Report is nil")

	// ErrPckCertChainNil error returned when PCK Certificate Chain is nil
	ErrPckCertChainNil = errors.New("PCK certificate chain is nil")

	// ErrTDQuoteBodyNil error returned when TD quote body is nil
	ErrTDQuoteBodyNil = errors.New("TD quote body is nil")

	// ErrTeeType error returned when TEE type is not TDX
	ErrTeeType = errors.New("TEE type is not TDX")

	// ErrAttestationKeyType error returned when attestation key is not of expected type
	ErrAttestationKeyType = errors.New("attestation key type not supported")

	// ErrHeaderNil error returned when header is nil
	ErrHeaderNil = errors.New("header is nil")
)

func clone(b []byte) []byte {
	result := make([]byte, len(b))
	copy(result, b)
	return result
}

// determineQuoteFormat returns the quote format version from the header.
func determineQuoteFormat(b []uint8) (uint32, error) {
	if len(b) < headerVersionEnd {
		return 0, fmt.Errorf("unable to determine quote format since bytes length is less than %d bytes", headerVersionEnd)
	}
	data := clone(b)
	header := &pb.Header{}
	header.Version = uint32(binary.LittleEndian.Uint16(data[headerVersionStart:headerVersionEnd]))
	return header.Version, nil
}

// QuoteToProto creates a Quote from the Intel's attestation quote byte array in Intel's ABI format.
// Supported quote formats - QuoteV4.
func QuoteToProto(b []uint8) (any, error) {
	quoteFormat, err := determineQuoteFormat(b)
	if err != nil {
		return nil, err
	}
	switch quoteFormat {
	case intelQuoteV4Version:
		return quoteToProtoV4(b)
	default:
		return nil, fmt.Errorf("quote format not supported")
	}
}

// quoteToProtoV4 creates a pb.QuoteV4 from the Intel's attestation quote byte array in Intel's ABI format.
func quoteToProtoV4(b []uint8) (*pb.QuoteV4, error) {
	data := clone(b) // Created an independent copy to make the interface less error-prone
	if len(data) < QuoteMinSize {
		return nil, fmt.Errorf("raw quote size is 0x%x, a TDX quote should have size a minimum size of 0x%x", len(data), QuoteMinSize)
	}
	quote := &pb.QuoteV4{}
	header, err := headerToProto(data[quoteHeaderStart:quoteHeaderEnd])
	if err != nil {
		return nil, err
	}

	tdQuoteBody, err := tdQuoteBodyToProto(data[quoteBodyStart:quoteBodyEnd])
	if err != nil {
		return nil, err
	}

	quote.SignedDataSize = binary.LittleEndian.Uint32(data[quoteSignedDataSizeStart:quoteSignedDataSizeEnd])

	additionalData := data[quoteSignedDataStart:]
	if uint32(len(additionalData)) < quote.GetSignedDataSize() {
		return nil, fmt.Errorf("size of signed data is 0x%x. Expected minimum size of 0x%x", len(additionalData), quote.GetSignedDataSize())
	}
	quoteSignedDataEnd := quoteSignedDataStart + quote.GetSignedDataSize()
	rawSignedData := data[quoteSignedDataStart:quoteSignedDataEnd]
	extraBytes := data[quoteSignedDataEnd:]
	signedData, err := signedDataToProto(rawSignedData)
	if err != nil {
		return nil, err
	}

	quote.Header = header
	quote.TdQuoteBody = tdQuoteBody
	quote.SignedData = signedData
	if len(extraBytes) > 0 {
		quote.ExtraBytes = extraBytes
	}

	if err := CheckQuoteV4(quote); err != nil {
		return nil, fmt.Errorf("parsing QuoteV4 failed: %v", err)
	}
	return quote, nil
}

func headerToProto(b []uint8) (*pb.Header, error) {
	data := clone(b) // Created an independent copy to make the interface less error-prone
	header := &pb.Header{}

	header.Version = uint32(binary.LittleEndian.Uint16(data[headerVersionStart:headerVersionEnd]))
	header.AttestationKeyType = uint32(binary.LittleEndian.Uint16(data[headerAttestationKeyTypeStart:headerAttestationKeyTypeEnd]))
	header.TeeType = binary.LittleEndian.Uint32(data[headerTeeTypeStart:headerTeeTypeEnd])
	header.PceSvn = data[headerPceSvnStart:headerPceSvnEnd]
	header.QeSvn = data[headerQeSvnStart:headerQeSvnEnd]
	header.QeVendorId = data[headerQeVendorIDStart:headerQeVendorIDEnd]
	header.UserData = data[headerUserDataStart:headerUserDataEnd]

	if err := checkHeader(header); err != nil {
		return nil, fmt.Errorf("parsing header failed: %v", err)
	}
	return header, nil
}

func tdQuoteBodyToProto(b []uint8) (*pb.TDQuoteBody, error) {
	data := clone(b) // Created an independent copy to make the interface less error-prone
	report := &pb.TDQuoteBody{}
	report.TeeTcbSvn = data[tdTeeTcbSvnStart:tdTeeTcbSvnEnd]
	report.MrSeam = data[tdMrSeamStart:tdMrSeamEnd]
	report.MrSignerSeam = data[tdMrSignerSeamStart:tdMrSignerSeamEnd]
	report.SeamAttributes = data[tdSeamAttributesStart:tdSeamAttributesEnd]
	report.TdAttributes = data[tdAttributesStart:tdAttributesEnd]
	report.Xfam = data[tdXfamStart:tdXfamEnd]
	report.MrTd = data[tdMrTdStart:tdMrTdEnd]
	report.MrConfigId = data[tdMrConfigIDStart:tdMrConfigIDEnd]
	report.MrOwner = data[tdMrOwnerStart:tdMrOwnerEnd]
	report.MrOwnerConfig = data[tdMrOwnerConfigStart:tdMrOwnerConfigEnd]
	report.ReportData = data[tdReportDataStart:tdReportDataEnd]
	rtmrsStart := tdRtmrsStart
	for i := 0; i < rtmrsCount; i++ {
		rtmrsEnd := rtmrsStart + RtmrSize
		arr := data[rtmrsStart:rtmrsEnd]
		report.Rtmrs = append(report.Rtmrs, arr)
		rtmrsStart += RtmrSize
	}

	if err := checkTDQuoteBody(report); err != nil {
		return nil, fmt.Errorf("parsing TD Quote Body failed: %v", err)
	}

	return report, nil
}

func signedDataToProto(b []uint8) (*pb.Ecdsa256BitQuoteV4AuthData, error) {
	data := clone(b) // Created an independent copy to make the interface less error-prone
	signedData := &pb.Ecdsa256BitQuoteV4AuthData{}
	signedData.Signature = data[signedDataSignatureStart:signedDataSignatureEnd]
	signedData.EcdsaAttestationKey = data[signedDataAttestationKeyStart:signedDataAttestationKeyEnd]

	certificationData, err := certificationDataToProto(data[signedDataCertificationDataStart:])
	if err != nil {
		return nil, err
	}

	signedData.CertificationData = certificationData

	if err := checkEcdsa256BitQuoteV4AuthData(signedData); err != nil {
		return nil, fmt.Errorf("parsing QuoteV4 AuthData failed: %v", err)
	}
	return signedData, nil
}

func certificationDataToProto(b []uint8) (*pb.CertificationData, error) {
	data := clone(b) // Created an independent copy to make the interface less error-prone
	certification := &pb.CertificationData{}

	certification.CertificateDataType = uint32(binary.LittleEndian.Uint16(data[certificateDataTypeStart:certificateDataTypeEnd]))
	certification.Size = binary.LittleEndian.Uint32(data[certificateSizeStart:certificateSizeEnd])
	rawCertificateData := data[certificateDataStart:]
	if uint32(len(rawCertificateData)) != certification.GetSize() {
		return nil, fmt.Errorf("size of certificate data is 0x%x. Expected size 0x%x", len(rawCertificateData), certification.GetSize())
	}

	qeReportCertificationData, err := qeReportCertificationDataToProto(rawCertificateData)
	if err != nil {
		return nil, err
	}

	certification.QeReportCertificationData = qeReportCertificationData

	if err := checkCertificationData(certification); err != nil {
		return nil, fmt.Errorf("parsing certification data failed: %v", err)
	}
	return certification, nil
}

func qeReportCertificationDataToProto(b []uint8) (*pb.QEReportCertificationData, error) {
	data := clone(b) // Created an independent copy to make the interface less error-prone
	qeReportCertificationData := &pb.QEReportCertificationData{}

	enclaveReport, err := enclaveReportToProto(data[enclaveReportStart:enclaveReportEnd])
	if err != nil {
		return nil, err
	}

	qeReportCertificationData.QeReport = enclaveReport
	qeReportCertificationData.QeReportSignature = data[qeReportCertificationDataSignatureStart:qeReportCertificationDataSignatureEnd]

	authData, authDataSize, err := qeAuthDataToProto(data[qeReportCertificationDataAuthDataStart:])
	if err != nil {
		return nil, err
	}

	qeReportCertificationData.QeAuthData = authData

	pckCertificateStart := qeReportCertificationDataAuthDataStart + authDataSize

	pckCertificateChain, err := pckCertificateChainToProto(data[pckCertificateStart:])
	if err != nil {
		return nil, err
	}

	qeReportCertificationData.PckCertificateChainData = pckCertificateChain

	if err := checkQeReportCertificationData(qeReportCertificationData); err != nil {
		return nil, fmt.Errorf("parsing QE Report Certification Data failed: %v", err)
	}
	return qeReportCertificationData, nil
}

func enclaveReportToProto(b []uint8) (*pb.EnclaveReport, error) {
	data := clone(b) // Created an independent copy to make the interface less error-prone
	enclaveReport := &pb.EnclaveReport{}

	enclaveReport.CpuSvn = data[qeCPUSvnStart:qeCPUSvnEnd]
	enclaveReport.MiscSelect = binary.LittleEndian.Uint32(data[qeMiscSelectStart:qeMiscSelectEnd])
	enclaveReport.Reserved1 = data[qeReserved1Start:qeReserved1End]
	enclaveReport.Attributes = data[qeAttributesStart:qeAttributesEnd]
	enclaveReport.MrEnclave = data[qeMrEnclaveStart:qeMrEnclaveEnd]
	enclaveReport.Reserved2 = data[qeReserved2Start:qeReserved2End]
	enclaveReport.MrSigner = data[qeMrSignerStart:qeMrSignerEnd]
	enclaveReport.Reserved3 = data[qeReserved3Start:qeReserved3End]
	enclaveReport.IsvProdId = uint32(binary.LittleEndian.Uint16(data[qeIsvProdIDStart:qeIsvProdIDEnd]))
	enclaveReport.IsvSvn = uint32(binary.LittleEndian.Uint16(data[qeIsvSvnStart:qeIsvSvnEnd]))
	enclaveReport.Reserved4 = data[qeReserved4Start:qeReserved4End]
	enclaveReport.ReportData = data[qeReportDataStart:qeReportDataEnd]

	if err := checkQeReport(enclaveReport); err != nil {
		return nil, fmt.Errorf("parsing QE Report failed: %v", err)
	}
	return enclaveReport, nil
}

func qeAuthDataToProto(b []uint8) (*pb.QeAuthData, uint32, error) {
	data := clone(b) // Created an independent copy to make the interface less error-prone
	authData := &pb.QeAuthData{}

	authData.ParsedDataSize = uint32(binary.LittleEndian.Uint16(data[authDataParsedDataSizeStart:authDataParsedDataSizeEnd]))
	authDataEnd := authDataParsedDataSizeEnd + authData.GetParsedDataSize()
	authData.Data = data[authDataStart:authDataEnd]
	if err := checkQeAuthData(authData); err != nil {
		return nil, 0, fmt.Errorf("parsing QE AuthData failed: %v", err)
	}
	return authData, authDataEnd, nil
}

func pckCertificateChainToProto(b []uint8) (*pb.PCKCertificateChainData, error) {
	data := clone(b) // Created an independent copy to make the interface less error-prone
	pckCertificateChain := &pb.PCKCertificateChainData{}

	pckCertificateChain.CertificateDataType = uint32(binary.LittleEndian.Uint16(data[pckCertChainCertificationDataTypeStart:pckCertChainCertificationDataTypeEnd]))
	pckCertificateChain.Size = binary.LittleEndian.Uint32(data[pckCertChainSizeStart:pckCertChainSizeEnd])
	pckCertificateChain.PckCertChain = data[pckCertChainDataStart:]

	if err := checkPCKCertificateChain(pckCertificateChain); err != nil {
		return nil, fmt.Errorf("parsing PCK certification chain failed: %v", err)
	}
	return pckCertificateChain, nil
}

func checkHeader(header *pb.Header) error {
	if header == nil {
		return ErrHeaderNil
	}
	if header.GetVersion() >= (1 << 16) {
		return fmt.Errorf("version field size must fit in 2 bytes , got %d", header.GetVersion())
	}
	if header.GetVersion() != QuoteVersion {
		return fmt.Errorf("version %d not supported", header.GetVersion())
	}
	if header.GetAttestationKeyType() >= (1 << 16) {
		return fmt.Errorf("attestation key type field size must fit in 2 bytes , got %d", header.GetAttestationKeyType())
	}
	if header.GetAttestationKeyType() != AttestationKeyType {
		return ErrAttestationKeyType
	}
	if header.GetTeeType() != TeeTDX {
		return ErrTeeType
	}

	if len(header.GetQeSvn()) != qeSvnSize {
		return fmt.Errorf("qeSvn size is %d bytes. Expected %d bytes", len(header.GetQeSvn()), qeSvnSize)
	}
	if len(header.GetPceSvn()) != pceSvnSize {
		return fmt.Errorf("pceSvn size is %d bytes. Expected %d bytes", len(header.GetPceSvn()), pceSvnSize)
	}
	if len(header.GetQeVendorId()) != QeVendorIDSize {
		return fmt.Errorf("qeVendorId size is %d bytes. Expected %d bytes", len(header.GetQeVendorId()), QeVendorIDSize)
	}
	if len(header.GetUserData()) != userDataSize {
		return fmt.Errorf("user data size is %d bytes. Expected %d bytes", len(header.GetUserData()), userDataSize)
	}

	return nil

}

func checkTDQuoteBody(tdQuoteBody *pb.TDQuoteBody) error {
	if tdQuoteBody == nil {
		return ErrTDQuoteBodyNil
	}
	if len(tdQuoteBody.GetTeeTcbSvn()) != TeeTcbSvnSize {
		return fmt.Errorf("teeTcbSvn size is %d bytes. Expected %d bytes", len(tdQuoteBody.GetTeeTcbSvn()), TeeTcbSvnSize)
	}
	if len(tdQuoteBody.GetMrSeam()) != MrSeamSize {
		return fmt.Errorf("mrSeam size is %d bytes. Expected %d bytes", len(tdQuoteBody.GetMrSeam()), MrSeamSize)
	}
	if len(tdQuoteBody.GetMrSignerSeam()) != mrSignerSeamSize {
		return fmt.Errorf("mrSignerSeam size is %d bytes. Expected %d bytes", len(tdQuoteBody.GetMrSignerSeam()), mrSignerSeamSize)
	}
	if len(tdQuoteBody.GetSeamAttributes()) != seamAttributesSize {
		return fmt.Errorf("seamAttributes size is %d bytes. Expected %d bytes", len(tdQuoteBody.GetSeamAttributes()), seamAttributesSize)
	}
	if len(tdQuoteBody.GetTdAttributes()) != TdAttributesSize {
		return fmt.Errorf("tdAttributes size is %d bytes. Expected %d bytes", len(tdQuoteBody.GetTdAttributes()), TdAttributesSize)
	}
	if len(tdQuoteBody.GetXfam()) != XfamSize {
		return fmt.Errorf("xfam size is %d bytes. Expected %d bytes", len(tdQuoteBody.GetXfam()), XfamSize)
	}
	if len(tdQuoteBody.GetMrTd()) != MrTdSize {
		return fmt.Errorf("mrTd size is %d bytes. Expected %d bytes", len(tdQuoteBody.GetMrTd()), MrTdSize)
	}
	if len(tdQuoteBody.GetMrConfigId()) != MrConfigIDSize {
		return fmt.Errorf("mrConfigId size is %d bytes. Expected %d bytes", len(tdQuoteBody.GetMrConfigId()), MrConfigIDSize)
	}
	if len(tdQuoteBody.GetMrOwner()) != MrOwnerSize {
		return fmt.Errorf("mrOwner size is %d bytes. Expected %d bytes", len(tdQuoteBody.GetMrOwner()), MrOwnerSize)
	}
	if len(tdQuoteBody.GetMrOwnerConfig()) != MrOwnerConfigSize {
		return fmt.Errorf("mrOwnerConfig size is %d bytes. Expected %d bytes", len(tdQuoteBody.GetMrOwnerConfig()), MrOwnerConfigSize)
	}
	if len(tdQuoteBody.GetRtmrs()) != rtmrsCount {
		return fmt.Errorf("rtmrs count is %d. Expected %d", len(tdQuoteBody.GetRtmrs()), rtmrsCount)
	}
	for i := 0; i < rtmrsCount; i++ {
		if len(tdQuoteBody.GetRtmrs()[i]) != RtmrSize {
			return fmt.Errorf("rtmr%d size is %d bytes. Expected %d bytes", i, len(tdQuoteBody.GetRtmrs()[i]), RtmrSize)
		}
	}
	return nil
}

func checkPCKCertificateChain(chain *pb.PCKCertificateChainData) error {
	if chain == nil {
		return ErrPckCertChainNil
	}
	if chain.GetCertificateDataType() >= (1 << 16) {
		return fmt.Errorf("certification data type expected to be of 2 bytes, got %d", chain.GetCertificateDataType())
	}
	if chain.GetCertificateDataType() != pckReportCertificationDataType {
		return fmt.Errorf("PCK certificate chain data type invalid, got %d, expected %d", chain.GetCertificateDataType(), pckReportCertificationDataType)
	}
	if chain.GetSize() != uint32(len(chain.GetPckCertChain())) {
		return fmt.Errorf("PCK certificate chain size is %d. Expected size %d", len(chain.GetPckCertChain()), chain.GetSize())
	}
	return nil
}

func checkQeReport(report *pb.EnclaveReport) error {
	if report == nil {
		return ErrQeReportNil
	}
	if len(report.GetCpuSvn()) != cpuSvnSize {
		return fmt.Errorf("cpuSvn size is %d bytes. Expected %d bytes", len(report.GetCpuSvn()), cpuSvnSize)
	}
	if len(report.GetReserved1()) != reserved1Size {
		return fmt.Errorf("reserved1 size is %d bytes. Expected %d bytes", len(report.GetReserved1()), reserved1Size)
	}
	if len(report.GetAttributes()) != attributesSize {
		return fmt.Errorf("attributes size is %d bytes. Expected %d bytes", len(report.GetAttributes()), attributesSize)
	}
	if len(report.GetMrEnclave()) != mrEnclaveSize {
		return fmt.Errorf("mrEnclave size is %d bytes. Expected %d bytes", len(report.GetMrEnclave()), mrEnclaveSize)
	}
	if len(report.GetReserved2()) != reserved2Size {
		return fmt.Errorf("reserved2 size is %d bytes. Expected %d bytes", len(report.GetReserved2()), reserved2Size)
	}
	if len(report.GetMrSigner()) != mrSignerSize {
		return fmt.Errorf("mrSigner size is %d bytes. Expected %d bytes", len(report.GetMrSigner()), mrSignerSize)
	}
	if len(report.GetReserved3()) != reserved3Size {
		return fmt.Errorf("reserved3 size is %d bytes. Expected %d bytes", len(report.GetReserved3()), reserved3Size)
	}
	if report.GetIsvProdId() >= (1 << 16) {
		return fmt.Errorf("isVProdId field size must fit in 2 bytes , got %d", report.GetIsvProdId())
	}
	if report.GetIsvSvn() >= (1 << 16) {
		return fmt.Errorf("isVSvn field size must fit in 2 bytes , got %d", report.GetIsvSvn())
	}
	if len(report.GetReserved4()) != reserved4Size {
		return fmt.Errorf("reserved4 size is %d bytes. Expected %d bytes", len(report.GetReserved4()), reserved4Size)
	}
	if len(report.GetReportData()) != ReportDataSize {
		return fmt.Errorf("report data size is %d bytes. Expected %d bytes", len(report.GetReportData()), ReportDataSize)
	}
	return nil
}

func checkQeAuthData(authData *pb.QeAuthData) error {
	if authData == nil {
		return ErrQeAuthDataNil
	}
	if authData.GetParsedDataSize() >= (1 << 16) {
		return fmt.Errorf("parsed data size field size must fit in 2 bytes , got %d", authData.GetParsedDataSize())
	}
	if authData.GetParsedDataSize() != uint32(len(authData.GetData())) {
		return fmt.Errorf("parsed data size is %d bytes. Expected %d bytes", len(authData.GetData()), authData.GetParsedDataSize())
	}
	return nil
}

func checkQeReportCertificationData(qeReport *pb.QEReportCertificationData) error {
	if qeReport == nil {
		return ErrQeReportCertificationDataNil
	}
	if err := checkQeReport(qeReport.GetQeReport()); err != nil {
		return fmt.Errorf("QE Report error: %v", err)
	}
	if len(qeReport.GetQeReportSignature()) != signatureSize {
		return fmt.Errorf("signature size is %d bytes. Expected %d bytes", len(qeReport.GetQeReportSignature()), signatureSize)
	}
	if err := checkQeAuthData(qeReport.GetQeAuthData()); err != nil {
		return fmt.Errorf("QE AuthData error: %v", err)
	}
	if err := checkPCKCertificateChain(qeReport.GetPckCertificateChainData()); err != nil {
		return fmt.Errorf("PCK certificate chain error: %v", err)
	}
	return nil
}

func checkCertificationData(certification *pb.CertificationData) error {
	if certification == nil {
		return ErrCertificationDataNil
	}
	if certification.GetCertificateDataType() >= (1 << 16) {
		return fmt.Errorf("certification data type field size must fit in 2 bytes , got %d", certification.GetCertificateDataType())
	}
	if certification.GetCertificateDataType() != qeReportCertificationDataType {
		return fmt.Errorf("certification data type invalid, got %d, expected %d", certification.GetCertificateDataType(), qeReportCertificationDataType)
	}
	if err := checkQeReportCertificationData(certification.GetQeReportCertificationData()); err != nil {
		return fmt.Errorf("QE Report certification data error: %v", err)
	}
	return nil
}

func checkEcdsa256BitQuoteV4AuthData(signedData *pb.Ecdsa256BitQuoteV4AuthData) error {
	if signedData == nil {
		return ErrQuoteV4AuthDataNil
	}
	if len(signedData.GetSignature()) != signatureSize {
		return fmt.Errorf("signature size is %d bytes. Expected %d bytes", len(signedData.GetSignature()), signatureSize)
	}
	if len(signedData.GetEcdsaAttestationKey()) != attestationKeySize {
		return fmt.Errorf("ecdsa attestation key size is %d bytes. Expected %d bytes", len(signedData.GetEcdsaAttestationKey()), attestationKeySize)
	}
	if err := checkCertificationData(signedData.GetCertificationData()); err != nil {
		return fmt.Errorf("certification data error: %v", err)
	}

	return nil
}

// CheckQuoteV4  validates a quote protobuf by ensuring all parameters meet their required size
func CheckQuoteV4(quote *pb.QuoteV4) error {
	if quote == nil {
		return ErrQuoteV4Nil
	}
	if err := checkHeader(quote.GetHeader()); err != nil {
		return fmt.Errorf("QuoteV4 Header error: %v", err)
	}
	if err := checkTDQuoteBody(quote.GetTdQuoteBody()); err != nil {
		return fmt.Errorf("QuoteV4 TD Quote Body error: %v", err)
	}

	if err := checkEcdsa256BitQuoteV4AuthData(quote.GetSignedData()); err != nil {
		return fmt.Errorf("QuoteV4 AuthData error: %v", err)
	}
	return nil
}

// EnclaveReportToAbiBytes translates the EnclaveReport back into its little-endian ABI format
func EnclaveReportToAbiBytes(report *pb.EnclaveReport) ([]byte, error) {
	if report == nil {
		return nil, ErrQeReportNil
	}
	if err := checkQeReport(report); err != nil {
		return nil, fmt.Errorf("QE Report invalid: %v", err)
	}

	data := make([]byte, qeReportSize)
	copy(data[qeCPUSvnStart:qeCPUSvnEnd], report.GetCpuSvn())
	binary.LittleEndian.PutUint32(data[qeMiscSelectStart:qeMiscSelectEnd], report.GetMiscSelect())
	copy(data[qeReserved1Start:qeReserved1End], report.GetReserved1())
	copy(data[qeAttributesStart:qeAttributesEnd], report.GetAttributes())
	copy(data[qeMrEnclaveStart:qeMrEnclaveEnd], report.GetMrEnclave())
	copy(data[qeReserved2Start:qeReserved2End], report.GetReserved2())
	copy(data[qeMrSignerStart:qeMrSignerEnd], report.GetMrSigner())
	copy(data[qeReserved3Start:qeReserved3End], report.GetReserved3())
	binary.LittleEndian.PutUint16(data[qeIsvProdIDStart:qeIsvProdIDEnd], uint16(report.GetIsvProdId()))
	binary.LittleEndian.PutUint16(data[qeIsvSvnStart:qeIsvSvnEnd], uint16(report.GetIsvSvn()))
	copy(data[qeReserved4Start:qeReserved4End], report.GetReserved4())
	copy(data[qeReportDataStart:qeReportDataEnd], report.GetReportData())
	return data, nil
}

// HeaderToAbiBytes translates the Header back into its little-endian ABI format
func HeaderToAbiBytes(header *pb.Header) ([]byte, error) {
	if header == nil {
		return nil, ErrHeaderNil
	}
	if err := checkHeader(header); err != nil {
		return nil, fmt.Errorf("header invalid: %v", err)
	}

	data := make([]byte, headerSize)
	binary.LittleEndian.PutUint16(data[headerVersionStart:headerVersionEnd], uint16(header.GetVersion()))
	binary.LittleEndian.PutUint16(data[headerAttestationKeyTypeStart:headerAttestationKeyTypeEnd], uint16(header.GetAttestationKeyType()))
	binary.LittleEndian.PutUint32(data[headerTeeTypeStart:headerTeeTypeEnd], (header.GetTeeType()))
	copy(data[headerPceSvnStart:headerPceSvnEnd], header.GetPceSvn())
	copy(data[headerQeSvnStart:headerQeSvnEnd], header.GetQeSvn())
	copy(data[headerQeVendorIDStart:headerQeVendorIDEnd], header.GetQeVendorId())
	copy(data[headerUserDataStart:headerUserDataEnd], header.GetUserData())

	return data, nil

}

// TdQuoteBodyToAbiBytes translates the TDQuoteBody back into its little-endian ABI format
func TdQuoteBodyToAbiBytes(tdQuoteBody *pb.TDQuoteBody) ([]byte, error) {

	if tdQuoteBody == nil {
		return nil, ErrTDQuoteBodyNil
	}
	if err := checkTDQuoteBody(tdQuoteBody); err != nil {
		return nil, fmt.Errorf("TD quote body invalid: %v", err)
	}

	data := make([]byte, tdQuoteBodySize)
	copy(data[tdTeeTcbSvnStart:tdTeeTcbSvnEnd], tdQuoteBody.GetTeeTcbSvn())
	copy(data[tdMrSeamStart:tdMrSeamEnd], tdQuoteBody.GetMrSeam())
	copy(data[tdMrSignerSeamStart:tdMrSignerSeamEnd], tdQuoteBody.GetMrSignerSeam())
	copy(data[tdSeamAttributesStart:tdSeamAttributesEnd], tdQuoteBody.GetSeamAttributes())
	copy(data[tdAttributesStart:tdAttributesEnd], tdQuoteBody.GetTdAttributes())
	copy(data[tdXfamStart:tdXfamEnd], tdQuoteBody.GetXfam())
	copy(data[tdMrTdStart:tdMrTdEnd], tdQuoteBody.GetMrTd())
	copy(data[tdMrConfigIDStart:tdMrConfigIDEnd], tdQuoteBody.GetMrConfigId())
	copy(data[tdMrOwnerStart:tdMrOwnerEnd], tdQuoteBody.GetMrOwner())
	copy(data[tdMrOwnerConfigStart:tdMrOwnerConfigEnd], tdQuoteBody.GetMrOwnerConfig())
	rtmrsStart := tdRtmrsStart
	for i := 0; i < rtmrsCount; i++ {
		rtmrsEnd := rtmrsStart + RtmrSize
		copy(data[rtmrsStart:rtmrsEnd], tdQuoteBody.GetRtmrs()[i])
		rtmrsStart += RtmrSize
	}
	copy(data[tdReportDataStart:tdReportDataEnd], tdQuoteBody.GetReportData())
	return data, nil
}

func pckCertificateChainToAbiBytes(pckCertificateChain *pb.PCKCertificateChainData) ([]byte, error) {
	if pckCertificateChain == nil {
		return nil, ErrPckCertChainNil
	}
	if err := checkPCKCertificateChain(pckCertificateChain); err != nil {
		return nil, fmt.Errorf("PCK certificate chain data invalid: %v", err)
	}

	data := make([]byte, pckCertificateChainKnownSize+pckCertificateChain.GetSize())
	binary.LittleEndian.PutUint16(data[pckCertChainCertificationDataTypeStart:pckCertChainCertificationDataTypeEnd], uint16(pckCertificateChain.GetCertificateDataType()))
	binary.LittleEndian.PutUint32(data[pckCertChainSizeStart:pckCertChainSizeEnd], pckCertificateChain.GetSize())
	copy(data[pckCertChainDataStart:], pckCertificateChain.GetPckCertChain())
	return data, nil
}

func qeAuthDataToAbiBytes(authData *pb.QeAuthData) ([]byte, error) {
	if authData == nil {
		return nil, ErrQeAuthDataNil
	}
	if err := checkQeAuthData(authData); err != nil {
		return nil, fmt.Errorf("QE AuthData invalid: %v", err)
	}

	data := make([]byte, qeAuthDataKnownSize+authData.GetParsedDataSize())
	binary.LittleEndian.PutUint16(data[authDataParsedDataSizeStart:authDataParsedDataSizeEnd], uint16(authData.GetParsedDataSize()))
	copy(data[authDataStart:], authData.GetData())
	return data, nil
}

func qeReportCertificationDataToAbiBytes(qeReport *pb.QEReportCertificationData) ([]byte, error) {
	if qeReport == nil {
		return nil, ErrQeReportCertificationDataNil
	}
	if err := checkQeReportCertificationData(qeReport); err != nil {
		return nil, fmt.Errorf("QE Report certification data invalid: %v", err)
	}

	data, err := EnclaveReportToAbiBytes(qeReport.GetQeReport())
	if err != nil {
		return nil, fmt.Errorf("enclave report to abi bytes conversion failed: %v", err)
	}
	qeReportSignatureData := clone(qeReport.GetQeReportSignature())
	data = append(data, qeReportSignatureData...)
	qeAuthData, err := qeAuthDataToAbiBytes(qeReport.GetQeAuthData())
	if err != nil {
		return nil, fmt.Errorf("QE AuthData to abi bytes conversion failed: %v", err)
	}
	data = append(data, qeAuthData...)

	pckCertificateChainData, err := pckCertificateChainToAbiBytes(qeReport.GetPckCertificateChainData())
	if err != nil {
		return nil, fmt.Errorf("PCK certificate chain to abi bytes conversion failed: %v", err)
	}
	data = append(data, pckCertificateChainData...)
	return data, nil
}

func certificationDataToAbiBytes(certification *pb.CertificationData) ([]byte, error) {
	if certification == nil {
		return nil, ErrCertificationDataNil
	}
	if err := checkCertificationData(certification); err != nil {
		return nil, fmt.Errorf("certification data invalid: %v", err)
	}
	data := make([]byte, certificationDataKnownSize)
	binary.LittleEndian.PutUint16(data[certificateDataTypeStart:certificateDataTypeEnd], uint16(certification.GetCertificateDataType()))
	binary.LittleEndian.PutUint32(data[certificateSizeStart:certificateSizeEnd], certification.GetSize())

	certificationData, err := qeReportCertificationDataToAbiBytes(certification.GetQeReportCertificationData())
	if err != nil {
		return nil, fmt.Errorf("QE Report certification data to abi bytes conversion failed: %v", err)
	}
	data = append(data, certificationData...)
	return data, nil
}

func signedDataToAbiBytes(signedData *pb.Ecdsa256BitQuoteV4AuthData) ([]byte, error) {
	if signedData == nil {
		return nil, ErrQuoteV4AuthDataNil
	}
	if err := checkEcdsa256BitQuoteV4AuthData(signedData); err != nil {
		return nil, fmt.Errorf("QuoteV4 AuthData invalid: %v", err)
	}
	data := make([]byte, quoteV4AuthDataKnownSize)
	copy(data[signedDataSignatureStart:signedDataSignatureEnd], signedData.GetSignature())
	copy(data[signedDataAttestationKeyStart:signedDataAttestationKeyEnd], signedData.GetEcdsaAttestationKey())

	certificationData, err := certificationDataToAbiBytes(signedData.GetCertificationData())
	if err != nil {
		return nil, fmt.Errorf("signed data certification data to abi bytes conversion failed: %v", err)
	}
	data = append(data, certificationData...)
	return data, nil
}

// QuoteToAbiBytes translates the Quote back into its little-endian ABI format.
// Supported quote formats - QuoteV4.
func QuoteToAbiBytes(quote any) ([]byte, error) {
	if quote == nil {
		return nil, ErrQuoteNil
	}
	switch q := quote.(type) {
	case *pb.QuoteV4:
		return quoteToAbiBytesV4(q)
	default:
		return nil, fmt.Errorf("unsupported quote type: %T", quote)
	}
}

// quoteToAbiBytesV4 translates the QuoteV4 back into its little-endian ABI format
func quoteToAbiBytesV4(quote *pb.QuoteV4) ([]byte, error) {
	if err := CheckQuoteV4(quote); err != nil {
		return nil, fmt.Errorf("QuoteV4 invalid: %v", err)
	}
	var data []byte

	headerData, err := HeaderToAbiBytes(quote.GetHeader())
	if err != nil {
		return nil, fmt.Errorf("header to abi bytes conversion failed: %v", err)
	}
	data = append(data, headerData...)

	tdReportData, err := TdQuoteBodyToAbiBytes(quote.GetTdQuoteBody())
	if err != nil {
		return nil, fmt.Errorf("TD quote body to abi bytes conversion failed: %v", err)
	}
	data = append(data, tdReportData...)

	signedDataSizeBytes := make([]byte, 0x04)
	binary.LittleEndian.PutUint32(signedDataSizeBytes[0x00:0x04], quote.GetSignedDataSize())
	data = append(data, signedDataSizeBytes...)

	signedData, err := signedDataToAbiBytes(quote.GetSignedData())
	if err != nil {
		return nil, fmt.Errorf("signed data to abi bytes conversion failed: %v", err)
	}
	data = append(data, signedData...)

	if quote.GetExtraBytes() != nil {
		data = append(data, quote.GetExtraBytes()...)
	}
	return data, nil
}

// SignatureToDER converts the signature to DER format
func SignatureToDER(x []byte) ([]byte, error) {
	if len(x) != signatureSize {
		return nil, fmt.Errorf("signature size is %d bytes. Expected %d bytes", len(x), signatureSize)
	}
	var b cryptobyte.Builder
	b.AddASN1(asn1.SEQUENCE, func(b *cryptobyte.Builder) {
		b.AddASN1BigInt(new(big.Int).SetBytes(ecdsaGetR(x)))
		b.AddASN1BigInt(new(big.Int).SetBytes(ecdsaGetS(x)))
	})
	return b.Bytes()
}
func ecdsaGetR(signature []byte) []byte {
	return signature[0x0:0x20]
}
func ecdsaGetS(signature []byte) []byte {
	return signature[0x20:0x40]
}
