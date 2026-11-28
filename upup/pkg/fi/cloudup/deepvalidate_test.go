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

	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/validation"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/util/pkg/vfs"
)

func TestDeepValidate_OK(t *testing.T) {
	c := buildDefaultCluster(t)
	var groups []*kopsapi.InstanceGroup
	for _, subnet := range c.Spec.Networking.Subnets {
		groups = append(groups, buildMinimalMasterInstanceGroup(subnet.Name))
		groups = append(groups, buildMinimalNodeInstanceGroup(subnet.Name))
	}
	err := validation.DeepValidate(c, groups, true, vfs.Context, nil)
	if err != nil {
		t.Fatalf("Expected no error from DeepValidate, got %v", err)
	}
}

func buildMinimalInstanceGroup(name string, role kopsapi.InstanceGroupRole, subnet string) *kopsapi.InstanceGroup {
	g := &kopsapi.InstanceGroup{}
	g.ObjectMeta.Name = name + "-" + subnet
	g.Spec.Role = role
	var one int32 = 1
	g.Spec.MinSize = &one
	g.Spec.MaxSize = &one
	g.Spec.Image = "my-image"
	g.Spec.Subnets = []string{subnet}
	return g
}

func TestDeepValidate_SplitControlPlane_OK(t *testing.T) {
	c := buildDefaultCluster(t)
	var groups []*kopsapi.InstanceGroup
	for _, subnet := range c.Spec.Networking.Subnets {
		groups = append(groups, buildMinimalInstanceGroup("apiserver", kopsapi.InstanceGroupRoleAPIServer, subnet.Name))
		groups = append(groups, buildMinimalInstanceGroup("etcd", kopsapi.InstanceGroupRoleEtcd, subnet.Name))
		groups = append(groups, buildMinimalInstanceGroup("kcm", kopsapi.InstanceGroupRoleKubeControllerManager, subnet.Name))
		groups = append(groups, buildMinimalInstanceGroup("scheduler", kopsapi.InstanceGroupRoleScheduler, subnet.Name))
		groups = append(groups, buildMinimalNodeInstanceGroup(subnet.Name))
	}
	err := validation.DeepValidate(c, groups, true, vfs.Context, nil)
	if err != nil {
		t.Fatalf("Expected no error from DeepValidate for split control plane, got %v", err)
	}
}

func TestDeepValidate_SplitControlPlane_MutualExclusion(t *testing.T) {
	c := buildDefaultCluster(t)
	var groups []*kopsapi.InstanceGroup
	groups = append(groups, buildMinimalMasterInstanceGroup("subnet-us-test-1a"))
	groups = append(groups, buildMinimalInstanceGroup("etcd", kopsapi.InstanceGroupRoleEtcd, "subnet-us-test-1a"))
	groups = append(groups, buildMinimalNodeInstanceGroup("subnet-us-test-1a"))
	expectErrorFromDeepValidate(t, c, groups, "cannot have both ControlPlane/Master InstanceGroups and split control plane InstanceGroups")
}

func TestDeepValidate_SplitControlPlane_MissingComponent(t *testing.T) {
	c := buildDefaultCluster(t)
	var groups []*kopsapi.InstanceGroup
	groups = append(groups, buildMinimalInstanceGroup("apiserver", kopsapi.InstanceGroupRoleAPIServer, "subnet-us-test-1a"))
	groups = append(groups, buildMinimalInstanceGroup("etcd", kopsapi.InstanceGroupRoleEtcd, "subnet-us-test-1a"))
	groups = append(groups, buildMinimalInstanceGroup("kcm", kopsapi.InstanceGroupRoleKubeControllerManager, "subnet-us-test-1a"))
	// Missing Scheduler!
	groups = append(groups, buildMinimalNodeInstanceGroup("subnet-us-test-1a"))
	expectErrorFromDeepValidate(t, c, groups, "must configure either a ControlPlane InstanceGroup or separate APIServer, Etcd, KubeControllerManager, and Scheduler InstanceGroups")
}

// TestDeepValidate_CompositeRole_RequiresGCE checks the cloud gate on composite roles. Only GCE
// wires up the per-role tags, load balancer membership and addressing a split control plane
// needs, so elsewhere a composite role has to be refused rather than silently ignored.
func TestDeepValidate_CompositeRole_RequiresGCE(t *testing.T) {
	featureflag.ParseFlags("+ExperimentalRoles")
	defer featureflag.ParseFlags("-ExperimentalRoles")

	c := buildDefaultCluster(t)
	var groups []*kopsapi.InstanceGroup
	// A complete split control plane, with the scheduler sharing a group with an API server.
	groups = append(groups, buildMinimalInstanceGroup("apiserver", "APIServer,Scheduler", "subnet-us-test-1a"))
	groups = append(groups, buildMinimalInstanceGroup("etcd", kopsapi.InstanceGroupRoleEtcd, "subnet-us-test-1a"))
	groups = append(groups, buildMinimalInstanceGroup("kcm", kopsapi.InstanceGroupRoleKubeControllerManager, "subnet-us-test-1a"))
	groups = append(groups, buildMinimalNodeInstanceGroup("subnet-us-test-1a"))
	expectErrorFromDeepValidate(t, c, groups, "combines several roles, which is only supported on GCE")
}

// TestDeepValidate_WellKnownServiceCoverage checks that the cluster refuses to come up when
// nothing serves an endpoint it cannot work without.
func TestDeepValidate_WellKnownServiceCoverage(t *testing.T) {
	c := buildDefaultCluster(t)
	var groups []*kopsapi.InstanceGroup

	apiserver := buildMinimalInstanceGroup("apiserver", kopsapi.InstanceGroupRoleAPIServer, "subnet-us-test-1a")
	// The only API server in the cluster, serving nothing: no internal endpoint for the nodes.
	apiserver.Spec.ServesWellKnownServices = []kopsapi.WellKnownService{}

	groups = append(groups, apiserver)
	groups = append(groups, buildMinimalInstanceGroup("etcd", kopsapi.InstanceGroupRoleEtcd, "subnet-us-test-1a"))
	groups = append(groups, buildMinimalInstanceGroup("kcm", kopsapi.InstanceGroupRoleKubeControllerManager, "subnet-us-test-1a"))
	groups = append(groups, buildMinimalInstanceGroup("scheduler", kopsapi.InstanceGroupRoleScheduler, "subnet-us-test-1a"))
	groups = append(groups, buildMinimalNodeInstanceGroup("subnet-us-test-1a"))

	expectErrorFromDeepValidate(t, c, groups, "no InstanceGroup serves the \"kube-apiserver-internal\" endpoint")
}

// TestDeepValidate_WellKnownServiceCoverage_Split checks the split-out shape the composite
// control plane needs: one API server group fronting external clients, another serving the
// cluster internally and hosting kops-controller.
func TestDeepValidate_WellKnownServiceCoverage_Split(t *testing.T) {
	c := buildDefaultCluster(t)
	var groups []*kopsapi.InstanceGroup

	external := buildMinimalInstanceGroup("external", kopsapi.InstanceGroupRoleAPIServer, "subnet-us-test-1a")
	external.Spec.ServesWellKnownServices = []kopsapi.WellKnownService{
		kopsapi.WellKnownServiceKubeAPIServerExternal,
	}
	internal := buildMinimalInstanceGroup("internal", kopsapi.InstanceGroupRoleAPIServer, "subnet-us-test-1a")
	internal.Spec.ServesWellKnownServices = []kopsapi.WellKnownService{
		kopsapi.WellKnownServiceKubeAPIServerInternal,
		kopsapi.WellKnownServiceKopsController,
	}

	groups = append(groups, external, internal)
	groups = append(groups, buildMinimalInstanceGroup("etcd", kopsapi.InstanceGroupRoleEtcd, "subnet-us-test-1a"))
	groups = append(groups, buildMinimalInstanceGroup("kcm", kopsapi.InstanceGroupRoleKubeControllerManager, "subnet-us-test-1a"))
	groups = append(groups, buildMinimalInstanceGroup("scheduler", kopsapi.InstanceGroupRoleScheduler, "subnet-us-test-1a"))
	groups = append(groups, buildMinimalNodeInstanceGroup("subnet-us-test-1a"))

	if err := validation.DeepValidate(c, groups, true, vfs.Context, nil); err != nil {
		t.Fatalf("Expected no error from DeepValidate for a split API server, got %v", err)
	}
}

// TestDeepValidate_HostedComponentCoverage checks that a cluster placing the scheduled
// components explicitly cannot leave one of them homeless. kOps places them by node label, so a
// component nothing claims is simply never scheduled, which is a cluster that silently never
// finishes coming up.
func TestDeepValidate_HostedComponentCoverage(t *testing.T) {
	build := func(components []kopsapi.ClusterComponent) []*kopsapi.InstanceGroup {
		apiserver := buildMinimalInstanceGroup("apiserver", kopsapi.InstanceGroupRoleAPIServer, "subnet-us-test-1a")
		apiserver.Spec.HostedComponents = components
		return []*kopsapi.InstanceGroup{
			apiserver,
			buildMinimalInstanceGroup("etcd", kopsapi.InstanceGroupRoleEtcd, "subnet-us-test-1a"),
			buildMinimalInstanceGroup("kcm", kopsapi.InstanceGroupRoleKubeControllerManager, "subnet-us-test-1a"),
			buildMinimalInstanceGroup("scheduler", kopsapi.InstanceGroupRoleScheduler, "subnet-us-test-1a"),
			buildMinimalNodeInstanceGroup("subnet-us-test-1a"),
		}
	}

	// buildDefaultCluster runs an external cloud-controller-manager, as any current cluster does,
	// so a complete claim set includes it.
	complete := []kopsapi.ClusterComponent{
		kopsapi.ClusterComponentKopsChannel,
		kopsapi.ClusterComponentKopsController,
		kopsapi.ClusterComponentCloudControllerManager,
	}

	t.Run("complete", func(t *testing.T) {
		c := buildDefaultCluster(t)
		if err := validation.DeepValidate(c, build(complete), true, vfs.Context, nil); err != nil {
			t.Fatalf("Expected no error from DeepValidate, got %v", err)
		}
	})

	t.Run("missing kops-controller", func(t *testing.T) {
		c := buildDefaultCluster(t)
		groups := build([]kopsapi.ClusterComponent{
			kopsapi.ClusterComponentKopsChannel,
			kopsapi.ClusterComponentCloudControllerManager,
		})
		expectErrorFromDeepValidate(t, c, groups, "no InstanceGroup hosts the \"kops-controller\" component")
	})

	t.Run("missing kops-channel", func(t *testing.T) {
		c := buildDefaultCluster(t)
		groups := build([]kopsapi.ClusterComponent{
			kopsapi.ClusterComponentKopsController,
			kopsapi.ClusterComponentCloudControllerManager,
		})
		expectErrorFromDeepValidate(t, c, groups, "no InstanceGroup hosts the \"kops-channel\" component")
	})

	t.Run("explicitly empty on one group still counts as in use", func(t *testing.T) {
		c := buildDefaultCluster(t)
		// The field is set, but claims nothing: the cluster would install the addons and never
		// schedule them.
		groups := build([]kopsapi.ClusterComponent{})
		expectErrorFromDeepValidate(t, c, groups, "no InstanceGroup hosts the \"kops-channel\" component")
	})

	t.Run("cloud-controller-manager required when the cluster runs one", func(t *testing.T) {
		c := buildDefaultCluster(t)
		groups := build([]kopsapi.ClusterComponent{
			kopsapi.ClusterComponentKopsChannel,
			kopsapi.ClusterComponentKopsController,
		})
		expectErrorFromDeepValidate(t, c, groups, "no InstanceGroup hosts the \"cloud-controller-manager\" component")
	})

	t.Run("cert-manager required when enabled", func(t *testing.T) {
		c := buildDefaultCluster(t)
		c.Spec.CertManager = &kopsapi.CertManagerConfig{Enabled: new(true)}
		expectErrorFromDeepValidate(t, c, build(complete), "no InstanceGroup hosts the \"cert-manager\" component")
	})

	t.Run("components the cluster will not install need no home", func(t *testing.T) {
		c := buildDefaultCluster(t)
		// cert-manager disabled, so not claiming it is fine. capi-manager is never required.
		c.Spec.CertManager = &kopsapi.CertManagerConfig{Enabled: new(false)}
		if err := validation.DeepValidate(c, build(complete), true, vfs.Context, nil); err != nil {
			t.Fatalf("Expected no error from DeepValidate, got %v", err)
		}
	})
}

func TestDeepValidate_NoNodeZones(t *testing.T) {
	c := buildDefaultCluster(t)
	var groups []*kopsapi.InstanceGroup
	groups = append(groups, buildMinimalMasterInstanceGroup("subnet-us-test-1a"))
	expectErrorFromDeepValidate(t, c, groups, "must configure at least one Node InstanceGroup")
}

func TestDeepValidate_NoMasterZones(t *testing.T) {
	c := buildDefaultCluster(t)
	var groups []*kopsapi.InstanceGroup
	groups = append(groups, buildMinimalNodeInstanceGroup("subnet-us-test-1a"))
	expectErrorFromDeepValidate(t, c, groups, "must configure either a ControlPlane InstanceGroup or separate APIServer, Etcd, KubeControllerManager, and Scheduler InstanceGroups")
}

func TestDeepValidate_BadZone(t *testing.T) {
	t.Skipf("Zone validation not checked by DeepValidate")
	c := buildDefaultCluster(t)
	c.Spec.Networking.Subnets = []kopsapi.ClusterSubnetSpec{
		{Name: "subnet-badzone", Zone: "us-test-1z", CIDR: "172.20.1.0/24"},
	}
	var groups []*kopsapi.InstanceGroup
	groups = append(groups, buildMinimalMasterInstanceGroup("subnet-us-test-1z"))
	groups = append(groups, buildMinimalNodeInstanceGroup("subnet-us-test-1z"))
	expectErrorFromDeepValidate(t, c, groups, "Zone is not a recognized AZ")
}

func TestDeepValidate_MixedRegion(t *testing.T) {
	t.Skipf("Region validation not checked by DeepValidate")
	c := buildDefaultCluster(t)
	c.Spec.Networking.Subnets = []kopsapi.ClusterSubnetSpec{
		{Name: "test1a", Zone: "us-test-1a", CIDR: "172.20.1.0/24"},
		{Name: "west1b", Zone: "us-west-1b", CIDR: "172.20.2.0/24"},
	}
	var groups []*kopsapi.InstanceGroup
	groups = append(groups, buildMinimalMasterInstanceGroup("subnet-us-test-1a"))
	groups = append(groups, buildMinimalNodeInstanceGroup("subnet-us-test-1a", "subnet-us-west-1b"))

	expectErrorFromDeepValidate(t, c, groups, "Clusters cannot span multiple regions")
}

func TestDeepValidate_RegionAsZone(t *testing.T) {
	t.Skipf("Region validation not checked by DeepValidate")
	c := buildDefaultCluster(t)
	c.Spec.Networking.Subnets = []kopsapi.ClusterSubnetSpec{
		{Name: "test1", Zone: "us-test-1", CIDR: "172.20.1.0/24"},
	}
	var groups []*kopsapi.InstanceGroup
	groups = append(groups, buildMinimalMasterInstanceGroup("subnet-us-test-1"))
	groups = append(groups, buildMinimalNodeInstanceGroup("subnet-us-test-1"))

	expectErrorFromDeepValidate(t, c, groups, "Region is not a recognized EC2 region: \"us-east-\" (check you have specified valid zones?)")
}

func TestDeepValidate_NotIncludedZone(t *testing.T) {
	c := buildDefaultCluster(t)
	var groups []*kopsapi.InstanceGroup
	groups = append(groups, buildMinimalMasterInstanceGroup("subnet-us-test-1d"))
	groups = append(groups, buildMinimalNodeInstanceGroup("subnet-us-test-1d"))

	expectErrorFromDeepValidate(t, c, groups, "spec.networking.subnets[0]: Not found: \"subnet-us-test-1d\"")
}

func TestDeepValidate_DuplicateZones(t *testing.T) {
	c := buildDefaultCluster(t)
	c.Spec.Networking.Subnets = []kopsapi.ClusterSubnetSpec{
		{Name: "dup1", Zone: "us-test-1a", CIDR: "172.20.1.0/24"},
		{Name: "dup1", Zone: "us-test-1a", CIDR: "172.20.2.0/24"},
	}
	var groups []*kopsapi.InstanceGroup
	groups = append(groups, buildMinimalMasterInstanceGroup("dup1"))
	groups = append(groups, buildMinimalNodeInstanceGroup("dup1"))
	expectErrorFromDeepValidate(t, c, groups, "spec.networking.subnets[1].name: Duplicate value: \"dup1\"")
}

func TestDeepValidate_ExtraMasterZone(t *testing.T) {
	c := buildDefaultCluster(t)
	c.Spec.Networking.Subnets = []kopsapi.ClusterSubnetSpec{
		{Name: "test1a", Zone: "us-test-1a", CIDR: "172.20.1.0/24", Type: kopsapi.SubnetTypePublic},
		{Name: "test1b", Zone: "us-test-1b", CIDR: "172.20.2.0/24", Type: kopsapi.SubnetTypePublic},
	}
	var groups []*kopsapi.InstanceGroup
	groups = append(groups, buildMinimalMasterInstanceGroup("subnet-us-test-1a"))
	groups = append(groups, buildMinimalMasterInstanceGroup("subnet-us-test-1b"))
	groups = append(groups, buildMinimalMasterInstanceGroup("subnet-us-test-1c"))
	groups = append(groups, buildMinimalNodeInstanceGroup("subnet-us-test-1a", "subnet-us-test-1b"))

	expectErrorFromDeepValidate(t, c, groups, "spec.networking.subnets[0]: Not found: \"subnet-us-test-1a\"")
}

func TestDeepValidate_EvenEtcdClusterSize(t *testing.T) {
	c := buildDefaultCluster(t)
	c.Spec.EtcdClusters = []kopsapi.EtcdClusterSpec{
		{
			Name: "main",
			Members: []kopsapi.EtcdMemberSpec{
				{Name: "us-test-1a", InstanceGroup: new("us-test-1a")},
				{Name: "us-test-1b", InstanceGroup: new("us-test-1b")},
			},
		},
	}

	var groups []*kopsapi.InstanceGroup
	groups = append(groups, buildMinimalMasterInstanceGroup("subnet-us-test-1a"))
	groups = append(groups, buildMinimalMasterInstanceGroup("subnet-us-test-1b"))
	groups = append(groups, buildMinimalMasterInstanceGroup("subnet-us-test-1c"))
	groups = append(groups, buildMinimalMasterInstanceGroup("subnet-us-test-1d"))
	groups = append(groups, buildMinimalNodeInstanceGroup("subnet-us-test-1a"))

	expectErrorFromDeepValidate(t, c, groups, "Should be an odd number of control-plane-zones for quorum. Use --zones and --control-plane-zones to declare node zones and control-plane zones separately")
}

func TestDeepValidate_MissingEtcdMember(t *testing.T) {
	c := buildDefaultCluster(t)
	c.Spec.EtcdClusters = []kopsapi.EtcdClusterSpec{
		{
			Name: "main",
			Members: []kopsapi.EtcdMemberSpec{
				{Name: "us-test-1a", InstanceGroup: new("us-test-1a")},
				{Name: "us-test-1b", InstanceGroup: new("us-test-1b")},
				{Name: "us-test-1c", InstanceGroup: new("us-test-1c")},
			},
		},
	}

	var groups []*kopsapi.InstanceGroup
	groups = append(groups, buildMinimalMasterInstanceGroup("subnet-us-test-1a"))
	groups = append(groups, buildMinimalMasterInstanceGroup("subnet-us-test-1b"))
	groups = append(groups, buildMinimalMasterInstanceGroup("subnet-us-test-1c"))
	groups = append(groups, buildMinimalMasterInstanceGroup("subnet-us-test-1d"))
	groups = append(groups, buildMinimalNodeInstanceGroup("subnet-us-test-1a"))

	expectErrorFromDeepValidate(t, c, groups, "spec.metadata.name: Forbidden: InstanceGroup \"master-subnet-us-test-1a\" with role ControlPlane must have a member in etcd (IG: \"us-test-1c\") cluster \"main\"")
}

func expectErrorFromDeepValidate(t *testing.T, c *kopsapi.Cluster, groups []*kopsapi.InstanceGroup, message string) {
	err := validation.DeepValidate(c, groups, true, vfs.Context, nil)
	if err == nil {
		t.Fatalf("Expected error %q from DeepValidate (strict=true), not no error raised", message)
	}
	actualMessage := fmt.Sprintf("%v", err)
	if !strings.Contains(actualMessage, message) {
		t.Fatalf("Expected error %q, got %q", message, actualMessage)
	}
}
