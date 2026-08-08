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
	"encoding/base64"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/linode/linodego/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/linode"
)

func TestInstanceFindDoesNotMutateDesiredTask(t *testing.T) {
	cluster := &kops.Cluster{ObjectMeta: metav1.ObjectMeta{Name: "example.k8s.local"}}
	userDataText := "#!/bin/bash\necho hello\n"
	userData := fi.Resource(fi.NewStringResource(userDataText))
	baseTags := []string{
		fmt.Sprintf("%s:%s", linode.TagKubernetesClusterName, cluster.Name),
		fmt.Sprintf("%s:%s", linode.TagKubernetesInstanceGroup, "nodes-us-east"),
		fmt.Sprintf("%s:%s", linode.TagKubernetesInstanceRole, string(kops.InstanceGroupRoleNode)),
	}
	userDataTag := fmt.Sprintf("%s:%s", linode.TagKubernetesInstanceUserData, generateUserDataHash(userDataText))

	client := &linode.MockLinodeClient{
		ListInstancesResponse: []linodego.Instance{{
			ID:     101,
			Label:  "nodes-us-east-1",
			Type:   "g6-standard-2",
			Image:  "linode/ubuntu24.04",
			Region: "us-east",
			Tags:   append(append([]string{}, baseTags...), userDataTag),
		}},
		ListInterfacesResponses: map[int][]linodego.LinodeInterface{
			101: {{Public: &linodego.PublicInterface{}, VPC: &linodego.VPCInterface{SubnetID: 42}}},
		},
	}
	cloud := &linode.MockLinodeCloud{Client_: client}
	ctx := newTestCloudupContext(t, cloud)
	ctx.T.Cluster = cluster
	requirePublicInterface := true

	task := &Instance{
		Name:                   new("nodes-us-east"),
		Region:                 "us-east",
		Type:                   "g6-standard-2",
		Image:                  "linode/ubuntu24.04",
		Count:                  3,
		Subnet:                 &Subnet{ID: new(42)},
		RequirePublicInterface: &requirePublicInterface,
		Tags:                   append([]string{}, baseTags...),
		UserData:               userData,
	}
	originalTags := append([]string{}, task.Tags...)

	actual, err := task.Find(ctx)
	if err != nil {
		t.Fatalf("Find returned error: %v", err)
	}
	if actual == nil {
		t.Fatalf("expected to find instances")
	}
	if got, want := actual.Count, 1; got != want {
		t.Fatalf("unexpected actual count: got %d, want %d", got, want)
	}
	if got, want := task.Count, 3; got != want {
		t.Fatalf("expected desired count to remain unchanged: got %d, want %d", got, want)
	}
	if task.NeedsUpdate != nil {
		t.Fatalf("expected desired task NeedsUpdate to remain nil, got %v", task.NeedsUpdate)
	}
	if !reflect.DeepEqual(task.Tags, originalTags) {
		t.Fatalf("expected desired task tags to remain unchanged: got %v, want %v", task.Tags, originalTags)
	}
	if len(actual.NeedsUpdate) != 0 {
		t.Fatalf("unexpected actual needsUpdate: %v", actual.NeedsUpdate)
	}
	if !reflect.DeepEqual(actual.Tags, originalTags) {
		t.Fatalf("expected actual tags to preserve desired base tags: got %v, want %v", actual.Tags, originalTags)
	}
}

func TestInstanceFindMarksMissingExpectedTagForUpdate(t *testing.T) {
	cluster := &kops.Cluster{ObjectMeta: metav1.ObjectMeta{Name: "example.k8s.local"}}
	userDataText := "#!/bin/bash\necho hello\n"
	userData := fi.Resource(fi.NewStringResource(userDataText))
	baseTags := []string{
		fmt.Sprintf("%s:%s", linode.TagKubernetesClusterName, cluster.Name),
		fmt.Sprintf("%s:%s", linode.TagKubernetesInstanceGroup, "nodes-us-east"),
		fmt.Sprintf("%s:%s", linode.TagKubernetesInstanceRole, string(kops.InstanceGroupRoleNode)),
	}
	userDataTag := fmt.Sprintf("%s:%s", linode.TagKubernetesInstanceUserData, generateUserDataHash(userDataText))

	client := &linode.MockLinodeClient{
		ListInstancesResponse: []linodego.Instance{{
			ID:     101,
			Label:  "nodes-us-east-1",
			Type:   "g6-standard-2",
			Image:  "linode/ubuntu24.04",
			Region: "us-east",
			Tags: []string{
				fmt.Sprintf("%s:%s", linode.TagKubernetesClusterName, cluster.Name),
				fmt.Sprintf("%s:%s", linode.TagKubernetesInstanceGroup, "nodes-us-east"),
				userDataTag,
			},
		}},
		ListInterfacesResponses: map[int][]linodego.LinodeInterface{
			101: {{Public: &linodego.PublicInterface{}, VPC: &linodego.VPCInterface{SubnetID: 42}}},
		},
	}
	cloud := &linode.MockLinodeCloud{Client_: client}
	ctx := newTestCloudupContext(t, cloud)
	ctx.T.Cluster = cluster
	requirePublicInterface := true

	task := &Instance{
		Name:                   new("nodes-us-east"),
		Region:                 "us-east",
		Type:                   "g6-standard-2",
		Image:                  "linode/ubuntu24.04",
		Subnet:                 &Subnet{ID: new(42)},
		RequirePublicInterface: &requirePublicInterface,
		Tags:                   append([]string{}, baseTags...),
		UserData:               userData,
	}

	actual, err := task.Find(ctx)
	if err != nil {
		t.Fatalf("Find returned error: %v", err)
	}
	if actual == nil {
		t.Fatalf("expected to find instances")
	}
	if got, want := actual.NeedsUpdate, []string{"nodes-us-east-1"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected actual needsUpdate: got %v, want %v", got, want)
	}
	if task.NeedsUpdate != nil {
		t.Fatalf("expected desired task NeedsUpdate to remain nil, got %v", task.NeedsUpdate)
	}
}

func TestInstanceFindMarksInterfaceMismatchForUpdate(t *testing.T) {
	cluster := &kops.Cluster{ObjectMeta: metav1.ObjectMeta{Name: "example.k8s.local"}}
	userDataText := "#!/bin/bash\necho hello\n"
	userData := fi.Resource(fi.NewStringResource(userDataText))
	baseTags := []string{
		fmt.Sprintf("%s:%s", linode.TagKubernetesClusterName, cluster.Name),
		fmt.Sprintf("%s:%s", linode.TagKubernetesInstanceGroup, "nodes-us-east"),
		fmt.Sprintf("%s:%s", linode.TagKubernetesInstanceRole, string(kops.InstanceGroupRoleNode)),
	}
	userDataTag := fmt.Sprintf("%s:%s", linode.TagKubernetesInstanceUserData, generateUserDataHash(userDataText))
	client := &linode.MockLinodeClient{
		ListInstancesResponse: []linodego.Instance{{
			ID:     101,
			Label:  "nodes-us-east-1",
			Type:   "g6-standard-2",
			Image:  "linode/ubuntu24.04",
			Region: "us-east",
			Tags:   append(append([]string{}, baseTags...), userDataTag),
		}},
		ListInterfacesResponses: map[int][]linodego.LinodeInterface{
			101: {{VPC: &linodego.VPCInterface{SubnetID: 99}}},
		},
	}
	cloud := &linode.MockLinodeCloud{Client_: client}
	ctx := newTestCloudupContext(t, cloud)
	ctx.T.Cluster = cluster
	requirePublicInterface := true

	task := &Instance{
		Name:                   new("nodes-us-east"),
		Region:                 "us-east",
		Type:                   "g6-standard-2",
		Image:                  "linode/ubuntu24.04",
		Subnet:                 &Subnet{ID: new(42)},
		RequirePublicInterface: &requirePublicInterface,
		Tags:                   append([]string{}, baseTags...),
		UserData:               userData,
	}

	actual, err := task.Find(ctx)
	if err != nil {
		t.Fatalf("Find returned error: %v", err)
	}
	if actual == nil {
		t.Fatalf("expected to find instances")
	}
	if got, want := actual.NeedsUpdate, []string{"nodes-us-east-1"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected actual needsUpdate: got %v, want %v", got, want)
	}
}

func TestInstanceCheckChangesCountValidation(t *testing.T) {
	for _, testCase := range []struct {
		name      string
		count     int
		wantError string
	}{
		{name: "zero allowed", count: 0},
		{name: "negative rejected", count: -1, wantError: "Count must be positive or 0"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			requirePublicInterface := false
			expected := &Instance{
				Name:                   new("nodes-us-east"),
				Region:                 "us-east",
				Type:                   "g6-standard-2",
				Image:                  "linode/ubuntu24.04",
				Count:                  testCase.count,
				Subnet:                 &Subnet{Name: new("example-k8s-local-subnet-a")},
				RequirePublicInterface: &requirePublicInterface,
				Tags:                   []string{"kops.k8s.io/cluster:example.k8s.local"},
				UserData:               fi.Resource(fi.NewStringResource("#!/bin/bash\n")),
				AuthorizedKeys: []*SSHKey{{
					Name: new("example-k8s-local-default"),
				}},
			}

			err := expected.CheckChanges(nil, expected, nil)
			if testCase.wantError == "" {
				if err != nil {
					t.Fatalf("expected count %d to be allowed, got error: %v", testCase.count, err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected count %d to be rejected", testCase.count)
			}
			if !strings.Contains(err.Error(), testCase.wantError) {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestInstanceRenderLinodeCreateWithPublicAndVPCInterfaces(t *testing.T) {
	requirePublicInterface := true
	publicKey := fi.Resource(fi.NewStringResource(testLinodeSSHPublicKey))
	userDataText := "#!/bin/bash\necho hello\n"
	userData := fi.Resource(fi.NewStringResource(userDataText))
	client := &linode.MockLinodeClient{
		CreateInstanceResponse: &linodego.Instance{ID: 101, Label: "nodes-us-east-1"},
	}
	target := linode.NewAPITarget(&linode.MockLinodeCloud{Client_: client})

	expected := &Instance{
		Name:                   new("nodes-us-east"),
		Region:                 "us-east",
		Type:                   "g6-standard-2",
		Image:                  "linode/ubuntu24.04",
		Count:                  1,
		Subnet:                 &Subnet{ID: new(42)},
		RequirePublicInterface: &requirePublicInterface,
		Tags:                   []string{"kops.k8s.io/cluster:example.k8s.local"},
		AuthorizedKeys:         []*SSHKey{{Name: new("example-k8s-local-default"), PublicKey: &publicKey}},
		UserData:               userData,
	}

	if err := (&Instance{}).RenderLinode(target, nil, expected, nil); err != nil {
		t.Fatalf("RenderLinode returned error: %v", err)
	}
	if got, want := client.CreateInstanceCalls, 1; got != want {
		t.Fatalf("unexpected create calls: got %d, want %d", got, want)
	}
	if got, want := client.LastCreateInstanceOpts.Region, "us-east"; got != want {
		t.Fatalf("unexpected create region: got %q, want %q", got, want)
	}
	if got, want := client.LastCreateInstanceOpts.Type, "g6-standard-2"; got != want {
		t.Fatalf("unexpected create type: got %q, want %q", got, want)
	}
	if got, want := client.LastCreateInstanceOpts.Image, "linode/ubuntu24.04"; got != want {
		t.Fatalf("unexpected create image: got %q, want %q", got, want)
	}
	if got, want := client.LastCreateInstanceOpts.InterfaceGeneration, linodego.GenerationLinode; got != want {
		t.Fatalf("unexpected interface generation: got %q, want %q", got, want)
	}
	if got, want := client.LastCreateInstanceOpts.AuthorizedKeys, []string{testLinodeSSHPublicKey}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected authorized keys: got %v, want %v", got, want)
	}
	if got, want := client.LastCreateInstanceOpts.Metadata.UserData, base64.StdEncoding.EncodeToString([]byte(userDataText)); got != want {
		t.Fatalf("unexpected metadata user_data: got %q, want %q", got, want)
	}
	if got, want := client.LastCreateInstanceOpts.Tags, []string{
		"kops.k8s.io/cluster:example.k8s.local",
		fmt.Sprintf("%s:%s", linode.TagKubernetesInstanceUserData, generateUserDataHash(userDataText)),
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected tags: got %v, want %v", got, want)
	}
	if got, want := len(client.LastCreateInstanceOpts.LinodeInterfaces), 2; got != want {
		t.Fatalf("unexpected interface count: got %d, want %d", got, want)
	}
	if client.LastCreateInstanceOpts.LinodeInterfaces[0].Public == nil || client.LastCreateInstanceOpts.LinodeInterfaces[1].VPC == nil {
		t.Fatalf("expected one public and one VPC interface")
	}
	if got, want := client.LastCreateInstanceOpts.LinodeInterfaces[1].VPC.SubnetID, 42; got != want {
		t.Fatalf("unexpected VPC subnet ID: got %d, want %d", got, want)
	}
	if !strings.HasPrefix(client.LastCreateInstanceOpts.Label, "nodes-us-east-") {
		t.Fatalf("unexpected created label: got %q", client.LastCreateInstanceOpts.Label)
	}
}

func TestInstanceRenderLinodeCreateWithVPCOnlyInterfaceAndExistingSSHKey(t *testing.T) {
	requirePublicInterface := false
	userDataText := "#!/bin/bash\necho hello\n"
	userData := fi.Resource(fi.NewStringResource(userDataText))
	client := &linode.MockLinodeClient{
		ListSSHKeysResponse:    []linodego.SSHKey{{ID: 12, Label: "example-k8s-local-default", SSHKey: testLinodeSSHPublicKey}},
		CreateInstanceResponse: &linodego.Instance{ID: 101, Label: "nodes-us-east-1"},
	}
	target := linode.NewAPITarget(&linode.MockLinodeCloud{Client_: client})

	expected := &Instance{
		Name:                   new("nodes-us-east"),
		Region:                 "us-east",
		Type:                   "g6-standard-2",
		Image:                  "linode/ubuntu24.04",
		Count:                  1,
		Subnet:                 &Subnet{ID: new(42)},
		RequirePublicInterface: &requirePublicInterface,
		Tags:                   []string{"kops.k8s.io/cluster:example.k8s.local"},
		AuthorizedKeys:         []*SSHKey{{Name: new("example-k8s-local-default")}},
		UserData:               userData,
	}

	if err := (&Instance{}).RenderLinode(target, nil, expected, nil); err != nil {
		t.Fatalf("RenderLinode returned error: %v", err)
	}
	if got, want := client.ListSSHKeysCalls, 1; got != want {
		t.Fatalf("unexpected SSH key list calls: got %d, want %d", got, want)
	}
	if got, want := len(client.LastCreateInstanceOpts.LinodeInterfaces), 1; got != want {
		t.Fatalf("unexpected interface count: got %d, want %d", got, want)
	}
	if client.LastCreateInstanceOpts.LinodeInterfaces[0].Public != nil {
		t.Fatalf("expected no public interface")
	}
	if got, want := client.LastCreateInstanceOpts.LinodeInterfaces[0].VPC.SubnetID, 42; got != want {
		t.Fatalf("unexpected VPC subnet ID: got %d, want %d", got, want)
	}
	if got, want := client.LastCreateInstanceOpts.AuthorizedKeys, []string{testLinodeSSHPublicKey}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected authorized keys: got %v, want %v", got, want)
	}
}

func TestBuildLinodeInstanceLabel(t *testing.T) {
	first, err := buildLinodeInstanceLabel("nodes-us-east")
	if err != nil {
		t.Fatalf("buildLinodeInstanceLabel returned error: %v", err)
	}
	second, err := buildLinodeInstanceLabel("nodes-us-east")
	if err != nil {
		t.Fatalf("buildLinodeInstanceLabel returned error: %v", err)
	}
	if first == second {
		t.Fatalf("expected distinct labels, got %q and %q", first, second)
	}
	if !strings.HasPrefix(first, "nodes-us-east-") {
		t.Fatalf("unexpected first label prefix: %q", first)
	}
	if len(first) > 64 {
		t.Fatalf("expected first label to fit Linode length limit, got %d characters", len(first))
	}

	trimmed, err := buildLinodeInstanceLabel(strings.Repeat("a", 80) + "-")
	if err != nil {
		t.Fatalf("buildLinodeInstanceLabel returned error: %v", err)
	}
	if len(trimmed) > 64 {
		t.Fatalf("expected trimmed label to fit Linode length limit, got %d characters", len(trimmed))
	}
}
