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
	"net"
	"strings"

	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"k8s.io/klog/v2"
	kopsv "k8s.io/kops"
	"k8s.io/kops/protokube/pkg/gossip"
	gossipscw "k8s.io/kops/protokube/pkg/gossip/scaleway"
	"k8s.io/kops/upup/pkg/fi/cloudup/scaleway"
)

// ScwCloudProvider defines the Scaleway Cloud volume implementation.
type ScwCloudProvider struct {
	scwClient *scw.Client
	server    *instance.Server
	serverIP  net.IP
}

var _ CloudProvider = &ScwCloudProvider{}

// NewScwCloudProvider returns a new Scaleway Cloud volume provider.
func NewScwCloudProvider() (*ScwCloudProvider, error) {
	metadataAPI := instance.NewMetadataAPI()
	metadata, err := metadataAPI.GetMetadata()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve server metadata: %s", err)
	}

	serverID := metadata.ID
	klog.V(4).Infof("Found ID of the running server: %v", serverID)

	zoneID := metadata.Location.ZoneID
	zone, err := scw.ParseZone(zoneID)
	if err != nil {
		return nil, fmt.Errorf("unable to parse Scaleway zone: %s", err)
	}
	klog.V(4).Infof("Found zone of the running server: %v", zone)

	privateIP := metadata.PrivateIP
	klog.V(4).Infof("Found first private net IP of the running server: %q", privateIP)

	scwClient, err := scw.NewClient(
		scw.WithUserAgent(scaleway.KopsUserAgentPrefix+kopsv.Version),
		scw.WithEnv(),
		scw.WithDefaultZone(zone),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating client: %w", err)
	}

	instanceAPI := instance.NewAPI(scwClient)
	server, err := instanceAPI.GetServer(&instance.GetServerRequest{
		ServerID: serverID,
		Zone:     zone,
	})
	if err != nil || server == nil {
		return nil, fmt.Errorf("failed to get the running server: %s", err)
	}
	klog.V(4).Infof("Found the running server: %q", server.Server.Name)

	s := &ScwCloudProvider{
		scwClient: scwClient,
		server:    server.Server,
		serverIP:  net.IP(privateIP),
	}

	return s, nil
}

func (s *ScwCloudProvider) InstanceID() string {
	return fmt.Sprintf("%s-%s", s.server.Name, s.server.ID)
}

func (s ScwCloudProvider) InstanceInternalIP() net.IP {
	return s.serverIP
}

func (s *ScwCloudProvider) GossipSeeds() (gossip.SeedProvider, error) {
	for _, tag := range s.server.Tags {
		if !strings.HasPrefix(tag, scaleway.TagClusterName) {
			continue
		}
		clusterName := strings.TrimPrefix(tag, scaleway.TagClusterName+"=")
		return gossipscw.NewSeedProvider(s.scwClient, clusterName)
	}
	return nil, fmt.Errorf("failed to find cluster name label for running server: %v", s.server.Tags)
}
