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

package kops

import (
	"testing"

	"github.com/blang/semver"
)

// Test_SemverOrdering is a test of semver ordering, but highlights the case that trips everyone one:
// 1.6.0-alpha.1 < 1.6.0, so you can't use >= 1.6.0 as the test for "1.6 series"
func Test_SemverOrdering(t *testing.T) {
	v160 := semver.MustParse("1.6.0")
	v160alpha1 := semver.MustParse("1.6.0-alpha.1")

	if !v160.GT(v160alpha1) {
		t.Errorf("semver 1.6.0-alpha.1 < 1.6.0 (as much as we would like this not to be true")
	}
}
