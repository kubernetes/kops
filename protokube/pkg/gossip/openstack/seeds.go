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

package gce

import (
	"context"
	"fmt"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/pagination"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/resolver"
	"k8s.io/kops/protokube/pkg/gossip"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
)

type SeedProvider struct {
	computeClient *gophercloud.ServiceClient
	projectID     string
	clusterName   string
}

var _ gossip.SeedProvider = &SeedProvider{}

func (p *SeedProvider) GetSeeds() ([]string, error) {
	predicate := func(*servers.Server) bool {
		return true
	}
	return p.discover(context.TODO(), predicate)
}

func (p *SeedProvider) discover(ctx context.Context, predicate func(*servers.Server) bool) ([]string, error) {
	var seeds []string

	err := servers.List(p.computeClient, servers.ListOpts{
		TenantID: p.projectID,
	}).EachPage(func(page pagination.Page) (bool, error) {
		var s []servers.Server
		err := servers.ExtractServersInto(page, &s)
		if err != nil {
			return false, err
		}

		for _, server := range s {
			if !predicate(&server) {
				continue
			}

			if clusterName, ok := server.Metadata[openstack.TagClusterName]; ok {
				// verify that the instance is from the same cluster
				if clusterName != p.clusterName {
					continue
				}

				var err error
				// find kopsNetwork from metadata, fallback to clustername
				ifName := clusterName
				if val, ok := server.Metadata[openstack.TagKopsNetwork]; ok {
					ifName = val
				}
				addr, err := openstack.GetServerFixedIP(&server, ifName)
				if err != nil {
					klog.Warningf("Failed to list seeds: %v", err)
					continue
				}
				seeds = append(seeds, addr)
			}
		}
		return true, nil
	})
	if err != nil {
		return seeds, fmt.Errorf("Failed to list servers while retrieving seeds: %v", err)
	}

	return seeds, nil
}

func NewSeedProvider(computeClient *gophercloud.ServiceClient, clusterName string, projectID string) (*SeedProvider, error) {
	return &SeedProvider{
		computeClient: computeClient,
		clusterName:   clusterName,
		projectID:     projectID,
	}, nil
}

var _ resolver.Resolver = &SeedProvider{}

// Resolve implements resolver.Resolve, providing name -> address resolution using cloud API discovery.
func (p *SeedProvider) Resolve(ctx context.Context, name string) ([]string, error) {
	klog.Infof("trying to resolve %q using SeedProvider", name)

	// TODO: Can we push this predicate down so we can filter server-side?
	// We assume we are trying to resolve a component that runs on the control plane
	isControlPlane := func(server *servers.Server) bool {
		switch server.Metadata[openstack.TagKopsRole] {
		case kops.InstanceGroupRoleControlPlane.ToLowerString():
			return true
		case string(kops.InstanceGroupRoleControlPlane):
			return true
		case "master":
			return true
		default:
			return false
		}
	}
	return p.discover(ctx, isControlPlane)
}
