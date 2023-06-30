//go:build !windows

// Copyright (c) 2019, Google LLC All rights reserved.
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

package tpm2

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/google/go-tpm/tpmutil"
)

// OpenTPM opens a channel to the TPM at the given path. If the file is a
// device, then it treats it like a normal TPM device, and if the file is a
// Unix domain socket, then it opens a connection to the socket.
//
// This function may also be invoked with no paths, as tpm2.OpenTPM(). In this
// case, the default paths on Linux (/dev/tpmrm0 then /dev/tpm0), will be used.
func OpenTPM(path ...string) (tpm io.ReadWriteCloser, err error) {
	switch len(path) {
	case 0:
		tpm, err = tpmutil.OpenTPM("/dev/tpmrm0")
		if errors.Is(err, os.ErrNotExist) {
			tpm, err = tpmutil.OpenTPM("/dev/tpm0")
		}
	case 1:
		tpm, err = tpmutil.OpenTPM(path[0])
	default:
		return nil, errors.New("cannot specify multiple paths to tpm2.OpenTPM")
	}
	if err != nil {
		return nil, err
	}

	// Make sure this is a TPM 2.0
	_, err = GetManufacturer(tpm)
	if err != nil {
		tpm.Close()
		return nil, fmt.Errorf("open %s: device is not a TPM 2.0", path)
	}
	return tpm, nil
}
