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
	"io"

	"github.com/google/go-tpm/tpmutil/tbs"
)

// winTPMBuffer is a ReadWriteCloser to access the TPM in Windows.
type winTPMBuffer struct {
	context   tbs.Context
	outBuffer []byte
}

// Executes the TPM command specified by commandBuffer (at Normal Priority), returning the number
// of bytes in the command and any error code returned by executing the TPM command. Command
// response can be read by calling Read().
func (rwc *winTPMBuffer) Write(commandBuffer []byte) (int, error) {
	// TPM spec defines longest possible response to be maxTPMResponse.
	rwc.outBuffer = rwc.outBuffer[:maxTPMResponse]

	outBufferLen, err := rwc.context.SubmitCommand(
		tbs.NormalPriority,
		commandBuffer,
		rwc.outBuffer,
	)

	if err != nil {
		rwc.outBuffer = rwc.outBuffer[:0]
		return 0, err
	}
	// Shrink outBuffer so it is length of response.
	rwc.outBuffer = rwc.outBuffer[:outBufferLen]
	return len(commandBuffer), nil
}

// Provides TPM response from the command called in the last Write call.
func (rwc *winTPMBuffer) Read(responseBuffer []byte) (int, error) {
	if len(rwc.outBuffer) == 0 {
		return 0, io.EOF
	}
	lenCopied := copy(responseBuffer, rwc.outBuffer)
	// Cut out the piece of slice which was just read out, maintaining original slice capacity.
	rwc.outBuffer = append(rwc.outBuffer[:0], rwc.outBuffer[lenCopied:]...)
	return lenCopied, nil
}

func (rwc *winTPMBuffer) Close() error {
	return rwc.context.Close()
}

// OpenTPM creates a new instance of a ReadWriteCloser which can interact with a
// Windows TPM.
func OpenTPM() (io.ReadWriteCloser, error) {
	tpmContext, err := tbs.CreateContext(tbs.TPMVersion20, tbs.IncludeTPM12|tbs.IncludeTPM20)
	rwc := &winTPMBuffer{
		context:   tpmContext,
		outBuffer: make([]byte, 0, maxTPMResponse),
	}
	return rwc, err
}

// FromContext creates a new instance of a ReadWriteCloser which can
// interact with a Windows TPM, using the specified TBS handle.
func FromContext(ctx tbs.Context) io.ReadWriteCloser {
	return &winTPMBuffer{
		context:   ctx,
		outBuffer: make([]byte, 0, maxTPMResponse),
	}
}
