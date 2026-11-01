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

import (
	"context"
	"testing"
)

func TestKubeStageLocation(t *testing.T) {
	cases := []struct {
		name               string
		stageLocation      string
		stageRegion        string
		stageEndpoint      string
		expectedBucket     string
		expectedBucketRoot string
		expectedMarkerURL  string
		expectError        bool
	}{
		{
			name:               "gcs bucket",
			stageLocation:      "gs://staging",
			expectedBucket:     "staging",
			expectedBucketRoot: "kubernetes",
			expectedMarkerURL:  "https://storage.googleapis.com/staging/kubernetes/latest.txt",
		},
		{
			name:               "gcs prefix",
			stageLocation:      "gs://k8s-staging-kops/pulls/job/pull-abc/",
			expectedBucket:     "k8s-staging-kops",
			expectedBucketRoot: "pulls/job/pull-abc/kubernetes",
			expectedMarkerURL:  "https://storage.googleapis.com/k8s-staging-kops/pulls/job/pull-abc/kubernetes/latest.txt",
		},
		{
			name:          "gcs without a bucket",
			stageLocation: "gs://",
			expectError:   true,
		},
		{
			name:               "s3 prefix",
			stageLocation:      "s3://staging/prefix",
			stageRegion:        "us-east-2",
			expectedBucket:     "s3://staging",
			expectedBucketRoot: "prefix/kubernetes",
			expectedMarkerURL:  "https://staging.s3.us-east-2.amazonaws.com/prefix/kubernetes/latest.txt",
		},
		{
			name:          "s3 at another endpoint",
			stageLocation: "s3://kops-e2e-staging/prefix",
			stageEndpoint: "https://sfo3.digitaloceanspaces.com",
			expectError:   true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b := &BuildOptions{StageLocation: tc.stageLocation, StageRegion: tc.stageRegion, StageEndpoint: tc.stageEndpoint}
			bucket, bucketRoot, markerURL, err := b.kubeStageLocation(context.Background())
			if tc.expectError {
				if err == nil {
					t.Fatalf("kubeStageLocation() succeeded, expected an error")
				}
				return
			}
			if err != nil {
				t.Fatalf("kubeStageLocation() = %v", err)
			}
			if bucket != tc.expectedBucket || bucketRoot != tc.expectedBucketRoot || markerURL != tc.expectedMarkerURL {
				t.Errorf("kubeStageLocation() = (%q, %q, %q), expected (%q, %q, %q)",
					bucket, bucketRoot, markerURL, tc.expectedBucket, tc.expectedBucketRoot, tc.expectedMarkerURL)
			}
		})
	}
}
