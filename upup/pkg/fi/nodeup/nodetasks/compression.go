/*
Copyright 2026 The Kubernetes Authors.

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

package nodetasks

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"io"
)

var gzipMagic = []byte{0x1f, 0x8b, 0x08}

func maybeGzipReader(r io.Reader) (io.Reader, io.Closer, error) {
	buffered := bufio.NewReader(r)
	header, err := buffered.Peek(len(gzipMagic))
	if err != nil && err != io.EOF {
		return nil, nil, err
	}
	if len(header) == len(gzipMagic) && bytes.Equal(header, gzipMagic) {
		gzipReader, err := gzip.NewReader(buffered)
		if err != nil {
			return nil, nil, err
		}
		return gzipReader, gzipReader, nil
	}
	return buffered, nil, nil
}
