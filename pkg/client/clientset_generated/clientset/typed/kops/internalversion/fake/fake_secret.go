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

// FakeSecrets implements SecretInterface
type FakeSecrets struct {
	Fake *FakeKops
	ns   string
}

var secretsResource = schema.GroupVersionResource{Group: "kops", Version: "", Resource: "secrets"}

var secretsKind = schema.GroupVersionKind{Group: "kops", Version: "", Kind: "Secret"}

func (c *FakeSecrets) Create(secret *kops.Secret) (result *kops.Secret, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(secretsResource, c.ns, secret), &kops.Secret{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kops.Secret), err
}

func (c *FakeSecrets) Update(secret *kops.Secret) (result *kops.Secret, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(secretsResource, c.ns, secret), &kops.Secret{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kops.Secret), err
}

func (c *FakeSecrets) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(secretsResource, c.ns, name), &kops.Secret{})

	return err
}

func (c *FakeSecrets) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(secretsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &kops.SecretList{})
	return err
}

func (c *FakeSecrets) Get(name string, options v1.GetOptions) (result *kops.Secret, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(secretsResource, c.ns, name), &kops.Secret{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kops.Secret), err
}

func (c *FakeSecrets) List(opts v1.ListOptions) (result *kops.SecretList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(secretsResource, secretsKind, c.ns, opts), &kops.SecretList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &kops.SecretList{}
	for _, item := range obj.(*kops.SecretList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested secrets.
func (c *FakeSecrets) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(secretsResource, c.ns, opts))

}

// Patch applies the patch and returns the patched secret.
func (c *FakeSecrets) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *kops.Secret, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(secretsResource, c.ns, name, data, subresources...), &kops.Secret{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kops.Secret), err
}
