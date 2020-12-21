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
	"context"
	"fmt"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2020-06-01/compute"
	v1 "k8s.io/api/core/v1"
)

func TestGetVMSSNameFromProviderID(t *testing.T) {
	testCases := []struct {
		providerID string
		vmssName   string
		success    bool
	}{
		{
			providerID: "azure:///subscriptions/7e232992-6f42-4554-a685-3e02278c3a8b/resourceGroups/cnatix-test/providers/Microsoft.Compute/virtualMachineScaleSets/master-eastus-1.masters.test-cluster.k8s.local/virtualMachines/0",
			vmssName:   "master-eastus-1.masters.test-cluster.k8s.local",
			success:    true,
		},
		{
			providerID: "azure:///subscriptions/7e232992-6f42-4554-a685-3e02278c3a8b/resourceGroups/cnatix-test/providers/Microsoft.Compute/virtualMachineScaleSets/master-eastus-1.masters.test-cluster.k8s.local",
			success:    false,
		},
		{
			providerID: "aws:///instanceID",
			success:    false,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
			vmssName, err := getVMSSNameFromProviderID(tc.providerID)
			if err != nil {
				if tc.success {
					t.Errorf("unexpected error: %s", err)
				}
				return
			}
			if !tc.success {
				t.Errorf("unexpected success")
			}
			if tc.vmssName != vmssName {
				t.Errorf("expected %s, but got %s", tc.vmssName, vmssName)
			}
		})
	}
}

func TestIdentifyNode(t *testing.T) {
	vmssName := "master-eastus-1.masters.test-cluster.k8s.local"
	igName := "master-eastus-1"

	client := &mockClient{
		vmsses: map[string]compute.VirtualMachineScaleSet{
			vmssName: {
				Name: &vmssName,
				Tags: map[string]*string{
					InstanceGroupNameTag: &igName,
				},
			},
		},
	}
	identifier := &nodeIdentifier{
		vmssGetter: client,
	}

	const providerIDPattern = "azure:///subscriptions/7e232992-6f42-4554-a685-3e02278c3a8b/resourceGroups/cnatix-test/providers/Microsoft.Compute/virtualMachineScaleSets/%s/virtualMachines/0"
	testCases := []struct {
		providerID string
		vmssName   string
		success    bool
	}{
		{
			providerID: fmt.Sprintf(providerIDPattern, vmssName),
			vmssName:   vmssName,
			success:    true,
		},
		{
			providerID: fmt.Sprintf(providerIDPattern, "differentName"),
			success:    false,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
			node := &v1.Node{
				Spec: v1.NodeSpec{
					ProviderID: tc.providerID,
				},
			}
			info, err := identifier.IdentifyNode(context.TODO(), node)
			if err != nil {
				if tc.success {
					t.Errorf("unexpected error: %s", err)
				}
				return
			}
			if !tc.success {
				t.Errorf("unexpected success")
			}
			if info.InstanceID != tc.vmssName {
				t.Errorf("expected %s, but got %s", tc.vmssName, info.InstanceID)
			}
		})
	}

}
