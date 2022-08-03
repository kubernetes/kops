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
	"encoding/json"
	"fmt"

	"go.uber.org/multierr"
	"k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/kubemanifest"
	"k8s.io/kops/upup/pkg/fi"
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

	objectsByGVK := make(map[schema.GroupVersionKind][]*kubemanifest.Object)
	for _, object := range objects {
		key := object.GetNamespace() + "/" + object.GetName()
		gv, err := schema.ParseGroupVersion(object.APIVersion())
		if err != nil || gv.Version == "" {
			return fmt.Errorf("failed to parse apiVersion %q in object %s", object.APIVersion(), key)
		}
		kind := object.Kind()
		if kind == "" {
			return fmt.Errorf("failed to find kind in object %s", key)
		}

		gvk := gv.WithKind(kind)
		objectsByGVK[gvk] = append(objectsByGVK[gvk], object)
	}

	var applyErrors error
	for gvk := range objectsByGVK {
		if err := p.applyObjectsOfKind(ctx, gvk, objectsByGVK[gvk]); err != nil {
			applyErrors = multierr.Append(applyErrors, fmt.Errorf("failed to apply objects of kind %s: %w", gvk, err))
		}
	}
	return applyErrors
}

func (p *Applier) applyObjectsOfKind(ctx context.Context, gvk schema.GroupVersionKind, expectedObjects []*kubemanifest.Object) error {
	klog.V(2).Infof("applying objects of kind: %v", gvk)

	restMapping, err := p.RESTMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return fmt.Errorf("unable to find resource for %s: %w", gvk, err)
	}

	if err := p.applyObjects(ctx, restMapping, expectedObjects); err != nil {
		return err
	}

	return nil
}

func (p *Applier) applyObjects(ctx context.Context, restMapping *meta.RESTMapping, expectedObjects []*kubemanifest.Object) error {
	var merr error

	for _, expectedObject := range expectedObjects {
		err := p.patchObject(ctx, restMapping, expectedObject)
		merr = multierr.Append(merr, err)
	}

	return merr
}

func (p *Applier) patchObject(ctx context.Context, restMapping *meta.RESTMapping, expectedObject *kubemanifest.Object) error {
	gvr := restMapping.Resource
	name := expectedObject.GetName()
	namespace := expectedObject.GetNamespace()
	key := namespace + "/" + name

	var resource dynamic.ResourceInterface

	if restMapping.Scope.Name() == meta.RESTScopeNameNamespace {
		if namespace == "" {
			return fmt.Errorf("namespace not set for namespace-scoped object %q", key)
		}
		resource = p.Client.Resource(gvr).Namespace(namespace)
	} else {
		if namespace != "" {
			return fmt.Errorf("namespace was set for cluster-scoped object %q", key)
		}
		resource = p.Client.Resource(gvr)
	}

	obj := expectedObject.ToUnstructured()

	jsonData, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("failed to marsal %q into json: %w", obj.GetName(), err)
	}

	{
		_, err := resource.Patch(ctx, obj.GetName(), types.ApplyPatchType, jsonData, v1.PatchOptions{FieldManager: "kops", Force: fi.Bool(true)})
		if err != nil {
			return fmt.Errorf("failed to patch object %q: %w", obj.GetName(), err)
		}
	}
	return nil
}
