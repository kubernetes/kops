/*
Copyright 2016 The Kubernetes Authors.

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
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/util/pkg/vfs"
)

type VFSClientset struct {
	basePath vfs.Path
}

var _ simple.Clientset = &VFSClientset{}

func (c *VFSClientset) Clusters() simple.ClusterInterface {
	return newClusterVFS(c.basePath)
}

func (c *VFSClientset) InstanceGroups(clusterName string) simple.InstanceGroupInterface {
	return newInstanceGroupVFS(c, clusterName)
}

func (c *VFSClientset) Federations() simple.FederationInterface {
	return newFederationVFS(c)
}

func NewVFSClientset(basePath vfs.Path) simple.Clientset {
	clientset := &VFSClientset{
		basePath: basePath,
	}
	return clientset
}
