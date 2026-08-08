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
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

func TestStagingBucketName(t *testing.T) {
	t.Setenv("BUILD_ID", "12345678901234567890")
	name, err := (Client{}).BucketName(context.Background(), BucketTypeStagingStore)
	if err != nil {
		t.Fatalf("building bucket name: %v", err)
	}
	if expected := "k8s-infra-kops-staging-12345678901234567890"; name != expected {
		t.Errorf("expected %q, got %q", expected, name)
	}
	if len(name) > 63 {
		t.Errorf("bucket name %q is longer than the 63 character limit", name)
	}
}

// Teardown can run without --build, so a missing staging bucket must be a no-op.
func TestDeleteMissingS3Bucket(t *testing.T) {
	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.RequestURI())
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?><Error><Code>NoSuchBucket</Code><Message>The specified bucket does not exist</Message></Error>`)
	}))
	defer server.Close()

	c := Client{s3Client: s3.New(s3.Options{
		Region:           defaultRegion,
		BaseEndpoint:     aws.String(server.URL),
		Credentials:      aws.AnonymousCredentials{},
		UsePathStyle:     true,
		RetryMaxAttempts: 1,
	})}

	if err := c.deleteS3Bucket(context.Background(), "does-not-exist", true); err != nil {
		t.Errorf("deleting a bucket that does not exist: %v", err)
	}
	if len(requests) != 1 || !strings.Contains(requests[0], "location") {
		t.Errorf("expected the bucket location request only, got %v", requests)
	}
}

func TestIsNoSuchBucket(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
		want bool
	}{
		{name: "typed", err: &types.NoSuchBucket{}, want: true},
		{name: "generic code", err: &smithy.GenericAPIError{Code: "NoSuchBucket"}, want: true},
		{name: "generic not found", err: &smithy.GenericAPIError{Code: "NotFound"}, want: true},
		{name: "other", err: errors.New("other")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := isNoSuchBucket(tc.err); got != tc.want {
				t.Errorf("isNoSuchBucket() = %v, expected %v", got, tc.want)
			}
		})
	}
}
