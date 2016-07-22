package cloudup

import (
	"fmt"
	"k8s.io/kops/upup/pkg/api"
	"k8s.io/kubernetes/pkg/util/sets"
	"strings"
	"testing"
)

func buildMinimalCluster() *api.Cluster {
	c := &api.Cluster{}
	c.Name = "testcluster.test.com"
	c.Spec.Zones = []*api.ClusterZoneSpec{
		{Name: "us-east-1a", CIDR: "172.20.1.0/24"},
		{Name: "us-east-1b", CIDR: "172.20.2.0/24"},
		{Name: "us-east-1c", CIDR: "172.20.3.0/24"},
	}
	c.Spec.NetworkCIDR = "172.20.0.0/16"
	c.Spec.NonMasqueradeCIDR = "100.64.0.0/10"
	c.Spec.CloudProvider = "aws"

	// Required to stop a call to cloud provider
	// TODO: Mock cloudprovider
	c.Spec.DNSZone = "test.com"

	return c
}

func TestPopulateCluster_Default_NoError(t *testing.T) {
	c := buildMinimalCluster()

	err := c.PerformAssignments()
	if err != nil {
		t.Fatalf("error from PerformAssignments: %v", err)
	}

	if len(c.Spec.EtcdClusters) == 0 {
		zones := sets.NewString()
		for _, z := range c.Spec.Zones {
			zones.Insert(z.Name)
		}
		etcdZones := zones.List()

		for _, etcdCluster := range EtcdClusters {
			etcd := &api.EtcdClusterSpec{}
			etcd.Name = etcdCluster
			for _, zone := range etcdZones {
				m := &api.EtcdMemberSpec{}
				m.Name = zone
				m.Zone = zone
				etcd.Members = append(etcd.Members, m)
			}
			c.Spec.EtcdClusters = append(c.Spec.EtcdClusters, etcd)
		}
	}

	registry := buildInmemoryClusterRegistry()
	_, err = PopulateClusterSpec(c, registry)
	if err != nil {
		t.Fatalf("Unexpected error from PopulateCluster: %v", err)
	}
}

func TestPopulateCluster_Custom_CIDR(t *testing.T) {
	c := buildMinimalCluster()
	c.Spec.NetworkCIDR = "172.20.2.0/24"
	c.Spec.Zones = []*api.ClusterZoneSpec{
		{Name: "us-east-1a", CIDR: "172.20.2.0/27"},
		{Name: "us-east-1b", CIDR: "172.20.2.32/27"},
		{Name: "us-east-1c", CIDR: "172.20.2.64/27"},
	}

	err := c.PerformAssignments()
	if err != nil {
		t.Fatalf("error from PerformAssignments: %v", err)
	}

	if len(c.Spec.EtcdClusters) == 0 {
		zones := sets.NewString()
		for _, z := range c.Spec.Zones {
			zones.Insert(z.Name)
		}
		etcdZones := zones.List()

		for _, etcdCluster := range EtcdClusters {
			etcd := &api.EtcdClusterSpec{}
			etcd.Name = etcdCluster
			for _, zone := range etcdZones {
				m := &api.EtcdMemberSpec{}
				m.Name = zone
				m.Zone = zone
				etcd.Members = append(etcd.Members, m)
			}
			c.Spec.EtcdClusters = append(c.Spec.EtcdClusters, etcd)
		}
	}

	registry := buildInmemoryClusterRegistry()
	full, err := PopulateClusterSpec(c, registry)
	if err != nil {
		t.Fatalf("Unexpected error from PopulateCluster: %v", err)
	}
	if full.Spec.NetworkCIDR != "172.20.2.0/24" {
		t.Fatalf("Unexpected NetworkCIDR: %v", full.Spec.NetworkCIDR)
	}
}

func TestPopulateCluster_Name_Required(t *testing.T) {
	c := buildMinimalCluster()
	c.Name = ""

	expectErrorFromPopulateCluster(t, c, "Name")
}

func TestPopulateCluster_Zone_Required(t *testing.T) {
	c := buildMinimalCluster()
	c.Spec.Zones = nil

	expectErrorFromPopulateCluster(t, c, "Zone")
}

func TestPopulateCluster_NetworkCIDR_Required(t *testing.T) {
	c := buildMinimalCluster()
	c.Spec.NetworkCIDR = ""

	expectErrorFromPopulateCluster(t, c, "NetworkCIDR")
}

func TestPopulateCluster_NonMasqueradeCIDR_Required(t *testing.T) {
	c := buildMinimalCluster()
	c.Spec.NonMasqueradeCIDR = ""

	expectErrorFromPopulateCluster(t, c, "NonMasqueradeCIDR")
}

func TestPopulateCluster_CloudProvider_Required(t *testing.T) {
	c := buildMinimalCluster()
	c.Spec.CloudProvider = ""

	expectErrorFromPopulateCluster(t, c, "CloudProvider")
}

func expectErrorFromPopulateCluster(t *testing.T, c *api.Cluster, message string) {
	registry := buildInmemoryClusterRegistry()
	_, err := PopulateClusterSpec(c, registry)
	if err == nil {
		t.Fatalf("Expected error from PopulateCluster")
	}
	actualMessage := fmt.Sprintf("%v", err)
	if !strings.Contains(actualMessage, message) {
		t.Fatalf("Expected error %q, got %q", message, actualMessage)
	}
}
