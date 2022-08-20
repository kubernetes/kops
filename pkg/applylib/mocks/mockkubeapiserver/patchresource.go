/*
Copyright 2022 The Kubernetes Authors.

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

package mockkubeapiserver

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

// patchResource is a request to patch a single resource
type patchResource struct {
	resourceRequestBase
}

// Run serves the http request
func (req *patchResource) Run(s *MockKubeAPIServer) error {
	gr := schema.GroupResource{Group: req.Group, Resource: req.Resource}

	id := types.NamespacedName{Namespace: req.Namespace, Name: req.Name}
	var existing *unstructured.Unstructured
	objects := s.objects[gr]
	if objects != nil {
		existing = objects.Objects[id]
	}

	bodyBytes, err := ioutil.ReadAll(req.r.Body)
	if err != nil {
		return err
	}

	body := &unstructured.Unstructured{}
	if err := body.UnmarshalJSON(bodyBytes); err != nil {
		return fmt.Errorf("failed to parse payload: %w", err)
	}

	// TODO: We need to implement patch properly
	klog.Infof("patch request %#v", string(bodyBytes))

	if existing == nil {
		// TODO: Only if server-side-apply
		if objects == nil {
			objects = &objectList{
				GroupResource: gr,
				Objects:       make(map[types.NamespacedName]*unstructured.Unstructured),
			}
			s.objects[gr] = objects
		}

		if req.SubResource != "" {
			// TODO: Is this correct for server-side-apply?
			return req.writeErrorResponse(http.StatusNotFound)
		}

		patched := body
		objects.Objects[id] = patched
		s.objectChanged(patched)
		return req.writeResponse(patched)
	}

	if req.SubResource == "" {
		if err := applyPatch(existing.Object, body.Object); err != nil {
			klog.Warningf("error from patch: %v", err)
			return err
		}
	} else {
		// TODO: We need to implement put properly
		return fmt.Errorf("unknown subresource %q", req.SubResource)
	}
	objects.Objects[id] = existing
	s.objectChanged(existing)
	return req.writeResponse(existing)
}

func applyPatch(existing, patch map[string]interface{}) error {
	for k, patchValue := range patch {
		existingValue := existing[k]
		switch patchValue := patchValue.(type) {
		case string, int64:
			existing[k] = patchValue
		case map[string]interface{}:
			if existingValue == nil {
				existing[k] = patchValue
			} else {
				existingMap, ok := existingValue.(map[string]interface{})
				if !ok {
					return fmt.Errorf("unexpected type mismatch, expected map got %T", existingValue)
				}
				if err := applyPatch(existingMap, patchValue); err != nil {
					return err
				}
			}
		default:
			return fmt.Errorf("type %T not handled in patch", patchValue)
		}
	}
	return nil
}
