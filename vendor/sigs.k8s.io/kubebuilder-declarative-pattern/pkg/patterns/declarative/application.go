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

// application.go manages an Application[1]
//
// [1] https://github.com/kubernetes-sigs/application
package declarative

import (
	"context"
	"errors"
	"sort"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/kubebuilder-declarative-pattern/pkg/patterns/declarative/pkg/manifest"
)

func transformApplication(ctx context.Context, instance DeclarativeObject, objects *manifest.Objects, labelMaker LabelMaker) error {
	app, err := ExtractApplication(objects)
	if err != nil {
		return err
	}
	if app == nil {
		return errors.New("cannot transformApplication without an app.k8s.io/Application in the manifest")
	}

	labels := labelMaker(ctx, instance)
	convertedLabels := map[string]interface{}{}
	for k, v := range labels {
		convertedLabels[k] = v
	}
	labelSelector := map[string]interface{}{"matchLabels": convertedLabels}
	app.SetNestedField(labelSelector, "spec", "selector")

	componentGroupKinds := []interface{}{}
	for _, gk := range uniqueGroupKind(objects) {
		componentGroupKinds = append(componentGroupKinds, map[string]interface{}{"group": gk.Group, "kind": gk.Kind})
	}
	app.SetNestedSlice(componentGroupKinds, "spec", "componentGroupKinds")

	return nil
}

// uniqueGroupKind returns all unique GroupKind defined in objects
func uniqueGroupKind(objects *manifest.Objects) []metav1.GroupKind {
	kinds := map[metav1.GroupKind]struct{}{}
	for _, o := range objects.Items {
		gk := o.GroupKind()
		kinds[metav1.GroupKind{Group: gk.Group, Kind: gk.Kind}] = struct{}{}
	}
	var unique []metav1.GroupKind
	for gk := range kinds {
		unique = append(unique, gk)
	}
	sort.Slice(unique, func(i, j int) bool {
		return unique[i].String() < unique[j].String()
	})
	return unique
}

// ExtractApplication extracts a single app.k8s.io/Application from objects.
//
// -  0 Application: (nil, nil)
// -  1 Application: (*app, nil)
// - >1 Application: (nil, err)
func ExtractApplication(objects *manifest.Objects) (*manifest.Object, error) {
	var app *manifest.Object
	for _, o := range objects.Items {
		if o.Group == "app.k8s.io" && o.Kind == "Application" {
			if app != nil {
				return nil, errors.New("multiple app.k8s.io/Application found in manifest")
			}
			app = o
		}
	}
	return app, nil
}
