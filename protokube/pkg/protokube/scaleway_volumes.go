/*
Copyright 2022 The Kubernetes Authors.

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

package protokube

import (
	"fmt"

	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	ipam "github.com/scaleway/scaleway-sdk-go/api/ipam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"k8s.io/klog/v2"
	kopsv "k8s.io/kops"
	"k8s.io/kops/protokube/pkg/gossip"
	gossipscw "k8s.io/kops/protokube/pkg/gossip/scaleway"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/scaleway"
)

// ScwCloudProvider defines the Scaleway Cloud volume implementation.
type ScwCloudProvider struct {
	scwClient *scw.Client
	server    *instance.Server
}

var _ CloudProvider = &ScwCloudProvider{}

// NewScwCloudProvider returns a new Scaleway Cloud volume provider.
func NewScwCloudProvider() (*ScwCloudProvider, error) {
	metadataAPI := instance.NewMetadataAPI()
	metadata, err := metadataAPI.GetMetadata()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve server metadata: %w", err)
	}

	serverID := metadata.ID
	klog.V(4).Infof("Found ID of the running server: %v", serverID)

	zone, err := scw.ParseZone(metadata.Location.ZoneID)
	if err != nil {
		return nil, fmt.Errorf("unable to parse Scaleway zone: %w", err)
	}
	klog.V(4).Infof("Found zone of the running server: %v", zone)

	region, err := zone.Region()
	if err != nil {
		return nil, fmt.Errorf("unable to parse Scaleway region: %w", err)
	}
	klog.V(4).Infof("Found region of the running server: %v", region)

	profile, err := scaleway.CreateValidScalewayProfile()
	if err != nil {
		return nil, err
	}
	scwClient, err := scw.NewClient(
		scw.WithProfile(profile),
		scw.WithUserAgent(scaleway.KopsUserAgentPrefix+kopsv.Version),
		scw.WithDefaultZone(zone),
		scw.WithDefaultRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("creating client for Protokube: %w", err)
	}

	instanceAPI := instance.NewAPI(scwClient)
	serverResponse, err := instanceAPI.GetServer(&instance.GetServerRequest{
		ServerID: serverID,
		Zone:     zone,
	})
	if err != nil || serverResponse.Server == nil {
		return nil, fmt.Errorf("failed to get the running server: %w", err)
	}
	server := serverResponse.Server
	klog.V(4).Infof("Found the running server: %q", server.Name)

	ips, err := ipam.NewAPI(scwClient).ListIPs(&ipam.ListIPsRequest{
		Region:     region,
		ResourceID: fi.PtrTo(serverID),
		IsIPv6:     fi.PtrTo(false),
		Zonal:      fi.PtrTo(zone.String()),
	}, scw.WithAllPages())
	if err != nil {
		return nil, fmt.Errorf("listing server's IPs: %w", err)
	}
	if ips.TotalCount < 1 {
		return nil, fmt.Errorf("expected at least 1 IP attached to the server %s", server.ID)
	}

	s := &ScwCloudProvider{
		scwClient: scwClient,
		server:    server,
	}

	return s, nil
}

func (s *ScwCloudProvider) InstanceID() string {
	return fmt.Sprintf("%s-%s", s.server.Name, s.server.ID)
}

func (s *ScwCloudProvider) GossipSeeds() (gossip.SeedProvider, error) {
	clusterName := scaleway.ClusterNameFromTags(s.server.Tags)
	if clusterName != "" {
		return gossipscw.NewSeedProvider(s.scwClient, clusterName)
	}
	return nil, fmt.Errorf("failed to find cluster name label for running server: %v", s.server.Tags)
}
