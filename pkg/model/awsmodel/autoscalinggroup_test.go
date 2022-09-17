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
	"fmt"
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/iam"
	"k8s.io/kops/pkg/testutils"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
	"k8s.io/kops/upup/pkg/fi/fitasks"
)

const (
	sshPublicKeyEntry = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCySdqIU+FhCWl3BNrAvPaOe5VfL2aCARUWwy91ZP+T7LBwFa9lhdttfjp/VX1D1/PVwntn2EhN079m8c2kfdmiZ/iCHqrLyIGSd+BOiCz0lT47znvANSfxYjLUuKrWWWeaXqerJkOsAD4PHchRLbZGPdbfoBKwtb/WT4GMRQmb9vmiaZYjsfdPPM9KkWI9ECoWFGjGehA8D+iYIPR711kRacb1xdYmnjHqxAZHFsb5L8wDWIeAyhy49cBD+lbzTiioq2xWLorXuFmXh6Do89PgzvHeyCLY6816f/kCX6wIFts8A2eaEHFL4rAOsuh6qHmSxGCR9peSyuRW8DxV725x justin@test"
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
	ig := buildNodeInstanceGroup("subnet-us-test-1a")
	ig.Spec.RootVolumeOptimization = fi.Bool(true)

	k := [][]byte{}
	k = append(k, []byte(sshPublicKeyEntry))

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
		BootstrapScriptBuilder: &model.BootstrapScriptBuilder{
			Lifecycle: fi.LifecycleSync,
			Cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					CloudProvider: kops.CloudProviderSpec{
						AWS: &kops.AWSSpec{},
					},
					Networking:        &kops.NetworkingSpec{},
					KubernetesVersion: "1.20.0",
				},
			},
		},
		Cluster: cluster,
	}

	c := &fi.ModelBuilderContext{
		Tasks: make(map[string]fi.Task),
	}

	// We need the CA for the bootstrap script
	caTask := &fitasks.Keypair{
		Name:    fi.String(fi.CertificateIDCA),
		Subject: "cn=kubernetes",
		Type:    "ca",
	}
	c.AddTask(caTask)
	for _, keypair := range []string{
		"etcd-clients-ca",
	} {
		task := &fitasks.Keypair{
			Name:    fi.String(keypair),
			Subject: "cn=" + keypair,
			Type:    "ca",
		}
		c.AddTask(task)
	}

	if err := b.Build(c); err != nil {
		t.Fatalf("error from Build: %v", err)
	}

	lc := c.Tasks["LaunchTemplate/nodes.testcluster.test.com"].(*awstasks.LaunchTemplate)

	if *lc.RootVolumeOptimization == false {
		t.Fatalf("RootVolumeOptimization was expected to be true, but was false")
	}
}

func TestAPIServerAdditionalSecurityGroupsWithNLB(t *testing.T) {
	const sgIDAPIServer = "sg-01234567890abcdef"

	cluster := buildMinimalCluster()
	cluster.Spec.API = &kops.AccessSpec{
		LoadBalancer: &kops.LoadBalancerAccessSpec{
			Class:                    kops.LoadBalancerClassNetwork,
			AdditionalSecurityGroups: []string{sgIDAPIServer},
		},
	}

	const (
		roleBastion = iota
		roleMaster
		roleNode
		_roleCount
	)
	igs := make([]*kops.InstanceGroup, _roleCount)
	// NB: (*AutoscalingGroupModelBuilder).buildLaunchTemplateTask expects there to be at least
	// one subnet specified in each InstanceGroup.
	subnets := []string{cluster.Spec.Subnets[0].Name}
	igs[roleBastion] = &kops.InstanceGroup{
		ObjectMeta: v1.ObjectMeta{
			Name: "bastion1",
		},
		Spec: kops.InstanceGroupSpec{
			Role:                     kops.InstanceGroupRoleBastion,
			Subnets:                  subnets,
			AdditionalSecurityGroups: []string{"sg-1234567890abcdef0"},
		},
	}
	igs[roleMaster] = &kops.InstanceGroup{
		ObjectMeta: v1.ObjectMeta{
			Name: "master1",
		},
		Spec: kops.InstanceGroupSpec{
			Role:                     kops.InstanceGroupRoleMaster,
			Subnets:                  subnets,
			AdditionalSecurityGroups: []string{"sg-234567890abcdef01"},
		},
	}
	igs[roleNode] = &kops.InstanceGroup{
		ObjectMeta: v1.ObjectMeta{
			Name: "node1",
		},
		Spec: kops.InstanceGroupSpec{
			Role:                     kops.InstanceGroupRoleNode,
			Subnets:                  subnets,
			AdditionalSecurityGroups: []string{"sg-34567890abcdef012"},
		},
	}

	b := AutoscalingGroupModelBuilder{
		AWSModelContext: &AWSModelContext{
			KopsModelContext: &model.KopsModelContext{
				IAMModelContext: iam.IAMModelContext{Cluster: cluster},
				SSHPublicKeys:   [][]byte{[]byte(sshPublicKeyEntry)},
				InstanceGroups:  igs,
			},
		},
		BootstrapScriptBuilder: &model.BootstrapScriptBuilder{
			KopsModelContext: &model.KopsModelContext{
				IAMModelContext: iam.IAMModelContext{Cluster: cluster},
				InstanceGroups:  igs,
			},
			Lifecycle: fi.LifecycleSync,
			Cluster:   cluster,
		},
		Cluster: cluster,
	}

	c := &fi.ModelBuilderContext{
		Tasks: make(map[string]fi.Task),
	}

	// We need the CA for the bootstrap script
	caTask := &fitasks.Keypair{
		Name:    fi.String(fi.CertificateIDCA),
		Subject: "cn=kubernetes",
		Type:    "ca",
	}
	c.AddTask(caTask)
	for _, keypair := range []string{
		"apiserver-aggregator-ca",
		"etcd-clients-ca",
		"etcd-manager-ca-events",
		"etcd-manager-ca-main",
		"etcd-peers-ca-events",
		"etcd-peers-ca-main",
		"service-account",
	} {
		task := &fitasks.Keypair{
			Name:    fi.String(keypair),
			Subject: "cn=" + keypair,
			Type:    "ca",
		}
		c.AddTask(task)
	}
	for _, keypair := range []string{
		"kubelet",
		"kube-proxy",
	} {
		task := &fitasks.Keypair{
			Name:    fi.String(keypair),
			Subject: "cn=" + keypair,
			Signer:  caTask,
			Type:    "client",
		}
		c.AddTask(task)
	}

	b.Build(c)

	hasSecurityGroup := func(lt *awstasks.LaunchTemplate, id string) bool {
		for _, sg := range lt.SecurityGroups {
			if sg.ID != nil && *sg.ID == id {
				return true
			}
		}
		return false
	}
	hasDesignatedSecurityGroup := func(lt *awstasks.LaunchTemplate) bool {
		return hasSecurityGroup(lt, sgIDAPIServer)
	}
	launchTemplateForGroup := func(t *testing.T, ig *kops.InstanceGroup) *awstasks.LaunchTemplate {
		t.Helper()
		subdomain := ig.Name
		if ig.Spec.Role == kops.InstanceGroupRoleMaster {
			subdomain = ig.Name + ".masters"
		}
		task, ok := c.Tasks[fmt.Sprintf("LaunchTemplate/%s.%s", subdomain, cluster.Name)]
		if !ok {
			t.Fatalf("No task available in model build context for InstanceGroup %q", ig.Name)
		}
		if task == nil {
			t.Fatalf("Task pointer in model build context for InstanceGroup %q is nil", ig.Name)
		}
		return task.(*awstasks.LaunchTemplate)
	}
	tests := []struct {
		ig          *kops.InstanceGroup
		expectHasSG bool
	}{
		{igs[roleBastion], false},
		{igs[roleMaster], true},
		{igs[roleNode], false},
	}
	for _, test := range tests {
		role := test.ig.Spec.Role
		t.Run(string(role), func(t *testing.T) {
			lt := launchTemplateForGroup(t, test.ig)
			if want, got := test.expectHasSG, hasDesignatedSecurityGroup(lt); got != want {
				t.Errorf("%q (role %q): launch template includes API server security group: want %t, got %t", test.ig.Name, role, want, got)
			}
			for _, sg := range test.ig.Spec.AdditionalSecurityGroups {
				if want, got := true, hasSecurityGroup(lt, sg); got != want {
					t.Errorf("%q (role %q): launch template includes additional security group %q: want %t, got %t", test.ig.Name, role, sg, want, got)
				}
			}
		})
	}
}
