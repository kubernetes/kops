/*
Copyright 2024 The Kubernetes Authors.

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

package kops

import "testing"

func TestClusterFieldHumanPath(t *testing.T) {
	grid := []struct {
		In       string
		Expected string
	}{
		{In: "spec.api.publicName", Expected: "spec.masterPublicName"},
		{In: "spec.api", Expected: "spec.api"},
	}

	for _, g := range grid {
		actual := HumanPathForClusterField(g.In)
		if actual != g.Expected {
			t.Errorf("HumanPathForClusterField(%q) = %q; expected %q", g.In, actual, g.Expected)
		}
	}
}
