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

package jsonutils

import (
	"encoding/json"
	"fmt"
	"sort"
)

// Transformer is used to transform JSON values
type Transformer struct {
	stringTransforms []func(path string, value string) (string, error)
	objectTransforms []func(path string, value map[string]any) error
	sliceTransforms  []func(path string, value []any) ([]any, error)
}

// NewTransformer is the constructor for a Transformer
func NewTransformer() *Transformer {
	return &Transformer{}
}

// AddStringTransform adds a function that will be called for each string value in the JSON tree
func (t *Transformer) AddStringTransform(fn func(path string, value string) (string, error)) {
	t.stringTransforms = append(t.stringTransforms, fn)
}

// AddObjectTransform adds a function that will be called for each object in the JSON tree
func (t *Transformer) AddObjectTransform(fn func(path string, value map[string]any) error) {
	t.objectTransforms = append(t.objectTransforms, fn)
}

// AddSliceTransform adds a function that will be called for each slice in the JSON tree
func (t *Transformer) AddSliceTransform(fn func(path string, value []any) ([]any, error)) {
	t.sliceTransforms = append(t.sliceTransforms, fn)
}

// Transform applies the transformations to the JSON tree
func (o *Transformer) Transform(v map[string]any) error {
	_, err := o.visitAny(v, "")
	return err
}

// visitAny is a helper function that visits any value in the JSON tree
func (o *Transformer) visitAny(v any, path string) (any, error) {
	if v == nil {
		return v, nil
	}
	switch v := v.(type) {
	case map[string]any:
		if err := o.visitMap(v, path); err != nil {
			return nil, err
		}
		return v, nil
	case []any:
		return o.visitSlice(v, path)
	case int64, float64, bool:
		return o.visitPrimitive(v, path)
	case string:
		return o.visitString(v, path)
	default:
		return nil, fmt.Errorf("unhandled type at path %q: %T", path, v)
	}
}

func (o *Transformer) visitMap(m map[string]any, path string) error {
	for _, fn := range o.objectTransforms {
		if err := fn(path, m); err != nil {
			return err
		}
	}

	for k, v := range m {
		childPath := path + "." + k

		v2, err := o.visitAny(v, childPath)
		if err != nil {
			return err
		}
		m[k] = v2
	}

	return nil
}

// visitSlice is a helper function that visits a slice in the JSON tree
func (o *Transformer) visitSlice(s []any, path string) (any, error) {
	for _, fn := range o.sliceTransforms {
		var err error
		s, err = fn(path+"[]", s)
		if err != nil {
			return nil, err
		}
	}

	for i, v := range s {
		v2, err := o.visitAny(v, path+"[]")
		if err != nil {
			return nil, err
		}
		s[i] = v2
	}

	return s, nil
}

// SortSlice sorts a slice of any values, ordered by their JSON representations.
// This is not very efficient, but is convenient for small slice where we don't know their types.
func SortSlice(s []any) ([]any, error) {
	type entry struct {
		o       any
		sortKey string
	}

	var entries []entry
	for i := range s {
		j, err := json.Marshal(s[i])
		if err != nil {
			return nil, fmt.Errorf("error converting to json: %w", err)
		}
		entries = append(entries, entry{o: s[i], sortKey: string(j)})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].sortKey < entries[j].sortKey
	})

	out := make([]any, 0, len(s))
	for i := range s {
		out = append(out, entries[i].o)
	}

	return out, nil
}

// visitPrimitive is a helper function that visits a primitive value in the JSON tree
func (o *Transformer) visitPrimitive(v any, _ string) (any, error) {
	return v, nil
}

// visitString is a helper function that visits a string value in the JSON tree
func (o *Transformer) visitString(v string, path string) (string, error) {
	for _, fn := range o.stringTransforms {
		var err error
		v, err = fn(path, v)
		if err != nil {
			return "", err
		}
	}
	return v, nil
}
