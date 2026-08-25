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

package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

type ociFixtureTransport func(*http.Request) (*http.Response, error)

func (f ociFixtureTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

// mockOCIFileRepository serves the staged assets for the complex cluster fixture.
func mockOCIFileRepository(t *testing.T) {
	t.Helper()
	digests := map[string]string{
		"nodeup:v1.37.0-beta.1-amd64":           "a899417ddb4fa0bba91c07c8d10b10747088a1067ca3dfa83097c1720eeda9ea",
		"nodeup:v1.37.0-beta.1-arm64":           "c2e424da38473c3ee1b9309e84e4167857fe8c01ea0df0328f49f6472aebac0e",
		"kubelet:v1.32.0-amd64":                 "5ad4965598773d56a37a8e8429c3dc3d86b4c5c26d8417ab333ae345c053dae2",
		"kubelet:v1.32.0-arm64":                 "bda9b2324c96693b38c41ecea051bab4c7c434be5683050b5e19025b50dbc0bf",
		"kubectl:v1.32.0-amd64":                 "646d58f6d98ee670a71d9cdffbf6625aeea2849d567f214bc43a35f8ccb7bf70",
		"kubectl:v1.32.0-arm64":                 "ba4004f98f3d3a7b7d2954ff0a424caa2c2b06b78c17b1dccf2acc76a311a896",
		"ecr-credential-provider:v1.31.7-amd64": "7644623e4ec9ad443ab352a8a5800a5180ee28741288be805286ba72bb8e7164",
		"ecr-credential-provider:v1.31.7-arm64": "1980e3a038cb16da48a137743b31fb81de6c0b59fa06c206c2bc20ce0a52f849",
		"cni-plugins:v1.6.2-amd64":              "b8e811578fb66023f90d2e238d80cec3bdfca4b44049af74c374d4fae0f9c090",
		"cni-plugins:v1.6.2-arm64":              "01e0e22acc7f7004e4588c1fe1871cc86d7ab562cd858e1761c4641d89ebfaa4",
		"containerd:v2.3.4-amd64":               "9d68969855fbf676cdb8ed758e420fb048d61f984f61de3e53eddfebe484d168",
		"containerd:v2.3.4-arm64":               "a985fbb7e18fc0362d31a055338f5d7b0e087a3e27f14c70d1c5965399a29f95",
		"runc:v1.4.3-amd64":                     "f6ae8efc0fa40079e1475e97cbe9d1bd3f106a28d6af78a11d9f1bd565515e60",
		"runc:v1.4.3-arm64":                     "633301e2e32f8a5ad54031aab4901eb00308bec677dd15faa2751e8f9dab5ca4",
	}
	original := remote.DefaultTransport
	t.Cleanup(func() { remote.DefaultTransport = original })
	remote.DefaultTransport = ociFixtureTransport(func(r *http.Request) (*http.Response, error) {
		if r.URL.Host != "registry.example.com" {
			return original.RoundTrip(r)
		}
		body := ""
		if r.URL.Path != "/v2/" {
			tag := strings.Replace(strings.TrimPrefix(r.URL.Path, "/v2/optional-prefix/"), "/manifests/", ":", 1)
			digest, found := digests[tag]
			if !found {
				return nil, fmt.Errorf("unexpected OCI fixture request: %s", r.URL)
			}
			body = fmt.Sprintf(`{"schemaVersion":2,"mediaType":%q,"layers":[{"digest":"sha256:%s"}]}`, types.OCIManifestSchema1, digest)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{string(types.OCIManifestSchema1)}},
			Body:       io.NopCloser(strings.NewReader(body)),
			Request:    r,
		}, nil
	})
}
