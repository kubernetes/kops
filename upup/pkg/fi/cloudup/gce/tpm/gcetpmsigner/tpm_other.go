//go:build !windows
// +build !windows

/*
Copyright 2021 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package gcetpmsigner

import (
	"fmt"
	"io"

	"github.com/google/go-tpm/tpm2"
)

var tpmPath = "/dev/tpm0"

func openTPM() (io.ReadWriteCloser, error) {
	rw, err := tpm2.OpenTPM(tpmPath)
	if err != nil {
		return nil, fmt.Errorf("tpm2.OpenTPM(%q): %w", tpmPath, err)
	}
	return rw, nil
}
