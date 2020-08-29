/*
Copyright 2019 The Kubernetes Authors.

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

package vfsclientset

import (
	"bytes"
	"fmt"
	"os"

	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/acls"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/pkg/kubemanifest"
	"k8s.io/kops/util/pkg/vfs"
)

type vfsAddonsClient struct {
	basePath vfs.Path

	clusterName string
	cluster     *kops.Cluster
}

var _ simple.AddonsClient = &vfsAddonsClient{}

func newAddonsVFS(c *VFSClientset, cluster *kops.Cluster) *vfsAddonsClient {
	if cluster == nil || cluster.Name == "" {
		klog.Fatalf("cluster / cluster.Name is required")
	}

	clusterName := cluster.Name

	r := &vfsAddonsClient{
		cluster:     cluster,
		clusterName: clusterName,
	}
	r.basePath = c.basePath.Join(clusterName, "clusteraddons")

	return r
}

// TODO: Offer partial replacement?
func (c *vfsAddonsClient) Replace(addons kubemanifest.ObjectList) error {
	b, err := addons.ToYAML()
	if err != nil {
		return err
	}

	configPath := c.basePath.Join("default")

	acl, err := acls.GetACL(configPath, c.cluster)
	if err != nil {
		return err
	}

	rs := bytes.NewReader(b)
	if err := configPath.WriteFile(rs, acl); err != nil {
		return fmt.Errorf("error writing addons file %s: %v", configPath, err)
	}

	return nil
}

func (c *vfsAddonsClient) List() (kubemanifest.ObjectList, error) {
	configPath := c.basePath.Join("default")

	b, err := configPath.ReadFile()
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error reading addons file %s: %v", configPath, err)
	}

	objects, err := kubemanifest.LoadObjectsFrom(b)
	if err != nil {
		return nil, err
	}

	return objects, nil
}
