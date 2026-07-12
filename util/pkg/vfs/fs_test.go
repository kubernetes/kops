/*
Copyright 2020 The Kubernetes Authors.

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

package vfs

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"k8s.io/kops/pkg/testutils/testcontext"
)

func TestCreateFile(t *testing.T) {
	ctx := testcontext.ForTest(t)
	TempDir := t.TempDir()

	tests := []struct {
		path string
		data []byte
	}{
		{
			path: filepath.Join(TempDir, "SubDir", "test1.tmp"),
			data: []byte("test data\nline 1\r\nline 2"),
		},
	}
	for _, test := range tests {
		fspath := &FSPath{test.path}
		// Create file
		err := fspath.CreateFile(ctx, bytes.NewReader(test.data), nil)
		if err != nil {
			t.Fatalf("Error writing file %s, error: %v", test.path, err)
		}

		// Create file again should result in error
		err = fspath.CreateFile(ctx, bytes.NewReader([]byte("data")), nil)
		if err != os.ErrExist {
			t.Errorf("Expected to get os.ErrExist, got: %v", err)
		}

		// Check file content
		data, err := fspath.ReadFile(ctx)
		if err != nil {
			t.Errorf("Error reading file %s, error: %v", test.path, err)
		}
		if !bytes.Equal(data, test.data) {
			t.Errorf("Expected file content %v, got %v", test.data, data)
		}
	}
}

func TestFSPathJoin(t *testing.T) {
	tests := []struct {
		base         string
		relativePath []string
		expectedPath string
		expectedBase string
	}{
		{
			base:         filepath.Join("state", "cluster.example.com"),
			relativePath: []string{"config"},
			expectedPath: filepath.Join("state", "cluster.example.com", "config"),
			expectedBase: "config",
		},
		{
			base:         filepath.Join("state", "cluster.example.com"),
			relativePath: []string{"secrets", "admin"},
			expectedPath: filepath.Join("state", "cluster.example.com", "secrets", "admin"),
			expectedBase: "admin",
		},
		{
			// Callers may pass slash-separated relative paths; filepath.Join normalizes them to
			// the OS path separator.
			base:         "state",
			relativePath: []string{"pki/private/ca"},
			expectedPath: filepath.Join("state", "pki", "private", "ca"),
			expectedBase: "ca",
		},
	}
	for _, test := range tests {
		joined := NewFSPath(test.base).Join(test.relativePath...)
		if joined.Path() != test.expectedPath {
			t.Errorf("Join(%q, %v): expected path %q, got %q", test.base, test.relativePath, test.expectedPath, joined.Path())
		}
		if joined.Base() != test.expectedBase {
			t.Errorf("Join(%q, %v): expected base %q, got %q", test.base, test.relativePath, test.expectedBase, joined.Base())
		}
	}
}

func TestWriteTo(t *testing.T) {
	ctx := testcontext.ForTest(t)

	TempDir := t.TempDir()

	tests := []struct {
		path string
		data []byte
	}{
		{
			path: filepath.Join(TempDir, "SubDir", "test1.tmp"),
			data: []byte("test data\nline 1\r\nline 2"),
		},
	}
	for _, test := range tests {
		var buf bytes.Buffer

		fspath := NewFSPath(test.path)
		// Create file
		err := fspath.CreateFile(ctx, bytes.NewReader(test.data), nil)
		if err != nil {
			t.Fatalf("Error writing file %s, error: %v", test.path, err)
		}

		// Write file to buf
		_, err = fspath.WriteTo(&buf)
		if err != nil {
			t.Fatalf("Error reading %s to buf, error: %v", test.path, err)
		}

		// Check buf content
		actually_bytes := buf.Bytes()
		if !bytes.Equal(test.data, actually_bytes) {
			t.Errorf("Expected %v, actually %v", test.data, actually_bytes)
		}
	}
}
