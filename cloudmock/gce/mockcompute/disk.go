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
	"fmt"
	"sync"

	compute "google.golang.org/api/compute/v1"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
)

type diskClient struct {
	// disks are disks keyed by project, zone, and disk name.
	disks map[string]map[string]map[string]*compute.Disk
	sync.Mutex
}

var _ gce.DiskClient = &diskClient{}

func newDiskClient() *diskClient {
	return &diskClient{
		disks: map[string]map[string]map[string]*compute.Disk{},
	}
}

func (c *diskClient) All() map[string]interface{} {
	c.Lock()
	defer c.Unlock()
	m := map[string]interface{}{}
	for _, zones := range c.disks {
		for _, disks := range zones {
			for n, disk := range disks {
				m[n] = disk
			}
		}
	}
	return m
}

func (c *diskClient) Insert(project, zone string, disk *compute.Disk) (*compute.Operation, error) {
	c.Lock()
	defer c.Unlock()
	zones, ok := c.disks[project]
	if !ok {
		zones = map[string]map[string]*compute.Disk{}
		c.disks[project] = zones
	}
	disks, ok := zones[zone]
	if !ok {
		disks = map[string]*compute.Disk{}
		zones[zone] = disks
	}
	disk.SelfLink = fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/zones/%s/disks/%s", project, zone, disk.Name)
	disk.Zone = zone
	disks[disk.Name] = disk
	return doneOperation(), nil
}

func (c *diskClient) Delete(project, zone, name string) (*compute.Operation, error) {
	c.Lock()
	defer c.Unlock()
	zones, ok := c.disks[project]
	if !ok {
		return nil, notFoundError()
	}
	disks, ok := zones[zone]
	if !ok {
		return nil, notFoundError()
	}
	if _, ok := disks[name]; !ok {
		return nil, notFoundError()
	}
	delete(disks, name)
	return doneOperation(), nil
}

func (c *diskClient) Get(project, zone, name string) (*compute.Disk, error) {
	c.Lock()
	defer c.Unlock()
	zones, ok := c.disks[project]
	if !ok {
		return nil, notFoundError()
	}
	disks, ok := zones[zone]
	if !ok {
		return nil, notFoundError()
	}
	disk, ok := disks[name]
	if !ok {
		return nil, notFoundError()
	}
	return disk, nil
}

func (c *diskClient) List(ctx context.Context, project, zone string) ([]*compute.Disk, error) {
	c.Lock()
	defer c.Unlock()
	zones, ok := c.disks[project]
	if !ok {
		return nil, nil
	}
	disks, ok := zones[zone]
	if !ok {
		return nil, nil
	}
	var l []*compute.Disk
	for _, d := range disks {
		l = append(l, d)
	}
	return l, nil
}

func (c *diskClient) AggregatedList(ctx context.Context, project string) ([]compute.DisksScopedList, error) {
	c.Lock()
	defer c.Unlock()
	zones, ok := c.disks[project]
	if !ok {
		return nil, nil
	}
	var allDisks []*compute.Disk
	for _, disks := range zones {
		for _, disk := range disks {
			allDisks = append(allDisks, disk)
		}
	}
	return []compute.DisksScopedList{
		{
			Disks: allDisks,
		},
	}, nil
}

func (c *diskClient) SetLabels(project, zone, name string, req *compute.ZoneSetLabelsRequest) error {
	c.Lock()
	defer c.Unlock()
	zones, ok := c.disks[project]
	if !ok {
		return notFoundError()
	}
	disks, ok := zones[zone]
	if !ok {
		return notFoundError()
	}
	disk, ok := disks[name]
	if !ok {
		return notFoundError()
	}
	disk.Labels = req.Labels
	return nil
}
