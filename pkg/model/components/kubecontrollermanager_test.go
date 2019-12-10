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
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/assets"
)

func buildCluster() *api.Cluster {

	return &api.Cluster{
		Spec: api.ClusterSpec{
			CloudProvider:     "aws",
			KubernetesVersion: "v1.4.0",
		},
	}
}

func Test_Build_KCM_Builder_Lower_Version(t *testing.T) {
	versions := []string{"v1.4.0", "v1.4.7", "v1.5.0"}

	for _, v := range versions {

		c := buildCluster()
		c.Spec.KubernetesVersion = v
		b := assets.NewAssetBuilder(c, "")

		kcm := &KubeControllerManagerOptionsBuilder{
			Context: &OptionsContext{
				AssetBuilder: b,
			},
		}

		err := kcm.BuildOptions(&c.Spec)
		if err != nil {
			t.Fatalf("unexpected error from BuildOptions: %v", err)
		}

		if c.Spec.KubeControllerManager.AttachDetachReconcileSyncPeriod != nil {
			t.Fatalf("AttachDetachReconcileSyncPeriod should not be set for old kubernetes version %s", c.Spec.KubernetesVersion)
		}
	}

}

func Test_Build_KCM_Builder_High_Enough_Version(t *testing.T) {
	versions := []string{"v1.4.8", "v1.5.2", "v1.9.0", "v2.4.0"}
	for _, v := range versions {

		c := buildCluster()
		c.Spec.KubernetesVersion = v
		b := assets.NewAssetBuilder(c, "")

		kcm := &KubeControllerManagerOptionsBuilder{
			Context: &OptionsContext{
				AssetBuilder: b,
			},
		}

		err := kcm.BuildOptions(&c.Spec)
		if err != nil {
			t.Fatalf("unexpected error from BuildOptions %s", err)
		}

		if c.Spec.KubeControllerManager.AttachDetachReconcileSyncPeriod.Duration != time.Minute {
			t.Fatalf("AttachDetachReconcileSyncPeriod should be set to 1m - %s, for k8s version %s", c.Spec.KubeControllerManager.AttachDetachReconcileSyncPeriod.Duration.String(), c.Spec.KubernetesVersion)
		}
	}

}

func Test_Build_KCM_Builder_Change_Duration(t *testing.T) {

	c := buildCluster()
	c.Spec.KubernetesVersion = "v1.5.2"
	b := assets.NewAssetBuilder(c, "")

	kcm := &KubeControllerManagerOptionsBuilder{
		Context: &OptionsContext{
			AssetBuilder: b,
		},
	}

	c.Spec.KubeControllerManager = &api.KubeControllerManagerConfig{
		AttachDetachReconcileSyncPeriod: &metav1.Duration{},
	}

	c.Spec.KubeControllerManager.AttachDetachReconcileSyncPeriod.Duration = time.Minute * 5

	err := kcm.BuildOptions(&c.Spec)
	if err != nil {
		t.Fatalf("unexpected error from BuildOptions %s", err)
	}

	if c.Spec.KubeControllerManager.AttachDetachReconcileSyncPeriod.Duration != time.Minute*5 {
		t.Fatalf("AttachDetachReconcileSyncPeriod should be set to 5m - %s, for k8s version %s", c.Spec.KubeControllerManager.AttachDetachReconcileSyncPeriod.Duration.String(), c.Spec.KubernetesVersion)
	}

}
