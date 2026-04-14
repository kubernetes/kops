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

func TestVolumeFindMatch(t *testing.T) {
	client := &fakeLinodeClient{
		listVolumesResponse: []linodego.Volume{
			{ID: 101, Label: "cp-0-etcd-main-example-k8s-local", Region: "us-ord", Size: 20, Tags: []string{"kops.k8s.io/cluster:example.k8s.local", "kops.k8s.io/etcd:main", "kops.k8s.io/instance-group:control-plane-us-ord"}},
			{ID: 102, Label: "other", Region: "us-ord", Size: 20, Tags: []string{"kops.k8s.io/cluster:other.k8s.local"}},
		},
	}
	cloud := &fakeLinodeCloud{client: client}
	ctx := newTestCloudupContext(t, cloud)

	task := &Volume{
		Name:   fi.PtrTo("cp-0.etcd-main.example.k8s.local"),
		Region: fi.PtrTo("us-ord"),
		SizeGB: fi.PtrTo(int64(20)),
		Tags: []string{
			"kops.k8s.io/cluster:example.k8s.local",
			"kops.k8s.io/etcd:main",
			"kops.k8s.io/instance-group:control-plane-us-ord",
		},
	}

	actual, err := task.Find(ctx)
	if err != nil {
		t.Fatalf("Find returned error: %v", err)
	}
	if actual == nil {
		t.Fatalf("expected to find matching volume")
	}
	if got, want := fi.ValueOf(actual.ID), 101; got != want {
		t.Fatalf("unexpected volume ID: got %d, want %d", got, want)
	}
	if got, want := fi.ValueOf(actual.Region), "us-ord"; got != want {
		t.Fatalf("unexpected region: got %q, want %q", got, want)
	}
	if got, want := fi.ValueOf(actual.SizeGB), int64(20); got != want {
		t.Fatalf("unexpected size: got %d, want %d", got, want)
	}
	if got, want := fi.ValueOf(actual.Name), fi.ValueOf(task.Name); got != want {
		t.Fatalf("expected task identity name to be preserved: got %q, want %q", got, want)
	}
	if got, want := fi.ValueOf(task.ID), 101; got != want {
		t.Fatalf("expected task ID to be propagated after Find: got %d, want %d", got, want)
	}
}

func TestVolumeFindListError(t *testing.T) {
	client := &fakeLinodeClient{listVolumesError: errors.New("api unavailable")}
	cloud := &fakeLinodeCloud{client: client}
	ctx := newTestCloudupContext(t, cloud)

	task := &Volume{Name: fi.PtrTo("cp-0.etcd-main.example.k8s.local")}
	_, err := task.Find(ctx)
	if err == nil {
		t.Fatalf("expected list error")
	}
	if !strings.Contains(err.Error(), "error listing Linode (Akamai) volumes") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVolumeRenderLinodeCreate(t *testing.T) {
	client := &fakeLinodeClient{}
	target := linode.NewAPITarget(&fakeLinodeCloud{client: client})

	expected := &Volume{
		Name:   fi.PtrTo("cp-0.etcd-main.example.k8s.local"),
		Region: fi.PtrTo("us-ord"),
		SizeGB: fi.PtrTo(int64(20)),
		Tags: []string{
			"kops.k8s.io/cluster:example.k8s.local",
			"kops.k8s.io/etcd:main",
			"kops.k8s.io/instance-group:control-plane-us-ord",
		},
	}

	if err := (&Volume{}).RenderLinode(target, nil, expected, nil); err != nil {
		t.Fatalf("RenderLinode returned error: %v", err)
	}

	if got, want := client.createVolumeCalls, 1; got != want {
		t.Fatalf("unexpected create calls: got %d, want %d", got, want)
	}
	if got, want := client.lastCreateVolumeOpts.Region, "us-ord"; got != want {
		t.Fatalf("unexpected region: got %q, want %q", got, want)
	}
	if got, want := client.lastCreateVolumeOpts.Size, 20; got != want {
		t.Fatalf("unexpected size: got %d, want %d", got, want)
	}
	if got, want := client.lastCreateVolumeOpts.Label, "cp-0-etcd-main-example-k8s-local"; got != want {
		t.Fatalf("unexpected label: got %q, want %q", got, want)
	}
}

func TestNormalizedVolumeLabel(t *testing.T) {
	longName := "Etcd MAIN volume for Control Plane 0 in very-long-cluster-name.with.many.parts.and.characters.example.k8s.local"
	label := normalizedVolumeLabel(longName)

	if len(label) > maxLinodeVolumeLabelLength {
		t.Fatalf("label too long: %d", len(label))
	}
	if label == "" {
		t.Fatalf("label should not be empty")
	}
	if strings.Contains(label, " ") {
		t.Fatalf("label should not contain spaces: %q", label)
	}
	if strings.Contains(label, ".") {
		t.Fatalf("label should not contain dots: %q", label)
	}

	regression := normalizedVolumeLabel("d.etcd-events.test-linode.k8s.local")
	if got, want := regression, "d-etcd-events-test-linode-k8s-lo"; got != want {
		t.Fatalf("unexpected normalized label for regression case: got %q, want %q", got, want)
	}
}
