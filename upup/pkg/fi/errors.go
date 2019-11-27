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

package fi

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"k8s.io/kops/util/pkg/reflectutils"
)

func RequiredField(key string) error {
	return fmt.Errorf("Field is required: %s", key)
}

func CannotChangeField(key string) error {
	return fmt.Errorf("Field cannot be changed: %s", key)
}

func FieldIsImmutable(newVal, oldVal interface{}, fldPath *field.Path) *field.Error {
	details := fmt.Sprintf("%s: old=%v new=%v", validation.FieldImmutableErrorMsg, reflectutils.FormatValue(oldVal), reflectutils.FormatValue(newVal))
	return field.Invalid(fldPath, newVal, details)
}
