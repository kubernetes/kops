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
	"fmt"
	"os"
	"reflect"
	"testing"

	"k8s.io/kops/util/pkg/architectures"
)

func TestRuncVersionUrlHash(t *testing.T) {
	tests := []struct {
		version string
		arch    architectures.Architecture
		hash    string
		url     string
		err     error
	}{
		{
			arch:    architectures.ArchitectureAmd64,
			version: "1.100.0",
			url:     "",
			hash:    "",
			err:     fmt.Errorf("unknown url and hash for runc version: amd64 - 1.100.0"),
		},
		{
			arch:    architectures.ArchitectureArm64,
			version: "1.100.0",
			url:     "",
			hash:    "",
			err:     fmt.Errorf("unknown url and hash for runc version: arm64 - 1.100.0"),
		},
		{
			arch:    architectures.ArchitectureAmd64,
			version: "1.1.0",
			url:     "https://github.com/opencontainers/runc/releases/download/v1.1.0/runc.amd64",
			hash:    "ab1c67fbcbdddbe481e48a55cf0ef9a86b38b166b5079e0010737fd87d7454bb",
			err:     nil,
		},
		{
			arch:    architectures.ArchitectureArm64,
			version: "1.1.0",
			url:     "https://github.com/opencontainers/runc/releases/download/v1.1.0/runc.arm64",
			hash:    "9ec8e68feabc4e7083a4cfa45ebe4d529467391e0b03ee7de7ddda5770b05e68",
			err:     nil,
		},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%s-%s", test.version, test.arch), func(t *testing.T) {
			url, hash, err := findRuncVersionUrlHash(test.arch, test.version)
			if !reflect.DeepEqual(err, test.err) {
				t.Errorf("actual error %q differs from expected error %q", err, test.err)
				return
			}
			if url != test.url {
				t.Errorf("actual url %q differs from expected url %q", url, test.url)
				return
			}
			if hash != test.hash {
				t.Errorf("actual hash %q differs from expected hash %q", hash, test.hash)
				return
			}
		})
	}
}

func TestRuncVersionUrl(t *testing.T) {
	tests := []struct {
		version string
		arch    architectures.Architecture
		url     string
		err     error
	}{
		{
			arch:    "",
			version: "1.1.0",
			url:     "",
			err:     fmt.Errorf("unknown arch: \"\""),
		},
		{
			arch:    "arm",
			version: "1.1.0",
			url:     "",
			err:     fmt.Errorf("unknown arch: \"arm\""),
		},
		{
			arch:    architectures.ArchitectureAmd64,
			version: "",
			url:     "",
			err:     fmt.Errorf("unable to parse version string: \"\""),
		},
		{
			arch:    architectures.ArchitectureArm64,
			version: "",
			url:     "",
			err:     fmt.Errorf("unable to parse version string: \"\""),
		},
		{
			arch:    architectures.ArchitectureAmd64,
			version: "1.0.0",
			url:     "",
			err:     fmt.Errorf("unsupported runc version: \"1.0.0\""),
		},
		{
			arch:    architectures.ArchitectureArm64,
			version: "1.0.0",
			url:     "",
			err:     fmt.Errorf("unsupported runc version: \"1.0.0\""),
		},
		{
			arch:    architectures.ArchitectureAmd64,
			version: "1.1.0",
			url:     "https://github.com/opencontainers/runc/releases/download/v1.1.0/runc.amd64",
			err:     nil,
		},
		{
			arch:    architectures.ArchitectureArm64,
			version: "1.1.0",
			url:     "https://github.com/opencontainers/runc/releases/download/v1.1.0/runc.arm64",
			err:     nil,
		},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%s-%s", test.version, test.arch), func(t *testing.T) {
			url, err := findRuncVersionUrl(test.arch, test.version)
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

func TestRuncVersionHash(t *testing.T) {
	tests := []struct {
		version string
		arch    architectures.Architecture
		hash    string
		err     error
	}{
		{
			arch:    "",
			version: "1.1.0",
			hash:    "",
			err:     fmt.Errorf("unknown arch: \"\""),
		},
		{
			arch:    "arm",
			version: "1.1.0",
			hash:    "",
			err:     fmt.Errorf("unknown arch: \"arm\""),
		},
		{
			arch:    architectures.ArchitectureAmd64,
			version: "",
			hash:    "",
			err:     fmt.Errorf("unable to parse version string: \"\""),
		},
		{
			arch:    architectures.ArchitectureArm64,
			version: "",
			hash:    "",
			err:     fmt.Errorf("unable to parse version string: \"\""),
		},
		{
			arch:    architectures.ArchitectureAmd64,
			version: "1.0.0",
			hash:    "",
			err:     fmt.Errorf("unsupported runc version: \"1.0.0\""),
		},
		{
			arch:    architectures.ArchitectureArm64,
			version: "1.0.0",
			hash:    "",
			err:     fmt.Errorf("unsupported runc version: \"1.0.0\""),
		},
		{
			arch:    architectures.ArchitectureAmd64,
			version: "1.100.0",
			hash:    "",
			err:     fmt.Errorf("unknown hash for runc version: amd64 - 1.100.0"),
		},
		{
			arch:    architectures.ArchitectureArm64,
			version: "1.100.0",
			hash:    "",
			err:     fmt.Errorf("unknown hash for runc version: arm64 - 1.100.0"),
		},
		{
			arch:    architectures.ArchitectureAmd64,
			version: "1.1.0",
			hash:    "ab1c67fbcbdddbe481e48a55cf0ef9a86b38b166b5079e0010737fd87d7454bb",
			err:     nil,
		},
		{
			arch:    architectures.ArchitectureArm64,
			version: "1.1.0",
			hash:    "9ec8e68feabc4e7083a4cfa45ebe4d529467391e0b03ee7de7ddda5770b05e68",
			err:     nil,
		},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%s-%s", test.version, test.arch), func(t *testing.T) {
			hash, err := findRuncVersionHash(test.arch, test.version)
			if !reflect.DeepEqual(err, test.err) {
				t.Errorf("actual error %q differs from expected error %q", err, test.err)
				return
			}
			if hash != test.hash {
				t.Errorf("actual hash %q differs from expected hash %q", hash, test.hash)
				return
			}
		})
	}
}

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
