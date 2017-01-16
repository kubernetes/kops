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

	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
)

type ClusterParams struct {
	CloudProvider     string
	KubernetesVersion string
	UpdatePolicy      string
}

func buildCluster(clusterArgs interface{}) *api.Cluster {

	if clusterArgs == nil {
		clusterArgs = ClusterParams{CloudProvider: "aws", KubernetesVersion: "1.4.0"}
	}

	cParams := clusterArgs.(ClusterParams)

	if cParams.CloudProvider == "" {
		cParams.CloudProvider = "aws"
	}

	if cParams.KubernetesVersion == "" {
		cParams.KubernetesVersion = "v1.4.0"
	}

	networking := &api.NetworkingSpec{
		CNI: &api.CNINetworkingSpec{},
	}

	return &api.Cluster{
		Spec: api.ClusterSpec{
			CloudProvider:     cParams.CloudProvider,
			KubernetesVersion: cParams.KubernetesVersion,
			Networking:        networking,
			UpdatePolicy:      fi.String(cParams.UpdatePolicy),
			Topology: &api.TopologySpec{
				Masters: api.TopologyPublic,
				Nodes:   api.TopologyPublic,
			},
		},
	}
}

func Test_Build_KCM_Builder_Lower_Version(t *testing.T) {
	c := buildCluster(nil)

	kcm := &KubeControllerManagerOptionsBuilder{
		Context: &OptionsContext{
			Cluster: c,
		},
	}

	spec := c.Spec
	err := kcm.BuildOptions(&spec)

	if err != nil {
		t.Fatalf("k-c-m builder errors: %v", err)
	}

	if spec.KubeControllerManager.AttachDetachReconcileSyncPeriod != nil {
		t.Fatalf("k-c-m builder cannot be set for k8s %s", spec.KubernetesVersion)
	}

}

func Test_Build_KCM_Builder_High_Enough_Version(t *testing.T) {
	c := buildCluster(ClusterParams{KubernetesVersion: "1.4.8"})

	kcm := &KubeControllerManagerOptionsBuilder{
		Context: &OptionsContext{
			Cluster: c,
		},
	}

	spec := c.Spec
	err := kcm.BuildOptions(&spec)

	if err != nil {
		t.Fatalf("k-c-m builder errors: %v", err)
	}

	if spec.KubeControllerManager.AttachDetachReconcileSyncPeriod.Duration != time.Minute {
		t.Fatalf("k-c-m builder should be set to 1m - %s", spec.KubeControllerManager.AttachDetachReconcileSyncPeriod.Duration.String())
	}

}
