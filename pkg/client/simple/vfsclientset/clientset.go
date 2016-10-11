package vfsclientset

import (
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/util/pkg/vfs"
)

type VFSClientset struct {
	basePath        vfs.Path
}

var _ simple.Clientset = &VFSClientset{}

func (c *VFSClientset) Clusters() simple.ClusterInterface {
	return &ClusterVFS{basePath: c.basePath}
}

func (c *VFSClientset) InstanceGroups(clusterName string) simple.InstanceGroupInterface {
	return newInstanceGroupVFS(c, clusterName)
}

func (c *VFSClientset) Federations() simple.FederationInterface {
	return newFederationVFS(c)
}

func NewVFSClientset(basePath vfs.Path) (simple.Clientset) {
	clientset := &VFSClientset{
		basePath: basePath,
	}
	return clientset
}
