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
)

const testSSHPublicKey = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCySdqIU+FhCWl3BNrAvPaOe5VfL2aCARUWwy91ZP+T7LBwFa9lhdttfjp/VX1D1/PVwntn2EhN079m8c2kfdmiZ/iCHqrLyIGSd+BOiCz0lT47znvANSfxYjLUuKrWWWeaXqerJkOsAD4PHchRLbZGPdbfoBKwtb/WT4GMRQmb9vmiaZYjsfdPPM9KkWI9ECoWFGjGehA8D+iYIPR711kRacb1xdYmnjHqxAZHFsb5L8wDWIeAyhy49cBD+lbzTiioq2xWLorXuFmXh6Do89PgzvHeyCLY6816f/kCX6wIFts8A2eaEHFL4rAOsuh6qHmSxGCR9peSyuRW8DxV725x justin@test"

func TestSSHKeyModelBuilderBuildWithPublicKey(t *testing.T) {
	sshKeyName := "custom.ssh:key"
	cluster := &kops.Cluster{
		ObjectMeta: metav1.ObjectMeta{Name: "example.k8s.local"},
		Spec:       kops.ClusterSpec{SSHKeyName: fi.PtrTo(sshKeyName)},
	}
	b := &SSHKeyModelBuilder{
		LinodeModelContext: &LinodeModelContext{KopsModelContext: &model.KopsModelContext{
			IAMModelContext: iam.IAMModelContext{Cluster: cluster},
			SSHPublicKeys:   [][]byte{[]byte(testSSHPublicKey)},
		}},
		Lifecycle: fi.LifecycleSync,
	}
	context := &fi.CloudupModelBuilderContext{Tasks: map[string]fi.CloudupTask{}}

	if err := b.Build(context); err != nil {
		t.Fatalf("Build returned error: %v", err)
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

func TestSSHKeyModelBuilderBuildWithExistingKeyName(t *testing.T) {
	sshKeyName := "existing.ssh:key"
	cluster := &kops.Cluster{
		ObjectMeta: metav1.ObjectMeta{Name: "example.k8s.local"},
		Spec:       kops.ClusterSpec{SSHKeyName: fi.PtrTo(sshKeyName)},
	}
	b := &SSHKeyModelBuilder{
		LinodeModelContext: &LinodeModelContext{KopsModelContext: &model.KopsModelContext{
			IAMModelContext: iam.IAMModelContext{Cluster: cluster},
		}},
		Lifecycle: fi.LifecycleSync,
	}
	context := &fi.CloudupModelBuilderContext{Tasks: map[string]fi.CloudupTask{}}

	if err := b.Build(context); err != nil {
		t.Fatalf("Build returned error: %v", err)
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

func TestSSHKeyModelBuilderBuildTruncatesLongGeneratedName(t *testing.T) {
	cluster := &kops.Cluster{
		ObjectMeta: metav1.ObjectMeta{Name: strings.Repeat("a", 32)},
	}
	b := &SSHKeyModelBuilder{
		LinodeModelContext: &LinodeModelContext{KopsModelContext: &model.KopsModelContext{
			IAMModelContext: iam.IAMModelContext{Cluster: cluster},
			SSHPublicKeys:   [][]byte{[]byte(testSSHPublicKey)},
		}},
		Lifecycle: fi.LifecycleSync,
	}
	context := &fi.CloudupModelBuilderContext{Tasks: map[string]fi.CloudupTask{}}

	if err := b.Build(context); err != nil {
		t.Fatalf("Build returned error: %v", err)
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
