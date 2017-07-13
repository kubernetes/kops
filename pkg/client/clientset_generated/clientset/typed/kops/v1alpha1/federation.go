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

package v1alpha1

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
	v1alpha1 "k8s.io/kops/pkg/apis/kops/v1alpha1"
	scheme "k8s.io/kops/pkg/client/clientset_generated/clientset/scheme"
)

// FederationsGetter has a method to return a FederationInterface.
// A group's client should implement this interface.
type FederationsGetter interface {
	Federations(namespace string) FederationInterface
}

// FederationInterface has methods to work with Federation resources.
type FederationInterface interface {
	Create(*v1alpha1.Federation) (*v1alpha1.Federation, error)
	Update(*v1alpha1.Federation) (*v1alpha1.Federation, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.Federation, error)
	List(opts v1.ListOptions) (*v1alpha1.FederationList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Federation, err error)
	FederationExpansion
}

// federations implements FederationInterface
type federations struct {
	client rest.Interface
	ns     string
}

// newFederations returns a Federations
func newFederations(c *KopsV1alpha1Client, namespace string) *federations {
	return &federations{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Create takes the representation of a federation and creates it.  Returns the server's representation of the federation, and an error, if there is any.
func (c *federations) Create(federation *v1alpha1.Federation) (result *v1alpha1.Federation, err error) {
	result = &v1alpha1.Federation{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("federations").
		Body(federation).
		Do().
		Into(result)
	return
}

// Update takes the representation of a federation and updates it. Returns the server's representation of the federation, and an error, if there is any.
func (c *federations) Update(federation *v1alpha1.Federation) (result *v1alpha1.Federation, err error) {
	result = &v1alpha1.Federation{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("federations").
		Name(federation.Name).
		Body(federation).
		Do().
		Into(result)
	return
}

// Delete takes name of the federation and deletes it. Returns an error if one occurs.
func (c *federations) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("federations").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *federations) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("federations").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Get takes name of the federation, and returns the corresponding federation object, and an error if there is any.
func (c *federations) Get(name string, options v1.GetOptions) (result *v1alpha1.Federation, err error) {
	result = &v1alpha1.Federation{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("federations").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Federations that match those selectors.
func (c *federations) List(opts v1.ListOptions) (result *v1alpha1.FederationList, err error) {
	result = &v1alpha1.FederationList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("federations").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested federations.
func (c *federations) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("federations").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Patch applies the patch and returns the patched federation.
func (c *federations) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Federation, err error) {
	result = &v1alpha1.Federation{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("federations").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
