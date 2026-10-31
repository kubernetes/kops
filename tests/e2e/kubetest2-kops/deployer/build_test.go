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

package deployer

import (
	"strings"
	"testing"

	"k8s.io/kops/tests/e2e/kubetest2-kops/builder"
)

// These cases do not reach the AWS APIs, which an s3:// stage location on AWS needs to create the
// bucket or to look up its region.
func TestVerifyBuildFlagsStageLocation(t *testing.T) {
	cases := []struct {
		name                string
		cloudProvider       string
		stageLocation       string
		stagingBucket       string
		expectedLocation    string
		expectedKopsBaseURL string
		expectedError       string
	}{
		{
			name:                "aws with a gs:// stage location",
			cloudProvider:       "aws",
			stageLocation:       "gs://k8s-staging-kops/pulls/job/pull-abc",
			expectedLocation:    "gs://k8s-staging-kops/pulls/job/pull-abc",
			expectedKopsBaseURL: "https://storage.googleapis.com/k8s-staging-kops/pulls/job/pull-abc",
		},
		{
			name:                "aws with a gs:// KOPS_STAGING_BUCKET",
			cloudProvider:       "aws",
			stagingBucket:       "gs://staging",
			expectedLocation:    "gs://staging",
			expectedKopsBaseURL: "https://storage.googleapis.com/staging",
		},
		{
			name:                "azure without a stage location",
			cloudProvider:       "azure",
			expectedLocation:    "gs://k8s-staging-kops/pulls/job/pull-abc",
			expectedKopsBaseURL: "https://storage.googleapis.com/k8s-staging-kops/pulls/job/pull-abc",
		},
		{
			name:          "aws with a KOPS_STAGING_BUCKET without a scheme",
			cloudProvider: "aws",
			stagingBucket: "staging",
			expectedError: "must be a gs:// or s3:// path",
		},
		{
			name:          "unsupported stage location",
			cloudProvider: "aws",
			stageLocation: "azureblob://account/staging",
			expectedError: "must be a gs:// or s3:// path",
		},
		{
			name:          "s3:// stage location on another cloud",
			cloudProvider: "gce",
			stageLocation: "s3://staging",
			expectedError: "requires --cloud-provider=aws",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("KOPS_BASE_URL", "")
			t.Setenv("KOPS_STAGING_BUCKET", tc.stagingBucket)
			t.Setenv("JOB_NAME", "job")
			t.Setenv("PULL_PULL_SHA", "abc")

			d := &deployer{
				CloudProvider: tc.cloudProvider,
				StageLocation: tc.stageLocation,
				BuildOptions:  &builder.BuildOptions{},
			}
			err := d.verifyBuildFlags()
			if tc.expectedError != "" {
				if err == nil || !strings.Contains(err.Error(), tc.expectedError) {
					t.Fatalf("verifyBuildFlags() = %v, expected an error containing %q", err, tc.expectedError)
				}
				return
			}
			if err != nil {
				t.Fatalf("verifyBuildFlags() = %v", err)
			}
			if d.BuildOptions.StageLocation != tc.expectedLocation {
				t.Errorf("StageLocation = %q, expected %q", d.BuildOptions.StageLocation, tc.expectedLocation)
			}
			if d.BuildOptions.StageRegion != "" {
				t.Errorf("StageRegion = %q, expected it unset for a gs:// stage location", d.BuildOptions.StageRegion)
			}
			if d.KopsBaseURL != tc.expectedKopsBaseURL {
				t.Errorf("KopsBaseURL = %q, expected %q", d.KopsBaseURL, tc.expectedKopsBaseURL)
			}
		})
	}
}
