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

func TestVerifyBuildFlagsWithExistingAWSStagingBucket(t *testing.T) {
	t.Setenv("BUILD_ID", "")
	t.Setenv("KOPS_STAGING_BUCKET", "s3://existing-staging")
	d := &deployer{
		CloudProvider: "aws",
		BuildOptions:  &builder.BuildOptions{},
	}
	if err := d.verifyBuildFlags(); err != nil {
		t.Fatalf("verifying build flags: %v", err)
	}
	if d.StageLocation != "s3://existing-staging" {
		t.Errorf("unexpected stage location %q", d.StageLocation)
	}
	if d.KopsBaseURL != d.StageLocation {
		t.Errorf("expected KopsBaseURL %q, got %q", d.StageLocation, d.KopsBaseURL)
	}
	if d.createStagingStore {
		t.Error("existing staging bucket must not be deleted during teardown")
	}
}

func TestVerifyBuildFlagsRejectsGCSStagingBucketOnAWS(t *testing.T) {
	t.Setenv("KOPS_STAGING_BUCKET", "gs://existing-staging")
	d := &deployer{
		CloudProvider: "aws",
		BuildOptions:  &builder.BuildOptions{},
	}
	err := d.verifyBuildFlags()
	if err == nil || !strings.Contains(err.Error(), "must be an s3://") {
		t.Fatalf("expected an S3 staging bucket error, got %v", err)
	}
}

func TestVerifyBuildFlagsRejectsS3StageLocation(t *testing.T) {
	d := &deployer{
		CloudProvider: "aws",
		StageLocation: "s3://staging",
		BuildOptions:  &builder.BuildOptions{},
	}
	err := d.verifyBuildFlags()
	if err == nil || !strings.Contains(err.Error(), "must be a gs:// path") {
		t.Fatalf("expected a gs:// stage-location error, got %v", err)
	}
}

// Without a BUILD_ID, separate kubetest2-kops invocations cannot derive the same staging bucket
// name, so builds stage to the GCS location as they did before S3 staging existed.
func TestVerifyBuildFlagsWithoutBuildIDStagesToGCS(t *testing.T) {
	t.Setenv("BUILD_ID", "")
	t.Setenv("KOPS_STAGING_BUCKET", "")
	t.Setenv("KOPS_BASE_URL", "")
	t.Setenv("JOB_NAME", "pull-kops-e2e-example")
	t.Setenv("PULL_PULL_SHA", "abcdef1")
	d := &deployer{
		CloudProvider: "aws",
		BuildOptions:  &builder.BuildOptions{},
	}
	if err := d.verifyBuildFlags(); err != nil {
		t.Fatalf("verifying build flags: %v", err)
	}
	if expected := "gs://k8s-staging-kops/pulls/pull-kops-e2e-example/pull-abcdef1"; d.StageLocation != expected {
		t.Errorf("expected stage location %q, got %q", expected, d.StageLocation)
	}
	if expected := "https://storage.googleapis.com/k8s-staging-kops/pulls/pull-kops-e2e-example/pull-abcdef1"; d.KopsBaseURL != expected {
		t.Errorf("expected KopsBaseURL %q, got %q", expected, d.KopsBaseURL)
	}
}
