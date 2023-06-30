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

package tpmutil

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// maxBytesBufferSize sets a sane upper bound on the size of a U32Bytes
// buffer. This limit exists to prevent a maliciously large size prefix
// from resulting in a massive memory allocation, potentially causing
// an OOM condition on the system.
// We expect no buffer from a TPM to approach 1Mb in size.
const maxBytesBufferSize uint32 = 1024 * 1024 // 1Mb.

// RawBytes is for Pack and RunCommand arguments that are already encoded.
// Compared to []byte, RawBytes will not be prepended with slice length during
// encoding.
type RawBytes []byte

// U16Bytes is a byte slice with a 16-bit header
type U16Bytes []byte

// TPMMarshal packs U16Bytes
func (b *U16Bytes) TPMMarshal(out io.Writer) error {
	size := len([]byte(*b))
	if err := binary.Write(out, binary.BigEndian, uint16(size)); err != nil {
		return err
	}

	n, err := out.Write(*b)
	if err != nil {
		return err
	}
	if n != size {
		return fmt.Errorf("unable to write all contents of U16Bytes")
	}
	return nil
}

// TPMUnmarshal unpacks a U16Bytes
func (b *U16Bytes) TPMUnmarshal(in io.Reader) error {
	var tmpSize uint16
	if err := binary.Read(in, binary.BigEndian, &tmpSize); err != nil {
		return err
	}
	size := int(tmpSize)

	if len(*b) >= size {
		*b = (*b)[:size]
	} else {
		*b = append(*b, make([]byte, size-len(*b))...)
	}

	n, err := in.Read(*b)
	if err != nil {
		return err
	}
	if n != size {
		return io.ErrUnexpectedEOF
	}
	return nil
}

// U32Bytes is a byte slice with a 32-bit header
type U32Bytes []byte

// TPMMarshal packs U32Bytes
func (b *U32Bytes) TPMMarshal(out io.Writer) error {
	size := len([]byte(*b))
	if err := binary.Write(out, binary.BigEndian, uint32(size)); err != nil {
		return err
	}

	n, err := out.Write(*b)
	if err != nil {
		return err
	}
	if n != size {
		return fmt.Errorf("unable to write all contents of U32Bytes")
	}
	return nil
}

// TPMUnmarshal unpacks a U32Bytes
func (b *U32Bytes) TPMUnmarshal(in io.Reader) error {
	var tmpSize uint32
	if err := binary.Read(in, binary.BigEndian, &tmpSize); err != nil {
		return err
	}

	if tmpSize > maxBytesBufferSize {
		return bytes.ErrTooLarge
	}
	// We can now safely cast to an int on 32-bit or 64-bit machines
	size := int(tmpSize)

	if len(*b) >= size {
		*b = (*b)[:size]
	} else {
		*b = append(*b, make([]byte, size-len(*b))...)
	}

	n, err := in.Read(*b)
	if err != nil {
		return err
	}
	if n != size {
		return fmt.Errorf("unable to read all contents in to U32Bytes")
	}
	return nil
}

// Tag is a command tag.
type Tag uint16

// Command is an identifier of a TPM command.
type Command uint32

// A commandHeader is the header for a TPM command.
type commandHeader struct {
	Tag  Tag
	Size uint32
	Cmd  Command
}

// ResponseCode is a response code returned by TPM.
type ResponseCode uint32

// RCSuccess is response code for successful command. Identical for TPM 1.2 and
// 2.0.
const RCSuccess ResponseCode = 0x000

// RCRetry is response code for TPM is busy.
const RCRetry ResponseCode = 0x922

// A responseHeader is a header for TPM responses.
type responseHeader struct {
	Tag  Tag
	Size uint32
	Res  ResponseCode
}

// A Handle is a reference to a TPM object.
type Handle uint32

// HandleValue returns the handle value. This behavior is intended to satisfy
// an interface that can be implemented by other, more complex types as well.
func (h Handle) HandleValue() uint32 {
	return uint32(h)
}

type handleList []Handle

func (l *handleList) TPMMarshal(_ io.Writer) error {
	return fmt.Errorf("TPMMarhsal on []Handle is not supported yet")
}

func (l *handleList) TPMUnmarshal(in io.Reader) error {
	var numHandles uint16
	if err := binary.Read(in, binary.BigEndian, &numHandles); err != nil {
		return err
	}

	// Make len(e) match size exactly.
	size := int(numHandles)
	if len(*l) >= size {
		*l = (*l)[:size]
	} else {
		*l = append(*l, make([]Handle, size-len(*l))...)
	}
	return binary.Read(in, binary.BigEndian, *l)
}

// SelfMarshaler allows custom types to override default encoding/decoding
// behavior in Pack, Unpack and UnpackBuf.
type SelfMarshaler interface {
	TPMMarshal(out io.Writer) error
	TPMUnmarshal(in io.Reader) error
}
