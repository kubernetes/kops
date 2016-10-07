package vfsclientset

import (
	"k8s.io/kops/pkg/client/simple"
	"github.com/golang/glog"
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
	if clusterName == "" {
		glog.Fatalf("clusterName is required")
	}
	clusterBasePath := c.basePath.Join(clusterName)

	return &InstanceGroupVFS{
		clusterBasePath: clusterBasePath,
	}
}

func NewVFSClientset(basePath vfs.Path) (simple.Clientset) {
	clientset := &VFSClientset{
		basePath: basePath,
	}
	return clientset
}
