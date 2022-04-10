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

package cmd

import (
	"fmt"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"

	_ "k8s.io/client-go/plugin/pkg/client/auth"

	certmanager "github.com/cert-manager/cert-manager/pkg/client/clientset/versioned"
)

type Factory interface {
	KubernetesClient() (kubernetes.Interface, error)
	CertManagerClient() (certmanager.Interface, error)
	RESTMapper() (*restmapper.DeferredDiscoveryRESTMapper, error)
	DynamicClient() (dynamic.Interface, error)
}

type DefaultFactory struct {
	ConfigFlags genericclioptions.ConfigFlags

	kubernetesClient  kubernetes.Interface
	certManagerClient certmanager.Interface

	cachedRESTConfig *rest.Config
	dynamicClient    dynamic.Interface
	restMapper       *restmapper.DeferredDiscoveryRESTMapper
}

var _ Factory = &DefaultFactory{}

func (f *DefaultFactory) restConfig() (*rest.Config, error) {
	if f.cachedRESTConfig == nil {
		restConfig, err := f.ConfigFlags.ToRESTConfig()
		if err != nil {
			return nil, fmt.Errorf("cannot load kubecfg settings: %w", err)
		}
		f.cachedRESTConfig = restConfig
	}
	return f.cachedRESTConfig, nil
}

func (f *DefaultFactory) KubernetesClient() (kubernetes.Interface, error) {
	if f.kubernetesClient == nil {
		restConfig, err := f.restConfig()
		if err != nil {
			return nil, err
		}
		k8sClient, err := kubernetes.NewForConfig(restConfig)
		if err != nil {
			return nil, fmt.Errorf("cannot build kube client: %w", err)
		}
		f.kubernetesClient = k8sClient
	}

	return f.kubernetesClient, nil
}

func (f *DefaultFactory) DynamicClient() (dynamic.Interface, error) {
	if f.dynamicClient == nil {
		restConfig, err := f.restConfig()
		if err != nil {
			return nil, fmt.Errorf("cannot load kubecfg settings: %w", err)
		}
		dynamicClient, err := dynamic.NewForConfig(restConfig)
		if err != nil {
			return nil, fmt.Errorf("cannot build dynamicClient client: %v", err)
		}
		f.dynamicClient = dynamicClient
	}

	return f.dynamicClient, nil
}

func (f *DefaultFactory) CertManagerClient() (certmanager.Interface, error) {
	if f.certManagerClient == nil {
		restConfig, err := f.restConfig()
		if err != nil {
			return nil, err
		}
		certManagerClient, err := certmanager.NewForConfig(restConfig)
		if err != nil {
			return nil, fmt.Errorf("cannot build kube client: %v", err)
		}
		f.certManagerClient = certManagerClient
	}

	return f.certManagerClient, nil
}

func (f *DefaultFactory) RESTMapper() (*restmapper.DeferredDiscoveryRESTMapper, error) {
	if f.restMapper == nil {
		discoveryClient, err := f.ConfigFlags.ToDiscoveryClient()
		if err != nil {
			return nil, err
		}

		restMapper := restmapper.NewDeferredDiscoveryRESTMapper(discoveryClient)

		f.restMapper = restMapper
	}

	return f.restMapper, nil
}
