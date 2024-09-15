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
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

// RootFS is the root fs path
var RootFS = "/"

// KubeBoot is the options for the protokube service
type KubeBoot struct {
	// Channels is a list of channel to apply
	Channels []string
	// InternalDNSSuffix is the dns zone we are living in
	InternalDNSSuffix string
	// Kubernetes holds a kubernetes client
	Kubernetes *KubernetesContext
	// Master indicates we are a master node
	Master bool

	// BootstrapMasterNodeLabels controls the initial application of node labels to our node
	// The node is found by matching NodeName
	BootstrapMasterNodeLabels bool

	// NodeName is the name of our node as it will be registered in k8s.
	// Used by BootstrapMasterNodeLabels
	NodeName string
}

// RunSyncLoop is responsible for provision the cluster
func (k *KubeBoot) RunSyncLoop() {
	ctx := context.Background()

	if k.Master {
		client, err := k.Kubernetes.KubernetesClient()
		if err != nil {
			panic(fmt.Sprintf("could not create kubernetes client: %v", err))
		}

		klog.Info("polling for apiserver readiness")
		for {
			_, err = client.CoreV1().Namespaces().Get(ctx, "kube-system", metav1.GetOptions{})
			if err == nil {
				klog.Info("successfully connected to the apiserver")
				break
			}
			klog.Infof("failed to connect to the apiserver (will sleep and retry): %v", err)
			time.Sleep(5 * time.Second)
		}
	}

	for {
		if err := k.syncOnce(ctx); err != nil {
			klog.Warningf("error during attempt to bootstrap (will sleep and retry): %v", err)
		}

		time.Sleep(1 * time.Minute)
	}
}

func (k *KubeBoot) syncOnce(ctx context.Context) error {
	if k.Master {
		for _, channel := range k.Channels {
			if err := applyChannel(channel); err != nil {
				klog.Warningf("error applying channel %q: %v", channel, err)
			}
		}
		if k.BootstrapMasterNodeLabels {
			if err := bootstrapMasterNodeLabels(ctx, k.Kubernetes, k.NodeName); err != nil {
				klog.Warningf("error bootstrapping master node labels: %v", err)
			}
		}
	}

	return nil
}
