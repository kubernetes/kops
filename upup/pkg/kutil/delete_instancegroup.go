package kutil

import (
	"fmt"

	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/upup/pkg/api"
	"k8s.io/kops/upup/pkg/fi"
)

// DeleteInstanceGroup removes the cloud resources for an InstanceGroup
type DeleteInstanceGroup struct {
	Cluster   *api.Cluster
	Cloud     fi.Cloud
	Clientset simple.Clientset
}

func (c *DeleteInstanceGroup) DeleteInstanceGroup(group *api.InstanceGroup) error {
	groups, err := FindCloudInstanceGroups(c.Cloud, c.Cluster, []*api.InstanceGroup{group}, false, nil)
	cig := groups[group.Name]
	if cig == nil {
		return fmt.Errorf("InstanceGroup not found in cloud")
	}
	if len(groups) != 1 {
		return fmt.Errorf("Multiple InstanceGroup resources found in cloud")
	}

	err = cig.Delete(c.Cloud)
	if err != nil {
		return fmt.Errorf("error deleting cloud resources for InstanceGroup: %v", err)
	}

	err = c.Clientset.InstanceGroups(c.Cluster.Name).Delete(group.Name, nil)
	if err != nil {
		return err
	}

	return nil
}
