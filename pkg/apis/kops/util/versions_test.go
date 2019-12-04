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

package util

import (
	"testing"
)

func Test_ParseKubernetesVersion(t *testing.T) {
	grid := map[string]string{
		"1.3.7":         "1.3.7",
		"v1.4.0-beta.8": "1.4.0-beta.8",
		"1.5.0":         "1.5.0",
		"https://storage.googleapis.com/kubernetes-release-dev/ci/v1.4.0-alpha.2.677+ea69570f61af8e/": "1.4.0",
	}
	for version, expected := range grid {
		sv, err := ParseKubernetesVersion(version)
		if err != nil {
			t.Errorf("ParseKubernetesVersion error parsing %q: %v", version, err)
		}

		actual := sv.String()
		if actual != expected {
			t.Errorf("version mismatch: %q -> %q but expected %q", version, actual, expected)
		}
	}
}

func Test_IsKubernetesGTEWithPatch(t *testing.T) {
	currentVersion, err := ParseKubernetesVersion("1.6.2")
	if err != nil {
		t.Fatalf("Error parsing version: %v", err)
	}

	grid := map[string]bool{
		"1.5.2": true,
		"1.6.2": true,
		"1.6.5": false,
		"1.7.8": false,
	}

	for v, expected := range grid {
		actual := IsKubernetesGTE(v, *currentVersion)
		if actual != expected {
			t.Errorf("expected %s to be >= than %s", v, currentVersion)
		}
	}
}

func Test_IsKubernetesGTEWithoutPatch(t *testing.T) {
	currentVersion, err := ParseKubernetesVersion("1.6")
	if err != nil {
		t.Fatalf("Error parsing version: %v", err)
	}

	grid := map[string]bool{
		"1.1": true,
		"1.2": true,
		"1.3": true,
		"1.6": true,
		"1.7": false,
	}

	for v, expected := range grid {
		actual := IsKubernetesGTE(v, *currentVersion)
		if actual != expected {
			t.Errorf("expected %s to be >= than %s", v, currentVersion)
		}
	}
}

func Test_IsKubernetesGTEWithPre(t *testing.T) {
	grid := map[string]bool{
		"1.6.1":         true,
		"1.6":           true,
		"1.6.0-alpha.1": true,
		"1.6.0-beta":    true,
		"1.5.9-alpha.1": false,
	}

	for v, expected := range grid {
		currentVersion, err := ParseKubernetesVersion(v)
		if err != nil {
			t.Fatalf("Error parsing version: %v", err)
		}

		actual := IsKubernetesGTE("1.6", *currentVersion)
		if actual != expected {
			t.Errorf("expected %s to be >= than %s", "1.6", currentVersion)
		}
	}
}
