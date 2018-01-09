/*
Copyright 2018 The Kubernetes Authors.

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
	kops "k8s.io/kops/pkg/apis/kops"
)

// FakeFederations implements FederationInterface
type FakeFederations struct {
	Fake *FakeKops
	ns   string
}

var federationsResource = schema.GroupVersionResource{Group: "kops", Version: "", Resource: "federations"}

var federationsKind = schema.GroupVersionKind{Group: "kops", Version: "", Kind: "Federation"}

// Get takes name of the federation, and returns the corresponding federation object, and an error if there is any.
func (c *FakeFederations) Get(name string, options v1.GetOptions) (result *kops.Federation, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(federationsResource, c.ns, name), &kops.Federation{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kops.Federation), err
}

// List takes label and field selectors, and returns the list of Federations that match those selectors.
func (c *FakeFederations) List(opts v1.ListOptions) (result *kops.FederationList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(federationsResource, federationsKind, c.ns, opts), &kops.FederationList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &kops.FederationList{}
	for _, item := range obj.(*kops.FederationList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested federations.
func (c *FakeFederations) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(federationsResource, c.ns, opts))

}

// Create takes the representation of a federation and creates it.  Returns the server's representation of the federation, and an error, if there is any.
func (c *FakeFederations) Create(federation *kops.Federation) (result *kops.Federation, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(federationsResource, c.ns, federation), &kops.Federation{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kops.Federation), err
}

// Update takes the representation of a federation and updates it. Returns the server's representation of the federation, and an error, if there is any.
func (c *FakeFederations) Update(federation *kops.Federation) (result *kops.Federation, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(federationsResource, c.ns, federation), &kops.Federation{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kops.Federation), err
}

// Delete takes name of the federation and deletes it. Returns an error if one occurs.
func (c *FakeFederations) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(federationsResource, c.ns, name), &kops.Federation{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeFederations) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(federationsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &kops.FederationList{})
	return err
}

// Patch applies the patch and returns the patched federation.
func (c *FakeFederations) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *kops.Federation, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(federationsResource, c.ns, name, data, subresources...), &kops.Federation{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kops.Federation), err
}
