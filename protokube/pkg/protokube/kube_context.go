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

package protokube

import (
	"fmt"
	"sync"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// KubernetesContext is the kubernetes context
type KubernetesContext struct {
	mutex     sync.Mutex
	k8sClient kubernetes.Interface
}

// NewKubernetesContext returns a new KubernetesContext
func NewKubernetesContext() *KubernetesContext {
	return &KubernetesContext{}
}

// KubernetesClient returns a new kubernetes api client
func (c *KubernetesContext) KubernetesClient() (kubernetes.Interface, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.k8sClient == nil {
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig

		configOverrides := &clientcmd.ConfigOverrides{
			ClusterDefaults: clientcmd.ClusterDefaults,
		}

		kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
		config, err := kubeConfig.ClientConfig()
		if err != nil {
			return nil, fmt.Errorf("cannot load kubecfg settings: %v", err)
		}

		k8sClient, err := kubernetes.NewForConfig(config)
		if err != nil {
			return nil, fmt.Errorf("cannot build kube client: %v", err)
		}
		c.k8sClient = k8sClient
	}

	return c.k8sClient, nil
}
