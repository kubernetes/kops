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

package cloudup

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"k8s.io/kops/util/pkg/architectures"
)

func TestCrioVersionUrl(t *testing.T) {
	tests := []struct {
		arch    architectures.Architecture
		version string
		url     string
		err     error
	}{
		{
			arch:    architectures.ArchitectureAmd64,
			version: "1.20.0",
			url:     "",
			err:     fmt.Errorf("crio version not supported"),
		},
		{
			arch:    architectures.ArchitectureAmd64,
			version: "1.21.0",
			url:     "https://storage.googleapis.com/k8s-conform-cri-o/artifacts/cri-o.amd64.v1.21.0.tar.gz",
			err:     nil,
		},
		{
			arch:    architectures.ArchitectureArm64,
			version: "1.21.0",
			url:     "https://storage.googleapis.com/k8s-conform-cri-o/artifacts/cri-o.arm64.v1.21.0.tar.gz",
			err:     nil,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s-%s", test.version, test.arch), func(t *testing.T) {
			url, err := findCrioVersionUrl(test.arch, test.version)
			fmt.Printf("the url is %s", url)
			if !reflect.DeepEqual(err, test.err) {
				t.Errorf("actual error %q differs from expected error %q", err, test.err)
				return
			}
			if url != test.url {
				t.Errorf("actual url %q differs from expected url %q", url, test.url)
				return
			}
		})
	}
}

func TestCrioVersionHash(t *testing.T) {
	tests := []struct {
		arch    architectures.Architecture
		version string
		hash    string
		err     error
	}{
		{
			arch:    architectures.ArchitectureAmd64,
			version: "1.20.0",
			hash:    "",
			err:     fmt.Errorf("version hash not found"),
		},
		{
			arch:    architectures.ArchitectureAmd64,
			version: "1.21.0",
			hash:    "75d1aefd93d9b6eea9627d20c1a8b65307e10396b7c28d440308bb351ca041bb",
			err:     nil,
		},
		{
			arch:    architectures.ArchitectureArm64,
			version: "1.21.0",
			hash:    "a3dc626ecd8ecd0561b6219abe49e5e64d86e5d45a1a6647eb10d9fd0dd3c8f7",
			err:     nil,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s-%s", test.version, test.arch), func(t *testing.T) {
			hash, err := findCrioVersionHash(test.arch, test.version)
			if !reflect.DeepEqual(err, test.err) {
				t.Errorf("actual error %q differs from expected error %q", err, test.err)
				return
			}
			if hash != test.hash {
				t.Errorf("actual url %q differs from expected url %q", hash, test.hash)
				return
			}
		})
	}
}

func TestCrioVersionsHashes(t *testing.T) {
	if os.Getenv("VERIFY_HASHES") == "" {
		t.Skip("VERIFY_HASHES not set, won't download & verify docker hashes")
	}

	for _, arch := range architectures.GetSupported() {
		for version, hash := range getCrioVersionHash(arch) {
			t.Run(version, func(t *testing.T) {
				url, _ := findCrioVersionUrl(arch, version)
				if err := verifyPackageHash(url, hash); err != nil {
					t.Errorf("error verifying package %q: %v", url, err)
				}
			})
		}
	}
}
