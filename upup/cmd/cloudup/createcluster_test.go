package main

import (
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kube-deploy/upup/pkg/fi"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup"
	"k8s.io/kube-deploy/upup/pkg/fi/vfs"
	"testing"
)

// TODO: Refactor CreateClusterCmd into pkg/fi/cloudup

func buildDefaultCreateCluster() *CreateClusterCmd {
	var err error

	c := &CreateClusterCmd{}

	c.Config = &cloudup.CloudConfig{}
	c.Config.ClusterName = "testcluster.mydomain.com"
	c.Config.NodeZones = []string{"us-east-1a", "us-east-1b", "us-east-1c"}
	c.Config.MasterZones = c.Config.NodeZones
	c.SSHPublicKey = "~/.ssh/id_rsa.pub"

	c.Config.CloudProvider = "aws"

	dryrun := false
	c.StateStore, err = fi.NewVFSStateStore(vfs.NewFSPath("test-state"), dryrun)
	if err != nil {
		glog.Fatalf("error building state store: %v", err)
	}

	return c
}

func expectErrorFromRun(t *testing.T, c *CreateClusterCmd, message string) {
	err := c.Run()
	if err == nil {
		t.Fatalf("Expected error from run")
	}
	actualMessage := fmt.Sprintf("%v", err)
	if actualMessage != message {
		t.Fatalf("Expected error %q, got %q", message, actualMessage)
	}
}

func TestCreateCluster_DuplicateZones(t *testing.T) {
	c := buildDefaultCreateCluster()
	c.Config.NodeZones = []string{"us-east-1a", "us-east-1b", "us-east-1b"}
	c.Config.MasterZones = []string{"us-east-1a"}
	expectErrorFromRun(t, c, "NodeZones contained a duplicate value:  us-east-1b")
}

func TestCreateCluster_NoClusterName(t *testing.T) {
	c := buildDefaultCreateCluster()
	c.Config.ClusterName = ""
	expectErrorFromRun(t, c, "-name is required (e.g. mycluster.myzone.com)")
}

func TestCreateCluster_NoCloud(t *testing.T) {
	c := buildDefaultCreateCluster()
	c.Config.CloudProvider = ""
	expectErrorFromRun(t, c, "-cloud is required (e.g. aws, gce)")
}

func TestCreateCluster_ExtraMasterZone(t *testing.T) {
	c := buildDefaultCreateCluster()
	c.Config.NodeZones = []string{"us-east-1a", "us-east-1c"}
	c.Config.MasterZones = []string{"us-east-1a", "us-east-1b", "us-east-1c"}
	expectErrorFromRun(t, c, "All MasterZones must (currently) also be NodeZones")
}

func TestCreateCluster_NoMasterZones(t *testing.T) {
	c := buildDefaultCreateCluster()
	c.Config.MasterZones = []string{}
	expectErrorFromRun(t, c, "must specify at least one MasterZone")
}

func TestCreateCluster_NoNodeZones(t *testing.T) {
	c := buildDefaultCreateCluster()
	c.Config.NodeZones = []string{}
	expectErrorFromRun(t, c, "must specify at least one NodeZone")
}

func TestCreateCluster_RegionAsZone(t *testing.T) {
	c := buildDefaultCreateCluster()
	c.Config.NodeZones = []string{"us-east-1"}
	c.Config.MasterZones = c.Config.NodeZones
	expectErrorFromRun(t, c, "Region is not a recognized EC2 region: \"us-east-\" (check you have specified valid zones?)")
}

func TestCreateCluster_BadZone(t *testing.T) {
	c := buildDefaultCreateCluster()
	c.Config.NodeZones = []string{"us-east-1z"}
	c.Config.MasterZones = c.Config.NodeZones
	expectErrorFromRun(t, c, "Zone is not a recognized AZ: \"us-east-1z\" (check you have specified a valid zone?)")
}

func TestCreateCluster_MixedRegion(t *testing.T) {
	c := buildDefaultCreateCluster()
	c.Config.NodeZones = []string{"us-west-1a", "us-west-2b", "us-west-2c"}
	c.Config.MasterZones = c.Config.NodeZones
	expectErrorFromRun(t, c, "Clusters cannot span multiple regions")
}

func TestCreateCluster_EvenEtcdClusterSize(t *testing.T) {
	c := buildDefaultCreateCluster()
	c.Config.NodeZones = []string{"us-east-1a", "us-east-1b", "us-east-1c", "us-east-1d"}
	c.Config.MasterZones = c.Config.NodeZones
	expectErrorFromRun(t, c, "There should be an odd number of master-zones, for etcd's quorum.  Hint: Use -zone and -master-zone to declare node zones and master zones separately.")
}
