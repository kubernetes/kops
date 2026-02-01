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

package env

import (
	"testing"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
)

func TestBuildSystemComponentEnvVars(t *testing.T) {
	tests := []struct {
		name      string
		spec      *kops.ClusterSpec
		envVar    string
		wantVal   string
		wantExist bool
	}{
		{
			name: "Openstack nil",
			spec: &kops.ClusterSpec{
				CloudProvider: kops.CloudProviderSpec{},
			},
			envVar:    "KOPS_OS_TLS_INSECURE_SKIP_VERIFY",
			wantExist: false,
		},
		{
			name: "Openstack InsecureSkipVerify nil",
			spec: &kops.ClusterSpec{
				CloudProvider: kops.CloudProviderSpec{
					Openstack: &kops.OpenstackSpec{},
				},
			},
			envVar:    "KOPS_OS_TLS_INSECURE_SKIP_VERIFY",
			wantExist: false,
		},
		{
			name: "Openstack InsecureSkipVerify false",
			spec: &kops.ClusterSpec{
				CloudProvider: kops.CloudProviderSpec{
					Openstack: &kops.OpenstackSpec{
						InsecureSkipVerify: fi.PtrTo(false),
					},
				},
			},
			envVar:    "KOPS_OS_TLS_INSECURE_SKIP_VERIFY",
			wantExist: false,
		},
		{
			name: "Openstack InsecureSkipVerify true",
			spec: &kops.ClusterSpec{
				CloudProvider: kops.CloudProviderSpec{
					Openstack: &kops.OpenstackSpec{
						InsecureSkipVerify: fi.PtrTo(true),
					},
				},
			},
			envVar:    "KOPS_OS_TLS_INSECURE_SKIP_VERIFY",
			wantVal:   "true",
			wantExist: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			vars := BuildSystemComponentEnvVars(tc.spec)
			val, ok := vars[tc.envVar]
			if ok != tc.wantExist {
				t.Errorf("Expected existence of key %q to be %v, but got %v", tc.envVar, tc.wantExist, ok)
			}
			if tc.wantExist && val != tc.wantVal {
				t.Errorf("Expected value of key %q to be %q, but got %q", tc.envVar, tc.wantVal, val)
			}
		})
	}
}
