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

package components

import (
	"testing"

	"k8s.io/kops/upup/pkg/fi"

	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/pkg/assets"
)

func buildCrioCluster(version string) *kopsapi.Cluster {
	return &kopsapi.Cluster{
		Spec: kopsapi.ClusterSpec{
			CloudProvider:     "aws",
			KubernetesVersion: version,
			Networking: &kopsapi.NetworkingSpec{
				Kubenet: &kopsapi.KubenetNetworkingSpec{},
			},
		},
	}
}

func TestCrioVersion(t *testing.T) {
	tests := []struct {
		kubernetesVersion string
		expected          string
		err               error
	}{
		{
			kubernetesVersion: "1.21.0",
			expected:          "1.21.0",
			err:               nil,
		},
	}

	for _, v := range tests {

		c := buildCrioCluster(v.kubernetesVersion)
		c.Spec.ContainerRuntime = "crio"
		b := assets.NewAssetBuilder(c, "")

		version, err := util.ParseKubernetesVersion(v.kubernetesVersion)
		if err != nil {
			t.Fatalf("unexpected error from ParseKubernetesVersion %s: %v", v.kubernetesVersion, err)
		}

		ob := CrioOptionsBuilder{
			&OptionsContext{
				AssetBuilder:      b,
				KubernetesVersion: *version,
			},
		}

		err = ob.BuildOptions(&c.Spec)

		if err != nil && v.err != err {
			t.Fatalf("Expected error but error was not thrown")
		}

		if fi.StringValue(c.Spec.Crio.Version) != v.expected {
			t.Fatalf("Did not get expected version")
		}

	}
}

func Test_Build_Crio_Supported_Version(t *testing.T) {
	kubernetesVersions := []string{"1.21.0"}

	for _, v := range kubernetesVersions {

		c := buildCrioCluster(v)
		c.Spec.ContainerRuntime = "crio"
		b := assets.NewAssetBuilder(c, "")

		version, err := util.ParseKubernetesVersion(v)
		if err != nil {
			t.Fatalf("unexpected error from ParseKubernetesVersion %s: %v", v, err)
		}

		ob := CrioOptionsBuilder{
			&OptionsContext{
				AssetBuilder:      b,
				KubernetesVersion: *version,
			},
		}

		err = ob.BuildOptions(&c.Spec)
		if err != nil {
			t.Fatalf("unexpected error from BuildOptions: %v", err)
		}

		if c.Spec.Crio.SkipInstall == true {
			t.Fatalf("expecting install when Kubernetes version >= 1.19: %s", v)
		}
	}
}

func Test_Build_Crio_Unsupported_Version(t *testing.T) {
	kubernetesVersions := []string{"1.18.0", "1.19.0"}

	for _, v := range kubernetesVersions {

		c := buildCrioCluster(v)
		c.Spec.ContainerRuntime = "crio"
		b := assets.NewAssetBuilder(c, "")

		version, err := util.ParseKubernetesVersion(v)
		if err != nil {
			t.Fatalf("unexpected error from ParseKubernetesVersion %s: %v", v, err)
		}

		ob := CrioOptionsBuilder{
			&OptionsContext{
				AssetBuilder:      b,
				KubernetesVersion: *version,
			},
		}

		err = ob.BuildOptions(&c.Spec)
		if err == nil {
			t.Fatalf("expected error when kubernetes version is not supported: %v", err)
		}
	}
}
