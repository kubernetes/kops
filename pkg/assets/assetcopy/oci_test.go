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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/util/pkg/hashing"
	"k8s.io/kops/util/pkg/vfs"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestHTTPSOnlyRegistryTransport(t *testing.T) {
	called := false
	transport := &httpsOnlyRegistryTransport{inner: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		called = true
		return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}, nil
	})}

	request, err := http.NewRequest(http.MethodGet, "http://localhost:5000/v2/", nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := transport.RoundTrip(request); err == nil || !strings.Contains(err.Error(), "must use HTTPS") {
		t.Fatalf("RoundTrip() error = %v, want HTTPS error", err)
	}
	if called {
		t.Fatal("HTTP request reached the registry transport")
	}

	request, err = http.NewRequest(http.MethodGet, "https://registry.example.com/v2/", nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := transport.RoundTrip(request); err != nil {
		t.Fatalf("RoundTrip() error = %v", err)
	}
	if !called {
		t.Fatal("HTTPS request did not reach the registry transport")
	}
}

func TestCopyOCIAsset(t *testing.T) {
	content := []byte("exact source bytes")
	source := filepath.Join(t.TempDir(), "asset")
	if err := os.WriteFile(source, content, 0o600); err != nil {
		t.Fatal(err)
	}
	hash, err := hashing.HashAlgorithmSHA256.Hash(bytes.NewReader(content))
	if err != nil {
		t.Fatal(err)
	}
	target, err := url.Parse("oci://registry.example.com/prefix/tool:v1.2.3-amd64")
	if err != nil {
		t.Fatal(err)
	}
	semanticTag := target.Host + target.Path

	originalGet, originalWrite := ociGet, ociWrite
	t.Cleanup(func() { ociGet, ociWrite = originalGet, originalWrite })
	var descriptor *remote.Descriptor
	writes := 0
	ociGet = func(ref name.Reference, _ ...remote.Option) (*remote.Descriptor, error) {
		if descriptor != nil {
			return descriptor, nil
		}
		return nil, &transport.Error{StatusCode: http.StatusNotFound}
	}
	ociWrite = func(ref name.Reference, image v1.Image, _ ...remote.Option) error {
		writes++
		if ref.Name() != semanticTag {
			return fmt.Errorf("unexpected tag %s", ref.Name())
		}
		layers, err := image.Layers()
		if err != nil {
			return err
		}
		if len(layers) != 1 {
			return fmt.Errorf("Layers() = %d, want 1", len(layers))
		}
		reader, err := layers[0].Compressed()
		if err != nil {
			return err
		}
		got, err := io.ReadAll(reader)
		_ = reader.Close()
		if err != nil {
			return err
		}
		if !bytes.Equal(got, content) {
			return fmt.Errorf("layer content = %q, want %q", got, content)
		}
		manifest, err := image.Manifest()
		if err != nil {
			return err
		}
		if manifest.MediaType != types.OCIManifestSchema1 || manifest.Config.MediaType != types.OCIConfigJSON {
			return fmt.Errorf("manifest is not OCI: %#v", manifest)
		}
		descriptor = descriptorForLayer(t, manifest.Layers[0].Digest.Hex)
		return nil
	}

	fileAsset := &assets.FileAsset{
		CanonicalURL: &url.URL{Path: source},
		DownloadURL:  target,
		SHAValue:     hash,
	}
	if err := Copy(nil, []*assets.FileAsset{fileAsset}, vfs.Context, nil); err != nil {
		t.Fatalf("Copy() error = %v", err)
	}
	if writes != 1 {
		t.Fatalf("writes = %d, want 1", writes)
	}
}

func TestCopyOCIAssetIsIdempotent(t *testing.T) {
	hash := strings.Repeat("a", 64)
	originalGet, originalWrite := ociGet, ociWrite
	t.Cleanup(func() { ociGet, ociWrite = originalGet, originalWrite })
	ociGet = func(_ name.Reference, _ ...remote.Option) (*remote.Descriptor, error) {
		return descriptorForLayer(t, hash), nil
	}
	ociWrite = func(ref name.Reference, _ v1.Image, _ ...remote.Option) error {
		t.Fatalf("unexpected write of existing tag %s", ref.Name())
		return nil
	}

	task := &copyOCIAsset{
		source: filepath.Join(t.TempDir(), "does-not-exist"),
		sha256: hash,
		target: "oci://registry.example.com/prefix/tool:v1.2.3-amd64",
		vfs:    vfs.Context,
	}
	if err := task.Run(); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestCopyOCIAssetRejectsTagCollision(t *testing.T) {
	hash := strings.Repeat("a", 64)
	originalGet, originalWrite := ociGet, ociWrite
	t.Cleanup(func() { ociGet, ociWrite = originalGet, originalWrite })
	semanticTag := "registry.example.com/tool:v1-amd64"
	ociGet = func(ref name.Reference, _ ...remote.Option) (*remote.Descriptor, error) {
		if ref.Name() != semanticTag {
			t.Fatalf("unexpected tag read: %s", ref.Name())
		}
		return descriptorForLayer(t, strings.Repeat("f", 64)), nil
	}
	ociWrite = func(ref name.Reference, _ v1.Image, _ ...remote.Option) error {
		t.Fatalf("conflicting tag was overwritten: %s", ref.Name())
		return nil
	}

	task := &copyOCIAsset{source: filepath.Join(t.TempDir(), "does-not-exist"), sha256: hash, target: "oci://registry.example.com/tool:v1-amd64", vfs: vfs.Context}
	err := task.Run()
	for _, want := range []string{
		semanticTag,
		"sha256:" + strings.Repeat("f", 64),
		"expected sha256:" + hash,
		"distinct asset version",
		"remove the conflicting tag only after verifying that it is no longer needed",
	} {
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Fatalf("Run() error = %v, want %q", err, want)
		}
	}
}

func TestCopyOCIAssetReportsConcurrentTagChange(t *testing.T) {
	content := []byte("concurrent source")
	hash, err := hashing.HashAlgorithmSHA256.Hash(bytes.NewReader(content))
	if err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(t.TempDir(), "asset")
	if err := os.WriteFile(source, content, 0o600); err != nil {
		t.Fatal(err)
	}

	originalGet, originalWrite := ociGet, ociWrite
	t.Cleanup(func() { ociGet, ociWrite = originalGet, originalWrite })
	semanticTag := "registry.example.com/tool:v1-amd64"
	semanticWritten := false
	ociGet = func(ref name.Reference, _ ...remote.Option) (*remote.Descriptor, error) {
		if ref.Name() != semanticTag {
			t.Fatalf("unexpected tag read: %s", ref.Name())
		}
		if semanticWritten {
			return descriptorForLayer(t, strings.Repeat("f", 64)), nil
		}
		return nil, &transport.Error{StatusCode: http.StatusNotFound}
	}
	ociWrite = func(ref name.Reference, _ v1.Image, _ ...remote.Option) error {
		if ref.Name() != semanticTag {
			t.Fatalf("unexpected tag write: %s", ref.Name())
		}
		semanticWritten = true
		return nil
	}

	task := &copyOCIAsset{source: source, sha256: hash.Hex(), target: "oci://" + semanticTag, vfs: vfs.Context}
	err = task.Run()
	if err == nil || !strings.Contains(err.Error(), semanticTag) ||
		!strings.Contains(err.Error(), "sha256:"+strings.Repeat("f", 64)) ||
		!strings.Contains(err.Error(), "expected sha256:"+hash.Hex()) {
		t.Fatalf("Run() error = %v, want concurrent tag collision with both digests", err)
	}
}

func descriptorForLayer(t *testing.T, digest string) *remote.Descriptor {
	t.Helper()
	manifest := v1.Manifest{
		SchemaVersion: 2,
		MediaType:     types.OCIManifestSchema1,
		Config: v1.Descriptor{
			MediaType: types.OCIConfigJSON,
			Digest:    v1.Hash{Algorithm: "sha256", Hex: strings.Repeat("0", 64)},
			Size:      2,
		},
		Layers: []v1.Descriptor{{
			MediaType: types.MediaType("application/octet-stream"),
			Digest:    v1.Hash{Algorithm: "sha256", Hex: digest},
			Size:      1,
		}},
	}
	b, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	return &remote.Descriptor{
		Descriptor: v1.Descriptor{MediaType: types.OCIManifestSchema1},
		Manifest:   b,
	}
}
