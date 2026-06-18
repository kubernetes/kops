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

package kops

import (
	"slices"
	"testing"
	// "github.com/stretchr/testify/assert"
)

func TestInstanceGroupRole_SubRoles(t *testing.T) {
	tests := []struct {
		name     string
		role     InstanceGroupRole
		expected []InstanceGroupSubRole
	}{
		{
			name:     "empty",
			role:     "",
			expected: []InstanceGroupSubRole{"unknown"},
		},
		{
			name:     "unknown",
			role:     "unknown",
			expected: []InstanceGroupSubRole{"unknown"},
		},
		{
			name:     "Control-plane",
			role:     InstanceGroupSubRoleControlPlane.Role(),
			expected: []InstanceGroupSubRole{canocicalControlPlane},
		},
		{
			name:     "APIServer",
			role:     InstanceGroupSubRoleAPIServer.Role(),
			expected: []InstanceGroupSubRole{canocicalAPIServer},
		},
		{
			name:     "Etcd",
			role:     InstanceGroupSubRoleEtcd.Role(),
			expected: []InstanceGroupSubRole{canocicalEtcd},
		},
		{
			name:     "Node",
			role:     InstanceGroupSubRoleNode.Role(),
			expected: []InstanceGroupSubRole{canocicalNode},
		},
		{
			name:     "Bastion",
			role:     InstanceGroupSubRoleBastion.Role(),
			expected: []InstanceGroupSubRole{canocicalBastion},
		},
		{
			name: "All control plane",
			role: "APIServer,Etcd,Scheduler,CloudControllerManager,KubeControllerManager",
			expected: []InstanceGroupSubRole{
				canocicalAPIServer,
				canocicalCloudControllerManager,
				canocicalEtcd,
				canocicalKubeControllerManager,
				canocicalScheduler,
			},
		},
		{
			name: "All control plane - reverse order",
			role: "KubeControllerManager,CloudControllerManager,Scheduler,Etcd,APIServer",
			expected: []InstanceGroupSubRole{
				canocicalAPIServer,
				canocicalCloudControllerManager,
				canocicalEtcd,
				canocicalKubeControllerManager,
				canocicalScheduler,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.role.SubRoles()
			if result := slices.Compare(actual, tc.expected); result != 0 {
				t.Errorf("%q.SubRoles() = %v, want %v, result %d", tc.role, actual, tc.expected, result)
			}
		})
	}
}

func TestInstanceGroupRole_HasX(t *testing.T) {
	tests := []struct {
		name            string
		role            InstanceGroupRole
		hasControlPlane bool
		hasAPIServer    bool
		hasEtcd         bool
		hasScheduler    bool
		hasKCM          bool
		hasCCM          bool
		hasNode         bool
		hasBastion      bool
	}{
		{
			name:            "empty",
			role:            "",
			hasControlPlane: false,
			hasAPIServer:    false,
			hasEtcd:         false,
			hasScheduler:    false,
			hasKCM:          false,
			hasCCM:          false,
			hasNode:         false,
			hasBastion:      false,
		},
		{
			name:            "unknown",
			role:            "unknown",
			hasControlPlane: false,
			hasAPIServer:    false,
			hasEtcd:         false,
			hasScheduler:    false,
			hasKCM:          false,
			hasCCM:          false,
			hasNode:         false,
			hasBastion:      false,
		},
		{
			name:            "Control-plane",
			role:            InstanceGroupSubRoleControlPlane.Role(),
			hasControlPlane: true,
			hasAPIServer:    false,
			hasEtcd:         false,
			hasScheduler:    false,
			hasKCM:          false,
			hasCCM:          false,
			hasNode:         false,
			hasBastion:      false,
		},
		{
			name:            "APIServer",
			role:            InstanceGroupSubRoleAPIServer.Role(),
			hasControlPlane: false,
			hasAPIServer:    true,
			hasEtcd:         false,
			hasScheduler:    false,
			hasKCM:          false,
			hasCCM:          false,
			hasNode:         false,
			hasBastion:      false,
		},
		{
			name:            "Etcd",
			role:            InstanceGroupSubRoleEtcd.Role(),
			hasControlPlane: false,
			hasAPIServer:    false,
			hasEtcd:         true,
			hasScheduler:    false,
			hasKCM:          false,
			hasCCM:          false,
			hasNode:         false,
			hasBastion:      false,
		},
		{
			name:            "Scheduler",
			role:            InstanceGroupSubRoleScheduler.Role(),
			hasControlPlane: false,
			hasAPIServer:    false,
			hasEtcd:         false,
			hasScheduler:    true,
			hasKCM:          false,
			hasCCM:          false,
			hasNode:         false,
			hasBastion:      false,
		},
		{
			name:            "KCM",
			role:            InstanceGroupSubRoleKubeControllerManager.Role(),
			hasControlPlane: false,
			hasAPIServer:    false,
			hasEtcd:         false,
			hasScheduler:    false,
			hasKCM:          true,
			hasCCM:          false,
			hasNode:         false,
			hasBastion:      false,
		},
		{
			name:            "CCM",
			role:            InstanceGroupSubRoleCloudControllerManager.Role(),
			hasControlPlane: false,
			hasAPIServer:    false,
			hasEtcd:         false,
			hasScheduler:    false,
			hasKCM:          false,
			hasCCM:          true,
			hasNode:         false,
			hasBastion:      false,
		},
		{
			name:            "Node",
			role:            InstanceGroupSubRoleNode.Role(),
			hasControlPlane: false,
			hasAPIServer:    false,
			hasEtcd:         false,
			hasScheduler:    false,
			hasKCM:          false,
			hasCCM:          false,
			hasNode:         true,
			hasBastion:      false,
		},
		{
			name:            "Bastion",
			role:            InstanceGroupSubRoleBastion.Role(),
			hasControlPlane: false,
			hasAPIServer:    false,
			hasEtcd:         false,
			hasScheduler:    false,
			hasKCM:          false,
			hasCCM:          false,
			hasNode:         false,
			hasBastion:      true,
		},
		{
			name:            "All control plane elements",
			role:            "APIServer,Etcd,Scheduler,CloudControllerManager,KubeControllerManager",
			hasControlPlane: false,
			hasAPIServer:    true,
			hasEtcd:         true,
			hasScheduler:    true,
			hasKCM:          true,
			hasCCM:          true,
			hasNode:         false,
			hasBastion:      false,
		},
		{
			name:            "No control plane",
			role:            "Node,Bastion",
			hasControlPlane: false,
			hasAPIServer:    false,
			hasEtcd:         false,
			hasScheduler:    false,
			hasKCM:          false,
			hasCCM:          false,
			hasNode:         true,
			hasBastion:      true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.role.HasControlPlane() != tc.hasControlPlane {
				t.Errorf("%q.HasControlPlane() = %v, want %v", tc.role, tc.role.HasControlPlane(), tc.hasControlPlane)
			}
			if tc.role.HasAPIServer() != tc.hasAPIServer {
				t.Errorf("%q.HasAPIServer() = %v, want %v", tc.role, tc.role.HasAPIServer(), tc.hasAPIServer)
			}
			if tc.role.HasEtcd() != tc.hasEtcd {
				t.Errorf("%q.HasEtcd() = %v, want %v", tc.role, tc.role.HasEtcd(), tc.hasEtcd)
			}
			if tc.role.HasScheduler() != tc.hasScheduler {
				t.Errorf("%q.HasScheduler() = %v, want %v", tc.role, tc.role.HasScheduler(), tc.hasScheduler)
			}
			if tc.role.HasKubeControllerManager() != tc.hasKCM {
				t.Errorf("%q.HasKubeControllerManager() = %v, want %v", tc.role, tc.role.HasKubeControllerManager(), tc.hasKCM)
			}
			if tc.role.HasCloudControllerManager() != tc.hasCCM {
				t.Errorf("%q.HasCloudControllerManager() = %v, want %v", tc.role, tc.role.HasCloudControllerManager(), tc.hasCCM)
			}
			if tc.role.HasNode() != tc.hasNode {
				t.Errorf("%q.HasNode() = %v, want %v", tc.role, tc.role.HasNode(), tc.hasNode)
			}
			if tc.role.HasBastion() != tc.hasBastion {
				t.Errorf("%q.HasBastion() = %v, want %v", tc.role, tc.role.HasBastion(), tc.hasBastion)
			}
		})
	}
}
