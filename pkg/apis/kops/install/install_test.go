/*
Copyright 2025 The Kubernetes Authors.

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

package install

import (
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestTypesExistInAllVersions(t *testing.T) {
	scheme := runtime.NewScheme()
	Install(scheme)

	group := "kops.k8s.io"
	internalVersion := schema.GroupVersion{Group: group, Version: runtime.APIVersionInternal}
	internalKindToGoType := scheme.KnownTypes(internalVersion)

	internalKinds := make(map[string]bool)
	for kind := range internalKindToGoType {
		internalKinds[kind] = true
	}

	for _, version := range []string{"v1alpha2", "v1alpha3"} {
		versionKinds := scheme.KnownTypes(schema.GroupVersion{Group: group, Version: version})
		for kind := range versionKinds {
			// Ignore ListOptions, DeleteOptions, etc.
			if strings.HasSuffix(kind, "Options") {
				continue
			}
			if !internalKinds[kind] {
				t.Errorf("version %s has kind %s, not found in internal API", version, kind)
			}
		}
	}
}
