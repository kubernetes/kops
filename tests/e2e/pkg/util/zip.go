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

package util

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// UnzipToTempDir will decompress the provided bytes into a temporary directory that is returned
func UnzipToTempDir(data []byte) (string, error) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", err
	}
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		return "", err
	}

	for _, r := range reader.File {
		fileNamePath := filepath.Join(dir, r.Name)
		if !strings.HasPrefix(fileNamePath, filepath.Clean(dir)+string(os.PathSeparator)) {
			return "", fmt.Errorf("invalid file path: %v", fileNamePath)
		}

		if r.FileInfo().IsDir() {
			os.MkdirAll(fileNamePath, os.ModePerm)
			continue
		}

		if err = os.MkdirAll(filepath.Dir(fileNamePath), os.ModePerm); err != nil {
			return "", err
		}

		output, err := os.OpenFile(fileNamePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, r.Mode())
		if err != nil {
			return "", err
		}

		fileReader, err := r.Open(ctx)
		if err != nil {
			return "", err
		}

		_, err = io.Copy(output, fileReader)

		output.Close()
		fileReader.Close()

		if err != nil {
			return "", err
		}
	}
	return dir, nil
}
