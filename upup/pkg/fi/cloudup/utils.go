package cloudup

import (
	"fmt"
	"k8s.io/kube-deploy/upup/pkg/api"
	"k8s.io/kube-deploy/upup/pkg/fi"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/gce"
	"strings"
)

func BuildCloud(cluster *api.Cluster) (fi.Cloud, error) {
	var cloud fi.Cloud

	region := ""
	project := ""

	switch cluster.Spec.CloudProvider {
	case "gce":
		{

			nodeZones := make(map[string]bool)
			for _, zone := range cluster.Spec.Zones {
				nodeZones[zone.Name] = true

				tokens := strings.Split(zone.Name, "-")
				if len(tokens) <= 2 {
					return nil, fmt.Errorf("Invalid GCE Zone: %v", zone.Name)
				}
				zoneRegion := tokens[0] + "-" + tokens[1]
				if region != "" && zoneRegion != region {
					return nil, fmt.Errorf("Clusters cannot span multiple regions")
				}

				region = zoneRegion
			}

			project = cluster.Spec.Project
			if project == "" {
				return nil, fmt.Errorf("project is required for GCE")
			}
			gceCloud, err := gce.NewGCECloud(region, project)
			if err != nil {
				return nil, err
			}

			cloud = gceCloud
		}

	case "aws":
		{

			nodeZones := make(map[string]bool)
			for _, zone := range cluster.Spec.Zones {
				if len(zone.Name) <= 2 {
					return nil, fmt.Errorf("Invalid AWS zone: %q", zone.Name)
				}

				nodeZones[zone.Name] = true

				zoneRegion := zone.Name[:len(zone.Name)-1]
				if region != "" && zoneRegion != region {
					return nil, fmt.Errorf("Clusters cannot span multiple regions")
				}

				region = zoneRegion
			}

			err := awsup.ValidateRegion(region)
			if err != nil {
				return nil, err
			}

			cloudTags := map[string]string{awsup.TagClusterName: cluster.Name}

			awsCloud, err := awsup.NewAWSCloud(region, cloudTags)
			if err != nil {
				return nil, err
			}

			var zoneNames []string
			for _, z := range cluster.Spec.Zones {
				zoneNames = append(zoneNames, z.Name)
			}
			err = awsCloud.ValidateZones(zoneNames)
			if err != nil {
				return nil, err
			}
			cloud = awsCloud
		}

	default:
		return nil, fmt.Errorf("unknown CloudProvider %q", cluster.Spec.CloudProvider)
	}
	return cloud, nil
}
