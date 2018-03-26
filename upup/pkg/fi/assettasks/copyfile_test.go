/*
Copyright 2017 The Kubernetes Authors.

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

package assettasks

import (
	"testing"
)

func Test_BuildVFSPath(t *testing.T) {

	grid := []struct {
		target  string
		vfsPath string
		pass    bool
	}{
		{
			"https://s3.amazonaws.com/k8s-for-greeks-kops/kubernetes-release/release/v1.7.2/bin/linux/amd64/kubectl",
			"s3://k8s-for-greeks-kops/kubernetes-release/release/v1.7.2/bin/linux/amd64/kubectl",
			true,
		},
		{
			"https://s3.cn-north-1.amazonaws.com/k8s-for-greeks-kops/kubernetes-release/release/v1.7.2/bin/linux/amd64/kubectl",
			"s3://k8s-for-greeks-kops/kubernetes-release/release/v1.7.2/bin/linux/amd64/kubectl",
			true,
		},
		{
			"https://s3-cn-north-1.amazonaws.com/k8s-for-greeks-kops/kubernetes-release/release/v1.7.2/bin/linux/amd64/kubectl",
			"s3://k8s-for-greeks-kops/kubernetes-release/release/v1.7.2/bin/linux/amd64/kubectl",
			true,
		},
		{
			"https://s3.k8s-for-greeks-kops.amazonaws.com/kubernetes-release/release/v1.7.2/bin/linux/amd64/kubectl",
			"s3://k8s-for-greeks-kops/kubernetes-release/release/v1.7.2/bin/linux/amd64/kubectl",
			true,
		},
		{
			"https://s3-k8s-for-greeks-kops.amazonaws.com/kubernetes-release/release/v1.7.2/bin/linux/amd64/kubectl",
			"s3://k8s-for-greeks-kops/kubernetes-release/release/v1.7.2/bin/linux/amd64/kubectl",
			true,
		},
		{
			"https://foo/k8s-for-greeks-kops/kubernetes-release/release/v1.7.2/bin/linux/amd64/kubectl",
			"",
			false,
		},
		{
			"kubernetes-release/release/v1.7.2/bin/linux/amd64/kubectl",
			"",
			false,
		},

		{
			"https://storage.googleapis.com/k8s-for-greeks-kops/kubernetes-release/release/v1.7.2/bin/linux/amd64/kubectl",
			"gs://k8s-for-greeks-kops/kubernetes-release/release/v1.7.2/bin/linux/amd64/kubectl",
			true,
		},
	}

	for _, test := range grid {
		path, err := buildVFSPath(test.target)
		if err != nil && test.pass {
			t.Errorf("error thrown, but expected to pass: %q, %v", test.target, err)
		}

		if test.pass && path != test.vfsPath {
			t.Errorf("incorrect url parsed %q", path)
		}
	}

}
