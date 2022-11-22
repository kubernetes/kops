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
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
)

// UnstructuredClient is a client that makes it easier to work with unstructured objects.
// It is similar to client.Client in controller-runtime, but is never cached.
type UnstructuredClient struct {
	// client is the dynamic kubernetes client used to apply objects to the k8s cluster.
	client dynamic.Interface
	// restMapper is used to map object kind to resources, and to know if objects are cluster-scoped.
	restMapper meta.RESTMapper
}

// NewUnstructuredClient constructs an UnstructuredClient
func NewUnstructuredClient(options Options) *UnstructuredClient {
	return &UnstructuredClient{
		client:     options.Client,
		restMapper: options.RESTMapper,
	}
}

// dynamicResource is a helper to get the resource for a gvk (with the namespace)
// It returns an error if a namespace is provided for a cluster-scoped resource,
// or no namespace is provided for a namespace-scoped resource.
func (c *UnstructuredClient) dynamicResource(ctx context.Context, gvk schema.GroupVersionKind, ns string) (dynamic.ResourceInterface, error) {
	restMapping, err := c.restMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, fmt.Errorf("error getting rest mapping for %v: %w", gvk, err)
	}
	gvr := restMapping.Resource

	switch restMapping.Scope.Name() {
	case meta.RESTScopeNameNamespace:
		if ns == "" {
			// TODO: Differentiate between server-fixable vs client-fixable errors?
			return nil, fmt.Errorf("namespace was not provided for namespace-scoped object %v", gvk)
		}
		return c.client.Resource(gvr).Namespace(ns), nil

	case meta.RESTScopeNameRoot:
		if ns != "" {
			// TODO: Differentiate between server-fixable vs client-fixable errors?
			return nil, fmt.Errorf("namespace %q was provided for cluster-scoped object %v", ns, gvk)
		}
		return c.client.Resource(gvr), nil

	default:
		// Internal error ... this is panic-level
		return nil, fmt.Errorf("unknown scope for gvk %s: %q", gvk, restMapping.Scope.Name())
	}
}

// Patch performs a Patch operation, used for server-side apply and client-side patch.
func (c *UnstructuredClient) Patch(ctx context.Context, gvk schema.GroupVersionKind, nn types.NamespacedName, patchType types.PatchType, data []byte, opt metav1.PatchOptions) (*unstructured.Unstructured, error) {
	dynamicResource, err := c.dynamicResource(ctx, gvk, nn.Namespace)
	if err != nil {
		return nil, err
	}

	name := nn.Name
	patched, err := dynamicResource.Patch(ctx, name, patchType, data, opt)
	if err != nil {
		return nil, fmt.Errorf("error patching object: %w", err)
	}
	return patched, nil
}

// Update performs an Update operation on the object.  Generally we should prefer server-side-apply.
func (c *UnstructuredClient) Update(ctx context.Context, obj *unstructured.Unstructured, opt metav1.UpdateOptions) (*unstructured.Unstructured, error) {
	gvk := obj.GroupVersionKind()
	dynamicResource, err := c.dynamicResource(ctx, gvk, obj.GetNamespace())
	if err != nil {
		return nil, err
	}

	updated, err := dynamicResource.Update(ctx, obj, opt)
	if err != nil {
		return nil, fmt.Errorf("error updating object: %w", err)
	}
	return updated, nil
}

// Get reads the specified object.
func (c *UnstructuredClient) Get(ctx context.Context, gvk schema.GroupVersionKind, nn types.NamespacedName) (*unstructured.Unstructured, error) {
	dynamicResource, err := c.dynamicResource(ctx, gvk, nn.Namespace)
	if err != nil {
		return nil, err
	}

	obj, err := dynamicResource.Get(ctx, nn.Name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not get existing object: %w", err)
	}

	return obj, nil
}
