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
	t.Skip("not going to run")
	grid := []struct {
		testFile   string
		expected   string
		sha        string
		asset      *FileAsset
		kopsAssets *kops.Assets
	}{
		//defaultCNIAssetK8s1_6           = "https://storage.googleapis.com/kubernetes-release/network-plugins/cni-0799f5732f2a11b329d9e3d51b9c8f2e3759f2ff.tar.gz"
		//defaultCNIAssetHashStringK8s1_6 = "1d9788b0f5420e1a219aad2cb8681823fc515e7c"
		{
			// FIXME - need https://s3.amazonaws.com/k8s-for-greeks-kops/kubernetes-release/release/v1.7.2/bin/linux/amd64/kubelet
			"https://gcr.io/kubernetes-release/release/v1.7.2/bin/linux/amd64/kubelet",
			"s3://clove-kops/kubernetes-release/release/v1.7.2/bin/linux/amd64/kubelet",
			"1d9788b0f5420e1a219aad2cb8681823fc515e7c",
			&FileAsset{
				File:              "s3://clove-kops/kubernetes-release/release/v1.7.2/bin/linux/amd64/kubelet",
				CanonicalLocation: "https://gcr.io/kubernetes-release/release/v1.7.2/bin/linux/amd64/kubelet",
			},
			&kops.Assets{
				FileRepository: s("s3://clove-kops"),
			},
		},
	}

	for _, g := range grid {
		// TODO FIXME
		builder := NewAssetBuilder(g.kopsAssets)

		actual, _, err := builder.RemapFileAndSHA(g.testFile, g.sha)
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
