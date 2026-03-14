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
	"k8s.io/apimachinery/pkg/types"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Service is a wrapper around the core Service type.
type Service struct {
	u *unstructured.Unstructured
}

// Name returns the name of the service.
func (s *Service) Name() string {
	return s.u.GetName()
}

// Namespace returns the namespace of the service.
func (s *Service) Namespace() string {
	return s.u.GetNamespace()
}

var serviceGVR = schema.GroupVersionResource{
	Group:    "",
	Version:  "v1",
	Resource: "services",
}

// ListServices lists all services in the cluster.
func (c *Client) ListServices(namespace string) []*Service {
	objectList, err := c.dynamicClient.Resource(serviceGVR).Namespace(namespace).List(c.ctx, metav1.ListOptions{})
	if err != nil {
		c.t.Fatalf("failed to list services: %v", err)
	}
	var out []*Service
	for i := range objectList.Items {
		out = append(out, &Service{u: &objectList.Items[i]})
	}
	return out
}

// HasService returns true if a service with the given name exists.
func (c *Client) HasService(id types.NamespacedName) bool {
	for _, service := range c.ListServices(id.Namespace) {
		if service.Name() == id.Name {
			return true
		}
	}
	return false
}
