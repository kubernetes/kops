/*
Copyright 2019 The Kubernetes Authors.

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
	"os"
	"testing"

	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/assets"
)

func Test_FindCNIAssetFromEnvironmentVariable(t *testing.T) {

	desiredCNIVersion := "https://storage.googleapis.com/kubernetes-release/network-plugins/cni-TEST-VERSION.tar.gz"
	os.Setenv(ENV_VAR_CNI_VERSION_URL, desiredCNIVersion)
	defer func() {
		os.Unsetenv(ENV_VAR_CNI_VERSION_URL)
	}()

	cluster := &api.Cluster{}
	cluster.Spec.KubernetesVersion = "v1.9.0"

	assetBuilder := assets.NewAssetBuilder(cluster, "")
	cniAsset, cniAssetHash, err := findCNIAssets(cluster, assetBuilder)

	if err != nil {
		t.Errorf("Unable to parse k8s version %s", err)
	}

	if cniAsset.String() != desiredCNIVersion {
		t.Errorf("Expected CNI version from Environment variable %q, but got %q instead", desiredCNIVersion, cniAsset)
	}

	if cniAssetHash != nil {
		t.Errorf("Expected Empty CNI Version Hash String, but got %v instead", cniAssetHash)
	}
}
