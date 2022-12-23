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

// Package linuxabi describes the /dev/sev-guest ioctl command ABI.
package linuxabi

import (
	"errors"
	"fmt"
	"reflect"
	"unsafe"
)

// EsResult is the status code type for Linux's GHCB communication results.
type EsResult int

// ioctl bits for x86-64
const (
	iocNrbits    = 8
	iocTypebits  = 8
	iocSizebits  = 14
	iocDirbits   = 2
	iocNrshift   = 0
	iocTypeshift = (iocNrshift + iocNrbits)
	iocSizeshift = (iocTypeshift + iocTypebits)
	iocDirshift  = (iocSizeshift + iocSizebits)
	iocWrite     = 1
	iocRead      = 2

	// Linux /dev/sev-guest ioctl interface
	iocTypeSnpGuestReq = 'S'
	iocSnpWithoutNr    = ((iocWrite | iocRead) << iocDirshift) |
		(iocTypeSnpGuestReq << iocTypeshift) |
		// unsafe.Sizeof(snpUserGuestRequest)
		(32 << iocSizeshift)

	// IocSnpGetReport is the ioctl command for getting an attestation report
	IocSnpGetReport = iocSnpWithoutNr | (0x0 << iocNrshift)

	// IocSnpGetDerivedKey is the ioctl command for getting a key derived from measured components and
	// either the VCEK or VMRK.
	IocSnpGetDerivedKey = iocSnpWithoutNr | (0x1 << iocNrshift)

	// IocSnpGetReport is the ioctl command for getting an extended attestation report that includes
	// certificate information.
	IocSnpGetExtendedReport = iocSnpWithoutNr | (0x2 << iocNrshift)

	// The message version for MSG_REPORT_REQ in the SNP API. Specified as 1.
	guestMsgVersion = 1

	// These numbers are from the uapi header sev_guest.h
	snpResportRespSize      = 4000
	msgReportReqHeaderSize  = 0x20
	SnpReportRespReportSize = snpResportRespSize - msgReportReqHeaderSize
)

const (
	// EsOk denotes success.
	EsOk EsResult = iota
	// EsUnsupported denotes that the requested operation is not supported.
	EsUnsupported
	// EsVmmError denotes that the virtual machine monitor was in an unexpected state.
	EsVmmError
	// EsDecodeFailed denotes that instruction decoding failed.
	EsDecodeFailed
	// EsException denotes that the GHCB communication caused an exception.
	EsException
	// EsRetry is the code for a retry instruction emulation
	EsRetry
)

// SevEsErr is an error that interprets SEV-ES guest-host communication results.
type SevEsErr struct {
	error
	Result EsResult
}

func (err SevEsErr) Error() string {
	if err.Result == EsUnsupported {
		return "requested operation not supported"
	}
	if err.Result == EsVmmError {
		return "unexpected state from the VMM"
	}
	if err.Result == EsDecodeFailed {
		return "instruction decoding failed"
	}
	if err.Result == EsException {
		return "instruction caused exception"
	}
	if err.Result == EsRetry {
		return "retry instruction emulation"
	}
	return "unknown error"
}

// SnpReportReqABI is Linux's sev-guest ioctl abi for sending a GET_REPORT request. See
// include/uapi/linux/sev-guest.h
type SnpReportReqABI struct {
	// ReportData to be included in the report
	ReportData [64]uint8

	// Vmpl is the SEV-SNP VMPL level to be included in the report.
	// The kernel must have access to the corresponding VMPCK.
	Vmpl uint32

	reserved [28]byte
}

// SnpReportRespABI is Linux's sev-guest ioctl abi for receiving a GET_REPORT response.
// The size is expected to be snpReportRespSize.
type SnpReportRespABI struct {
	Status     uint32
	ReportSize uint32
	reserved   [0x20 - 8]byte
	// Data is the response data, see SEV-SNP spec for the format
	Data [SnpReportRespReportSize]uint8
}

// ABI returns the same object since it doesn't need a separate representation across the interface.
func (r *SnpReportReqABI) ABI() BinaryConversion { return r }

// Pointer returns a pointer to the object itself.
func (r *SnpReportReqABI) Pointer() unsafe.Pointer {
	return unsafe.Pointer(r)
}

// Finish is a no-op.
func (r *SnpReportReqABI) Finish(b BinaryConvertible) error { return nil }

// ABI returns the same object since it doesn't need a separate representation across the interface.
func (r *SnpReportRespABI) ABI() BinaryConversion { return r }

// Pointer returns a pointer to the object itself.
func (r *SnpReportRespABI) Pointer() unsafe.Pointer {
	return unsafe.Pointer(r)
}

// Finish checks the status of the message and translates it to a Golang error.
func (r *SnpReportRespABI) Finish(b BinaryConvertible) error {
	if r.Status != 0 {
		switch r.Status {
		case 0x16: // Value from MSG_REPORT_RSP specification in SNP API.
			return errors.New("get_report had invalid parameters")
		default:
			return fmt.Errorf("unknown status: 0x%x", r.Status)
		}
	}
	return nil
}

// SnpDerivedKeyReqABI is the ABI representation of a request to the SEV guest device to derive a
// key from specified information.
type SnpDerivedKeyReqABI struct {
	// RootKeySelect is all reserved bits except bit 0 for UseVMRK (1) or UseVCEK (0).
	RootKeySelect    uint32
	reserved         uint32
	GuestFieldSelect uint64
	// Vmpl to mix into the key. Must be greater than or equal to current Vmpl.
	Vmpl uint32
	// GuestSVN to mix into the key. Must be less than or equal to GuestSVN at launch.
	GuestSVN uint32
	// TCBVersion to mix into the key. Must be less than or equal to the CommittedTcb.
	TCBVersion uint64
}

// Pointer returns a pointer to the object.
func (r *SnpDerivedKeyReqABI) Pointer() unsafe.Pointer { return unsafe.Pointer(r) }

// Finish is a no-op.
func (r *SnpDerivedKeyReqABI) Finish(BinaryConvertible) error { return nil }

// ABI returns the ABI representation of this object.
func (r *SnpDerivedKeyReqABI) ABI() BinaryConversion { return r }

// SnpDerivedKeyRespABI represents the response to an SnpDerivedKeyReq.
type SnpDerivedKeyRespABI struct {
	Status   uint32
	reserved [0x20 - 4]byte
	Data     [32]byte
}

// ABI returns the object itself.
func (r *SnpDerivedKeyRespABI) ABI() BinaryConversion { return r }

// Pointer returns a pointer to the object itself.
func (r *SnpDerivedKeyRespABI) Pointer() unsafe.Pointer { return unsafe.Pointer(r) }

// Finish is a no-op.
func (r *SnpDerivedKeyRespABI) Finish(BinaryConvertible) error {
	switch r.Status {
	case 0:
		return nil
	case 0x16:
		return errors.New("msg_key_req error: invalid parameters")
	default:
		return fmt.Errorf("msg_key_req unknown status code: 0x%x", r.Status)
	}
}

// SnpExtendedReportReqABI is Linux's sev-guest ioctl abi for sending a GET_EXTENDED_REPORT request.
type SnpExtendedReportReqABI struct {
	Data SnpReportReqABI

	// Where to copy the certificate blob.
	CertsAddress unsafe.Pointer

	// length of the certificate blob
	CertsLength uint32
}

// SnpExtendedReportReq is close to Linux's sev-guest ioctl abi for sending a GET_EXTENDED_REPORT request,
// but uses safer types for the Ioctl interface.
type SnpExtendedReportReq struct {
	Data SnpReportReqABI

	// Certs receives the certificate blob after the extended report request.
	Certs []byte

	// CertsLength is the length of the certificate blob.
	CertsLength uint32
}

// Pointer returns a pointer so the object itself.
func (r *SnpExtendedReportReqABI) Pointer() unsafe.Pointer {
	return unsafe.Pointer(r)
}

// Finish writes back the changed CertsLength value.
func (r *SnpExtendedReportReqABI) Finish(b BinaryConvertible) error {
	s, ok := b.(*SnpExtendedReportReq)
	if !ok {
		return fmt.Errorf("Finish argument is %v. Expects a *SnpExtendedReportReq", reflect.TypeOf(b))
	}
	s.CertsLength = r.CertsLength
	return nil
}

// ABI returns an object that can cross the ABI boundary and copy back changes to the original
// object.
func (r *SnpExtendedReportReq) ABI() BinaryConversion {
	var certsAddress unsafe.Pointer
	if len(r.Certs) != 0 {
		certsAddress = unsafe.Pointer(&r.Certs[0])
	}
	return &SnpExtendedReportReqABI{
		Data:         r.Data,
		CertsAddress: certsAddress,
		CertsLength:  r.CertsLength,
	}
}

// SnpUserGuestRequestABI is Linux's sev-guest ioctl abi for issuing a guest message.
type SnpUserGuestRequestABI struct {
	GuestMsgVersion uint32
	// Request and response structure address.
	ReqData  unsafe.Pointer
	RespData unsafe.Pointer
	// firmware error code on failure (see psp-sev.h in Linux kernel)
	FwErr uint64
}

type snpUserGuestRequestConversion struct {
	abi      SnpUserGuestRequestABI
	reqConv  BinaryConversion
	respConv BinaryConversion
}

// SnpUserGuestRequest is Linux's sev-guest ioctl interface for issuing a guest message. The
// types here enhance runtime safety when using Ioctl as an interface.
type SnpUserGuestRequest struct {
	// Request and response structure address.
	ReqData  BinaryConvertible
	RespData BinaryConvertible
	// firmware error code on failure (see psp-sev.h in Linux kernel)
	FwErr uint64
}

// ABI returns an object that can cross the ABI boundary and copy back changes to the original
// object.
func (r *SnpUserGuestRequest) ABI() BinaryConversion {
	result := &snpUserGuestRequestConversion{
		reqConv:  r.ReqData.ABI(),
		respConv: r.RespData.ABI(),
	}
	result.abi.GuestMsgVersion = guestMsgVersion
	result.abi.ReqData = result.reqConv.Pointer()
	result.abi.RespData = result.respConv.Pointer()
	return result
}

// Pointer returns a pointer to the object that crosses the ABI boundary.
func (r *snpUserGuestRequestConversion) Pointer() unsafe.Pointer {
	return unsafe.Pointer(&r.abi)
}

// Finish writes back the FwErr and any changes to the request or response objects.
func (r *snpUserGuestRequestConversion) Finish(b BinaryConvertible) error {
	s, ok := b.(*SnpUserGuestRequest)
	if !ok {
		return fmt.Errorf("Finish argument is %v. Expects a *SnpUserGuestRequestSafe", reflect.TypeOf(b))
	}
	if err := r.reqConv.Finish(s.ReqData); err != nil {
		return fmt.Errorf("could not finalize request data: %v", err)
	}
	if err := r.respConv.Finish(s.RespData); err != nil {
		return fmt.Errorf("could not finalize response data: %v", err)
	}
	s.FwErr = r.abi.FwErr
	return nil
}

// BinaryConversion is an interface that abstracts a "stand-in" object that passes through an ABI
// boundary and can finalize changes to the original object.
type BinaryConversion interface {
	Pointer() unsafe.Pointer
	Finish(BinaryConvertible) error
}

// BinaryConvertible is an interface for an object that can produce a partner BinaryConversion
// object to allow its representation to pass the ABI boundary.
type BinaryConvertible interface {
	ABI() BinaryConversion
}
