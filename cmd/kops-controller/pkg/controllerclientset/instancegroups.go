/*
Copyright 2024 The Kubernetes Authors.

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

package controllerclientset

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/klog/v2"
	kopsapi "k8s.io/kops/pkg/apis/kops"
	kopsinternalversion "k8s.io/kops/pkg/client/clientset_generated/clientset/typed/kops/internalversion"
	"k8s.io/kops/pkg/client/simple/vfsclientset"
	"k8s.io/kops/util/pkg/vfs"
)

type instanceGroups struct {
	base        vfsclientset.VFSClientBase
	clusterName string
}

var _ kopsinternalversion.InstanceGroupInterface = &instanceGroups{}

func newInstanceGroups(vfsContext *vfs.VFSContext, cluster *kopsapi.Cluster, clusterBasePath vfs.Path) *instanceGroups {
	if cluster == nil || cluster.Name == "" {
		klog.Fatalf("cluster / cluster.Name is required")
	}

	clusterName := cluster.Name
	kind := "InstanceGroup"

	r := &instanceGroups{
		// cluster:     cluster,
		clusterName: clusterName,
	}
	// We don't expect to need encoding
	var storeVersion runtime.GroupVersioner
	r.base.Init(kind, vfsContext, clusterBasePath.Join("instancegroup"), storeVersion)
	return r
}

func (c *instanceGroups) Get(ctx context.Context, name string, options metav1.GetOptions) (*kopsapi.InstanceGroup, error) {
	if options.ResourceVersion != "" {
		return nil, fmt.Errorf("ResourceVersion not supported in InstanceGroupVFS::Get")
	}

	o, err := c.base.Find(ctx, name)
	if err != nil {
		return nil, err
	}
	if o == nil {
		return nil, errors.NewNotFound(schema.GroupResource{Group: kopsapi.GroupName, Resource: "InstanceGroup"}, name)
	}
	ig := o.(*kopsapi.InstanceGroup)
	c.addLabels(ig)

	return ig, nil
}

func (c *instanceGroups) addLabels(ig *kopsapi.InstanceGroup) {
	if ig.ObjectMeta.Labels == nil {
		ig.ObjectMeta.Labels = make(map[string]string)
	}
	ig.ObjectMeta.Labels[kopsapi.LabelClusterName] = c.clusterName
}

func (c *instanceGroups) List(ctx context.Context, options metav1.ListOptions) (*kopsapi.InstanceGroupList, error) {
	list := &kopsapi.InstanceGroupList{}
	items, err := c.base.List(ctx, list.Items, options)
	if err != nil {
		return nil, err
	}
	list.Items = items.([]kopsapi.InstanceGroup)
	for i := range list.Items {
		c.addLabels(&list.Items[i])
	}
	return list, nil
}

func (c *instanceGroups) Create(ctx context.Context, g *kopsapi.InstanceGroup, opts metav1.CreateOptions) (*kopsapi.InstanceGroup, error) {
	return nil, fmt.Errorf("InstanceGroups::Create not supported for server-side client")
}

func (c *instanceGroups) Update(ctx context.Context, g *kopsapi.InstanceGroup, opts metav1.UpdateOptions) (*kopsapi.InstanceGroup, error) {
	return nil, fmt.Errorf("InstanceGroups::Update not supported for server-side client")
}

func (c *instanceGroups) Delete(ctx context.Context, name string, options metav1.DeleteOptions) error {
	return fmt.Errorf("InstanceGroups::Delete not supported for server-side client")
}

func (r *instanceGroups) DeleteCollection(ctx context.Context, options metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	return fmt.Errorf("InstanceGroups::DeleteCollection not supported for server-side client")
}

func (r *instanceGroups) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return nil, fmt.Errorf("InstanceGroups::Watch not supported for server-side client")
}

func (r *instanceGroups) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *kopsapi.InstanceGroup, err error) {
	return nil, fmt.Errorf("InstanceGroups::Patch not supported for server-side client")
}
