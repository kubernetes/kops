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

package manifest

import (
	"fmt"

	jsonpatch "github.com/evanphx/json-patch"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (objects *Objects) Patch(patches []*unstructured.Unstructured) error {
	log := log.Log

	for i, o := range objects.Items {
		log.WithValues("object", o).Info("applying patches")

		patched, err := apply(o.UnstructuredObject(), patches)
		if err != nil {
			return fmt.Errorf("applying patch to object (%v): %e", o.UnstructuredObject().GetName(), err)
		}

		log.WithValues("patched", patched).V(2).Info("applying patches")

		patchedObject, err := NewObject(patched)
		if err != nil {
			return err
		}
		objects.Items[i] = patchedObject
	}

	log.WithValues("patches count", len(patches)).WithValues("objects count", len(objects.Items)).Info("applied patches")

	return nil
}

func apply(base *unstructured.Unstructured, patches []*unstructured.Unstructured) (*unstructured.Unstructured, error) {
	merged := base.DeepCopy()

	for _, p := range patches {
		if p.GetObjectKind().GroupVersionKind() != base.GetObjectKind().GroupVersionKind() {
			continue
		}
		if p.GetName() != base.GetName() || p.GetNamespace() != base.GetNamespace() {
			continue
		}

		versionedObj, err := scheme.Scheme.New(base.GetObjectKind().GroupVersionKind())
		switch {
		case runtime.IsNotRegisteredError(err):
			// Use JSON merge patch to handle types w/o schema
			baseBytes, err := merged.MarshalJSON()
			if err != nil {
				return nil, err
			}
			patchBytes, err := p.MarshalJSON()
			if err != nil {
				return nil, err
			}
			mergedBytes, err := jsonpatch.MergePatch(baseBytes, patchBytes)
			if err != nil {
				return nil, err
			}
			err = merged.UnmarshalJSON(mergedBytes)
			if err != nil {
				return nil, err
			}
		case err != nil:
			return nil, err
		default:
			// Use Strategic-Merge-Patch to handle types w/ schema
			// TODO: Change this to use the new Merge package.
			// Store the name of the base object, because this name may have been munged.
			// Apply this name to the patched object.
			lookupPatchMeta, err := strategicpatch.NewPatchMetaFromStruct(versionedObj)
			if err != nil {
				return nil, err
			}
			merged.Object, err = strategicpatch.StrategicMergeMapPatchUsingLookupPatchMeta(
				merged.Object,
				p.Object,
				lookupPatchMeta)
			if err != nil {
				return nil, err
			}
		}
	}

	return merged, nil
}
