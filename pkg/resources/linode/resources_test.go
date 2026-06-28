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
	"errors"
	"reflect"
	"sort"
	"testing"

	"github.com/linode/linodego"
	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/upup/pkg/fi/cloudup/linode"
)

func TestListResources(t *testing.T) {
	client := &linode.MockLinodeClient{
		ListVPCsResponse: []linodego.VPC{
			{ID: 701, Label: "example-k8s-local", Region: "us-east"},
			{ID: 702, Label: "example-k8s-local", Region: "us-west"},
			{ID: 703, Label: "other-k8s-local", Region: "us-east"},
		},
		ListSSHKeysResponse: []linodego.SSHKey{
			{ID: 801, Label: "kubernetes-example-k8s-local-aa-bb"},
			{ID: 802, Label: "custom-key"},
		},
	}
	cloud := &linode.MockLinodeCloud{Client_: client, Region_: "us-east"}

	resourceMap, err := ListResources(cloud, resources.ClusterInfo{Name: "example.k8s.local"})
	if err != nil {
		t.Fatalf("ListResources returned error: %v", err)
	}

	wantKeys := []string{"ssh-key:801", "vpc:701"}
	if gotKeys := sortedResourceKeys(resourceMap); !reflect.DeepEqual(gotKeys, wantKeys) {
		t.Fatalf("unexpected resources\nwant: %v\n got: %v", wantKeys, gotKeys)
	}

	if r := resourceMap["vpc:701"]; r == nil {
		t.Fatalf("missing vpc:701")
	} else if got, want := r.Name, "example-k8s-local"; got != want {
		t.Fatalf("unexpected VPC Name: got %q, want %q", got, want)
	} else if got, want := r.Type, resourceTypeVPC; got != want {
		t.Fatalf("unexpected VPC Type: got %q, want %q", got, want)
	}

	if r := resourceMap["ssh-key:801"]; r == nil {
		t.Fatalf("missing ssh-key:801")
	} else if got, want := r.Name, "kubernetes-example-k8s-local-aa-bb"; got != want {
		t.Fatalf("unexpected SSH key Name: got %q, want %q", got, want)
	} else if got, want := r.Type, resourceTypeSSHKey; got != want {
		t.Fatalf("unexpected SSH key Type: got %q, want %q", got, want)
	}
}

func TestListResources_PropagatesErrors(t *testing.T) {
	client := &linode.MockLinodeClient{ListVPCsError: errors.New("VPC API down")}
	cloud := &linode.MockLinodeCloud{Client_: client, Region_: "us-east"}

	_, err := ListResources(cloud, resources.ClusterInfo{Name: "example.k8s.local"})
	if err == nil {
		t.Fatalf("expected error when listing VPCs")
	}
}

func TestDeleteVPC(t *testing.T) {
	client := &linode.MockLinodeClient{}
	cloud := &linode.MockLinodeCloud{Client_: client}

	tracker := &resources.Resource{Name: "example-k8s-local", ID: "701", Type: resourceTypeVPC}
	if err := deleteVPC(cloud, tracker); err != nil {
		t.Fatalf("deleteVPC returned error: %v", err)
	}

	if !reflect.DeepEqual(client.DeletedVPCIDs, []int{701}) {
		t.Fatalf("unexpected deleted VPC IDs: %v", client.DeletedVPCIDs)
	}
}

func TestDeleteVPC_NotFound(t *testing.T) {
	client := &linode.MockLinodeClient{DeleteVPCError: &linodego.Error{Code: 404, Message: "not found"}}
	cloud := &linode.MockLinodeCloud{Client_: client}

	tracker := &resources.Resource{Name: "example-k8s-local", ID: "701", Type: resourceTypeVPC}
	if err := deleteVPC(cloud, tracker); err != nil {
		t.Fatalf("deleteVPC returned error for not found response: %v", err)
	}
}

func TestDeleteSSHKey(t *testing.T) {
	client := &linode.MockLinodeClient{}
	cloud := &linode.MockLinodeCloud{Client_: client}

	tracker := &resources.Resource{Name: "kubernetes-example-k8s-local-aa-bb", ID: "801", Type: resourceTypeSSHKey}
	if err := deleteSSHKey(cloud, tracker); err != nil {
		t.Fatalf("deleteSSHKey returned error: %v", err)
	}

	if !reflect.DeepEqual(client.DeletedSSHKeyIDs, []int{801}) {
		t.Fatalf("unexpected deleted SSH key IDs: %v", client.DeletedSSHKeyIDs)
	}
}

func TestDeleteSSHKey_NotFound(t *testing.T) {
	client := &linode.MockLinodeClient{DeleteSSHKeyError: &linodego.Error{Code: 404, Message: "not found"}}
	cloud := &linode.MockLinodeCloud{Client_: client}

	tracker := &resources.Resource{Name: "kubernetes-example-k8s-local-aa-bb", ID: "801", Type: resourceTypeSSHKey}
	if err := deleteSSHKey(cloud, tracker); err != nil {
		t.Fatalf("deleteSSHKey returned error for not found response: %v", err)
	}
}

func sortedResourceKeys(m map[string]*resources.Resource) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
