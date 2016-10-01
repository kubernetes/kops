package cloudup

import (
	"k8s.io/kops/upup/pkg/api"
	"k8s.io/kops/upup/pkg/fi"
	"testing"
)

func TestBuildTags_CloudProvider_AWS(t *testing.T) {
	c := &api.Cluster{
		Spec: api.ClusterSpec{
			CloudProvider:     "aws",
			KubernetesVersion: "v1.3.5",
		},
	}

	tags, err := buildCloudupTags(c)
	if err != nil {
		t.Fatalf("buildCloudupTags error: %v", err)
	}

	if _, found := tags["_aws"]; !found {
		t.Fatalf("tag _aws not found")
	}

	nodeUpTags, err := buildNodeupTags(api.InstanceGroupRoleNode, c, tags)
	if err != nil {
		t.Fatalf("buildNodeupTags error: %v", err)
	}

	if !stringSliceContains(nodeUpTags, "_aws") {
		t.Fatalf("nodeUpTag _aws not found")
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
		c := &api.Cluster{
			Spec: api.ClusterSpec{
				CloudProvider:     "aws",
				KubernetesVersion: version,
			},
		}

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
	c := &api.Cluster{
		Spec: api.ClusterSpec{
			CloudProvider:     "aws",
			KubernetesVersion: "v1.3.5",
			UpdatePolicy:      nil,
		},
	}

	tags, err := buildCloudupTags(c)
	if err != nil {
		t.Fatalf("buildCloudupTags error: %v", err)
	}

	nodeUpTags, err := buildNodeupTags(api.InstanceGroupRoleNode, c, tags)
	if err != nil {
		t.Fatalf("buildNodeupTags error: %v", err)
	}

	if !stringSliceContains(nodeUpTags, "_automatic_upgrades") {
		t.Fatalf("nodeUpTag _automatic_upgrades not found")
	}
}

func TestBuildTags_UpdatePolicy_None(t *testing.T) {
	c := &api.Cluster{
		Spec: api.ClusterSpec{
			CloudProvider:     "aws",
			KubernetesVersion: "v1.3.5",
			UpdatePolicy:      fi.String(api.UpdatePolicyExternal),
		},
	}

	tags, err := buildCloudupTags(c)
	if err != nil {
		t.Fatalf("buildTags error: %v", err)
	}

	nodeUpTags, err := buildNodeupTags(api.InstanceGroupRoleNode, c, tags)
	if err != nil {
		t.Fatalf("buildNodeupTags error: %v", err)
	}

	if stringSliceContains(nodeUpTags, "_automatic_upgrades") {
		t.Fatalf("nodeUpTag _automatic_upgrades found unexpectedly")
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
