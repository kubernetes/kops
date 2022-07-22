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
	"encoding/json"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
)

// isHealthy reports whether the object should be considered "healthy"
// TODO: Replace with kstatus library
func isHealthy(u *unstructured.Unstructured) bool {
	ready := true

	statusConditions, found, err := unstructured.NestedFieldNoCopy(u.Object, "status", "conditions")
	if err != nil || !found {
		klog.Infof("status conditions not found for %s", u.GroupVersionKind())
		return true
	}

	statusConditionsList, ok := statusConditions.([]interface{})
	if !ok {
		klog.Warningf("expected status.conditions to be list, got %T", statusConditions)
		return true
	}
	for i := range statusConditionsList {
		condition := statusConditionsList[i]
		conditionMap, ok := condition.(map[string]interface{})
		if !ok {
			klog.Warningf("expected status.conditions[%d] to be map, got %T", i, condition)
			return true
		}

		conditionType := ""
		conditionStatus := ""
		for k, v := range conditionMap {
			switch k {
			case "type":
				s, ok := v.(string)
				if !ok {
					klog.Warningf("expected status.conditions[].type to be string, got %T", v)
				} else {
					conditionType = s
				}
			case "status":
				s, ok := v.(string)
				if !ok {
					klog.Warningf("expected status.conditions[].status to be string, got %T", v)
				} else {
					conditionStatus = s
				}
			}
		}

		// TODO: Check conditionType?

		switch conditionStatus {
		case "True":
			// ready

		case "False":
			j, _ := json.Marshal(condition)
			klog.Infof("status.conditions indicates object is not ready: %v", string(j))
			ready = false

		case "":
			klog.Warningf("ignoring status.conditions[] type %q with unknown status %q", conditionType, conditionStatus)
		}
	}

	klog.Infof("isHealthy %s %s/%s => %v", u.GroupVersionKind(), u.GetNamespace(), u.GetName(), ready)
	return ready
}
