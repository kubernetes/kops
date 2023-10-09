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

//go:build linux

// Package client provides an interface to the Intel TDX guest device commands.
package client

import (
	"fmt"

	"unsafe"

	labi "github.com/google/go-tdx-guest/client/linuxabi"
	"golang.org/x/sys/unix"
)

// defaultTdxGuestDevicePath is the platform's usual device path to the TDX guest.
const defaultTdxGuestDevicePath = "/dev/tdx_guest"

// LinuxDevice implements the Device interface with Linux ioctls.
type LinuxDevice struct {
	fd int
}

// Open opens the TDX guest device from a given path
func (d *LinuxDevice) Open(path string) error {
	fd, err := unix.Open(path, unix.O_RDWR|unix.O_SYNC, 0)
	if err != nil {
		d.fd = -1
		return fmt.Errorf("could not open Intel TDX guest device at %q: %v", path, err)
	}
	d.fd = fd
	return nil
}

// OpenDevice opens the TDX guest device.
func OpenDevice() (*LinuxDevice, error) {
	result := &LinuxDevice{}
	path := *tdxGuestPath
	if UseDefaultTdxGuestDevice() {
		path = defaultTdxGuestDevicePath
	}
	if err := result.Open(path); err != nil {
		return nil, err
	}
	return result, nil
}

// Close closes the TDX guest device.
func (d *LinuxDevice) Close() error {
	if d.fd == -1 { // Not open
		return nil
	}
	if err := unix.Close(d.fd); err != nil {
		return err
	}
	// Prevent double-close.
	d.fd = -1
	return nil
}

// Ioctl sends a command with its wrapped request and response values to the Linux device.
func (d *LinuxDevice) Ioctl(command uintptr, req any) (uintptr, error) {
	if d.fd == -1 {
		return 0, fmt.Errorf("intel TDX Guest Device is not open")
	}
	switch sreq := req.(type) {
	case *labi.TdxQuoteReq:
		abi := sreq.ABI()
		result, _, errno := unix.Syscall(unix.SYS_IOCTL, uintptr(d.fd), command, uintptr(abi.Pointer()))
		abi.Finish(sreq)
		if errno == unix.EBUSY {
			return 0, errno
		}
		if errno != 0 {
			return 0, errno
		}
		return result, nil
	case *labi.TdxReportReq:
		result, _, errno := unix.Syscall(unix.SYS_IOCTL, uintptr(d.fd), command, uintptr(unsafe.Pointer(req.(*labi.TdxReportReq))))
		if errno != 0 {
			return 0, errno
		}
		return result, nil
	}
	return 0, fmt.Errorf("unexpected request value: %v", req)
}
