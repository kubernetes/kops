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

	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/pkg/assets"
)

func buildSchedulerConfigMapCluster() *api.Cluster {
	usePolicyConfigMap := true

	return &api.Cluster{
		Spec: api.ClusterSpec{
			CloudProvider:     "aws",
			KubernetesVersion: "v1.4.0",
			KubeScheduler: &api.KubeSchedulerConfig{
				UsePolicyConfigMap: &usePolicyConfigMap,
			},
		},
	}
}

func Test_Build_Scheduler_Without_PolicyConfigMap(t *testing.T) {
	versions := []string{"v1.6.0", "v1.6.4", "v1.7.0", "v1.7.4"}
	b := assets.NewAssetBuilder(nil)

	for _, v := range versions {

		c := buildCluster()

		version, err := util.ParseKubernetesVersion(v)

		if err != nil {
			t.Fatalf("unexpected error from ParseKubernetesVersion %s: %v", v, err)
		}

		ks := &KubeSchedulerOptionsBuilder{
			&OptionsContext{
				AssetBuilder:      b,
				KubernetesVersion: *version,
			},
		}

		spec := c.Spec

		spec.KubernetesVersion = v
		err = ks.BuildOptions(&spec)

		if err != nil {
			t.Fatalf("unexpected error from BuildOptions: %v", err)
		}
	}

}
func Test_Build_Scheduler_PolicyConfigMap_Unsupported_Version(t *testing.T) {
	versions := []string{"v1.6.0", "v1.6.4"}
	b := assets.NewAssetBuilder(nil)

	for _, v := range versions {

		c := buildSchedulerConfigMapCluster()

		version, err := util.ParseKubernetesVersion(v)

		if err != nil {
			t.Fatalf("unexpected error from ParseKubernetesVersion %s: %v", v, err)
		}

		ks := &KubeSchedulerOptionsBuilder{
			&OptionsContext{
				AssetBuilder:      b,
				KubernetesVersion: *version,
			},
		}

		spec := c.Spec

		spec.KubernetesVersion = v
		err = ks.BuildOptions(&spec)

		if err == nil {
			t.Fatalf("error is expected, but none are returned")
		}
	}

}

func Test_Build_Scheduler_PolicyConfigMap_Supported_Version(t *testing.T) {
	versions := []string{"v1.7.0", "v1.7.4", "v1.8.0"}
	b := assets.NewAssetBuilder(nil)

	for _, v := range versions {

		c := buildSchedulerConfigMapCluster()

		version, err := util.ParseKubernetesVersion(v)

		if err != nil {
			t.Fatalf("unexpected error from ParseKubernetesVersion %s: %v", v, err)
		}

		ks := &KubeSchedulerOptionsBuilder{
			&OptionsContext{
				AssetBuilder:      b,
				KubernetesVersion: *version,
			},
		}

		spec := c.Spec

		spec.KubernetesVersion = v
		err = ks.BuildOptions(&spec)

		if err != nil {
			t.Fatalf("unexpected error from BuildOptions %s: %v", v, err)
		}
	}

}
