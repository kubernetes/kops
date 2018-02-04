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

package controller

import (
	"context"

	"github.com/golang/glog"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	clusterapiclient "k8s.io/kube-deploy/cluster-api/client"
)

// NodeWatcher watches updates to core "Node" objects and takes action on their
// events. Currently, its only responsibility is to find nodes that have a
// "machine" annotation and link the corresponding "Machine" object to them.
//
// The "machine" annotation is an implementation detail of how the two objects
// can get linked together, but it is not required behavior. However, in the
// event that a Machine.Spec update requires replacing the Node, this can allow
// for faster turn-around time by allowing a new Node to be created with a new
// name while the old node is being deleted.
//
// Currently, these annotations are added by the node itself as part of its
// bootup script after "kubeadm join" succeeds.
type NodeWatcher struct {
	nodeClient    *kubernetes.Clientset
	machineClient clusterapiclient.MachinesInterface
	linkedNodes   map[string]bool
}

func NewNodeWatcher(kubeconfig string) (*NodeWatcher, error) {
	nodeClient, err := nodeClient(kubeconfig)
	if err != nil {
		return nil, err
	}

	machineClient, err := machineClient(kubeconfig)
	if err != nil {
		return nil, err
	}

	return &NodeWatcher{
		nodeClient:    nodeClient,
		machineClient: machineClient,
		linkedNodes:   make(map[string]bool),
	}, nil
}

func (c *NodeWatcher) Run() error {
	glog.Infof("Running node watcher...")

	return c.run(context.Background())
}

func (c *NodeWatcher) run(ctx context.Context) error {
	source := cache.NewListWatchFromClient(c.nodeClient.CoreV1().RESTClient(), "nodes", corev1.NamespaceAll, fields.Everything())

	_, informer := cache.NewInformer(
		source,
		&corev1.Node{},
		0,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    c.onAdd,
			UpdateFunc: c.onUpdate,
			DeleteFunc: c.onDelete,
		},
	)

	informer.Run(ctx.Done())
	return nil
}

func (c *NodeWatcher) onAdd(obj interface{}) {
	node := obj.(*corev1.Node)
	glog.Infof("node created: %s\n", node.ObjectMeta.Name)
	c.link(node)
}

func (c *NodeWatcher) onUpdate(oldObj, newObj interface{}) {
	newNode := newObj.(*corev1.Node)
	glog.V(2).Infof("node updated: %s\n", newNode.ObjectMeta.Name)
	c.link(newNode)
}

func (c *NodeWatcher) onDelete(obj interface{}) {
	node := obj.(*corev1.Node)
	glog.Infof("node deleted: %s\n", node.ObjectMeta.Name)
	c.unlink(node)
}

func (c *NodeWatcher) link(node *corev1.Node) {
	if val, _ := c.linkedNodes[node.ObjectMeta.Name]; val {
		return
	}

	if val, ok := node.ObjectMeta.Annotations["machine"]; ok {
		machine, err := c.machineClient.Get(val, metav1.GetOptions{})
		if err != nil {
			glog.Errorf("Error getting machine %v: %v\n", val, err)
			return
		}

		machine.Status.NodeRef = objectRef(node)

		if _, err := c.machineClient.Update(machine); err != nil {
			glog.Errorf("Error updating machine to link to node: %v\n", err)
		} else {
			glog.Infof("Successfully linked machine %s to node %s\n",
				machine.ObjectMeta.Name, node.ObjectMeta.Name)
			c.linkedNodes[node.ObjectMeta.Name] = true
		}
	}
}

func (c *NodeWatcher) unlink(node *corev1.Node) {
	if val, ok := node.ObjectMeta.Annotations["machine"]; ok {
		machine, err := c.machineClient.Get(val, metav1.GetOptions{})
		if err != nil {
			glog.Errorf("Error getting machine %v: %v\n", val, err)
			return
		}

		// This machine has no link to remove
		if machine.Status.NodeRef == nil {
			return
		}

		// This machine was linked to a different node, don't unlink them
		if machine.Status.NodeRef.Name != node.ObjectMeta.Name {
			return
		}

		machine.Status.NodeRef = nil

		if _, err := c.machineClient.Update(machine); err != nil {
			glog.Errorf("Error updating machine %s to unlink node %s: %v\n",
				machine.ObjectMeta.Name, node.ObjectMeta.Name, err)
		} else {
			glog.Infof("Successfully unlinked node %s from machine %s\n",
				node.ObjectMeta.Name, machine.ObjectMeta.Name)
		}
	}
}

func objectRef(node *corev1.Node) *corev1.ObjectReference {
	return &corev1.ObjectReference{
		Kind:      "Node",
		Namespace: node.ObjectMeta.Namespace,
		Name:      node.ObjectMeta.Name,
	}
}
