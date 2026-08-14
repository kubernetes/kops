/*
Copyright The Kubernetes Authors.

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

package linodemodel

import (
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/iam"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/linode"
	"k8s.io/kops/upup/pkg/fi/cloudup/linodetasks"
	"k8s.io/kops/upup/pkg/fi/fitasks"
)

func TestInstanceModelBuilderBuildUsesIGSubnetAndAllowsZeroSize(t *testing.T) {
	cluster := &kops.Cluster{
		ObjectMeta: metav1.ObjectMeta{Name: "example.k8s.local"},
		Spec: kops.ClusterSpec{
			CloudProvider: kops.CloudProviderSpec{Linode: &kops.LinodeSpec{}},
			Networking: kops.NetworkingSpec{Subnets: []kops.ClusterSubnetSpec{
				{Name: "subnet-a", Region: "us-east", CIDR: "172.16.1.0/24", Type: kops.SubnetTypePublic},
				{Name: "subnet-b", Region: "us-east", CIDR: "172.16.2.0/24", Type: kops.SubnetTypePublic},
			}},
		},
	}
	ig := &kops.InstanceGroup{
		ObjectMeta: metav1.ObjectMeta{Name: "nodes-us-east"},
		Spec: kops.InstanceGroupSpec{
			Role:        kops.InstanceGroupRoleNode,
			Subnets:     []string{"subnet-b"},
			MachineType: "g6-standard-2",
			Image:       "linode/ubuntu24.04",
			MinSize:     new(int32),
		},
	}
	context := newLinodeInstanceModelBuilderContext(cluster)
	addBootstrapPrerequisites(context)

	networkBuilder := &NetworkModelBuilder{
		LinodeModelContext: &LinodeModelContext{KopsModelContext: contextModel(cluster, []*kops.InstanceGroup{ig})},
		Lifecycle:          fi.LifecycleSync,
	}
	if err := networkBuilder.Build(context); err != nil {
		t.Fatalf("NetworkModelBuilder.Build returned error: %v", err)
	}

	publicKey := fi.Resource(fi.NewStringResource(testSSHPublicKey))
	context.AddTask(&linodetasks.SSHKey{
		Name:      new("example-k8s-local-default"),
		Lifecycle: fi.LifecycleSync,
		PublicKey: &publicKey,
	})

	builder := &InstanceModelBuilder{
		LinodeModelContext: &LinodeModelContext{KopsModelContext: contextModel(cluster, []*kops.InstanceGroup{ig})},
		Lifecycle:          fi.LifecycleSync,
		BootstrapScriptBuilder: &model.BootstrapScriptBuilder{
			Lifecycle:        fi.LifecycleSync,
			KopsModelContext: contextModel(cluster, []*kops.InstanceGroup{ig}),
		},
	}

	if err := builder.Build(context); err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	instanceTask := findInstanceTask(t, context, ig.Name)
	if got, want := fi.ValueOf(instanceTask.Subnet.Name), linode.NormalizeLinodeLabel(cluster.Name+"-subnet-b"); got != want {
		t.Fatalf("unexpected subnet task name: got %q, want %q", got, want)
	}
	if got, want := instanceTask.Region, "us-east"; got != want {
		t.Fatalf("unexpected region: got %q, want %q", got, want)
	}
	if got, want := instanceTask.Count, 0; got != want {
		t.Fatalf("unexpected count: got %d, want %d", got, want)
	}
	if got, want := fi.ValueOf(instanceTask.RequirePublicInterface), true; got != want {
		t.Fatalf("unexpected RequirePublicInterface: got %t, want %t", got, want)
	}
	if instanceTask.UserData == nil {
		t.Fatalf("expected userdata resource")
	}
	if got, want := len(instanceTask.AuthorizedKeys), 1; got != want {
		t.Fatalf("unexpected authorized key count: got %d, want %d", got, want)
	}
}

func TestInstanceModelBuilderBuildRejectsMultipleIGSubnets(t *testing.T) {
	minSize := int32(1)
	cluster := &kops.Cluster{
		ObjectMeta: metav1.ObjectMeta{Name: "example.k8s.local"},
		Spec: kops.ClusterSpec{
			CloudProvider: kops.CloudProviderSpec{Linode: &kops.LinodeSpec{}},
			Networking: kops.NetworkingSpec{Subnets: []kops.ClusterSubnetSpec{
				{Name: "subnet-a", Region: "us-east", CIDR: "172.16.1.0/24", Type: kops.SubnetTypePublic},
				{Name: "subnet-b", Region: "us-east", CIDR: "172.16.2.0/24", Type: kops.SubnetTypePublic},
			}},
		},
	}
	ig := &kops.InstanceGroup{
		ObjectMeta: metav1.ObjectMeta{Name: "nodes-us-east"},
		Spec: kops.InstanceGroupSpec{
			Role:        kops.InstanceGroupRoleNode,
			Subnets:     []string{"subnet-a", "subnet-b"},
			MachineType: "g6-standard-2",
			Image:       "linode/ubuntu24.04",
			MinSize:     &minSize,
		},
	}
	context := newLinodeInstanceModelBuilderContext(cluster)
	addBootstrapPrerequisites(context)

	networkBuilder := &NetworkModelBuilder{
		LinodeModelContext: &LinodeModelContext{KopsModelContext: contextModel(cluster, []*kops.InstanceGroup{ig})},
		Lifecycle:          fi.LifecycleSync,
	}
	if err := networkBuilder.Build(context); err != nil {
		t.Fatalf("NetworkModelBuilder.Build returned error: %v", err)
	}

	builder := &InstanceModelBuilder{
		LinodeModelContext: &LinodeModelContext{KopsModelContext: contextModel(cluster, []*kops.InstanceGroup{ig})},
		Lifecycle:          fi.LifecycleSync,
		BootstrapScriptBuilder: &model.BootstrapScriptBuilder{
			Lifecycle:        fi.LifecycleSync,
			KopsModelContext: contextModel(cluster, []*kops.InstanceGroup{ig}),
		},
	}

	err := builder.Build(context)
	if err == nil {
		t.Fatalf("expected validation error")
	}
	if !strings.Contains(err.Error(), "expected exactly one subnet for InstanceGroup") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInstanceModelBuilderBuildDerivesPrivateSubnetInterfacePolicy(t *testing.T) {
	cluster := &kops.Cluster{
		ObjectMeta: metav1.ObjectMeta{Name: "example.k8s.local"},
		Spec: kops.ClusterSpec{
			CloudProvider: kops.CloudProviderSpec{Linode: &kops.LinodeSpec{}},
			Networking: kops.NetworkingSpec{Subnets: []kops.ClusterSubnetSpec{{
				Name:   "subnet-a",
				Region: "us-east",
				CIDR:   "172.16.1.0/24",
				Type:   kops.SubnetTypePrivate,
			}}},
		},
	}
	associatePublicIP := true
	ig := &kops.InstanceGroup{
		ObjectMeta: metav1.ObjectMeta{Name: "nodes-us-east"},
		Spec: kops.InstanceGroupSpec{
			Role:              kops.InstanceGroupRoleNode,
			Subnets:           []string{"subnet-a"},
			MachineType:       "g6-standard-2",
			Image:             "linode/ubuntu24.04",
			MinSize:           new(int32),
			AssociatePublicIP: &associatePublicIP,
		},
	}
	context := newLinodeInstanceModelBuilderContext(cluster)
	addBootstrapPrerequisites(context)

	networkBuilder := &NetworkModelBuilder{
		LinodeModelContext: &LinodeModelContext{KopsModelContext: contextModel(cluster, []*kops.InstanceGroup{ig})},
		Lifecycle:          fi.LifecycleSync,
	}
	if err := networkBuilder.Build(context); err != nil {
		t.Fatalf("NetworkModelBuilder.Build returned error: %v", err)
	}

	publicKey := fi.Resource(fi.NewStringResource(testSSHPublicKey))
	context.AddTask(&linodetasks.SSHKey{
		Name:      new("example-k8s-local-default"),
		Lifecycle: fi.LifecycleSync,
		PublicKey: &publicKey,
	})

	builder := &InstanceModelBuilder{
		LinodeModelContext: &LinodeModelContext{KopsModelContext: contextModel(cluster, []*kops.InstanceGroup{ig})},
		Lifecycle:          fi.LifecycleSync,
		BootstrapScriptBuilder: &model.BootstrapScriptBuilder{
			Lifecycle:        fi.LifecycleSync,
			KopsModelContext: contextModel(cluster, []*kops.InstanceGroup{ig}),
		},
	}

	if err := builder.Build(context); err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	instanceTask := findInstanceTask(t, context, ig.Name)
	if got, want := fi.ValueOf(instanceTask.RequirePublicInterface), false; got != want {
		t.Fatalf("unexpected RequirePublicInterface: got %t, want %t", got, want)
	}
}

func contextModel(cluster *kops.Cluster, instanceGroups []*kops.InstanceGroup) *model.KopsModelContext {
	return &model.KopsModelContext{
		IAMModelContext:   iam.IAMModelContext{Cluster: cluster},
		AllInstanceGroups: instanceGroups,
		InstanceGroups:    instanceGroups,
	}
}

func newLinodeInstanceModelBuilderContext(cluster *kops.Cluster) *fi.CloudupModelBuilderContext {
	return &fi.CloudupModelBuilderContext{Tasks: map[string]fi.CloudupTask{}}
}

func addBootstrapPrerequisites(context *fi.CloudupModelBuilderContext) {
	context.AddTask(&fitasks.Keypair{
		Name:    new(fi.CertificateIDCA),
		Subject: "cn=kubernetes",
		Type:    "ca",
	})
}

func findInstanceTask(t *testing.T, context *fi.CloudupModelBuilderContext, name string) *linodetasks.Instance {
	t.Helper()
	for _, task := range context.Tasks {
		instanceTask, ok := task.(*linodetasks.Instance)
		if !ok {
			continue
		}
		if fi.ValueOf(instanceTask.Name) == name {
			return instanceTask
		}
	}

	t.Fatalf("Instance task %q not found", name)
	return nil
}
