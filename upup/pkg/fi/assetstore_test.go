/*
Copyright The Kubernetes Authors.

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

package fi

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/klauspost/compress/zstd"
)

func TestAddArchive(t *testing.T) {
	// The layout of a gVisor release tarball: binaries at the top level and
	// sidecar binaries in a subdirectory.
	files := map[string]string{
		"runsc":                    "runsc contents",
		"containerd-shim-runsc-v1": "shim contents",
		"gvisor-bin/gvisor_sentry": "sentry contents",
	}

	tests := []struct {
		name        string
		compression archiveCompression
		compress    func(io.Writer) (io.WriteCloser, error)
	}{
		{
			name:        "gzip",
			compression: archiveGzip,
			compress: func(w io.Writer) (io.WriteCloser, error) {
				return gzip.NewWriter(w), nil
			},
		},
		{
			name:        "zstd",
			compression: archiveZstd,
			compress: func(w io.Writer) (io.WriteCloser, error) {
				return zstd.NewWriter(w)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dir := t.TempDir()
			archiveFile := filepath.Join(dir, "archive")
			writeTestArchive(t, archiveFile, files, test.compress)

			store := NewAssetStore(dir)
			source := &Source{URL: "https://example.com/archive"}
			if err := store.addArchive(source, archiveFile, test.compression); err != nil {
				t.Fatalf("addArchive: %v", err)
			}

			for assetPath, want := range files {
				matches := store.FindMatches(regexp.MustCompile("^" + regexp.QuoteMeta(assetPath) + "$"))
				if len(matches) != 1 {
					t.Fatalf("found %d assets for %q, want 1", len(matches), assetPath)
				}
				for _, resource := range matches {
					got, err := ResourceAsString(resource)
					if err != nil {
						t.Fatalf("reading %q: %v", assetPath, err)
					}
					if got != want {
						t.Errorf("%q contains %q, want %q", assetPath, got, want)
					}
				}
			}
		})
	}
}

func writeTestArchive(t *testing.T, archiveFile string, files map[string]string, compress func(io.Writer) (io.WriteCloser, error)) {
	t.Helper()

	f, err := os.Create(archiveFile)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	cw, err := compress(f)
	if err != nil {
		t.Fatal(err)
	}
	tw := tar.NewWriter(cw)

	if err := tw.WriteHeader(&tar.Header{Name: "gvisor-bin/", Typeflag: tar.TypeDir, Mode: 0o755}); err != nil {
		t.Fatal(err)
	}
	for name, contents := range files {
		if err := tw.WriteHeader(&tar.Header{Name: name, Typeflag: tar.TypeReg, Mode: 0o755, Size: int64(len(contents))}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(contents)); err != nil {
			t.Fatal(err)
		}
	}

	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := cw.Close(); err != nil {
		t.Fatal(err)
	}
}
