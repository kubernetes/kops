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
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/linode/linodego"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/linode"
)

func newTestCloudupContext(t *testing.T, cloud linode.LinodeCloud) *fi.CloudupContext {
	t.Helper()
	target := linode.NewAPITarget(cloud)
	c, err := fi.NewCloudupContext(context.Background(), fi.DeletionProcessingModeDeleteIncludingDeferred, target, nil, cloud, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("NewCloudupContext returned error: %v", err)
	}
	return c
}

func TestVPCFindMatchByName(t *testing.T) {
	client := &linode.MockLinodeClient{
		ListVPCsResponse: []linodego.VPC{
			{ID: 101, Label: "example-k8s-local", Description: "kOps cluster VPC", Region: "us-east"},
			{ID: 102, Label: "other", Region: "us-east"},
		},
	}
	cloud := &linode.MockLinodeCloud{Client_: client}
	ctx := newTestCloudupContext(t, cloud)

	task := &VPC{Name: new("example-k8s-local")}
	actual, err := task.Find(ctx)
	if err != nil {
		t.Fatalf("Find returned error: %v", err)
	}
	if actual == nil {
		t.Fatalf("expected to find VPC")
	}
	if got, want := fi.ValueOf(actual.ID), 101; got != want {
		t.Fatalf("unexpected VPC ID: got %d, want %d", got, want)
	}
	if got, want := fi.ValueOf(actual.Description), "kOps cluster VPC"; got != want {
		t.Fatalf("unexpected description: got %q, want %q", got, want)
	}
	if got, want := fi.ValueOf(actual.Region), "us-east"; got != want {
		t.Fatalf("unexpected region: got %q, want %q", got, want)
	}
	if got, want := fi.ValueOf(task.ID), 101; got != want {
		t.Fatalf("expected task ID to be propagated after Find: got %d, want %d", got, want)
	}
}

func TestVPCFindMatchByNameAndRegion(t *testing.T) {
	client := &linode.MockLinodeClient{
		ListVPCsResponse: []linodego.VPC{
			{ID: 101, Label: "example-k8s-local", Region: "us-west"},
			{ID: 102, Label: "example-k8s-local", Region: "us-east"},
		},
	}
	cloud := &linode.MockLinodeCloud{Client_: client}
	ctx := newTestCloudupContext(t, cloud)

	task := &VPC{Name: new("example-k8s-local"), Region: new("us-east")}
	actual, err := task.Find(ctx)
	if err != nil {
		t.Fatalf("Find returned error: %v", err)
	}
	if actual == nil {
		t.Fatalf("expected to find VPC")
	}
	if got, want := fi.ValueOf(actual.ID), 102; got != want {
		t.Fatalf("unexpected VPC ID: got %d, want %d", got, want)
	}
}

func TestVPCFindListError(t *testing.T) {
	client := &linode.MockLinodeClient{ListVPCsError: errors.New("api unavailable")}
	cloud := &linode.MockLinodeCloud{Client_: client}
	ctx := newTestCloudupContext(t, cloud)

	_, err := (&VPC{Name: new("example-k8s-local")}).Find(ctx)
	if err == nil {
		t.Fatalf("expected list error")
	}
	if !strings.Contains(err.Error(), "error listing Linode (Akamai) VPCs") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVPCFindDuplicateName(t *testing.T) {
	client := &linode.MockLinodeClient{
		ListVPCsResponse: []linodego.VPC{
			{ID: 101, Label: "example-k8s-local", Region: "us-east"},
			{ID: 102, Label: "example-k8s-local", Region: "us-west"},
		},
	}
	cloud := &linode.MockLinodeCloud{Client_: client}
	ctx := newTestCloudupContext(t, cloud)

	_, err := (&VPC{Name: new("example-k8s-local")}).Find(ctx)
	if err == nil {
		t.Fatalf("expected duplicate name error")
	}
	if !strings.Contains(err.Error(), "found multiple Linode (Akamai) VPCs named") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVPCRenderLinodeCreate(t *testing.T) {
	client := &linode.MockLinodeClient{CreateVPCResponse: &linodego.VPC{ID: 42, Label: "example-k8s-local", Region: "us-east"}}
	target := linode.NewAPITarget(&linode.MockLinodeCloud{Client_: client})

	expected := &VPC{
		Name:        new("example-k8s-local"),
		Description: new("kOps cluster VPC"),
		Region:      new("us-east"),
	}

	if err := (&VPC{}).RenderLinode(target, nil, expected, nil); err != nil {
		t.Fatalf("RenderLinode returned error: %v", err)
	}
	if got, want := client.CreateVPCCalls, 1; got != want {
		t.Fatalf("unexpected create calls: got %d, want %d", got, want)
	}
	if got, want := client.LastCreateVPCOpts.Label, "example-k8s-local"; got != want {
		t.Fatalf("unexpected create label: got %q, want %q", got, want)
	}
	if got, want := client.LastCreateVPCOpts.Description, "kOps cluster VPC"; got != want {
		t.Fatalf("unexpected create description: got %q, want %q", got, want)
	}
	if got, want := client.LastCreateVPCOpts.Region, "us-east"; got != want {
		t.Fatalf("unexpected create region: got %q, want %q", got, want)
	}
	if got, want := fi.ValueOf(expected.ID), 42; got != want {
		t.Fatalf("expected task ID to be populated from create response: got %d, want %d", got, want)
	}
}

func TestVPCRenderLinodeUpdate(t *testing.T) {
	client := &linode.MockLinodeClient{UpdateVPCResponse: &linodego.VPC{ID: 42, Label: "new-name", Description: "new description"}}
	target := linode.NewAPITarget(&linode.MockLinodeCloud{Client_: client})

	actual := &VPC{ID: new(42), Name: new("old-name"), Description: new("old"), Region: new("us-east")}
	expected := &VPC{ID: new(42), Name: new("new-name"), Description: new("new description"), Region: new("us-east")}
	changes := &VPC{Name: expected.Name, Description: expected.Description}

	if err := (&VPC{}).RenderLinode(target, actual, expected, changes); err != nil {
		t.Fatalf("RenderLinode returned error: %v", err)
	}
	if got, want := client.UpdateVPCCalls, 1; got != want {
		t.Fatalf("unexpected update calls: got %d, want %d", got, want)
	}
	if got, want := client.UpdatedVPCIDs, []int{42}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected updated IDs: got %v, want %v", got, want)
	}
	if got, want := client.LastUpdateVPCOpts.Label, "new-name"; got != want {
		t.Fatalf("unexpected update label: got %q, want %q", got, want)
	}
	if got, want := client.LastUpdateVPCOpts.Description, "new description"; got != want {
		t.Fatalf("unexpected update description: got %q, want %q", got, want)
	}
}

func TestVPCCheckChangesRejectsRegionChange(t *testing.T) {
	actual := &VPC{Name: new("example-k8s-local"), ID: new(42), Region: new("us-east")}
	expected := &VPC{Name: new("example-k8s-local"), ID: new(42), Region: new("us-west")}
	changes := &VPC{Region: expected.Region}

	if err := (&VPC{}).CheckChanges(actual, expected, changes); err == nil {
		t.Fatalf("expected region change to be rejected")
	}
}
