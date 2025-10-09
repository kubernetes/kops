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

package builders

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func makeRef(u *unstructured.Unstructured) map[string]any {
	apiVersion, kind := u.GroupVersionKind().ToAPIVersionAndKind()
	ref := map[string]any{
		"name":       u.GetName(),
		"apiVersion": apiVersion,
		"kind":       kind,
	}
	return ref
}

// func setOwnerRef(u *unstructured.Unstructured, owner client.Object) {
// 	apiVersion, kind := owner.GetObjectKind().GroupVersionKind().ToAPIVersionAndKind()

// 	u.SetOwnerReferences([]metav1.OwnerReference{
// 		{
// 			APIVersion: apiVersion,
// 			Kind:       kind,
// 			Name:       owner.GetName(),
// 			UID:        owner.GetUID(),
// 			Controller: PtrTo(true),
// 		},
// 	})
// }

func PtrTo[T any](t T) *T {
	return &t
}
