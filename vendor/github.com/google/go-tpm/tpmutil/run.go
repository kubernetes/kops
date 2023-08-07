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

// Package tpmutil provides common utility functions for both TPM 1.2 and TPM
// 2.0 devices.
package tpmutil

import (
	"errors"
	"io"
	"os"
	"time"
)

// maxTPMResponse is the largest possible response from the TPM. We need to know
// this because we don't always know the length of the TPM response, and
// /dev/tpm insists on giving it all back in a single value rather than
// returning a header and a body in separate responses.
const maxTPMResponse = 4096

// RunCommandRaw executes the given raw command and returns the raw response.
// Does not check the response code except to execute retry logic.
func RunCommandRaw(rw io.ReadWriter, inb []byte) ([]byte, error) {
	if rw == nil {
		return nil, errors.New("nil TPM handle")
	}

	// f(t) = (2^t)ms, up to 2s
	var backoffFac uint
	var rh responseHeader
	var outb []byte

	for {
		if _, err := rw.Write(inb); err != nil {
			return nil, err
		}

		// If the TPM is a real device, it may not be ready for reading
		// immediately after writing the command. Wait until the file
		// descriptor is ready to be read from.
		if f, ok := rw.(*os.File); ok {
			if err := poll(f); err != nil {
				return nil, err
			}
		}

		outb = make([]byte, maxTPMResponse)
		outlen, err := rw.Read(outb)
		if err != nil {
			return nil, err
		}
		// Resize the buffer to match the amount read from the TPM.
		outb = outb[:outlen]

		_, err = Unpack(outb, &rh)
		if err != nil {
			return nil, err
		}

		// If TPM is busy, retry the command after waiting a few ms.
		if rh.Res == RCRetry {
			if backoffFac < 11 {
				dur := (1 << backoffFac) * time.Millisecond
				time.Sleep(dur)
				backoffFac++
			} else {
				return nil, err
			}
		} else {
			break
		}
	}

	return outb, nil
}

// RunCommand executes cmd with given tag and arguments. Returns TPM response
// body (without response header) and response code from the header. Returned
// error may be nil if response code is not RCSuccess; caller should check
// both.
func RunCommand(rw io.ReadWriter, tag Tag, cmd Command, in ...interface{}) ([]byte, ResponseCode, error) {
	inb, err := packWithHeader(commandHeader{tag, 0, cmd}, in...)
	if err != nil {
		return nil, 0, err
	}

	outb, err := RunCommandRaw(rw, inb)
	if err != nil {
		return nil, 0, err
	}

	var rh responseHeader
	read, err := Unpack(outb, &rh)
	if err != nil {
		return nil, 0, err
	}
	if rh.Res != RCSuccess {
		return nil, rh.Res, nil
	}

	return outb[read:], rh.Res, nil
}
