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

//go:build darwin

package client

import (
	"fmt"

	spb "github.com/google/go-sev-guest/proto/sevsnp"
)

// DefaultSevGuestDevicePath is the platform's usual device path to the SEV guest.
const DefaultSevGuestDevicePath = "unknown"

// MacOSDevice implements the Device interface with Linux ioctls.
type MacOSDevice struct{}

// Open is not supported on MacOS.
func (*MacOSDevice) Open(_ string) error {
	return fmt.Errorf("MacOS is unsupported")
}

// OpenDevice fails on MacOS.
func OpenDevice() (*MacOSDevice, error) {
	return nil, fmt.Errorf("MacOS is unsupported")
}

// Close is not supported on MacOS.
func (*MacOSDevice) Close() error {
	return fmt.Errorf("MacOS is unsupported")
}

// Ioctl is not supported on MacOS.
func (*MacOSDevice) Ioctl(_ uintptr, _ any) (uintptr, error) {
	return 0, fmt.Errorf("MacOS is unsupported")
}

// Product is not supported on MacOS.
func (*MacOSDevice) Product() *spb.SevProduct {
	return &spb.SevProduct{}
}

// MacOSQuoteProvider implements the QuoteProvider interface with Linux's configfs-tsm.
type MacOSQuoteProvider struct{}

// IsSupported checks if the quote provider is supported.
func (*MacOSQuoteProvider) IsSupported() bool {
	return false
}

// GetRawQuote returns byte format attestation plus certificate table via ConfigFS.
func (*MacOSQuoteProvider) GetRawQuote(reportData [64]byte) ([]byte, error) {
	return nil, fmt.Errorf("MacOS is unsupported")
}

// GetRawQuoteAtLevel returns byte format attestation plus certificate table via ConfigFS.
func (*MacOSQuoteProvider) GetRawQuoteAtLevel(reportData [64]byte, level uint) ([]byte, error) {
	return nil, fmt.Errorf("MacOS is unsupported")
}

// GetQuoteProvider returns a supported SEV-SNP QuoteProvider.
func GetQuoteProvider() (QuoteProvider, error) {
	return nil, fmt.Errorf("MacOS is unsupported")
}

// GetLeveledQuoteProvider returns a supported SEV-SNP LeveledQuoteProvider.
func GetLeveledQuoteProvider() (LeveledQuoteProvider, error) {
	return nil, fmt.Errorf("MacOS is unsupported")
}
