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
	"reflect"
	"testing"
)

func TestPublishConfig(t *testing.T) {
	for _, tc := range []struct {
		name     string
		location string
		target   string
		env      []string
		wantErr  bool
	}{
		{
			name:     "GCS",
			location: "gs://bucket/path/",
			target:   "gcs-publish-ci",
			env:      []string{"GCS_LOCATION=gs://bucket/path/"},
		},
		{
			name:     "S3",
			location: "s3://bucket/path/",
			target:   "s3-publish-ci",
			env: []string{
				"UPLOAD_DEST=s3://bucket/path/",
				"UPLOAD_ARGS=",
			},
		},
		{
			name:     "unsupported",
			location: "https://example.com/path/",
			wantErr:  true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			target, env, err := publishConfig(tc.location)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected an error")
				}
				return
			}
			if err != nil {
				t.Fatalf("building publish config: %v", err)
			}
			if target != tc.target {
				t.Errorf("expected target %q, got %q", tc.target, target)
			}
			if !reflect.DeepEqual(env, tc.env) {
				t.Errorf("expected env %v, got %v", tc.env, env)
			}
		})
	}
}

func TestPublishedBaseURL(t *testing.T) {
	for _, tc := range []struct {
		value    string
		expected string
	}{
		{
			value:    "https://storage.googleapis.com/bucket/path/version",
			expected: "https://storage.googleapis.com/bucket/path/version",
		},
		{
			value:    "s3://bucket/path/version",
			expected: "s3://bucket/path/version",
		},
		{
			// The publish target joins a location that already ends in a slash.
			value:    "https://storage.googleapis.com/bucket/path//version",
			expected: "https://storage.googleapis.com/bucket/path/version",
		},
	} {
		actual, err := publishedBaseURL(tc.value)
		if err != nil {
			t.Fatalf("converting %q: %v", tc.value, err)
		}
		if actual != tc.expected {
			t.Errorf("expected %q, got %q", tc.expected, actual)
		}
	}
}
