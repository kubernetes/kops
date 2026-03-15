/*
Copyright The Kubernetes Authors.

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

package kubeobjects

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DaemonSet is a wrapper around the apps/v1 DaemonSet type.
type DaemonSet struct {
	u *unstructured.Unstructured
}

// Name returns the name of the daemonset.
func (d *DaemonSet) Name() string {
	return d.u.GetName()
}

// Namespace returns the namespace of the daemonset.
func (d *DaemonSet) Namespace() string {
	return d.u.GetNamespace()
}

// NumberReady returns the number of ready pods in the daemonset.
func (d *DaemonSet) NumberReady() int64 {
	val, _, _ := unstructured.NestedInt64(d.u.Object, "status", "numberReady")
	return val
}

// DesiredNumberScheduled returns the desired number of pods in the daemonset.
func (d *DaemonSet) DesiredNumberScheduled() int64 {
	val, _, _ := unstructured.NestedInt64(d.u.Object, "status", "desiredNumberScheduled")
	return val
}

var daemonSetGVR = schema.GroupVersionResource{
	Group:    "apps",
	Version:  "v1",
	Resource: "daemonsets",
}

// ListDaemonSets lists all daemonsets in the given namespace.
func (c *Client) ListDaemonSets(namespace string) []*DaemonSet {
	objectList, err := c.dynamicClient.Resource(daemonSetGVR).Namespace(namespace).List(c.ctx, metav1.ListOptions{})
	if err != nil {
		c.t.Fatalf("failed to list daemonsets: %v", err)
	}
	var out []*DaemonSet
	for i := range objectList.Items {
		out = append(out, &DaemonSet{u: &objectList.Items[i]})
	}
	return out
}
