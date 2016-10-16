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
	"k8s.io/kubernetes/pkg/api/v1"
)

// PopulateExternalIP sets the external IP on the node which is our node
func PopulateExternalIP(kubeContext *KubernetesContext, nodeName string) error {
	client, err := kubeContext.KubernetesClient()
	if err != nil {
		return err
	}

	glog.V(2).Infof("Querying k8s for node %q", nodeName)
	node, err := client.Core().Nodes().Get(nodeName)
	if err != nil {
		return fmt.Errorf("error querying for node %q: %v", nodeName, err)
	}

	if node == nil {
		return fmt.Errorf("could not find node %q", nodeName)
	}

	var externalIPs []string
	var internalIPs []string
	for i := range node.Status.Addresses {
		a := &node.Status.Addresses[i]
		if a.Type == v1.NodeExternalIP {
			externalIPs = append(externalIPs, a.Address)
		}
		if a.Type == v1.NodeInternalIP {
			internalIPs = append(internalIPs, a.Address)
		}
	}

	if len(externalIPs) > 0 {
		glog.Infof("Node has external ips: %v", externalIPs)
		return nil
	}

	// TODO: source from somewhere else?
	// (but if we do, be careful that we don't delay for long enough that the node will have been updated)
	if len(internalIPs) == 0 {
		return fmt.Errorf("no external ips for node; but no internal IPs to copy")
	}

	for _, a := range internalIPs {
		node.Status.Addresses = append(node.Status.Addresses, v1.NodeAddress{
			Type:    v1.NodeExternalIP,
			Address: a,
		})
	}

	_, err = client.Core().Nodes().UpdateStatus(node)
	if err != nil {
		return fmt.Errorf("error updating node status: %v", err)
	}

	glog.Infof("updated node status with external addresses: %v", internalIPs)
	return nil
}
