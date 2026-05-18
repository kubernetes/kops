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
	"encoding/base64"
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/linode/linodego"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/linode"
)

func TestInstanceFindMatch(t *testing.T) {
	client := &linode.MockLinodeClient{
		ListInstancesResponse: []linodego.Instance{
			{ID: 1001, Label: "nodes-example-1", Region: "us-east", Type: "g6-standard-2", Image: "linode/ubuntu24.04", Tags: []string{"kops.k8s.io/cluster:example.k8s.local", "kops.k8s.io/instance-group:nodes-us-east"}},
			{ID: 1002, Label: "nodes-example-2", Region: "us-east", Type: "g6-standard-2", Image: "linode/ubuntu24.04", Tags: []string{"kops.k8s.io/cluster:example.k8s.local", "kops.k8s.io/instance-group:nodes-us-east"}},
			{ID: 2001, Label: "other-ig", Region: "us-east", Type: "g6-standard-2", Image: "linode/ubuntu24.04", Tags: []string{"kops.k8s.io/cluster:example.k8s.local", "kops.k8s.io/instance-group:control-plane-us-east"}},
		},
	}
	cloud := &linode.MockLinodeCloud{Client_: client}
	ctx := newTestCloudupContext(t, cloud)

	task := &Instance{
		Name:  fi.PtrTo("nodes-us-east.example.k8s.local"),
		Tags:  []string{"kops.k8s.io/cluster:example.k8s.local", "kops.k8s.io/instance-group:nodes-us-east"},
		Count: 2,
	}

	actual, err := task.Find(ctx)
	if err != nil {
		t.Fatalf("Find returned error: %v", err)
	}
	if actual == nil {
		t.Fatalf("expected to find matching instances")
	}
	if got, want := actual.Count, 2; got != want {
		t.Fatalf("unexpected count: got %d, want %d", got, want)
	}
	if got, want := fi.ValueOf(actual.Region), "us-east"; got != want {
		t.Fatalf("unexpected region: got %q, want %q", got, want)
	}
	if got, want := fi.ValueOf(actual.Type), "g6-standard-2"; got != want {
		t.Fatalf("unexpected type: got %q, want %q", got, want)
	}
}

func TestInstanceFindListError(t *testing.T) {
	client := &linode.MockLinodeClient{ListInstancesError: errors.New("api unavailable")}
	cloud := &linode.MockLinodeCloud{Client_: client}
	ctx := newTestCloudupContext(t, cloud)

	task := &Instance{Tags: []string{"kops.k8s.io/cluster:example.k8s.local"}}
	_, err := task.Find(ctx)
	if err == nil {
		t.Fatalf("expected list error")
	}
	if !strings.Contains(err.Error(), "error listing Linode (Akamai) instances") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInstanceRenderLinodeCreate(t *testing.T) {
	client := &linode.MockLinodeClient{}
	target := linode.NewAPITarget(&linode.MockLinodeCloud{Client_: client})

	userData := fi.Resource(fi.NewStringResource("#cloud-config\nruncmd:\n  - echo hello\n"))
	publicKey := fi.Resource(fi.NewStringResource(testOpenSSHPublicKey))
	expected := &Instance{
		Name:          fi.PtrTo("nodes.us-east.example.k8s.local"),
		Region:        fi.PtrTo("us-east"),
		Type:          fi.PtrTo("g6-standard-2"),
		Image:         fi.PtrTo("linode/ubuntu24.04"),
		Count:         2,
		Tags:          []string{"kops.k8s.io/cluster:example.k8s.local", "kops.k8s.io/instance-group:nodes-us-east", "kops.k8s.io/instance-role:Node"},
		AuthorizedKey: &publicKey,
		UserData:      &userData,
	}

	if err := (&Instance{}).RenderLinode(target, nil, expected, nil); err != nil {
		t.Fatalf("RenderLinode returned error: %v", err)
	}

	if got, want := client.CreateInstanceCalls, 2; got != want {
		t.Fatalf("unexpected create calls: got %d, want %d", got, want)
	}
	if got, want := client.LastCreateInstanceOpts.Region, "us-east"; got != want {
		t.Fatalf("unexpected region: got %q, want %q", got, want)
	}
	if got, want := client.LastCreateInstanceOpts.Type, "g6-standard-2"; got != want {
		t.Fatalf("unexpected type: got %q, want %q", got, want)
	}
	if got, want := client.LastCreateInstanceOpts.Image, "linode/ubuntu24.04"; got != want {
		t.Fatalf("unexpected image: got %q, want %q", got, want)
	}
	if got := client.LastCreateInstanceOpts.RootPass; got == "" {
		t.Fatalf("expected root password to be populated")
	}
	if got, want := client.LastCreateInstanceOpts.PrivateIP, true; got != want {
		t.Fatalf("unexpected private IP setting: got %t, want %t", got, want)
	}
	if got, want := len(client.LastCreateInstanceOpts.AuthorizedKeys), 1; got != want {
		t.Fatalf("unexpected authorized key count: got %d, want %d", got, want)
	}
	if got, want := client.LastCreateInstanceOpts.AuthorizedKeys[0], testOpenSSHPublicKey; got != want {
		t.Fatalf("unexpected authorized key: got %q, want %q", got, want)
	}
	if client.LastCreateInstanceOpts.Metadata == nil {
		t.Fatalf("expected metadata to be configured")
	}
	decodedUserData, err := base64.StdEncoding.DecodeString(client.LastCreateInstanceOpts.Metadata.UserData)
	if err != nil {
		t.Fatalf("failed to decode metadata user data: %v", err)
	}
	if got, want := string(decodedUserData), "#cloud-config\nruncmd:\n  - echo hello\n"; got != want {
		t.Fatalf("unexpected user data payload: got %q, want %q", got, want)
	}
}

func TestInstanceRenderLinodeScaleDownNotSupported(t *testing.T) {
	client := &linode.MockLinodeClient{}
	target := linode.NewAPITarget(&linode.MockLinodeCloud{Client_: client})

	actual := &Instance{Count: 2}
	expected := &Instance{
		Name:   fi.PtrTo("nodes.example.k8s.local"),
		Region: fi.PtrTo("us-east"),
		Type:   fi.PtrTo("g6-standard-2"),
		Image:  fi.PtrTo("linode/ubuntu24.04"),
		Count:  1,
	}

	err := (&Instance{}).RenderLinode(target, actual, expected, nil)
	if err == nil {
		t.Fatalf("expected scale-down error")
	}
	if !strings.Contains(err.Error(), "decreasing Linode (Akamai) instance count") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMakeInstanceLabel(t *testing.T) {
	label := makeInstanceLabel("nodes__us..east@@example.k8s.local", 17)

	if len(label) > 64 {
		t.Fatalf("label too long: %d characters", len(label))
	}
	if ok, err := regexp.MatchString(`^[a-z0-9][a-z0-9._-]*[a-z0-9]$`, label); err != nil || !ok {
		t.Fatalf("label does not satisfy Linode format: %q", label)
	}
	if strings.Contains(label, "--") || strings.Contains(label, "__") || strings.Contains(label, "..") {
		t.Fatalf("label should not contain repeated separators: %q", label)
	}
}
