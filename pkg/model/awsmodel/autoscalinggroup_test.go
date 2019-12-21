/*
Copyright 2017 The Kubernetes Authors.

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

package awsmodel

import (
	"testing"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
)

func buildMinimalCluster() *kops.Cluster {
	c := &kops.Cluster{}
	c.ObjectMeta.Name = "testcluster.test.com"
	c.Spec.KubernetesVersion = "1.14.6"
	c.Spec.Subnets = []kops.ClusterSubnetSpec{
		{Name: "subnet-us-mock-1a", Zone: "us-mock-1a", CIDR: "172.20.1.0/24", Type: kops.SubnetTypePrivate},
	}

	c.Spec.KubernetesAPIAccess = []string{"0.0.0.0/0"}
	c.Spec.SSHAccess = []string{"0.0.0.0/0"}

	// Default to public topology
	c.Spec.Topology = &kops.TopologySpec{
		Masters: kops.TopologyPublic,
		Nodes:   kops.TopologyPublic,
	}
	c.Spec.NetworkCIDR = "172.20.0.0/16"
	c.Spec.NonMasqueradeCIDR = "100.64.0.0/10"
	c.Spec.CloudProvider = "aws"

	c.Spec.ConfigBase = "s3://unittest-bucket/"

	// Required to stop a call to cloud provider
	// TODO: Mock cloudprovider
	c.Spec.DNSZone = "test.com"

	return c
}

func buildNodeInstanceGroup(subnets ...string) *kops.InstanceGroup {
	g := &kops.InstanceGroup{}
	g.ObjectMeta.Name = "nodes"
	g.Spec.Role = kops.InstanceGroupRoleNode
	g.Spec.Subnets = subnets

	return g
}

// Tests that RootVolumeOptimization flag gets added to the awstasks
func TestRootVolumeOptimizationFlag(t *testing.T) {
	cluster := buildMinimalCluster()
	ig := buildNodeInstanceGroup("subnet-us-mock-1a")
	ig.Spec.RootVolumeOptimization = fi.Bool(true)

	k := [][]byte{}
	k = append(k, []byte("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCySdqIU+FhCWl3BNrAvPaOe5VfL2aCARUWwy91ZP+T7LBwFa9lhdttfjp/VX1D1/PVwntn2EhN079m8c2kfdmiZ/iCHqrLyIGSd+BOiCz0lT47znvANSfxYjLUuKrWWWeaXqerJkOsAD4PHchRLbZGPdbfoBKwtb/WT4GMRQmb9vmiaZYjsfdPPM9KkWI9ECoWFGjGehA8D+iYIPR711kRacb1xdYmnjHqxAZHFsb5L8wDWIeAyhy49cBD+lbzTiioq2xWLorXuFmXh6Do89PgzvHeyCLY6816f/kCX6wIFts8A2eaEHFL4rAOsuh6qHmSxGCR9peSyuRW8DxV725x justin@test"))

	igs := []*kops.InstanceGroup{}
	igs = append(igs, ig)

	b := AutoscalingGroupModelBuilder{
		AWSModelContext: &AWSModelContext{
			KopsModelContext: &model.KopsModelContext{
				SSHPublicKeys:  k,
				Cluster:        cluster,
				InstanceGroups: igs,
			},
		},
	}

	c := &fi.ModelBuilderContext{
		Tasks: make(map[string]fi.Task),
	}

	b.Build(c)

	lc := c.Tasks["LaunchConfiguration/nodes.testcluster.test.com"].(*awstasks.LaunchConfiguration)

	if *lc.RootVolumeOptimization == false {
		t.Fatalf("RootVolumeOptimization was expected to be true, but was false")
	}
}
