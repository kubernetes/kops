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
