/*
Copyright 2017 The Kubernetes Authors.

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

package instancegroup

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/storage"
	"k8s.io/apiserver/pkg/storage/names"

	"k8s.io/kops/pkg/apis/kops"
)

type instanceGroupStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

func NewStrategy(typer runtime.ObjectTyper) instanceGroupStrategy {
	return instanceGroupStrategy{typer, names.SimpleNameGenerator}
}

func (instanceGroupStrategy) NamespaceScoped() bool {
	return true
}

func (instanceGroupStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
}

func (instanceGroupStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
}

func (instanceGroupStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	return field.ErrorList{}
	// return validation.ValidateServiceInjection(obj.(*serviceinjection.ServiceInjection))
}

func (instanceGroupStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (instanceGroupStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (instanceGroupStrategy) Canonicalize(obj runtime.Object) {
}

func (instanceGroupStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return field.ErrorList{}
	// return validation.ValidateServiceInjectionUpdate(obj.(*serviceinjection.ServiceInjection), old.(*serviceinjection.ServiceInjection))
}

func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, bool, error) {
	instanceGroup, ok := obj.(*kops.InstanceGroup)
	if !ok {
		return nil, nil, false, fmt.Errorf("given object is not an InstanceGroup.")
	}
	return labels.Set(instanceGroup.Labels), InstanceGroupToSelectableFields(instanceGroup), instanceGroup.Initializers != nil, nil
}

// MatchInstanceGroup is the filter used by the generic etcd backend to watch events
// from etcd to clients of the apiserver only interested in specific labels/fields.
func MatchInstanceGroup(label labels.Selector, field fields.Selector) storage.SelectionPredicate {
	return storage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

// InstanceGroupToSelectableFields returns a field set that represents the object.
func InstanceGroupToSelectableFields(obj *kops.InstanceGroup) fields.Set {
	return generic.ObjectMetaFieldsSet(&obj.ObjectMeta, true)
}
