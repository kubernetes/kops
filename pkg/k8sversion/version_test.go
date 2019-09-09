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

package k8sversion

import "testing"

func TestParse(t *testing.T) {
	grid := []struct {
		Input    string
		Expected string
	}{
		{Input: "1.1.0", Expected: "1.1.0"},
		{Input: "1.2.0", Expected: "1.2.0"},
		{Input: "1.3.0", Expected: "1.3.0"},
		{Input: "1.4.0", Expected: "1.4.0"},
		{Input: "1.5.0", Expected: "1.5.0"},
		{Input: "1.6.0", Expected: "1.6.0"},
		{Input: "1.7.0", Expected: "1.7.0"},
		{Input: "1.8.0", Expected: "1.8.0"},
		{Input: "1.9.0", Expected: "1.9.0"},
		{Input: "1.10.0", Expected: "1.10.0"},
		{Input: "v1.1.0-alpha1", Expected: "1.1.0-alpha1"},
		{Input: "1.11.0", Expected: "1.11.0"},
		{Input: "1.12.0", Expected: "1.12.0"},
		{Input: "1.13.0", Expected: "1.13.0"},
		{Input: "1.14.0", Expected: "1.14.0"},
		{Input: "1.15.0", Expected: "1.15.0"},
		{Input: "1.16.0", Expected: "1.16.0"},
		{Input: "https://example.com/v1.8.0-downloads", Expected: "1.8.0"},
	}

	for _, g := range grid {
		actual, err := Parse(g.Input)
		if err != nil {
			t.Errorf("error parsing %q: %v", g.Input, err)
			continue
		}
		if actual.String() != g.Expected {
			t.Errorf("unexpected result parsing %q: actual=%q expected=%q", g.Input, actual.String(), g.Expected)
			continue
		}
	}
}
