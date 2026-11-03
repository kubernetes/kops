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

// objectStore is an S3-compatible bucket that the AWS CLI uploads the staged artifacts to, and that
// serves them over https://.
type objectStore struct {
	bucket string
	prefix string
	// awsArgs select the endpoint of the bucket and the ACL of the uploaded objects.
	awsArgs []string
	// awsEnv is the environment of the AWS CLI, or nil to inherit it.
	awsEnv []string
	// objectKey returns the key of the object served at urlPath.
	objectKey func(urlPath string) string
	// publicURL returns the https:// URL that serves urlPath.
	publicURL func(ctx context.Context, urlPath string) (string, error)
}

// s3Store returns the objectStore of the s3:// stage location.
func (b *BuildOptions) s3Store() (*objectStore, error) {
	if b.StageEndpoint != "" {
		bucket, prefix, err := aws.SplitS3URL(b.StageLocation)
		if err != nil {
			return nil, err
		}
		return endpointStore(b.StageEndpoint, bucket, prefix)
	}

	bucket, prefix, err := b.s3StageLocation()
	if err != nil {
		return nil, err
	}
	return &objectStore{
		bucket:    bucket,
		prefix:    prefix,
		awsArgs:   []string{"--region", b.StageRegion},
		objectKey: krel.S3ObjectKey,
		publicURL: func(ctx context.Context, urlPath string) (string, error) {
			return aws.PublicObjectURL(ctx, b.StageRegion, bucket, urlPath)
		},
	}, nil
}

// publishKops is the object store counterpart of "make gcs-publish-ci". The upload runs outside of
// make, because the Makefile unexports the AWS credential env vars.
func (b *BuildOptions) publishKops(ctx context.Context, env []string, store *objectStore) error {
	if err := b.runKopsMake("version-dist-ci", env); err != nil {
		return err
	}
	version, err := kopsVersion(b.KopsRoot, env)
	if err != nil {
		return err
	}

	uploadDir := filepath.Join(b.KopsRoot, ".build", "upload")
	versionPath := path.Join(store.prefix, version)
	if err := store.sync(filepath.Join(uploadDir, "kops", version), versionPath); err != nil {
		return err
	}

	kopsBaseURL, err := store.publicURL(ctx, versionPath)
	if err != nil {
		return err
	}
	// Like gcs-publish-ci, write the marker to where Build() reads the kops base URL from.
	latestPath := filepath.Join(uploadDir, "latest-ci.txt")
	if err := os.WriteFile(latestPath, []byte(kopsBaseURL+"\n"), 0o600); err != nil {
		return fmt.Errorf("failed to WriteFile(%q): %w", latestPath, err)
	}
	return store.copy(latestPath, path.Join(store.prefix, "latest-ci.txt"))
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

// sync uploads the contents of dir to the objects served under urlPath.
func (s *objectStore) sync(dir, urlPath string) error {
	args := append([]string{"s3", "sync", "--no-progress"}, s.awsArgs...)
	return s.runAWSCLI(append(args, dir, s.uri(urlPath)+"/")...)
}

// copy uploads file to the object served at urlPath.
func (s *objectStore) copy(file, urlPath string) error {
	args := append([]string{"s3", "cp", "--no-progress"}, s.awsArgs...)
	return s.runAWSCLI(append(args, file, s.uri(urlPath))...)
}

// uri returns the s3:// URI of the object served at urlPath.
func (s *objectStore) uri(urlPath string) string {
	return "s3://" + s.bucket + "/" + s.objectKey(urlPath)
}

// runAWSCLI runs the AWS CLI, which, like the gcloud CLI used for GCS, uploads large files in
// parallel parts. Its environment holds the credentials.
func (s *objectStore) runAWSCLI(args ...string) error {
	cmd := exec.Command("aws", args...)
	if s.awsEnv != nil {
		cmd.SetEnv(s.awsEnv...)
	}
	exec.InheritOutput(cmd)
	klog.Infof("Executing %q", "aws "+strings.Join(args, " "))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("running aws %s: %w", strings.Join(args[:2], " "), err)
	}
	return nil
}
