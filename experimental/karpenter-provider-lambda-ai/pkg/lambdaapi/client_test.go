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

package lambdaapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func setupTestServer(handler http.HandlerFunc) (*httptest.Server, *Client) {
	server := httptest.NewServer(handler)
	client := NewClient("test-api-key")
	client.baseURL = server.URL // Override baseURL to point to the test server
	return server, client
}

func TestListInstanceTypes(t *testing.T) {
	expectedTypes := map[string]InstanceType{
		"gpu_1x_a100": {
			Name:              "gpu_1x_a100",
			Description:       "1x A100 (40 GB SXM4)",
			PriceCentsPerHour: 110,
			Specs: Specs{
				VCPUs:        12,
				MemoryGib:    200,
				GPUs:         1,
				GPUMemoryGib: 40,
			},
		},
	}

	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/instance-types" {
			t.Errorf("expected path /instance-types, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-api-key" {
			t.Errorf("expected Authorization header Bearer test-api-key, got %s", r.Header.Get("Authorization"))
		}

		resp := ListInstanceTypesResponse{Data: expectedTypes}
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	types, err := client.ListInstanceTypes()
	if err != nil {
		t.Fatalf("ListInstanceTypes failed: %v", err)
	}

	if !reflect.DeepEqual(types, expectedTypes) {
		t.Errorf("expected %v, got %v", expectedTypes, types)
	}
}

func TestListRegions(t *testing.T) {
	expectedRegions := []Region{
		{Name: "us-east-1", Description: "US East (N. Virginia)"},
	}

	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/regions" {
			t.Errorf("expected path /regions, got %s", r.URL.Path)
		}

		resp := ListRegionsResponse{Data: expectedRegions}
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	regions, err := client.ListRegions()
	if err != nil {
		t.Fatalf("ListRegions failed: %v", err)
	}

	if !reflect.DeepEqual(regions, expectedRegions) {
		t.Errorf("expected %v, got %v", expectedRegions, regions)
	}
}

func TestLaunchInstance(t *testing.T) {
	req := LaunchRequest{
		RegionName:       "us-east-1",
		InstanceTypeName: "gpu_1x_a100",
		SSHKeyNames:      []string{"my-key"},
		Quantity:         1,
		Name:             "test-instance",
	}
	expectedIDs := []string{"instance-123"}

	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/instance-operations/launch" {
			t.Errorf("expected path /instance-operations/launch, got %s", r.URL.Path)
		}

		var receivedReq LaunchRequest
		if err := json.NewDecoder(r.Body).Decode(&receivedReq); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if !reflect.DeepEqual(receivedReq, req) {
			t.Errorf("expected request %v, got %v", req, receivedReq)
		}

		resp := LaunchResponse{Data: struct {
			InstanceIDs []string `json:"instance_ids"`
		}{InstanceIDs: expectedIDs}}
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	ids, err := client.LaunchInstance(req)
	if err != nil {
		t.Fatalf("LaunchInstance failed: %v", err)
	}

	if !reflect.DeepEqual(ids, expectedIDs) {
		t.Errorf("expected %v, got %v", expectedIDs, ids)
	}
}

func TestTerminateInstances(t *testing.T) {
	instanceIDs := []string{"instance-123", "instance-456"}

	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/instance-operations/terminate" {
			t.Errorf("expected path /instance-operations/terminate, got %s", r.URL.Path)
		}

		var receivedReq TerminateRequest
		if err := json.NewDecoder(r.Body).Decode(&receivedReq); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if !reflect.DeepEqual(receivedReq.InstanceIDs, instanceIDs) {
			t.Errorf("expected instance IDs %v, got %v", instanceIDs, receivedReq.InstanceIDs)
		}

		resp := TerminateResponse{Data: struct {
			TerminatedInstances []Instance `json:"terminated_instances"`
		}{TerminatedInstances: []Instance{{ID: "instance-123"}, {ID: "instance-456"}}}}
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	err := client.TerminateInstances(instanceIDs)
	if err != nil {
		t.Fatalf("TerminateInstances failed: %v", err)
	}
}

func TestListInstances(t *testing.T) {
	expectedInstances := []Instance{
		{ID: "instance-123", Name: "test-instance", Status: "active"},
	}

	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/instances" {
			t.Errorf("expected path /instances, got %s", r.URL.Path)
		}

		resp := ListInstancesResponse{Data: expectedInstances}
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	instances, err := client.ListInstances()
	if err != nil {
		t.Fatalf("ListInstances failed: %v", err)
	}

	if !reflect.DeepEqual(instances, expectedInstances) {
		t.Errorf("expected %v, got %v", expectedInstances, instances)
	}
}

func TestListSSHKeys(t *testing.T) {
	expectedKeys := []SSHKey{
		{Name: "my-key", PublicKey: "ssh-rsa AAA..."},
	}

	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/ssh-keys" {
			t.Errorf("expected path /ssh-keys, got %s", r.URL.Path)
		}

		resp := ListSSHKeysResponse{Data: expectedKeys}
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	keys, err := client.ListSSHKeys()
	if err != nil {
		t.Fatalf("ListSSHKeys failed: %v", err)
	}

	if !reflect.DeepEqual(keys, expectedKeys) {
		t.Errorf("expected %v, got %v", expectedKeys, keys)
	}
}

func TestAPIError(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": {"message": "Bad Request"}}`))
	})
	defer server.Close()

	_, err := client.ListRegions()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	expectedErrorSubstring := "API request failed with status 400"
	if err.Error()[:len(expectedErrorSubstring)] != expectedErrorSubstring {
		t.Errorf("expected error to start with %q, got %q", expectedErrorSubstring, err.Error())
	}
}
