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
	"errors"

	"github.com/golang/glog"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	clusterv1 "k8s.io/kube-deploy/cluster-api/api/cluster/v1alpha1"
	"k8s.io/kube-deploy/cluster-api/client"
	"k8s.io/kube-deploy/cluster-api-gcp/cloud"
	"k8s.io/kube-deploy/cluster-api-gcp/util"
	apiutil "k8s.io/kube-deploy/cluster-api/util"
)

type MachineController struct {
	config        *Configuration
	restClient    *rest.RESTClient
	kubeClientSet *kubernetes.Clientset
	clusterClient *client.ClusterAPIV1Alpha1Client
	actuator      cloud.MachineActuator
	nodeWatcher   *NodeWatcher
	machineClient client.MachinesInterface
	runner        *asyncRunner
}

func NewMachineController(config *Configuration) *MachineController {
	restClient, err := restClient(config.Kubeconfig)
	if err != nil {
		glog.Fatalf("error creating rest client: %v", err)
	}

	kubeClientSet, err := kubeClientSet(config.Kubeconfig)
	if err != nil {
		glog.Fatalf("error creating kube client set: %v", err)
	}

	clusterClient := client.New(restClient)

	machineClient, err := machineClient(config.Kubeconfig)
	if err != nil {
		glog.Fatalf("error creating machine client: %v", err)
	}

	// Determine cloud type from cluster CRD when available
	actuator, err := cloud.NewMachineActuator(config.Cloud, config.KubeadmToken, machineClient)
	if err != nil {
		glog.Fatalf("error creating machine actuator: %v", err)
	}

	nodeWatcher, err := NewNodeWatcher(config.Kubeconfig)
	if err != nil {
		glog.Fatalf("error creating node watcher: %v", err)
	}

	return &MachineController{
		config:        config,
		restClient:    restClient,
		kubeClientSet: kubeClientSet,
		clusterClient: clusterClient,
		actuator:      actuator,
		nodeWatcher:   nodeWatcher,
		machineClient: machineClient,
		runner:        newAsyncRunner(),
	}
}

func (c *MachineController) Run() error {
	glog.Infof("Running ...")

	// Run leader election

	go func() {
		c.nodeWatcher.Run()
	}()

	return c.run(context.Background())
}

func (c *MachineController) run(ctx context.Context) error {
	source := cache.NewListWatchFromClient(c.restClient, "machines", apiv1.NamespaceAll, fields.Everything())

	_, informer := cache.NewInformer(
		source,
		&clusterv1.Machine{},
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

func (c *MachineController) onAdd(obj interface{}) {
	machine := obj.(*clusterv1.Machine)
	glog.Infof("machine object created: %s\n", machine.ObjectMeta.Name)

	c.runner.runAsync(machine.ObjectMeta.Name, func(){
		err := c.reconcile(machine)
		if err != nil {
			glog.Errorf("processing machine object %s create failed: %v", machine.ObjectMeta.Name, err)
		} else {
			glog.Infof("processing machine object %s create succeded.", machine.ObjectMeta.Name)
		}
	})
}

func (c *MachineController) onUpdate(oldObj, newObj interface{}) {
	oldMachine := oldObj.(*clusterv1.Machine)
	newMachine := newObj.(*clusterv1.Machine)
	glog.Infof("machine object updated: %s\n", oldMachine.ObjectMeta.Name)

	c.runner.runAsync(newMachine.ObjectMeta.Name, func(){
		err := 	c.reconcile(newMachine)
		if err != nil {
			glog.Errorf("processing machine object %s update failed: %v", newMachine.ObjectMeta.Name, err)
		} else {
			glog.Infof("processing machine object %s update succeded.", newMachine.ObjectMeta.Name)
		}
	})
}

func (c *MachineController) onDelete(obj interface{}) {
	machine := obj.(*clusterv1.Machine)
	glog.Infof("machine object deleted: %s\n", machine.ObjectMeta.Name)

	if ignored(machine) {
		return
	}

	c.runner.runAsync(machine.ObjectMeta.Name, func(){
		err := c.reconcile(machine)
		if err != nil {
			glog.Errorf("processing machine object %s delete failed: %v", machine.ObjectMeta.Name, err)
		} else {
			glog.Infof("processing machine object %s delete succeded.", machine.ObjectMeta.Name)
		}
	})
}

func ignored(machine *clusterv1.Machine) bool {
	if apiutil.IsMaster(machine) {
		glog.Infof("Ignoring master machine\n")
		return true
	}
	return false
}

// Reconcile for the given machine the current desired state (in form of machine CRD)
// and the current actual state (in form of actual machine status
func (c *MachineController) reconcile(machine *clusterv1.Machine) error {
	desiredMachine, err := util.GetCurrentMachineIfExists(c.machineClient, machine)
	if err != nil {
		return err
	}

	if desiredMachine == nil {
		// CRD deleted. Delete machine.
		glog.Infof("reconciling machine object %v triggers idempotent delete.", machine.ObjectMeta.Name)
		return c.delete(machine)
	}

	exist, err := c.actuator.Exists(machine)
	if err != nil {
		return err
	}

	if !exist {
		// CRD created. Machine does not yet exist.
		glog.Infof("reconciling machine object %v triggers idempotent create.", machine.ObjectMeta.Name)
		return c.create(desiredMachine)
	}

	glog.Infof("reconciling machine object %v triggers idempotent update.", machine.ObjectMeta.Name)
	return c.update(desiredMachine)
}

func (c *MachineController) create(machine *clusterv1.Machine) error {
	cluster, err := c.getCluster()
	if err != nil {
		return err
	}

	return c.actuator.Create(cluster, machine)
}

func (c *MachineController) delete(machine *clusterv1.Machine) error {
	c.kubeClientSet.CoreV1().Nodes().Delete(machine.ObjectMeta.Name, &metav1.DeleteOptions{})
	if err := c.actuator.Delete(machine); err != nil {
		return err
	}
	// Do a second node cleanup after the delete completes in case the node joined the cluster
	// while the deletion of the machine was mid-way.
	c.kubeClientSet.CoreV1().Nodes().Delete(machine.ObjectMeta.Name, &metav1.DeleteOptions{})
	return nil
}

func (c *MachineController) update(new_machine *clusterv1.Machine) error {
	cluster, err := c.getCluster()
	if err != nil {
		return err
	}

	// TODO: Assume single master for now.
	// TODO: Assume we never change the role for the machines. (Master->Node, Node->Master, etc)
	return c.actuator.Update(cluster, new_machine)
}

//TODO: we should cache this locally and update with an informer
func (c *MachineController) getCluster() (*clusterv1.Cluster, error) {
	clusters, err := c.clusterClient.Clusters().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	switch len(clusters.Items) {
	case 0:
		return nil, errors.New("no clusters defined")
	case 1:
		return &clusters.Items[0], nil
	default:
		return nil, errors.New("multiple clusters defined")
	}
}
