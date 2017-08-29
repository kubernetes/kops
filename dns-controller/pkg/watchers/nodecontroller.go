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

package watchers

import (
	"fmt"
	"time"

	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/kops/dns-controller/pkg/dns"
	"k8s.io/kops/dns-controller/pkg/util"
	kopsutil "k8s.io/kops/pkg/apis/kops/util"
)

// NodeController watches for nodes
type NodeController struct {
	util.Stoppable
	kubeClient kubernetes.Interface
	scope      dns.Scope
}

// newNodeController creates a nodeController
func NewNodeController(kubeClient kubernetes.Interface, dns dns.Context) (*NodeController, error) {
	scope, err := dns.CreateScope("node")
	if err != nil {
		return nil, fmt.Errorf("error building dns scope: %v", err)
	}
	c := &NodeController{
		kubeClient: kubeClient,
		scope:      scope,
	}

	return c, nil
}

// Run starts the NodeController.
func (c *NodeController) Run() {
	glog.Infof("starting node controller")

	stopCh := c.StopChannel()
	go c.runWatcher(stopCh)

	<-stopCh
	glog.Infof("shutting down node controller")
}

func (c *NodeController) runWatcher(stopCh <-chan struct{}) {
	runOnce := func() (bool, error) {
		var listOpts metav1.ListOptions
		glog.V(4).Infof("querying without field filter")

		// Note we need to watch all the nodes, to set up alias targets
		nodeList, err := c.kubeClient.CoreV1().Nodes().List(listOpts)
		if err != nil {
			return false, fmt.Errorf("error listing nodes: %v", err)
		}
		for i := range nodeList.Items {
			node := &nodeList.Items[i]
			glog.V(4).Infof("node: %v", node.Name)
			c.updateNodeRecords(node)
		}
		c.scope.MarkReady()

		listOpts.Watch = true
		listOpts.ResourceVersion = nodeList.ResourceVersion
		watcher, err := c.kubeClient.CoreV1().Nodes().Watch(listOpts)
		if err != nil {
			return false, fmt.Errorf("error watching nodes: %v", err)
		}
		ch := watcher.ResultChan()
		for {
			select {
			case <-stopCh:
				glog.Infof("Got stop signal")
				return true, nil
			case event, ok := <-ch:
				if !ok {
					glog.Infof("node watch channel closed")
					return false, nil
				}

				node := event.Object.(*v1.Node)
				glog.V(4).Infof("node changed: %s %v", event.Type, node.Name)

				switch event.Type {
				case watch.Added, watch.Modified:
					c.updateNodeRecords(node)

				case watch.Deleted:
					c.scope.Replace( /* no namespace for nodes */ node.Name, nil)
				}
			}
		}
	}

	for {
		stop, err := runOnce()
		if stop {
			return
		}

		if err != nil {
			glog.Warningf("Unexpected error in event watch, will retry: %v", err)
			time.Sleep(10 * time.Second)
		}
	}
}

func (c *NodeController) updateNodeRecords(node *v1.Node) {
	var records []dns.Record

	//dnsLabel := node.Labels[LabelNameDns]
	//if dnsLabel != "" {
	//	var ips []string
	//	for _, a := range node.Status.Addresses {
	//		if a.Type != v1.NodeExternalIP {
	//			continue
	//		}
	//		ips = append(ips, a.Address)
	//	}
	//	tokens := strings.Split(dnsLabel, ",")
	//	for _, token := range tokens {
	//		token = strings.TrimSpace(token)
	//
	//		// Assume a FQDN A record
	//		fqdn := token
	//		for _, ip := range ips {
	//			records = append(records, dns.Record{
	//				RecordType: dns.RecordTypeA,
	//				FQDN: fqdn,
	//				Value: ip,
	//			})
	//		}
	//	}
	//}
	//
	//dnsLabelInternal := node.Annotations[AnnotationNameDnsInternal]
	//if dnsLabelInternal != "" {
	//	var ips []string
	//	for _, a := range node.Status.Addresses {
	//		if a.Type != v1.NodeInternalIP {
	//			continue
	//		}
	//		ips = append(ips, a.Address)
	//	}
	//	tokens := strings.Split(dnsLabelInternal, ",")
	//	for _, token := range tokens {
	//		token = strings.TrimSpace(token)
	//
	//		// Assume a FQDN A record
	//		fqdn := dns.EnsureDotSuffix(token)
	//		for _, ip := range ips {
	//			records = append(records, dns.Record{
	//				RecordType: dns.RecordTypeA,
	//				FQDN: fqdn,
	//				Value: ip,
	//			})
	//		}
	//	}
	//}

	// Alias targets

	// node/<name>/internal -> InternalIP
	for _, a := range node.Status.Addresses {
		if a.Type != v1.NodeInternalIP {
			continue
		}
		records = append(records, dns.Record{
			RecordType:  dns.RecordTypeA,
			FQDN:        "node/" + node.Name + "/internal",
			Value:       a.Address,
			AliasTarget: true,
		})
	}

	// node/<name>/external -> ExternalIP
	for _, a := range node.Status.Addresses {
		if a.Type != v1.NodeExternalIP {
			continue
		}
		records = append(records, dns.Record{
			RecordType:  dns.RecordTypeA,
			FQDN:        "node/" + node.Name + "/external",
			Value:       a.Address,
			AliasTarget: true,
		})
	}

	// node/role=<role>/external -> ExternalIP
	// node/role=<role>/internal -> InternalIP
	{
		role := kopsutil.GetNodeRole(node)
		// Default to node
		if role == "" {
			role = "node"
		}

		for _, a := range node.Status.Addresses {
			var roleType string
			if a.Type == v1.NodeInternalIP {
				roleType = dns.RoleTypeInternal
			} else if a.Type == v1.NodeExternalIP {
				roleType = dns.RoleTypeExternal
			}
			records = append(records, dns.Record{
				RecordType:  dns.RecordTypeA,
				FQDN:        dns.AliasForNodesInRole(role, roleType),
				Value:       a.Address,
				AliasTarget: true,
			})
		}
	}

	c.scope.Replace( /* no namespace for nodes */ node.Name, records)
}
