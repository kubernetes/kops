package protokube

import (
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kubernetes/pkg/client/clientset_generated/release_1_3"
	"k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
	"sync"
)

type KubernetesContext struct {
	mutex  sync.Mutex
	client *release_1_3.Clientset
}

func NewKubernetesContext() *KubernetesContext {
	return &KubernetesContext{}
}

func (c *KubernetesContext) KubernetesClient() (*release_1_3.Clientset, error) {
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

		k8sClient, err := release_1_3.NewForConfig(clientConfig)
		if err != nil {
			return nil, fmt.Errorf("cannot build kube client: %v", err)
		}
		c.client = k8sClient
	}
	return c.client, nil
}
