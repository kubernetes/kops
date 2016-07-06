package cloudup

import (
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/api"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/vfs"
	k8sapi "k8s.io/kubernetes/pkg/api"
	"os"
	"path"
	"strings"
	"testing"
)

func buildDefaultCreateCluster() *CreateClusterCmd {
	var err error

	memfs := vfs.NewMemFSContext()
	memfs.MarkClusterReadable()

	c := &CreateClusterCmd{}

	// TODO: We should actually just specify the minimums here, and run in though the default logic
	c.Cluster = &api.Cluster{}
	c.Cluster.Name = "testcluster.mydomain.com"
	c.Cluster.Spec.Zones = []*api.ClusterZoneSpec{
		{Name: "us-east-1a", CIDR: "172.20.1.0/24"},
		{Name: "us-east-1b", CIDR: "172.20.2.0/24"},
		{Name: "us-east-1c", CIDR: "172.20.3.0/24"},
		{Name: "us-east-1d", CIDR: "172.20.4.0/24"},
	}
	c.InstanceGroups = append(c.InstanceGroups, buildNodeInstanceGroup("us-east-1a"))
	c.InstanceGroups = append(c.InstanceGroups, buildMasterInstanceGroup("us-east-1a"))
	c.SSHPublicKey = path.Join(os.Getenv("HOME"), ".ssh", "id_rsa.pub")

	c.Cluster.Spec.Kubelet = &api.KubeletConfig{}
	c.Cluster.Spec.KubeControllerManager = &api.KubeControllerManagerConfig{}
	c.Cluster.Spec.KubeDNS = &api.KubeDNSConfig{}
	c.Cluster.Spec.KubeAPIServer = &api.KubeAPIServerConfig{}
	c.Cluster.Spec.KubeProxy = &api.KubeProxyConfig{}
	c.Cluster.Spec.Docker = &api.DockerConfig{}

	c.Cluster.Spec.NetworkCIDR = "172.20.0.0/16"

	c.Cluster.Spec.NonMasqueradeCIDR = "100.64.0.0/10"
	c.Cluster.Spec.Kubelet.NonMasqueradeCIDR = c.Cluster.Spec.NonMasqueradeCIDR

	c.Cluster.Spec.ServiceClusterIPRange = "100.64.1.0/24"
	c.Cluster.Spec.KubeAPIServer.ServiceClusterIPRange = c.Cluster.Spec.ServiceClusterIPRange

	c.Cluster.Spec.KubeDNS.ServerIP = "100.64.1.10"
	c.Cluster.Spec.Kubelet.ClusterDNS = c.Cluster.Spec.KubeDNS.ServerIP

	c.Cluster.Spec.CloudProvider = "aws"
	c.Cluster.Spec.Kubelet.CloudProvider = c.Cluster.Spec.CloudProvider
	c.Cluster.Spec.KubeAPIServer.CloudProvider = c.Cluster.Spec.CloudProvider
	c.Cluster.Spec.KubeControllerManager.CloudProvider = c.Cluster.Spec.CloudProvider

	c.Target = "dryrun"

	dryrun := false
	c.StateStore, err = fi.NewVFSStateStore(vfs.NewMemFSPath(memfs, "test-statestore"), c.Cluster.Name, dryrun)
	if err != nil {
		glog.Fatalf("error building state store: %v", err)
	}

	return c
}

func buildNodeInstanceGroup(zones ...string) *api.InstanceGroup {
	g := &api.InstanceGroup{
		ObjectMeta: k8sapi.ObjectMeta{Name: "nodes-" + strings.Join(zones, "-")},
		Spec: api.InstanceGroupSpec{
			Role:  api.InstanceGroupRoleNode,
			Zones: zones,
		},
	}
	return g
}

func buildMasterInstanceGroup(zones ...string) *api.InstanceGroup {
	g := &api.InstanceGroup{
		ObjectMeta: k8sapi.ObjectMeta{Name: "master-" + strings.Join(zones, "-")},
		Spec: api.InstanceGroupSpec{
			Role:  api.InstanceGroupRoleMaster,
			Zones: zones,
		},
	}
	return g
}

func expectErrorFromRun(t *testing.T, c *CreateClusterCmd, message string) {
	err := c.Run()
	if err == nil {
		t.Fatalf("Expected error from run")
	}
	actualMessage := fmt.Sprintf("%v", err)
	if !strings.Contains(actualMessage, message) {
		t.Fatalf("Expected error %q, got %q", message, actualMessage)
	}
}

func TestCreateCluster_DuplicateZones(t *testing.T) {
	c := buildDefaultCreateCluster()
	c.Cluster.Spec.Zones = []*api.ClusterZoneSpec{
		{Name: "us-east-1a", CIDR: "172.20.1.0/24"},
		{Name: "us-east-1a", CIDR: "172.20.2.0/24"},
	}
	expectErrorFromRun(t, c, "Zones contained a duplicate value: us-east-1a")
}

func TestCreateCluster_NoClusterName(t *testing.T) {
	c := buildDefaultCreateCluster()
	c.Cluster.Name = ""
	expectErrorFromRun(t, c, "ClusterName is required")
}

func TestCreateCluster_NoCloud(t *testing.T) {
	c := buildDefaultCreateCluster()
	c.Cluster.Spec.CloudProvider = ""
	expectErrorFromRun(t, c, "-cloud is required (e.g. aws, gce)")
}

func TestCreateCluster_ExtraMasterZone(t *testing.T) {
	c := buildDefaultCreateCluster()
	c.Cluster.Spec.Zones = []*api.ClusterZoneSpec{
		{Name: "us-east-1a", CIDR: "172.20.1.0/24"},
		{Name: "us-east-1b", CIDR: "172.20.2.0/24"},
	}
	c.InstanceGroups = []*api.InstanceGroup{}
	c.InstanceGroups = append(c.InstanceGroups, buildNodeInstanceGroup("us-east-1a", "us-east-1b"))
	c.InstanceGroups = append(c.InstanceGroups, buildMasterInstanceGroup("us-east-1a", "us-east-1b", "us-east-1c"))
	expectErrorFromRun(t, c, "is not configured as a Zone in the cluster")
}

func TestCreateCluster_NoMasterZones(t *testing.T) {
	c := buildDefaultCreateCluster()
	c.InstanceGroups = []*api.InstanceGroup{}
	c.InstanceGroups = append(c.InstanceGroups, buildNodeInstanceGroup("us-east-1a"))
	expectErrorFromRun(t, c, "must configure at least one Master InstanceGroup")
}

func TestCreateCluster_NoNodeZones(t *testing.T) {
	c := buildDefaultCreateCluster()
	c.InstanceGroups = []*api.InstanceGroup{}
	c.InstanceGroups = append(c.InstanceGroups, buildMasterInstanceGroup("us-east-1a"))
	expectErrorFromRun(t, c, "must configure at least one Node InstanceGroup")
}

func TestCreateCluster_RegionAsZone(t *testing.T) {
	c := buildDefaultCreateCluster()
	c.Cluster.Spec.Zones = []*api.ClusterZoneSpec{
		{Name: "us-east-1", CIDR: "172.20.1.0/24"},
	}
	c.InstanceGroups = []*api.InstanceGroup{
		buildNodeInstanceGroup("us-east-1"),
		buildMasterInstanceGroup("us-east-1"),
	}
	expectErrorFromRun(t, c, "Region is not a recognized EC2 region: \"us-east-\" (check you have specified valid zones?)")
}

func TestCreateCluster_NotIncludedZone(t *testing.T) {
	c := buildDefaultCreateCluster()
	c.InstanceGroups = []*api.InstanceGroup{
		buildNodeInstanceGroup("us-east-1e"),
		buildMasterInstanceGroup("us-east-1a"),
	}
	expectErrorFromRun(t, c, "not configured as a Zone in the cluster")
}

func TestCreateCluster_BadZone(t *testing.T) {
	c := buildDefaultCreateCluster()
	c.Cluster.Spec.Zones = []*api.ClusterZoneSpec{
		{Name: "us-east-1z", CIDR: "172.20.1.0/24"},
	}
	c.InstanceGroups = []*api.InstanceGroup{
		buildNodeInstanceGroup("us-east-1z"),
		buildMasterInstanceGroup("us-east-1z"),
	}
	expectErrorFromRun(t, c, "Zone is not a recognized AZ: \"us-east-1z\" (check you have specified a valid zone?)")
}

func TestCreateCluster_MixedRegion(t *testing.T) {
	c := buildDefaultCreateCluster()
	c.Cluster.Spec.Zones = []*api.ClusterZoneSpec{
		{Name: "us-east-1a", CIDR: "172.20.1.0/24"},
		{Name: "us-west-1b", CIDR: "172.20.2.0/24"},
	}
	c.InstanceGroups = []*api.InstanceGroup{
		buildNodeInstanceGroup("us-east-1a", "us-west-1b"),
		buildMasterInstanceGroup("us-east-1a"),
	}
	expectErrorFromRun(t, c, "Clusters cannot span multiple regions")
}

func TestCreateCluster_EvenEtcdClusterSize(t *testing.T) {
	c := buildDefaultCreateCluster()
	c.InstanceGroups = []*api.InstanceGroup{}
	c.InstanceGroups = append(c.InstanceGroups, buildNodeInstanceGroup("us-east-1a"))
	c.InstanceGroups = append(c.InstanceGroups, buildMasterInstanceGroup("us-east-1a", "us-east-1b", "us-east-1c", "us-east-1d"))
	c.Cluster.Spec.EtcdClusters = []*api.EtcdClusterSpec{
		{
			Name: "main",
			Members: []*api.EtcdMemberSpec{
				{Name: "us-east-1a", Zone: "us-east-1a"},
				{Name: "us-east-1b", Zone: "us-east-1b"},
			},
		},
	}
	expectErrorFromRun(t, c, "There should be an odd number of master-zones, for etcd's quorum.  Hint: Use --zones and --master-zones to declare node zones and master zones separately.")
}
