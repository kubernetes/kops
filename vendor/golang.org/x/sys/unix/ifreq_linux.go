// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux
// +build linux

package unix

import (
	"bytes"
	"unsafe"
)

// Helpers for dealing with ifreq since it contains a union and thus requires a
// lot of unsafe.Pointer casts to use properly.

// An Ifreq is a type-safe wrapper around the raw ifreq struct. An Ifreq
// contains an interface name and a union of arbitrary data which can be
// accessed using the Ifreq's methods. To create an Ifreq, use the NewIfreq
// function.
//
// Use the Name method to access the stored interface name. The union data
// fields can be get and set using the following methods:
//   - Uint16/SetUint16: flags
//   - Uint32/SetUint32: ifindex, metric, mtu
type Ifreq struct{ raw ifreq }

// NewIfreq creates an Ifreq with the input network interface name after
// validating the name does not exceed IFNAMSIZ-1 (trailing NULL required)
// bytes.
func NewIfreq(name string) (*Ifreq, error) {
	// Leave room for terminating NULL byte.
	if len(name) >= IFNAMSIZ {
		return nil, EINVAL
	}

	var ifr ifreq
	copy(ifr.Ifrn[:], name)

	return &Ifreq{raw: ifr}, nil
}

// TODO(mdlayher): get/set methods for sockaddr, char array, etc.

// Name returns the interface name associated with the Ifreq.
func (ifr *Ifreq) Name() string {
	// BytePtrToString requires a NULL terminator or the program may crash. If
	// one is not present, just return the empty string.
	if !bytes.Contains(ifr.raw.Ifrn[:], []byte{0x00}) {
		return ""
	}

	return BytePtrToString(&ifr.raw.Ifrn[0])
}

// Uint16 returns the Ifreq union data as a C short/Go uint16 value.
func (ifr *Ifreq) Uint16() uint16 {
	return *(*uint16)(unsafe.Pointer(&ifr.raw.Ifru[:2][0]))
}

// SetUint16 sets a C short/Go uint16 value as the Ifreq's union data.
func (ifr *Ifreq) SetUint16(v uint16) {
	ifr.clear()
	*(*uint16)(unsafe.Pointer(&ifr.raw.Ifru[:2][0])) = v
}

// Uint32 returns the Ifreq union data as a C int/Go uint32 value.
func (ifr *Ifreq) Uint32() uint32 {
	return *(*uint32)(unsafe.Pointer(&ifr.raw.Ifru[:4][0]))
}

// SetUint32 sets a C int/Go uint32 value as the Ifreq's union data.
func (ifr *Ifreq) SetUint32(v uint32) {
	ifr.clear()
	*(*uint32)(unsafe.Pointer(&ifr.raw.Ifru[:4][0])) = v
}

// clear zeroes the ifreq's union field to prevent trailing garbage data from
// being sent to the kernel if an ifreq is reused.
func (ifr *Ifreq) clear() {
	for i := range ifr.raw.Ifru {
		ifr.raw.Ifru[i] = 0
	}
}

// TODO(mdlayher): export as IfreqData? For now we can provide helpers such as
// IoctlGetEthtoolDrvinfo which use these APIs under the hood.

// An ifreqData is an Ifreq which carries pointer data. To produce an ifreqData,
// use the Ifreq.withData method.
type ifreqData struct {
	name [IFNAMSIZ]byte
	// A type separate from ifreq is required in order to comply with the
	// unsafe.Pointer rules since the "pointer-ness" of data would not be
	// preserved if it were cast into the byte array of a raw ifreq.
	data unsafe.Pointer
	// Pad to the same size as ifreq.
	_ [len(ifreq{}.Ifru) - SizeofPtr]byte
}

// withData produces an ifreqData with the pointer p set for ioctls which require
// arbitrary pointer data.
func (ifr Ifreq) withData(p unsafe.Pointer) ifreqData {
	return ifreqData{
		name: ifr.raw.Ifrn,
		data: p,
	}
}
