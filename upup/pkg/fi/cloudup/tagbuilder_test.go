package cloudup

import (
	"k8s.io/kops/upup/pkg/api"
	"testing"
)

func TestBuildTags_CloudProvider_AWS(t *testing.T) {
	c := &api.Cluster{
		Spec: api.ClusterSpec{
			CloudProvider: "aws",
		},
	}

	tags, err := buildClusterTags(c)
	if err != nil {
		t.Fatalf("buildTags error: %v", err)
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

func stringSliceContains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if needle == s {
			return true
		}
	}
	return false
}
