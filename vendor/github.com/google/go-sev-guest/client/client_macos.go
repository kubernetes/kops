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
)

// DefaultSevGuestDevicePath is the platform's usual device path to the SEV guest.
const DefaultSevGuestDevicePath = "unknown"

// MacOSDevice implements the Device interface with Linux ioctls.
type MacOSDevice struct{}

// Open is not supported on MacOS.
func (d *MacOSDevice) Open(path string) error {
	return fmt.Errorf("MacOS is unsupported")
}

// OpenDevice fails on MacOS.
func OpenDevice() (*MacOSDevice, error) {
	return nil, fmt.Errorf("MacOS is unsupported")
}

// Close is not supported on MacOS.
func (d *MacOSDevice) Close() error {
	return fmt.Errorf("MacOS is unsupported")
}

// Ioctl is not supported on MacOS.
func (d *MacOSDevice) Ioctl(command uintptr, req any) (uintptr, error) {
	return 0, fmt.Errorf("MacOS is unsupported")
}
