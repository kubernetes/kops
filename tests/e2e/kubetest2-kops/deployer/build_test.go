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
	"os"
	"path/filepath"
	"strings"
	"testing"

	"k8s.io/kops/tests/e2e/kubetest2-kops/builder"
)

func TestVerifyBuildFlagsWithExistingAWSStagingBucket(t *testing.T) {
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

func TestVerifyBuildFlagsRejectsInvalidS3Usage(t *testing.T) {
	for _, tc := range []struct {
		name          string
		cloudProvider string
		buildK8s      bool
		errorText     string
	}{
		{
			name:          "non-AWS provider",
			cloudProvider: "gce",
			errorText:     "only supported on AWS",
		},
		{
			name:          "Kubernetes build",
			cloudProvider: "aws",
			buildK8s:      true,
			errorText:     "requires a gs://",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.buildK8s {
				gopath := t.TempDir()
				t.Setenv("GOPATH", gopath)
				if err := os.MkdirAll(filepath.Join(gopath, "src", "k8s.io", "kubernetes"), 0o755); err != nil {
					t.Fatalf("creating Kubernetes checkout: %v", err)
				}
			}
			d := &deployer{
				CloudProvider: tc.cloudProvider,
				StageLocation: "s3://staging",
				BuildOptions:  &builder.BuildOptions{BuildKubernetes: tc.buildK8s},
			}
			err := d.verifyBuildFlags()
			if err == nil || !strings.Contains(err.Error(), tc.errorText) {
				t.Fatalf("expected error containing %q, got %v", tc.errorText, err)
			}
		})
	}
}
