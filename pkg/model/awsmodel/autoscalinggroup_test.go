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
	"k8s.io/kops/pkg/model/iam"
	"k8s.io/kops/pkg/testutils"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
)

func buildMinimalCluster() *kops.Cluster {
	return testutils.BuildMinimalCluster("testcluster.test.com")

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
				IAMModelContext: iam.IAMModelContext{Cluster: cluster},
				SSHPublicKeys:   k,
				InstanceGroups:  igs,
			},
		},
	}

	c := &fi.ModelBuilderContext{
		Tasks: make(map[string]fi.Task),
	}

	b.Build(c)

	lc := c.Tasks["LaunchTemplate/nodes.testcluster.test.com"].(*awstasks.LaunchTemplate)

	if *lc.RootVolumeOptimization == false {
		t.Fatalf("RootVolumeOptimization was expected to be true, but was false")
	}
}
