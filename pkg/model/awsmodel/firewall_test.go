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

package awsmodel

import (
	"testing"
)

func TestJoinSuffixes(t *testing.T) {
	grid := []struct {
		src      SecurityGroupInfo
		dest     SecurityGroupInfo
		expected string
	}{
		{
			src:      SecurityGroupInfo{Suffix: ""},
			dest:     SecurityGroupInfo{Suffix: ""},
			expected: "",
		},
		{
			src:      SecurityGroupInfo{Suffix: "srcSuffix"},
			dest:     SecurityGroupInfo{Suffix: ""},
			expected: "srcSuffix-default",
		},
		{
			src:      SecurityGroupInfo{Suffix: ""},
			dest:     SecurityGroupInfo{Suffix: "destSuffix"},
			expected: "-defaultdestSuffix",
		},
	}

	for _, g := range grid {
		actual := JoinSuffixes(g.src, g.dest)
		if actual != g.expected {
			t.Errorf("unexpected result.  expected %q, got %q", g.expected, actual)
		}
	}
}
