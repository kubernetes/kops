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
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"k8s.io/klog/v2"
	"k8s.io/kops/tests/e2e/kubetest2-kops/aws"
	krel "k8s.io/release/pkg/build"
	"sigs.k8s.io/kubetest2/pkg/exec"
)

// The objects staged in S3 are served over https:// from a publicly readable bucket, like the GCS
// staging bucket. Their keys follow krel.S3ObjectKey, the convention krel uses for the Kubernetes
// artifacts, so that the URLs that kops builds from a version containing a "+" resolve.

// publishKopsToS3 is the S3 counterpart of "make gcs-publish-ci". The upload runs outside of make,
// because the Makefile unexports the AWS credential env vars.
func (b *BuildOptions) publishKopsToS3(ctx context.Context, env []string) error {
	bucket, prefix, err := b.s3StageLocation()
	if err != nil {
		return err
	}

	if err := b.runKopsMake("version-dist-ci", env); err != nil {
		return err
	}
	version, err := kopsVersion(b.KopsRoot, env)
	if err != nil {
		return err
	}

	uploadDir := filepath.Join(b.KopsRoot, ".build", "upload")
	versionPath := path.Join(prefix, version)
	if err := s3Sync(filepath.Join(uploadDir, "kops", version), bucket, versionPath, b.StageRegion); err != nil {
		return err
	}

	kopsBaseURL, err := aws.PublicObjectURL(ctx, b.StageRegion, bucket, versionPath)
	if err != nil {
		return err
	}
	// Like gcs-publish-ci, write the marker to where Build() reads the kops base URL from.
	latestPath := filepath.Join(uploadDir, "latest-ci.txt")
	if err := os.WriteFile(latestPath, []byte(kopsBaseURL+"\n"), 0o600); err != nil {
		return fmt.Errorf("failed to WriteFile(%q): %w", latestPath, err)
	}
	return s3Copy(latestPath, bucket, path.Join(prefix, "latest-ci.txt"), b.StageRegion)
}

// s3KubeStageLocation is the s3:// counterpart of kubeStageLocation.
func (b *BuildOptions) s3KubeStageLocation(ctx context.Context) (bucket, bucketRoot, markerURL string, err error) {
	s3Bucket, prefix, err := b.s3StageLocation()
	if err != nil {
		return "", "", "", err
	}

	bucketRoot = path.Join(prefix, "kubernetes")
	markerURL, err = aws.PublicObjectURL(ctx, b.StageRegion, s3Bucket, path.Join(bucketRoot, "latest.txt"))
	if err != nil {
		return "", "", "", err
	}
	return "s3://" + s3Bucket, bucketRoot, markerURL, nil
}

func (b *BuildOptions) s3StageLocation() (bucket, prefix string, err error) {
	if b.StageRegion == "" {
		return "", "", errors.New("the region of the s3:// stage location is not set")
	}
	return aws.SplitS3URL(b.StageLocation)
}

// kopsVersion returns the version that the Makefile assigns to a build in the same environment.
func kopsVersion(kopsRoot string, env []string) (string, error) {
	cmd := exec.Command("tools/get_version.sh")
	cmd.SetDir(kopsRoot)
	cmd.SetEnv(env...)
	lines, err := exec.OutputLines(cmd)
	if err != nil {
		return "", fmt.Errorf("running tools/get_version.sh: %w", err)
	}
	return parseKopsVersion(lines)
}

// parseKopsVersion extracts the version from the output of tools/get_version.sh.
func parseKopsVersion(lines []string) (string, error) {
	for _, line := range lines {
		if version, ok := strings.CutPrefix(line, "VERSION "); ok {
			if version = strings.TrimSpace(version); version != "" {
				return version, nil
			}
		}
	}
	return "", fmt.Errorf("no VERSION in the output of tools/get_version.sh: %q", lines)
}

// s3Sync uploads the contents of dir to the objects served under urlPath in bucket.
func s3Sync(dir, bucket, urlPath, region string) error {
	return runAWSCLI("s3", "sync", "--no-progress", "--region", region, dir, s3URI(bucket, urlPath)+"/")
}

// s3Copy uploads file to the object served at urlPath in bucket.
func s3Copy(file, bucket, urlPath, region string) error {
	return runAWSCLI("s3", "cp", "--no-progress", "--region", region, file, s3URI(bucket, urlPath))
}

// s3URI returns the s3:// URI of the object that S3 serves at urlPath in bucket.
func s3URI(bucket, urlPath string) string {
	return "s3://" + bucket + "/" + krel.S3ObjectKey(urlPath)
}

// runAWSCLI runs the AWS CLI, which, like the gcloud CLI used for GCS, uploads large files in
// parallel parts. It inherits the environment, which holds the AWS credentials.
func runAWSCLI(args ...string) error {
	cmd := exec.Command("aws", args...)
	exec.InheritOutput(cmd)
	klog.Infof("Executing %q", "aws "+strings.Join(args, " "))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("running aws %s: %w", strings.Join(args[:2], " "), err)
	}
	return nil
}
