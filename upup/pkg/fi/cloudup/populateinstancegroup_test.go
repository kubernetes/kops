package cloudup

import (
	"fmt"
	api "k8s.io/kops/pkg/apis/kops"
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

	channel := &api.Channel{}

	expectErrorFromPopulateInstanceGroup(t, cluster, g, channel, "Name")
}

func TestPopulateInstanceGroup_Role_Required(t *testing.T) {
	cluster := buildMinimalCluster()
	g := buildMinimalNodeInstanceGroup()
	g.Spec.Role = ""

	channel := &api.Channel{}

	expectErrorFromPopulateInstanceGroup(t, cluster, g, channel, "Role")
}

func expectErrorFromPopulateInstanceGroup(t *testing.T, cluster *api.Cluster, g *api.InstanceGroup, channel *api.Channel, message string) {
	_, err := PopulateInstanceGroupSpec(cluster, g, channel)
	if err == nil {
		t.Fatalf("Expected error from PopulateInstanceGroup")
	}
	actualMessage := fmt.Sprintf("%v", err)
	if !strings.Contains(actualMessage, message) {
		t.Fatalf("Expected error %q, got %q", message, actualMessage)
	}
}
