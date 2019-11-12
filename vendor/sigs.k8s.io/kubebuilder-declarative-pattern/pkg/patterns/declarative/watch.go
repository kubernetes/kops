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
	"sort"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"sigs.k8s.io/kubebuilder-declarative-pattern/pkg/patterns/declarative/pkg/manifest"
	"sigs.k8s.io/kubebuilder-declarative-pattern/pkg/patterns/declarative/pkg/watch"
)

type Source interface {
	SetSink(sink Sink)
}

type DynamicWatch interface {
	// Add registers a watch for changes to 'trigger' filtered by 'options' to raise an event on 'target'
	Add(trigger schema.GroupVersionKind, options metav1.ListOptions, target metav1.ObjectMeta) error
}

// WatchAll creates a Watch on ctrl for all objects reconciled by recnl
func WatchAll(config *rest.Config, ctrl controller.Controller, recnl Source, labelMaker LabelMaker) (chan struct{}, error) {
	if labelMaker == nil {
		return nil, fmt.Errorf("labelMaker is required to scope watches")
	}

	dw, events, err := watch.NewDynamicWatch(*config)
	if err != nil {
		return nil, fmt.Errorf("creating dynamic watch: %v", err)
	}
	src := &source.Channel{Source: events}
	// Inject a stop channel that will never close. The controller does not have a concept of
	// shutdown, so there is no oppritunity to stop the watch.
	stopCh := make(chan struct{})
	src.InjectStopChannel(stopCh)
	if err := ctrl.Watch(src, &handler.EnqueueRequestForObject{}); err != nil {
		return nil, fmt.Errorf("setting up dynamic watch on the controller: %v", err)
	}
	recnl.SetSink(&watchAll{dw, labelMaker, make(map[string]struct{})})
	return stopCh, nil
}

type watchAll struct {
	dw         DynamicWatch
	labelMaker LabelMaker
	registered map[string]struct{}
}

func (w *watchAll) Notify(ctx context.Context, dest DeclarativeObject, objs *manifest.Objects) error {
	log := log.Log

	labelSelector := strings.Builder{}
	for k, v := range w.labelMaker(ctx, dest) {
		if labelSelector.Len() != 0 {
			labelSelector.WriteRune(',')
		}
		fmt.Fprintf(&labelSelector, "%s=%s", k, fields.EscapeValue(v))
	}

	notify := metav1.ObjectMeta{Name: dest.GetName(), Namespace: dest.GetNamespace()}
	filter := metav1.ListOptions{LabelSelector: labelSelector.String()}

	for _, gvk := range uniqueGroupVersionKind(objs) {
		key := fmt.Sprintf("%s,%s,%s", gvk.String(), labelSelector.String(), dest.GetNamespace())
		if _, ok := w.registered[key]; ok {
			continue
		}

		err := w.dw.Add(gvk, filter, notify)
		if err != nil {
			log.WithValues("GroupVersionKind", gvk.String()).Error(err, "adding watch")
			continue
		}

		w.registered[key] = struct{}{}
	}
	return nil
}

// uniqueGroupVersionKind returns all unique GroupVersionKind defined in objects
func uniqueGroupVersionKind(objects *manifest.Objects) []schema.GroupVersionKind {
	kinds := map[schema.GroupVersionKind]struct{}{}
	for _, o := range objects.Items {
		kinds[o.GroupVersionKind()] = struct{}{}
	}
	var unique []schema.GroupVersionKind
	for gvk := range kinds {
		unique = append(unique, gvk)
	}
	sort.Slice(unique, func(i, j int) bool {
		return unique[i].String() < unique[j].String()
	})
	return unique
}
