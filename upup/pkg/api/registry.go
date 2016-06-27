package api

import (
	"fmt"
	"k8s.io/kube-deploy/upup/pkg/fi"
	"k8s.io/kubernetes/pkg/api/unversioned"
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
