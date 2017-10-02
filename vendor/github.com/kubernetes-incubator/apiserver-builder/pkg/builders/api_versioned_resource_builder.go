/*
Copyright 2017 The Kubernetes Authors.

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

package builders

import (
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/generic/registry"
	"k8s.io/apiserver/pkg/registry/rest"
)

//
// Versioned Kind Builder builds a versioned resource using unversioned strategy
//

// NewApiResource returns a new versionedResourceBuilder for registering endpoints for
// resources that are persisted to storage.
// strategy - unversionedBuilder from calling NewUnversionedXXX()
// new - function for creating new empty VERSIONED instances - e.g. func() runtime.Object { return &Deployment{} }
// newList - function for creating an empty list of VERSIONED instances - e.g. func() runtime.Object { return &DeploymentList{} }
// storeBuilder - builder for creating the store
func NewApiResource(
	unversionedBuilder UnversionedResourceBuilder,
	schemeFns SchemeFns,
	new, newList func() runtime.Object,
	storeBuilder StorageBuilder) *versionedResourceBuilder {

	if storeBuilder == nil {
		storeBuilder = StorageStrategySingleton
	}

	return &versionedResourceBuilder{
		unversionedBuilder, schemeFns, new, newList, storeBuilder, nil, nil,
	}
}

// NewApiResourceWithStorage returns a new versionedResourceBuilder for registering endpoints that
// does not require standard storage (e.g. subresources reuses the storage for the parent resource).
// strategy - unversionedBuilder from calling NewUnversionedXXX()
// new - function for creating new empty VERSIONED instances - e.g. func() runtime.Object { return &Deployment{} }
// storage - storage for manipulating the resource
func NewApiResourceWithStorage(
	unversionedBuilder UnversionedResourceBuilder,
	schemeFns SchemeFns,
	new, newList func() runtime.Object,
	RESTFunc func() rest.Storage) *versionedResourceBuilder {
	v := &versionedResourceBuilder{
		unversionedBuilder, schemeFns, new, newList, nil, RESTFunc, nil,
	}
	if new == nil {
		panic(fmt.Errorf("Cannot call NewApiResourceWithStorage with nil new function."))
	}
	if RESTFunc == nil {
		panic(fmt.Errorf("Cannot call NewApiResourceWithStorage with nil RESTFunc function."))
	}
	return v
}

type versionedResourceBuilder struct {
	Unversioned UnversionedResourceBuilder
	SchemeFns   SchemeFns

	// NewFunc returns an empty unversioned instance of a resource
	NewFunc func() runtime.Object

	// NewListFunc returns and empty unversioned instance of a resource List
	NewListFunc func() runtime.Object

	// StorageBuilder is used to modify the default storage, mutually exclusive with RESTFunc
	StorageBuilder StorageBuilder

	// RESTFunc returns a rest.Storage implementation, mutually exclusive with StorageBuilder
	RESTFunc func() rest.Storage

	Storage rest.StandardStorage
}

func (b *versionedResourceBuilder) New() runtime.Object {
	if b.NewFunc == nil {
		return nil
	}
	return b.NewFunc()
}

func (b *versionedResourceBuilder) NewList() runtime.Object {
	if b.NewListFunc == nil {
		return nil
	}
	return b.NewListFunc()
}

type StorageWrapper struct {
	registry.Store
}

func (s StorageWrapper) Create(ctx request.Context, obj runtime.Object, includeUninitialized bool) (runtime.Object, error) {
	return s.Store.Create(ctx, obj, false)
}

func (b *versionedResourceBuilder) Build(
	group string,
	optionsGetter generic.RESTOptionsGetter) rest.StandardStorage {

	// Set a default strategy
	wcs := 1000
	store := &StorageWrapper{registry.Store{
		Copier:            Scheme,
		NewFunc:           b.Unversioned.New,     // Use the unversioned type
		NewListFunc:       b.Unversioned.NewList, // Use the unversioned type
		QualifiedResource: b.getGroupResource(group),
		WatchCacheSize:    &wcs,
	}}

	// Use default, requires
	options := &generic.StoreOptions{RESTOptions: optionsGetter}

	if b.StorageBuilder != nil {
		// Allow overriding the storage defaults
		b.StorageBuilder.Build(b.StorageBuilder, store, options)
	}

	if err := store.CompleteWithOptions(options); err != nil {
		panic(err) // TODO: Propagate error up
	}
	b.Storage = store
	return store
}

func (b *versionedResourceBuilder) GetStandardStorage() rest.StandardStorage {
	return b.Storage
}

// getGroupResource returns the GroupResource for this Resource and the provided Group
// group is the group the resource belongs to
func (b *versionedResourceBuilder) getGroupResource(group string) schema.GroupResource {
	return schema.GroupResource{group, b.Unversioned.GetName()}

}

// registerEndpoints registers the REST endpoints for this resource in the registry
// group is the group to register the resource under
// optionsGetter is the RESTOptionsGetter provided by a server.Config
// registry is the server.APIGroupInfo VersionedResourcesStorageMap used to register REST endpoints
func (b *versionedResourceBuilder) registerEndpoints(
	group string,
	optionsGetter generic.RESTOptionsGetter,
	registry map[string]rest.Storage) {

	// Register the endpoint
	path := b.Unversioned.GetPath()
	if len(path) > 0 {
		// Subresources appear after the resource
		path = b.Unversioned.GetName() + "/" + path
	} else {
		path = b.Unversioned.GetName()
	}

	if b.RESTFunc != nil {
		// Use the REST implementation directly.
		registry[path] = b.RESTFunc()
	} else {
		// Create a new REST implementation wired to storage.
		registry[path] = b.
			Build(group, optionsGetter)
	}
}
