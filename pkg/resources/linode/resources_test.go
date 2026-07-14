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

	"github.com/linode/linodego/v2"
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
		ListVPCSubnetsResponse: []linodego.VPCSubnet{
			{ID: 901, Label: "example-k8s-local-us-east", IPv4: "172.16.1.0/16"},
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

	wantKeys := []string{"ssh-key:801", "subnet:901", "vpc:701"}
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

	if r := resourceMap["subnet:901"]; r == nil {
		t.Fatalf("missing subnet:901")
	} else if got, want := r.Name, "example-k8s-local-us-east"; got != want {
		t.Fatalf("unexpected subnet Name: got %q, want %q", got, want)
	} else if got, want := r.Type, resourceTypeSubnet; got != want {
		t.Fatalf("unexpected subnet Type: got %q, want %q", got, want)
	} else if got, want := r.Blocks, []string{"vpc:701"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected subnet Blocks: got %v, want %v", got, want)
	} else if got, ok := r.Obj.(linodego.VPCSubnet); !ok {
		t.Fatalf("unexpected subnet Obj type: %T", r.Obj)
	} else if got.Label != "example-k8s-local-us-east" {
		t.Fatalf("unexpected subnet Obj label: got %q, want %q", got.Label, "example-k8s-local-us-east")
	}

	expectedListOptions, err := linode.ListOptionsForLabel("example-k8s-local")
	if err != nil {
		t.Fatalf("ListOptionsForLabel returned error: %v", err)
	}
	if client.LastListVPCsOpts == nil {
		t.Fatalf("expected VPC list options to be recorded")
	}
	if got, want := client.LastListVPCsOpts.Filter, expectedListOptions.Filter; got != want {
		t.Fatalf("unexpected VPC list filter: got %q, want %q", got, want)
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

func TestDeleteSubnet(t *testing.T) {
	client := &linode.MockLinodeClient{}
	cloud := &linode.MockLinodeCloud{Client_: client}

	tracker := &resources.Resource{Name: "example-k8s-local-us-east", ID: "901", Type: resourceTypeSubnet, Obj: linodego.VPCSubnet{ID: 901, Label: "example-k8s-local-us-east"}}
	if err := deleteSubnet(701, cloud, tracker); err != nil {
		t.Fatalf("deleteSubnet returned error: %v", err)
	}

	if !reflect.DeepEqual(client.DeletedVPCSubnetIDs, []int{901}) {
		t.Fatalf("unexpected deleted subnet IDs: %v", client.DeletedVPCSubnetIDs)
	}
	if !reflect.DeepEqual(client.DeletedVPCSubnetVPCIDs, []int{701}) {
		t.Fatalf("unexpected deleted subnet VPC IDs: %v", client.DeletedVPCSubnetVPCIDs)
	}
}

func TestDeleteSubnet_NotFound(t *testing.T) {
	client := &linode.MockLinodeClient{DeleteVPCSubnetError: &linodego.Error{Code: 404, Message: "not found"}}
	cloud := &linode.MockLinodeCloud{Client_: client}

	tracker := &resources.Resource{Name: "example-k8s-local-us-east", ID: "901", Type: resourceTypeSubnet, Obj: linodego.VPCSubnet{ID: 901, Label: "example-k8s-local-us-east"}}
	if err := deleteSubnet(701, cloud, tracker); err != nil {
		t.Fatalf("deleteSubnet returned error for not found response: %v", err)
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
