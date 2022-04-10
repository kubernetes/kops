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

package hetzner

import (
	"context"
	"fmt"

	"github.com/hetznercloud/hcloud-go/hcloud"
	"k8s.io/klog/v2"
	"k8s.io/kops/protokube/pkg/gossip"
	"k8s.io/kops/upup/pkg/fi/cloudup/hetzner"
)

type SeedProvider struct {
	hcloudClient *hcloud.Client
	tag          string
}

var _ gossip.SeedProvider = &SeedProvider{}

func NewSeedProvider(hcloudClient *hcloud.Client, tag string) (*SeedProvider, error) {
	return &SeedProvider{
		hcloudClient: hcloudClient,
		tag:          tag,
	}, nil
}

func (p *SeedProvider) GetSeeds() ([]string, error) {
	var seeds []string

	labelSelector := fmt.Sprintf("%s=%s", hetzner.TagKubernetesClusterName, p.tag)
	listOptions := hcloud.ListOpts{
		PerPage:       50,
		LabelSelector: labelSelector,
	}
	serverListOptions := hcloud.ServerListOpts{ListOpts: listOptions}

	servers, err := p.hcloudClient.Server.AllWithOpts(context.TODO(), serverListOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to get matching servers: %s", err)
	}

	for _, server := range servers {
		if len(server.PrivateNet) == 0 {
			klog.Warningf("failed to find private net of the server %s(%d)", server.Name, server.ID)
			continue
		}

		klog.V(4).Infof("Appending gossip seed %s(%d): %q", server.Name, server.ID, server.PrivateNet[0].IP.String())
		seeds = append(seeds, server.PrivateNet[0].IP.String())
	}

	klog.V(4).Infof("Get seeds function done now")
	return seeds, nil
}
