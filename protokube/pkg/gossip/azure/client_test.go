/*
Copyright 2020 The Kubernetes Authors.

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

package azure

import (
	"os"
	"reflect"
	"testing"
)

func TestUnmarshalMetadata(t *testing.T) {
	data, err := os.ReadFile("testdata/metadata.json")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	metadata, err := unmarshalInstanceMetadata(data)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	mc := metadata.Compute
	if a, e := mc.ResourceGroupName, "macikgo-test-may-23"; a != e {
		t.Errorf("expected resource group name %s, but got %s", e, a)
	}
	if a, e := mc.SubscriptionID, "8d10da13-8125-4ba9-a717-bf7490507b3d"; a != e {
		t.Errorf("expected resource group name %s, but got %s", e, a)
	}
	actualTags, err := mc.GetTags()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	expectedTags := map[string]string{
		"baz": "bash",
		"foo": "bar",
	}
	if a, e := actualTags, expectedTags; !reflect.DeepEqual(a, e) {
		t.Errorf("expected resource group name %s, but got %s", e, a)
	}

	mn := metadata.Network
	if a, e := len(mn.Interfaces), 1; a != e {
		t.Errorf("expected %d interfaces, but got %d", e, a)
	}
	ipAddrs := mn.Interfaces[0].IPv4.IPAddresses
	if a, e := len(ipAddrs), 1; a != e {
		t.Errorf("expected %d IP addresses, but got %d", e, a)
	}
	if a, e := ipAddrs[0].PrivateIPAddress, "172.16.32.8"; a != e {
		t.Errorf("expected private IP address %s, but got %s", e, a)
	}
	if a, e := ipAddrs[0].PublicIPAddress, "52.136.124.5"; a != e {
		t.Errorf("expected public IP address %s, but got %s", e, a)
	}
}
