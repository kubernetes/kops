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
	"fmt"
	"strconv"
)

// Elements in the balanced list
type SimpleDevice struct {
	zone             int
	nodeId, deviceId string
}

// Pretty pring a SimpleDevice
func (s *SimpleDevice) String() string {
	return fmt.Sprintf("{Z:%v N:%v D:%v}",
		s.zone,
		s.nodeId,
		s.deviceId)
}

// Simple Devices so that we have no pointers and no race conditions
type SimpleDevices []SimpleDevice

// A node is a collection of devices
type SimpleNode []*SimpleDevice

// A zone is a collection of nodes
type SimpleZone []SimpleNode

// The allocation ring will contain a map composed of all
// the devices available in the cluster.  Call Rebalance()
// for it to create a balanced list.
type SimpleAllocatorRing struct {

	// Map [zone] to [node] to slice of SimpleDevices
	ring         map[int]map[string][]*SimpleDevice
	balancedList SimpleDevices
}

// Create a new simple ring
func NewSimpleAllocatorRing() *SimpleAllocatorRing {
	s := &SimpleAllocatorRing{}
	s.ring = make(map[int]map[string][]*SimpleDevice)

	return s
}

// Convert the ring map into a consumable list of lists.
// This allows the rebalancer to go through the lists and remove
// elements as it balances
func (s *SimpleAllocatorRing) createZoneLists() []SimpleZone {
	zones := make([]SimpleZone, 0)

	for _, n := range s.ring {

		zone := make([]SimpleNode, 0)
		for _, d := range n {
			zone = append(zone, d)
		}
		zones = append(zones, zone)
	}

	return zones
}

// Add a device to the ring map
func (s *SimpleAllocatorRing) Add(d *SimpleDevice) {

	if nodes, ok := s.ring[d.zone]; ok {
		if _, ok := nodes[d.nodeId]; ok {
			nodes[d.nodeId] = append(nodes[d.nodeId], d)
		} else {
			nodes[d.nodeId] = []*SimpleDevice{d}
		}
	} else {
		s.ring[d.zone] = make(map[string][]*SimpleDevice)
		s.ring[d.zone][d.nodeId] = []*SimpleDevice{d}
	}

	s.balancedList = nil
}

// Remove device from the ring map
func (s *SimpleAllocatorRing) Remove(d *SimpleDevice) {

	if nodes, ok := s.ring[d.zone]; ok {
		if devices, ok := nodes[d.nodeId]; ok {
			for index, device := range devices {
				if device.deviceId == d.deviceId {
					// Found device, now delete it from the ring map
					nodes[d.nodeId] = append(nodes[d.nodeId][:index], nodes[d.nodeId][index+1:]...)

					if len(nodes[d.nodeId]) == 0 {
						delete(nodes, d.nodeId)
					}
					if len(s.ring[d.zone]) == 0 {
						delete(s.ring, d.zone)
					}
				}
			}
		}
	}

	s.balancedList = nil
}

// Rebalance the ring and place the rebalanced list
// into balancedList.
// The idea is to setup an array/slice where each continguous SimpleDevice
// is from either a different zone, or node.
func (s *SimpleAllocatorRing) Rebalance() {

	if s.balancedList != nil {
		return
	}

	// Copy map data to slices
	zones := s.createZoneLists()

	// Create a list
	list := make(SimpleDevices, 0)

	// Populate the list
	var device *SimpleDevice
	for i := 0; len(zones) != 0; i++ {
		zone := i % len(zones)
		node := i % len(zones[zone])

		// pop device
		device, zones[zone][node] = zones[zone][node][len(zones[zone][node])-1], zones[zone][node][:len(zones[zone][node])-1]
		list = append(list, *device)

		// delete node
		if len(zones[zone][node]) == 0 {
			zones[zone] = append(zones[zone][:node], zones[zone][node+1:]...)

			// delete zone
			if len(zones[zone]) == 0 {
				zones = append(zones[:zone], zones[zone+1:]...)
			}
		}
	}

	s.balancedList = list
}

// Use a uuid to point at a position in the ring.  Return a list of devices
// from that point in the ring.
func (s *SimpleAllocatorRing) GetDeviceList(uuid string) SimpleDevices {

	if s.balancedList == nil {
		s.Rebalance()
	}
	if len(s.balancedList) == 0 {
		return SimpleDevices{}
	}

	// Create a new list to avoid race conditions
	devices := make(SimpleDevices, len(s.balancedList))
	copy(devices, s.balancedList)

	// Instead of using 8 characters to convert to a int32, use 7 which avoids
	// negative numbers
	index64, err := strconv.ParseInt(uuid[:7], 16, 32)
	if err != nil {
		logger.Err(err)
		return devices
	}

	// Point to a position on the ring
	index := int(index64) % len(s.balancedList)

	// Return a list according to the position in the list
	return append(devices[index:], devices[:index]...)

}
