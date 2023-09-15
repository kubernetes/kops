// Copyright 2023 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build darwin

// Package client provides an interface to the Intel TDX guest device commands.
package client

import (
	"fmt"
)

// defaultTdxGuestDevicePath is the platform's usual device path to the TDX guest.
const defaultTdxGuestDevicePath = "unknown"

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
