/*
Copyright 2020 The Kubernetes Authors.

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

package aws

import (
	"errors"
	"math/rand"
	"sort"
)

var allZones = []string{
	"ap-northeast-1a",
	"ap-northeast-1c",
	"ap-northeast-1d",
	// AZ does not exist, so we're breaking the 3 AZs per region target here
	//"ap-northeast-2b",
	"ap-northeast-2a",
	"ap-northeast-2c",
	"ap-northeast-2d",
	// Disabled until etcd-manager supports the region and the AMIs used in testing are present
	//"ap-northeast-3a",
	//"ap-northeast-3b",
	//"ap-northeast-3c",
	"ap-south-1a",
	"ap-south-1b",
	"ap-south-1c",
	"ap-southeast-1a",
	"ap-southeast-1b",
	"ap-southeast-1c",
	"ap-southeast-2a",
	"ap-southeast-2b",
	"ap-southeast-2c",
	"ca-central-1a",
	"ca-central-1b",
	"ca-central-1d",
	"eu-central-1a",
	"eu-central-1b",
	"eu-central-1c",
	// Disabled until region limits are increased https://github.com/kubernetes/k8s.io/issues/1921
	//"eu-north-1a",
	//"eu-north-1b",
	//"eu-north-1c",
	"eu-west-1a",
	"eu-west-1b",
	"eu-west-1c",
	"eu-west-2a",
	"eu-west-2b",
	"eu-west-2c",
	"eu-west-3a",
	"eu-west-3b",
	"eu-west-3c",
	"sa-east-1a",
	"sa-east-1b",
	"sa-east-1c",
	"us-east-1a",
	"us-east-1b",
	"us-east-1c",
	"us-east-1d",
	"us-east-1e",
	"us-east-1f",
	"us-east-2a",
	"us-east-2b",
	"us-east-2c",
	"us-west-1a",
	"us-west-1b",
	//"us-west-1c", AZ does not exist, so we're breaking the 3 AZs per region target here
	"us-west-2a",
	"us-west-2b",
	"us-west-2c",
	"us-west-2d",
}

// ErrNoEligibleRegion indicates the requested number of zones is not available in any region
var ErrNoEligibleRegion = errors.New("No eligible AWS region found with enough zones")

// RandomZones returns a random set of availability zones within a region
func RandomZones(count int) ([]string, error) {
	regions := make(map[string][]string)
	for _, zone := range allZones {
		region := zone[:len(zone)-1]
		regions[region] = append(regions[region], zone)
	}
	eligibleRegions := make([][]string, 0)
	for _, zones := range regions {
		if len(zones) >= count {
			eligibleRegions = append(eligibleRegions, zones)
		}
	}
	if len(eligibleRegions) == 0 {
		return nil, ErrNoEligibleRegion
	}
	chosenRegion := eligibleRegions[rand.Int()%len(eligibleRegions)]

	chosenZones := make([]string, 0)
	randIndexes := rand.Perm(len(chosenRegion))
	for i := 0; i < count; i++ {
		chosenZones = append(chosenZones, chosenRegion[randIndexes[i]])
	}
	sort.Strings(chosenZones)
	return chosenZones, nil
}
