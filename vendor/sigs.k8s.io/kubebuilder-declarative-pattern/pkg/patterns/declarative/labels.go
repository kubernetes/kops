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

package declarative

import (
	"context"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"sigs.k8s.io/kubebuilder-declarative-pattern/pkg/patterns/declarative/pkg/manifest"
)

// AddLabels returns an ObjectTransform that adds labels to all the objects
func AddLabels(labels map[string]string) ObjectTransform {
	return func(ctx context.Context, o DeclarativeObject, manifest *manifest.Objects) error {
		log := log.Log
		// TODO: Add to selectors and labels in templates?
		for _, o := range manifest.Items {
			log.WithValues("object", o).WithValues("labels", labels).V(1).Info("add labels to object")
			o.AddLabels(labels)
		}

		return nil
	}
}

// SourceLabel returns a fixed label based on the type and name of the DeclarativeObject
func SourceLabel(scheme *runtime.Scheme) LabelMaker {
	return func(ctx context.Context, o DeclarativeObject) map[string]string {
		log := log.Log

		gvk := o.GetObjectKind().GroupVersionKind()
		gvk, err := apiutil.GVKForObject(o, scheme)

		if err != nil {
			log.WithValues("object", o).WithValues("GroupVersionKind", gvk).Error(err, "can't map GroupVersionKind")
			return map[string]string{}
		}

		if gvk.Group == "" || gvk.Kind == "" {
			log.WithValues("object", o).WithValues("GroupVersionKind", gvk).Info("GroupVersionKind is invalid")
			return map[string]string{}
		}

		return map[string]string{
			fmt.Sprintf("%s/%s", gvk.Group, strings.ToLower(gvk.Kind)): o.GetName(),
		}
	}
}
