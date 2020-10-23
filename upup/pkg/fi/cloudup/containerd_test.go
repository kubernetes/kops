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

func TestContainerdVersionUrlHash(t *testing.T) {
	tests := []struct {
		version string
		arch    architectures.Architecture
		hash    string
		url     string
		err     error
	}{
		{
			arch:    architectures.ArchitectureAmd64,
			version: "1.3.4",
			url:     "https://storage.googleapis.com/cri-containerd-release/cri-containerd-1.3.4.linux-amd64.tar.gz",
			hash:    "4616971c3ad21c24f2f2320fa1c085577a91032a068dd56a41c7c4b71a458087",
			err:     nil,
		},
		{
			arch:    architectures.ArchitectureArm64,
			version: "1.3.4",
			url:     "https://download.docker.com/linux/static/stable/aarch64/docker-19.03.11.tgz",
			hash:    "9cd49fe82f6b7ec413b04daef35bc0c87b01d6da67611e5beef36291538d3145",
			err:     nil,
		},
		{
			arch:    architectures.ArchitectureAmd64,
			version: "1.3.7",
			url:     "https://download.docker.com/linux/static/stable/x86_64/docker-19.03.13.tgz",
			hash:    "ddb13aff1fcdcceb710bf71a210169b9c1abfd7420eeaf42cf7975f8fae2fcc8",
			err:     nil,
		},
		{
			arch:    architectures.ArchitectureArm64,
			version: "1.3.7",
			url:     "https://download.docker.com/linux/static/stable/aarch64/docker-19.03.13.tgz",
			hash:    "bdf080af7d6f383ad80e415e9c1952a63c7038c149dc673b7598bfca4d3311ec",
			err:     nil,
		},
		{
			arch:    architectures.ArchitectureAmd64,
			version: "1.4.1",
			url:     "https://github.com/containerd/containerd/releases/download/v1.4.1/cri-containerd-cni-1.4.1-linux-amd64.tar.gz",
			hash:    "757efb93a4f3161efc447a943317503d8a7ded5cb4cc0cba3f3318d7ce1542ed",
			err:     nil,
		},
		{
			arch:    architectures.ArchitectureArm64,
			version: "1.4.1",
			url:     "https://download.docker.com/linux/static/stable/aarch64/docker-19.03.11.tgz",
			hash:    "9cd49fe82f6b7ec413b04daef35bc0c87b01d6da67611e5beef36291538d3145",
			err:     nil,
		},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%s-%s", test.version, test.arch), func(t *testing.T) {
			url, hash, err := findContainerdVersionUrlHash(test.arch, test.version)
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

func TestContainerdVersionUrl(t *testing.T) {
	tests := []struct {
		version string
		arch    architectures.Architecture
		url     string
		err     error
	}{
		{
			arch:    "",
			version: "1.4.1",
			url:     "",
			err:     fmt.Errorf("unknown arch: \"\""),
		},
		{
			arch:    "arm",
			version: "1.4.1",
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
			version: "1.1.1",
			url:     "",
			err:     fmt.Errorf("unsupported legacy containerd version: \"1.1.1\""),
		},
		{
			arch:    architectures.ArchitectureArm64,
			version: "1.1.1",
			url:     "",
			err:     fmt.Errorf("unsupported legacy containerd version: \"1.1.1\""),
		},
		{
			arch:    architectures.ArchitectureAmd64,
			version: "1.3.5",
			url:     "https://storage.googleapis.com/cri-containerd-release/cri-containerd-1.3.5.linux-amd64.tar.gz",
			err:     nil,
		},
		{
			arch:    architectures.ArchitectureArm64,
			version: "1.3.5",
			url:     "",
			err:     fmt.Errorf("unknown url for containerd version: arm64 - 1.3.5"),
		},
		{
			arch:    architectures.ArchitectureAmd64,
			version: "1.3.4",
			url:     "https://storage.googleapis.com/cri-containerd-release/cri-containerd-1.3.4.linux-amd64.tar.gz",
			err:     nil,
		},
		{
			arch:    architectures.ArchitectureArm64,
			version: "1.3.4",
			url:     "https://download.docker.com/linux/static/stable/aarch64/docker-19.03.11.tgz",
			err:     nil,
		},
		{
			arch:    architectures.ArchitectureAmd64,
			version: "1.4.1",
			url:     "https://github.com/containerd/containerd/releases/download/v1.4.1/cri-containerd-cni-1.4.1-linux-amd64.tar.gz",
			err:     nil,
		},
		{
			arch:    architectures.ArchitectureArm64,
			version: "1.4.1",
			url:     "https://download.docker.com/linux/static/stable/aarch64/docker-19.03.11.tgz",
			err:     nil,
		},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%s-%s", test.version, test.arch), func(t *testing.T) {
			url, err := findContainerdVersionUrl(test.arch, test.version)
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

func TestContainerdVersionHash(t *testing.T) {
	tests := []struct {
		version string
		arch    architectures.Architecture
		hash    string
		err     error
	}{
		{
			arch:    "",
			version: "1.4.1",
			hash:    "",
			err:     fmt.Errorf("unknown arch: \"\""),
		},
		{
			arch:    "arm",
			version: "1.4.1",
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
			err:     fmt.Errorf("unsupported legacy containerd version: \"1.1.1\""),
		},
		{
			arch:    architectures.ArchitectureArm64,
			version: "1.1.1",
			hash:    "",
			err:     fmt.Errorf("unsupported legacy containerd version: \"1.1.1\""),
		},
		{
			arch:    architectures.ArchitectureAmd64,
			version: "1.3.5",
			hash:    "",
			err:     fmt.Errorf("unknown hash for containerd version: amd64 - 1.3.5"),
		},
		{
			arch:    architectures.ArchitectureArm64,
			version: "1.3.5",
			hash:    "",
			err:     fmt.Errorf("unknown hash for containerd version: arm64 - 1.3.5"),
		},
		{
			arch:    architectures.ArchitectureAmd64,
			version: "1.4.1",
			hash:    "757efb93a4f3161efc447a943317503d8a7ded5cb4cc0cba3f3318d7ce1542ed",
			err:     nil,
		},
		{
			arch:    architectures.ArchitectureArm64,
			version: "1.4.1",
			hash:    "9cd49fe82f6b7ec413b04daef35bc0c87b01d6da67611e5beef36291538d3145",
			err:     nil,
		},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%s-%s", test.version, test.arch), func(t *testing.T) {
			hash, err := findContainerdVersionHash(test.arch, test.version)
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

func TestContainerdVersionsHashesAmd64(t *testing.T) {
	if os.Getenv("VERIFY_HASHES") == "" {
		t.Skip("VERIFY_HASHES not set, won't download & verify docker hashes")
	}

	for version, hash := range findAllContainerdHashesAmd64() {
		t.Run(version+"-amd64", func(t *testing.T) {
			url, _ := findContainerdVersionUrl(architectures.ArchitectureAmd64, version)
			if err := verifyPackageHash(url, hash); err != nil {
				t.Errorf("error verifying package %q: %v", url, err)
			}
		})
	}
}
