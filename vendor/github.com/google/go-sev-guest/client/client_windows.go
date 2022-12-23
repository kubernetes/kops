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
)

// WindowsDevice implements the Device interface with Linux ioctls.
type WindowsDevice struct{}

// Open is not supported on Windows.
func (d *WindowsDevice) Open(path string) error {
	return fmt.Errorf("Windows is unsupported")
}

// OpenDevice fails on Windows.
func OpenDevice() (*WindowsDevice, error) {
	return nil, fmt.Errorf("Windows is unsupported")
}

// Close is not supported on Windows.
func (d *WindowsDevice) Close() error {
	return fmt.Errorf("Windows is unsupported")
}

// Ioctl is not supported on Windows.
func (d *WindowsDevice) Ioctl(command uintptr, req any) (uintptr, error) {
	// The GuestAttestation library on Windows is closed source.
	return 0, fmt.Errorf("Windows is unsupported")
}
