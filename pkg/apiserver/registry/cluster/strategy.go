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

package cluster

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

	"k8s.io/klog"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/validation"
)

type clusterStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

func NewStrategy(typer runtime.ObjectTyper) clusterStrategy {
	return clusterStrategy{typer, names.SimpleNameGenerator}
}

func (clusterStrategy) NamespaceScoped() bool {
	return true
}

func (clusterStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
}

func (clusterStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
}

func (clusterStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	return field.ErrorList{}
	// return validation.ValidateServiceInjection(obj.(*serviceinjection.ServiceInjection))
}

func (clusterStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (clusterStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (clusterStrategy) Canonicalize(obj runtime.Object) {
}

func (clusterStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	klog.Warningf("Performing cluster update without status validation")
	var status *kops.ClusterStatus
	return validation.ValidateClusterUpdate(obj.(*kops.Cluster), status, old.(*kops.Cluster))
}

func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, bool, error) {
	cluster, ok := obj.(*kops.Cluster)
	if !ok {
		return nil, nil, false, fmt.Errorf("given object is not a Cluster.")
	}
	return labels.Set(cluster.Labels), ClusterToSelectableFields(cluster), cluster.Initializers != nil, nil
}

// MatchCluster is the filter used by the generic etcd backend to watch events
// from etcd to clients of the apiserver only interested in specific labels/fields.
func MatchCluster(label labels.Selector, field fields.Selector) storage.SelectionPredicate {
	return storage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

// ClusterToSelectableFields returns a field set that represents the object.
func ClusterToSelectableFields(obj *kops.Cluster) fields.Set {
	return generic.ObjectMetaFieldsSet(&obj.ObjectMeta, true)
}
