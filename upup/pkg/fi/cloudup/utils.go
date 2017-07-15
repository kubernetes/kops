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

package cloudup

import (
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/vsphere"
	"k8s.io/kubernetes/federation/pkg/dnsprovider"
	"k8s.io/kubernetes/federation/pkg/dnsprovider/providers/aws/route53"
	"strings"
)

func BuildCloud(cluster *kops.Cluster) (fi.Cloud, error) {
	var cloud fi.Cloud

	region := ""
	project := ""

	switch cluster.Spec.CloudProvider {
	case "gce":
		{
			nodeZones := make(map[string]bool)
			for _, subnet := range cluster.Spec.Subnets {
				nodeZones[subnet.Zone] = true

				tokens := strings.Split(subnet.Zone, "-")
				if len(tokens) <= 2 {
					return nil, fmt.Errorf("Invalid GCE Zone: %v", subnet.Zone)
				}
				zoneRegion := tokens[0] + "-" + tokens[1]
				if region != "" && zoneRegion != region {
					return nil, fmt.Errorf("Clusters cannot span multiple regions (found zone %q, but region is %q)", subnet.Zone, region)
				}

				region = zoneRegion
			}

			project = cluster.Spec.Project
			if project == "" {
				return nil, fmt.Errorf("project is required for GCE - try gcloud config get-value project")
			}

			labels := map[string]string{gce.GceLabelNameKubernetesCluster: gce.SafeClusterName(cluster.ObjectMeta.Name)}

			gceCloud, err := gce.NewGCECloud(region, project, labels)
			if err != nil {
				return nil, err
			}

			cloud = gceCloud
		}

	case "aws":
		{
			region, err := awsup.FindRegion(cluster)
			if err != nil {
				return nil, err
			}

			err = awsup.ValidateRegion(region)
			if err != nil {
				return nil, err
			}

			cloudTags := map[string]string{awsup.TagClusterName: cluster.ObjectMeta.Name}

			awsCloud, err := awsup.NewAWSCloud(region, cloudTags)
			if err != nil {
				return nil, err
			}

			var zoneNames []string
			for _, subnet := range cluster.Spec.Subnets {
				zoneNames = append(zoneNames, subnet.Zone)
			}
			err = awsup.ValidateZones(zoneNames, awsCloud)
			if err != nil {
				return nil, err
			}
			cloud = awsCloud
		}
	case "vsphere":
		{
			vsphereCloud, err := vsphere.NewVSphereCloud(&cluster.Spec)
			if err != nil {
				return nil, err
			}
			cloud = vsphereCloud
		}
	case "digitalocean":
		{
			return nil, fmt.Errorf("digitalocean not supported yet!")
		}

	default:
		return nil, fmt.Errorf("unknown CloudProvider %q", cluster.Spec.CloudProvider)
	}
	return cloud, nil
}

func FindDNSHostedZone(dns dnsprovider.Interface, clusterDNSName string, dnsType kops.DNSType) (string, error) {
	glog.V(2).Infof("Querying for all DNS zones to find match for %q", clusterDNSName)

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
					if fi.BoolValue(hostedZone.Config.PrivateZone) {
						zoneDNSType = kops.DNSTypePrivate
					}
					if zoneDNSType != dnsType {
						glog.Infof("Found matching hosted zone %q, but it was %q and we require %q", zoneName, zoneDNSType, dnsType)
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
		//glog.Warningf("No matching hosted zones found; will created %q", suffix)
		//return suffix, nil
		return "", fmt.Errorf("No matching hosted zones found for %q; please create one (e.g. %q) first", clusterDNSName, suffix)
	}

	if len(maxLengthZones) == 1 {
		id := maxLengthZones[0].ID()
		id = strings.TrimPrefix(id, "/hostedzone/")
		return id, nil
	}

	return "", fmt.Errorf("Found multiple hosted zones matching cluster %q; please specify the ID of the zone to use", clusterDNSName)
}
