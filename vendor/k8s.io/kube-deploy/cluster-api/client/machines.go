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

package client

import (
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	scheme "k8s.io/client-go/kubernetes/scheme"
	rest "k8s.io/client-go/rest"
	clusterv1 "k8s.io/kube-deploy/cluster-api/api/cluster/v1alpha1"
)

type MachinesGetter interface {
	Machines() MachinesInterface
}

// MachinesInterface has methods to work with Machine resources.
type MachinesInterface interface {
	Create(*clusterv1.Machine) (*clusterv1.Machine, error)
	Update(*clusterv1.Machine) (*clusterv1.Machine, error)
	Delete(string, *metav1.DeleteOptions) error
	List(metav1.ListOptions) (*clusterv1.MachineList, error)
	Get(string, metav1.GetOptions) (*clusterv1.Machine, error)
}

// machines implements MachinesInterface
type machines struct {
	client rest.Interface
}

// newMachines returns a machines
func newMachines(c *ClusterAPIV1Alpha1Client) *machines {
	return &machines{
		client: c.RESTClient(),
	}
}

func (c *machines) Create(machine *clusterv1.Machine) (result *clusterv1.Machine, err error) {
	result = &clusterv1.Machine{}
	err = c.client.Post().
		Namespace(apiv1.NamespaceDefault).
		Resource(clusterv1.MachinesCRDPlural).
		Body(machine).
		Do().
		Into(result)
	return
}

func (c *machines) Update(machine *clusterv1.Machine) (result *clusterv1.Machine, err error) {
	result = &clusterv1.Machine{}
	err = c.client.Put().
		Namespace(apiv1.NamespaceDefault).
		Resource(clusterv1.MachinesCRDPlural).
		Name(machine.Name).
		Body(machine).
		Do().
		Into(result)
	return
}

func (c *machines) Delete(name string, options *metav1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(apiv1.NamespaceDefault).
		Resource(clusterv1.MachinesCRDPlural).
		Name(name).
		Body(options).
		Do().
		Error()
}

// List takes label and field selectors, and returns the list of machines that match those selectors.
func (c *machines) List(opts metav1.ListOptions) (result *clusterv1.MachineList, err error) {
	result = &clusterv1.MachineList{}
	err = c.client.Get().
		Namespace(apiv1.NamespaceDefault).
		Resource(clusterv1.MachinesCRDPlural).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

func (c *machines) Get(name string, options metav1.GetOptions) (result *clusterv1.Machine, err error) {
	result = &clusterv1.Machine{}
	err = c.client.Get().
		Namespace(apiv1.NamespaceDefault).
		Resource(clusterv1.MachinesCRDPlural).
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}
