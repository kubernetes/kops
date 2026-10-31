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
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"
)

func TestParseKopsVersion(t *testing.T) {
	cases := []struct {
		name        string
		lines       []string
		expected    string
		expectError bool
	}{
		{
			name:     "release version",
			lines:    []string{"VERSION 1.37.0-alpha.1", "IMAGE_TAG 1.37.0-alpha.1"},
			expected: "1.37.0-alpha.1",
		},
		{
			name:     "CI version",
			lines:    []string{"VERSION 1.37.0-alpha.1+v1.37.0-alpha.0-12-gabcdef0", "IMAGE_TAG 1.37.0-alpha.1-v1.37.0-alpha.0-12-gabcdef0"},
			expected: "1.37.0-alpha.1+v1.37.0-alpha.0-12-gabcdef0",
		},
		{
			name:        "no version",
			lines:       []string{"Build was tagged with 1.37.0, but kops-version.go had version 1.37.1"},
			expectError: true,
		},
		{
			name:        "empty version",
			lines:       []string{"VERSION ", "IMAGE_TAG "},
			expectError: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := parseKopsVersion(tc.lines)
			if tc.expectError {
				if err == nil {
					t.Fatalf("parseKopsVersion() = %q, expected an error", actual)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseKopsVersion() = %v", err)
			}
			if actual != tc.expected {
				t.Errorf("parseKopsVersion() = %q, expected %q", actual, tc.expected)
			}
		})
	}
}

// The S3 upload must use the version that the Makefile assigns to the build, which comes from
// tools/get_version.sh. Without CI set, that is the release version in kops-version.go.
func TestKopsVersionMatchesKopsVersionGo(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is required by tools/get_version.sh")
	}
	kopsRoot, err := filepath.Abs(filepath.Join("..", "..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	src, err := os.ReadFile(filepath.Join(kopsRoot, "kops-version.go"))
	if err != nil {
		t.Fatal(err)
	}
	match := regexp.MustCompile(`KOPS_RELEASE_VERSION\s*=\s*"([^"]+)"`).FindSubmatch(src)
	if match == nil {
		t.Fatal("KOPS_RELEASE_VERSION not found in kops-version.go")
	}

	actual, err := kopsVersion(kopsRoot, []string{"PATH=" + os.Getenv("PATH")})
	if err != nil {
		t.Fatalf("kopsVersion() = %v", err)
	}
	if expected := string(match[1]); actual != expected {
		t.Errorf("kopsVersion() = %q, expected %q", actual, expected)
	}
}

func TestS3URI(t *testing.T) {
	cases := []struct {
		urlPath  string
		expected string
	}{
		{
			urlPath:  "latest-ci.txt",
			expected: "s3://staging/latest-ci.txt",
		},
		{
			urlPath:  "prefix/1.37.0-alpha.1+v1.37.0-alpha.0-12-gabcdef0",
			expected: "s3://staging/prefix/1.37.0-alpha.1 v1.37.0-alpha.0-12-gabcdef0",
		},
		{
			urlPath:  "kubernetes/v1.37.0-alpha.1.123+0123456789ab",
			expected: "s3://staging/kubernetes/v1.37.0-alpha.1.123 0123456789ab",
		},
	}
	for _, tc := range cases {
		if actual := s3URI("staging", tc.urlPath); actual != tc.expected {
			t.Errorf("s3URI(%q) = %q, expected %q", tc.urlPath, actual, tc.expected)
		}
	}
}

// An s3:// stage location is useless without its region, which the URLs of the artifacts include.
// The region must be checked before anything is built.
func TestBuildToS3RequiresRegion(t *testing.T) {
	for _, buildKubernetes := range []bool{false, true} {
		b := &BuildOptions{StageLocation: "s3://staging", BuildKubernetes: buildKubernetes}
		if _, err := b.Build(); err == nil {
			t.Errorf("Build() with BuildKubernetes=%v succeeded, expected an error for an s3:// stage location without a region", buildKubernetes)
		}
	}
}
