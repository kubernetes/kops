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

package cloudinstances

import (
	"fmt"
	"testing"
)

func TestToAzureVMName(t *testing.T) {
	testCases := []struct {
		providerID string
		vmName     string
		success    bool
	}{
		{
			providerID: "azure:///subscriptions/<subscription ID>/resourceGroups/my-group/providers/Microsoft.Compute/virtualMachineScaleSets/nodes.my-cluster.k8s.local/virtualMachines/0",
			vmName:     "nodes.my-cluster.k8s.local_0",
			success:    true,
		},
		{
			providerID: "foo/bar",
			success:    false,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
			vmName, err := toAzureVMName(tc.providerID)
			if err != nil {
				if tc.success {
					t.Errorf("unexpected error %s", err)
				}
				return
			}
			if !tc.success {
				t.Fatalf("unexpected success")
			}
			if vmName != tc.vmName {
				t.Errorf("expected %s, but got %s", tc.vmName, vmName)
			}
		})
	}
}
