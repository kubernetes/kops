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

package main

import (
	"reflect"
	"testing"
)

func TestRemoveSharedPrefix(t *testing.T) {
	grid := []struct {
		Input  []string
		Output []string
	}{
		{
			Input:  []string{"a", "b", "c"},
			Output: []string{"a", "b", "c"},
		},
		{
			Input:  []string{"za", "zb", "zc"},
			Output: []string{"a", "b", "c"},
		},
		{
			Input:  []string{"zza", "zzb", "zzc"},
			Output: []string{"a", "b", "c"},
		},
		{
			Input:  []string{"zza", "zzb", ""},
			Output: []string{"zza", "zzb", ""},
		},
		{
			Input:  []string{"us-test-1a-1", "us-test-1b-1", "us-test-1a-2", "us-test-1b-2", "us-test-1a-3"},
			Output: []string{"a-1", "b-1", "a-2", "b-2", "a-3"},
		},
	}
	for _, g := range grid {
		actual := trimCommonPrefix(g.Input)
		if !reflect.DeepEqual(actual, g.Output) {
			t.Errorf("unexpected result from %q.  actual=%v, expected=%v", g.Input, actual, g.Output)
		}
	}
}
