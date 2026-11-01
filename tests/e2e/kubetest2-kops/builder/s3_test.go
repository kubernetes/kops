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
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
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

func TestObjectStoreURIs(t *testing.T) {
	t.Setenv("S3_ACCESS_KEY_ID", "access")
	t.Setenv("S3_SECRET_ACCESS_KEY", "secret")

	s3, err := (&BuildOptions{StageLocation: "s3://staging/prefix", StageRegion: "us-east-2"}).s3Store()
	if err != nil {
		t.Fatalf("s3Store() = %v", err)
	}
	spaces, err := (&BuildOptions{StageLocation: "s3://kops-e2e-staging/prefix/", StageEndpoint: "https://sfo3.digitaloceanspaces.com"}).s3Store()
	if err != nil {
		t.Fatalf("s3Store() with an endpoint = %v", err)
	}

	cases := []struct {
		name        string
		store       *objectStore
		urlPath     string
		expectedURI string
		expectedURL string
	}{
		{
			name:        "s3 marker",
			store:       s3,
			urlPath:     "prefix/latest-ci.txt",
			expectedURI: "s3://staging/prefix/latest-ci.txt",
			expectedURL: "https://staging.s3.us-east-2.amazonaws.com/prefix/latest-ci.txt",
		},
		{
			name:        "s3 kops version",
			store:       s3,
			urlPath:     "prefix/1.37.0-alpha.1+v1.37.0-alpha.0-12-gabcdef0",
			expectedURI: "s3://staging/prefix/1.37.0-alpha.1 v1.37.0-alpha.0-12-gabcdef0",
			expectedURL: "https://staging.s3.us-east-2.amazonaws.com/prefix/1.37.0-alpha.1+v1.37.0-alpha.0-12-gabcdef0",
		},
		{
			name:        "spaces kops version keeps the plus",
			store:       spaces,
			urlPath:     "prefix/1.37.0-alpha.1+v1.37.0-alpha.0-12-gabcdef0",
			expectedURI: "s3://kops-e2e-staging/prefix/1.37.0-alpha.1+v1.37.0-alpha.0-12-gabcdef0",
			expectedURL: "https://kops-e2e-staging.sfo3.digitaloceanspaces.com/prefix/1.37.0-alpha.1+v1.37.0-alpha.0-12-gabcdef0",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if actual := tc.store.uri(tc.urlPath); actual != tc.expectedURI {
				t.Errorf("uri(%q) = %q, expected %q", tc.urlPath, actual, tc.expectedURI)
			}
			actual, err := tc.store.publicURL(context.Background(), tc.urlPath)
			if err != nil {
				t.Fatalf("publicURL(%q) = %v", tc.urlPath, err)
			}
			if actual != tc.expectedURL {
				t.Errorf("publicURL(%q) = %q, expected %q", tc.urlPath, actual, tc.expectedURL)
			}
		})
	}

	if spaces.prefix != "prefix" {
		t.Errorf("spaces prefix = %q, expected %q", spaces.prefix, "prefix")
	}
}

// Without the keys of the endpoint, the upload would fail only after kops is built.
func TestBuildToEndpointRequiresKeys(t *testing.T) {
	t.Setenv("S3_ACCESS_KEY_ID", "")
	t.Setenv("S3_SECRET_ACCESS_KEY", "")

	b := &BuildOptions{StageLocation: "s3://kops-e2e-staging", StageEndpoint: "https://sfo3.digitaloceanspaces.com"}
	if _, err := b.Build(); err == nil || !strings.Contains(err.Error(), "S3_ACCESS_KEY_ID") {
		t.Errorf("Build() = %v, expected an error about the missing keys", err)
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
