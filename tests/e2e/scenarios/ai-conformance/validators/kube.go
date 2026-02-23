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

package validators

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DynamicClient returns a dynamic client for the kubernetes cluster.
func (h *ValidatorHarness) DynamicClient() dynamic.Interface {
	if h.dynamicClient == nil {
		dynamicClient, err := dynamic.NewForConfig(h.restConfig)
		if err != nil {
			h.Fatalf("failed to create dynamic client: %v", err)
		}
		h.dynamicClient = dynamicClient
	}
	return h.dynamicClient
}

// DeviceClass is a wrapper around the DRA DeviceClass type.
type DeviceClass struct {
	u *unstructured.Unstructured
}

// Name returns the name of the device class.
func (d *DeviceClass) Name() string {
	return d.u.GetName()
}

var deviceClassGVR = schema.GroupVersionResource{
	Group:    "resource.k8s.io",
	Version:  "v1",
	Resource: "deviceclasses",
}

// ListDeviceClasses lists all device classes in the cluster.
func (h *ValidatorHarness) ListDeviceClasses() []*DeviceClass {
	objectList, err := h.DynamicClient().Resource(deviceClassGVR).List(h.Context(), metav1.ListOptions{})
	if err != nil {
		h.Fatalf("failed to list device classes: %v", err)
	}
	var out []*DeviceClass
	for i := range objectList.Items {
		out = append(out, &DeviceClass{u: &objectList.Items[i]})
	}
	return out
}

// HasDeviceClass returns true if a device class with the given name exists.
func (h *ValidatorHarness) HasDeviceClass(name string) bool {
	for _, deviceClass := range h.ListDeviceClasses() {
		if deviceClass.Name() == name {
			return true
		}
	}
	return false
}

// ResourceSlice is a wrapper around the DRA ResourceSlice type.
type ResourceSlice struct {
	u *unstructured.Unstructured
}

// Name returns the name of the resource slice.
func (d *ResourceSlice) Name() string {
	return d.u.GetName()
}

var resourceSliceGVR = schema.GroupVersionResource{
	Group:    "resource.k8s.io",
	Version:  "v1",
	Resource: "resourceslices",
}

// ListResourceSlices lists all resource slices in the cluster.
func (h *ValidatorHarness) ListResourceSlices() []*ResourceSlice {
	objectList, err := h.DynamicClient().Resource(resourceSliceGVR).List(h.Context(), metav1.ListOptions{})
	if err != nil {
		h.Fatalf("failed to list resource slices: %v", err)
	}
	var out []*ResourceSlice
	for i := range objectList.Items {
		out = append(out, &ResourceSlice{u: &objectList.Items[i]})
	}
	return out
}
