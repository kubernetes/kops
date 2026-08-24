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
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi/cloudup/linode"
	"k8s.io/kops/util/pkg/vfs"
)

func TestPerformAssignmentsAssignsLinodeSubnetCIDRs(t *testing.T) {
	cluster := &kops.Cluster{
		ObjectMeta: metav1.ObjectMeta{Name: "example.k8s.local"},
		Spec: kops.ClusterSpec{
			KubernetesVersion: "1.36.0",
			CloudProvider:     kops.CloudProviderSpec{Linode: &kops.LinodeSpec{}},
			Networking: kops.NetworkingSpec{Subnets: []kops.ClusterSubnetSpec{
				{Name: "us-ord", Region: "us-ord", Zone: "us-ord", Type: kops.SubnetTypePrivate},
				{Name: "utility-us-ord", Region: "us-ord", Zone: "us-ord", Type: kops.SubnetTypeUtility},
			}},
		},
	}

	if err := PerformAssignments(cluster, vfs.NewTestingVFSContext(), &linode.MockLinodeCloud{}); err != nil {
		t.Fatalf("PerformAssignments returned error: %v", err)
	}

	if got, want := cluster.Spec.Networking.NetworkCIDR, defaultLinodeNetworkCIDR; got != want {
		t.Fatalf("unexpected Linode network CIDR: got %q, want %q", got, want)
	}
	if got, want := cluster.Spec.Networking.Subnets[0].CIDR, "10.0.128.0/17"; got != want {
		t.Fatalf("unexpected private subnet CIDR: got %q, want %q", got, want)
	}
	if got, want := cluster.Spec.Networking.Subnets[1].CIDR, "10.0.0.0/20"; got != want {
		t.Fatalf("unexpected utility subnet CIDR: got %q, want %q", got, want)
	}
}

func TestPopulateClusterSpec_Proxy(t *testing.T) {
	_, c := buildMinimalCluster()

	c.Spec.Networking.EgressProxy = &kops.EgressProxySpec{
		ProxyExcludes: "google.com",
		HTTPProxy: kops.HTTPProxy{
			Host: "52.205.179.249",
			Port: 3128,
		},
	}

	c.Spec.Networking.NonMasqueradeCIDR = "100.64.0.1/10"
	c.Spec.Networking.NetworkCIDR = "192.168.0.0/20"
	var err error
	c.Spec.Networking.EgressProxy, err = assignProxy(c)
	if err != nil {
		t.Fatalf("unable to assign proxy, %v", err)
	}

	expectedExcludes := "google.com,127.0.0.1,localhost,api.testcluster.test.com,testcluster.test.com,100.64.0.2,100.64.0.1/10,169.254.169.254,192.168.0.0/20"
	if c.Spec.Networking.EgressProxy.ProxyExcludes != expectedExcludes {
		t.Fatalf("Incorrect proxy excludes set: %v, expected %v", c.Spec.Networking.EgressProxy.ProxyExcludes, expectedExcludes)
	}

	c.Spec.Networking.EgressProxy = &kops.EgressProxySpec{
		HTTPProxy: kops.HTTPProxy{
			Host: "52.205.179.249",
			Port: 3128,
		},
	}

	c.Spec.Networking.NonMasqueradeCIDR = "100.64.0.0/10"
	c.Spec.Networking.NetworkCIDR = "192.168.0.0/20"
	c.Spec.Networking.EgressProxy.ProxyExcludes = ""

	c.Spec.Networking.EgressProxy, err = assignProxy(c)
	if err != nil {
		t.Fatalf("unable to assign proxy, %v", err)
	}

	expectedExcludes = "127.0.0.1,localhost,api.testcluster.test.com,testcluster.test.com,100.64.0.1,100.64.0.0/10,169.254.169.254,192.168.0.0/20"
	if c.Spec.Networking.EgressProxy.ProxyExcludes != expectedExcludes {
		t.Fatalf("Incorrect proxy excludes set: %v, expected %v", c.Spec.Networking.EgressProxy.ProxyExcludes, expectedExcludes)
	}

	c.Spec.Networking.NonMasqueradeCIDR = "172.16.0.5/12"
	c.Spec.Networking.NetworkCIDR = "192.168.0.0/20"
	c.Spec.CloudProvider = kops.CloudProviderSpec{
		GCE: &kops.GCESpec{},
	}
	c.Spec.Networking.EgressProxy.ProxyExcludes = ""
	c.Spec.Networking.EgressProxy, err = assignProxy(c)
	if err != nil {
		t.Fatalf("unable to assign proxy, %v", err)
	}

	expectedExcludes = "127.0.0.1,localhost,api.testcluster.test.com,testcluster.test.com,172.16.0.6,172.16.0.5/12,192.168.0.0/20"
	if c.Spec.Networking.EgressProxy.ProxyExcludes != expectedExcludes {
		t.Fatalf("Incorrect proxy excludes set: %v", c.Spec.Networking.EgressProxy.ProxyExcludes)
	}

	// idempotency test
	c.Spec.Networking.EgressProxy, err = assignProxy(c)
	if err != nil {
		t.Fatalf("unable to assign proxy, %v", err)
	}

	expectedExcludes = "127.0.0.1,localhost,api.testcluster.test.com,testcluster.test.com,172.16.0.6,172.16.0.5/12,192.168.0.0/20"
	if c.Spec.Networking.EgressProxy.ProxyExcludes != expectedExcludes {
		t.Fatalf("Incorrect proxy excludes set during idempotency check: %v    should have been %v", c.Spec.Networking.EgressProxy.ProxyExcludes, expectedExcludes)
	}
}
