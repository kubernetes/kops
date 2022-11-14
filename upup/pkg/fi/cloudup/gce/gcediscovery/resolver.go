/*
Copyright 2017 The Kubernetes Authors.

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

package gcediscovery

import (
	"context"
	"fmt"
	"strings"

	"cloud.google.com/go/compute/metadata"
	compute "google.golang.org/api/compute/v1"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/resolver"
	"k8s.io/kops/protokube/pkg/gossip"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce/gcemetadata"
)

type Discovery struct {
	compute          *compute.Service
	projectID        string
	region           string
	zone             string
	clusterName      string
	allZonesInRegion []string
}

var _ gossip.SeedProvider = &Discovery{}

func (r *Discovery) GetSeeds() ([]string, error) {
	var seeds []string

	// We are only finding seeds here; we don't need every result.
	const maxResults = 100

	ctx := context.TODO()

	if err := r.findInstances(ctx, func(i *compute.Instance) (bool, error) {
		// TODO: Expose multiple IPs topologies?

		for _, ni := range i.NetworkInterfaces {
			// TODO: Check e.g. Network

			if ni.NetworkIP != "" {
				seeds = append(seeds, ni.NetworkIP)
			}
		}

		if len(seeds) >= maxResults {
			return false, nil
		}
		return true, nil
	}); err != nil {
		return nil, err
	}

	return seeds, nil
}

func (r *Discovery) findInstances(ctx context.Context, callback func(*compute.Instance) (bool, error)) error {
	// TODO: Does it suffice to just query one zone (as long as we sort so it is always the first)?
	// Or does that introduce edges cases where we have partitions / cliques

	for _, zoneName := range r.allZonesInRegion {
		pageToken := ""
		for {
			listCall := r.compute.Instances.List(r.projectID, zoneName).Context(ctx)

			// TODO: Filter by tags (but doesn't seem to be possible)
			// TODO: Restrict the fields returned (but be sure to include nextPageToken!)

			if pageToken != "" {
				listCall.PageToken(pageToken)
			}

			res, err := listCall.Do()
			if err != nil {
				return err
			}
			pageToken = res.NextPageToken
			for _, i := range res.Items {
				if !gcemetadata.InstanceMatchesClusterName(r.clusterName, i) {
					continue
				}

				keepGoing, err := callback(i)
				if err != nil {
					return err
				}
				if !keepGoing {
					// We immediately stop (even if there are still zones to visit)
					return nil
				}
			}

			if pageToken == "" {
				break
			}
		}
	}

	return nil
}

// ProjectID returns the GCP project ID we are running in.
func (r *Discovery) ProjectID() string {
	return r.projectID
}

// Zone returns the GCP zone we are running in (e.g. us-central-1a).
func (r *Discovery) Zone() string {
	return r.zone
}

// Region returns the GCP region we are running in (e.g. us-central-1).
func (r *Discovery) Region() string {
	return r.region
}

// ClusterName returns the kOps cluster-name we are part of.
func (r *Discovery) ClusterName() string {
	return r.clusterName
}

// Compute returns the GCP compute service we built.
func (r *Discovery) Compute() *compute.Service {
	return r.compute
}

// New builds a Discovery.
func New() (*Discovery, error) {
	ctx := context.Background()

	computeService, err := compute.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("error building compute API client: %v", err)
	}

	myZoneName, err := metadata.Zone()
	if err != nil {
		return nil, fmt.Errorf("failed to get zone from metadata: %w", err)
	}

	projectID, err := metadata.ProjectID()
	if err != nil {
		return nil, fmt.Errorf("failed to get project id from metadata: %w", err)
	}
	projectID = strings.TrimSpace(projectID)

	zones, err := computeService.Zones.List(projectID).Do()
	if err != nil {
		return nil, fmt.Errorf("error querying for GCE zones: %w", err)
	}

	// Find our zone
	var myZone *compute.Zone
	for _, zone := range zones.Items {
		if myZoneName == zone.Name {
			myZone = zone
		}
	}
	if myZone == nil {
		return nil, fmt.Errorf("failed to find zone %q", myZoneName)
	}

	region := lastComponent(myZone.Region)

	// Find all the zones in our region
	var zoneNames []string
	for _, zone := range zones.Items {
		regionName := lastComponent(zone.Region)
		if regionName != region {
			continue
		}
		zoneNames = append(zoneNames, zone.Name)
	}

	clusterName, err := metadata.InstanceAttributeValue(gcemetadata.MetadataKeyClusterName)
	if err != nil {
		return nil, fmt.Errorf("error reading cluster-name attribute from GCE: %w", err)
	}
	clusterName = strings.TrimSpace(clusterName)
	if clusterName == "" {
		return nil, fmt.Errorf("cluster-name metadata was empty")
	}
	klog.Infof("Found cluster-name=%q", clusterName)

	return &Discovery{
		compute:          computeService,
		region:           region,
		projectID:        projectID,
		zone:             myZoneName,
		clusterName:      clusterName,
		allZonesInRegion: zoneNames,
	}, nil
}

// Returns the last component of a URL, i.e. anything after the last slash
// If there is no slash, returns the whole string
func lastComponent(s string) string {
	lastSlash := strings.LastIndex(s, "/")
	if lastSlash != -1 {
		s = s[lastSlash+1:]
	}
	return s
}

var _ resolver.Resolver = &Discovery{}

// Resolve implements resolver.Resolve, providing name -> address resolution using GCE discovery.
func (r *Discovery) Resolve(ctx context.Context, name string) ([]string, error) {
	var records []string
	klog.Infof("trying to resolve %q using GCEResolver", name)

	var requiredTags []string

	// We assume we are trying to resolve a component that runs on the control plane
	requiredTags = append(requiredTags, gce.TagForRole(r.clusterName, kops.InstanceGroupRoleControlPlane))

	if err := r.findInstances(ctx, func(i *compute.Instance) (bool, error) {
		// Make sure the instance has any required tags
		for _, requiredTag := range requiredTags {
			hasTag := false
			if i.Tags != nil {
				for _, tag := range i.Tags.Items {
					if requiredTag == tag {
						hasTag = true
					}
				}
			}
			if !hasTag {
				return true, nil
			}
		}

		// TODO: Expose multiple IPs topologies?
		for _, ni := range i.NetworkInterfaces {
			// TODO: Check e.g. Network

			if ni.NetworkIP != "" {
				records = append(records, ni.NetworkIP)
			}
		}

		return true, nil
	}); err != nil {
		return nil, err
	}

	return records, nil
}
