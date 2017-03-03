//
// Copyright (c) 2015 The heketi Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package glusterfs

import (
	"github.com/boltdb/bolt"
	"github.com/heketi/heketi/pkg/utils"
	"sort"
	"sync"
)

type MockAllocator struct {
	clustermap map[string]sort.StringSlice
	lock       sync.Mutex
	db         bolt.DB
}

func NewMockAllocator(db *bolt.DB) *MockAllocator {
	d := &MockAllocator{}
	d.clustermap = make(map[string]sort.StringSlice)

	var clusters []string
	err := db.View(func(tx *bolt.Tx) error {
		var err error
		clusters, err = ClusterList(tx)
		if err != nil {
			return err
		}

		for _, cluster := range clusters {
			err := d.addDevicesFromDb(tx, cluster)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil
	}

	return d
}

func (d *MockAllocator) AddDevice(cluster *ClusterEntry,
	node *NodeEntry,
	device *DeviceEntry) error {

	d.lock.Lock()
	defer d.lock.Unlock()

	clusterId := cluster.Info.Id
	deviceId := device.Info.Id

	if devicelist, ok := d.clustermap[clusterId]; ok {
		devicelist = append(devicelist, deviceId)
		devicelist.Sort()
		d.clustermap[clusterId] = devicelist
	} else {
		d.clustermap[clusterId] = sort.StringSlice{deviceId}
	}

	return nil
}

func (d *MockAllocator) RemoveDevice(cluster *ClusterEntry,
	node *NodeEntry,
	device *DeviceEntry) error {

	d.lock.Lock()
	defer d.lock.Unlock()

	clusterId := cluster.Info.Id
	deviceId := device.Info.Id

	d.clustermap[clusterId] = utils.SortedStringsDelete(d.clustermap[clusterId], deviceId)

	return nil
}

func (d *MockAllocator) RemoveCluster(clusterId string) error {
	// Save in the object
	d.lock.Lock()
	defer d.lock.Unlock()

	delete(d.clustermap, clusterId)

	return nil
}

func (d *MockAllocator) GetNodes(clusterId, brickId string) (<-chan string,
	chan<- struct{}, <-chan error) {

	// Initialize channels
	device, done := make(chan string), make(chan struct{})

	// Make sure to make a buffered channel for the error, so we can
	// set it and return
	errc := make(chan error, 1)

	d.lock.Lock()
	devicelist := d.clustermap[clusterId]
	d.lock.Unlock()

	// Start generator in a new goroutine
	go func() {
		defer func() {
			errc <- nil
			close(device)
		}()

		for _, id := range devicelist {
			select {
			case device <- id:
			case <-done:
				return
			}
		}

	}()

	return device, done, errc
}

func (d *MockAllocator) addDevicesFromDb(tx *bolt.Tx, clusterId string) error {
	// Get data from the DB
	devicelist := make(sort.StringSlice, 0)

	// Get cluster info
	cluster, err := NewClusterEntryFromId(tx, clusterId)
	if err != nil {
		return err
	}

	for _, nodeId := range cluster.Info.Nodes {
		node, err := NewNodeEntryFromId(tx, nodeId)
		if err != nil {
			return err
		}

		devicelist = append(devicelist, node.Devices...)
	}

	// We have to sort the list so that later we can search and delete an entry
	devicelist.Sort()

	// Save in the object
	d.lock.Lock()
	defer d.lock.Unlock()

	d.clustermap[clusterId] = devicelist
	return nil
}
