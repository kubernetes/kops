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
	"ap-northeast-2a",
	//"ap-northeast-2b" - AZ does not exist, so we"re breaking the 3 AZs per region target here
	"ap-northeast-2c",
	"ap-south-1a",
	"ap-south-1b",
	"ap-southeast-1a",
	"ap-southeast-1b",
	"ap-southeast-1c",
	"ap-southeast-2a",
	"ap-southeast-2b",
	"ap-southeast-2c",
	"eu-central-1a",
	"eu-central-1b",
	"eu-central-1c",
	"eu-west-1a",
	"eu-west-1b",
	"eu-west-1c",
	"eu-west-2a",
	"eu-west-2b",
	"eu-west-2c",
	//"eu-west-3a", documented to not support c4 family
	//"eu-west-3b", documented to not support c4 family
	//"eu-west-3c", documented to not support c4 family
	//"us-east-1a", // temporarily removing due to lack of quota test-infra#10043
	//"us-east-1b", // temporarily removing due to lack of quota test-infra#10043
	//"us-east-1c", // temporarily removing due to lack of quota test-infra#10043
	//"us-east-1d", // limiting to 3 zones to not overallocate
	//"us-east-1e", // limiting to 3 zones to not overallocate
	//"us-east-1f", // limiting to 3 zones to not overallocate
	//"us-east-2a", InsufficientInstanceCapacity for c4.large 2018-05-30
	//"us-east-2b", InsufficientInstanceCapacity for c4.large 2018-05-30
	//"us-east-2c", InsufficientInstanceCapacity for c4.large 2018-05-30
	"us-west-1a",
	"us-west-1b",
	//"us-west-1c", AZ does not exist, so we"re breaking the 3 AZs per region target here
	//"us-west-2a", // temporarily removing due to lack of quota test-infra#10043
	//"us-west-2b", // temporarily removing due to lack of quota test-infra#10043
	//"us-west-2c", // temporarily removing due to lack of quota test-infra#10043
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
