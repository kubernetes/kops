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
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
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
