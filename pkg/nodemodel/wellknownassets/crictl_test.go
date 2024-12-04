/*
Copyright 2024 The Kubernetes Authors.

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
	"testing"

	"k8s.io/kops/pkg/apis/kops"
	kopsmodel "k8s.io/kops/pkg/apis/kops/model"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/util/pkg/architectures"
	"k8s.io/kops/util/pkg/vfs"
)

func Test_FindCrictlVersionHash(t *testing.T) {
	desiredCrictlURL := "https://github.com/kubernetes-sigs/cri-tools/releases/download/v1.29.0/crictl-v1.29.0-linux-amd64.tar.gz"
	desiredCirctlHash := "sha256:d16a1ffb3938f5a19d5c8f45d363bd091ef89c0bc4d44ad16b933eede32fdcbb"

	cluster := &kops.Cluster{}
	cluster.Spec.KubernetesVersion = "v1.29.0"

	ig := &kops.InstanceGroup{}

	igModel, err := kopsmodel.ForInstanceGroup(cluster, ig)
	if err != nil {
		t.Fatalf("building instance group model: %v", err)
	}

	assetBuilder := assets.NewAssetBuilder(vfs.Context, cluster.Spec.Assets, false)

	crictlAsset, err := FindCrictlAsset(igModel, assetBuilder, architectures.ArchitectureAmd64)
	if err != nil {
		t.Errorf("Unable to parse crictl version %s", err)
	}
	if crictlAsset.DownloadURL.String() != desiredCrictlURL {
		t.Errorf("Expected crictl version %q, but got %q instead", desiredCrictlURL, crictlAsset.DownloadURL)
	}
	if crictlAsset.SHAValue.String() != desiredCirctlHash {
		t.Errorf("Expected crictl version hash %q, but got %q instead", desiredCirctlHash, crictlAsset.SHAValue)
	}
}
