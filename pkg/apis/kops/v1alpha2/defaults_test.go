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

package v1alpha2

import "testing"

// TestSetDefaultsClusterSpecAuthorization checks that an omitted or empty
// authorization defaults to RBAC, and that an explicit AlwaysAllow is preserved.
func TestSetDefaultsClusterSpecAuthorization(t *testing.T) {
	grid := []struct {
		name            string
		authorization   *AuthorizationSpec
		wantRBAC        bool
		wantAlwaysAllow bool
	}{
		{
			name:          "omitted defaults to RBAC",
			authorization: nil,
			wantRBAC:      true,
		},
		{
			name:          "empty defaults to RBAC",
			authorization: &AuthorizationSpec{},
			wantRBAC:      true,
		},
		{
			name:            "explicit AlwaysAllow is preserved",
			authorization:   &AuthorizationSpec{AlwaysAllow: &AlwaysAllowAuthorizationSpec{}},
			wantAlwaysAllow: true,
		},
		{
			name:          "explicit RBAC is preserved",
			authorization: &AuthorizationSpec{RBAC: &RBACAuthorizationSpec{}},
			wantRBAC:      true,
		},
	}
	for _, g := range grid {
		t.Run(g.name, func(t *testing.T) {
			spec := &ClusterSpec{Authorization: g.authorization}

			SetDefaults_ClusterSpec(spec)

			if spec.Authorization == nil {
				t.Fatalf("Authorization was not defaulted")
			}
			if got := spec.Authorization.RBAC != nil; got != g.wantRBAC {
				t.Errorf("Authorization.RBAC set = %v, want %v", got, g.wantRBAC)
			}
			if got := spec.Authorization.AlwaysAllow != nil; got != g.wantAlwaysAllow {
				t.Errorf("Authorization.AlwaysAllow set = %v, want %v", got, g.wantAlwaysAllow)
			}
		})
	}
}
