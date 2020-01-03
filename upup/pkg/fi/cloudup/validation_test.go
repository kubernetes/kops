/*
Copyright 2019 The Kubernetes Authors.

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
	"fmt"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog"
	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/validation"
	"k8s.io/kops/upup/pkg/fi"
)

const MockAWSRegion = "us-mock-1"

func buildDefaultCluster(t *testing.T) *kopsapi.Cluster {
	c := buildMinimalCluster()

	err := PerformAssignments(c)
	if err != nil {
		t.Fatalf("error from PerformAssignments: %v", err)
	}

	if len(c.Spec.EtcdClusters) == 0 {
		zones := sets.NewString()
		for _, z := range c.Spec.Subnets {
			zones.Insert(z.Zone)
		}
		etcdZones := zones.List()

		for _, etcdCluster := range EtcdClusters {
			etcd := &kopsapi.EtcdClusterSpec{}
			etcd.Name = etcdCluster
			for _, zone := range etcdZones {
				m := &kopsapi.EtcdMemberSpec{}
				m.Name = zone
				m.InstanceGroup = fi.String(zone)
				etcd.Members = append(etcd.Members, m)
			}
			c.Spec.EtcdClusters = append(c.Spec.EtcdClusters, etcd)
		}
	}

	fullSpec, err := mockedPopulateClusterSpec(c)
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
	if err := validation.ValidateCluster(c, false); err != nil {
		klog.Infof("Cluster: %v", c)
		t.Fatalf("Validate gave unexpected error (strict=false): %v", err)
	}
	if err := validation.ValidateCluster(c, true); err != nil {
		t.Fatalf("Validate gave unexpected error (strict=true): %v", err)
	}
}

func TestValidateFull_ClusterName_InvalidDNS_NoDot(t *testing.T) {
	c := buildDefaultCluster(t)
	c.ObjectMeta.Name = "test"
	expectErrorFromValidate(t, c, "DNS name")
}

func TestValidateFull_ClusterName_InvalidDNS_Invalid(t *testing.T) {
	c := buildDefaultCluster(t)
	c.ObjectMeta.Name = "test.-"
	expectErrorFromValidate(t, c, "DNS name")
}

func TestValidateFull_ClusterName_Required(t *testing.T) {
	c := buildDefaultCluster(t)
	c.ObjectMeta.Name = ""
	expectErrorFromValidate(t, c, "Name")
}

func TestValidateFull_UpdatePolicy_Valid(t *testing.T) {
	c := buildDefaultCluster(t)
	c.Spec.UpdatePolicy = fi.String(kopsapi.UpdatePolicyExternal)
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
	c.Spec.Networking = &kopsapi.NetworkingSpec{
		Classic: &kopsapi.ClassicNetworkingSpec{},
	}

	expectErrorFromValidate(t, c, "spec.Networking")
}

func Test_Validate_Kubenet_With_14(t *testing.T) {
	c := buildDefaultCluster(t)
	c.Spec.KubernetesVersion = "1.4.1"
	c.Spec.Networking = &kopsapi.NetworkingSpec{
		Kubenet: &kopsapi.KubenetNetworkingSpec{},
	}

	expectNoErrorFromValidate(t, c)
}

func TestValidate_ClusterName_Import(t *testing.T) {
	c := buildDefaultCluster(t)

	// When we import a cluster, it likely won't have a valid name until we convert it
	c.ObjectMeta.Annotations = make(map[string]string)
	c.ObjectMeta.Annotations[kopsapi.AnnotationNameManagement] = kopsapi.AnnotationValueManagementImported
	c.ObjectMeta.Name = "kubernetes"

	expectNoErrorFromValidate(t, c)
}

func TestValidate_ContainerRegistry_and_ContainerProxy_exclusivity(t *testing.T) {
	c := buildDefaultCluster(t)

	assets := new(kopsapi.Assets)
	c.Spec.Assets = assets

	expectNoErrorFromValidate(t, c)

	registry := "https://registry.example.com/"
	c.Spec.Assets.ContainerRegistry = &registry
	expectNoErrorFromValidate(t, c)

	proxy := "https://proxy.example.com/"
	c.Spec.Assets.ContainerProxy = &proxy
	expectErrorFromValidate(t, c, "ContainerProxy cannot be used in conjunction with ContainerRegistry")

	c.Spec.Assets.ContainerRegistry = nil
	expectNoErrorFromValidate(t, c)

}

func expectErrorFromValidate(t *testing.T, c *kopsapi.Cluster, message string) {
	err := validation.ValidateCluster(c, false)
	if err == nil {
		t.Fatalf("Expected error from Validate")
	}
	actualMessage := fmt.Sprintf("%v", err)
	if !strings.Contains(actualMessage, message) {
		t.Fatalf("Expected error %q, got %q", message, actualMessage)
	}
}

func expectNoErrorFromValidate(t *testing.T, c *kopsapi.Cluster) {
	err := validation.ValidateCluster(c, false)
	if err != nil {
		t.Fatalf("Unexpected error from Validate: %v", err)
	}
}
