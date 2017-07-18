/*
Copyright 2016 The Kubernetes Authors.

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

package apiserver

import (
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/apiserver/pkg/server"
	genericapiserver "k8s.io/apiserver/pkg/server"

	"k8s.io/kops/pkg/apis/kops"
	_ "k8s.io/kops/pkg/apis/kops/install"
	"k8s.io/kops/pkg/apis/kops/v1alpha2"
	registrycluster "k8s.io/kops/pkg/apiserver/registry/cluster"
	registryinstancegroup "k8s.io/kops/pkg/apiserver/registry/instancegroup"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func init() {
	// we need to add the options to empty v1
	// TODO fix the server code to avoid this
	metav1.AddToGroupVersion(kops.Scheme, schema.GroupVersion{Version: "v1"})

	// TODO: keep the generic API server from wanting this
	unversioned := schema.GroupVersion{Group: "", Version: "v1"}
	kops.Scheme.AddUnversionedTypes(unversioned,
		&metav1.Status{},
		&metav1.APIVersions{},
		&metav1.APIGroupList{},
		&metav1.APIGroup{},
		&metav1.APIResourceList{},
	)
}

type Config struct {
	GenericConfig *server.Config

	// RESTOptionsGetter is used to construct storage for a particular resource
	RESTOptionsGetter generic.RESTOptionsGetter
}

// APIDiscoveryServer contains state for a Kubernetes cluster master/api server.
type APIDiscoveryServer struct {
	GenericAPIServer *server.GenericAPIServer
}

type completedConfig struct {
	*Config
}

// Complete fills in any fields not set that are required to have valid data. It's mutating the receiver.
func (c *Config) Complete() completedConfig {
	c.GenericConfig.Complete()

	c.GenericConfig.Version = &version.Info{
		Major: "1",
		Minor: "0",
	}

	return completedConfig{c}
}

// SkipComplete provides a way to construct a server instance without config completion.
func (c *Config) SkipComplete() completedConfig {
	return completedConfig{c}
}

// New returns a new instance of APIDiscoveryServer from the given config.
func (c completedConfig) New() (*APIDiscoveryServer, error) {
	genericServer, err := c.Config.GenericConfig.SkipComplete().New("kops-apiserver", genericapiserver.EmptyDelegate) // completion is done in Complete, no need for a second time
	if err != nil {
		return nil, err
	}

	s := &APIDiscoveryServer{
		GenericAPIServer: genericServer,
	}

	apiGroupInfo := server.NewDefaultAPIGroupInfo(kops.GroupName, kops.Registry, kops.Scheme, kops.ParameterCodec, kops.Codecs)

	apiGroupInfo.GroupMeta.GroupVersion = v1alpha2.SchemeGroupVersion
	v1alpha2storage := map[string]rest.Storage{}
	v1alpha2storage["clusters"] = registrycluster.NewREST(c.RESTOptionsGetter)
	//v1alpha2storage["clusters/full"] = registrycluster.NewREST(c.RESTOptionsGetter)
	v1alpha2storage["instancegroups"] = registryinstancegroup.NewREST(c.RESTOptionsGetter)
	apiGroupInfo.VersionedResourcesStorageMap["v1alpha2"] = v1alpha2storage

	if err := s.GenericAPIServer.InstallAPIGroup(&apiGroupInfo); err != nil {
		return nil, err
	}

	return s, nil
}
