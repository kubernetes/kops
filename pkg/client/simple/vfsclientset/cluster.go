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
	"fmt"
	"github.com/golang/glog"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/registry"
	"k8s.io/kops/pkg/apis/kops/v1alpha1"
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/util/pkg/vfs"
	k8sapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"os"
	"strings"
	"time"
)

type ClusterVFS struct {
	commonVFS
}

func newClusterVFS(basePath vfs.Path) *ClusterVFS {
	c := &ClusterVFS{}
	c.init("Cluster", basePath, v1alpha1.SchemeGroupVersion)
	defaultReadVersion := v1alpha1.SchemeGroupVersion.WithKind("Cluster")
	c.defaultReadVersion = &defaultReadVersion
	return c
}

var _ simple.ClusterInterface = &ClusterVFS{}

func (c *ClusterVFS) Get(name string) (*api.Cluster, error) {
	return c.find(name)
}

// Deprecated, but we need this for now..
func (c *ClusterVFS) ConfigBase(clusterName string) (vfs.Path, error) {
	if clusterName == "" {
		return nil, fmt.Errorf("clusterName is required")
	}
	configPath := c.basePath.Join(clusterName)
	return configPath, nil
}

func (c *ClusterVFS) List(options k8sapi.ListOptions) (*api.ClusterList, error) {
	names, err := c.listNames()
	if err != nil {
		return nil, err
	}

	var items []api.Cluster

	for _, clusterName := range names {
		cluster, err := c.find(clusterName)
		if err != nil {
			glog.Warningf("cluster %q found in state store listing, but cannot be loaded: %v", clusterName, err)
			continue
		}

		if cluster == nil {
			glog.Warningf("cluster %q found in state store listing, but doesn't exist now", clusterName)
			continue
		}

		items = append(items, *cluster)
	}

	return &api.ClusterList{Items: items}, nil
}

func (r *ClusterVFS) Create(c *api.Cluster) (*api.Cluster, error) {
	err := c.Validate(false)
	if err != nil {
		return nil, err
	}

	if c.ObjectMeta.CreationTimestamp.IsZero() {
		c.ObjectMeta.CreationTimestamp = unversioned.NewTime(time.Now().UTC())
	}

	clusterName := c.ObjectMeta.Name
	if clusterName == "" {
		return nil, fmt.Errorf("clusterName is required")
	}

	err = r.writeConfig(r.basePath.Join(clusterName, registry.PathCluster), c, vfs.WriteOptionCreate)
	if err != nil {
		if os.IsExist(err) {
			return nil, err
		}
		return nil, fmt.Errorf("error writing Cluster %q: %v", c.ObjectMeta.Name, err)
	}

	return c, nil
}

func (r *ClusterVFS) Update(c *api.Cluster) (*api.Cluster, error) {
	err := c.Validate(false)
	if err != nil {
		return nil, err
	}

	clusterName := c.ObjectMeta.Name
	if clusterName == "" {
		return nil, fmt.Errorf("clusterName is required")
	}

	err = r.writeConfig(r.basePath.Join(clusterName, registry.PathCluster), c, vfs.WriteOptionOnlyIfExists)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, err
		}
		return nil, fmt.Errorf("error writing Cluster: %v", err)
	}

	return c, nil
}

// List returns a slice containing all the cluster names
// It skips directories that don't look like clusters
func (r *ClusterVFS) listNames() ([]string, error) {
	paths, err := r.basePath.ReadTree()
	if err != nil {
		return nil, fmt.Errorf("error reading state store: %v", err)
	}

	var keys []string
	for _, p := range paths {
		relativePath, err := vfs.RelativePath(r.basePath, p)
		if err != nil {
			return nil, err
		}
		if !strings.HasSuffix(relativePath, "/config") {
			continue
		}
		key := strings.TrimSuffix(relativePath, "/config")
		keys = append(keys, key)
	}
	return keys, nil
}

func (r *ClusterVFS) find(clusterName string) (*api.Cluster, error) {
	if clusterName == "" {
		return nil, fmt.Errorf("clusterName is required")
	}
	configPath := r.basePath.Join(clusterName, registry.PathCluster)

	o, err := r.readConfig(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error reading cluster configuration %q: %v", clusterName, err)
	}

	c := o.(*api.Cluster)

	if c.ObjectMeta.Name == "" {
		c.ObjectMeta.Name = clusterName
	}
	if c.ObjectMeta.Name != clusterName {
		glog.Warningf("Name of cluster does not match: %q vs %q", c.ObjectMeta.Name, clusterName)
	}

	// TODO: Split this out into real version updates / schema changes
	if c.Spec.ConfigBase == "" {
		configBase, err := r.ConfigBase(clusterName)
		if err != nil {
			return nil, fmt.Errorf("error building ConfigBase for cluster: %v", err)
		}
		c.Spec.ConfigBase = configBase.Path()
	}

	return c, nil
}
