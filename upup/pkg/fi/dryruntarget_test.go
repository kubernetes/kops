/*
Copyright 2017 The Kubernetes Authors.

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

package fi

import (
	"reflect"
	"testing"
)

func Test_tryResourceAsString(t *testing.T) {
	var sr *StringResource
	grid := []struct {
		Resource interface{}
		Expected string
	}{
		{
			Resource: NewStringResource("hello"),
			Expected: "hello",
		},
		{
			Resource: sr,
			Expected: "",
		},
		{
			Resource: nil,
			Expected: "",
		},
	}
	for i, g := range grid {
		v := reflect.ValueOf(g.Resource)
		actual, _ := tryResourceAsString(v)
		if actual != g.Expected {
			t.Errorf("unexpected result from %d.  Expected=%q, got %q", i, g.Expected, actual)
		}
	}
}
