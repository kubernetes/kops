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
	"net/url"
	"reflect"
	"testing"

	"k8s.io/kops"
	"k8s.io/kops/util/pkg/hashing"
)

func Test_BuildMirroredAsset(t *testing.T) {
	tests := []struct {
		url      string
		hash     string
		expected []string
	}{
		{
			url: "https://kubeupv2.s3.amazonaws.com/kops/%s/images/protokube-linux-amd64",
			expected: []string{
				"https://artifacts.k8s.io/binaries/kops/1.19.0-alpha.5/images/protokube-linux-amd64",
				"https://github.com/kubernetes/kops/releases/download/v1.19.0-alpha.5/images-protokube-linux-amd64",
				"https://kubeupv2.s3.amazonaws.com/kops/1.19.0-alpha.5/images/protokube-linux-amd64",
			},
		},
		{
			url: "https://kubeupv2.s3.amazonaws.com/kops/%s/images/protokube-linux-arm64",
			expected: []string{
				"https://artifacts.k8s.io/binaries/kops/1.19.0-alpha.5/images/protokube-linux-arm64",
				"https://github.com/kubernetes/kops/releases/download/v1.19.0-alpha.5/images-protokube-linux-arm64",
				"https://kubeupv2.s3.amazonaws.com/kops/1.19.0-alpha.5/images/protokube-linux-arm64",
			},
		},
		{
			url: "https://kubeupv2.s3.amazonaws.com/kops/%s/linux/amd64/nodeup",
			expected: []string{
				"https://artifacts.k8s.io/binaries/kops/1.19.0-alpha.5/linux/amd64/nodeup",
				"https://github.com/kubernetes/kops/releases/download/v1.19.0-alpha.5/nodeup-linux-amd64",
				"https://kubeupv2.s3.amazonaws.com/kops/1.19.0-alpha.5/linux/amd64/nodeup",
			},
		},
		{
			url: "https://kubeupv2.s3.amazonaws.com/kops/%s/linux/arm64/nodeup",
			expected: []string{
				"https://artifacts.k8s.io/binaries/kops/1.19.0-alpha.5/linux/arm64/nodeup",
				"https://github.com/kubernetes/kops/releases/download/v1.19.0-alpha.5/nodeup-linux-arm64",
				"https://kubeupv2.s3.amazonaws.com/kops/1.19.0-alpha.5/linux/arm64/nodeup",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.url, func(t *testing.T) {
			h, _ := hashing.FromString("0000000000000000000000000000000000000000000000000000000000000000")
			u, err := url.Parse(fmt.Sprintf(tc.url, kops.Version))
			if err != nil {
				t.Errorf("cannot parse URL: %s", fmt.Sprintf(tc.url, kops.Version))
				return
			}
			actual := BuildMirroredAsset(u, h)

			if !reflect.DeepEqual(actual.Locations, tc.expected) {
				t.Errorf("Locations differ:\nActual: %+v\nExpect: %+v", actual.Locations, tc.expected)
				return
			}
		})
	}
}
