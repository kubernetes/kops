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

package fake

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
	v1alpha1 "k8s.io/kube-deploy/cluster-api/api/cluster/v1alpha1"
)

// FakeMachines implements MachineInterface
type FakeMachines struct {
	Fake *FakeClusterV1alpha1
}

var machinesResource = schema.GroupVersionResource{Group: "cluster.k8s.io", Version: "v1alpha1", Resource: "machines"}

var machinesKind = schema.GroupVersionKind{Group: "cluster.k8s.io", Version: "v1alpha1", Kind: "Machine"}

// Get takes name of the machine, and returns the corresponding machine object, and an error if there is any.
func (c *FakeMachines) Get(name string, options v1.GetOptions) (result *v1alpha1.Machine, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(machinesResource, name), &v1alpha1.Machine{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Machine), err
}

// List takes label and field selectors, and returns the list of Machines that match those selectors.
func (c *FakeMachines) List(opts v1.ListOptions) (result *v1alpha1.MachineList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(machinesResource, machinesKind, opts), &v1alpha1.MachineList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.MachineList{}
	for _, item := range obj.(*v1alpha1.MachineList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested machines.
func (c *FakeMachines) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(machinesResource, opts))
}

// Create takes the representation of a machine and creates it.  Returns the server's representation of the machine, and an error, if there is any.
func (c *FakeMachines) Create(machine *v1alpha1.Machine) (result *v1alpha1.Machine, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(machinesResource, machine), &v1alpha1.Machine{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Machine), err
}

// Update takes the representation of a machine and updates it. Returns the server's representation of the machine, and an error, if there is any.
func (c *FakeMachines) Update(machine *v1alpha1.Machine) (result *v1alpha1.Machine, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(machinesResource, machine), &v1alpha1.Machine{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Machine), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeMachines) UpdateStatus(machine *v1alpha1.Machine) (*v1alpha1.Machine, error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateSubresourceAction(machinesResource, "status", machine), &v1alpha1.Machine{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Machine), err
}

// Delete takes name of the machine and deletes it. Returns an error if one occurs.
func (c *FakeMachines) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteAction(machinesResource, name), &v1alpha1.Machine{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeMachines) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(machinesResource, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.MachineList{})
	return err
}

// Patch applies the patch and returns the patched machine.
func (c *FakeMachines) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Machine, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(machinesResource, name, data, subresources...), &v1alpha1.Machine{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Machine), err
}
