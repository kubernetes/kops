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

package aws

import (
	"context"
	"testing"
)

func TestBucketNameStagingStore(t *testing.T) {
	// BUILD_ID makes the name deterministic, so that a later kubetest2 invocation in the same
	// job can delete the staging bucket. It also avoids the STS call.
	t.Setenv("BUILD_ID", "1849320394810294272")

	actual, err := Client{}.BucketName(context.Background(), BucketTypeStagingStore)
	if err != nil {
		t.Fatalf("BucketName() = %v", err)
	}
	if expected := "k8s-infra-kops-staging-1849320394810294272"; actual != expected {
		t.Errorf("BucketName() = %q, expected %q", actual, expected)
	}
}

func TestSplitS3URL(t *testing.T) {
	cases := []struct {
		location       string
		expectedBucket string
		expectedPrefix string
		expectError    bool
	}{
		{location: "s3://bucket", expectedBucket: "bucket"},
		{location: "s3://bucket/", expectedBucket: "bucket"},
		{location: "s3://bucket/a/b/", expectedBucket: "bucket", expectedPrefix: "a/b"},
		{location: "gs://bucket/a", expectError: true},
		{location: "s3:///a", expectError: true},
	}
	for _, tc := range cases {
		t.Run(tc.location, func(t *testing.T) {
			bucket, prefix, err := SplitS3URL(tc.location)
			if tc.expectError {
				if err == nil {
					t.Fatalf("SplitS3URL(%q) = (%q, %q), expected an error", tc.location, bucket, prefix)
				}
				return
			}
			if err != nil {
				t.Fatalf("SplitS3URL(%q) = %v", tc.location, err)
			}
			if bucket != tc.expectedBucket || prefix != tc.expectedPrefix {
				t.Errorf("SplitS3URL(%q) = (%q, %q), expected (%q, %q)", tc.location, bucket, prefix, tc.expectedBucket, tc.expectedPrefix)
			}
		})
	}
}

func TestPublicObjectURL(t *testing.T) {
	cases := []struct {
		name     string
		region   string
		bucket   string
		urlPath  string
		expected string
	}{
		{
			name:     "regional endpoint keeps a literal plus",
			region:   "us-east-2",
			bucket:   "k8s-infra-kops-staging-1",
			urlPath:  "kubernetes/v1.37.0-alpha.1.123+0123456789ab",
			expected: "https://k8s-infra-kops-staging-1.s3.us-east-2.amazonaws.com/kubernetes/v1.37.0-alpha.1.123+0123456789ab",
		},
		{
			name:     "bucket root",
			region:   "us-west-2",
			bucket:   "staging",
			urlPath:  "",
			expected: "https://staging.s3.us-west-2.amazonaws.com/",
		},
		{
			name:     "partition DNS suffix",
			region:   "cn-north-1",
			bucket:   "staging",
			urlPath:  "/1.37.0-alpha.1+v1.37.0-alpha.0-1-gabcdef/",
			expected: "https://staging.s3.cn-north-1.amazonaws.com.cn/1.37.0-alpha.1+v1.37.0-alpha.0-1-gabcdef",
		},
		{
			name:     "bucket that cannot be a subdomain",
			region:   "us-east-2",
			bucket:   "staging.example.com",
			urlPath:  "a/b",
			expected: "https://s3.us-east-2.amazonaws.com/staging.example.com/a/b",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := PublicObjectURL(context.Background(), tc.region, tc.bucket, tc.urlPath)
			if err != nil {
				t.Fatalf("PublicObjectURL() = %v", err)
			}
			if actual != tc.expected {
				t.Errorf("PublicObjectURL() = %q, expected %q", actual, tc.expected)
			}
		})
	}
}

func TestEndpointObjectURL(t *testing.T) {
	actual, err := EndpointObjectURL("https://sfo3.digitaloceanspaces.com", "kops-e2e-staging", "/pulls/job/1.37.0-alpha.1+v1.37.0-alpha.0-1-gabcdef/")
	if err != nil {
		t.Fatalf("EndpointObjectURL() = %v", err)
	}
	if expected := "https://kops-e2e-staging.sfo3.digitaloceanspaces.com/pulls/job/1.37.0-alpha.1+v1.37.0-alpha.0-1-gabcdef"; actual != expected {
		t.Errorf("EndpointObjectURL() = %q, expected %q", actual, expected)
	}

	if _, err := EndpointObjectURL("sfo3.digitaloceanspaces.com", "kops-e2e-staging", ""); err == nil {
		t.Errorf("EndpointObjectURL() with an endpoint without a scheme succeeded, expected an error")
	}
}
