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
	"fmt"
	"testing"
)

func TestGetVMSSNameFromProviderID(t *testing.T) {
	testCases := []struct {
		providerID string
		vmName     string
		success    bool
	}{
		{
			providerID: "azure:///subscriptions/7e232992-6f42-4554-a685-3e02278c3a8b/resourceGroups/cnatix-test/providers/Microsoft.Compute/virtualMachineScaleSets/master-eastus-1.masters.test-cluster.k8s.local/virtualMachines/0",
			vmName:     "master-eastus-1.masters.test-cluster.k8s.local_0",
			success:    true,
		},
		{
			providerID: "azure:///subscriptions/7e232992-6f42-4554-a685-3e02278c3a8b/resourceGroups/cnatix-test/providers/Microsoft.Compute/virtualMachines/general-purpose-ckptq",
			vmName:     "general-purpose-ckptq",
			success:    true,
		},
		{
			providerID: "azure:///subscriptions/7e232992-6f42-4554-a685-3e02278c3a8b/resourceGroups/cnatix-test/providers/Microsoft.Compute/virtualMachineScaleSets/master-eastus-1.masters.test-cluster.k8s.local",
			success:    false,
		},
		{
			providerID: "azure:///subscriptions/7e232992-6f42-4554-a685-3e02278c3a8b/resourceGroups/cnatix-test/providers/Microsoft.Compute/virtualMachines",
			success:    false,
		},
		{
			providerID: "aws:///instanceID",
			success:    false,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
			vmName, err := getVMNameFromProviderID(tc.providerID)
			if err != nil {
				if tc.success {
					t.Errorf("unexpected error: %s", err)
				}
				return
			}
			if !tc.success {
				t.Errorf("unexpected success")
			}
			if tc.vmName != vmName {
				t.Errorf("expected %s, but got %s", tc.vmName, vmName)
			}
		})
	}
}
