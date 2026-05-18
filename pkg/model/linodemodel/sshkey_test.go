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

package linodemodel

import (
	"regexp"
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/iam"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/linodetasks"
)

var linodeSSHKeyLabelRegex = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

const testOpenSSHPublicKey = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCySdqIU+FhCWl3BNrAvPaOe5VfL2aCARUWwy91ZP+T7LBwFa9lhdttfjp/VX1D1/PVwntn2EhN079m8c2kfdmiZ/iCHqrLyIGSd+BOiCz0lT47znvANSfxYjLUuKrWWWeaXqerJkOsAD4PHchRLbZGPdbfoBKwtb/WT4GMRQmb9vmiaZYjsfdPPM9KkWI9ECoWFGjGehA8D+iYIPR711kRacb1xdYmnjHqxAZHFsb5L8wDWIeAyhy49cBD+lbzTiioq2xWLorXuFmXh6Do89PgzvHeyCLY6816f/kCX6wIFts8A2eaEHFL4rAOsuh6qHmSxGCR9peSyuRW8DxV725x justin@test"

func TestSSHKeyModelBuilderBuild(t *testing.T) {
	cluster := &kops.Cluster{ObjectMeta: metav1.ObjectMeta{Name: "example.k8s.local"}}

	builder := &SSHKeyModelBuilder{
		LinodeModelContext: &LinodeModelContext{
			KopsModelContext: &model.KopsModelContext{
				IAMModelContext: iam.IAMModelContext{Cluster: cluster},
				SSHPublicKeys:   [][]byte{[]byte(testOpenSSHPublicKey)},
			},
		},
		Lifecycle: fi.LifecycleSync,
	}

	context := &fi.CloudupModelBuilderContext{Tasks: map[string]fi.CloudupTask{}}
	if err := builder.Build(context); err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	if got, want := len(context.Tasks), 1; got != want {
		t.Fatalf("unexpected task count: got %d, want %d", got, want)
	}

	for _, task := range context.Tasks {
		sshKeyTask, ok := task.(*linodetasks.SSHKey)
		if !ok {
			t.Fatalf("unexpected task type: %T", task)
		}

		got := fi.ValueOf(sshKeyTask.Name)
		if !strings.HasPrefix(got, "kubernetes-example-k8s-local-") {
			t.Fatalf("unexpected generated name prefix: %q", got)
		}
		if len(got) > 64 {
			t.Fatalf("generated name exceeded max length: %d", len(got))
		}
		if !linodeSSHKeyLabelRegex.MatchString(got) {
			t.Fatalf("generated name has invalid characters: %q", got)
		}
		if sshKeyTask.PublicKey == nil {
			t.Fatalf("expected public key resource to be set")
		}
		if got, want := sshKeyTask.Lifecycle, fi.LifecycleSync; got != want {
			t.Fatalf("unexpected lifecycle: got %q, want %q", got, want)
		}
	}
}

func TestSSHKeyModelBuilderBuild_CustomNameNormalized(t *testing.T) {
	customName := "custom.key:name"
	cluster := &kops.Cluster{ObjectMeta: metav1.ObjectMeta{Name: "example.k8s.local"}}
	cluster.Spec.SSHKeyName = fi.PtrTo(customName)

	builder := &SSHKeyModelBuilder{
		LinodeModelContext: &LinodeModelContext{
			KopsModelContext: &model.KopsModelContext{
				IAMModelContext: iam.IAMModelContext{Cluster: cluster},
				SSHPublicKeys:   [][]byte{[]byte(testOpenSSHPublicKey)},
			},
		},
		Lifecycle: fi.LifecycleSync,
	}

	context := &fi.CloudupModelBuilderContext{Tasks: map[string]fi.CloudupTask{}}
	if err := builder.Build(context); err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	for _, task := range context.Tasks {
		sshKeyTask, ok := task.(*linodetasks.SSHKey)
		if !ok {
			t.Fatalf("unexpected task type: %T", task)
		}

		if got, want := fi.ValueOf(sshKeyTask.Name), "custom-key-name"; got != want {
			t.Fatalf("unexpected normalized custom name: got %q, want %q", got, want)
		}
	}
}
