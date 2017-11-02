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

package assets

import (
	"testing"

	"k8s.io/kops/pkg/apis/kops"
)

func TestRemap_File(t *testing.T) {
	grid := []struct {
		testFile   string
		expected   string
		asset      *FileAsset
		kopsAssets *kops.Assets
	}{
		{
			// FIXME - need https://s3.amazonaws.com/k8s-for-greeks-kops/kubernetes-release/release/v1.7.2/bin/linux/amd64/kubelet
			"https://gcr.io/kubernetes-release/release/v1.7.2/bin/linux/amd64/kubelet",
			"s3://k8s-for-greeks-kops/kubernetes-release/release/v1.7.2/bin/linux/amd64/kubelet",
			&FileAsset{
				File:              "s3://k8s-for-greeks-kops/kubernetes-release/release/v1.7.2/bin/linux/amd64/kubelet",
				CanonicalLocation: "https://gcr.io/kubernetes-release/release/v1.7.2/bin/linux/amd64/kubelet",
			},
			&kops.Assets{
				FileRepository: s("s3://k8s-for-greeks-kops"),
			},
		},
	}

	for _, g := range grid {
		builder := NewAssetBuilder(g.kopsAssets)

		actual, err := builder.RemapFile(g.testFile)
		if err != nil {
			t.Errorf("err occurred: %v", err)
		}
		if actual != g.expected {
			t.Errorf("results did not match.  actual=%q expected=%q", actual, g.expected)
		}
	}
}

func s(s string) *string {
	return &s
}
