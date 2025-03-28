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

//go:build windows

package client

import (
	"fmt"

	spb "github.com/google/go-sev-guest/proto/sevsnp"
)

// WindowsDevice implements the Device interface with Linux ioctls.
type WindowsDevice struct{}

// Open is not supported on Windows.
func (*WindowsDevice) Open(_ string) error {
	return fmt.Errorf("Windows is unsupported")
}

// OpenDevice fails on Windows.
func OpenDevice() (*WindowsDevice, error) {
	return nil, fmt.Errorf("Windows is unsupported")
}

// Close is not supported on Windows.
func (*WindowsDevice) Close() error {
	return fmt.Errorf("Windows is unsupported")
}

// Ioctl is not supported on Windows.
func (*WindowsDevice) Ioctl(_ uintptr, _ any) (uintptr, error) {
	// The GuestAttestation library on Windows is closed source.
	return 0, fmt.Errorf("Windows is unsupported")
}

// Product is not supported on Windows.
func (*WindowsDevice) Product() *spb.SevProduct {
	return &spb.SevProduct{}
}

// WindowsQuoteProvider implements the QuoteProvider interface with Linux's configfs-tsm.
type WindowsQuoteProvider struct{}

// IsSupported checks if the quote provider is supported.
func (*WindowsQuoteProvider) IsSupported() bool {
	return false
}

// GetRawQuote returns byte format attestation plus certificate table via ConfigFS.
func (*WindowsQuoteProvider) GetRawQuote(reportData [64]byte) ([]byte, error) {
	return nil, fmt.Errorf("Windows is unsupported")
}

// GetRawQuoteAtLevel returns byte format attestation plus certificate table via ConfigFS.
func (*WindowsQuoteProvider) GetRawQuoteAtLevel(reportData [64]byte, level uint) ([]byte, error) {
	return nil, fmt.Errorf("Windows is unsupported")
}

// GetQuoteProvider returns a supported SEV-SNP QuoteProvider.
func GetQuoteProvider() (QuoteProvider, error) {
	return nil, fmt.Errorf("Windows is unsupported")
}

// GetLeveledQuoteProvider returns a supported SEV-SNP LeveledQuoteProvider.
func GetLeveledQuoteProvider() (LeveledQuoteProvider, error) {
	return nil, fmt.Errorf("Windows is unsupported")
}
