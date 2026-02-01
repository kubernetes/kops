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

package vfs

import (
	"testing"
)

func TestOpenstackConfig_GetInsecureSkipVerify(t *testing.T) {
	tests := []struct {
		name   string
		envVal string
		envSet bool // if false, don't set the env var (simulate unset)
		want   bool
	}{
		{
			name:   "Not set",
			envSet: false,
			want:   false,
		},
		{
			name:   "Set to empty string",
			envVal: "",
			envSet: true,
			want:   false,
		},
		{
			name:   "Set to true",
			envVal: "true",
			envSet: true,
			want:   true,
		},
		{
			name:   "Set to 1",
			envVal: "1",
			envSet: true,
			want:   true,
		},
		{
			name:   "Set to false",
			envVal: "false",
			envSet: true,
			want:   false,
		},
		{
			name:   "Set to 0",
			envVal: "0",
			envSet: true,
			want:   false,
		},
		{
			name:   "Set to other",
			envVal: "foo",
			envSet: true,
			want:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.envSet {
				t.Setenv("KOPS_OS_TLS_INSECURE_SKIP_VERIFY", tc.envVal)
			} else {
				// Ensure it's not set from the environment where tests are running
				t.Setenv("KOPS_OS_TLS_INSECURE_SKIP_VERIFY", "")
			}

			oc := OpenstackConfig{}
			got := oc.GetInsecureSkipVerify()
			if got != tc.want {
				t.Errorf("GetInsecureSkipVerify() = %v, want %v", got, tc.want)
			}
		})
	}
}
