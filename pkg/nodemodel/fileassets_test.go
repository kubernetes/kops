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

package nodemodel

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/model"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/util/pkg/architectures"
	"k8s.io/kops/util/pkg/vfs"
)

type kubernetesAssetTestTransport func(*http.Request) (*http.Response, error)

func (f kubernetesAssetTestTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestKubernetesURLAssetsToOCI(t *testing.T) {
	const baseURL = "https://upstream.example.com/ci/v1.37.0-alpha.0.123+abcdef"
	const tag = "v1.37.0-alpha.0.123-abcdef"
	digest := strings.Repeat("a", 64)
	t.Setenv("KOPS_BASE_URL", "")
	for _, test := range []struct {
		name      string
		version   string
		getAssets bool
	}{
		{name: "update", version: baseURL},
		{name: "stage", version: baseURL, getAssets: true},
		{name: "update trailing slash", version: baseURL + "/"},
		{name: "stage trailing slash", version: baseURL + "/", getAssets: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			originalRegistry, originalHTTP := remote.DefaultTransport, http.DefaultTransport
			t.Cleanup(func() { remote.DefaultTransport, http.DefaultTransport = originalRegistry, originalHTTP })
			transport := kubernetesAssetTestTransport(func(r *http.Request) (*http.Response, error) {
				body := ""
				switch {
				case test.getAssets && r.URL.Host == "upstream.example.com" && strings.HasSuffix(r.URL.Path, ".sha256"):
					body = digest
				case !test.getAssets && r.URL.Host == "registry.example.com" && r.URL.Path == "/v2/":
				case !test.getAssets && r.URL.Host == "registry.example.com" && strings.Contains(r.URL.Path, "/manifests/"):
					body = fmt.Sprintf(`{"schemaVersion":2,"mediaType":%q,"config":{"mediaType":%q,"digest":"sha256:%s","size":2},"layers":[{"digest":"sha256:%s","size":42}]}`,
						types.OCIManifestSchema1, types.OCIConfigJSON, strings.Repeat("b", 64), digest)
				default:
					t.Errorf("unexpected request: %s", r.URL)
					return nil, fmt.Errorf("unexpected request: %s", r.URL)
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{"Content-Type": []string{string(types.OCIManifestSchema1)}},
					Body:       io.NopCloser(strings.NewReader(body)),
					Request:    r,
				}, nil
			})
			remote.DefaultTransport, http.DefaultTransport = transport, transport

			cluster := &kops.Cluster{Spec: kops.ClusterSpec{
				KubernetesVersion: test.version,
				ConfigStore:       kops.ConfigStoreSpec{Base: t.TempDir()},
				Networking:        kops.NetworkingSpec{Cilium: &kops.CiliumNetworkingSpec{}},
				Containerd:        &kops.ContainerdConfig{SkipInstall: true},
			}}
			ig, err := model.ForInstanceGroup(cluster, &kops.InstanceGroup{})
			if err != nil {
				t.Fatal(err)
			}
			repository := "oci://registry.example.com/kubernetes"
			builder := assets.NewAssetBuilder(vfs.NewVFSContext(), &kops.AssetsSpec{FileRepository: &repository}, test.getAssets)
			if _, err := BuildKubernetesFileAssets(ig, builder); err != nil {
				t.Errorf("building Kubernetes binaries: %v", err)
			}
			if _, err := NewNodeUpConfigBuilder(cluster, builder, ""); err != nil {
				t.Errorf("building Kubernetes component images: %v", err)
			}

			want := make(map[string]string)
			for _, family := range []string{"kubelet", "kubectl", "kube-proxy", "kube-apiserver", "kube-controller-manager", "kube-scheduler"} {
				filename := family
				if family != "kubelet" && family != "kubectl" {
					filename += ".tar"
				}
				for _, arch := range architectures.GetSupported() {
					location := repository + "/" + family + ":" + tag + "-" + string(arch)
					want[location] = baseURL + "/bin/linux/" + string(arch) + "/" + filename
				}
			}
			seen := make(map[string]bool)
			for _, asset := range builder.FileAssets() {
				location := asset.DownloadURL.String()
				if canonical, found := want[location]; !found || canonical != asset.CanonicalURL.String() {
					t.Errorf("unexpected remapping: %s -> %s", asset.CanonicalURL, location)
				}
				if asset.SHAValue.Hex() != digest {
					t.Errorf("hash for %s = %s, want %s", location, asset.SHAValue, digest)
				}
				seen[location] = true
			}
			if len(seen) != len(want) {
				t.Errorf("got %d distinct asset locations, want %d", len(seen), len(want))
			}
		})
	}
}
