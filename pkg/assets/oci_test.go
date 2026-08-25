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

package assets

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/util/pkg/hashing"
	"k8s.io/kops/util/pkg/vfs"
)

type ociTestTransport func(*http.Request) (*http.Response, error)

func (f ociTestTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func mockOCITransport(t *testing.T, handler func(*http.Request) (*http.Response, error)) {
	t.Helper()
	originalRegistry, originalHTTP := remote.DefaultTransport, http.DefaultTransport
	t.Cleanup(func() { remote.DefaultTransport, http.DefaultTransport = originalRegistry, originalHTTP })
	remote.DefaultTransport = ociTestTransport(handler)
	http.DefaultTransport = ociTestTransport(handler)
}

func ociTestResponse(r *http.Request, code int, body string) *http.Response {
	return &http.Response{
		StatusCode: code,
		Header:     http.Header{"Content-Type": []string{string(types.OCIManifestSchema1)}},
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    r,
	}
}

func ociTestManifest(layers string) string {
	return fmt.Sprintf(`{"schemaVersion":2,"mediaType":%q,"config":{"mediaType":%q,"digest":"sha256:%s","size":2},"layers":%s}`,
		types.OCIManifestSchema1, types.OCIConfigJSON, strings.Repeat("c", 64), layers)
}

func TestOCIUpdateSelectsHashWithoutUpstream(t *testing.T) {
	digest := strings.Repeat("a", 64)
	for _, test := range []struct {
		name      string
		canonical string
		knownHash *hashing.Hash
		pinned    string
	}{
		{name: "nodeup without a checksum", canonical: "https://upstream.example.com/nodeup.xz"},
		{name: "cached checksum", canonical: "https://upstream.example.com/nodeup.xz"},
		{name: "supplied SHA256", canonical: "https://upstream.example.com/containerd.tgz", knownHash: hashing.MustFromString(strings.Repeat("b", 64)), pinned: strings.Repeat("b", 64)},
		{name: "supplied hash overrides embedded checksum", canonical: "https://dl.k8s.io/release/v1.32.0/bin/linux/amd64/kubelet", knownHash: hashing.MustFromString(strings.Repeat("b", 64)), pinned: strings.Repeat("b", 64)},
		{name: "embedded checksum", canonical: "https://dl.k8s.io/release/v1.32.0/bin/linux/amd64/kubelet", pinned: "5ad4965598773d56a37a8e8429c3dc3d86b4c5c26d8417ab333ae345c053dae2"},
	} {
		t.Run(test.name, func(t *testing.T) {
			resetDownloadedFileHashes(t)
			canonical := mustParseURL(t, test.canonical)
			if test.name == "cached checksum" {
				downloadedFileHashes.Store(canonical.String(), hashing.MustFromString(strings.Repeat("d", 64)))
			}
			manifestRequests := 0
			mockOCITransport(t, func(r *http.Request) (*http.Response, error) {
				if test.pinned != "" {
					t.Errorf("unexpected network request for pinned asset: %s", r.URL)
					return ociTestResponse(r, http.StatusForbidden, ""), nil
				}
				if r.URL.Scheme != "https" || r.URL.Host != "registry.example.com" {
					t.Errorf("unexpected upstream request: %s", r.URL)
					return ociTestResponse(r, http.StatusForbidden, ""), nil
				}
				switch r.URL.Path {
				case "/v2/":
					response := ociTestResponse(r, http.StatusUnauthorized, "")
					response.Header.Set("WWW-Authenticate", `Bearer realm="https://registry.example.com/token",service="assets"`)
					return response, nil
				case "/token":
					if got := r.URL.Query().Get("scope"); got != "repository:prefix/nodeup:pull" {
						t.Errorf("token scope = %q", got)
					}
					return ociTestResponse(r, http.StatusOK, `{"token":"anonymous-pull-token"}`), nil
				case "/v2/prefix/nodeup/manifests/v1.37.0-amd64":
					if r.Header.Get("Authorization") != "Bearer anonymous-pull-token" {
						t.Error("missing anonymous pull token")
					}
					manifestRequests++
					return ociTestResponse(r, http.StatusOK, ociTestManifest(fmt.Sprintf(`[{"digest":"sha256:%s","size":42}]`, digest))), nil
				default:
					t.Errorf("unexpected request: %s", r.URL)
					return ociTestResponse(r, http.StatusForbidden, ""), nil
				}
			})
			repository := "oci://registry.example.com/prefix"
			wantHash, wantRequests := digest, 1
			if test.pinned != "" {
				wantHash, wantRequests = test.pinned, 0
			}
			for range 2 {
				builder := NewAssetBuilder(vfs.NewVFSContext(), &kops.AssetsSpec{FileRepository: &repository}, false)
				asset, err := builder.RemapFileWithInfo(canonical, test.knownHash, FileAssetInfo{Family: "nodeup", Version: "1.37.0", Architecture: "amd64"})
				if err != nil {
					t.Fatal(err)
				}
				if got := BuildMirroredAsset(asset).CompactString(); got != wantHash+"@"+repository+"/nodeup:v1.37.0-amd64" {
					t.Fatalf("asset = %q", got)
				}
				if len(builder.FileAssets()) != 1 {
					t.Fatal("cached lookup did not register the asset with the new builder")
				}
			}
			if manifestRequests != wantRequests {
				t.Fatalf("manifest requests = %d, want %d", manifestRequests, wantRequests)
			}
		})
	}
}

func TestOCIUpdateRejectsInvalidManifestsWithoutUpstreamFallback(t *testing.T) {
	for _, test := range []struct {
		name string
		body string
		code int
		want string
	}{
		{name: "missing tag", code: http.StatusNotFound, want: "run 'kops get assets --copy'"},
		{name: "unauthorized", code: http.StatusForbidden, want: "reading staged OCI asset"},
		{name: "malformed manifest", body: "{", want: "decoding OCI asset"},
		{name: "no layers", body: ociTestManifest("[]"), want: "exactly one file layer"},
		{name: "multiple layers", body: ociTestManifest("[{},{}]"), want: "exactly one file layer"},
		{name: "missing digest", body: ociTestManifest("[{}]"), want: "file layer must use SHA-256"},
		{name: "invalid SHA256", body: ociTestManifest(`[{"digest":"sha256:bad"}]`), want: "decoding OCI asset"},
		{name: "wrong algorithm", body: ociTestManifest(fmt.Sprintf(`[{"digest":"sha512:%s"}]`, strings.Repeat("a", 128))), want: "decoding OCI asset"},
	} {
		t.Run(test.name, func(t *testing.T) {
			resetDownloadedFileHashes(t)
			mockOCITransport(t, func(r *http.Request) (*http.Response, error) {
				if r.URL.Host != "registry.example.com" {
					t.Errorf("unexpected upstream fallback: %s", r.URL)
				}
				if r.URL.Path == "/v2/" {
					return ociTestResponse(r, http.StatusOK, ""), nil
				}
				code := test.code
				if code == 0 {
					code = http.StatusOK
				}
				return ociTestResponse(r, code, test.body), nil
			})
			repository := "oci://registry.example.com/prefix"
			builder := NewAssetBuilder(vfs.NewVFSContext(), &kops.AssetsSpec{FileRepository: &repository}, false)
			_, err := builder.RemapFileWithInfo(mustParseURL(t, "https://upstream.example.com/nodeup.xz"), nil, FileAssetInfo{Family: "nodeup", Version: "1.37.0"})
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want %q", err, test.want)
			}
			if len(builder.FileAssets()) != 0 {
				t.Fatal("invalid asset was registered")
			}
		})
	}
}

func TestOCIUpdateRequiresDiscoverableTag(t *testing.T) {
	for _, version := range []string{"", strings.Repeat("a", 129)} {
		repository := "oci://registry.example.com/prefix"
		builder := NewAssetBuilder(nil, &kops.AssetsSpec{FileRepository: &repository}, false)
		_, err := builder.RemapFileWithInfo(mustParseURL(t, "https://upstream.example.com/tool"), nil, FileAssetInfo{Family: "tool", Version: version})
		if err == nil || !strings.Contains(err.Error(), "explicit SHA-256 is required") {
			t.Fatalf("version %q: error = %v", version, err)
		}
	}
}

func TestOCIUpdateHashTagPreservesPin(t *testing.T) {
	mockOCITransport(t, func(r *http.Request) (*http.Response, error) {
		t.Errorf("unexpected network request for pinned asset: %s", r.URL)
		return ociTestResponse(r, http.StatusForbidden, ""), nil
	})
	repository := "oci://registry.example.com/prefix"
	for _, test := range []struct {
		name      string
		canonical string
		knownHash *hashing.Hash
		want      string
	}{
		{name: "explicit", canonical: "https://upstream.example.com/tool", knownHash: hashing.MustFromString(strings.Repeat("a", 64)), want: strings.Repeat("a", 64)},
		{name: "embedded", canonical: "https://dl.k8s.io/release/v1.32.0/bin/linux/amd64/kubelet", want: "5ad4965598773d56a37a8e8429c3dc3d86b4c5c26d8417ab333ae345c053dae2"},
	} {
		t.Run(test.name, func(t *testing.T) {
			resetDownloadedFileHashes(t)
			location := repository + "/tool:sha256-" + test.want + "-amd64"
			// A previously resolved registry digest must not override a trusted pin.
			downloadedFileHashes.Store(location, hashing.MustFromString(strings.Repeat("b", 64)))
			builder := NewAssetBuilder(nil, &kops.AssetsSpec{FileRepository: &repository}, false)
			asset, err := builder.RemapFileWithInfo(mustParseURL(t, test.canonical), test.knownHash, FileAssetInfo{Family: "tool", Architecture: "amd64"})
			if err != nil {
				t.Fatal(err)
			}
			if got := BuildMirroredAsset(asset).CompactString(); got != test.want+"@"+location {
				t.Fatalf("asset = %q, want pin %s at %s", got, test.want, location)
			}
		})
	}
}

func TestOCIUpdateRejectsNonSHA256Pin(t *testing.T) {
	mockOCITransport(t, func(r *http.Request) (*http.Response, error) {
		t.Errorf("unexpected network request for unsupported pin: %s", r.URL)
		return ociTestResponse(r, http.StatusForbidden, ""), nil
	})
	for _, size := range []int{32, 40} {
		hash := hashing.MustFromString(strings.Repeat("a", size))
		t.Run(string(hash.Algorithm), func(t *testing.T) {
			repository := "oci://registry.example.com/prefix"
			builder := NewAssetBuilder(nil, &kops.AssetsSpec{FileRepository: &repository}, false)
			canonical := mustParseURL(t, "https://dl.k8s.io/release/v1.32.0/bin/linux/amd64/kubelet")
			_, err := builder.RemapFileWithInfo(canonical, hash, FileAssetInfo{Family: "kubelet", Version: "1.32.0", Architecture: "amd64"})
			if err == nil || !strings.Contains(err.Error(), "requires a SHA-256 for updates") {
				t.Fatalf("error = %v, want SHA-256 requirement", err)
			}
			if len(builder.FileAssets()) != 0 {
				t.Fatal("asset with unsupported pin was registered")
			}
		})
	}
}

func TestHTTPSOnlyTransport(t *testing.T) {
	called := false
	transport := &HTTPSOnlyTransport{Inner: ociTestTransport(func(request *http.Request) (*http.Response, error) {
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
