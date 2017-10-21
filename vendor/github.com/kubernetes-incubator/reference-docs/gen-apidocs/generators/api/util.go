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

import (
	"fmt"
	"strings"

	"errors"
	"github.com/go-openapi/spec"
)

func GetGroupVersionKind() {

}

// GetDefinitionVersionKind returns the api version and kind for the spec.  This is the primary key of a Definition.
func GetDefinitionVersionKind(s spec.Schema) (string, string, string) {
	// Get the reference for complex types
	if IsDefinition(s) {
		s := fmt.Sprintf("%s", s.SchemaProps.Ref.GetPointer())
		s = strings.Replace(s, "/definitions/", "", -1)
		name := strings.Split(s, ".")

		var group, version, kind string
		if name[len(name)-3] == "api" {
			// e.g. "io.k8s.apimachinery.pkg.api.resource.Quantity"
			group = "core"
			version = name[len(name)-2]
			kind = name[len(name)-1]
		} else if name[len(name)-4] == "api" {
			// e.g. "io.k8s.api.core.v1.Pod"
			group = name[len(name)-3]
			version = name[len(name)-2]
			kind = name[len(name)-1]
		} else if name[len(name)-4] == "apis" {
			// e.g. "io.k8s.apimachinery.pkg.apis.meta.v1.Status"
			group = name[len(name)-3]
			version = name[len(name)-2]
			kind = name[len(name)-1]
		} else if name[len(name)-3] == "util" || name[len(name)-3] == "pkg" {
			// e.g. io.k8s.apimachinery.pkg.util.intstr.IntOrString
			// e.g. io.k8s.apimachinery.pkg.runtime.RawExtension
			return "", "", ""
		} else {
			panic(errors.New(fmt.Sprintf("Could not locate group for %s", name)))
		}
		return group, version, kind
	}
	// Recurse if type is array
	if IsArray(s) {
		return GetDefinitionVersionKind(*s.Items.Schema)
	}
	return "", "", ""
}

// GetTypeName returns the display name of a Schema.  This is the api kind for definitions and the type for
// primitive types.  Arrays of objects have "array" appended.
func GetTypeName(s spec.Schema) string {
	// Get the reference for complex types
	if IsDefinition(s) {
		_, _, name := GetDefinitionVersionKind(s)
		return name
	}
	// Recurse if type is array
	if IsArray(s) {
		return fmt.Sprintf("%s array", GetTypeName(*s.Items.Schema))
	}
	// Get the value for primitive types
	if len(s.Type) > 0 {
		return fmt.Sprintf("%s", s.Type[0])
	}
	panic(fmt.Errorf("No type found for object %v", s))
}

// IsArray returns true if the type is an array type.
func IsArray(s spec.Schema) bool {
	//if s == nil {
	//	return false
	//}
	return len(s.Type) > 0 && s.Type[0] == "array"
}

// IsDefinition returns true if Schema is a complex type that should have a Definition.
func IsDefinition(s spec.Schema) bool {
	return len(s.SchemaProps.Ref.GetPointer().String()) > 0
}
