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

package wellknownassets

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"sync/atomic"
	"testing"

	"k8s.io/kops"
	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/util/pkg/architectures"
	"k8s.io/kops/util/pkg/hashing"
	"k8s.io/kops/util/pkg/vfs"
)

func TestBaseURL_OverridesVersionFromKopsBaseURL(t *testing.T) {
	origVersion := kops.Version
	t.Cleanup(func() {
		kops.Version = origVersion
		kopsBaseURL = nil
	})

	tests := []struct {
		name            string
		kopsBaseURL     string
		expectedVersion string
	}{
		{
			name:            "postsubmit URL",
			kopsBaseURL:     "https://storage.googleapis.com/k8s-staging-kops/kops/releases/1.35.0-beta.2+v1.37.0-beta.1-384-gf369c3ab16",
			expectedVersion: "1.35.0-beta.2+v1.37.0-beta.1-384-gf369c3ab16",
		},
		{
			name:            "postsubmit URL with trailing slash",
			kopsBaseURL:     "https://storage.googleapis.com/k8s-staging-kops/kops/releases/1.35.0-beta.2+v1.37.0-beta.1-384-gf369c3ab16/",
			expectedVersion: "1.35.0-beta.2+v1.37.0-beta.1-384-gf369c3ab16",
		},
		{
			name:            "CI URL",
			kopsBaseURL:     "https://storage.googleapis.com/k8s-staging-kops/kops/ci/1.35.0-beta.2+abc123",
			expectedVersion: "1.35.0-beta.2+abc123",
		},
		{
			name:            "CI URL with trailing slash",
			kopsBaseURL:     "https://storage.googleapis.com/k8s-staging-kops/kops/ci/1.35.0-beta.2+abc123/",
			expectedVersion: "1.35.0-beta.2+abc123",
		},
		{
			name:            "release URL",
			kopsBaseURL:     "https://artifacts.k8s.io/binaries/kops/1.35.0",
			expectedVersion: "1.35.0",
		},
		{
			name:            "same version as binary does not override",
			kopsBaseURL:     fmt.Sprintf("https://example.com/kops/%s", origVersion),
			expectedVersion: origVersion,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			kops.Version = origVersion
			kopsBaseURL = nil
			t.Setenv("KOPS_BASE_URL", tc.kopsBaseURL)

			_, err := BaseURL()
			if err != nil {
				t.Fatalf("BaseURL() error: %v", err)
			}
			if kops.Version != tc.expectedVersion {
				t.Errorf("kops.Version = %q, want %q", kops.Version, tc.expectedVersion)
			}
		})
	}
}

func Test_BuildMirroredAsset(t *testing.T) {
	tests := []struct {
		url      string
		hash     string
		expected []string
	}{
		{
			url: "https://artifacts.k8s.io/binaries/kops/%s/linux/amd64/nodeup",
			expected: []string{
				"https://artifacts.k8s.io/binaries/kops/" + kops.Version + "/linux/amd64/nodeup",
				"https://github.com/kubernetes/kops/releases/download/v" + kops.Version + "/nodeup-linux-amd64",
			},
		},
		{
			url: "https://artifacts.k8s.io/binaries/kops/%s/linux/arm64/nodeup",
			expected: []string{
				"https://artifacts.k8s.io/binaries/kops/" + kops.Version + "/linux/arm64/nodeup",
				"https://github.com/kubernetes/kops/releases/download/v" + kops.Version + "/nodeup-linux-arm64",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.url, func(t *testing.T) {
			h := hashing.MustFromString("0000000000000000000000000000000000000000000000000000000000000000")
			u, err := url.Parse(fmt.Sprintf(tc.url, kops.Version))
			if err != nil {
				t.Errorf("cannot parse URL: %s", fmt.Sprintf(tc.url, kops.Version))
				return
			}
			asset := &assets.FileAsset{
				DownloadURL:  u,
				CanonicalURL: u,
				SHAValue:     h,
			}
			actual := assets.BuildMirroredAsset(asset)

			if !reflect.DeepEqual(actual.Locations, tc.expected) {
				t.Errorf("Locations differ:\nActual: %+v\nExpect: %+v", actual.Locations, tc.expected)
				return
			}
		})
	}
}

func TestNodeUpAssetRegistersWithEveryAssetBuilder(t *testing.T) {
	hashes := map[architectures.Architecture]string{
		architectures.ArchitectureAmd64: "5555555555555555555555555555555555555555555555555555555555555555",
		architectures.ArchitectureArm64: "6666666666666666666666666666666666666666666666666666666666666666",
	}

	var requests atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for arch, hash := range hashes {
			if r.URL.Path == fmt.Sprintf("/kops/%s/linux/%s/nodeup.sha256", kops.Version, arch) {
				requests.Add(1)
				fmt.Fprintf(w, "%s  nodeup\n", hash)
				return
			}
		}
		// The VFS retries 404 and 5xx responses.
		http.Error(w, "not found", http.StatusForbidden)
	}))
	t.Cleanup(server.Close)

	kopsBaseURL = nil
	t.Cleanup(func() {
		kopsBaseURL = nil
	})
	// Keep kops.Version unchanged when BaseURL parses KOPS_BASE_URL.
	t.Setenv("KOPS_BASE_URL", fmt.Sprintf("%s/kops/%s", server.URL, kops.Version))

	vfsContext := vfs.NewVFSContext()

	// Both update and get-assets builders must register nodeup.
	for _, getAssets := range []bool{false, true} {
		t.Run(fmt.Sprintf("getAssets=%t", getAssets), func(t *testing.T) {
			assetBuilder := assets.NewAssetBuilder(vfsContext, &kopsapi.AssetsSpec{}, getAssets)

			for _, arch := range []architectures.Architecture{architectures.ArchitectureAmd64, architectures.ArchitectureArm64} {
				asset, err := NodeUpAsset(assetBuilder, arch)
				if err != nil {
					t.Fatalf("NodeUpAsset(%s) error: %v", arch, err)
				}
				expectedLocation := fmt.Sprintf("%s/kops/%s/linux/%s/nodeup", server.URL, kops.Version, arch)
				if !reflect.DeepEqual(asset.Locations, []string{expectedLocation}) {
					t.Errorf("unexpected nodeup locations for %s: %v", arch, asset.Locations)
				}
				if asset.Hash.Hex() != hashes[arch] {
					t.Errorf("unexpected nodeup hash for %s: actual %q, expected %q", arch, asset.Hash.Hex(), hashes[arch])
				}
			}

			var registered []string
			for _, fileAsset := range assetBuilder.FileAssets() {
				registered = append(registered, fileAsset.CanonicalURL.String())
			}
			expected := []string{
				fmt.Sprintf("%s/kops/%s/linux/amd64/nodeup", server.URL, kops.Version),
				fmt.Sprintf("%s/kops/%s/linux/arm64/nodeup", server.URL, kops.Version),
			}
			if !reflect.DeepEqual(registered, expected) {
				t.Errorf("unexpected registered file assets:\nActual: %v\nExpect: %v", registered, expected)
			}
		})
	}

	// Each architecture's checksum is downloaded once across both builders.
	if actual := requests.Load(); actual != int64(len(hashes)) {
		t.Errorf("expected %d checksum requests, got %d", len(hashes), actual)
	}
}
