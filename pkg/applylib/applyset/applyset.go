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

package applyset

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
)

// ApplySet is a set of objects that we want to apply to the cluster.
//
// An ApplySet has a few cases which it tries to optimize for:
// * We can change the objects we're applying
// * We want to watch the objects we're applying / be notified of changes
// * We want to know when the objects we apply are "healthy"
// * We expose a "try once" method to better support running from a controller.
//
// TODO: Pluggable health functions.
// TODO: Pruning
type ApplySet struct {
	// client is the dynamic kubernetes client used to apply objects to the k8s cluster.
	client dynamic.Interface
	// restMapper is used to map object kind to resources, and to know if objects are cluster-scoped.
	restMapper meta.RESTMapper
	// patchOptions holds the options used when applying, in particular the fieldManager
	patchOptions metav1.PatchOptions

	// mutex guards trackers
	mutex sync.Mutex
	// trackers is a (mutable) pointer to the (immutable) objectTrackerList, containing a list of objects we are applying.
	trackers *objectTrackerList
}

// Options holds the parameters for building an ApplySet.
type Options struct {
	// Client is the dynamic kubernetes client used to apply objects to the k8s cluster.
	Client dynamic.Interface
	// RESTMapper is used to map object kind to resources, and to know if objects are cluster-scoped.
	RESTMapper meta.RESTMapper
	// PatchOptions holds the options used when applying, in particular the fieldManager
	PatchOptions metav1.PatchOptions
}

// New constructs a new ApplySet
func New(options Options) (*ApplySet, error) {
	a := &ApplySet{
		client:       options.Client,
		restMapper:   options.RESTMapper,
		patchOptions: options.PatchOptions,
	}
	a.trackers = &objectTrackerList{}
	return a, nil
}

// SetDesiredObjects is used to replace the desired state of all the objects.
// Any objects not specified are removed from the "desired" set.
func (a *ApplySet) SetDesiredObjects(objects []ApplyableObject) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	newTrackers := a.trackers.setDesiredObjects(objects)
	a.trackers = newTrackers

	return nil
}

// ApplyOnce will make one attempt to apply all objects and observe their health.
// It does not wait for the objects to become healthy, but will report their health.
//
// TODO: Limit the amount of time this takes, particularly if we have thousands of objects.
//
//	We don't _have_ to try to apply all objects if it is taking too long.
//
// TODO: We re-apply every object every iteration; we should be able to do better.
func (a *ApplySet) ApplyOnce(ctx context.Context) (*ApplyResults, error) {
	// snapshot the state
	a.mutex.Lock()
	trackers := a.trackers
	a.mutex.Unlock()

	results := &ApplyResults{total: len(trackers.items)}

	for i := range trackers.items {
		tracker := &trackers.items[i]
		obj := tracker.desired

		name := obj.GetName()
		ns := obj.GetNamespace()
		gvk := obj.GroupVersionKind()
		nn := types.NamespacedName{Namespace: ns, Name: name}

		restMapping, err := a.restMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			results.applyError(gvk, nn, fmt.Errorf("error getting rest mapping for %v: %w", gvk, err))
			continue
		}
		gvr := restMapping.Resource

		var dynamicResource dynamic.ResourceInterface

		switch restMapping.Scope.Name() {
		case meta.RESTScopeNameNamespace:
			if ns == "" {
				// TODO: Differentiate between server-fixable vs client-fixable errors?
				results.applyError(gvk, nn, fmt.Errorf("namespace was not provided for namespace-scoped object %v", gvk))
				continue
			}
			dynamicResource = a.client.Resource(gvr).Namespace(ns)

		case meta.RESTScopeNameRoot:
			if ns != "" {
				// TODO: Differentiate between server-fixable vs client-fixable errors?
				results.applyError(gvk, nn, fmt.Errorf("namespace %q was provided for cluster-scoped object %v", obj.GetNamespace(), gvk))
				continue
			}
			dynamicResource = a.client.Resource(gvr)

		default:
			// Internal error ... this is panic-level
			return nil, fmt.Errorf("unknown scope for gvk %s: %q", gvk, restMapping.Scope.Name())
		}

		j, err := json.Marshal(obj)
		if err != nil {
			// TODO: Differentiate between server-fixable vs client-fixable errors?
			results.applyError(gvk, nn, fmt.Errorf("failed to marshal object to JSON: %w", err))
			continue
		}

		lastApplied, err := dynamicResource.Patch(ctx, name, types.ApplyPatchType, j, a.patchOptions)
		if err != nil {
			results.applyError(gvk, nn, fmt.Errorf("error from apply: %w", err))
			continue
		}

		tracker.lastApplied = lastApplied
		results.applySuccess(gvk, nn)
		tracker.isHealthy = isHealthy(lastApplied)
		results.reportHealth(gvk, nn, tracker.isHealthy)
	}
	return results, nil
}
