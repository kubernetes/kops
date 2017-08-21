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
	kops "k8s.io/kops/pkg/apis/kops"
)

// FakeSSHSecrets implements SSHSecretInterface
type FakeSSHSecrets struct {
	Fake *FakeKops
	ns   string
}

var sshsecretsResource = schema.GroupVersionResource{Group: "kops", Version: "", Resource: "sshsecrets"}

var sshsecretsKind = schema.GroupVersionKind{Group: "kops", Version: "", Kind: "SSHSecret"}

func (c *FakeSSHSecrets) Create(sSHSecret *kops.SSHSecret) (result *kops.SSHSecret, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(sshsecretsResource, c.ns, sSHSecret), &kops.SSHSecret{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kops.SSHSecret), err
}

func (c *FakeSSHSecrets) Update(sSHSecret *kops.SSHSecret) (result *kops.SSHSecret, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(sshsecretsResource, c.ns, sSHSecret), &kops.SSHSecret{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kops.SSHSecret), err
}

func (c *FakeSSHSecrets) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(sshsecretsResource, c.ns, name), &kops.SSHSecret{})

	return err
}

func (c *FakeSSHSecrets) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(sshsecretsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &kops.SSHSecretList{})
	return err
}

func (c *FakeSSHSecrets) Get(name string, options v1.GetOptions) (result *kops.SSHSecret, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(sshsecretsResource, c.ns, name), &kops.SSHSecret{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kops.SSHSecret), err
}

func (c *FakeSSHSecrets) List(opts v1.ListOptions) (result *kops.SSHSecretList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(sshsecretsResource, sshsecretsKind, c.ns, opts), &kops.SSHSecretList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &kops.SSHSecretList{}
	for _, item := range obj.(*kops.SSHSecretList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested sSHSecrets.
func (c *FakeSSHSecrets) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(sshsecretsResource, c.ns, opts))

}

// Patch applies the patch and returns the patched sSHSecret.
func (c *FakeSSHSecrets) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *kops.SSHSecret, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(sshsecretsResource, c.ns, name, data, subresources...), &kops.SSHSecret{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kops.SSHSecret), err
}
