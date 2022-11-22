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

	apierrors "k8s.io/apimachinery/pkg/api/errors"
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

	client := &UnstructuredClient{
		client:     a.client,
		restMapper: a.restMapper,
	}

	results := &ApplyResults{total: len(trackers.items)}

	for i := range trackers.items {
		tracker := &trackers.items[i]
		expectedObject := tracker.desired

		name := expectedObject.GetName()
		ns := expectedObject.GetNamespace()
		gvk := expectedObject.GroupVersionKind()
		nn := types.NamespacedName{Namespace: ns, Name: name}

		currentObj, err := client.Get(ctx, gvk, nn)
		if err != nil {
			if !apierrors.IsNotFound(err) {
				results.applyError(gvk, nn, err)
				continue
			}
		}

		// If the object exists, we need to update any client-side-apply field-managers
		// Otherwise we often end up with old and new objects combined, which
		// is unexpected and can be invalid.
		if currentObj != nil {
			managedFields := &ManagedFieldsMigrator{
				NewManager: "kops",
				Client:     client,
			}
			if err := managedFields.Migrate(ctx, currentObj); err != nil {
				results.applyError(gvk, nn, err)
				continue
			}
		}

		j, err := json.Marshal(expectedObject)
		if err != nil {
			// TODO: Differentiate between server-fixable vs client-fixable errors?
			results.applyError(gvk, nn, fmt.Errorf("failed to marshal object to JSON: %w", err))
			continue
		}

		lastApplied, err := client.Patch(ctx, gvk, nn, types.ApplyPatchType, j, a.patchOptions)
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
