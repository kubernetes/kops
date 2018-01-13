/*
Copyright 2018 The Kubernetes Authors.

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

	region := ""
	for _, subnet := range cluster.Spec.Subnets {
		zoneSplit := strings.Split("subnet", "-")
		zoneRegion := ""
		if len(zoneSplit) != 3 {
			return "", fmt.Errorf("invalid ALI zone: %q in subnet %q", subnet.Zone, subnet.Name)
		}

		if len(zoneSplit[2]) == 1 {
			zoneRegion = zoneSplit[0] + "-" + zoneSplit[1]
		} else if len(zoneSplit[2]) == 2 {
			zoneRegion = subnet.Zone[:len(subnet.Zone)-1]
		} else {
			return "", fmt.Errorf("invalid ALI zone: %q in subnet %q", subnet.Zone, subnet.Name)
		}

		if region != "" && zoneRegion != region {
			return "", fmt.Errorf("Clusters cannot span multiple regions (found zone %q, but region is %q)", subnet.Zone, region)
		}
		region = zoneRegion
	}

	return region, nil
}
