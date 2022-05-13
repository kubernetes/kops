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

package validation

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func ValidateAdditionalObject(ctx context.Context, fieldPath *field.Path, u *unstructured.Unstructured) field.ErrorList {
	var errors field.ErrorList

	gvk := u.GroupVersionKind()

	// Note: we use unstructured because:
	// 1) it means we don't have to depend on types / validation code from multiple projects
	// 2) we can more easily differentiate whether a field is set.
	// 3) we can support partial configuration specification - just the fields we care about.
	//
	// It would be nice to be able to consume validation code e.g. via a container,
	// so we could be more extensible.
	errors = append(errors, validateAdditionalObjectKubescheduler(ctx, fieldPath, gvk, u)...)
	return errors
}

func validateAdditionalObjectKubescheduler(ctx context.Context, fieldPath *field.Path, gvk schema.GroupVersionKind, u *unstructured.Unstructured) field.ErrorList {
	var errors field.ErrorList

	if gvk.Kind == "KubeSchedulerConfiguration" && gvk.Group == "kubescheduler.config.k8s.io" {
		kubeconfig, found, err := unstructured.NestedString(u.Object, "clientConnection", "kubeconfig")
		if err != nil {
			errors = append(errors, field.Invalid(fieldPath.Child("clientConnection", "kubeconfig"), u, fmt.Sprintf("error reading field: %v", err)))
		}
		if found && kubeconfig != "" {
			errors = append(errors, field.Invalid(fieldPath.Child("clientConnection", "kubeconfig"), kubeconfig, "value is controlled by kOps and should not be set"))
		}
	}

	return errors
}
