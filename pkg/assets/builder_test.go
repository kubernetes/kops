/*
Copyright 2017 The Kubernetes Authors.

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

package assets

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/testutils/golden"
	"k8s.io/kops/util/pkg/hashing"
)

func buildAssetBuilder(t *testing.T) *AssetBuilder {
	return NewAssetBuilder(nil, &kops.AssetsSpec{}, false)
}

func TestValidate_RemapImage_ContainerProxy_AppliesToDockerHub(t *testing.T) {
	builder := buildAssetBuilder(t)

	proxyURL := "proxy.example.com/"
	image := "weaveworks/weave-kube"
	expected := "proxy.example.com/weaveworks/weave-kube"

	builder.assetsLocation.ContainerProxy = &proxyURL

	remapped := builder.RemapImage(image)
	if remapped != expected {
		t.Errorf("Error remapping image (Expecting: %s, got %s)", expected, remapped)
	}
}

func TestValidate_RemapImage_ContainerProxy_AppliesToSimplifiedDockerHub(t *testing.T) {
	builder := buildAssetBuilder(t)

	proxyURL := "proxy.example.com/"
	image := "debian"
	expected := "proxy.example.com/debian"

	builder.assetsLocation.ContainerProxy = &proxyURL

	remapped := builder.RemapImage(image)
	if remapped != expected {
		t.Errorf("Error remapping image (Expecting: %s, got %s)", expected, remapped)
	}
}

func TestValidate_RemapImage_ContainerProxy_AppliesToSimplifiedKubernetesURL(t *testing.T) {
	builder := buildAssetBuilder(t)

	proxyURL := "proxy.example.com/"
	image := "registry.k8s.io/kube-apiserver"
	expected := "proxy.example.com/kube-apiserver"

	builder.assetsLocation.ContainerProxy = &proxyURL

	remapped := builder.RemapImage(image)
	if remapped != expected {
		t.Errorf("Error remapping image (Expecting: %s, got %s)", expected, remapped)
	}
}

func TestValidate_RemapImage_ContainerProxy_AppliesToLegacyKubernetesURL(t *testing.T) {
	builder := buildAssetBuilder(t)

	proxyURL := "proxy.example.com/"
	image := "gcr.io/google_containers/kube-apiserver"
	expected := "proxy.example.com/google_containers/kube-apiserver"

	builder.assetsLocation.ContainerProxy = &proxyURL

	remapped := builder.RemapImage(image)
	if remapped != expected {
		t.Errorf("Error remapping image (Expecting: %s, got %s)", expected, remapped)
	}
}

func TestValidate_RemapImage_ContainerProxy_AppliesToImagesWithTags(t *testing.T) {
	builder := buildAssetBuilder(t)

	proxyURL := "proxy.example.com/"
	image := "registry.k8s.io/kube-apiserver:1.2.3"
	expected := "proxy.example.com/kube-apiserver:1.2.3"

	builder.assetsLocation.ContainerProxy = &proxyURL

	remapped := builder.RemapImage(image)
	if remapped != expected {
		t.Errorf("Error remapping image (Expecting: %s, got %s)", expected, remapped)
	}
}

func TestValidate_RemapImage_ContainerRegistry_MappingMultipleTimesConverges(t *testing.T) {
	builder := buildAssetBuilder(t)

	mirrorURL := "proxy.example.com"
	image := "kube-apiserver:1.2.3"
	expected := "proxy.example.com/kube-apiserver:1.2.3"

	builder.assetsLocation.ContainerRegistry = &mirrorURL

	remapped := image
	iterations := make([]map[int]int, 2)
	for i := range iterations {
		remapped := builder.RemapImage(remapped)
		if remapped != expected {
			t.Errorf("Error remapping image (Expecting: %s, got %s, iteration: %d)", expected, remapped, i)
		}
	}
}

func TestRemapEmptySection(t *testing.T) {
	builder := buildAssetBuilder(t)

	testdir := "testdata"

	key := "emptysection"

	inputPath := filepath.Join(testdir, key+".input.yaml")
	expectedPath := filepath.Join(testdir, key+".expected.yaml")

	input, err := os.ReadFile(inputPath)
	if err != nil {
		t.Errorf("error reading file %q: %v", inputPath, err)
	}

	actual, err := builder.RemapManifest(input)
	if err != nil {
		t.Errorf("error remapping manifest %q: %v", inputPath, err)
	}

	golden.AssertMatchesFile(t, string(actual), expectedPath)
}

func TestAssetBuilderConcurrentCollection(t *testing.T) {
	builder := buildAssetBuilder(t)
	knownHash := hashing.MustFromString("sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")

	originalImageDigestEnabled := featureflag.ImageDigest.Enabled()
	featureflag.ParseFlags("-ImageDigest")
	t.Cleanup(func() {
		if originalImageDigestEnabled {
			featureflag.ParseFlags("ImageDigest")
		} else {
			featureflag.ParseFlags("-ImageDigest")
		}
	})

	const count = 64

	var wg sync.WaitGroup
	wg.Add(count * 4)

	for i := 0; i < count; i++ {
		go func(i int) {
			defer wg.Done()
			builder.RemapImage(fmt.Sprintf("registry.k8s.io/example/image-%d:latest", i))
		}(i)

		go func(i int) {
			defer wg.Done()
			u := fmt.Sprintf("https://example.com/assets/file-%d", i)
			fileURL, err := url.Parse(u)
			if err != nil {
				t.Errorf("error parsing url %q: %v", u, err)
				return
			}
			if _, err := builder.RemapFile(fileURL, knownHash); err != nil {
				t.Errorf("error remapping file %q: %v", u, err)
			}
		}(i)

		go func(i int) {
			defer wg.Done()
			builder.AddStaticManifest(&StaticManifest{
				Key:      fmt.Sprintf("manifest-%d", i),
				Path:     fmt.Sprintf("manifests/static/manifest-%d.yaml", i),
				Contents: []byte(fmt.Sprintf("manifest-%d", i)),
			})
		}(i)

		go func(i int) {
			defer wg.Done()
			builder.AddStaticFile(&StaticFile{
				Path:    fmt.Sprintf("/etc/kubernetes/static-file-%d", i),
				Content: fmt.Sprintf("content-%d", i),
			})
		}(i)
	}

	wg.Wait()

	if got := len(builder.ImageAssets()); got != count {
		t.Fatalf("expected %d image assets, got %d", count, got)
	}
	if got := len(builder.FileAssets()); got != count {
		t.Fatalf("expected %d file assets, got %d", count, got)
	}
	if got := len(builder.StaticManifests()); got != count {
		t.Fatalf("expected %d static manifests, got %d", count, got)
	}
	if got := len(builder.StaticFiles()); got != count {
		t.Fatalf("expected %d static files, got %d", count, got)
	}

	imageAssets := builder.ImageAssets()
	for i := 1; i < len(imageAssets); i++ {
		prev := imageAssets[i-1]
		curr := imageAssets[i]
		if prev.CanonicalLocation > curr.CanonicalLocation {
			t.Fatalf("image assets not sorted by canonical location: %q > %q", prev.CanonicalLocation, curr.CanonicalLocation)
		}
	}

	fileAssets := builder.FileAssets()
	for i := 1; i < len(fileAssets); i++ {
		prev := fileAssets[i-1]
		curr := fileAssets[i]
		prevCanonical := prev.CanonicalURL.String()
		currCanonical := curr.CanonicalURL.String()
		if prevCanonical > currCanonical {
			t.Fatalf("file assets not sorted by canonical url: %q > %q", prevCanonical, currCanonical)
		}
	}

	staticManifests := builder.StaticManifests()
	for i := 1; i < len(staticManifests); i++ {
		prev := staticManifests[i-1]
		curr := staticManifests[i]
		if prev.Key > curr.Key {
			t.Fatalf("static manifests not sorted by key: %q > %q", prev.Key, curr.Key)
		}
	}

	staticFiles := builder.StaticFiles()
	for i := 1; i < len(staticFiles); i++ {
		prev := staticFiles[i-1]
		curr := staticFiles[i]
		if prev.Path > curr.Path {
			t.Fatalf("static files not sorted: %q > %q", prev.Path, curr.Path)
		}
	}
}

func TestRemapFile_OCIRepository(t *testing.T) {
	builder := buildAssetBuilder(t)

	fileRepository := "oci://registry.example.com/assets"
	builder.assetsLocation.FileRepository = &fileRepository

	// CI builds are staged under paths with characters that are not allowed in OCI repository names,
	// such as `+`.
	canonicalURL, err := url.Parse("https://example.com/kops/1.37.0-alpha.2+v1.37.0-alpha.1/linux/amd64/nodeup")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	hash, err := hashing.FromString("833723369ad345a88dd85d61b1e77336d56e61b864557ded71b92b6e34158e6a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	asset, err := builder.RemapFile(canonicalURL, hash)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "oci://registry.example.com/assets/kops/1.37.0-alpha.2_v1.37.0-alpha.1/linux/amd64/nodeup"
	if a := asset.DownloadURL.String(); a != expected {
		t.Errorf("unexpected remapped file (expecting: %s, got %s)", expected, a)
	}
}

func TestSanitizeOCIRepository(t *testing.T) {
	grid := []struct {
		input    string
		expected string
	}{
		{
			input:    "/binaries/kops/1.35.0/linux/amd64/nodeup",
			expected: "binaries/kops/1.35.0/linux/amd64/nodeup",
		},
		{
			input:    "/kops/1.37.0-alpha.2+v1.37.0-alpha.1/linux/amd64/nodeup",
			expected: "kops/1.37.0-alpha.2_v1.37.0-alpha.1/linux/amd64/nodeup",
		},
		{
			// Repository path components cannot start with a separator or contain separator runs, and are
			// lowercased.
			input:    "/kops/+build/Upper/a++b/nodeup",
			expected: "kops/build/upper/a_b/nodeup",
		},
	}

	for _, g := range grid {
		if actual := sanitizeOCIRepository(g.input); actual != g.expected {
			t.Errorf("sanitizeOCIRepository(%q): expected %q, got %q", g.input, g.expected, actual)
		}
	}
}

func TestRemapFile_OCIRequiresSHA256(t *testing.T) {
	builder := buildAssetBuilder(t)

	fileRepository := "oci://registry.example.com/assets"
	builder.assetsLocation.FileRepository = &fileRepository

	canonicalURL, err := url.Parse("https://example.com/kops/nodeup")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sha1Hash, err := hashing.FromString("da39a3ee5e6b4b0d3255bfef95601890afd80709")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = builder.RemapFile(canonicalURL, sha1Hash)
	if err == nil || !strings.Contains(err.Error(), "must have a sha256 hash") {
		t.Fatalf("expected a sha256 requirement error, got: %v", err)
	}
}

func TestRemapFile_OCISharedHost(t *testing.T) {
	builder := buildAssetBuilder(t)

	// The registry shares a host with the asset's https source; the differing scheme must still
	// trigger the remap.
	fileRepository := "oci://example.com/assets"
	builder.assetsLocation.FileRepository = &fileRepository

	canonicalURL, err := url.Parse("https://example.com/kops/nodeup")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	hash, err := hashing.FromString("833723369ad345a88dd85d61b1e77336d56e61b864557ded71b92b6e34158e6a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	asset, err := builder.RemapFile(canonicalURL, hash)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "oci://example.com/assets/kops/nodeup"
	if a := asset.DownloadURL.String(); a != expected {
		t.Errorf("unexpected remapped file (expecting: %s, got %s)", expected, a)
	}
}
