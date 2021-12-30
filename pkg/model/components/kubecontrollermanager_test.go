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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/upup/pkg/fi"
)

func buildCluster() *api.Cluster {
	return &api.Cluster{
		Spec: api.ClusterSpec{
			CloudProvider: api.CloudProviderSpec{
				AWS: &api.AWSSpec{},
			},
			KubernetesVersion: "v1.20.0",
			KubeAPIServer:     &api.KubeAPIServerConfig{},
		},
	}
}

func Test_Build_KCM_Builder(t *testing.T) {
	versions := []string{"v1.11.0", "v2.4.0"}
	for _, v := range versions {

		c := buildCluster()
		c.Spec.KubernetesVersion = v
		b := assets.NewAssetBuilder(c, false)

		kcm := &KubeControllerManagerOptionsBuilder{
			OptionsContext: &OptionsContext{
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
	b := assets.NewAssetBuilder(c, false)

	kcm := &KubeControllerManagerOptionsBuilder{
		OptionsContext: &OptionsContext{
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

func Test_Build_KCM_Builder_CIDR_Mask_Size(t *testing.T) {
	grid := []struct {
		PodCIDR          string
		ClusterCIDR      string
		ExpectedMaskSize *int32
	}{
		{
			PodCIDR:          "100.64.1.0/11",
			ExpectedMaskSize: nil,
		},
		{
			PodCIDR:          "2001:DB8::/32",
			ExpectedMaskSize: fi.Int32(48),
		},
		{
			PodCIDR:          "2001:DB8::/65",
			ExpectedMaskSize: fi.Int32(81),
		},
		{
			PodCIDR:          "2001:DB8::/32",
			ClusterCIDR:      "2001:DB8::/65",
			ExpectedMaskSize: fi.Int32(81),
		},
		{
			PodCIDR:          "2001:DB8::/95",
			ExpectedMaskSize: fi.Int32(111),
		},
		{
			PodCIDR:          "2001:DB8::/96",
			ExpectedMaskSize: fi.Int32(112),
		},
		{
			PodCIDR:          "2001:DB8::/97",
			ExpectedMaskSize: fi.Int32(112),
		},
		{
			PodCIDR:          "2001:DB8::/98",
			ExpectedMaskSize: fi.Int32(113),
		},
		{
			PodCIDR:          "2001:DB8::/99",
			ExpectedMaskSize: fi.Int32(113),
		},
		{
			PodCIDR:          "2001:DB8::/100",
			ExpectedMaskSize: fi.Int32(114),
		},
	}
	for _, tc := range grid {
		t.Run(tc.PodCIDR+":"+tc.ClusterCIDR, func(t *testing.T) {
			c := buildCluster()
			b := assets.NewAssetBuilder(c, false)

			kcm := &KubeControllerManagerOptionsBuilder{
				OptionsContext: &OptionsContext{
					AssetBuilder: b,
				},
			}

			c.Spec.PodCIDR = tc.PodCIDR
			c.Spec.KubeControllerManager = &api.KubeControllerManagerConfig{
				ClusterCIDR: tc.ClusterCIDR,
			}

			err := kcm.BuildOptions(&c.Spec)
			require.NoError(t, err)

			assert.Equal(t, tc.ExpectedMaskSize, c.Spec.KubeControllerManager.NodeCIDRMaskSize)
		})
	}
}
