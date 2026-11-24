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

package main

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kopsapi "k8s.io/kops/pkg/apis/kops"
)

func TestSelectInstanceGroupsToValidate(t *testing.T) {
	items := []kopsapi.InstanceGroup{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "control-plane-a"},
			Spec:       kopsapi.InstanceGroupSpec{Role: kopsapi.InstanceGroupRoleControlPlane},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "control-plane-b"},
			Spec:       kopsapi.InstanceGroupSpec{Role: kopsapi.InstanceGroupRoleControlPlane},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "nodes-a"},
			Spec:       kopsapi.InstanceGroupSpec{Role: kopsapi.InstanceGroupRoleNode},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "addons"},
			Spec:       kopsapi.InstanceGroupSpec{Role: kopsapi.InstanceGroupRoleNode},
		},
	}

	tests := []struct {
		name      string
		names     []string
		roles     []string
		wantNames []string
		wantErr   string
	}{
		{
			name:      "single instance group by name",
			names:     []string{"addons"},
			wantNames: []string{"addons"},
		},
		{
			name:      "multiple instance groups preserve order and deduplicate",
			names:     []string{"addons", "nodes-a", "addons"},
			wantNames: []string{"addons", "nodes-a"},
		},
		{
			name:      "by role case-insensitive",
			roles:     []string{"control-plane"},
			wantNames: []string{"control-plane-a", "control-plane-b"},
		},
		{
			name:      "combined names and roles deduplicate",
			names:     []string{"addons", "control-plane-b"},
			roles:     []string{"control-plane"},
			wantNames: []string{"addons", "control-plane-b", "control-plane-a"},
		},
		{
			name:    "unknown instance group name",
			names:   []string{"missing"},
			wantErr: `InstanceGroup "missing" not found`,
		},
		{
			name:    "invalid role",
			roles:   []string{"worker"},
			wantErr: `invalid instance group role "worker"`,
		},
		{
			name:    "role with no matching instance groups",
			roles:   []string{"bastion"},
			wantErr: "no matching InstanceGroup objects found",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := selectInstanceGroupsToValidate(items, tc.names, tc.roles)
			if tc.wantErr != "" {
				require.EqualError(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			var gotNames []string
			for _, ig := range got {
				gotNames = append(gotNames, ig.ObjectMeta.Name)
			}
			assert.Equal(t, tc.wantNames, gotNames)
		})
	}
}

func TestValidateClusterFlagNormalization(t *testing.T) {
	cmd := NewCmdValidateCluster(nil, io.Discard)
	err := cmd.ParseFlags([]string{
		"--ig=addons",
		"--instance-groups=nodes-a",
		"--role=control-plane",
		"--roles=node",
	})
	require.NoError(t, err)

	igs, err := cmd.Flags().GetStringSlice("instance-group")
	require.NoError(t, err)
	assert.Equal(t, []string{"addons", "nodes-a"}, igs)

	roles, err := cmd.Flags().GetStringSlice("instance-group-roles")
	require.NoError(t, err)
	assert.Equal(t, []string{"control-plane", "node"}, roles)
}
