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

// Package tbs provides an low-level interface directly mapping to Windows
// Tbs.dll system library commands:
// https://docs.microsoft.com/en-us/windows/desktop/TBS/tpm-base-services-portal
// Public field descriptions contain links to the high-level Windows documentation.
package tbs

import (
	"fmt"
	"syscall"
	"unsafe"
)

// Context references the current TPM context
type Context uintptr

// Version of TPM being used by the application.
type Version uint32

// Flag indicates TPM versions that are supported by the application.
type Flag uint32

// CommandPriority is used to determine which pending command to submit whenever the TPM is free.
type CommandPriority uint32

// Command parameters:
// https://github.com/tpn/winsdk-10/blob/master/Include/10.0.10240.0/shared/tbs.h
const (
	// https://docs.microsoft.com/en-us/windows/desktop/api/Tbs/ns-tbs-tdtbs_context_params2
	// OR flags to use multiple.
	RequestRaw   Flag = 1 << iota // Add flag to request raw context
	IncludeTPM12                  // Add flag to support TPM 1.2
	IncludeTPM20                  // Add flag to support TPM 2

	TPMVersion12 Version = 1 // For TPM 1.2 applications
	TPMVersion20 Version = 2 // For TPM 2 applications or applications using multiple TPM versions

	// https://docs.microsoft.com/en-us/windows/desktop/tbs/command-scheduling
	// https://docs.microsoft.com/en-us/windows/desktop/api/Tbs/nf-tbs-tbsip_submit_command#parameters
	LowPriority    CommandPriority = 100 // For low priority application use
	NormalPriority CommandPriority = 200 // For normal priority application use
	HighPriority   CommandPriority = 300 // For high priority application use
	SystemPriority CommandPriority = 400 // For system tasks that access the TPM

	commandLocalityZero uint32 = 0 // Windows currently only supports TBS_COMMAND_LOCALITY_ZERO.
)

// Error is the return type of all functions in this package.
type Error uint32

func (err Error) Error() string {
	if description, ok := errorDescriptions[err]; ok {
		return fmt.Sprintf("TBS Error 0x%X: %s", uint32(err), description)
	}
	return fmt.Sprintf("Unrecognized TBS Error 0x%X", uint32(err))
}

func getError(err uintptr) error {
	// tbs.dll uses 0x0 as the return value for success.
	if err == 0 {
		return nil
	}
	return Error(err)
}

// TBS Return Codes:
// https://docs.microsoft.com/en-us/windows/desktop/TBS/tbs-return-codes
const (
	ErrInternalError          Error = 0x80284001
	ErrBadParameter           Error = 0x80284002
	ErrInvalidOutputPointer   Error = 0x80284003
	ErrInvalidContext         Error = 0x80284004
	ErrInsufficientBuffer     Error = 0x80284005
	ErrIOError                Error = 0x80284006
	ErrInvalidContextParam    Error = 0x80284007
	ErrServiceNotRunning      Error = 0x80284008
	ErrTooManyTBSContexts     Error = 0x80284009
	ErrTooManyResources       Error = 0x8028400A
	ErrServiceStartPending    Error = 0x8028400B
	ErrPPINotSupported        Error = 0x8028400C
	ErrCommandCanceled        Error = 0x8028400D
	ErrBufferTooLarge         Error = 0x8028400E
	ErrTPMNotFound            Error = 0x8028400F
	ErrServiceDisabled        Error = 0x80284010
	ErrNoEventLog             Error = 0x80284011
	ErrAccessDenied           Error = 0x80284012
	ErrProvisioningNotAllowed Error = 0x80284013
	ErrPPIFunctionUnsupported Error = 0x80284014
	ErrOwnerauthNotFound      Error = 0x80284015
)

var errorDescriptions = map[Error]string{
	ErrInternalError:          "An internal software error occurred.",
	ErrBadParameter:           "One or more parameter values are not valid.",
	ErrInvalidOutputPointer:   "A specified output pointer is bad.",
	ErrInvalidContext:         "The specified context handle does not refer to a valid context.",
	ErrInsufficientBuffer:     "The specified output buffer is too small.",
	ErrIOError:                "An error occurred while communicating with the TPM.",
	ErrInvalidContextParam:    "A context parameter that is not valid was passed when attempting to create a TBS context.",
	ErrServiceNotRunning:      "The TBS service is not running and could not be started.",
	ErrTooManyTBSContexts:     "A new context could not be created because there are too many open contexts.",
	ErrTooManyResources:       "A new virtual resource could not be created because there are too many open virtual resources.",
	ErrServiceStartPending:    "The TBS service has been started but is not yet running.",
	ErrPPINotSupported:        "The physical presence interface is not supported.",
	ErrCommandCanceled:        "The command was canceled.",
	ErrBufferTooLarge:         "The input or output buffer is too large.",
	ErrTPMNotFound:            "A compatible Trusted Platform Module (TPM) Security Device cannot be found on this computer.",
	ErrServiceDisabled:        "The TBS service has been disabled.",
	ErrNoEventLog:             "The TBS event log is not available.",
	ErrAccessDenied:           "The caller does not have the appropriate rights to perform the requested operation.",
	ErrProvisioningNotAllowed: "The TPM provisioning action is not allowed by the specified flags.",
	ErrPPIFunctionUnsupported: "The Physical Presence Interface of this firmware does not support the requested method.",
	ErrOwnerauthNotFound:      "The requested TPM OwnerAuth value was not found.",
}

// Tbs.dll provides an API for making calls to the TPM:
// https://docs.microsoft.com/en-us/windows/desktop/TBS/tpm-base-services-portal
var (
	tbsDLL           = syscall.NewLazyDLL("Tbs.dll")
	tbsGetDeviceInfo = tbsDLL.NewProc("Tbsi_GetDeviceInfo")
	tbsCreateContext = tbsDLL.NewProc("Tbsi_Context_Create")
	tbsContextClose  = tbsDLL.NewProc("Tbsip_Context_Close")
	tbsSubmitCommand = tbsDLL.NewProc("Tbsip_Submit_Command")
	tbsGetTCGLog     = tbsDLL.NewProc("Tbsi_Get_TCG_Log")
)

// Returns the address of the beginning of a slice or 0 for a nil slice.
func sliceAddress(s []byte) uintptr {
	if len(s) == 0 {
		return 0
	}
	return uintptr(unsafe.Pointer(&(s[0])))
}

// DeviceInfo is TPM_DEVICE_INFO from tbs.h
type DeviceInfo struct {
	StructVersion    uint32
	TPMVersion       Version
	TPMInterfaceType uint32
	TPMImpRevision   uint32
}

// GetDeviceInfo gets the DeviceInfo of the current TPM:
// https://docs.microsoft.com/en-us/windows/win32/api/tbs/nf-tbs-tbsi_getdeviceinfo
func GetDeviceInfo() (*DeviceInfo, error) {
	info := DeviceInfo{}
	// TBS_RESULT Tbsi_GetDeviceInfo(
	//   UINT32 Size,
	//   PVOID  Info
	// );
	if err := tbsGetDeviceInfo.Find(); err != nil {
		return nil, err
	}
	result, _, _ := tbsGetDeviceInfo.Call(
		unsafe.Sizeof(info),
		uintptr(unsafe.Pointer(&info)),
	)
	return &info, getError(result)
}

// CreateContext creates a new TPM context:
// https://docs.microsoft.com/en-us/windows/desktop/api/Tbs/nf-tbs-tbsi_context_create
func CreateContext(version Version, flag Flag) (Context, error) {
	var context Context
	params := struct {
		Version
		Flag
	}{version, flag}
	// TBS_RESULT Tbsi_Context_Create(
	//   _In_  PCTBS_CONTEXT_PARAMS pContextParams,
	//   _Out_ PTBS_HCONTEXT        *phContext
	// );
	if err := tbsCreateContext.Find(); err != nil {
		return context, err
	}
	result, _, _ := tbsCreateContext.Call(
		uintptr(unsafe.Pointer(&params)),
		uintptr(unsafe.Pointer(&context)),
	)
	return context, getError(result)
}

// Close closes an existing TPM context:
// https://docs.microsoft.com/en-us/windows/desktop/api/Tbs/nf-tbs-tbsip_context_close
func (context Context) Close() error {
	// TBS_RESULT Tbsip_Context_Close(
	//   _In_ TBS_HCONTEXT hContext
	// );
	if err := tbsContextClose.Find(); err != nil {
		return err
	}
	result, _, _ := tbsContextClose.Call(uintptr(context))
	return getError(result)
}

// SubmitCommand sends commandBuffer to the TPM, returning the number of bytes
// written to responseBuffer. ErrInsufficientBuffer is returned if the
// responseBuffer is too short. ErrInvalidOutputPointer is returned if the
// responseBuffer is nil. On failure, the returned length is unspecified.
// https://docs.microsoft.com/en-us/windows/desktop/api/Tbs/nf-tbs-tbsip_submit_command
func (context Context) SubmitCommand(
	priority CommandPriority,
	commandBuffer []byte,
	responseBuffer []byte,
) (uint32, error) {
	responseBufferLen := uint32(len(responseBuffer))

	// TBS_RESULT Tbsip_Submit_Command(
	//   _In_          TBS_HCONTEXT         hContext,
	//   _In_          TBS_COMMAND_LOCALITY Locality,
	//   _In_          TBS_COMMAND_PRIORITY Priority,
	//   _In_    const PCBYTE               *pabCommand,
	//   _In_          UINT32               cbCommand,
	//   _Out_         PBYTE                *pabResult,
	//   _Inout_       UINT32               *pcbOutput
	// );
	if err := tbsSubmitCommand.Find(); err != nil {
		return 0, err
	}
	result, _, _ := tbsSubmitCommand.Call(
		uintptr(context),
		uintptr(commandLocalityZero),
		uintptr(priority),
		sliceAddress(commandBuffer),
		uintptr(len(commandBuffer)),
		sliceAddress(responseBuffer),
		uintptr(unsafe.Pointer(&responseBufferLen)),
	)
	return responseBufferLen, getError(result)
}

// GetTCGLog gets the system event log, returning the number of bytes written
// to logBuffer. If logBuffer is nil, the size of the TCG log is returned.
// ErrInsufficientBuffer is returned if the logBuffer is too short. On failure,
// the returned length is unspecified.
// https://docs.microsoft.com/en-us/windows/desktop/api/Tbs/nf-tbs-tbsi_get_tcg_log
func (context Context) GetTCGLog(logBuffer []byte) (uint32, error) {
	logBufferLen := uint32(len(logBuffer))

	// TBS_RESULT Tbsi_Get_TCG_Log(
	//   TBS_HCONTEXT hContext,
	//   PBYTE        pOutputBuf,
	//   PUINT32      pOutputBufLen
	// );
	if err := tbsGetTCGLog.Find(); err != nil {
		return 0, err
	}
	result, _, _ := tbsGetTCGLog.Call(
		uintptr(context),
		sliceAddress(logBuffer),
		uintptr(unsafe.Pointer(&logBufferLen)),
	)
	return logBufferLen, getError(result)
}
