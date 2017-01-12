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

package protokube

import (
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kubernetes/pkg/client/clientset_generated/release_1_5"
	"k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
	"sync"
)

type KubernetesContext struct {
	mutex  sync.Mutex
	client *release_1_5.Clientset
}

func NewKubernetesContext() *KubernetesContext {
	return &KubernetesContext{}
}

func (c *KubernetesContext) KubernetesClient() (*release_1_5.Clientset, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.client == nil {
		config := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			clientcmd.NewDefaultClientConfigLoadingRules(),
			&clientcmd.ConfigOverrides{})
		clientConfig, err := config.ClientConfig()
		if err != nil {
			if clientcmd.IsEmptyConfig(err) {
				glog.V(2).Infof("No client config found; will use default config")
				clientConfig, err = clientcmd.DefaultClientConfig.ClientConfig()
				if err != nil {
					return nil, fmt.Errorf("cannot build default kube config settings: %v", err)
				}
			} else {
				return nil, fmt.Errorf("cannot load kubecfg settings: %v", err)
			}
		}

		k8sClient, err := release_1_5.NewForConfig(clientConfig)
		if err != nil {
			return nil, fmt.Errorf("cannot build kube client: %v", err)
		}
		c.client = k8sClient
	}
	return c.client, nil
}
