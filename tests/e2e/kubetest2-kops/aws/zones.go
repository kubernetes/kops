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
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"math/rand/v2"
	"os"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"k8s.io/klog/v2"
)

var allZones = []string{
	"ap-northeast-1a",
	"ap-northeast-1c",
	"ap-northeast-1d",
	// AZ does not exist, so we're breaking the 3 AZs per region target here
	//"ap-northeast-2b",
	"ap-northeast-2a",
	"ap-northeast-2c",
	// t4g instances are not available
	//"ap-northeast-2d",
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
	// t4g instances are not available
	//"us-east-1e",
	"us-east-1f",
	"us-east-2a",
	"us-east-2b",
	"us-east-2c",
	// Newer accounts don't have us-west-1c and one other zone is constrained so we ignore the entire region
	//"us-west-1a",
	//"us-west-1b",
	//"us-west-1c",
	"us-west-2a",
	"us-west-2b",
	"us-west-2c",
	"us-west-2d",
}

// ErrNoEligibleRegion indicates the requested number of zones is not available in any region
var ErrNoEligibleRegion = errors.New("No eligible AWS region found with enough zones")

// newRand returns a seeded Rand. If the BUILD_ID environment variable is set it
// is used as the seed so that zone selection is deterministic for a given build.
// Otherwise a randomly seeded Rand is returned.
func newRand() *rand.Rand {
	if buildID := os.Getenv("BUILD_ID"); buildID != "" {
		h := fnv.New64a()
		h.Write([]byte(buildID))
		return rand.New(rand.NewPCG(h.Sum64(), 0))
	}
	return rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
}

// RandomZones returns a random set of availability zones within a region,
// ensuring the provided instance types are available in those zones.
func RandomZones(count int, instanceTypes []string) ([]string, error) {
	rng := newRand()
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
	// Sort so that seeded selection is deterministic regardless of map iteration order.
	sort.Slice(eligibleRegions, func(i, j int) bool {
		return eligibleRegions[i][0] < eligibleRegions[j][0]
	})

	// Try regions in a random order so that a single region without the
	// requested instance types does not deterministically fail.
	regionOrder := rng.Perm(len(eligibleRegions))
	var lastErr error
	for _, idx := range regionOrder {
		regionZones := eligibleRegions[idx]
		region := regionZones[0][:len(regionZones[0])-1]

		candidateZones := regionZones
		if len(instanceTypes) > 0 {
			filtered, err := zonesWithInstanceTypes(region, regionZones, instanceTypes)
			if err != nil {
				klog.Warningf("skipping region %s: could not describe instance type offerings: %v", region, err)
				lastErr = err
				continue
			}
			if len(filtered) < count {
				klog.V(2).Infof("region %s has only %d of %d zones offering %v, trying next region", region, len(filtered), count, instanceTypes)
				continue
			}
			candidateZones = filtered
		}

		chosenZones := make([]string, 0, count)
		randIndexes := rng.Perm(len(candidateZones))
		for i := 0; i < count; i++ {
			chosenZones = append(chosenZones, candidateZones[randIndexes[i]])
		}
		sort.Strings(chosenZones)
		return chosenZones, nil
	}

	if lastErr != nil {
		return nil, fmt.Errorf("no eligible region found with instance types %v: %w", instanceTypes, lastErr)
	}
	return nil, fmt.Errorf("no eligible region found with %d zones offering instance types %v", count, instanceTypes)
}

// zonesWithInstanceTypes returns the subset of zones in which every one of
// instanceTypes is offered, according to DescribeInstanceTypeOfferings.
func zonesWithInstanceTypes(region string, zones []string, instanceTypes []string) ([]string, error) {
	ctx := context.Background()
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("loading AWS config for region %s: %w", region, err)
	}
	client := ec2.NewFromConfig(cfg)

	zoneSet := make(map[string]bool, len(zones))
	for _, z := range zones {
		zoneSet[z] = true
	}

	// zoneOfferings[zone] is the set of requested instance types offered in that zone.
	zoneOfferings := make(map[string]map[string]bool, len(zones))
	paginator := ec2.NewDescribeInstanceTypeOfferingsPaginator(client, &ec2.DescribeInstanceTypeOfferingsInput{
		LocationType: ec2types.LocationTypeAvailabilityZone,
		Filters: []ec2types.Filter{
			{
				Name:   aws.String("instance-type"),
				Values: instanceTypes,
			},
			{
				Name:   aws.String("location"),
				Values: zones,
			},
		},
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("describing instance type offerings in %s: %w", region, err)
		}
		for _, offering := range page.InstanceTypeOfferings {
			if offering.Location == nil {
				continue
			}
			zone := *offering.Location
			if !zoneSet[zone] {
				continue
			}
			if zoneOfferings[zone] == nil {
				zoneOfferings[zone] = make(map[string]bool, len(instanceTypes))
			}
			zoneOfferings[zone][string(offering.InstanceType)] = true
		}
	}

	matching := make([]string, 0, len(zones))
	for _, zone := range zones {
		if len(zoneOfferings[zone]) == len(instanceTypes) {
			matching = append(matching, zone)
		}
	}
	return matching, nil
}
