package vfsclientset

import (
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/upup/pkg/api"
	k8sapi "k8s.io/kubernetes/pkg/api"
	"fmt"
	"k8s.io/kops/upup/pkg/api/registry"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"time"
	"k8s.io/kops/util/pkg/vfs"
	"strings"
	"os"
)

type ClusterVFS struct {
	basePath vfs.Path
}

var _ simple.ClusterInterface = &ClusterVFS{}

func (c *ClusterVFS) Get(name string) (*api.Cluster, error) {
	return c.find(name)
}

// Deprecated, but we need this for now..
func (c*ClusterVFS) ConfigBase(clusterName string) (vfs.Path, error) {
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
			return nil, err
		}

		if cluster == nil {
			return nil, fmt.Errorf("cluster not found %q", clusterName)
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

	if c.CreationTimestamp.IsZero() {
		c.CreationTimestamp = unversioned.NewTime(time.Now().UTC())
	}

	configPath := r.basePath.Join(c.Name, registry.PathCluster)

	err = registry.WriteConfig(configPath, c, vfs.WriteOptionCreate)
	if err != nil {
		return nil, fmt.Errorf("error writing Cluster: %v", err)
	}

	return c, nil
}

func (r *ClusterVFS) Update(c *api.Cluster) (*api.Cluster, error) {
	err := c.Validate(false)
	if err != nil {
		return nil, err
	}

	configPath := r.basePath.Join(c.Name, registry.PathCluster)

	err = registry.WriteConfig(configPath, c, vfs.WriteOptionOnlyIfExists)
	if err != nil {
		return nil, fmt.Errorf("error writing cluster %q: %v", c.Name, err)
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
	c := &api.Cluster{}
	err := registry.ReadConfig(configPath, c)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error reading cluster configuration %q: %v", clusterName, err)
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
