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
	"io"
	"strings"
	"testing"

	"github.com/linode/linodego/v2"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/linode"
	"k8s.io/kops/util/pkg/vfs"
)

func TestSubnetFindMatchByName(t *testing.T) {
	client := &linode.MockLinodeClient{
		ListVPCSubnetsResponse: []linodego.VPCSubnet{
			{ID: 101, Label: "example-k8s-local-us-east", IPv4: "172.16.1.0/16"},
			{ID: 102, Label: "other", IPv4: "172.16.2.0/16"},
		},
	}
	cloud := &linode.MockLinodeCloud{Client_: client}
	ctx := newTestCloudupContext(t, cloud)

	vpc := &VPC{ID: new(42)}
	task := &Subnet{Name: new("example-k8s-local-us-east"), VPC: vpc}
	actual, err := task.Find(ctx)
	if err != nil {
		t.Fatalf("Find returned error: %v", err)
	}
	if actual == nil {
		t.Fatalf("expected to find subnet")
	}
	if got, want := fi.ValueOf(actual.ID), 101; got != want {
		t.Fatalf("unexpected subnet ID: got %d, want %d", got, want)
	}
	if got, want := fi.ValueOf(actual.IPv4), "172.16.1.0/16"; got != want {
		t.Fatalf("unexpected subnet IPv4: got %q, want %q", got, want)
	}
	if actual.VPC != vpc {
		t.Fatalf("expected matched subnet to preserve VPC reference")
	}
	if got, want := fi.ValueOf(task.ID), 101; got != want {
		t.Fatalf("expected task ID to be propagated after Find: got %d, want %d", got, want)
	}
	if got, want := fi.ValueOf(actual.Name), "example-k8s-local-us-east"; got != want {
		t.Fatalf("unexpected subnet name: got %q, want %q", got, want)
	}
	if got, want := client.LastListVPCSubnetsVPCID, 42; got != want {
		t.Fatalf("unexpected listed VPC ID: got %d, want %d", got, want)
	}
	if client.LastListVPCSubnetsOpts != nil {
		t.Fatalf("expected subnet lookup to list all subnets in the VPC")
	}
}

func TestSubnetFindIgnoresMatchingSubnetInOtherVPC(t *testing.T) {
	client := &linode.MockLinodeClient{
		ListVPCSubnetsResponses: map[int][]linodego.VPCSubnet{
			42: {
				{ID: 101, Label: "example-k8s-local-us-east", IPv4: "172.16.1.0/16"},
			},
			99: {
				{ID: 202, Label: "example-k8s-local-us-east", IPv4: "172.16.2.0/16"},
			},
		},
	}
	cloud := &linode.MockLinodeCloud{Client_: client}
	ctx := newTestCloudupContext(t, cloud)

	vpc := &VPC{ID: new(42)}
	task := &Subnet{Name: new("example-k8s-local-us-east"), VPC: vpc}
	actual, err := task.Find(ctx)
	if err != nil {
		t.Fatalf("Find returned error: %v", err)
	}
	if actual == nil {
		t.Fatalf("expected to find subnet in requested VPC")
	}
	if got, want := fi.ValueOf(actual.ID), 101; got != want {
		t.Fatalf("unexpected subnet ID: got %d, want %d", got, want)
	}
	if got, want := fi.ValueOf(actual.IPv4), "172.16.1.0/16"; got != want {
		t.Fatalf("unexpected subnet IPv4: got %q, want %q", got, want)
	}
	if got, want := client.LastListVPCSubnetsVPCID, 42; got != want {
		t.Fatalf("unexpected listed VPC ID: got %d, want %d", got, want)
	}
}

func TestSubnetFindRequiresVPCID(t *testing.T) {
	cloud := &linode.MockLinodeCloud{Client_: &linode.MockLinodeClient{}}
	ctx := newTestCloudupContext(t, cloud)

	_, err := (&Subnet{Name: new("example-k8s-local-us-east"), VPC: &VPC{}}).Find(ctx)
	if err == nil {
		t.Fatalf("expected VPC ID error")
	}
	if !strings.Contains(err.Error(), "Subnet.VPC.ID is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSubnetRunDryRunBeforeVPCHasID(t *testing.T) {
	cloud := &linode.MockLinodeCloud{Client_: &linode.MockLinodeClient{}}
	vpc := &VPC{Name: new("example-k8s-local"), Lifecycle: fi.LifecycleSync, Region: new("us-ord")}
	subnet := &Subnet{Name: new("example-k8s-local-subnet-a"), Lifecycle: fi.LifecycleSync, IPv4: new("172.16.1.0/24"), VPC: vpc}

	tasks := map[string]fi.CloudupTask{}
	modelContext := &fi.CloudupModelBuilderContext{Tasks: tasks}
	modelContext.AddTask(vpc)
	modelContext.AddTask(subnet)

	assetBuilder := assets.NewAssetBuilder(vfs.Context, nil, false)
	target := fi.NewCloudupDryRunTarget(assetBuilder, true, io.Discard)
	ctx, err := fi.NewCloudupContext(context.Background(), fi.DeletionProcessingModeDeleteIncludingDeferred, target, nil, cloud, nil, nil, nil, modelContext.Tasks)
	if err != nil {
		t.Fatalf("NewCloudupContext returned error: %v", err)
	}

	if err := ctx.RunTasks(fi.RunTasksOptions{}); err != nil {
		t.Fatalf("RunTasks returned error: %v", err)
	}
	if got, want := cloud.Client().(*linode.MockLinodeClient).ListVPCSubnetsCalls, 0; got != want {
		t.Fatalf("unexpected subnet list calls during dry-run: got %d, want %d", got, want)
	}
}

func TestSubnetFindRequiresVPC(t *testing.T) {
	cloud := &linode.MockLinodeCloud{Client_: &linode.MockLinodeClient{}}
	ctx := newTestCloudupContext(t, cloud)

	_, err := (&Subnet{Name: new("example-k8s-local-us-east")}).Find(ctx)
	if err == nil {
		t.Fatalf("expected VPC error")
	}
	if !strings.Contains(err.Error(), "Subnet.VPC is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSubnetFindMatchesByIPv4WhenLabelChanges(t *testing.T) {
	client := &linode.MockLinodeClient{
		ListVPCSubnetsResponse: []linodego.VPCSubnet{
			{ID: 101, Label: "old-subnet", IPv4: "172.16.1.0/16"},
		},
	}
	cloud := &linode.MockLinodeCloud{Client_: client}
	ctx := newTestCloudupContext(t, cloud)

	task := &Subnet{Name: new("new-subnet"), IPv4: new("172.16.1.0/16"), VPC: &VPC{ID: new(42)}}
	actual, err := task.Find(ctx)
	if err != nil {
		t.Fatalf("Find returned error: %v", err)
	}
	if actual == nil {
		t.Fatalf("expected to find subnet by IPv4")
	}
	if got, want := fi.ValueOf(actual.ID), 101; got != want {
		t.Fatalf("unexpected subnet ID: got %d, want %d", got, want)
	}
	if got, want := fi.ValueOf(actual.Name), "old-subnet"; got != want {
		t.Fatalf("unexpected subnet name: got %q, want %q", got, want)
	}
	if got, want := client.ListVPCSubnetsCalls, 1; got != want {
		t.Fatalf("unexpected subnet list calls: got %d, want %d", got, want)
	}
	if client.LastListVPCSubnetsOpts != nil {
		t.Fatalf("expected subnet lookup to list all subnets in the VPC")
	}
}

func TestSubnetFindDuplicateName(t *testing.T) {
	client := &linode.MockLinodeClient{
		ListVPCSubnetsResponse: []linodego.VPCSubnet{
			{ID: 101, Label: "example-k8s-local-us-east"},
			{ID: 102, Label: "example-k8s-local-us-east"},
		},
	}
	cloud := &linode.MockLinodeCloud{Client_: client}
	ctx := newTestCloudupContext(t, cloud)

	_, err := (&Subnet{Name: new("example-k8s-local-us-east"), VPC: &VPC{ID: new(42)}}).Find(ctx)
	if err == nil {
		t.Fatalf("expected duplicate name error")
	}
	if !strings.Contains(err.Error(), "found multiple Linode (Akamai) VPC Subnets named") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSubnetFindListError(t *testing.T) {
	client := &linode.MockLinodeClient{ListVPCSubnetsError: errors.New("api unavailable")}
	cloud := &linode.MockLinodeCloud{Client_: client}
	ctx := newTestCloudupContext(t, cloud)

	_, err := (&Subnet{Name: new("example-k8s-local-us-east"), VPC: &VPC{ID: new(42)}}).Find(ctx)
	if err == nil {
		t.Fatalf("expected list error")
	}
	if !strings.Contains(err.Error(), "error listing Linode (Akamai) VPC Subnets") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSubnetRenderLinodeCreate(t *testing.T) {
	client := &linode.MockLinodeClient{CreateVPCSubnetResponse: &linodego.VPCSubnet{ID: 42, Label: "example-k8s-local-us-east", IPv4: "172.16.1.0/16"}}
	target := linode.NewAPITarget(&linode.MockLinodeCloud{Client_: client})

	expected := &Subnet{
		Name: new("example-k8s-local-us-east"),
		IPv4: new("172.16.1.0/16"),
		VPC:  &VPC{ID: new(7)},
	}

	if err := (&Subnet{}).RenderLinode(target, nil, expected, nil); err != nil {
		t.Fatalf("RenderLinode returned error: %v", err)
	}
	if got, want := client.CreateVPCSubnetCalls, 1; got != want {
		t.Fatalf("unexpected create calls: got %d, want %d", got, want)
	}
	if got, want := client.LastCreateVPCSubnetOpts.Label, "example-k8s-local-us-east"; got != want {
		t.Fatalf("unexpected create label: got %q, want %q", got, want)
	}
	if got, want := client.LastCreateVPCSubnetOpts.IPv4, "172.16.1.0/16"; got != want {
		t.Fatalf("unexpected create IPv4: got %q, want %q", got, want)
	}
	if got, want := client.LastCreateVPCSubnetVPCID, 7; got != want {
		t.Fatalf("unexpected create VPC ID: got %d, want %d", got, want)
	}
	if got, want := fi.ValueOf(expected.ID), 42; got != want {
		t.Fatalf("expected task ID to be populated from create response: got %d, want %d", got, want)
	}
}

func TestSubnetRenderLinodeUpdate(t *testing.T) {
	client := &linode.MockLinodeClient{UpdateVPCSubnetResponse: &linodego.VPCSubnet{ID: 42, Label: "renamed-subnet", IPv4: "172.16.1.0/16"}}
	target := linode.NewAPITarget(&linode.MockLinodeCloud{Client_: client})

	actual := &Subnet{ID: new(42), Name: new("old-subnet"), IPv4: new("172.16.1.0/16"), VPC: &VPC{ID: new(7)}}
	expected := &Subnet{ID: new(42), Name: new("renamed-subnet"), IPv4: new("172.16.1.0/16"), VPC: &VPC{ID: new(7)}}
	changes := &Subnet{Name: expected.Name}

	if err := (&Subnet{}).RenderLinode(target, actual, expected, changes); err != nil {
		t.Fatalf("RenderLinode returned error: %v", err)
	}
	if got, want := client.UpdateVPCSubnetCalls, 1; got != want {
		t.Fatalf("unexpected update calls: got %d, want %d", got, want)
	}
	if got, want := client.LastUpdateVPCSubnetVPCID, 7; got != want {
		t.Fatalf("unexpected update VPC ID: got %d, want %d", got, want)
	}
	if got, want := client.LastUpdateVPCSubnetID, 42; got != want {
		t.Fatalf("unexpected update subnet ID: got %d, want %d", got, want)
	}
	if got, want := client.LastUpdateVPCSubnetOpts.Label, "renamed-subnet"; got != want {
		t.Fatalf("unexpected update label: got %q, want %q", got, want)
	}
	if got, want := fi.ValueOf(expected.ID), 42; got != want {
		t.Fatalf("expected task ID to stay populated after update: got %d, want %d", got, want)
	}
}

func TestSubnetCheckChangesRejectsIPv4Change(t *testing.T) {
	actual := &Subnet{Name: new("example-k8s-local-us-east"), ID: new(42), IPv4: new("172.16.1.0/16"), VPC: &VPC{ID: new(7)}}
	expected := &Subnet{Name: new("example-k8s-local-us-east"), ID: new(42), IPv4: new("172.16.2.0/16"), VPC: &VPC{ID: new(7)}}
	changes := &Subnet{IPv4: expected.IPv4}

	if err := (&Subnet{}).CheckChanges(actual, expected, changes); err == nil {
		t.Fatalf("expected IPv4 change to be rejected")
	}
}

func TestSubnetCheckChangesRejectsVPCChange(t *testing.T) {
	actual := &Subnet{Name: new("example-k8s-local-us-east"), ID: new(42), IPv4: new("172.16.1.0/16"), VPC: &VPC{ID: new(7)}}
	expected := &Subnet{Name: new("example-k8s-local-us-east"), ID: new(42), IPv4: new("172.16.1.0/16"), VPC: &VPC{ID: new(8)}}
	changes := &Subnet{VPC: expected.VPC}

	if err := (&Subnet{}).CheckChanges(actual, expected, changes); err == nil {
		t.Fatalf("expected VPC change to be rejected")
	}
}
