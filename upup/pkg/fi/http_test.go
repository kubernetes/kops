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
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"k8s.io/kops/util/pkg/hashing"
)

func TestDownloadURLRejectsNon2xxAndPreservesDestination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusFound)
		_, _ = w.Write([]byte("redirect body"))
	}))
	defer server.Close()

	dest := filepath.Join(t.TempDir(), "download")
	if err := os.WriteFile(dest, []byte("original"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if _, err := DownloadURL(server.URL, dest, nil); err == nil {
		t.Fatalf("DownloadURL() expected error")
	}

	actual, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(actual) != "original" {
		t.Fatalf("download destination = %q, expected original contents", actual)
	}
}

func TestDownloadURLToWriterVerifiesHash(t *testing.T) {
	body := []byte("payload")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(body)
	}))
	defer server.Close()

	expectedHash, err := hashing.HashAlgorithmSHA256.Hash(bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	var output bytes.Buffer
	actualHash, err := downloadURLToWriter(context.TODO(), server.URL, &output, expectedHash)
	if err != nil {
		t.Fatalf("downloadURLToWriter() error = %v", err)
	}
	if !actualHash.Equal(expectedHash) {
		t.Fatalf("downloadURLToWriter() hash = %v, expected %v", actualHash, expectedHash)
	}
	if !bytes.Equal(output.Bytes(), body) {
		t.Fatalf("downloadURLToWriter() body = %q, expected %q", output.Bytes(), body)
	}
}

func TestDownloadURLToWriterRejectsHashMismatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("payload"))
	}))
	defer server.Close()

	wrongHash, err := hashing.HashAlgorithmSHA256.Hash(bytes.NewReader([]byte("different")))
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	var output bytes.Buffer
	if _, err := downloadURLToWriter(context.TODO(), server.URL, &output, wrongHash); err == nil {
		t.Fatalf("downloadURLToWriter() expected hash mismatch error")
	}
}
