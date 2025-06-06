/*
Copyright 2025 The Kubernetes Authors.

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

package elemento

import (
	"fmt"

	"k8s.io/kops/pkg/apis/kops"
)

// FindRegion determines the region from the zones specified in the cluster
func FindRegion(cluster *kops.Cluster) (string, error) {
	var region string
	// Elemento intends the zone parameter to be used as a continent in order
	// to let elemento systems determine the best region/zones for the cluster
	// based on parameters like specs and price.
	// Elemento also supports deploy of cluster on multiple regions and providers
	// Supported zones are: asia, europe, north-america, south-america, africa,
	// oceania, antartica ;)
	
	for _, subnet := range cluster.Spec.Networking.Subnets {
		var zoneRegion string
		if subnet.Zone == "asia" || 
		   subnet.Zone == "europe" || 
		   subnet.Zone == "north-america" || 
		   subnet.Zone == "south-america" || 
		   subnet.Zone == "africa" || 
		   subnet.Zone == "oceania" || 
		   subnet.Zone == "antartica" {
			// TODO: check if it is works out
			zoneRegion = subnet.Zone
		} else {
			return "", fmt.Errorf("unknown zone %q for elemento cloud, known zones are asia, europe, north-america, south-america, africa, oceania, antartica", subnet.Zone)
		}

		region = zoneRegion
	}

	return region, nil
}
