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

package linodetasks

import (
	"errors"
	"strings"
	"testing"

	"github.com/linode/linodego"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/linode"
)

const testOpenSSHPublicKey = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCySdqIU+FhCWl3BNrAvPaOe5VfL2aCARUWwy91ZP+T7LBwFa9lhdttfjp/VX1D1/PVwntn2EhN079m8c2kfdmiZ/iCHqrLyIGSd+BOiCz0lT47znvANSfxYjLUuKrWWWeaXqerJkOsAD4PHchRLbZGPdbfoBKwtb/WT4GMRQmb9vmiaZYjsfdPPM9KkWI9ECoWFGjGehA8D+iYIPR711kRacb1xdYmnjHqxAZHFsb5L8wDWIeAyhy49cBD+lbzTiioq2xWLorXuFmXh6Do89PgzvHeyCLY6816f/kCX6wIFts8A2eaEHFL4rAOsuh6qHmSxGCR9peSyuRW8DxV725x justin@test"

func newTestPublicKeyResource() *fi.Resource {
	resource := fi.Resource(fi.NewStringResource(testOpenSSHPublicKey))
	return &resource
}

func TestSSHKeyFindMatch(t *testing.T) {
	publicKey := newTestPublicKeyResource()
	client := &linode.MockLinodeClient{
		ListSSHKeysResponse: []linodego.SSHKey{{
			ID:     123,
			Label:  "kubernetes.example.k8s.local-1234",
			SSHKey: testOpenSSHPublicKey,
		}},
	}
	cloud := &linode.MockLinodeCloud{Client_: client}
	ctx := newTestCloudupContext(t, cloud)

	task := &SSHKey{
		Name:      fi.PtrTo("kubernetes.example.k8s.local-1234"),
		PublicKey: publicKey,
		Lifecycle: fi.LifecycleSync,
	}

	actual, err := task.Find(ctx)
	if err != nil {
		t.Fatalf("Find returned error: %v", err)
	}
	if actual == nil {
		t.Fatalf("expected to find SSH key")
	}
	if got, want := fi.ValueOf(actual.ID), 123; got != want {
		t.Fatalf("unexpected ID: got %d, want %d", got, want)
	}
}

func TestSSHKeyFindDuplicate(t *testing.T) {
	client := &linode.MockLinodeClient{
		ListSSHKeysResponse: []linodego.SSHKey{
			{ID: 1, Label: "kubernetes.example.k8s.local-1234", SSHKey: "ssh-rsa AAAA test"},
			{ID: 2, Label: "kubernetes.example.k8s.local-1234", SSHKey: "ssh-rsa AAAA test"},
		},
	}
	cloud := &linode.MockLinodeCloud{Client_: client}
	ctx := newTestCloudupContext(t, cloud)

	task := &SSHKey{Name: fi.PtrTo("kubernetes.example.k8s.local-1234")}
	_, err := task.Find(ctx)
	if err == nil {
		t.Fatalf("expected duplicate name error")
	}
	if !strings.Contains(err.Error(), "found multiple SSH keys named") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSSHKeyFindPublicKeyMismatch(t *testing.T) {
	publicKey := newTestPublicKeyResource()
	client := &linode.MockLinodeClient{
		ListSSHKeysResponse: []linodego.SSHKey{{
			ID:     123,
			Label:  "kubernetes.example.k8s.local-1234",
			SSHKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDbadkey",
		}},
	}
	cloud := &linode.MockLinodeCloud{Client_: client}
	ctx := newTestCloudupContext(t, cloud)

	task := &SSHKey{Name: fi.PtrTo("kubernetes.example.k8s.local-1234"), PublicKey: publicKey}
	_, err := task.Find(ctx)
	if err == nil {
		t.Fatalf("expected mismatch error")
	}
	if !strings.Contains(err.Error(), "public key data did not match") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSSHKeyFindListError(t *testing.T) {
	client := &linode.MockLinodeClient{ListSSHKeysError: errors.New("api unavailable")}
	cloud := &linode.MockLinodeCloud{Client_: client}
	ctx := newTestCloudupContext(t, cloud)

	task := &SSHKey{Name: fi.PtrTo("kubernetes.example.k8s.local-1234")}
	_, err := task.Find(ctx)
	if err == nil {
		t.Fatalf("expected list error")
	}
	if !strings.Contains(err.Error(), "error listing Linode (Akamai) SSH keys") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSSHKeyRenderLinodeCreate(t *testing.T) {
	publicKey := newTestPublicKeyResource()
	client := &linode.MockLinodeClient{CreateSSHKeyResponse: &linodego.SSHKey{ID: 42, Label: "kubernetes.example.k8s.local-1234"}}
	cloud := &linode.MockLinodeCloud{Client_: client}
	target := linode.NewAPITarget(cloud)

	expected := &SSHKey{Name: fi.PtrTo("kubernetes.example.k8s.local-1234"), PublicKey: publicKey}
	err := (&SSHKey{}).RenderLinode(target, nil, expected, nil)
	if err != nil {
		t.Fatalf("RenderLinode returned error: %v", err)
	}
	if got, want := client.CreateSSHKeyCalls, 1; got != want {
		t.Fatalf("unexpected create calls: got %d, want %d", got, want)
	}
	if got, want := client.LastCreateSSHKeyOpts.Label, "kubernetes.example.k8s.local-1234"; got != want {
		t.Fatalf("unexpected create label: got %q, want %q", got, want)
	}
	if fi.ValueOf(expected.ID) != 42 {
		t.Fatalf("expected task ID to be populated from create response")
	}
}

func TestSSHKeyRenderLinodeNoopWhenActualExists(t *testing.T) {
	publicKey := newTestPublicKeyResource()
	client := &linode.MockLinodeClient{}
	cloud := &linode.MockLinodeCloud{Client_: client}
	target := linode.NewAPITarget(cloud)

	actual := &SSHKey{Name: fi.PtrTo("kubernetes.example.k8s.local-1234"), ID: fi.PtrTo(11)}
	expected := &SSHKey{Name: fi.PtrTo("kubernetes.example.k8s.local-1234"), PublicKey: publicKey}
	err := (&SSHKey{}).RenderLinode(target, actual, expected, nil)
	if err != nil {
		t.Fatalf("RenderLinode returned error: %v", err)
	}
	if got := client.CreateSSHKeyCalls; got != 0 {
		t.Fatalf("unexpected create calls: got %d, want 0", got)
	}
}

func TestSSHKeyRenderLinodeRequiresPublicKey(t *testing.T) {
	client := &linode.MockLinodeClient{}
	cloud := &linode.MockLinodeCloud{Client_: client}
	target := linode.NewAPITarget(cloud)

	expected := &SSHKey{Name: fi.PtrTo("kubernetes.example.k8s.local-1234")}
	err := (&SSHKey{}).RenderLinode(target, nil, expected, nil)
	if err == nil {
		t.Fatalf("expected missing PublicKey error")
	}
	if !strings.Contains(err.Error(), "PublicKey") {
		t.Fatalf("unexpected error: %v", err)
	}
}
