package cloudup

import (
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/api"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kubernetes/pkg/util/sets"
	"strings"
	"testing"
)

const MockAWSRegion = "us-mock-1"

func buildDefaultCluster(t *testing.T) *api.Cluster {
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
				m.Zone = fi.String(zone)
				etcd.Members = append(etcd.Members, m)
			}
			c.Spec.EtcdClusters = append(c.Spec.EtcdClusters, etcd)
		}
	}

	awsup.InstallMockAWSCloud(MockAWSRegion, "abcd")

	fullSpec, err := PopulateClusterSpec(c)
	if err != nil {
		t.Fatalf("error from PopulateClusterSpec: %v", err)
	}

	//// TODO: We should actually just specify the minimums here, and run in though the default logic
	//c.Cluster = &api.Cluster{}
	//c.Cluster.Name = "testcluster.mydomain.com"

	//c.InstanceGroups = append(c.InstanceGroups, buildNodeInstanceGroup("us-mock-1a"))
	//c.InstanceGroups = append(c.InstanceGroups, buildMasterInstanceGroup("us-mock-1a"))
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

func TestValidateFull_UpdatePolicy_Valid(t *testing.T) {
	c := buildDefaultCluster(t)
	c.Spec.UpdatePolicy = fi.String(api.UpdatePolicyExternal)
	expectNoErrorFromValidate(t, c)
}

func TestValidateFull_UpdatePolicy_Invalid(t *testing.T) {
	c := buildDefaultCluster(t)
	c.Spec.UpdatePolicy = fi.String("not-a-real-value")
	expectErrorFromValidate(t, c, "UpdatePolicy")
}

func Test_Validate_No_Classic_With_14(t *testing.T) {
	c := buildDefaultCluster(t)
	c.Spec.KubernetesVersion = "1.4.1"
	c.Spec.Networking = &api.NetworkingSpec{
		Classic: &api.ClassicNetworkingSpec{},
	}

	expectErrorFromValidate(t, c, "Spec.Networking")
}

func Test_Validate_Kubenet_With_14(t *testing.T) {
	c := buildDefaultCluster(t)
	c.Spec.KubernetesVersion = "1.4.1"
	c.Spec.Networking = &api.NetworkingSpec{
		Kubenet: &api.KubenetNetworkingSpec{},
	}

	expectNoErrorFromValidate(t, c)
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

func expectNoErrorFromValidate(t *testing.T, c *api.Cluster) {
	err := c.Validate(false)
	if err != nil {
		t.Fatalf("Unexpected error from Validate: %v", err)
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
//		{Name: "us-mock-1a", CIDR: "172.20.1.0/24"},
//		{Name: "us-mock-1b", CIDR: "172.20.2.0/24"},
//		{Name: "us-mock-1c", CIDR: "172.20.3.0/24"},
//		{Name: "us-mock-1d", CIDR: "172.20.4.0/24"},
//	}
//	c.InstanceGroups = append(c.InstanceGroups, buildNodeInstanceGroup("us-mock-1a"))
//	c.InstanceGroups = append(c.InstanceGroups, buildMasterInstanceGroup("us-mock-1a"))
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
