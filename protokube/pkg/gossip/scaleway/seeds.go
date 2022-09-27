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

package scaleway

import (
	"fmt"

	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"k8s.io/klog/v2"
	"k8s.io/kops/protokube/pkg/gossip"
	"k8s.io/kops/upup/pkg/fi/cloudup/scaleway"
)

type SeedProvider struct {
	scwClient *scw.Client
	tag       string
}

var _ gossip.SeedProvider = &SeedProvider{}

func NewSeedProvider(scwClient *scw.Client, clusterName string) (*SeedProvider, error) {
	return &SeedProvider{
		scwClient: scwClient,
		tag:       clusterName,
	}, nil
}

func (p *SeedProvider) GetSeeds() ([]string, error) {
	var seeds []string

	instanceAPI := instance.NewAPI(p.scwClient)
	zone, ok := p.scwClient.GetDefaultZone()
	if !ok {
		return nil, fmt.Errorf("could not determine default region from client")
	}
	servers, err := instanceAPI.ListServers(&instance.ListServersRequest{
		Zone: zone,
		Tags: []string{fmt.Sprintf("%s=%s", scaleway.TagClusterName, p.tag)},
	}, scw.WithAllPages())
	if err != nil {
		return nil, fmt.Errorf("failed to get matching servers: %s", err)
	}

	for _, server := range servers.Servers {
		if server.PrivateIP == nil || *server.PrivateIP == "" {
			klog.Warningf("failed to find private ip of the server %s(%s)", server.Name, server.ID)
			continue
		}

		klog.V(4).Infof("Appending gossip seed %s(%s): %q", server.Name, server.ID, *server.PrivateIP)
		seeds = append(seeds, *server.PrivateIP)
	}

	klog.V(4).Infof("Get seeds function done now")
	return seeds, nil
}
