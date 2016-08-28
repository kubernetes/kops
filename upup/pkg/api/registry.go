package api

import (
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/utils"
	"k8s.io/kops/upup/pkg/fi/vfs"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"os"
	"strings"
	"time"
)

// Path for the user-specifed cluster spec
const PathCluster = "config"

// Path for completed cluster spec in the state store
const PathClusterCompleted = "cluster.spec"

// ClusterRegistry is a store of the specs for a group of clusters
type ClusterRegistry struct {
	// basePath is the parent path, each cluster is stored in a directory under this
	basePath vfs.Path
}

func NewClusterRegistry(basePath vfs.Path) *ClusterRegistry {
	registry := &ClusterRegistry{
		basePath: basePath,
	}

	return registry
}

// List returns a slice containing all the cluster names
func (r *ClusterRegistry) List() ([]string, error) {

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

func (r *ClusterRegistry) Find(clusterName string) (*Cluster, error) {
	if clusterName == "" {
		return nil, fmt.Errorf("clusterName is required")
	}
	stateStore := r.stateStore(clusterName)

	c := &Cluster{}
	err := stateStore.ReadConfig(PathCluster, c)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error reading cluster configuration %q: %v", clusterName, err)
	}
	return c, nil
}

func (r *ClusterRegistry) InstanceGroups(clusterName string) (*InstanceGroupRegistry, error) {
	if clusterName == "" {
		return nil, fmt.Errorf("clusterName is required")
	}
	stateStore := r.stateStore(clusterName)

	return &InstanceGroupRegistry{
		stateStore:      stateStore,
		clusterRegistry: r,
	}, nil
}

func (r *ClusterRegistry) WriteCompletedConfig(config *Cluster, writeOptions ...fi.WriteOption) error {
	clusterName := config.Name
	if clusterName == "" {
		return fmt.Errorf("clusterName is required")
	}
	stateStore := r.stateStore(clusterName)

	return stateStore.WriteConfig(PathClusterCompleted, config, writeOptions...)
}

// ReadCompletedConfig reads the "full" cluster spec for the given cluster
func (r *ClusterRegistry) ReadCompletedConfig(clusterName string) (*Cluster, error) {
	if clusterName == "" {
		return nil, fmt.Errorf("clusterName is required")
	}
	stateStore := r.stateStore(clusterName)

	cluster := &Cluster{}
	err := stateStore.ReadConfig(PathClusterCompleted, cluster)
	if err != nil {
		return nil, err
	}
	return cluster, nil
}

func (r *ClusterRegistry) ConfigurationPath(clusterName string) (vfs.Path, error) {
	basePath, err := r.ClusterBase(clusterName)
	if err != nil {
		return nil, err
	}
	return basePath.Join(PathClusterCompleted), nil
}

func (r *ClusterRegistry) ClusterBase(clusterName string) (vfs.Path, error) {
	if clusterName == "" {
		return nil, fmt.Errorf("clusterName is required")
	}
	stateStore := r.stateStore(clusterName)

	return stateStore.VFSPath(), nil
}

func (r *ClusterRegistry) Create(g *Cluster) error {
	clusterName := g.Name
	if clusterName == "" {
		return fmt.Errorf("Name is required")
	}
	stateStore := r.stateStore(clusterName)

	if g.CreationTimestamp.IsZero() {
		g.CreationTimestamp = unversioned.NewTime(time.Now().UTC())
	}

	err := stateStore.WriteConfig(PathCluster, g, fi.WriteOptionCreate)
	if err != nil {
		return fmt.Errorf("error writing cluster: %v", err)
	}

	return nil
}

func (r *ClusterRegistry) Update(g *Cluster) error {
	clusterName := g.Name
	if clusterName == "" {
		return fmt.Errorf("Name is required")
	}

	stateStore := r.stateStore(clusterName)

	err := stateStore.WriteConfig(PathCluster, g, fi.WriteOptionOnlyIfExists)
	if err != nil {
		return fmt.Errorf("error writing cluster %q: %v", g.Name, err)
	}

	return nil
}

func (r *ClusterRegistry) stateStore(clusterName string) fi.StateStore {
	if clusterName == "" {
		panic("clusterName is required")
	}

	stateStore := fi.NewVFSStateStore(r.basePath, clusterName)
	return stateStore
}

func (r *ClusterRegistry) KeyStore(clusterName string) fi.CAStore {
	s := r.stateStore(clusterName)
	return s.CA()
}

func (r *ClusterRegistry) SecretStore(clusterName string) fi.SecretStore {
	s := r.stateStore(clusterName)
	return s.Secrets()
}

type InstanceGroupRegistry struct {
	clusterRegistry *ClusterRegistry
	stateStore      fi.StateStore
}

func (r *InstanceGroupRegistry) List() ([]string, error) {
	keys, err := r.stateStore.ListChildren("instancegroup")
	if err != nil {
		return nil, fmt.Errorf("error listing instancegroups in state store: %v", err)
	}
	return keys, nil
}

func (r *InstanceGroupRegistry) Find(name string) (*InstanceGroup, error) {
	group := &InstanceGroup{}
	err := r.stateStore.ReadConfig("instancegroup/"+name, group)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error reading instancegroup configuration %q: %v", name, err)
	}
	return group, nil
}

func (r *InstanceGroupRegistry) Delete(name string) (bool, error) {
	p := r.stateStore.VFSPath().Join("instancegroup", name)
	err := p.Remove()
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("error deleting instancegroup configuration %q: %v", name, err)
	}
	return true, nil
}

func (r *InstanceGroupRegistry) ReadAll() ([]*InstanceGroup, error) {
	names, err := r.List()
	if err != nil {
		return nil, err
	}

	var instancegroups []*InstanceGroup
	for _, name := range names {
		g, err := r.Find(name)
		if err != nil {
			return nil, err
		}

		if g == nil {
			glog.Warningf("instancegroup was listed, but then not found %q", name)
		}

		instancegroups = append(instancegroups, g)
	}

	return instancegroups, nil
}

func (r *InstanceGroupRegistry) Create(g *InstanceGroup) error {
	if g.Name == "" {
		return fmt.Errorf("Name is required")
	}

	if g.CreationTimestamp.IsZero() {
		g.CreationTimestamp = unversioned.NewTime(time.Now().UTC())
	}

	err := r.stateStore.WriteConfig("instancegroup/"+g.Name, g, fi.WriteOptionCreate)
	if err != nil {
		return fmt.Errorf("error writing instancegroup: %v", err)
	}

	return nil
}

func (r *InstanceGroupRegistry) Update(g *InstanceGroup) error {
	if g.Name == "" {
		return fmt.Errorf("Name is required")
	}

	err := r.stateStore.WriteConfig("instancegroup/"+g.Name, g, fi.WriteOptionOnlyIfExists)
	if err != nil {
		return fmt.Errorf("error writing instancegroup %q: %v", g.Name, err)
	}

	return nil
}

func CreateClusterConfig(clusterRegistry *ClusterRegistry, cluster *Cluster, groups []*InstanceGroup) error {
	// Check for instancegroup Name duplicates before writing
	{
		names := map[string]bool{}
		for i, ns := range groups {
			if ns.Name == "" {
				return fmt.Errorf("InstanceGroup #%d did not have a Name", i+1)
			}
			if names[ns.Name] {
				return fmt.Errorf("Duplicate InstanceGroup Name found: %q", ns.Name)
			}
			names[ns.Name] = true
		}
	}

	clusterName := cluster.Name

	igRegistry, err := clusterRegistry.InstanceGroups(clusterName)
	if err != nil {
		return fmt.Errorf("error getting instance group registry: %v", err)
	}

	err = clusterRegistry.Create(cluster)
	if err != nil {
		return err
	}

	for _, ig := range groups {
		err = igRegistry.Create(ig)
		if err != nil {
			return fmt.Errorf("error writing updated instancegroup configuration: %v", err)
		}
	}

	return nil
}

func UpdateClusterConfig(clusterRegistry *ClusterRegistry, cluster *Cluster, groups []*InstanceGroup) error {
	// Check for instancegroup Name duplicates before writing
	// TODO: Move to deep-validate, DRY with CreateClusterConfig
	{
		names := map[string]bool{}
		for i, ns := range groups {
			if ns.Name == "" {
				return fmt.Errorf("InstanceGroup #%d did not have a Name", i+1)
			}
			if names[ns.Name] {
				return fmt.Errorf("Duplicate InstanceGroup Name found: %q", ns.Name)
			}
			names[ns.Name] = true
		}
	}

	clusterName := cluster.Name

	igRegistry, err := clusterRegistry.InstanceGroups(clusterName)
	if err != nil {
		return fmt.Errorf("error getting instance group registry: %v", err)
	}

	err = clusterRegistry.Update(cluster)
	if err != nil {
		return err
	}

	for _, ig := range groups {
		err = igRegistry.Update(ig)
		if err != nil {
			return fmt.Errorf("error writing updated instancegroup configuration: %v", err)
		}
	}

	return nil
}

func (r *ClusterRegistry) DeleteAllClusterState(clusterName string) error {
	if clusterName == "" {
		return fmt.Errorf("clusterName is required")
	}

	stateStore := r.stateStore(clusterName)

	basePath := stateStore.VFSPath()
	paths, err := basePath.ReadTree()
	if err != nil {
		return fmt.Errorf("error listing files in state store: %v", err)
	}

	for _, path := range paths {
		relativePath, err := vfs.RelativePath(basePath, path)
		if err != nil {
			return err
		}
		if relativePath == "config" || relativePath == "cluster.spec" {
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

func ParseYaml(data []byte, dest interface{}) error {
	// Yaml can't parse empty strings
	configString := string(data)
	configString = strings.TrimSpace(configString)

	if configString != "" {
		err := utils.YamlUnmarshal([]byte(configString), dest)
		if err != nil {
			return fmt.Errorf("error parsing configuration: %v", err)
		}
	}

	return nil
}

func ToYaml(dest interface{}) ([]byte, error) {
	data, err := utils.YamlMarshal(dest)
	if err != nil {
		return nil, fmt.Errorf("error converting to yaml: %v", err)
	}

	return data, nil
}
