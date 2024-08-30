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

package cloudup

import (
	"context"
	"fmt"
	"strings"

	"k8s.io/klog/v2"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider/providers/aws/route53"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"
	"k8s.io/kops/upup/pkg/fi/cloudup/do"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/hetzner"
	"k8s.io/kops/upup/pkg/fi/cloudup/metal"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
	"k8s.io/kops/upup/pkg/fi/cloudup/scaleway"
)

func BuildCloud(cluster *kops.Cluster) (fi.Cloud, error) {
	var cloud fi.Cloud
	ctx := context.TODO()

	region := ""
	project := ""

	switch cluster.GetCloudProvider() {
	case kops.CloudProviderGCE:
		{
			for _, subnet := range cluster.Spec.Networking.Subnets {
				if subnet.Region != "" {
					region = subnet.Region
				}
			}
			if region == "" {
				return nil, fmt.Errorf("on GCE, subnets must include Regions")
			}

			project = cluster.Spec.CloudProvider.GCE.Project
			if project == "" {
				return nil, fmt.Errorf("project is required for GCE - try gcloud config get-value project")
			}

			clusterLabel := gce.LabelForCluster(cluster.ObjectMeta.Name)
			labels := map[string]string{clusterLabel.Key: clusterLabel.Value}

			gceCloud, err := gce.NewGCECloud(region, project, labels)
			if err != nil {
				return nil, err
			}

			cloud = gceCloud
		}

	case kops.CloudProviderAWS:
		{
			region, err := awsup.FindRegion(cluster)
			if err != nil {
				return nil, err
			}

			err = awsup.ValidateRegion(ctx, region)
			if err != nil {
				return nil, err
			}

			cloudTags := map[string]string{awsup.TagClusterName: cluster.ObjectMeta.Name}

			awsCloud, err := awsup.NewAWSCloud(region, cloudTags)
			if err != nil {
				return nil, err
			}

			var zoneNames []string
			for _, subnet := range cluster.Spec.Networking.Subnets {
				zoneNames = append(zoneNames, subnet.Zone)
			}
			err = awsup.ValidateZones(zoneNames, awsCloud)
			if err != nil {
				return nil, err
			}
			cloud = awsCloud
		}
	case kops.CloudProviderDO:
		{
			// for development purposes we're going to assume
			// single region setups for DO. Reconsider this logic
			// when setting up multi-region kubernetes clusters on DO
			region := cluster.Spec.Networking.Subnets[0].Zone
			doCloud, err := do.NewDOCloud(region)
			if err != nil {
				return nil, fmt.Errorf("error initializing digitalocean cloud: %s", err)
			}

			cloud = doCloud
		}

	case kops.CloudProviderHetzner:
		{
			region, err := hetzner.FindRegion(cluster)
			if err != nil {
				return nil, err
			}

			hetznerCloud, err := hetzner.NewHetznerCloud(region)
			if err != nil {
				return nil, fmt.Errorf("error initializing hetzner cloud: %s", err)
			}

			cloud = hetznerCloud
		}

	case kops.CloudProviderOpenstack:
		{
			osc, err := openstack.NewOpenstackCloud(cluster, "build-cloud")
			if err != nil {
				return nil, err
			}
			var zoneNames []string
			for _, subnet := range cluster.Spec.Networking.Subnets {
				if !fi.ArrayContains(zoneNames, subnet.Zone) {
					zoneNames = append(zoneNames, subnet.Zone)
				}
			}
			osc.UseZones(zoneNames)
			cloud = osc
		}

	case kops.CloudProviderAzure:
		{
			for _, subnet := range cluster.Spec.Networking.Subnets {
				if subnet.Region != "" {
					region = subnet.Region
				}
			}
			if region == "" {
				return nil, fmt.Errorf("on Azure, subnets must include Regions")
			}

			cloudTags := map[string]string{azure.TagClusterName: cluster.ObjectMeta.Name}

			subscriptionID := cluster.Spec.CloudProvider.Azure.SubscriptionID
			resourceGroupName := cluster.Spec.CloudProvider.Azure.ResourceGroupName
			azureCloud, err := azure.NewAzureCloud(subscriptionID, resourceGroupName, region, cloudTags)
			if err != nil {
				return nil, err
			}

			cloud = azureCloud
		}
	case kops.CloudProviderScaleway:
		{
			zone, err := scaleway.ParseZoneFromClusterSpec(cluster.Spec)
			if err != nil {
				return nil, fmt.Errorf("error initializing Scaleway cloud: %w", err)
			}
			region, err := scaleway.ParseRegionFromZone(zone)
			if err != nil {
				return nil, fmt.Errorf("error initializing Scaleway cloud: %w", err)
			}

			cloudTags := map[string]string{
				scaleway.TagClusterName: cluster.ObjectMeta.Name,
				"zone":                  string(zone),
				"region":                string(region),
			}
			for k, v := range cluster.Spec.CloudLabels {
				cloudTags[k] = v
			}

			scwCloud, err := scaleway.NewScwCloud(cloudTags)
			if err != nil {
				return nil, fmt.Errorf("error initializing Scaleway cloud: %w", err)
			}

			cloud = scwCloud
		}
	case kops.CloudProviderMetal:
		metalCloud, err := metal.NewCloud()
		if err != nil {
			return nil, fmt.Errorf("error initializing Metal cloud: %w", err)
		}
		cloud = metalCloud
	default:
		return nil, fmt.Errorf("unknown CloudProvider %q", cluster.GetCloudProvider())
	}
	return cloud, nil
}

func FindDNSHostedZone(dns dnsprovider.Interface, clusterDNSName string, dnsType kops.DNSType) (string, error) {
	klog.V(2).Infof("Querying for all DNS zones to find match for %q", clusterDNSName)

	clusterDNSName = "." + strings.TrimSuffix(clusterDNSName, ".")

	zonesProvider, ok := dns.Zones()
	if !ok {
		return "", fmt.Errorf("dns provider %T does not support zones", dns)
	}

	allZones, err := zonesProvider.List()
	if err != nil {
		return "", fmt.Errorf("error querying zones: %v", err)
	}

	var zones []dnsprovider.Zone
	for _, z := range allZones {
		zoneName := "." + strings.TrimSuffix(z.Name(), ".")

		if !strings.HasSuffix(clusterDNSName, zoneName) {
			continue
		}

		if dnsType != "" {
			if awsZone, ok := z.(*route53.Zone); ok {
				hostedZone := awsZone.Route53HostedZone()
				if hostedZone.Config != nil {
					zoneDNSType := kops.DNSTypePublic
					if hostedZone.Config.PrivateZone {
						zoneDNSType = kops.DNSTypePrivate
					}
					if zoneDNSType != dnsType {
						klog.Infof("Found matching hosted zone %q, but it was %q and we require %q", zoneName, zoneDNSType, dnsType)
						continue
					}
				}
			}
		}

		zones = append(zones, z)
	}

	// Find the longest zones
	maxLength := -1
	maxLengthZones := []dnsprovider.Zone{}
	for _, z := range zones {
		zoneName := "." + strings.TrimSuffix(z.Name(), ".")

		n := len(zoneName)
		if n < maxLength {
			continue
		}

		if n > maxLength {
			maxLength = n
			maxLengthZones = []dnsprovider.Zone{}
		}

		maxLengthZones = append(maxLengthZones, z)
	}

	if len(maxLengthZones) == 0 {
		// We make this an error because you have to set up DNS delegation anyway
		tokens := strings.Split(clusterDNSName, ".")
		suffix := strings.Join(tokens[len(tokens)-2:], ".")
		// klog.Warningf("No matching hosted zones found; will created %q", suffix)
		// return suffix, nil
		return "", fmt.Errorf("No matching hosted zones found for %q; please create one (e.g. %q) first", clusterDNSName, suffix)
	}

	if len(maxLengthZones) == 1 {
		id := maxLengthZones[0].ID()
		id = strings.TrimPrefix(id, "/hostedzone/")
		return id, nil
	}

	return "", fmt.Errorf("Found multiple hosted zones matching cluster %q; please specify the ID of the zone to use", clusterDNSName)
}
