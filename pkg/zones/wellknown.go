/*
Copyright 2021 The Kubernetes Authors.

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

package zones

import (
	"sort"

	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

// These lists allow us to infer from certain well-known zones to a cloud
// Note it is safe to "overmap" zones that don't exist: we'll check later if the zones actually exist

var gceZones = []string{
	"asia-east1-a",
	"asia-east1-b",
	"asia-east1-c",
	"asia-east1-d",

	"asia-east2-a",
	"asia-east2-b",
	"asia-east2-c",

	"asia-northeast1-a",
	"asia-northeast1-b",
	"asia-northeast1-c",
	"asia-northeast1-d",

	"asia-northeast2-a",
	"asia-northeast2-b",
	"asia-northeast2-c",

	"asia-northeast3-a",
	"asia-northeast3-b",
	"asia-northeast3-c",

	"asia-south1-a",
	"asia-south1-b",
	"asia-south1-c",

	"asia-southeast1-a",
	"asia-southeast1-b",

	"asia-southeast2-a",
	"asia-southeast2-b",
	"asia-southeast2-c",

	"australia-southeast1-a",
	"australia-southeast1-b",
	"australia-southeast1-c",

	"europe-north1-a",
	"europe-north1-b",
	"europe-north1-c",

	"europe-west1-a",
	"europe-west1-b",
	"europe-west1-c",
	"europe-west1-d",
	"europe-west1-e",

	"europe-west2-a",
	"europe-west2-b",
	"europe-west2-c",

	"europe-west3-a",
	"europe-west3-b",
	"europe-west3-c",

	"europe-west4-a",
	"europe-west4-b",
	"europe-west4-c",

	"europe-west6-a",
	"europe-west6-b",
	"europe-west6-c",

	"us-central1-a",
	"us-central1-b",
	"us-central1-c",
	"us-central1-d",
	"us-central1-e",
	"us-central1-f",
	"us-central1-g",
	"us-central1-h",

	"us-east1-a",
	"us-east1-b",
	"us-east1-c",
	"us-east1-d",

	"us-east4-a",
	"us-east4-b",
	"us-east4-c",

	"us-west1-a",
	"us-west1-b",
	"us-west1-c",
	"us-west1-d",

	"us-west2-a",
	"us-west2-b",
	"us-west2-c",

	"us-west3-a",
	"us-west3-b",
	"us-west3-c",

	"us-west4-a",
	"us-west4-b",
	"us-west4-c",

	"northamerica-northeast1-a",
	"northamerica-northeast1-b",
	"northamerica-northeast1-c",

	"southamerica-east1-a",
	"southamerica-east1-b",
	"southamerica-east1-c",
}

var doZones = []string{
	"nyc1",
	"nyc3",

	"sfo3",

	"ams3",

	"tor1",

	"sgp1",

	"lon1",

	"fra1",

	"blr1",
}

var hetznerZones = []string{
	// eu-central
	"fsn1",
	"nbg1",
	"hel1",
	// us-east
	"ash",
}

var azureZones = []string{
	"asia",
	"asiapacific",
	"australia",
	"australiacentral",
	"australiacentral2",
	"australiaeast",
	"australiasoutheast",
	"brazil",
	"brazilsouth",
	"brazilsoutheast",
	"canada",
	"canadacentral",
	"canadaeast",
	"centralindia",
	"centralus",
	"centraluseuap",
	"centralusstage",
	"eastasia",
	"eastasiastage",
	"eastus",
	"eastus2",
	"eastus2euap",
	"eastus2stage",
	"eastusstage",
	"europe",
	"francecentral",
	"francesouth",
	"germanynorth",
	"germanywestcentral",
	"global",
	"india",
	"japan",
	"japaneast",
	"japanwest",
	"koreacentral",
	"koreasouth",
	"northcentralus",
	"northcentralusstage",
	"northeurope",
	"norwayeast",
	"norwaywest",
	"southafricanorth",
	"southafricawest",
	"southcentralus",
	"southcentralusstage",
	"southeastasia",
	"southeastasiastage",
	"southindia",
	"switzerlandnorth",
	"switzerlandwest",
	"uaecentral",
	"uaenorth",
	"uk",
	"uksouth",
	"ukwest",
	"unitedstates",
	"westcentralus",
	"westeurope",
	"westindia",
	"westus",
	"westus2",
	"westus2stage",
	"westusstage",
}

func WellKnownZonesForCloud(matchCloud kops.CloudProviderID, prefix string) []string {
	var found []string
	switch matchCloud {
	case kops.CloudProviderAWS:
		prefix = strings.ToLower(prefix)
		for _, partition := range endpoints.DefaultResolver().(endpoints.EnumPartitions).Partitions() {
			for _, region := range partition.Regions() {
				regionName := strings.ToLower(region.ID())
				if prefix == regionName || strings.HasPrefix(prefix, regionName+"-") {
					// If the prefix is a region name or a Local Zone or a Wavelength Zone,
					// return all its matching zones as the completion options.
					awsCloud, err := awsup.NewAWSCloud(regionName, map[string]string{})
					if err != nil {
						continue
					}
					var zones *ec2.DescribeAvailabilityZonesOutput
					zones, err = awsCloud.EC2().DescribeAvailabilityZones(&ec2.DescribeAvailabilityZonesInput{
						AllAvailabilityZones: aws.Bool(true),
					})
					if err != nil {
						continue
					}
					for _, zone := range zones.AvailabilityZones {
						found = append(found, *zone.ZoneName)
					}
				} else if strings.HasPrefix(regionName, prefix) {
					// Return the region name as the completion option. After the user completes
					// that much, the code will then look up the specific zone options.
					found = append(found, regionName)
				} else {
					// If the zone name is in the form of single-letter zones
					// belonging to a region, that's good enough.
					if len(prefix) == len(regionName)+1 && strings.HasPrefix(prefix, regionName) {
						found = append(found, prefix)
					}
				}
			}
		}
	case kops.CloudProviderAzure:
		found = azureZones
	case kops.CloudProviderDO:
		found = doZones
	case kops.CloudProviderGCE:
		found = gceZones
	case kops.CloudProviderHetzner:
		found = hetznerZones
	default:
		return nil
	}

	sort.Strings(found)
	return found
}
