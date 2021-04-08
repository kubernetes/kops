/*
Copyright 2021 The Kubernetes Authors.

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

package mockcompute

import (
	"context"

	compute "google.golang.org/api/compute/v1"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
)

type zoneClient struct {
	// zones are zones keyed by project and zone name.
	zones map[string]map[string]*compute.Zone
}

var _ gce.ZoneClient = &zoneClient{}

func newZoneClient(project string) *zoneClient {
	return &zoneClient{
		zones: map[string]map[string]*compute.Zone{
			project: {
				"us-test1-a": {
					Name:   "us-test1-a",
					Region: "https://www.googleapis.com/compute/v1/projects/testproject/regions/us-test1",
				},
			},
		},
	}
}

func (c *zoneClient) All() map[string]interface{} {
	m := map[string]interface{}{}
	for _, zones := range c.zones {
		for n, z := range zones {
			m[n] = z
		}
	}
	return m
}

func (c *zoneClient) List(ctx context.Context, project string) ([]*compute.Zone, error) {
	zones, ok := c.zones[project]
	if !ok {
		return nil, nil
	}
	var l []*compute.Zone
	for _, z := range zones {
		l = append(l, z)
	}
	return l, nil
}
