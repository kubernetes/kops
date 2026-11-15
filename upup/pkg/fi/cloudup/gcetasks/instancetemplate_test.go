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

package gcetasks

import (
	"cmp"
	"testing"
)

func TestGuessMachineTypeInfo(t *testing.T) {
	tests := []struct {
		machineType       string
		supportsMigration bool
	}{
		{machineType: "", supportsMigration: true},
		{machineType: "n4-standard-4", supportsMigration: true},
		{machineType: "c4a-standard-4", supportsMigration: true},
		{machineType: "e2-medium", supportsMigration: true},
		// GPU machine families cannot be live migrated.
		{machineType: "a3-highgpu-8g", supportsMigration: false},
		{machineType: "a4x-highgpu-4g", supportsMigration: false},
		{machineType: "g2-standard-4", supportsMigration: false},
		// Bare metal machine types cannot be live migrated.
		{machineType: "c4a-highmem-96-metal", supportsMigration: false},
		{machineType: "c4a-standard-96-metal", supportsMigration: false},
		{machineType: "c3-standard-192-metal", supportsMigration: false},
		{machineType: "c4-highmem-288-lssd-metal", supportsMigration: false},
		{machineType: "z3-highmem-192-highlssd-metal", supportsMigration: false},
	}

	for _, tc := range tests {
		t.Run(cmp.Or(tc.machineType, "empty"), func(t *testing.T) {
			info, err := guessMachineTypeInfo(tc.machineType)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if info.SupportsMigration != tc.supportsMigration {
				t.Errorf("SupportsMigration = %v, want %v", info.SupportsMigration, tc.supportsMigration)
			}

			wantOnHostMaintenance := "MIGRATE"
			if !tc.supportsMigration {
				wantOnHostMaintenance = "TERMINATE"
			}
			scheduling := buildScheduling(info, new(false), nil, nil)
			if scheduling.OnHostMaintenance != wantOnHostMaintenance {
				t.Errorf("OnHostMaintenance = %q, want %q", scheduling.OnHostMaintenance, wantOnHostMaintenance)
			}
		})
	}
}
