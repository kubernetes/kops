/*
Copyright 2021 The Kubernetes Authors.

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
	"k8s.io/kops/channels/pkg/api"
	"k8s.io/kops/pkg/kubemanifest"
)

type Pruner struct {
	Client     dynamic.Interface
	RESTMapper *restmapper.DeferredDiscoveryRESTMapper
}

// Prune prunes objects not in the manifest, according to PruneSpec.
func (p *Pruner) Prune(ctx context.Context, manifest []byte, spec *api.PruneSpec) error {
	klog.Infof("Prune spec: %v", spec)

	if spec == nil {
		return nil
	}

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

	for i := range spec.Kinds {
		pruneKind := &spec.Kinds[i]
		gk := schema.GroupKind{Group: pruneKind.Group, Kind: pruneKind.Kind}
		if err := p.pruneObjectsOfKind(ctx, gk, pruneKind, objectsByKind[gk]); err != nil {
			return fmt.Errorf("failed to prune objects of kind %s: %w", gk, err)
		}
	}

	return nil
}

func (p *Pruner) pruneObjectsOfKind(ctx context.Context, gk schema.GroupKind, spec *api.PruneKindSpec, keepObjects []*kubemanifest.Object) error {
	klog.Infof("pruning objects of kind: %v", gk)

	restMapping, err := p.RESTMapper.RESTMapping(gk)
	if err != nil {
		return fmt.Errorf("unable to find resource for %s: %w", gk, err)
	}

	gvr := restMapping.Resource

	var listOptions v1.ListOptions
	listOptions.LabelSelector = spec.LabelSelector
	listOptions.FieldSelector = spec.FieldSelector

	baseResource := p.Client.Resource(gvr)
	if len(spec.Namespaces) == 0 {
		objects, err := baseResource.List(ctx, listOptions)
		if err != nil {
			return fmt.Errorf("error listing objects: %w", err)
		}
		if err := p.pruneObjects(ctx, gvr, objects, keepObjects); err != nil {
			return err
		}
	} else {
		for _, namespace := range spec.Namespaces {
			resource := baseResource.Namespace(namespace)
			actualObjects, err := resource.List(ctx, listOptions)
			if err != nil {
				return fmt.Errorf("error listing objects in namespace %s: %w", namespace, err)
			}
			if err := p.pruneObjects(ctx, gvr, actualObjects, keepObjects); err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *Pruner) pruneObjects(ctx context.Context, gvr schema.GroupVersionResource, actualObjects *unstructured.UnstructuredList, keepObjects []*kubemanifest.Object) error {
	keepMap := make(map[string]*kubemanifest.Object)
	for _, keepObject := range keepObjects {
		key := keepObject.GetNamespace() + "/" + keepObject.GetName()
		keepMap[key] = keepObject
	}

	for _, actualObject := range actualObjects.Items {
		name := actualObject.GetName()
		namespace := actualObject.GetNamespace()
		key := namespace + "/" + name
		if _, found := keepMap[key]; found {
			// Object is in manifest, don't delete
			continue
		}

		klog.Infof("pruning %s %s", gvr, key)

		var resource dynamic.ResourceInterface
		if namespace != "" {
			resource = p.Client.Resource(gvr).Namespace(namespace)
		} else {
			resource = p.Client.Resource(gvr)
		}

		var opts v1.DeleteOptions
		if err := resource.Delete(ctx, name, opts); err != nil {
			return fmt.Errorf("failed to delete %s: %w", key, err)
		}
	}

	return nil
}
