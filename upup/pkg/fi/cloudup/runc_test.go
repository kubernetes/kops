/*
Copyright 2022 The Kubernetes Authors.

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
	"os"
	"testing"

	"k8s.io/kops/util/pkg/architectures"
)

func TestRuncVersionsHashesAmd64(t *testing.T) {
	if os.Getenv("VERIFY_HASHES") == "" {
		t.Skip("VERIFY_HASHES not set, won't download & verify runc hashes")
	}

	for version, hash := range findAllRuncHashesAmd64() {
		t.Run(version+"-amd64", func(t *testing.T) {
			url, _ := findRuncVersionUrl(architectures.ArchitectureAmd64, version)
			if err := verifyPackageHash(url, hash); err != nil {
				t.Errorf("error verifying package %q: %v", url, err)
			}
		})
	}
}

func TestRuncVersionsHashesArm64(t *testing.T) {
	if os.Getenv("VERIFY_HASHES") == "" {
		t.Skip("VERIFY_HASHES not set, won't download & verify runc hashes")
	}

	for version, hash := range findAllRuncHashesArm64() {
		t.Run(version+"-arm64", func(t *testing.T) {
			url, _ := findRuncVersionUrl(architectures.ArchitectureArm64, version)
			if err := verifyPackageHash(url, hash); err != nil {
				t.Errorf("error verifying package %q: %v", url, err)
			}
		})
	}
}
