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

package yandex

import (
	"context"
	"fmt"
	"strings"

	"cloud.google.com/go/compute/metadata"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"k8s.io/klog/v2"
	"k8s.io/kops/protokube/pkg/gossip"
)

type SeedProvider struct {
	sdk *ycsdk.SDK
	tag string
}

var _ gossip.SeedProvider = &SeedProvider{}

func NewSeedProvider(sdk *ycsdk.SDK, tag string) (*SeedProvider, error) {
	sdk, err := ycsdk.Build(context.TODO(), ycsdk.Config{
		Credentials: ycsdk.InstanceServiceAccount(),
	})
	if err != nil {
		return nil, err
	}
	return &SeedProvider{
		sdk: sdk,
		tag: tag,
	}, nil
}

func (p *SeedProvider) GetSeeds() ([]string, error) {
	var seeds []string
	zone, err := metadata.Get("instance/zone") // .Zone() returns only zone
	if err != nil {
		return nil, err
	}
	klog.Infof("metadata.Zone(): %s ", zone)
	data := strings.Split(zone, "/")
	if len(data) == 0 {
		return nil, fmt.Errorf("error getting folderId from zone identifier (metadata): %s", zone)
	}
	folderId := data[1]
	// TODO(YuraBeznos): make names of the nodes special and filter by regex
	// TODO(YuraBeznos): right now we can use regex on name   "name": "master-ru-central1-a", (master or nodes, a,b,c zones)
	// filter := fmt.Sprintf("name=\"%s\"", *v.Name) // only filter by name supported atm 08.2022
	resp, err := p.sdk.Compute().Instance().List(context.TODO(), &compute.ListInstancesRequest{
		FolderId: folderId,
		// Filter:   filter,
		PageSize: 100,
	})
	if err != nil {
		return nil, err
	}
	for _, instance := range resp.Instances {
		seeds = append(seeds, instance.NetworkInterfaces[0].PrimaryV4Address.GetAddress())
	}
	klog.Infof("seeds: %s ", seeds)
	return seeds, nil
}
