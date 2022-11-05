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

package gce

import (
	"errors"
	"math/rand"
	"sort"
)

var allZones = []string{
	// Starting with the us- zones (we can add the others later)
	"us-central1-a",
	"us-central1-b",
	"us-central1-c",
	"us-central1-f",
	"us-east1-b",
	"us-east1-c",
	"us-east1-d",
	"us-east4-a",
	"us-east4-b",
	"us-east4-c",
	"us-west1-a",
	"us-west1-b",
	"us-west1-c",
	"us-west2-a",
	"us-west2-b",
	"us-west2-c",
	"us-west3-a",
	"us-west3-b",
	"us-west3-c",
	"us-west4-a",
	"us-west4-b",
	"us-west4-c",
}

// ErrNoEligibleRegion indicates the requested number of zones is not available in any region
var ErrNoEligibleRegion = errors.New("No eligible GCP region found with enough zones")

// RandomZones returns a random set of availability zones within a region
func RandomZones(count int) ([]string, error) {
	regions := make(map[string][]string)
	for _, zone := range allZones {
		region := zone[:len(zone)-2]
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
