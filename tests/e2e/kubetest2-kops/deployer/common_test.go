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

package deployer

import "testing"

func TestMaybeGSURL(t *testing.T) {
	cases := []struct {
		name          string
		cloudProvider string
		baseURL       string
		expected      string
	}{
		{
			name:          "gce staged artifacts from a CI build",
			cloudProvider: "gce",
			baseURL:       "https://storage.googleapis.com/k8s-staging-kops/kops/releases/1.37.0-alpha.2+v1.37.0-alpha.1-42-gdeadbeef01",
			expected:      "gs://k8s-staging-kops/kops/releases/1.37.0-alpha.2+v1.37.0-alpha.1-42-gdeadbeef01",
		},
		{
			name:          "gce staged artifacts from a release branch without gs support are left alone",
			cloudProvider: "gce",
			baseURL:       "https://storage.googleapis.com/k8s-staging-kops/kops/releases/1.35.3+v1.35.1-39-gc89e13599b",
			expected:      "https://storage.googleapis.com/k8s-staging-kops/kops/releases/1.35.3+v1.35.1-39-gc89e13599b",
		},
		{
			name:          "gce with no parseable version is left alone",
			cloudProvider: "gce",
			baseURL:       "https://storage.googleapis.com/k8s-staging-kops/kops/latest/",
			expected:      "https://storage.googleapis.com/k8s-staging-kops/kops/latest/",
		},
		{
			name:          "aws staged artifacts are left alone",
			cloudProvider: "aws",
			baseURL:       "https://storage.googleapis.com/k8s-staging-kops/kops/releases/1.37.0-alpha.2+v1.37.0-alpha.1-42-gdeadbeef01",
			expected:      "https://storage.googleapis.com/k8s-staging-kops/kops/releases/1.37.0-alpha.2+v1.37.0-alpha.1-42-gdeadbeef01",
		},
		{
			name:          "gce with a non-GCS url",
			cloudProvider: "gce",
			baseURL:       "https://artifacts.k8s.io/binaries/kops/1.37.0/",
			expected:      "https://artifacts.k8s.io/binaries/kops/1.37.0/",
		},
		{
			name:          "gce with an already converted url",
			cloudProvider: "gce",
			baseURL:       "gs://k8s-staging-kops/kops/releases/1.37.0-alpha.2+v1.37.0-alpha.1-42-gdeadbeef01",
			expected:      "gs://k8s-staging-kops/kops/releases/1.37.0-alpha.2+v1.37.0-alpha.1-42-gdeadbeef01",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := &deployer{CloudProvider: tc.cloudProvider}
			if actual := d.maybeGSURL(tc.baseURL); actual != tc.expected {
				t.Errorf("maybeGSURL(%q) = %q, expected %q", tc.baseURL, actual, tc.expected)
			}
		})
	}
}
