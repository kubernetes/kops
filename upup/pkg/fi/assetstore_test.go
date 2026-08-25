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

package fi

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsTarGzip(t *testing.T) {
	var archive bytes.Buffer
	gz := gzip.NewWriter(&archive)
	tw := tar.NewWriter(gz)
	content := []byte("tool")
	if err := tw.WriteHeader(&tar.Header{Name: "bin/tool", Mode: 0o755, Size: int64(len(content))}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}

	for _, test := range []struct {
		name string
		data []byte
		want bool
	}{
		{name: "tar gzip", data: archive.Bytes(), want: true},
		{name: "plain file", data: []byte("not an archive")},
	} {
		t.Run(test.name, func(t *testing.T) {
			filename := filepath.Join(t.TempDir(), "blob")
			if err := os.WriteFile(filename, test.data, 0o600); err != nil {
				t.Fatal(err)
			}
			got, err := isTarGzip(filename)
			if err != nil {
				t.Fatal(err)
			}
			if got != test.want {
				t.Errorf("isTarGzip() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestAssetIdentity(t *testing.T) {
	for _, test := range []struct {
		name     string
		location string
		wantKey  string
		wantPath string
	}{
		{
			name:     "HTTP",
			location: "https://example.com/assets/tool.tgz",
			wantKey:  "tool.tgz",
			wantPath: "https://example.com/assets/tool.tgz",
		},
		{
			name:     "OCI",
			location: "oci://registry.example.com/optional-prefix/runc:v1.3.5-arm64",
			wantKey:  "runc",
			wantPath: "/runc",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			key, assetPath, err := assetIdentity(test.location)
			if err != nil {
				t.Fatal(err)
			}
			if key != test.wantKey || assetPath != test.wantPath {
				t.Errorf("assetIdentity() = %q, %q, want %q, %q", key, assetPath, test.wantKey, test.wantPath)
			}
		})
	}
}

func TestAssetStoreRejectsUntaggedOCIAsset(t *testing.T) {
	id := strings.Repeat("0", 64) + "@oci://registry.example.com/kubelet"
	err := NewAssetStore(t.TempDir()).Add(context.Background(), id)
	if err == nil || !strings.Contains(err.Error(), "must include a tag") {
		t.Fatalf("Add() error = %v", err)
	}
}
