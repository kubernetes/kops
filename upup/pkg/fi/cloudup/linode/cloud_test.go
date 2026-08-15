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

package linode

import (
	"reflect"
	"strings"
	"testing"

	"github.com/linode/linodego/v2"
	"k8s.io/kops/pkg/cloudinstances"
)

func TestListOptionsForTags(t *testing.T) {
	tests := []struct {
		name string
		tags []string
		want string
	}{
		{
			name: "single tag",
			tags: []string{"kops.k8s.io/cluster:example.k8s.local"},
			want: `{"+and":[{"tags":{"+contains":"kops.k8s.io/cluster:example.k8s.local"}}]}`,
		},
		{
			name: "multiple tags",
			tags: []string{"kops.k8s.io/cluster:example.k8s.local", "kops.k8s.io/instance-group:control-plane"},
			want: `{"+and":[{"tags":{"+contains":"kops.k8s.io/cluster:example.k8s.local"}},{"tags":{"+contains":"kops.k8s.io/instance-group:control-plane"}}]}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			options, err := ListOptionsForTags(test.tags...)
			if err != nil {
				t.Fatalf("ListOptionsForTags returned error: %v", err)
			}
			if got := options.Filter; got != test.want {
				t.Fatalf("unexpected filter: got %q, want %q", got, test.want)
			}
		})
	}
}

func TestCloudDeleteInstance(t *testing.T) {
	for _, testCase := range []struct {
		name        string
		id          string
		deleteError error
		wantError   string
		wantDeleted []int
	}{
		{name: "success", id: "101", wantDeleted: []int{101}},
		{name: "invalid id", id: "not-an-int", wantError: "error parsing Akamai (Linode) instance ID"},
		{name: "not found", id: "101", deleteError: &linodego.Error{Code: 404, Message: "not found"}, wantDeleted: []int{101}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			client := &MockLinodeClient{DeleteInstanceError: testCase.deleteError}
			cloud := &Cloud{client: client}

			err := cloud.DeleteInstance(&cloudinstances.CloudInstance{ID: testCase.id})
			if testCase.wantError != "" {
				if err == nil {
					t.Fatalf("expected error for instance ID %q", testCase.id)
				}
				if !strings.Contains(err.Error(), testCase.wantError) {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("DeleteInstance returned error: %v", err)
			}
			if !reflect.DeepEqual(client.DeletedInstanceIDs, testCase.wantDeleted) {
				t.Fatalf("unexpected deleted instance IDs: got %v, want %v", client.DeletedInstanceIDs, testCase.wantDeleted)
			}
		})
	}
}
