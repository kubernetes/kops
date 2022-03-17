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

package do

import (
	"errors"
	"math/rand"
	"time"
)

var allZones = []string{
	"nyc1",
	"nyc3",
	"sfo3",
	"tor1",
	"lon1",
	"sgp1",
	"blr1",
	"sfo3",
}

// ErrNoEligibleRegion indicates the requested number of zones is not available in any region
var ErrNoEligibleRegion = errors.New("No eligible DO region found with enough zones")

// ErrMoreThanOneZone indicates the requested number of zones was more than one
var ErrMoreThanOneZone = errors.New("More than 1 zone is chosen. DO only works with 1 zone")

// RandomZones returns a random set of availability zones within a region
func RandomZones(count int) ([]string, error) {
	if count > 1 {
		return nil, ErrMoreThanOneZone
	}

	rand.Seed(time.Now().UnixNano())
	n := rand.Intn(1000) % len(allZones)
	chosenZone := allZones[n]

	chosenZones := make([]string, 0)
	chosenZones = append(chosenZones, chosenZone)

	return chosenZones, nil
}
