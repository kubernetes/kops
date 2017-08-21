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

// SSHSecretsGetter has a method to return a SSHSecretInterface.
// A group's client should implement this interface.
type SSHSecretsGetter interface {
	SSHSecrets(namespace string) SSHSecretInterface
}

// SSHSecretInterface has methods to work with SSHSecret resources.
type SSHSecretInterface interface {
	Create(*v1alpha1.SSHSecret) (*v1alpha1.SSHSecret, error)
	Update(*v1alpha1.SSHSecret) (*v1alpha1.SSHSecret, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.SSHSecret, error)
	List(opts v1.ListOptions) (*v1alpha1.SSHSecretList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.SSHSecret, err error)
	SSHSecretExpansion
}

// sSHSecrets implements SSHSecretInterface
type sSHSecrets struct {
	client rest.Interface
	ns     string
}

// newSSHSecrets returns a SSHSecrets
func newSSHSecrets(c *KopsV1alpha1Client, namespace string) *sSHSecrets {
	return &sSHSecrets{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Create takes the representation of a sSHSecret and creates it.  Returns the server's representation of the sSHSecret, and an error, if there is any.
func (c *sSHSecrets) Create(sSHSecret *v1alpha1.SSHSecret) (result *v1alpha1.SSHSecret, err error) {
	result = &v1alpha1.SSHSecret{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("sshsecrets").
		Body(sSHSecret).
		Do().
		Into(result)
	return
}

// Update takes the representation of a sSHSecret and updates it. Returns the server's representation of the sSHSecret, and an error, if there is any.
func (c *sSHSecrets) Update(sSHSecret *v1alpha1.SSHSecret) (result *v1alpha1.SSHSecret, err error) {
	result = &v1alpha1.SSHSecret{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("sshsecrets").
		Name(sSHSecret.Name).
		Body(sSHSecret).
		Do().
		Into(result)
	return
}

// Delete takes name of the sSHSecret and deletes it. Returns an error if one occurs.
func (c *sSHSecrets) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("sshsecrets").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *sSHSecrets) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("sshsecrets").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Get takes name of the sSHSecret, and returns the corresponding sSHSecret object, and an error if there is any.
func (c *sSHSecrets) Get(name string, options v1.GetOptions) (result *v1alpha1.SSHSecret, err error) {
	result = &v1alpha1.SSHSecret{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("sshsecrets").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of SSHSecrets that match those selectors.
func (c *sSHSecrets) List(opts v1.ListOptions) (result *v1alpha1.SSHSecretList, err error) {
	result = &v1alpha1.SSHSecretList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("sshsecrets").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested sSHSecrets.
func (c *sSHSecrets) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("sshsecrets").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Patch applies the patch and returns the patched sSHSecret.
func (c *sSHSecrets) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.SSHSecret, err error) {
	result = &v1alpha1.SSHSecret{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("sshsecrets").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
