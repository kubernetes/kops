/*
Copyright 2019 The Kubernetes Authors.

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
	"context"
	"fmt"
	"strings"

	"github.com/digitalocean/godo"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/resolver"
	"k8s.io/kops/protokube/pkg/gossip"
	"k8s.io/kops/upup/pkg/fi/cloudup/do"
)

type SeedProvider struct {
	godoClient *godo.Client
	tag        string
}

var _ gossip.SeedProvider = &SeedProvider{}

func (p *SeedProvider) GetSeeds() ([]string, error) {
	predicate := func(*godo.Droplet) bool {
		return true
	}
	return p.discover(context.TODO(), predicate)
}

func (p *SeedProvider) discover(ctx context.Context, predicate func(*godo.Droplet) bool) ([]string, error) {
	var seeds []string

	droplets, _, err := p.godoClient.Droplets.List(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("Droplets.ListByTag returned error: %w", err)
	}

	for _, droplet := range droplets {
		if !predicate(&droplet) {
			continue
		}

		for _, dropTag := range droplet.Tags {
			klog.V(4).Infof("Get Seeds - droplet found=%s,SeedProvider Tag=%s", dropTag, p.tag)
			if strings.Contains(dropTag, strings.Replace(p.tag, ".", "-", -1)) {
				klog.V(4).Infof("Tag matched for droplet tag =%s. Getting private IP", p.tag)
				ip, err := droplet.PrivateIPv4()
				if err == nil {
					klog.V(4).Infof("Appending a seed for cluster tag:%s, with ip=%s", p.tag, ip)
					seeds = append(seeds, ip)
				} else {
					klog.V(4).Infof("Ah ...Private IP failed for tag=%s, error=%v", p.tag, err)
				}
			} else {
				klog.V(4).Infof("Tag NOT matched for droplet tag =%s. and pTag=%s", dropTag, p.tag)
			}
		}
	}

	klog.V(4).Infof("Get seeds function done now")
	return seeds, nil
}

func NewSeedProvider(godoClient *godo.Client, tag string) (*SeedProvider, error) {
	klog.V(4).Infof("Trying new seed provider with cluster tag:%s", tag)

	return &SeedProvider{
		godoClient: godoClient,
		tag:        tag,
	}, nil
}

var _ resolver.Resolver = &SeedProvider{}

// Resolve implements resolver.Resolve, providing name -> address resolution using cloud API discovery.
func (p *SeedProvider) Resolve(ctx context.Context, name string) ([]string, error) {
	klog.Infof("trying to resolve %q using SeedProvider", name)

	// We assume we are trying to resolve a component that runs on the control plane
	isControlPlane := func(droplet *godo.Droplet) bool {
		for _, dropTag := range droplet.Tags {
			if strings.HasPrefix(dropTag, do.TagKubernetesClusterMasterPrefix+":") {
				return true
			}
		}
		return false
	}
	return p.discover(ctx, isControlPlane)
}
