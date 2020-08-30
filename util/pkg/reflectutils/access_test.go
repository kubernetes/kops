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
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

type fakeEnum string

type fakeObject struct {
	Spec   fakeObjectSpec   `json:"spec"`
	Status fakeObjectStatus `json:"status"`
}

func (o *fakeObject) String() string {
	b, err := json.Marshal(o)
	if err != nil {
		return fmt.Sprintf("(error:%v)", err)
	}
	return string(b)
}

type fakeObjectSpec struct {
	Containers []fakeObjectContainers `json:"containers"`
}

type fakeObjectContainers struct {
	Image     string               `json:"image"`
	Policy    *fakeObjectPolicy    `json:"policy"`
	Resources *fakeObjectResources `json:"resources"`

	IntPointer   *int32 `json:"intPointer"`
	Int32Pointer *int32 `json:"int32Pointer"`
	Int64Pointer *int64 `json:"int64Pointer"`

	Int   int32 `json:"int"`
	Int32 int32 `json:"int32"`
	Int64 int64 `json:"int64"`

	Enum      fakeEnum   `json:"enum"`
	EnumSlice []fakeEnum `json:"enumSlice"`
}

type fakeObjectPolicy struct {
	Name  string `json:"name"`
	Allow bool   `json:"allow"`
}

type fakeObjectResources struct {
	Limits   map[string]int `json:"limits"`
	Requests map[string]int `json:"requests"`
}

type fakeObjectStatus struct {
}

func toJSON(s string) []byte {
	s = strings.ReplaceAll(s, "'", "\"")
	return []byte(s)
}

func TestSet(t *testing.T) {
	grid := []struct {
		Name     string
		Input    string
		Expected string
		Path     string
		Value    string
	}{
		{
			Name:     "simple setting",
			Input:    "{ 'spec': { 'containers': [ {} ] } }",
			Expected: "{ 'spec': { 'containers': [ { 'image': 'hello-world' } ] } }",
			Path:     "spec.containers[0].image",
			Value:    "hello-world",
		},
		{
			Name:     "setting with wildcard",
			Input:    "{ 'spec': { 'containers': [ {} ] } }",
			Expected: "{ 'spec': { 'containers': [ { 'image': 'hello-world' } ] } }",
			Path:     "spec.containers[*].image",
			Value:    "hello-world",
		},
		{
			Name:     "creating missing objects",
			Input:    "{ 'spec': { 'containers': [ {} ] } }",
			Expected: "{ 'spec': { 'containers': [ { 'policy': { 'name': 'allowed' } } ] } }",
			Path:     "spec.containers[0].policy.name",
			Value:    "allowed",
		},
		{
			Name:     "set int",
			Input:    "{ 'spec': { 'containers': [ {} ] } }",
			Expected: "{ 'spec': { 'containers': [ { 'int': 123 } ] } }",
			Path:     "spec.containers[0].int",
			Value:    "123",
		},
		{
			Name:     "set int32",
			Input:    "{ 'spec': { 'containers': [ {} ] } }",
			Expected: "{ 'spec': { 'containers': [ { 'int32': 123 } ] } }",
			Path:     "spec.containers[0].int32",
			Value:    "123",
		},
		{
			Name:     "set int64",
			Input:    "{ 'spec': { 'containers': [ {} ] } }",
			Expected: "{ 'spec': { 'containers': [ { 'int64': 123 } ] } }",
			Path:     "spec.containers[0].int64",
			Value:    "123",
		},
		{
			Name:     "set int pointer",
			Input:    "{ 'spec': { 'containers': [ {} ] } }",
			Expected: "{ 'spec': { 'containers': [ { 'intPointer': 123 } ] } }",
			Path:     "spec.containers[0].intPointer",
			Value:    "123",
		},
		{
			Name:     "set int32 pointer",
			Input:    "{ 'spec': { 'containers': [ {} ] } }",
			Expected: "{ 'spec': { 'containers': [ { 'int32Pointer': 123 } ] } }",
			Path:     "spec.containers[0].int32Pointer",
			Value:    "123",
		},
		{
			Name:     "set int64 pointer",
			Input:    "{ 'spec': { 'containers': [ {} ] } }",
			Expected: "{ 'spec': { 'containers': [ { 'int64Pointer': 123 } ] } }",
			Path:     "spec.containers[0].int64Pointer",
			Value:    "123",
		},
		{
			Name:     "set enum",
			Input:    "{ 'spec': { 'containers': [ {} ] } }",
			Expected: "{ 'spec': { 'containers': [ { 'enum': 'ABC' } ] } }",
			Path:     "spec.containers[0].enum",
			Value:    "ABC",
		},
		{
			Name:     "set enum slice",
			Input:    "{ 'spec': { 'containers': [ {} ] } }",
			Expected: "{ 'spec': { 'containers': [ { 'enumSlice': [ 'ABC', 'DEF' ] } ] } }",
			Path:     "spec.containers[0].enumSlice",
			Value:    "ABC,DEF",
		},
		{
			Name:     "unset int pointer",
			Input:    "{ 'spec': { 'containers': [ { 'int64Pointer': 123 } ] } }",
			Expected: "{ 'spec': { 'containers': [ { } ] } }",
			Path:     "spec.containers[0].int64Pointer",
			Value:    "",
		},
		{
			Name:     "unset struct pointer",
			Input:    "{ 'spec': { 'containers': [ { 'resources': { 'limits': { 'cpu': 1 } } } ] } }",
			Expected: "{ 'spec': { 'containers': [ { } ] } }",
			Path:     "spec.containers[0].resources",
			Value:    "",
		},

		// Not sure if we should do this...
		// {
		// 	Name:     "creating missing array elements",
		// 	Input:    "{}",
		// 	Expected: "{ 'spec': { 'containers': [ { 'policy': { 'name': 'allowed' } } ] } }",
		// 	Path:     "spec.containers[0].policy.name",
		// 	Value:    "allowed",
		// },
	}

	for _, g := range grid {
		g := g

		t.Run(g.Name, func(t *testing.T) {

			c := &fakeObject{}
			if err := json.Unmarshal(toJSON(g.Input), c); err != nil {
				t.Fatalf("failed to unmarshal input: %v", err)

			}

			if err := SetString(c, g.Path, g.Value); err != nil {
				t.Fatalf("error from SetString: %v", err)
			}

			// Changed in-place
			actual := c

			expected := &fakeObject{}
			if err := json.Unmarshal(toJSON(g.Expected), expected); err != nil {
				t.Fatalf("failed to unmarshal expected: %v", err)
			}

			if !reflect.DeepEqual(c, expected) {
				t.Fatalf("comparison failed; expected %+v, was %+v", expected, actual)
			}
		})
	}
}
