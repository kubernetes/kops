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

package components

import (
	"testing"

	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/util/pkg/vfs"
)

func buildContainerdCluster(version string) *kopsapi.Cluster {
	return &kopsapi.Cluster{
		Spec: kopsapi.ClusterSpec{
			CloudProvider: kopsapi.CloudProviderSpec{
				AWS: &kopsapi.AWSSpec{},
			},
			KubernetesVersion: version,
			Networking: kopsapi.NetworkingSpec{
				Kubenet: &kopsapi.KubenetNetworkingSpec{},
			},
		},
	}
}

func Test_Build_Containerd_Supported_Version(t *testing.T) {
	kubernetesVersions := []string{"1.18.0", "1.18.3"}

	for _, v := range kubernetesVersions {

		c := buildContainerdCluster(v)
		b := assets.NewAssetBuilder(vfs.Context, c.Spec.Assets, c.Spec.KubernetesVersion, false)

		version, err := util.ParseKubernetesVersion(v)
		if err != nil {
			t.Fatalf("unexpected error from ParseKubernetesVersion %s: %v", v, err)
		}

		ob := &ContainerdOptionsBuilder{
			&OptionsContext{
				AssetBuilder:      b,
				KubernetesVersion: *version,
			},
		}

		err = ob.BuildOptions(c)
		if err != nil {
			t.Fatalf("unexpected error from BuildOptions: %v", err)
		}

		if c.Spec.Containerd.SkipInstall == true {
			t.Fatalf("expecting install when Kubernetes version >= 1.11: %s", v)
		}
	}
}
