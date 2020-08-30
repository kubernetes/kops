/*
Copyright 2020 The Kubernetes Authors.

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

package reflectutils

import (
	"testing"
)

func TestParseFieldPath(t *testing.T) {
	grid := []struct{ V string }{
		{"Spec.Containers.Image"},
		{"Spec.Containers[0].Image"},
		{"Spec.Containers[*].Image"},
	}

	for _, g := range grid {
		s := g.V
		t.Run("test "+s, func(t *testing.T) {
			fp, err := ParseFieldPath(s)
			if err != nil {
				t.Fatalf("error parsing field path %q: %v", s, err)
			}

			s2 := fp.String()
			if s != s2 {
				t.Fatalf("field path %q did not round-trip, string value was %q", s, s2)
			}
		})
	}
}
