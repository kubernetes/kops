/*
Copyright 2016 The Kubernetes Authors.

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
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"testing"
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
			Topology:          &api.TopologySpec{
				Masters: api.TopologyPublic,
				Nodes:   api.TopologyPublic,
			},
		},
	}
}
func TestBuildTags_CloudProvider_AWS(t *testing.T) {

	c := buildCluster(nil)

	tags, err := buildCloudupTags(c)
	if err != nil {
		t.Fatalf("buildCloudupTags error: %v", err)
	}

	if _, found := tags["_aws"]; !found {
		t.Fatal("tag _aws not found")
	}

	if _, found := tags["_networking_cni"]; !found {
		t.Fatal("tag _networking_cni not found")
	}

	nodeUpTags, err := buildNodeupTags(api.InstanceGroupRoleNode, c, tags)
	if err != nil {
		t.Fatalf("buildNodeupTags error: %v", err)
	}

	if !stringSliceContains(nodeUpTags, "_aws") {
		t.Fatal("nodeUpTag _aws not found")
	}
}

func TestBuildTags_KubernetesVersions(t *testing.T) {
	grid := map[string]string{
		"1.3.7":         "_k8s_1_3",
		"v1.4.0-beta.8": "_k8s_1_4",
		"1.5.0":         "_k8s_1_5",
		"https://storage.googleapis.com/kubernetes-release-dev/ci/v1.4.0-alpha.2.677+ea69570f61af8e/": "_k8s_1_4",
	}
	for version, tag := range grid {
		c := buildCluster(ClusterParams{KubernetesVersion: version})

		tags, err := buildCloudupTags(c)
		if err != nil {
			t.Fatalf("buildCloudupTags error: %v", err)
		}

		if _, found := tags[tag]; !found {
			t.Fatalf("tag %q not found for %q: %v", tag, version, tags)
		}
	}
}

func TestBuildTags_UpdatePolicy_Nil(t *testing.T) {
	c := buildCluster(nil)

	tags, err := buildCloudupTags(c)
	if err != nil {
		t.Fatalf("buildCloudupTags error: %v", err)
	}

	nodeUpTags, err := buildNodeupTags(api.InstanceGroupRoleNode, c, tags)
	if err != nil {
		t.Fatalf("buildNodeupTags error: %v", err)
	}

	if !stringSliceContains(nodeUpTags, "_automatic_upgrades") {
		t.Fatal("nodeUpTag _automatic_upgrades not found")
	}
}

func TestBuildTags_UpdatePolicy_None(t *testing.T) {
	c := buildCluster(ClusterParams{CloudProvider: "aws", UpdatePolicy: api.UpdatePolicyExternal})

	tags, err := buildCloudupTags(c)
	if err != nil {
		t.Fatalf("buildTags error: %v", err)
	}

	nodeUpTags, err := buildNodeupTags(api.InstanceGroupRoleNode, c, tags)
	if err != nil {
		t.Fatalf("buildNodeupTags error: %v", err)
	}

	if stringSliceContains(nodeUpTags, "_automatic_upgrades") {
		t.Fatal("nodeUpTag _automatic_upgrades found unexpectedly")
	}
}

func stringSliceContains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if needle == s {
			return true
		}
	}
	return false
}
