/*
Copyright 2024 The Kubernetes Authors.

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
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/klog/v2"
)

type zoneDetails struct {
	name     string
	id       string
	zoneType string
}

type zoneCache struct {
	cloud             *Cloud
	mutex             sync.RWMutex
	zoneNameToDetails map[string]zoneDetails
}

func (z *zoneCache) getZoneIDByZoneName(zoneName string) (string, error) {
	zoneNameToDetails, err := z.getZoneDetailsByNames([]string{zoneName})
	if err != nil {
		return "", err
	}

	zoneDetail, ok := zoneNameToDetails[zoneName]
	if !ok {
		return "", fmt.Errorf("Could not get zone ID from zone name %s", zoneName)
	}

	return zoneDetail.id, nil
}

// Get the zone details by zone names and load from the cache if available as
// zone information should never change.
func (z *zoneCache) getZoneDetailsByNames(zoneNames []string) (map[string]zoneDetails, error) {
	if len(zoneNames) == 0 {
		return map[string]zoneDetails{}, nil
	}

	if z.shouldPopulateCache(zoneNames) {
		// Populate the cache if it hasn't been populated yet
		err := z.populate()
		if err != nil {
			return nil, err
		}
	}

	z.mutex.RLock()
	defer z.mutex.RUnlock()

	requestedZoneDetails := map[string]zoneDetails{}
	for _, zone := range zoneNames {
		if zoneDetails, ok := z.zoneNameToDetails[zone]; ok {
			requestedZoneDetails[zone] = zoneDetails
		} else {
			klog.Warningf("Could not find zone %s", zone)
		}
	}

	return requestedZoneDetails, nil
}

func (z *zoneCache) shouldPopulateCache(zoneNames []string) bool {
	z.mutex.RLock()
	defer z.mutex.RUnlock()

	if len(z.zoneNameToDetails) == 0 {
		// Populate the cache if it hasn't been populated yet
		return true
	}

	// Make sure that we know about all of the AZs we're looking for.
	for _, zone := range zoneNames {
		if _, ok := z.zoneNameToDetails[zone]; !ok {
			klog.Infof("AZ %s not found in zone cache.", zone)
			return true
		}
	}

	return false
}

// Populates the zone cache. If cache is already populated, it will overwrite entries,
// which is useful when accounts get access to new zones.
func (z *zoneCache) populate() error {
	z.mutex.Lock()
	defer z.mutex.Unlock()

	azRequest := &ec2.DescribeAvailabilityZonesInput{}
	zones, err := z.cloud.ec2.DescribeAvailabilityZones(azRequest)
	if err != nil {
		return fmt.Errorf("error describe availability zones: %q", err)
	}

	// Initialize the map if it's unset
	if z.zoneNameToDetails == nil {
		z.zoneNameToDetails = make(map[string]zoneDetails)
	}

	for _, zone := range zones {
		name := aws.StringValue(zone.ZoneName)
		z.zoneNameToDetails[name] = zoneDetails{
			name:     name,
			id:       aws.StringValue(zone.ZoneId),
			zoneType: aws.StringValue(zone.ZoneType),
		}
	}

	return nil
}
