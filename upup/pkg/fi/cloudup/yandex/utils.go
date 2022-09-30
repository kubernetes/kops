/*
Copyright 2022 The Kubernetes Authors.

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

package yandex

import (
	"fmt"

	"k8s.io/kops/pkg/apis/kops"
)

// FindRegion determines the region from the zones specified in the cluster
func FindRegion(cluster *kops.Cluster) (string, error) {
	region := ""

	nodeZones := make(map[string]bool)
	for _, subnet := range cluster.Spec.Subnets {
		// zone name require to be full like ru-central1-a
		if len(subnet.Zone) <= 1 {
			return "", fmt.Errorf("invalid Yandex.Cloud zone: %q in subnet %q", subnet.Zone, subnet.Name)
		}

		nodeZones[subnet.Zone] = true

		// region-zone like ru-central1-a
		zoneRegion := subnet.Zone[:len(subnet.Zone)-2]
		if region != "" && zoneRegion != region {
			return "", fmt.Errorf("error Clusters cannot span multiple regions (found zone %q, but region is %q)", subnet.Zone, region)
		}

		region = zoneRegion
	}

	return region, nil
}
