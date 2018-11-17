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
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/apiserver/pkg/server"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/install"
	registrycluster "k8s.io/kops/pkg/apiserver/registry/cluster"
	registryinstancegroup "k8s.io/kops/pkg/apiserver/registry/instancegroup"
)

var (
	Scheme = runtime.NewScheme()
	Codecs = serializer.NewCodecFactory(Scheme)
)

func init() {
	install.Install(Scheme)

	// we need to add the options to empty v1
	// TODO fix the server code to avoid this
	metav1.AddToGroupVersion(Scheme, schema.GroupVersion{Version: "v1"})

	// TODO: keep the generic API server from wanting this
	unversioned := schema.GroupVersion{Group: "", Version: "v1"}
	Scheme.AddUnversionedTypes(unversioned,
		&metav1.Status{},
		&metav1.APIVersions{},
		&metav1.APIGroupList{},
		&metav1.APIGroup{},
		&metav1.APIResourceList{},
	)
}

type ExtraConfig struct {
	// Place you custom config here.
}

type Config struct {
	GenericConfig *genericapiserver.RecommendedConfig
	ExtraConfig   ExtraConfig
}

// KopsServer contains state for a Kubernetes cluster master/api server.
type KopsServer struct {
	GenericAPIServer *server.GenericAPIServer
}

type completedConfig struct {
	GenericConfig genericapiserver.CompletedConfig
	ExtraConfig   *ExtraConfig
}

type CompletedConfig struct {
	// Embed a private pointer that cannot be instantiated outside of this package.
	*completedConfig
}

// Complete fills in any fields not set that are required to have valid data. It's mutating the receiver.
func (cfg *Config) Complete() CompletedConfig {
	c := completedConfig{
		cfg.GenericConfig.Complete(),
		&cfg.ExtraConfig,
	}

	c.GenericConfig.Version = &version.Info{
		Major: "1",
		Minor: "0",
	}

	return CompletedConfig{&c}
}

// New returns a new instance of KopsServer from the given config.
func (c completedConfig) New() (*KopsServer, error) {
	genericServer, err := c.GenericConfig.New("kops-apiserver", genericapiserver.NewEmptyDelegate())
	if err != nil {
		return nil, err
	}

	s := &KopsServer{
		GenericAPIServer: genericServer,
	}

	apiGroupInfo := server.NewDefaultAPIGroupInfo(kops.GroupName, Scheme, metav1.ParameterCodec, Codecs)

	//	apiGroupInfo.GroupMeta.GroupVersion = v1alpha2.SchemeGroupVersion

	// {
	// 	v1alpha1storage := map[string]rest.Storage{}
	// 	v1alpha1storage["clusters"], err = registrycluster.NewREST(Scheme, c.GenericConfig.RESTOptionsGetter)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("error initializing clusters: %v", err)
	// 	}
	// 	//v1alpha2stv1alpha1storageorage["clusters/full"] = registrycluster.NewREST(c.RESTOptionsGetter)
	// 	v1alpha1storage["instancegroups"], err = registryinstancegroup.NewREST(Scheme, c.GenericConfig.RESTOptionsGetter)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("error initializing instancegroups: %v", err)
	// 	}
	// 	apiGroupInfo.VersionedResourcesStorageMap["v1alpha1"] = v1alpha1storage
	// }

	{
		v1alpha2storage := map[string]rest.Storage{}
		v1alpha2storage["clusters"], err = registrycluster.NewREST(Scheme, c.GenericConfig.RESTOptionsGetter)
		if err != nil {
			return nil, fmt.Errorf("error initializing clusters: %v", err)
		}
		//v1alpha2storage["clusters/full"] = registrycluster.NewREST(c.RESTOptionsGetter)
		v1alpha2storage["instancegroups"], err = registryinstancegroup.NewREST(Scheme, c.GenericConfig.RESTOptionsGetter)
		if err != nil {
			return nil, fmt.Errorf("error initializing instancegroups: %v", err)
		}
		apiGroupInfo.VersionedResourcesStorageMap["v1alpha2"] = v1alpha2storage
	}

	if err := s.GenericAPIServer.InstallAPIGroup(&apiGroupInfo); err != nil {
		return nil, err
	}

	return s, nil
}
