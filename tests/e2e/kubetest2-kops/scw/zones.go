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

package scaleway

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func zoneIsInZones(zone scw.Zone, zones []scw.Zone) bool {
	for _, z := range zones {
		if zone == z {
			return true
		}
	}
	return false
}

// RandomZones returns a random availability zone among the ones where the products needed are available
func RandomZones(count int) ([]string, error) {
	if count > 1 {
		return nil, fmt.Errorf("expected 1 zone, got %d", count)
	}

	instanceAPI := &instance.API{}
	instanceAPIZones := instanceAPI.Zones()
	lbAPI := &lb.ZonedAPI{}
	lbAPIZones := lbAPI.Zones()

	allAvailableZones := []scw.Zone(nil)
	for _, zone := range append(instanceAPIZones, lbAPIZones...) {
		if !zoneIsInZones(zone, allAvailableZones) && zoneIsInZones(zone, instanceAPIZones) && zoneIsInZones(zone, lbAPIZones) {
			allAvailableZones = append(allAvailableZones, zone)
		}
	}

	rand.Seed(time.Now().UnixNano())
	n := rand.Intn(1000) % len(allAvailableZones)
	chosenZone := allAvailableZones[n]
	chosenZones := []string{chosenZone.String()}

	return chosenZones, nil
}
