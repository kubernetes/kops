/*
Copyright 2019 The Kubernetes Authors.

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

package utils

import (
	"compress/gzip"
	"io"
	"os"
)

// UngzipFile extracts a .gzip file
func UngzipFile(gzipPath string, destPath string) error {
	reader, err := os.Open(gzipPath)
	if err != nil {
		return err
	}
	defer reader.Close()

	writer, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer writer.Close()

	archive, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}
	defer archive.Close()

	_, err = io.Copy(writer, archive)
	return err
}
