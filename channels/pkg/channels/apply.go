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

package channels

import (
	"context"
	"fmt"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/kubemanifest"
)

type Applier struct {
	Client     dynamic.Interface
	RESTMapper *restmapper.DeferredDiscoveryRESTMapper
}

// Apply applies the manifest to the cluster.
func (p *Applier) Apply(ctx context.Context, manifest []byte) error {
	objects, err := kubemanifest.LoadObjectsFrom(manifest)
	if err != nil {
		return fmt.Errorf("failed to parse objects: %w", err)
	}

	objectsByKind := make(map[schema.GroupKind][]*kubemanifest.Object)
	for _, object := range objects {
		gv, err := schema.ParseGroupVersion(object.APIVersion())
		if err != nil || gv.Version == "" {
			return fmt.Errorf("failed to parse apiVersion %q", object.APIVersion())
		}
		kind := object.Kind()
		if kind == "" {
			return fmt.Errorf("failed to find kind in object")
		}

		gvk := gv.WithKind(kind)
		gk := gvk.GroupKind()
		objectsByKind[gk] = append(objectsByKind[gk], object)
	}

	for gk := range objectsByKind {
		if err := p.applyObjectsOfKind(ctx, gk, objectsByKind[gk]); err != nil {
			return fmt.Errorf("failed to apply objects of kind %s: %w", gk, err)
		}
	}
	return nil
}

func (p *Applier) applyObjectsOfKind(ctx context.Context, gk schema.GroupKind, expectedObjects []*kubemanifest.Object) error {
	klog.V(2).Infof("applying objects of kind: %v", gk)

	restMapping, err := p.RESTMapper.RESTMapping(gk)
	if err != nil {
		return fmt.Errorf("unable to find resource for %s: %w", gk, err)
	}

	gvr := restMapping.Resource

	baseResource := p.Client.Resource(gvr)

	actualObjects, err := baseResource.List(ctx, v1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error listing objects: %w", err)
	}
	if err := p.applyObjects(ctx, gvr, actualObjects, expectedObjects); err != nil {
		return err
	}

	return nil
}

func (p *Applier) applyObjects(ctx context.Context, gvr schema.GroupVersionResource, actualObjects *unstructured.UnstructuredList, expectedObjects []*kubemanifest.Object) error {
	actualMap := make(map[string]unstructured.Unstructured)
	for _, actualObject := range actualObjects.Items {
		key := actualObject.GetNamespace() + "/" + actualObject.GetName()
		actualMap[key] = actualObject
	}

	for _, expectedObjects := range expectedObjects {
		name := expectedObjects.GetName()
		namespace := expectedObjects.GetNamespace()
		key := namespace + "/" + name

		var resource dynamic.ResourceInterface
		if namespace != "" {
			resource = p.Client.Resource(gvr).Namespace(namespace)
		} else {
			resource = p.Client.Resource(gvr)
		}

		obj := expectedObjects.ToUnstructured()

		if actual, found := actualMap[key]; found {
			klog.V(2).Infof("updating %s %s", gvr, key)
			var opts v1.UpdateOptions
			obj.SetResourceVersion(actual.GetResourceVersion())
			if _, err := resource.Update(ctx, obj, opts); err != nil {
				return fmt.Errorf("failed to create %s: %w", key, err)
			}
		} else {
			klog.V(2).Infof("creating %s %s", gvr, key)
			var opts v1.CreateOptions
			if _, err := resource.Create(ctx, obj, opts); err != nil {
				return fmt.Errorf("failed to create %s: %w", key, err)
			}
		}

	}

	return nil
}
