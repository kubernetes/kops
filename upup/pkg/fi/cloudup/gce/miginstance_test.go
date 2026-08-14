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

package gce

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	compute "google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

func TestGetManagedInstanceStopsPaginationAfterMatch(t *testing.T) {
	const (
		project    = "test-project"
		zone       = "test-zone"
		mig        = "test-mig"
		instanceID = uint64(1234567890)
	)

	var pageTokens []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pageToken := r.URL.Query().Get("pageToken")
		pageTokens = append(pageTokens, pageToken)

		var response *compute.InstanceGroupManagersListManagedInstancesResponse
		switch pageToken {
		case "":
			response = &compute.InstanceGroupManagersListManagedInstancesResponse{
				ManagedInstances: []*compute.ManagedInstance{{Id: 1}},
				NextPageToken:    "second",
			}
		case "second":
			response = &compute.InstanceGroupManagersListManagedInstancesResponse{
				ManagedInstances: []*compute.ManagedInstance{{Id: instanceID}},
				NextPageToken:    "third",
			}
		default:
			http.Error(w, "unexpected page token "+pageToken, http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}))
	t.Cleanup(server.Close)

	computeService, err := compute.NewService(context.Background(), option.WithEndpoint(server.URL+"/"), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("building compute client: %v", err)
	}

	instance := &compute.Instance{
		Id:   instanceID,
		Zone: "https://www.googleapis.com/compute/v1/projects/" + project + "/zones/" + zone,
	}
	member, err := getManagedInstance(context.Background(), computeService, project, mig, instance)
	if err != nil {
		t.Fatalf("getting managed instance: %v", err)
	}
	if member.Id != instanceID {
		t.Errorf("expected instance ID %d, got %d", instanceID, member.Id)
	}

	if len(pageTokens) != 2 || pageTokens[0] != "" || pageTokens[1] != "second" {
		t.Errorf("expected requests for the first two pages, got page tokens %q", pageTokens)
	}
}
