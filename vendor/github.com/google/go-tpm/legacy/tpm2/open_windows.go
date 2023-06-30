//go:build windows

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

package tpm2

import (
	"fmt"
	"io"

	"github.com/google/go-tpm/tpmutil"
	"github.com/google/go-tpm/tpmutil/tbs"
)

// OpenTPM opens a channel to the TPM.
func OpenTPM() (io.ReadWriteCloser, error) {
	info, err := tbs.GetDeviceInfo()
	if err != nil {
		return nil, err
	}

	if info.TPMVersion != tbs.TPMVersion20 {
		return nil, fmt.Errorf("openTPM: device is not a TPM 2.0")
	}

	return tpmutil.OpenTPM()
}
