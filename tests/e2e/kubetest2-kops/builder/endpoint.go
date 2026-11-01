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
	"fmt"
	"os"
	"strings"

	"k8s.io/kops/tests/e2e/kubetest2-kops/aws"
)

// endpointStore returns the objectStore of a bucket at an S3-compatible endpoint outside AWS, such
// as DigitalOcean Spaces. It uses the keys that kops reads for such endpoints.
func endpointStore(endpoint, bucket, prefix string) (*objectStore, error) {
	accessKey, secretKey := os.Getenv("S3_ACCESS_KEY_ID"), os.Getenv("S3_SECRET_ACCESS_KEY")
	if accessKey == "" || secretKey == "" {
		return nil, fmt.Errorf("an s3:// stage location at %s requires the S3_ACCESS_KEY_ID and S3_SECRET_ACCESS_KEY env vars", endpoint)
	}

	return &objectStore{
		bucket: bucket,
		prefix: prefix,
		// The endpoint selects the region, the region only signs the requests. Unlike on S3, the
		// objects are only public with a public-read ACL.
		awsArgs: []string{"--region", "us-east-1", "--endpoint-url", endpoint, "--acl", "public-read"},
		awsEnv: append(os.Environ(),
			"AWS_ACCESS_KEY_ID="+accessKey,
			"AWS_SECRET_ACCESS_KEY="+secretKey,
			// Newer AWS CLIs add checksums by default, which S3-compatible stores may not support.
			"AWS_REQUEST_CHECKSUM_CALCULATION=when_required",
			"AWS_RESPONSE_CHECKSUM_VALIDATION=when_required",
		),
		// Unlike S3, Ceph-based stores like Spaces keep a literal "+" in a URL path.
		objectKey: func(urlPath string) string {
			return strings.Trim(urlPath, "/")
		},
		publicURL: func(_ context.Context, urlPath string) (string, error) {
			return aws.EndpointObjectURL(endpoint, bucket, urlPath)
		},
	}, nil
}
