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

package fi

import (
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/util/sets"
)

func TestBuildChangesMaps(t *testing.T) {
	type task struct {
		Values map[string]*string
	}
	for _, tt := range []struct {
		name       string
		actual     map[string]*string
		expected   map[string]*string
		wantChange bool
	}{
		{name: "nil maps"},
		{name: "empty maps", actual: map[string]*string{}, expected: map[string]*string{}},
		{name: "nil to empty", expected: map[string]*string{}},
		{name: "empty to nil", actual: map[string]*string{}},
		{name: "equal values", actual: map[string]*string{"a": new("x")}, expected: map[string]*string{"a": new("x")}},
		{name: "equal nil values", actual: map[string]*string{"a": nil}, expected: map[string]*string{"a": nil}},
		{name: "replace key with nil value", actual: map[string]*string{"a": nil}, expected: map[string]*string{"b": nil}, wantChange: true},
		{name: "replace value", actual: map[string]*string{"a": new("x")}, expected: map[string]*string{"a": new("y")}, wantChange: true},
		{name: "nil to value", actual: map[string]*string{"a": nil}, expected: map[string]*string{"a": new("x")}, wantChange: true},
		{name: "value to nil", actual: map[string]*string{"a": new("x")}, expected: map[string]*string{"a": nil}, wantChange: true},
		{name: "add", actual: map[string]*string{}, expected: map[string]*string{"a": new("x")}, wantChange: true},
		{name: "add to nil", expected: map[string]*string{"a": new("x")}, wantChange: true},
		{name: "clear to empty", actual: map[string]*string{"a": new("x")}, expected: map[string]*string{}, wantChange: true},
		{name: "clear to nil", actual: map[string]*string{"a": new("x")}, wantChange: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			actual := &task{Values: tt.actual}
			expected := &task{Values: tt.expected}
			changes := &task{}
			if got := BuildChanges(actual, expected, changes); got != tt.wantChange {
				t.Fatalf("BuildChanges = %v, want %v", got, tt.wantChange)
			}
			want := &task{}
			if tt.wantChange {
				want = expected
			}
			if !reflect.DeepEqual(changes, want) {
				t.Errorf("changes = %#v, want %#v", changes, want)
			}
		})
	}
}

func TestBuildChangesSets(t *testing.T) {
	type task struct {
		Members sets.Set[string]
	}
	for _, tt := range []struct {
		name       string
		actual     sets.Set[string]
		expected   sets.Set[string]
		wantChange bool
	}{
		{name: "nil sets"},
		{name: "empty sets", actual: sets.New[string](), expected: sets.New[string]()},
		{name: "nil to empty", expected: sets.New[string]()},
		{name: "empty to nil", actual: sets.New[string]()},
		{name: "reordered", actual: sets.New("a", "b"), expected: sets.New("b", "a")},
		{name: "replace", actual: sets.New("a"), expected: sets.New("b"), wantChange: true},
		{name: "add", actual: sets.New("a"), expected: sets.New("a", "b"), wantChange: true},
		{name: "remove", actual: sets.New("a", "b"), expected: sets.New("a"), wantChange: true},
		{name: "clear to empty", actual: sets.New("a"), expected: sets.New[string](), wantChange: true},
		{name: "clear to nil", actual: sets.New("a"), wantChange: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			actual := &task{Members: tt.actual}
			expected := &task{Members: tt.expected}
			changes := &task{}
			if got := BuildChanges(actual, expected, changes); got != tt.wantChange {
				t.Fatalf("BuildChanges = %v, want %v", got, tt.wantChange)
			}
			want := &task{}
			if tt.wantChange {
				want = expected
			}
			if !reflect.DeepEqual(changes, want) {
				t.Errorf("changes = %#v, want %#v", changes, want)
			}
		})
	}
}

func TestBuildChangesMapResources(t *testing.T) {
	type task struct {
		Values map[string]Resource
	}
	for _, tt := range []struct {
		name       string
		expected   Resource
		wantChange bool
	}{
		{name: "same contents", expected: NewBytesResource([]byte("contents"))},
		{name: "different contents", expected: NewBytesResource([]byte("different")), wantChange: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			// Different resource types must still compare by their contents.
			actual := &task{Values: map[string]Resource{"file": NewStringResource("contents")}}
			expected := &task{Values: map[string]Resource{"file": tt.expected}}
			changes := &task{}
			if got := BuildChanges(actual, expected, changes); got != tt.wantChange {
				t.Fatalf("BuildChanges = %v, want %v", got, tt.wantChange)
			}
			want := &task{}
			if tt.wantChange {
				want = expected
			}
			if !reflect.DeepEqual(changes, want) {
				t.Errorf("changes = %#v, want %#v", changes, want)
			}
		})
	}
}
