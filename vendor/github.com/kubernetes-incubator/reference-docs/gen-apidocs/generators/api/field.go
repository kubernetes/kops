/*
Copyright 2016 The Kubernetes Authors.

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

package api

import "strings"

type Fields []*Field

func (a Fields) Len() int           { return len(a) }
func (a Fields) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Fields) Less(i, j int) bool { return a[i].Name < a[j].Name }

type Field struct {
	Name        string
	Type        string
	Description string
	DescriptionWithEntities string
	// Optional Definition for complex types
	Definition *Definition

	// Patch semantics
	PatchStrategy string
	PatchMergeKey string
}

func (f Field) Link() string {
	if f.Definition != nil {
		return strings.Replace(f.Type, f.Definition.Name, f.Definition.MdLink(), -1)
	} else {
		return f.Type
	}
}
