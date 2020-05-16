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
	"context"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/registry"
	kopsinternalversion "k8s.io/kops/pkg/client/clientset_generated/clientset/typed/kops/internalversion"
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/secrets"
	"k8s.io/kops/util/pkg/vfs"
)

type VFSClientset struct {
	basePath vfs.Path
}

var _ simple.Clientset = &VFSClientset{}

func (c *VFSClientset) clusters() *ClusterVFS {
	return newClusterVFS(c.basePath)
}

// GetCluster implements the GetCluster method of simple.Clientset for a VFS-backed state store
func (c *VFSClientset) GetCluster(ctx context.Context, name string) (*kops.Cluster, error) {
	return c.clusters().Get(name, metav1.GetOptions{})
}

// UpdateCluster implements the UpdateCluster method of simple.Clientset for a VFS-backed state store
func (c *VFSClientset) UpdateCluster(ctx context.Context, cluster *kops.Cluster, status *kops.ClusterStatus) (*kops.Cluster, error) {
	return c.clusters().Update(cluster, status)
}

// CreateCluster implements the CreateCluster method of simple.Clientset for a VFS-backed state store
func (c *VFSClientset) CreateCluster(ctx context.Context, cluster *kops.Cluster) (*kops.Cluster, error) {
	return c.clusters().Create(cluster)
}

// ListClusters implements the ListClusters method of simple.Clientset for a VFS-backed state store
func (c *VFSClientset) ListClusters(ctx context.Context, options metav1.ListOptions) (*kops.ClusterList, error) {
	return c.clusters().List(options)
}

// ConfigBaseFor implements the ConfigBaseFor method of simple.Clientset for a VFS-backed state store
func (c *VFSClientset) ConfigBaseFor(cluster *kops.Cluster) (vfs.Path, error) {
	if cluster.Spec.ConfigBase != "" {
		return vfs.Context.BuildVfsPath(cluster.Spec.ConfigBase)
	}
	return c.clusters().configBase(cluster.Name)
}

// InstanceGroupsFor implements the InstanceGroupsFor method of simple.Clientset for a VFS-backed state store
func (c *VFSClientset) InstanceGroupsFor(cluster *kops.Cluster) kopsinternalversion.InstanceGroupInterface {
	return newInstanceGroupVFS(c, cluster)
}

func (c *VFSClientset) SecretStore(cluster *kops.Cluster) (fi.SecretStore, error) {
	configBase, err := registry.ConfigBase(cluster)
	if err != nil {
		return nil, err
	}
	basedir := configBase.Join("secrets")
	return secrets.NewVFSSecretStore(cluster, basedir), nil
}

func (c *VFSClientset) KeyStore(cluster *kops.Cluster) (fi.CAStore, error) {
	configBase, err := registry.ConfigBase(cluster)
	if err != nil {
		return nil, err
	}
	basedir := configBase.Join("pki")
	return fi.NewVFSCAStore(cluster, basedir), nil
}

func (c *VFSClientset) SSHCredentialStore(cluster *kops.Cluster) (fi.SSHCredentialStore, error) {
	configBase, err := registry.ConfigBase(cluster)
	if err != nil {
		return nil, err
	}
	basedir := configBase.Join("pki")
	return fi.NewVFSSSHCredentialStore(cluster, basedir), nil
}

func DeleteAllClusterState(basePath vfs.Path) error {
	paths, err := basePath.ReadTree()
	if err != nil {
		return fmt.Errorf("error listing files in state store: %v", err)
	}

	for _, path := range paths {
		relativePath, err := vfs.RelativePath(basePath, path)
		if err != nil {
			return err
		}

		if relativePath == "" {
			continue
		}

		if relativePath == "config" || relativePath == "cluster.spec" {
			continue
		}
		if strings.HasPrefix(relativePath, "addons/") {
			continue
		}
		if strings.HasPrefix(relativePath, "pki/") {
			continue
		}
		if strings.HasPrefix(relativePath, "secrets/") {
			continue
		}
		if strings.HasPrefix(relativePath, "instancegroup/") {
			continue
		}
		if strings.HasPrefix(relativePath, "manifests/") {
			continue
		}
		// TODO: offer an option _not_ to delete backups?
		if strings.HasPrefix(relativePath, "backups/") {
			continue
		}

		return fmt.Errorf("refusing to delete: unknown file found: %s", path)
	}

	for _, path := range paths {
		err = path.Remove()
		if err != nil {
			return fmt.Errorf("error deleting cluster file %s: %v", path, err)
		}
	}

	return nil
}

func (c *VFSClientset) DeleteCluster(ctx context.Context, cluster *kops.Cluster) error {
	configBase, err := registry.ConfigBase(cluster)
	if err != nil {
		return err
	}

	return DeleteAllClusterState(configBase)
}

func NewVFSClientset(basePath vfs.Path) simple.Clientset {
	vfsClientset := &VFSClientset{
		basePath: basePath,
	}
	return vfsClientset
}
