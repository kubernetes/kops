// Copyright 2018 Google LLC All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package layout

import (
	"fmt"
	"io"
	"os"

	v1 "github.com/google/go-containerregistry/pkg/v1"
)

// Blob returns a blob with the given hash from the Path.
func (l Path) Blob(h v1.Hash) (io.ReadCloser, error) {
	return l.openBlob(h)
}

// Bytes is a convenience function to return a blob from the Path as
// a byte slice.
func (l Path) Bytes(h v1.Hash) ([]byte, error) {
	f, err := l.openBlob(h)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}

func (l Path) blobPath(h v1.Hash) string {
	return l.path("blobs", h.Algorithm, h.Hex)
}

func (l Path) openBlob(h v1.Hash) (*os.File, error) {
	p := l.blobPath(h)
	info, err := os.Lstat(p)
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("layout blob %s is a symlink", h)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("layout blob %s is not a regular file", h)
	}
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	closeFile := true
	defer func() {
		if closeFile {
			f.Close()
		}
	}()
	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !stat.Mode().IsRegular() {
		return nil, fmt.Errorf("layout blob %s is not a regular file", h)
	}
	if !os.SameFile(info, stat) {
		return nil, fmt.Errorf("layout blob %s changed while opening", h)
	}
	closeFile = false
	return f, nil
}
