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
	"testing"

	"github.com/linode/linodego"
)

func TestNewProvider(t *testing.T) {
	client := linodego.NewClient(nil)
	provider := NewProvider(&client)
	if provider == nil {
		t.Fatal("NewProvider returned nil")
	}

	zones, ok := provider.Zones()
	if !ok {
		t.Fatal("Provider should support zones")
	}
	if zones == nil {
		t.Fatal("Zones should not be nil")
	}
}

func TestZoneNew(t *testing.T) {
	client := linodego.NewClient(nil)
	z := &zones{client: &client}

	zone, err := z.New("example.com")
	if err != nil {
		t.Fatalf("Failed to create new zone: %v", err)
	}
	if zone == nil {
		t.Fatal("New zone should not be nil")
	}
	if zone.Name() != "example.com" {
		t.Errorf("Expected zone name 'example.com', got '%s'", zone.Name())
	}
	if zone.ID() != "example.com" {
		t.Errorf("Expected zone ID 'example.com', got '%s'", zone.ID())
	}
}

func TestResourceRecordSets(t *testing.T) {
	client := linodego.NewClient(nil)
	z := &zone{
		name:   "example.com",
		id:     "example.com",
		client: &client,
	}

	rrsets, ok := z.ResourceRecordSets()
	if !ok {
		t.Fatal("Zone should support resource record sets")
	}
	if rrsets == nil {
		t.Fatal("ResourceRecordSets should not be nil")
	}

	// Test creating a new record set
	rrs := rrsets.New("test.example.com", []string{"192.0.2.1"}, 300, "A")
	if rrs == nil {
		t.Fatal("New ResourceRecordSet should not be nil")
	}
	if rrs.Name() != "test.example.com" {
		t.Errorf("Expected name 'test.example.com', got '%s'", rrs.Name())
	}
	if rrs.Ttl() != 300 {
		t.Errorf("Expected TTL 300, got %d", rrs.Ttl())
	}
	if string(rrs.Type()) != "A" {
		t.Errorf("Expected type 'A', got '%s'", rrs.Type())
	}

	// Test empty data returns nil
	nilRrs := rrsets.New("test.example.com", []string{}, 300, "A")
	if nilRrs != nil {
		t.Error("ResourceRecordSet with empty data should return nil")
	}
}

func TestChangeset(t *testing.T) {
	client := linodego.NewClient(nil)
	z := &zone{
		name:   "example.com",
		id:     "example.com",
		client: &client,
	}
	rrsets := &resourceRecordSets{
		zone:   z,
		client: &client,
	}

	changeset := rrsets.StartChangeset()
	if changeset == nil {
		t.Fatal("StartChangeset should not return nil")
	}

	if !changeset.IsEmpty() {
		t.Error("New changeset should be empty")
	}

	// Test adding records
	rrs := rrsets.New("test.example.com", []string{"192.0.2.1"}, 300, "A")
	changeset.Add(rrs)
	if changeset.IsEmpty() {
		t.Error("Changeset with additions should not be empty")
	}

	// Test removals
	changeset2 := rrsets.StartChangeset()
	changeset2.Remove(rrs)
	if changeset2.IsEmpty() {
		t.Error("Changeset with removals should not be empty")
	}

	// Test upserts
	changeset3 := rrsets.StartChangeset()
	changeset3.Upsert(rrs)
	if changeset3.IsEmpty() {
		t.Error("Changeset with upserts should not be empty")
	}
}
