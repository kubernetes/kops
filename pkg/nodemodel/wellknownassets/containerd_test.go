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
	"reflect"
	"testing"

	"k8s.io/kops/util/pkg/architectures"
)

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
			version: "1.6.5",
			url:     "https://github.com/containerd/containerd/releases/download/v1.6.5/containerd-1.6.5-linux-amd64.tar.gz",
			err:     nil,
		},
		{
			arch:    architectures.ArchitectureArm64,
			version: "1.5.5",
			url:     "",
			err:     fmt.Errorf("unknown url for containerd version: arm64 - 1.5.5"),
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
