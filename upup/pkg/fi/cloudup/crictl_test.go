package cloudup

import (
	"testing"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/util/pkg/architectures"
	"k8s.io/kops/util/pkg/vfs"
)

func Test_FindCrictlVersionHash(t *testing.T) {
	desiredCrictlURL := "https://github.com/kubernetes-sigs/cri-tools/releases/download/v1.29.0/crictl-v1.29.0-linux-amd64.tar.gz"
	desiredCirctlHash := "sha256:d16a1ffb3938f5a19d5c8f45d363bd091ef89c0bc4d44ad16b933eede32fdcbb"

	cluster := &kops.Cluster{}
	cluster.Spec.KubernetesVersion = "v1.29.0"

	assetBuilder := assets.NewAssetBuilder(vfs.Context, cluster.Spec.Assets, cluster.Spec.KubernetesVersion, false)
	crictlAsset, crictlAssetHash, err := findCrictlAsset(cluster, assetBuilder, architectures.ArchitectureAmd64)
	if err != nil {
		t.Errorf("Unable to parse crictl version %s", err)
	}
	if crictlAsset.String() != desiredCrictlURL {
		t.Errorf("Expected crictl version %q, but got %q instead", desiredCrictlURL, crictlAsset)
	}
	if crictlAssetHash.String() != desiredCirctlHash {
		t.Errorf("Expected crictl version hash %q, but got %q instead", desiredCirctlHash, crictlAssetHash)
	}
}
