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

func Test_VFSPath(t *testing.T) {
	grid := []struct {
		Input          string
		ExpectedResult string
		ExpectError    bool
	}{
		{
			Input:          "s3.amazonaws.com/bucket",
			ExpectedResult: "s3://bucket",
			ExpectError:    false,
		},
		{
			Input:          "s3-bucket.amazonaws.com",
			ExpectedResult: "s3://bucket",
			ExpectError:    false,
		},
		{
			Input:          "s3-bucket.amazonaws.com/path",
			ExpectedResult: "s3://bucket/path",
			ExpectError:    false,
		},
		{
			Input:          "s3.bucket.amazonaws.com",
			ExpectedResult: "s3://bucket",
			ExpectError:    false,
		},
		{
			Input:          "s3.bucket_foo-bar.abc.amazonaws.com/path",
			ExpectedResult: "s3://bucket_foo-bar.abc/path",
			ExpectError:    false,
		},
		{
			Input:          "s3-us-west-2.amazonaws.com/bucket/path",
			ExpectedResult: "s3://bucket/path",
			ExpectError:    false,
		},
		{
			Input:          "s3-us-west-2.amazonaws.com/bucket/path",
			ExpectedResult: "s3://bucket/path",
			ExpectError:    false,
		},
		{
			Input:          "s3.cn-north-1.amazonaws.com.cn/bucket",
			ExpectedResult: "s3://bucket",
			ExpectError:    false,
		},
		{
			Input:          "s3.cn-north-1.amazonaws.com.cn/bucket/path",
			ExpectedResult: "s3://bucket/path",
			ExpectError:    false,
		},
		{
			Input:          "https://s3.amazonaws.com/bucket",
			ExpectedResult: "s3://bucket",
			ExpectError:    false,
		},
		{
			Input:          "http://s3.amazonaws.com/bucket",
			ExpectedResult: "s3://bucket",
			ExpectError:    false,
		},
		{
			Input:          "https://bucket-name.s3.us-east-1.amazonaws.com/path",
			ExpectedResult: "s3://bucket-name/path",
			ExpectError:    false,
		},
		{
			Input:          "example.com/bucket",
			ExpectedResult: "",
			ExpectError:    true,
		},
		{
			Input:          "https://example.com/bucket",
			ExpectedResult: "",
			ExpectError:    true,
		},
		{
			Input:          "storage.googleapis.com",
			ExpectedResult: "",
			ExpectError:    true,
		},
		{
			Input:          "storage.googleapis.com/foo/bar",
			ExpectedResult: "",
			ExpectError:    true,
		},
		{
			Input:          "https://storage.googleapis.com",
			ExpectedResult: "",
			ExpectError:    true,
		},
		{
			Input:          "https://storage.googleapis.com/foo/bar",
			ExpectedResult: "",
			ExpectError:    true,
		},
	}
	for _, g := range grid {
		vfsPath, err := VFSPath(g.Input)
		if !g.ExpectError {
			if err != nil {
				t.Fatalf("unexpected error parsing vfs path: %v", err)
			}
			if vfsPath != g.ExpectedResult {
				t.Fatalf("s3 url does not match expected result (%v): %v", g.ExpectedResult, g.Input)
			}
		} else {
			if err == nil {
				t.Fatalf("unexpected error parsing %q", g.Input)
			}
		}
	}
}
