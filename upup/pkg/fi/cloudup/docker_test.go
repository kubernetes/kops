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
	"path"
	"path/filepath"
	"reflect"
	"testing"

	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/architectures"
)

func TestDockerVersionUrlHash(t *testing.T) {
	tests := []struct {
		version string
		arch    architectures.Architecture
		hash    string
		url     string
		err     error
	}{
		{
			arch:    architectures.ArchitectureAmd64,
			version: "19.03.13",
			url:     "https://download.docker.com/linux/static/stable/x86_64/docker-19.03.13.tgz",
			hash:    "ddb13aff1fcdcceb710bf71a210169b9c1abfd7420eeaf42cf7975f8fae2fcc8",
			err:     nil,
		},
		{
			arch:    architectures.ArchitectureArm64,
			version: "19.03.13",
			url:     "https://download.docker.com/linux/static/stable/aarch64/docker-19.03.13.tgz",
			hash:    "bdf080af7d6f383ad80e415e9c1952a63c7038c149dc673b7598bfca4d3311ec",
			err:     nil,
		},
		{
			arch:    architectures.ArchitectureAmd64,
			version: "18.06.3",
			url:     "https://download.docker.com/linux/static/stable/x86_64/docker-18.06.3-ce.tgz",
			hash:    "346f9394393ee8db5f8bd1e229ee9d90e5b36931bdd754308b2ae68884dd6822",
			err:     nil,
		},
		{
			arch:    architectures.ArchitectureArm64,
			version: "18.06.3",
			url:     "https://download.docker.com/linux/static/stable/aarch64/docker-18.06.3-ce.tgz",
			hash:    "defb2ccc95c0825833216c8b9e0e15baaa51bcedb3efc1f393f5352d184dead4",
			err:     nil,
		},
		{
			arch:    architectures.ArchitectureAmd64,
			version: "17.03.1",
			url:     "https://download.docker.com/linux/static/stable/x86_64/docker-17.03.1-ce.tgz",
			hash:    "3e070e7b34e99cf631f44d0ff5cf9a127c0b8af5c53dfc3e1fce4f9615fbf603",
			err:     nil,
		},
		{
			arch:    architectures.ArchitectureArm64,
			version: "17.03.1",
			url:     "https://download.docker.com/linux/static/stable/aarch64/docker-17.09.0-ce.tgz",
			hash:    "2af5d112ab514d9b0b84d9e7360a5e7633e88b7168d1bbfc16c6532535cb0123",
			err:     nil,
		},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%s-%s", test.version, test.arch), func(t *testing.T) {
			url, hash, err := findDockerVersionUrlHash(test.arch, test.version)
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

func TestDockerVersionUrl(t *testing.T) {
	tests := []struct {
		version string
		arch    architectures.Architecture
		url     string
		err     error
	}{
		{
			arch:    "",
			version: "19.03.13",
			url:     "",
			err:     fmt.Errorf("unknown arch: \"\""),
		},
		{
			arch:    "arm",
			version: "19.03.13",
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
			version: "18.06.3",
			url:     "https://download.docker.com/linux/static/stable/x86_64/docker-18.06.3-ce.tgz",
			err:     nil,
		},
		{
			arch:    architectures.ArchitectureArm64,
			version: "18.06.3",
			url:     "https://download.docker.com/linux/static/stable/aarch64/docker-18.06.3-ce.tgz",
			err:     nil,
		},
		{
			arch:    architectures.ArchitectureAmd64,
			version: "19.03.13",
			url:     "https://download.docker.com/linux/static/stable/x86_64/docker-19.03.13.tgz",
			err:     nil,
		},
		{
			arch:    architectures.ArchitectureArm64,
			version: "19.03.13",
			url:     "https://download.docker.com/linux/static/stable/aarch64/docker-19.03.13.tgz",
			err:     nil,
		},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%s-%s", test.version, test.arch), func(t *testing.T) {
			url, err := findDockerVersionUrl(test.arch, test.version)
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

func TestDockerVersionHash(t *testing.T) {
	tests := []struct {
		version string
		arch    architectures.Architecture
		hash    string
		err     error
	}{
		{
			arch:    "",
			version: "19.03.13",
			hash:    "",
			err:     fmt.Errorf("unknown arch: \"\""),
		},
		{
			arch:    "arm",
			version: "19.03.13",
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
			version: "1.1.1",
			hash:    "",
			err:     fmt.Errorf("unsupported legacy Docker version: \"1.1.1\""),
		},
		{
			arch:    architectures.ArchitectureArm64,
			version: "1.1.1",
			hash:    "",
			err:     fmt.Errorf("unsupported legacy Docker version: \"1.1.1\""),
		},
		{
			arch:    architectures.ArchitectureAmd64,
			version: "19.03.13",
			hash:    "ddb13aff1fcdcceb710bf71a210169b9c1abfd7420eeaf42cf7975f8fae2fcc8",
			err:     nil,
		},
		{
			arch:    architectures.ArchitectureArm64,
			version: "19.03.13",
			hash:    "bdf080af7d6f383ad80e415e9c1952a63c7038c149dc673b7598bfca4d3311ec",
			err:     nil,
		},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%s-%s", test.version, test.arch), func(t *testing.T) {
			hash, err := findDockerVersionHash(test.arch, test.version)
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

func TestDockerVersionsHashesAmd64(t *testing.T) {
	if os.Getenv("VERIFY_HASHES") == "" {
		t.Skip("VERIFY_HASHES not set, won't download & verify docker hashes")
	}

	for version, hash := range findAllDockerHashesAmd64() {
		t.Run(version+"-amd64", func(t *testing.T) {
			url, _ := findDockerVersionUrl(architectures.ArchitectureAmd64, version)
			if err := verifyPackageHash(url, hash); err != nil {
				t.Errorf("error verifying package %q: %v", url, err)
			}
		})
	}
}

func TestDockerVersionsHashesArm64(t *testing.T) {
	if os.Getenv("VERIFY_HASHES") == "" {
		t.Skip("VERIFY_HASHES not set, won't download & verify docker hashes")
	}

	for version, hash := range findAllDockerHashesArm64() {
		t.Run(version+"-arm64", func(t *testing.T) {
			url, _ := findDockerVersionUrl(architectures.ArchitectureArm64, version)
			if err := verifyPackageHash(url, hash); err != nil {
				t.Errorf("error verifying package %q: %v", url, err)
			}
		})
	}
}

func verifyPackageHash(u string, h string) error {
	name := fmt.Sprintf("%s-%s", h, path.Base(u))
	path := filepath.Join("/tmp", name)

	actualHash, err := fi.DownloadURL(u, path, nil)
	if err != nil {
		return err
	}

	err = os.Remove(path)
	if err != nil {
		return err
	}

	if h != actualHash.Hex() {
		return fmt.Errorf("actual hash %q differs from expected hash %q", actualHash.Hex(), h)
	}

	return nil
}
