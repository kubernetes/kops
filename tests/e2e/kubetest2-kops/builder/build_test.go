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

package builder

import "testing"

func TestKubeStageLocation(t *testing.T) {
	tests := []struct {
		name          string
		stageLocation string
		wantBucket    string
		wantGCSRoot   string
		wantMarkerURL string
		wantErr       bool
	}{
		{
			name:          "bucket root",
			stageLocation: "gs://project-staging-1234",
			wantBucket:    "project-staging-1234",
			wantGCSRoot:   "kubernetes",
			wantMarkerURL: "https://storage.googleapis.com/project-staging-1234/kubernetes/latest.txt",
		},
		{
			name:          "bucket root with trailing slash",
			stageLocation: "gs://project-staging-1234/",
			wantBucket:    "project-staging-1234",
			wantGCSRoot:   "kubernetes",
			wantMarkerURL: "https://storage.googleapis.com/project-staging-1234/kubernetes/latest.txt",
		},
		{
			name:          "path inside a shared bucket",
			stageLocation: "gs://k8s-staging-kops/pulls/pull-kops-kubernetes-e2e-ubuntu-ec2-build/pull-0d6ca211",
			wantBucket:    "k8s-staging-kops",
			wantGCSRoot:   "pulls/pull-kops-kubernetes-e2e-ubuntu-ec2-build/pull-0d6ca211/kubernetes",
			wantMarkerURL: "https://storage.googleapis.com/k8s-staging-kops/pulls/pull-kops-kubernetes-e2e-ubuntu-ec2-build/pull-0d6ca211/kubernetes/latest.txt",
		},
		{
			name:          "not a gs:// path",
			stageLocation: "s3://bucket/path",
			wantErr:       true,
		},
		{
			name:          "missing bucket name",
			stageLocation: "gs:///path",
			wantErr:       true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			bucket, gcsRoot, markerURL, err := kubeStageLocation(tc.stageLocation)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got bucket=%q gcsRoot=%q markerURL=%q", bucket, gcsRoot, markerURL)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if bucket != tc.wantBucket {
				t.Errorf("bucket: got %q, want %q", bucket, tc.wantBucket)
			}
			if gcsRoot != tc.wantGCSRoot {
				t.Errorf("gcsRoot: got %q, want %q", gcsRoot, tc.wantGCSRoot)
			}
			if markerURL != tc.wantMarkerURL {
				t.Errorf("markerURL: got %q, want %q", markerURL, tc.wantMarkerURL)
			}
		})
	}
}
