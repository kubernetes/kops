package api

import (
	"fmt"
	"k8s.io/kube-deploy/upup/pkg/fi"
	"k8s.io/kube-deploy/upup/pkg/fi/vfs"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"strings"
	"time"
)

func WriteConfig(stateStore fi.StateStore, cluster *Cluster, groups []*InstanceGroup) error {
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
	if cluster.CreationTimestamp.IsZero() {
		cluster.CreationTimestamp = unversioned.NewTime(time.Now().UTC())
	}
	err := stateStore.WriteConfig("config", cluster)
	if err != nil {
		return fmt.Errorf("error writing updated cluster configuration: %v", err)
	}

	for _, ns := range groups {
		if ns.CreationTimestamp.IsZero() {
			ns.CreationTimestamp = unversioned.NewTime(time.Now().UTC())
		}
		err = stateStore.WriteConfig("instancegroup/"+ns.Name, ns)
		if err != nil {
			return fmt.Errorf("error writing updated instancegroup configuration: %v", err)
		}
	}

	return nil
}

func ReadConfig(stateStore fi.StateStore) (*Cluster, []*InstanceGroup, error) {
	cluster := &Cluster{}
	err := stateStore.ReadConfig("config", cluster)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading cluster configuration: %v", err)
	}

	var instanceGroups []*InstanceGroup
	keys, err := stateStore.ListChildren("instancegroup")
	if err != nil {
		return nil, nil, fmt.Errorf("error listing instancegroups in state store: %v", err)
	}
	for _, key := range keys {
		group := &InstanceGroup{}
		err = stateStore.ReadConfig("instancegroup/"+key, group)
		if err != nil {
			return nil, nil, fmt.Errorf("error reading instancegroup configuration %q: %v", key, err)
		}
		instanceGroups = append(instanceGroups, group)
	}

	return cluster, instanceGroups, nil
}

func DeleteConfig(stateStore fi.StateStore) error {
	paths, err := stateStore.VFSPath().ReadTree()
	if err != nil {
		return fmt.Errorf("error listing files in state store: %v", err)
	}

	for _, path := range paths {
		relativePath, err := vfs.RelativePath(stateStore.VFSPath(), path)
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
