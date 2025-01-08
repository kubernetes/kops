/*
Copyright 2019 The Kubernetes Authors.

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

package vfs

import "testing"

func Test_S3Path_Parse(t *testing.T) {
	grid := []struct {
		Input          string
		ExpectError    bool
		ExpectedBucket string
		ExpectedPath   string
	}{
		{
			Input:          "s3://bucket",
			ExpectedBucket: "bucket",
			ExpectedPath:   "",
		},
		{
			Input:          "s3://bucket/path",
			ExpectedBucket: "bucket",
			ExpectedPath:   "path",
		},
		{
			Input:          "s3://bucket2/path/subpath",
			ExpectedBucket: "bucket2",
			ExpectedPath:   "path/subpath",
		},
		{
			Input:       "s3:///bucket/path/subpath",
			ExpectError: true,
		},
	}
	for _, g := range grid {
		s3path, err := Context.buildS3Path(g.Input)
		if !g.ExpectError {
			if err != nil {
				t.Fatalf("unexpected error parsing s3 path: %v", err)
			}
			if s3path.bucket != g.ExpectedBucket {
				t.Fatalf("unexpected s3 path: %v", s3path)
			}
			if s3path.key != g.ExpectedPath {
				t.Fatalf("unexpected s3 path: %v", s3path)
			}
		} else {
			if err == nil {
				t.Fatalf("unexpected error parsing %q", g.Input)
			}
		}
	}
}

func TestGetHTTPsUrl(t *testing.T) {
	grid := []struct {
		Path        string
		Dualstack   bool
		Region      string
		ExpectedURL string
	}{
		{
			Path:        "s3://bucket",
			Region:      "us-east-1",
			ExpectedURL: "https://bucket.s3.us-east-1.amazonaws.com",
		},
		{
			Path:        "s3://bucket.with.forced.path.style/subpath",
			Region:      "us-east-1",
			ExpectedURL: "https://s3.us-east-1.amazonaws.com/bucket.with.forced.path.style/subpath",
		},
		{
			Path:        "s3://bucket/path",
			Region:      "us-east-2",
			ExpectedURL: "https://bucket.s3.us-east-2.amazonaws.com/path",
		},
		{
			Path:        "s3://bucket2/path/subpath",
			Region:      "us-east-1",
			ExpectedURL: "https://bucket2.s3.us-east-1.amazonaws.com/path/subpath",
		},
		{
			Path:        "s3://bucket2-ds/path/subpath",
			Dualstack:   true,
			Region:      "us-east-1",
			ExpectedURL: "https://bucket2-ds.s3.dualstack.us-east-1.amazonaws.com/path/subpath",
		},
		{
			Path:        "s3://bucket2-cn/path/subpath",
			Region:      "cn-north-1",
			ExpectedURL: "https://bucket2-cn.s3.cn-north-1.amazonaws.com.cn/path/subpath",
		},
		{
			Path:        "s3://bucket2-cn-ds/path/subpath",
			Dualstack:   true,
			Region:      "cn-north-1",
			ExpectedURL: "https://bucket2-cn-ds.s3.dualstack.cn-north-1.amazonaws.com.cn/path/subpath",
		},
		{
			Path:        "s3://bucket2-gov/path/subpath",
			Region:      "us-gov-west-1",
			ExpectedURL: "https://bucket2-gov.s3.us-gov-west-1.amazonaws.com/path/subpath",
		},
		{
			Path:        "s3://bucket2-gov-ds/path/subpath",
			Dualstack:   true,
			Region:      "us-gov-west-1",
			ExpectedURL: "https://bucket2-gov-ds.s3.dualstack.us-gov-west-1.amazonaws.com/path/subpath",
		},
	}
	for _, g := range grid {
		t.Run(g.Path, func(t *testing.T) {
			// Must be nonempty in order to force S3_REGION usage
			// rather than querying S3 for the region.
			t.Setenv("S3_ENDPOINT", "1")
			t.Setenv("S3_REGION", g.Region)
			s3path, _ := Context.buildS3Path(g.Path)
			url, err := s3path.GetHTTPsUrl(g.Dualstack)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if url != g.ExpectedURL {
				t.Fatalf("expected url: %v vs actual url: %v", g.ExpectedURL, url)
			}
		})
	}
}
