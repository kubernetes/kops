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

package assetcopy

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/registry"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/static"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"k8s.io/kops/util/pkg/vfs"
)

func TestCopyFileToOCI_RoundTrip(t *testing.T) {
	server := httptest.NewServer(registry.New(registry.Logger(log.New(io.Discard, "", 0))))
	defer server.Close()
	host := strings.TrimPrefix(server.URL, "http://")

	content := []byte("test file asset content")
	digest := sha256.Sum256(content)
	sha := hex.EncodeToString(digest[:])

	sourceFile := filepath.Join(t.TempDir(), "nodeup")
	if err := os.WriteFile(sourceFile, content, 0o600); err != nil {
		t.Fatalf("writing source file: %v", err)
	}

	task := &CopyFileToOCI{
		Name:       sourceFile,
		SourceFile: sourceFile,
		TargetRef:  "oci://" + host + "/assets/binaries/nodeup",
		SHA:        sha,
		VFSContext: vfs.Context,
	}
	if err := task.Run(); err != nil {
		t.Fatalf("pushing file to the registry: %v", err)
	}

	// Nodes download the layer blob directly, addressed by the file's hash.
	response, err := http.Get(server.URL + "/v2/assets/binaries/nodeup/blobs/sha256:" + sha)
	if err != nil {
		t.Fatalf("downloading blob: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("unexpected response downloading blob: HTTP %s", response.Status)
	}
	downloaded, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("reading blob: %v", err)
	}
	if !bytes.Equal(downloaded, content) {
		t.Errorf("downloaded blob does not match the source file")
	}

	// A second run finds the already-pushed artifact and is a no-op.
	if err := task.Run(); err != nil {
		t.Fatalf("re-pushing file to the registry: %v", err)
	}
}

func TestCopyFileToOCI_RejectsNormalizingReference(t *testing.T) {
	// A registry host without a dot or port would be normalized to a Docker Hub repository on push,
	// while nodes would pull from that literal host.
	task := &CopyFileToOCI{
		TargetRef: "oci://docker.io/assets",
		SHA:       strings.Repeat("0", 64),
	}
	err := task.Run()
	if err == nil || !strings.Contains(err.Error(), "is not the location nodes download from") {
		t.Fatalf("expected a reference mismatch error, got: %v", err)
	}
}

func TestCopyFileToOCI_OverwritesStaleTag(t *testing.T) {
	server := httptest.NewServer(registry.New(registry.Logger(log.New(io.Discard, "", 0))))
	defer server.Close()
	host := strings.TrimPrefix(server.URL, "http://")

	content := []byte("test file asset content")
	digest := sha256.Sum256(content)
	sha := hex.EncodeToString(digest[:])

	sourceFile := filepath.Join(t.TempDir(), "nodeup")
	if err := os.WriteFile(sourceFile, content, 0o600); err != nil {
		t.Fatalf("writing source file: %v", err)
	}

	// Pre-push different content under the tag the copy task uses.
	staleRef, err := name.NewTag(host + "/assets/binaries/nodeup:" + sha)
	if err != nil {
		t.Fatalf("parsing stale reference: %v", err)
	}
	staleLayer := static.NewLayer([]byte("stale content"), types.MediaType("application/octet-stream"))
	staleImage, err := mutate.AppendLayers(empty.Image, staleLayer)
	if err != nil {
		t.Fatalf("building stale artifact: %v", err)
	}
	if err := remote.Write(staleRef, staleImage); err != nil {
		t.Fatalf("pushing stale artifact: %v", err)
	}

	task := &CopyFileToOCI{
		Name:       sourceFile,
		SourceFile: sourceFile,
		TargetRef:  "oci://" + host + "/assets/binaries/nodeup",
		SHA:        sha,
		VFSContext: vfs.Context,
	}
	if err := task.Run(); err != nil {
		t.Fatalf("pushing file to the registry: %v", err)
	}

	// The stale tag must not have been trusted; the expected blob must now exist.
	response, err := http.Get(server.URL + "/v2/assets/binaries/nodeup/blobs/sha256:" + sha)
	if err != nil {
		t.Fatalf("downloading blob: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("unexpected response downloading blob: HTTP %s", response.Status)
	}
}
