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

package linodetasks

import (
	"errors"
	"strings"
	"testing"

	"github.com/linode/linodego"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/linode"
)

const testLinodeSSHPublicKey = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCySdqIU+FhCWl3BNrAvPaOe5VfL2aCARUWwy91ZP+T7LBwFa9lhdttfjp/VX1D1/PVwntn2EhN079m8c2kfdmiZ/iCHqrLyIGSd+BOiCz0lT47znvANSfxYjLUuKrWWWeaXqerJkOsAD4PHchRLbZGPdbfoBKwtb/WT4GMRQmb9vmiaZYjsfdPPM9KkWI9ECoWFGjGehA8D+iYIPR711kRacb1xdYmnjHqxAZHFsb5L8wDWIeAyhy49cBD+lbzTiioq2xWLorXuFmXh6Do89PgzvHeyCLY6816f/kCX6wIFts8A2eaEHFL4rAOsuh6qHmSxGCR9peSyuRW8DxV725x justin@test"

func newSSHKeyResource(t *testing.T, contents string) *fi.Resource {
	t.Helper()
	r := fi.Resource(fi.NewStringResource(contents))
	return &r
}

func TestSSHKeyFindMatchByName(t *testing.T) {
	publicKey := newSSHKeyResource(t, testLinodeSSHPublicKey)
	client := &linode.MockLinodeClient{
		ListSSHKeysResponse: []linodego.SSHKey{
			{ID: 101, Label: "example-k8s-local", SSHKey: testLinodeSSHPublicKey},
			{ID: 102, Label: "other", SSHKey: testLinodeSSHPublicKey},
		},
	}
	cloud := &linode.MockLinodeCloud{Client_: client}
	ctx := newTestCloudupContext(t, cloud)

	task := &SSHKey{Name: fi.PtrTo("example-k8s-local"), PublicKey: publicKey}
	actual, err := task.Find(ctx)
	if err != nil {
		t.Fatalf("Find returned error: %v", err)
	}
	if actual == nil {
		t.Fatalf("expected to find SSH key")
	}
	if got, want := fi.ValueOf(actual.ID), 101; got != want {
		t.Fatalf("unexpected SSH key ID: got %d, want %d", got, want)
	}
	if actual.PublicKey == nil {
		t.Fatalf("expected matched SSH key to carry public key resource")
	}
}

func TestSSHKeyFindPublicKeyMismatch(t *testing.T) {
	publicKey := newSSHKeyResource(t, testLinodeSSHPublicKey)
	client := &linode.MockLinodeClient{
		ListSSHKeysResponse: []linodego.SSHKey{{ID: 101, Label: "example-k8s-local", SSHKey: "ssh-rsa mismatch"}},
	}
	cloud := &linode.MockLinodeCloud{Client_: client}
	ctx := newTestCloudupContext(t, cloud)

	_, err := (&SSHKey{Name: fi.PtrTo("example-k8s-local"), PublicKey: publicKey}).Find(ctx)
	if err == nil {
		t.Fatalf("expected public key mismatch error")
	}
	if !strings.Contains(err.Error(), "public key data did not match") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSSHKeyFindDuplicateName(t *testing.T) {
	client := &linode.MockLinodeClient{
		ListSSHKeysResponse: []linodego.SSHKey{
			{ID: 101, Label: "example-k8s-local"},
			{ID: 102, Label: "example-k8s-local"},
		},
	}
	cloud := &linode.MockLinodeCloud{Client_: client}
	ctx := newTestCloudupContext(t, cloud)

	_, err := (&SSHKey{Name: fi.PtrTo("example-k8s-local")}).Find(ctx)
	if err == nil {
		t.Fatalf("expected duplicate name error")
	}
	if !strings.Contains(err.Error(), "found multiple SSH keys named") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSSHKeyFindListError(t *testing.T) {
	client := &linode.MockLinodeClient{ListSSHKeysError: errors.New("api unavailable")}
	cloud := &linode.MockLinodeCloud{Client_: client}
	ctx := newTestCloudupContext(t, cloud)

	_, err := (&SSHKey{Name: fi.PtrTo("example-k8s-local")}).Find(ctx)
	if err == nil {
		t.Fatalf("expected list error")
	}
	if !strings.Contains(err.Error(), "error listing Linode (Akamai) SSH keys") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSSHKeyRenderLinodeCreate(t *testing.T) {
	publicKey := newSSHKeyResource(t, testLinodeSSHPublicKey)
	client := &linode.MockLinodeClient{CreateSSHKeyResponse: &linodego.SSHKey{ID: 42, Label: "example-k8s-local"}}
	target := linode.NewAPITarget(&linode.MockLinodeCloud{Client_: client})

	expected := &SSHKey{Name: fi.PtrTo("example-k8s-local"), PublicKey: publicKey}
	if err := (&SSHKey{}).RenderLinode(target, nil, expected, nil); err != nil {
		t.Fatalf("RenderLinode returned error: %v", err)
	}
	if got, want := client.CreateSSHKeyCalls, 1; got != want {
		t.Fatalf("unexpected create calls: got %d, want %d", got, want)
	}
	if got, want := client.LastCreateSSHKeyOpts.Label, "example-k8s-local"; got != want {
		t.Fatalf("unexpected create label: got %q, want %q", got, want)
	}
	if got, want := client.LastCreateSSHKeyOpts.SSHKey, testLinodeSSHPublicKey; got != want {
		t.Fatalf("unexpected create public key: got %q, want %q", got, want)
	}
	if got, want := fi.ValueOf(expected.ID), 42; got != want {
		t.Fatalf("expected task ID to be populated from create response: got %d, want %d", got, want)
	}
}

func TestSSHKeyCheckChangesRejectsPublicKeyChange(t *testing.T) {
	actualPublicKey := newSSHKeyResource(t, testLinodeSSHPublicKey)
	expectedPublicKey := newSSHKeyResource(t, testLinodeSSHPublicKey+"-changed")
	actual := &SSHKey{Name: fi.PtrTo("example-k8s-local"), ID: fi.PtrTo(42), PublicKey: actualPublicKey}
	expected := &SSHKey{Name: fi.PtrTo("example-k8s-local"), ID: fi.PtrTo(42), PublicKey: expectedPublicKey}
	changes := &SSHKey{PublicKey: expected.PublicKey}

	if err := (&SSHKey{}).CheckChanges(actual, expected, changes); err == nil {
		t.Fatalf("expected public key change to be rejected")
	}
}
