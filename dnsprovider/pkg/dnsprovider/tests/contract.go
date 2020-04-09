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

package tests

import (
	"testing"

	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider/rrstype"
)

// TestContract verifies the general ResourceRecordChangeset contract
func TestContract(t *testing.T, rrsets dnsprovider.ResourceRecordSets) {

	{
		changeset := rrsets.StartChangeset()
		if !changeset.IsEmpty() {
			t.Fatalf("expected new changeset to be empty")
		}

		rrs := changeset.ResourceRecordSets().New("foo", []string{"192.168.0.1"}, 1, rrstype.A)
		changeset.Add(rrs)
		if changeset.IsEmpty() {
			t.Fatalf("expected changeset not to be empty after add")
		}
	}

	{
		changeset := rrsets.StartChangeset()
		if !changeset.IsEmpty() {
			t.Fatalf("expected new changeset to be empty")
		}

		rrs := changeset.ResourceRecordSets().New("foo", []string{"192.168.0.1"}, 1, rrstype.A)
		changeset.Remove(rrs)
		if changeset.IsEmpty() {
			t.Fatalf("expected changeset not to be empty after remove")
		}
	}

	{
		changeset := rrsets.StartChangeset()
		if !changeset.IsEmpty() {
			t.Fatalf("expected new changeset to be empty")
		}

		rrs := changeset.ResourceRecordSets().New("foo", []string{"192.168.0.1"}, 1, rrstype.A)
		changeset.Upsert(rrs)
		if changeset.IsEmpty() {
			t.Fatalf("expected changeset not to be empty after upsert")
		}
	}

}
