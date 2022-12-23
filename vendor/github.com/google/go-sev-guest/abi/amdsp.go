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

package abi

import "fmt"

// SevFirmwareStatus is the type of all AMD-SP firmware status codes, as documented in the SEV API
// https://www.amd.com/system/files/TechDocs/55766_SEV-KM_API_Specification.pdf
type SevFirmwareStatus int

// Unexported errors are not expected to leave the kernel.
const (
	// Success denotes successful completion of a firmware command.
	Success SevFirmwareStatus = 0
	// InvalidPlatformState is the code for the platform to be in the wrong state for a given command.
	InvalidPlatformState = 1
	// InvalidGuestState is the code for the guest to be in the wrong state for a given command.
	InvalidGuestState = 2
	// Platform owner error unexpected by guest command.
	// invalidConfig = 3
	// InvalidLength is the code for a provided buffer size is too small to complete the command.
	InvalidLength = 4
	// Platform owner error unexpected by guest command.
	// alreadyOwned = 5
	// Platform owner error unexpected by guest command.
	// invalidCertificate = 6
	// PolicyFailure is the code for when the guest policy disallows the command.
	PolicyFailure = 7
	// Inactive is the code for when a command is sent for a guest, but the guest is inactive.
	Inactive = 8
	// InvalidAddress is the code for when a provided address is invalid.
	InvalidAddress = 9
	// User error expected at launch, unexpected here.
	// badSignature = 10
	// User error expected at launch, unexpected here.
	// badMeasurement = 11
	// Kernel error, unexpected.
	// asidOwned = 12
	// Kernel error, unexpected.
	// invalidAsid = 13
	// Kernel error, unexpected.
	// wbinvdRequired = 14
	// Kernel error, unexpected.
	// dfFlushRequired = 15
	// Kernel error, unexpected.
	// invalidGuest = 16
	// InvalidCommand is the code for when the command code is invalid.
	InvalidCommand = 17
	// Kernel error, unexpected.
	// active = 18
	// HwErrorPlatform is the code for when the hardware failed but it's okay to update its buffers.
	HwErrorPlatform = 19
	// HwErrorUnsafe is the code for when the hardware failed and it's unsafe to update its buffers.
	HwErrorUnsafe = 20
	// Unsupported is for an unsupported feature.
	Unsupported = 21
	// InvalidParam is the code for an invalid parameter in a command.
	InvalidParam = 22
	// ResourceLimit is the code for when the firmware has reached a resource limit and can't complete the command.
	ResourceLimit = 23
	// SecureDataInvalid is the code for when a hardware integrity check has failed.
	SecureDataInvalid = 24
	// InvalidPageSize indicates an RMP error with the recorded page size.
	InvalidPageSize = 25
	// InvalidPageState indicates an RMP error with the recorded page state.
	InvalidPageState = 26
	// InvalidMdataEntry indicates an RMP error with the recorded metadata.
	InvalidMdataEntry = 27
	// InvalidPageOwner indicates an RMP error with ASID mismatch between accessors.
	InvalidPageOwner = 28
	// AeadOflow indicates that firmware memory capacity is reached in the AEAD cryptographic algorithm.
	AeadOflow = 29
	// Skip code 0x1E since AeaedOflow is 0x1D and rbModeExited is 0x1F.
	// reserved1e = 30
	// Kernel error, unexpected.
	// rbModeExited = 31
	// Kernel error, unexpected.
	// rmpInitRequired = 32
	// Platform management error, unexpected.
	// badSvn = 33
	// Platform management error, unexpected.
	// badVersion = 34
	// Platform management error, unexpected.
	// shutdownRequired = 35
	// Platform management error, unexpected.
	// updateFailed = 36
	// Platform management error, unexpected.
	// restoreRequired = 37
)

// GuestRequestInvalidLength is set by the ccp driver and not the AMD-SP when an guest extended
// request provides too few pages for the firmware to populate with data.
const GuestRequestInvalidLength SevFirmwareStatus = 0x100000000

// SevFirmwareErr is an error that interprets firmware status codes from the AMD secure processor.
type SevFirmwareErr struct {
	error
	Status SevFirmwareStatus
}

func (e SevFirmwareErr) Error() string {
	if e.Status == Success {
		return "success"
	}
	if e.Status == InvalidPlatformState {
		return "platform state is invalid for this command"
	}
	if e.Status == InvalidGuestState {
		return "guest state is invalid for this command"
	}
	if e.Status == InvalidLength {
		return "memory buffer is too small (library bug, please report)"
	}
	if e.Status == PolicyFailure {
		return "request is not allowed by guest policy"
	}
	if e.Status == Inactive {
		return "guest is inactive"
	}
	if e.Status == InvalidAddress {
		return "address provided is invalid (library bug, please report)"
	}
	if e.Status == InvalidCommand {
		return "invalid command (library bug, please report)"
	}
	if e.Status == HwErrorPlatform {
		return "hardware condition has occurred affecting the platform (report to sysadmin)"
	}
	if e.Status == HwErrorUnsafe {
		return "hardware condition has occurred affecting the platform. Buffers unsafe (report to sysadmin)"
	}
	if e.Status == Unsupported {
		return "unsupported feature"
	}
	if e.Status == InvalidParam {
		return "invalid parameter (library bug, please report)"
	}
	if e.Status == ResourceLimit {
		return "SEV firmware has run out of recources necessary to complete the command"
	}
	if e.Status == SecureDataInvalid {
		return "part-specific SEV data failed integrity checks (report to sysadmin)"
	}
	if e.Status == InvalidPageSize {
		return "RMP: invalid page size"
	}
	if e.Status == InvalidPageState {
		return "RMP: invalid page state"
	}
	if e.Status == InvalidMdataEntry {
		return "RMP: invalid recorded metadata"
	}
	if e.Status == InvalidPageOwner {
		return "RMP: ASID mismatch between accessors"
	}
	if e.Status == AeadOflow {
		return "AMD-SP firmware memory would be over capacity for AEAD use"
	}
	if e.Status == GuestRequestInvalidLength {
		return "too few extended guest request data pages"
	}
	return fmt.Sprintf("unexpected firmware status (see SEV API spec): %x", uint64(e.Status))
}
