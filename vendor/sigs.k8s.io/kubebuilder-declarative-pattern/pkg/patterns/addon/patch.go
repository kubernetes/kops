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

package addon

import (
	"bytes"
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/log"

	addonsv1alpha1 "sigs.k8s.io/kubebuilder-declarative-pattern/pkg/patterns/addon/pkg/apis/v1alpha1"
	"sigs.k8s.io/kubebuilder-declarative-pattern/pkg/patterns/declarative"
	"sigs.k8s.io/kubebuilder-declarative-pattern/pkg/patterns/declarative/pkg/manifest"
)

// ApplyPatches is an ObjectTransform to apply Patches specified on the Addon object to the manifest
// This transform requires the DeclarativeObject to implement addonsv1alpha1.Patchable
func ApplyPatches(ctx context.Context, object declarative.DeclarativeObject, objects *manifest.Objects) error {
	log := log.Log

	p, ok := object.(addonsv1alpha1.Patchable)
	if !ok {
		return fmt.Errorf("provided object (%T) does not implement Patchable type", object)
	}

	var patches []*unstructured.Unstructured

	for _, p := range p.PatchSpec().Patches {
		// Object is nil, Raw  is populated (with json, even when input was yaml)
		r := bytes.NewReader(p.Raw)
		decoder := yaml.NewYAMLOrJSONDecoder(r, 1024)
		patch := &unstructured.Unstructured{}

		if err := decoder.Decode(patch); err != nil {
			return fmt.Errorf("error parsing json into unstructured object: %v", err)
		}
		log.WithValues("patch", patch).V(1).Info("parsed patch")

		patches = append(patches, patch)
	}

	return objects.Patch(patches)
}
