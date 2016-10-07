package registry

import (
	"fmt"
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/upup/pkg/api"
)

func CreateClusterConfig(clientset simple.Clientset, cluster *api.Cluster, groups []*api.InstanceGroup) error {
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

	_, err := clientset.Clusters().Create(cluster)
	if err != nil {
		return err
	}

	for _, ig := range groups {
		_, err = clientset.InstanceGroups(cluster.Name).Create(ig)
		if err != nil {
			return fmt.Errorf("error writing updated instancegroup configuration: %v", err)
		}
	}

	return nil
}
