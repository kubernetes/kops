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

const testSSHPublicKey = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCySdqIU+FhCWl3BNrAvPaOe5VfL2aCARUWwy91ZP+T7LBwFa9lhdttfjp/VX1D1/PVwntn2EhN079m8c2kfdmiZ/iCHqrLyIGSd+BOiCz0lT47znvANSfxYjLUuKrWWWeaXqerJkOsAD4PHchRLbZGPdbfoBKwtb/WT4GMRQmb9vmiaZYjsfdPPM9KkWI9ECoWFGjGehA8D+iYIPR711kRacb1xdYmnjHqxAZHFsb5L8wDWIeAyhy49cBD+lbzTiioq2xWLorXuFmXh6Do89PgzvHeyCLY6816f/kCX6wIFts8A2eaEHFL4rAOsuh6qHmSxGCR9peSyuRW8DxV725x justin@test"

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
	secondIG := &kops.InstanceGroup{
		ObjectMeta: metav1.ObjectMeta{Name: "nodes-us-east-2"},
		Spec: kops.InstanceGroupSpec{
			Role:        kops.InstanceGroupRoleNode,
			Subnets:     []string{"subnet-a"},
			MachineType: "g6-standard-2",
			Image:       "linode/ubuntu24.04",
			MinSize:     new(int32),
		},
	}
	context := newLinodeInstanceModelBuilderContext(cluster)
	addBootstrapPrerequisites(context)
	modelContext := contextModel(cluster, []*kops.InstanceGroup{ig, secondIG})
	modelContext.SSHPublicKeys = [][]byte{[]byte(testSSHPublicKey)}

	networkBuilder := &NetworkModelBuilder{
		LinodeModelContext: &LinodeModelContext{KopsModelContext: modelContext},
		Lifecycle:          fi.LifecycleSync,
	}
	if err := networkBuilder.Build(context); err != nil {
		t.Fatalf("NetworkModelBuilder.Build returned error: %v", err)
	}

	builder := &InstanceModelBuilder{
		LinodeModelContext: &LinodeModelContext{KopsModelContext: modelContext},
		Lifecycle:          fi.LifecycleSync,
		SSHKeyLifecycle:    fi.LifecycleSync,
		BootstrapScriptBuilder: &model.BootstrapScriptBuilder{
			Lifecycle:        fi.LifecycleSync,
			KopsModelContext: modelContext,
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
	secondInstanceTask := findInstanceTask(t, context, secondIG.Name)
	if got, want := len(secondInstanceTask.AuthorizedKeys), 1; got != want {
		t.Fatalf("unexpected second instance authorized key count: got %d, want %d", got, want)
	}
	if instanceTask.AuthorizedKeys[0] != secondInstanceTask.AuthorizedKeys[0] {
		t.Fatalf("expected instance groups to share the SSH key task")
	}

	sshKeyTaskCount := 0
	for _, task := range context.Tasks {
		if _, ok := task.(*linodetasks.SSHKey); ok {
			sshKeyTaskCount++
		}
	}
	if got, want := sshKeyTaskCount, 1; got != want {
		t.Fatalf("unexpected SSH key task count: got %d, want %d", got, want)
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

func TestInstanceModelBuilderBuildSSHKeyTaskWithPublicKey(t *testing.T) {
	sshKeyName := "custom.ssh:key"
	cluster := &kops.Cluster{
		ObjectMeta: metav1.ObjectMeta{Name: "example.k8s.local"},
		Spec:       kops.ClusterSpec{SSHKeyName: new(sshKeyName)},
	}
	b := &InstanceModelBuilder{
		LinodeModelContext: &LinodeModelContext{KopsModelContext: &model.KopsModelContext{
			IAMModelContext: iam.IAMModelContext{Cluster: cluster},
			SSHPublicKeys:   [][]byte{[]byte(testSSHPublicKey)},
		}},
		SSHKeyLifecycle: fi.LifecycleSync,
	}
	context := &fi.CloudupModelBuilderContext{Tasks: map[string]fi.CloudupTask{}}

	if _, err := b.buildSSHKeyTask(context); err != nil {
		t.Fatalf("buildSSHKeyTask returned error: %v", err)
	}

	if got, want := len(context.Tasks), 1; got != want {
		t.Fatalf("unexpected task count: got %d, want %d", got, want)
	}

	for _, task := range context.Tasks {
		sshKey, ok := task.(*linodetasks.SSHKey)
		if !ok {
			t.Fatalf("expected SSHKey task, got %T", task)
		}
		if got, want := fi.ValueOf(sshKey.Name), linode.NormalizeLinodeLabel(sshKeyName); got != want {
			t.Fatalf("unexpected SSH key name: got %q, want %q", got, want)
		}
		if sshKey.PublicKey == nil {
			t.Fatalf("expected SSH public key resource")
		}
		publicKey, err := fi.ResourceAsString(*sshKey.PublicKey)
		if err != nil {
			t.Fatalf("ResourceAsString returned error: %v", err)
		}
		if got, want := publicKey, testSSHPublicKey; got != want {
			t.Fatalf("unexpected SSH public key: got %q, want %q", got, want)
		}
		if got, want := sshKey.Lifecycle, fi.LifecycleSync; got != want {
			t.Fatalf("unexpected lifecycle: got %q, want %q", got, want)
		}
	}
}

func TestInstanceModelBuilderBuildSSHKeyTaskWithExistingKeyName(t *testing.T) {
	sshKeyName := "existing.ssh:key"
	cluster := &kops.Cluster{
		ObjectMeta: metav1.ObjectMeta{Name: "example.k8s.local"},
		Spec:       kops.ClusterSpec{SSHKeyName: new(sshKeyName)},
	}
	b := &InstanceModelBuilder{
		LinodeModelContext: &LinodeModelContext{KopsModelContext: &model.KopsModelContext{
			IAMModelContext: iam.IAMModelContext{Cluster: cluster},
		}},
		SSHKeyLifecycle: fi.LifecycleSync,
	}
	context := &fi.CloudupModelBuilderContext{Tasks: map[string]fi.CloudupTask{}}

	if _, err := b.buildSSHKeyTask(context); err != nil {
		t.Fatalf("buildSSHKeyTask returned error: %v", err)
	}

	if got, want := len(context.Tasks), 1; got != want {
		t.Fatalf("unexpected task count: got %d, want %d", got, want)
	}

	for _, task := range context.Tasks {
		sshKey, ok := task.(*linodetasks.SSHKey)
		if !ok {
			t.Fatalf("expected SSHKey task, got %T", task)
		}
		if got, want := fi.ValueOf(sshKey.Name), linode.NormalizeLinodeLabel(sshKeyName); got != want {
			t.Fatalf("unexpected SSH key name: got %q, want %q", got, want)
		}
		if sshKey.PublicKey != nil {
			t.Fatalf("expected existing key task to omit public key data")
		}
	}
}

func TestInstanceModelBuilderBuildSSHKeyTaskTruncatesLongGeneratedName(t *testing.T) {
	cluster := &kops.Cluster{
		ObjectMeta: metav1.ObjectMeta{Name: strings.Repeat("a", 32)},
	}
	b := &InstanceModelBuilder{
		LinodeModelContext: &LinodeModelContext{KopsModelContext: &model.KopsModelContext{
			IAMModelContext: iam.IAMModelContext{Cluster: cluster},
			SSHPublicKeys:   [][]byte{[]byte(testSSHPublicKey)},
		}},
		SSHKeyLifecycle: fi.LifecycleSync,
	}
	context := &fi.CloudupModelBuilderContext{Tasks: map[string]fi.CloudupTask{}}

	if _, err := b.buildSSHKeyTask(context); err != nil {
		t.Fatalf("buildSSHKeyTask returned error: %v", err)
	}

	prefix := linode.NormalizeLinodeLabel("kubernetes." + cluster.ObjectMeta.Name)
	for _, task := range context.Tasks {
		sshKey, ok := task.(*linodetasks.SSHKey)
		if !ok {
			t.Fatalf("expected SSHKey task, got %T", task)
		}
		name := fi.ValueOf(sshKey.Name)
		if got, want := len(name), 64; got != want {
			t.Fatalf("unexpected SSH key name length: got %d, want %d", got, want)
		}
		if !strings.HasPrefix(name, prefix+"-") {
			t.Fatalf("unexpected SSH key name prefix: got %q, want prefix %q", name, prefix+"-")
		}
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
