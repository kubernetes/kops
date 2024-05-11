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

package wellknownassets

import (
	"fmt"
	"reflect"
	"testing"

	"k8s.io/kops/util/pkg/architectures"
)

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
