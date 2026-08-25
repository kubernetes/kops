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
	"reflect"
	"strings"
	"testing"

	"github.com/linode/linodego/v2"
	"k8s.io/kops/pkg/truncate"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/linode"
)

func TestVolumeFindByLabel(t *testing.T) {
	client := &linode.MockLinodeClient{
		ListVolumesResponse: []linodego.Volume{{
			ID:     101,
			Label:  "example-k8s-local-etcd-main",
			Region: "us-east",
			Size:   20,
			Tags:   []string{"kops.k8s.io/cluster:example.k8s.local"},
		}},
	}
	ctx := newTestCloudupContext(t, &linode.MockLinodeCloud{Client_: client})
	task := &Volume{Name: new("example-k8s-local-etcd-main")}

	actual, err := task.Find(ctx)
	if err != nil {
		t.Fatalf("Find returned error: %v", err)
	}
	if actual == nil {
		t.Fatalf("expected to find volume")
	}
	if got, want := fi.ValueOf(actual.ID), 101; got != want {
		t.Fatalf("unexpected volume ID: got %d, want %d", got, want)
	}
	if got, want := fi.ValueOf(task.ID), 101; got != want {
		t.Fatalf("expected task ID to be propagated after Find: got %d, want %d", got, want)
	}
	if got, want := fi.ValueOf(actual.Region), "us-east"; got != want {
		t.Fatalf("unexpected volume region: got %q, want %q", got, want)
	}
	if got, want := fi.ValueOf(actual.SizeGB), 20; got != want {
		t.Fatalf("unexpected volume size: got %d, want %d", got, want)
	}

	expectedListOptions, err := linode.ListOptionsForLabel("example-k8s-local-etcd-main")
	if err != nil {
		t.Fatalf("ListOptionsForLabel returned error: %v", err)
	}
	if client.LastListVolumesOpts == nil {
		t.Fatalf("expected volume list options to be recorded")
	}
	if got, want := client.LastListVolumesOpts.Filter, expectedListOptions.Filter; got != want {
		t.Fatalf("unexpected volume list filter: got %q, want %q", got, want)
	}
}

func TestVolumeFindListError(t *testing.T) {
	client := &linode.MockLinodeClient{ListVolumesError: errors.New("api unavailable")}
	ctx := newTestCloudupContext(t, &linode.MockLinodeCloud{Client_: client})

	_, err := (&Volume{Name: new("example-k8s-local-etcd-main")}).Find(ctx)
	if err == nil {
		t.Fatalf("expected list error")
	}
	if !strings.Contains(err.Error(), "error listing Akamai (Linode) volumes") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVolumeRenderLinodeCreate(t *testing.T) {
	client := &linode.MockLinodeClient{CreateVolumeResponse: &linodego.Volume{ID: 42, Label: "example-k8s-local-etcd-main"}}
	target := linode.NewAPITarget(&linode.MockLinodeCloud{Client_: client})
	expected := &Volume{
		Name:   new("example-k8s-local-etcd-main"),
		Region: new("us-east"),
		SizeGB: new(20),
		Tags:   []string{"kops.k8s.io/cluster:example.k8s.local"},
	}

	if err := (&Volume{}).RenderLinode(target, nil, expected, nil); err != nil {
		t.Fatalf("RenderLinode returned error: %v", err)
	}
	if got, want := client.CreateVolumeCalls, 1; got != want {
		t.Fatalf("unexpected create calls: got %d, want %d", got, want)
	}
	if got, want := client.LastCreateVolumeOpts.Label, "example-k8s-local-etcd-main"; got != want {
		t.Fatalf("unexpected create label: got %q, want %q", got, want)
	}
	if got, want := client.LastCreateVolumeOpts.Region, "us-east"; got != want {
		t.Fatalf("unexpected create region: got %q, want %q", got, want)
	}
	if got, want := client.LastCreateVolumeOpts.Size, 20; got != want {
		t.Fatalf("unexpected create size: got %d, want %d", got, want)
	}
	if got, want := client.LastCreateVolumeOpts.Tags, expected.Tags; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected create tags: got %v, want %v", got, want)
	}
	if got, want := fi.ValueOf(expected.ID), 42; got != want {
		t.Fatalf("expected task ID to be populated from create response: got %d, want %d", got, want)
	}
}

func TestVolumeRenderLinodeCreateNormalizesLabel(t *testing.T) {
	client := &linode.MockLinodeClient{CreateVolumeResponse: &linodego.Volume{ID: 42}}
	target := linode.NewAPITarget(&linode.MockLinodeCloud{Client_: client})
	expected := &Volume{
		Name:   new("d.etcd-main.kops.linode.example.com"),
		Region: new("us-east"),
		SizeGB: new(20),
	}

	if err := (&Volume{}).RenderLinode(target, nil, expected, nil); err != nil {
		t.Fatalf("RenderLinode returned error: %v", err)
	}
	label := client.LastCreateVolumeOpts.Label
	if len(label) > 32 {
		t.Fatalf("expected label to fit Linode's length limit: %q", label)
	}
	if strings.ContainsAny(label, ".:") {
		t.Fatalf("expected label to contain only Linode-supported characters: %q", label)
	}
	want := truncate.TruncateString(linode.NormalizeLinodeLabel(fi.ValueOf(expected.Name)), truncate.TruncateStringOptions{MaxLength: 32})
	if label != want {
		t.Fatalf("unexpected normalized label: %q", label)
	}
}

func TestVolumeFindNormalizesLabel(t *testing.T) {
	name := "d.etcd-main.kops.linode.example.com"
	label := truncate.TruncateString(linode.NormalizeLinodeLabel(name), truncate.TruncateStringOptions{MaxLength: 32})
	client := &linode.MockLinodeClient{ListVolumesResponse: []linodego.Volume{{
		ID:     101,
		Label:  label,
		Region: "us-east",
		Size:   20,
	}}}
	ctx := newTestCloudupContext(t, &linode.MockLinodeCloud{Client_: client})
	task := &Volume{Name: new(name)}

	actual, err := task.Find(ctx)
	if err != nil {
		t.Fatalf("Find returned error: %v", err)
	}
	if actual == nil {
		t.Fatalf("expected to find volume")
	}
	if got, want := fi.ValueOf(actual.Name), name; got != want {
		t.Fatalf("unexpected canonical name: got %q, want %q", got, want)
	}

	expectedListOptions, err := linode.ListOptionsForLabel(label)
	if err != nil {
		t.Fatalf("ListOptionsForLabel returned error: %v", err)
	}
	if got, want := client.LastListVolumesOpts.Filter, expectedListOptions.Filter; got != want {
		t.Fatalf("unexpected volume list filter: got %q, want %q", got, want)
	}
}

func TestVolumeRenderLinodeResize(t *testing.T) {
	client := &linode.MockLinodeClient{}
	target := linode.NewAPITarget(&linode.MockLinodeCloud{Client_: client})
	actual := &Volume{ID: new(42), Name: new("example-k8s-local-etcd-main"), SizeGB: new(20)}
	expected := &Volume{SizeGB: new(30)}
	changes := &Volume{SizeGB: expected.SizeGB}

	if err := (&Volume{}).RenderLinode(target, actual, expected, changes); err != nil {
		t.Fatalf("RenderLinode returned error: %v", err)
	}
	if got, want := client.ResizeVolumeCalls, 1; got != want {
		t.Fatalf("unexpected resize calls: got %d, want %d", got, want)
	}
	if got, want := client.ResizedVolumeIDs, []int{42}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected resized volume IDs: got %v, want %v", got, want)
	}
	if got, want := client.LastResizeVolumeOpts.Size, 30; got != want {
		t.Fatalf("unexpected resize size: got %d, want %d", got, want)
	}
	if got, want := fi.ValueOf(expected.ID), 42; got != want {
		t.Fatalf("expected task ID to stay populated after resize: got %d, want %d", got, want)
	}
}

func TestVolumeRenderLinodeResizeError(t *testing.T) {
	client := &linode.MockLinodeClient{ResizeVolumeError: errors.New("resize API down")}
	target := linode.NewAPITarget(&linode.MockLinodeCloud{Client_: client})
	actual := &Volume{ID: new(42), Name: new("example-k8s-local-etcd-main"), SizeGB: new(20)}
	expected := &Volume{SizeGB: new(30)}
	changes := &Volume{SizeGB: expected.SizeGB}

	err := (&Volume{}).RenderLinode(target, actual, expected, changes)
	if err == nil {
		t.Fatalf("expected resize error")
	}
	if !strings.Contains(err.Error(), "resize API down") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVolumeCheckChangesRejectsUnsupportedChanges(t *testing.T) {
	actual := &Volume{
		ID:     new(42),
		Name:   new("example-k8s-local-etcd-main"),
		Region: new("us-east"),
		SizeGB: new(20),
		Tags:   []string{"kops.k8s.io/cluster:example.k8s.local"},
	}

	for _, testCase := range []struct {
		name     string
		expected *Volume
		changes  *Volume
	}{
		{
			name:     "tags",
			expected: &Volume{Tags: []string{"kops.k8s.io/cluster:other.k8s.local"}},
			changes:  &Volume{Tags: []string{"kops.k8s.io/cluster:other.k8s.local"}},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			if err := (&Volume{}).CheckChanges(actual, testCase.expected, testCase.changes); err == nil {
				t.Fatalf("expected %s change to be rejected", testCase.name)
			}
		})
	}
}

func TestVolumeCheckChangesRejectsSizeDecrease(t *testing.T) {
	actual := &Volume{SizeGB: new(20)}
	expected := &Volume{SizeGB: new(10)}
	changes := &Volume{SizeGB: expected.SizeGB}

	err := (&Volume{}).CheckChanges(actual, expected, changes)
	if err == nil {
		t.Fatalf("expected size decrease to be rejected")
	}
	if !strings.Contains(err.Error(), "SizeGB cannot be decreased") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVolumeCheckChangesValidatesMinimumSize(t *testing.T) {
	for _, testCase := range []struct {
		name      string
		sizeGB    int
		wantError bool
	}{
		{name: "below minimum", sizeGB: 9, wantError: true},
		{name: "minimum", sizeGB: 10},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			expected := &Volume{
				Name:   new("example-k8s-local-etcd-main"),
				Region: new("us-east"),
				SizeGB: new(testCase.sizeGB),
			}

			err := (&Volume{}).CheckChanges(nil, expected, expected)
			if testCase.wantError {
				if err == nil {
					t.Fatalf("expected size %d to be rejected", testCase.sizeGB)
				}
				if !strings.Contains(err.Error(), "SizeGB must be at least 10 GB") {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected size %d to be accepted, got error: %v", testCase.sizeGB, err)
			}
		})
	}
}
