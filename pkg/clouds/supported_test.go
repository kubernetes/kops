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

package clouds

import (
	"os"
	"testing"

	"k8s.io/kops/pkg/apis/kops"
)

func TestGuessCloudForPath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		env     string
		envVal  string
		want    kops.CloudProviderID
		wantErr bool
	}{
		{
			name: "azureblob",
			path: "azureblob://container/path",
			want: kops.CloudProviderAzure,
		},
		{
			name: "do",
			path: "do://bucket/path",
			want: kops.CloudProviderDO,
		},
		{
			name: "hos",
			path: "hos://something",
			want: kops.CloudProviderHetzner,
		},
		{
			name: "gs",
			path: "gs://bucket",
			want: kops.CloudProviderGCE,
		},
		{
			name: "scw",
			path: "scw://bucket",
			want: kops.CloudProviderScaleway,
		},
		{
			name: "swift",
			path: "swift://container",
			want: kops.CloudProviderOpenstack,
		},
		{
			name: "s3_aws",
			path: "s3://bucket",
			env:  "HCLOUD_TOKEN",
			want: kops.CloudProviderAWS,
		},
		{
			name:   "s3_hcloud",
			path:   "s3://bucket",
			env:    "HCLOUD_TOKEN",
			envVal: "token",
			want:   kops.CloudProviderHetzner,
		},
		{
			name:    "unknown",
			path:    "file://local/path",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.env != "" {
				os.Setenv(tc.env, tc.envVal)
			} else {
				os.Unsetenv("HCLOUD_TOKEN")
			}

			got, err := GuessCloudForPath(tc.path)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q, got nil (got=%q)", tc.path, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}
