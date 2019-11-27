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

package aliup

import (
	"fmt"
	"strings"

	"k8s.io/kops/pkg/apis/kops"
)

// FindRegion determines the region from the zones specified in the cluster
func FindRegion(cluster *kops.Cluster) (string, error) {

	zones := []string{}
	for _, subnet := range cluster.Spec.Subnets {
		zones = append(zones, subnet.Zone)
	}
	return getRegionByZones(zones)

}

func getRegionByZones(zones []string) (string, error) {
	region := ""

	for _, zone := range zones {
		zoneSplit := strings.Split(zone, "-")
		zoneRegion := ""
		if len(zoneSplit) != 3 {
			return "", fmt.Errorf("invalid ALI zone: %q ", zone)
		}

		if len(zoneSplit[2]) == 1 {
			zoneRegion = zoneSplit[0] + "-" + zoneSplit[1]
		} else if len(zoneSplit[2]) == 2 {
			zoneRegion = zone[:len(zone)-1]
		} else {
			return "", fmt.Errorf("invalid ALI zone: %q ", zone)
		}

		if region != "" && zoneRegion != region {
			return "", fmt.Errorf("clusters cannot span multiple regions (found zone %q, but region is %q)", zone, region)
		}
		region = zoneRegion
	}

	return region, nil
}
