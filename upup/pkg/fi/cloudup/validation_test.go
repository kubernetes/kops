package cloudup

import (
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/api"
	"k8s.io/kops/upup/pkg/fi/vfs"
	"k8s.io/kubernetes/pkg/util/sets"
	"strings"
	"testing"
)

func buildInmemoryClusterRegistry() *api.ClusterRegistry {
	memfs := vfs.NewMemFSContext()
	memfs.MarkClusterReadable()
	basePath := vfs.NewMemFSPath(memfs, "test-statestore")
	return api.NewClusterRegistry(basePath)
}

func buildDefaultCluster(t *testing.T) *api.Cluster {
	registry := buildInmemoryClusterRegistry()

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

	fullSpec, err := PopulateClusterSpec(c, registry)
	if err != nil {
		t.Fatalf("error from PopulateClusterSpec: %v", err)
	}

	//// TODO: We should actually just specify the minimums here, and run in though the default logic
	//c.Cluster = &api.Cluster{}
	//c.Cluster.Name = "testcluster.mydomain.com"

	//c.InstanceGroups = append(c.InstanceGroups, buildNodeInstanceGroup("us-east-1a"))
	//c.InstanceGroups = append(c.InstanceGroups, buildMasterInstanceGroup("us-east-1a"))
	//c.SSHPublicKey = path.Join(os.Getenv("HOME"), ".ssh", "id_rsa.pub")
	//
	//c.Cluster.Spec.Kubelet = &api.KubeletConfig{}
	//c.Cluster.Spec.KubeControllerManager = &api.KubeControllerManagerConfig{}
	//c.Cluster.Spec.KubeDNS = &api.KubeDNSConfig{}
	//c.Cluster.Spec.KubeAPIServer = &api.KubeAPIServerConfig{}
	//c.Cluster.Spec.KubeProxy = &api.KubeProxyConfig{}
	//c.Cluster.Spec.Docker = &api.DockerConfig{}
	//
	//c.Cluster.Spec.NetworkCIDR = "172.20.0.0/16"
	//
	//c.Cluster.Spec.NonMasqueradeCIDR = "100.64.0.0/10"
	//c.Cluster.Spec.Kubelet.NonMasqueradeCIDR = c.Cluster.Spec.NonMasqueradeCIDR
	//
	//c.Cluster.Spec.ServiceClusterIPRange = "100.64.1.0/24"
	//c.Cluster.Spec.KubeAPIServer.ServiceClusterIPRange = c.Cluster.Spec.ServiceClusterIPRange
	//
	//c.Cluster.Spec.KubeDNS.ServerIP = "100.64.1.10"
	//c.Cluster.Spec.Kubelet.ClusterDNS = c.Cluster.Spec.KubeDNS.ServerIP
	//
	//c.Cluster.Spec.CloudProvider = "aws"
	//c.Cluster.Spec.Kubelet.CloudProvider = c.Cluster.Spec.CloudProvider
	//c.Cluster.Spec.KubeAPIServer.CloudProvider = c.Cluster.Spec.CloudProvider
	//c.Cluster.Spec.KubeControllerManager.CloudProvider = c.Cluster.Spec.CloudProvider
	//
	//c.Target = "dryrun"
	//
	//basePath := vfs.NewMemFSPath(memfs, "test-statestore")
	//c.ClusterRegistry = api.NewClusterRegistry(basePath)

	return fullSpec
}

func TestValidateFull_Default_Validates(t *testing.T) {
	c := buildDefaultCluster(t)
	err := c.Validate(false)
	if err != nil {
		glog.Infof("Cluster: %v", c)
		t.Fatalf("Validate gave unexpected error (strict=false): %v", err)
	}
	err = c.Validate(true)
	if err != nil {
		t.Fatalf("Validate gave unexpected error (strict=true): %v", err)
	}
}

func TestValidateFull_ClusterName_InvalidDNS_NoDot(t *testing.T) {
	c := buildDefaultCluster(t)
	c.Name = "test"
	expectErrorFromValidate(t, c, "DNS name")
}

func TestValidateFull_ClusterName_InvalidDNS_Invalid(t *testing.T) {
	c := buildDefaultCluster(t)
	c.Name = "test.-"
	expectErrorFromValidate(t, c, "DNS name")
}

func TestValidateFull_ClusterName_Required(t *testing.T) {
	c := buildDefaultCluster(t)
	c.Name = ""
	expectErrorFromValidate(t, c, "Name")
}

func expectErrorFromValidate(t *testing.T, c *api.Cluster, message string) {
	err := c.Validate(false)
	if err == nil {
		t.Fatalf("Expected error from Validate")
	}
	actualMessage := fmt.Sprintf("%v", err)
	if !strings.Contains(actualMessage, message) {
		t.Fatalf("Expected error %q, got %q", message, actualMessage)
	}
}

//
//import (
//	"fmt"
//	"k8s.io/kops/upup/pkg/api"
//	"k8s.io/kops/upup/pkg/fi/vfs"
//	k8sapi "k8s.io/kubernetes/pkg/api"
//	"os"
//	"path"
//	"strings"
//	"testing"
//)
//
//func buildDefaultCreateCluster() *CreateClusterCmd {
//	memfs := vfs.NewMemFSContext()
//	memfs.MarkClusterReadable()
//
//	c := &CreateClusterCmd{}
//
//	// TODO: We should actually just specify the minimums here, and run in though the default logic
//	c.Cluster = &api.Cluster{}
//	c.Cluster.Name = "testcluster.mydomain.com"
//	c.Cluster.Spec.Zones = []*api.ClusterZoneSpec{
//		{Name: "us-east-1a", CIDR: "172.20.1.0/24"},
//		{Name: "us-east-1b", CIDR: "172.20.2.0/24"},
//		{Name: "us-east-1c", CIDR: "172.20.3.0/24"},
//		{Name: "us-east-1d", CIDR: "172.20.4.0/24"},
//	}
//	c.InstanceGroups = append(c.InstanceGroups, buildNodeInstanceGroup("us-east-1a"))
//	c.InstanceGroups = append(c.InstanceGroups, buildMasterInstanceGroup("us-east-1a"))
//	c.SSHPublicKey = path.Join(os.Getenv("HOME"), ".ssh", "id_rsa.pub")
//
//	c.Cluster.Spec.Kubelet = &api.KubeletConfig{}
//	c.Cluster.Spec.KubeControllerManager = &api.KubeControllerManagerConfig{}
//	c.Cluster.Spec.KubeDNS = &api.KubeDNSConfig{}
//	c.Cluster.Spec.KubeAPIServer = &api.KubeAPIServerConfig{}
//	c.Cluster.Spec.KubeProxy = &api.KubeProxyConfig{}
//	c.Cluster.Spec.Docker = &api.DockerConfig{}
//
//	c.Cluster.Spec.NetworkCIDR = "172.20.0.0/16"
//
//	c.Cluster.Spec.NonMasqueradeCIDR = "100.64.0.0/10"
//	c.Cluster.Spec.Kubelet.NonMasqueradeCIDR = c.Cluster.Spec.NonMasqueradeCIDR
//
//	c.Cluster.Spec.ServiceClusterIPRange = "100.64.1.0/24"
//	c.Cluster.Spec.KubeAPIServer.ServiceClusterIPRange = c.Cluster.Spec.ServiceClusterIPRange
//
//	c.Cluster.Spec.KubeDNS.ServerIP = "100.64.1.10"
//	c.Cluster.Spec.Kubelet.ClusterDNS = c.Cluster.Spec.KubeDNS.ServerIP
//
//	c.Cluster.Spec.CloudProvider = "aws"
//	c.Cluster.Spec.Kubelet.CloudProvider = c.Cluster.Spec.CloudProvider
//	c.Cluster.Spec.KubeAPIServer.CloudProvider = c.Cluster.Spec.CloudProvider
//	c.Cluster.Spec.KubeControllerManager.CloudProvider = c.Cluster.Spec.CloudProvider
//
//	c.Target = "dryrun"
//
//	basePath := vfs.NewMemFSPath(memfs, "test-statestore")
//	c.ClusterRegistry = api.NewClusterRegistry(basePath)
//
//	return c
//}
//func expectErrorFromRun(t *testing.T, c *CreateClusterCmd, message string) {
//	err := c.Run()
//	if err == nil {
//		t.Fatalf("Expected error from run")
//	}
//	actualMessage := fmt.Sprintf("%v", err)
//	if !strings.Contains(actualMessage, message) {
//		t.Fatalf("Expected error %q, got %q", message, actualMessage)
//	}
//}
