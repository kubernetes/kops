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
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

// ApplyResults contains the results of an Apply operation.
type ApplyResults struct {
	total             int
	applySuccessCount int
	applyFailCount    int
	healthyCount      int
	unhealthyCount    int
}

// AllApplied is true if the desired state has been successfully applied for all objects.
// Note: you likely also want to check AllHealthy, if you want to be sure the objects are "ready".
func (r *ApplyResults) AllApplied() bool {
	r.checkInvariants()

	return r.applyFailCount == 0
}

// AllHealthy is true if all the objects have been applied and have converged to a "ready" state.
// Note that this is only meaningful if AllApplied is true.
func (r *ApplyResults) AllHealthy() bool {
	r.checkInvariants()

	return r.unhealthyCount == 0
}

// checkInvariants is an internal function that warns if the object doesn't match the expected invariants.
func (r *ApplyResults) checkInvariants() {
	if r.total != (r.applySuccessCount + r.applyFailCount) {
		klog.Warningf("consistency error (apply counts): %#v", r)
	} else if r.total != (r.healthyCount + r.unhealthyCount) {
		// This "invariant" only holds when all objects could be applied
		klog.Warningf("consistency error (healthy counts): %#v", r)
	}
}

// applyError records that the apply of an object failed with an error.
func (r *ApplyResults) applyError(gvk schema.GroupVersionKind, nn types.NamespacedName, err error) {
	r.applyFailCount++
	klog.Warningf("error from apply on %s %s: %v", gvk, nn, err)
}

// applySuccess records that an object was applied and this succeeded.
func (r *ApplyResults) applySuccess(gvk schema.GroupVersionKind, nn types.NamespacedName) {
	r.applySuccessCount++
}

// reportHealth records the health of an object.
func (r *ApplyResults) reportHealth(gvk schema.GroupVersionKind, nn types.NamespacedName, isHealthy bool) {
	if isHealthy {
		r.healthyCount++
	} else {
		r.unhealthyCount++
	}
}
