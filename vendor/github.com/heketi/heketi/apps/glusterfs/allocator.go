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

type Allocator interface {

	// Inform the brick allocator to include device
	AddDevice(c *ClusterEntry, n *NodeEntry, d *DeviceEntry) error

	// Inform the brick allocator to not use the specified device
	RemoveDevice(c *ClusterEntry, n *NodeEntry, d *DeviceEntry) error

	// Remove cluster information from allocator
	RemoveCluster(clusterId string) error

	// Returns a generator, done, and error channel.
	// The generator returns the location for the brick, then the possible locations
	// of its replicas. The caller must close() the done channel when it no longer
	// needs to read from the generator.
	GetNodes(clusterId, brickId string) (<-chan string,
		chan<- struct{}, <-chan error)
}
