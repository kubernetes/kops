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

//go:build linux || freebsd || openbsd || netbsd

// Package client provides an interface to the AMD SEV-SNP guest device commands.
package client

import (
	"flag"
	"fmt"
	"strconv"
	"time"

	"github.com/google/go-configfs-tsm/configfs/linuxtsm"
	"github.com/google/go-configfs-tsm/report"
	"github.com/google/go-sev-guest/abi"
	labi "github.com/google/go-sev-guest/client/linuxabi"
	spb "github.com/google/go-sev-guest/proto/sevsnp"
	"golang.org/x/sys/unix"
)

const (
	// defaultSevGuestDevicePath is the platform's usual device path to the SEV guest.
	defaultSevGuestDevicePath = "/dev/sev-guest"
	installURL                = "https://github.com/google/go-sev-guest/blob/main/INSTALL.md"
)

// These flags should not be needed for long term health of the project as the Linux kernel
// catches up with throttling-awareness.
var (
	throttleDuration = flag.Duration("self_throttle_duration", 2*time.Second, "Rate-limit library-initiated device commands to this duration")
	burstMax         = flag.Int("self_throttle_burst", 1, "Rate-limit library-initiated device commands to this many commands per duration")
	defaultVMPL      = flag.String("default_vmpl", "", "Default VMPL to use for attestation (empty for driver default)")
)

// LinuxDevice implements the Device interface with Linux ioctls.
type LinuxDevice struct {
	fd      int
	lastCmd time.Time
	burst   int
}

// Open opens the SEV-SNP guest device from a given path
func (d *LinuxDevice) Open(path string) error {
	fd, err := unix.Open(path, unix.O_RDWR, 0)
	if err != nil {
		d.fd = -1
		return fmt.Errorf("could not open AMD SEV guest device at %s (see %s): %v", path, installURL, err)
	}
	d.fd = fd
	return nil
}

// OpenDevice opens the SEV-SNP guest device.
func OpenDevice() (*LinuxDevice, error) {
	result := &LinuxDevice{}
	path := *sevGuestPath
	if UseDefaultSevGuest() {
		path = defaultSevGuestDevicePath
	}
	if err := result.Open(path); err != nil {
		return nil, err
	}
	return result, nil
}

// Close closes the SEV-SNP guest device.
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
	// TODO(Issue #40): Remove the workaround to the ENOTTY lockout when throttled
	// in Linux 6.1 by throttling ourselves first.
	if d.burst == 0 {
		sinceLast := time.Since(d.lastCmd)
		// Self-throttle for tests without guest OS throttle detection
		if sinceLast < *throttleDuration {
			time.Sleep(*throttleDuration - sinceLast)
		}
	}
	switch sreq := req.(type) {
	case *labi.SnpUserGuestRequest:
		abi := sreq.ABI()
		result, _, errno := unix.Syscall(unix.SYS_IOCTL, uintptr(d.fd), command, uintptr(abi.Pointer()))
		abi.Finish(sreq)
		d.burst = (d.burst + 1) % *burstMax
		if d.burst == 0 {
			d.lastCmd = time.Now()
		}

		// TODO(Issue #5): remove the work around for the kernel bug that writes
		// uninitialized memory back on non-EIO.
		if errno != unix.EIO {
			sreq.FwErr = 0
		}
		if errno != 0 {
			return 0, errno
		}
		return result, nil
	}
	return 0, fmt.Errorf("unexpected request value: %v", req)
}

// Product returns the current CPU's associated AMD SEV product information.
func (d *LinuxDevice) Product() *spb.SevProduct {
	return abi.SevProduct()
}

// LinuxIoctlQuoteProvider implements the QuoteProvider interface to fetch
// attestation quote via the deprecated /dev/sev-guest ioctl.
type LinuxIoctlQuoteProvider struct{}

// IsSupported checks if TSM client can be created to use /dev/sev-guest ioctl.
func (p *LinuxIoctlQuoteProvider) IsSupported() bool {
	d, err := OpenDevice()
	if err != nil {
		return false
	}
	d.Close()
	return true
}

// GetRawQuoteAtLevel returns byte format attestation plus certificate table via /dev/sev-guest ioctl.
func (p *LinuxIoctlQuoteProvider) GetRawQuoteAtLevel(reportData [64]byte, level uint) ([]uint8, error) {
	d, err := OpenDevice()
	if err != nil {
		return nil, err
	}
	defer d.Close()
	// If there are no certificates, then just return the raw report.
	length, err := queryCertificateLength(d, int(level))
	if err != nil {
		return GetRawReportAtVmpl(d, reportData, int(level))
	}
	certs := make([]byte, length)
	report, _, err := getExtendedReportIn(d, reportData, int(level), certs)
	if err != nil {
		return nil, err
	}
	// Mix the platform info in with the auxblob.
	extended, err := abi.ExtendedPlatformCertTable(certs)
	if err != nil {
		return nil, fmt.Errorf("invalid certificate table: %v", err)
	}
	return append(report, extended...), nil
}

// GetRawQuote returns byte format attestation plus certificate table via /dev/sev-guest ioctl.
func (p *LinuxIoctlQuoteProvider) GetRawQuote(reportData [64]byte) ([]uint8, error) {
	if *defaultVMPL == "" {
		return p.GetRawQuoteAtLevel(reportData, 0)
	}
	vmpl, err := strconv.ParseUint(*defaultVMPL, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("bad default_vmpl: %q", *defaultVMPL)
	}
	return p.GetRawQuoteAtLevel(reportData, uint(vmpl))
}

// Product returns the current CPU's associated AMD SEV product information.
//
// Deprecated: Use ExtraPlatformInfoGUID from the cert table.
func (*LinuxIoctlQuoteProvider) Product() *spb.SevProduct {
	return abi.SevProduct()
}

// LinuxConfigFsQuoteProvider implements the QuoteProvider interface to fetch
// attestation quote via ConfigFS.
type LinuxConfigFsQuoteProvider struct{}

// IsSupported checks if TSM client can be created to use ConfigFS system.
func (p *LinuxConfigFsQuoteProvider) IsSupported() bool {
	c, err := linuxtsm.MakeClient()
	if err != nil {
		return false
	}
	r, err := report.Create(c, &report.Request{})
	if err != nil {
		return false
	}
	provider, err := r.ReadOption("provider")
	return err == nil && string(provider) == "sev_guest\n"
}

// GetRawQuoteAtLevel returns byte format attestation plus certificate table via ConfigFS.
func (p *LinuxConfigFsQuoteProvider) GetRawQuoteAtLevel(reportData [64]byte, level uint) ([]uint8, error) {
	req := &report.Request{
		InBlob:     reportData[:],
		GetAuxBlob: true,
		Privilege: &report.Privilege{
			Level: level,
		},
	}
	resp, err := linuxtsm.GetReport(req)
	if err != nil {
		return nil, err
	}
	// Mix the platform info in with the auxblob.
	extended, err := abi.ExtendedPlatformCertTable(resp.AuxBlob)
	if err != nil {
		return nil, fmt.Errorf("invalid certificate table: %v", err)
	}
	return append(resp.OutBlob, extended...), nil
}

// GetRawQuote returns byte format attestation plus certificate table via ConfigFS.
func (p *LinuxConfigFsQuoteProvider) GetRawQuote(reportData [64]byte) ([]uint8, error) {
	req := &report.Request{
		InBlob:     reportData[:],
		GetAuxBlob: true,
	}
	if *defaultVMPL != "" {
		vmpl, err := strconv.ParseUint(*defaultVMPL, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("bad default_vmpl: %q", *defaultVMPL)
		}
		req.Privilege = &report.Privilege{
			Level: uint(vmpl),
		}
	}
	resp, err := linuxtsm.GetReport(req)
	if err != nil {
		return nil, err
	}
	// Mix the platform info in with the auxblob.
	extended, err := abi.ExtendedPlatformCertTable(resp.AuxBlob)
	if err != nil {
		return nil, fmt.Errorf("invalid certificate table: %v", err)
	}
	return append(resp.OutBlob, extended...), nil
}

// Product returns the current CPU's associated AMD SEV product information.
//
// Deprecated: Use ExtraPlatformInfoGUID from the cert table.
func (*LinuxConfigFsQuoteProvider) Product() *spb.SevProduct {
	return abi.SevProduct()
}

// GetQuoteProvider returns a supported SEV-SNP QuoteProvider.
func GetQuoteProvider() (QuoteProvider, error) {
	var provider QuoteProvider
	provider = &LinuxConfigFsQuoteProvider{}
	if provider.IsSupported() {
		return provider, nil
	}
	provider = &LinuxIoctlQuoteProvider{}
	if provider.IsSupported() {
		return provider, nil
	}
	return nil, fmt.Errorf("no supported SEV-SNP QuoteProvider found")
}

// GetLeveledQuoteProvider returns a supported SEV-SNP LeveledQuoteProvider.
func GetLeveledQuoteProvider() (LeveledQuoteProvider, error) {
	var provider LeveledQuoteProvider
	provider = &LinuxConfigFsQuoteProvider{}
	if provider.IsSupported() {
		return provider, nil
	}
	provider = &LinuxIoctlQuoteProvider{}
	if provider.IsSupported() {
		return provider, nil
	}
	return nil, fmt.Errorf("no supported SEV-SNP LeveledQuoteProvider found")
}
