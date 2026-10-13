/*
Copyright 2026 The Kubernetes Authors.

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

package gcemodel

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/iam"
	"k8s.io/kops/pkg/testutils"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/gcetasks"
)

// buildInternalLBWhenUpdatingOnlyNodes builds the internal API load balancer of a cluster
// with the given instance groups, when only the nodes are being updated.
func buildInternalLBWhenUpdatingOnlyNodes(t *testing.T, instanceGroups ...*kops.InstanceGroup) (*APILoadBalancerBuilder, *fi.CloudupModelBuilderContext) {
	cluster := testutils.BuildMinimalClusterGCE("minimal-gce.example.com", "testproject")
	cluster.Spec.API.LoadBalancer = &kops.LoadBalancerAccessSpec{Type: kops.LoadBalancerTypeInternal}

	nodes := newInstanceGroup("nodes", kops.InstanceGroupRoleNode)
	b := &APILoadBalancerBuilder{
		GCEModelContext: &GCEModelContext{
			KopsModelContext: &model.KopsModelContext{
				IAMModelContext:   iam.IAMModelContext{Cluster: cluster},
				AllInstanceGroups: append(instanceGroups, nodes),
				InstanceGroups:    []*kops.InstanceGroup{nodes},
			},
		},
		Lifecycle: fi.LifecycleSync,
	}
	c := &fi.CloudupModelBuilderContext{
		Tasks: make(map[string]fi.CloudupTask),
		// The MIGs of the instance groups not being updated must not change, whatever the overrides.
		LifecycleOverrides: map[string]fi.Lifecycle{"InstanceGroupManager": fi.LifecycleSync},
	}
	if err := b.Build(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return b, c
}

func newInstanceGroup(name string, role kops.InstanceGroupRole) *kops.InstanceGroup {
	return &kops.InstanceGroup{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: kops.InstanceGroupSpec{
			Role:    role,
			Subnets: []string{"us-test1-a"},
			Zones:   []string{"us-test1-a"},
		},
	}
}

// assertLinksIgnoredMIGs checks that a backend service links to the MIGs of the given instance
// groups, which are not being updated, and that these MIGs are never changed.
func assertLinksIgnoredMIGs(t *testing.T, b *APILoadBalancerBuilder, c *fi.CloudupModelBuilderContext, backendService string, instanceGroups ...*kops.InstanceGroup) {
	task, ok := c.Tasks["BackendService/"+b.NameForBackendService(backendService)].(*gcetasks.BackendService)
	if !ok {
		t.Fatalf("backend service %q not found", backendService)
	}
	var linked []string
	for _, igm := range task.InstanceGroupManagers {
		linked = append(linked, fi.ValueOf(igm.Name))
	}
	if len(linked) != len(instanceGroups) {
		t.Fatalf("backend service %q links to %v, expected the MIGs of %d instance groups", backendService, linked, len(instanceGroups))
	}
	for i, ig := range instanceGroups {
		name := gce.NameForInstanceGroupManager(b.Cluster.ObjectMeta.Name, ig.ObjectMeta.Name, "us-test1-a")
		if linked[i] != name {
			t.Errorf("backend service %q links to %q, expected %q", backendService, linked[i], name)
		}
		igm, ok := c.Tasks["InstanceGroupManager/"+name].(*gcetasks.InstanceGroupManager)
		if !ok {
			t.Errorf("MIG %q not found", name)
		} else if igm.Lifecycle != fi.LifecycleIgnore {
			t.Errorf("MIG %q has lifecycle %q, expected %q", name, igm.Lifecycle, fi.LifecycleIgnore)
		}
	}
}

func TestInternalLoadBalancerWhenUpdatingOnlyNodes(t *testing.T) {
	controlPlane := newInstanceGroup("master-us-test1-a", kops.InstanceGroupRoleControlPlane)
	b, c := buildInternalLBWhenUpdatingOnlyNodes(t, controlPlane)

	assertLinksIgnoredMIGs(t, b, c, "api", controlPlane)
	if c.Tasks["ForwardingRule/"+b.NameForForwardingRule("api-us-test1-a")] == nil {
		t.Errorf("forwarding rule of the API not found")
	}
}

func TestEtcdLoadBalancerWhenUpdatingOnlyNodes(t *testing.T) {
	controlPlane := newInstanceGroup("master-us-test1-a", kops.InstanceGroupRoleControlPlane)
	etcd := newInstanceGroup("etcd-us-test1-a", kops.InstanceGroupRoleEtcd)
	b, c := buildInternalLBWhenUpdatingOnlyNodes(t, controlPlane, etcd)

	// kops-controller gets its own backend service, because etcd does not run on all the MIGs of the API.
	assertLinksIgnoredMIGs(t, b, c, "kops-controller", controlPlane)
	assertLinksIgnoredMIGs(t, b, c, "etcd", controlPlane, etcd)
	if c.Tasks["ForwardingRule/"+b.NameForForwardingRule("etcd-us-test1-a")] == nil {
		t.Errorf("forwarding rule of etcd not found")
	}
}
