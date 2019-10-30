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

	"k8s.io/klog"
	"k8s.io/kops/pkg/resources/digitalocean"
	"k8s.io/kops/protokube/pkg/gossip"
)

type SeedProvider struct {
	cloud *digitalocean.Cloud
	tag   string
}

var _ gossip.SeedProvider = &SeedProvider{}

func (p *SeedProvider) GetSeeds() ([]string, error) {
	var seeds []string

	droplets, _, err := p.cloud.Droplets().List(context.TODO(), nil)

	if err != nil {
		return nil, fmt.Errorf("Droplets.ListByTag returned error: %v", err)
	}

	for _, droplet := range droplets {
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

func NewSeedProvider(cloud *digitalocean.Cloud, tag string) (*SeedProvider, error) {
	klog.V(4).Infof("Trying new seed provider with cluster tag:%s", tag)

	return &SeedProvider{
		cloud: cloud,
		tag:   tag,
	}, nil
}
