package cloudup

import (
	"fmt"
	"k8s.io/kops/upup/pkg/api"
	"strings"
	"testing"
)

func buildMinimalNodeInstanceGroup(zones ...string) *api.InstanceGroup {
	g := &api.InstanceGroup{}
	g.Name = "nodes"
	g.Spec.Role = api.InstanceGroupRoleNode
	g.Spec.Zones = zones

	return g
}

func buildMinimalMasterInstanceGroup(zones ...string) *api.InstanceGroup {
	g := &api.InstanceGroup{}
	g.Name = "master"
	g.Spec.Role = api.InstanceGroupRoleMaster
	g.Spec.Zones = zones

	return g
}

func TestPopulateInstanceGroup_Name_Required(t *testing.T) {
	cluster := buildMinimalCluster()
	g := buildMinimalNodeInstanceGroup()
	g.Name = ""

	expectErrorFromPopulateInstanceGroup(t, cluster, g, "Name")
}

func TestPopulateInstanceGroup_Role_Required(t *testing.T) {
	cluster := buildMinimalCluster()
	g := buildMinimalNodeInstanceGroup()
	g.Spec.Role = ""

	expectErrorFromPopulateInstanceGroup(t, cluster, g, "Role")
}

func expectErrorFromPopulateInstanceGroup(t *testing.T, cluster *api.Cluster, g *api.InstanceGroup, message string) {
	_, err := PopulateInstanceGroupSpec(cluster, g)
	if err == nil {
		t.Fatalf("Expected error from PopulateInstanceGroup")
	}
	actualMessage := fmt.Sprintf("%v", err)
	if !strings.Contains(actualMessage, message) {
		t.Fatalf("Expected error %q, got %q", message, actualMessage)
	}
}
