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

package executors

type Executor interface {
	PeerProbe(exec_host, newnode string) error
	PeerDetach(exec_host, detachnode string) error
	DeviceSetup(host, device, vgid string) (*DeviceInfo, error)
	DeviceTeardown(host, device, vgid string) error
	BrickCreate(host string, brick *BrickRequest) (*BrickInfo, error)
	BrickDestroy(host string, brick *BrickRequest) error
	BrickDestroyCheck(host string, brick *BrickRequest) error
	VolumeCreate(host string, volume *VolumeRequest) (*VolumeInfo, error)
	VolumeDestroy(host string, volume string) error
	VolumeDestroyCheck(host, volume string) error
	VolumeExpand(host string, volume *VolumeRequest) (*VolumeInfo, error)
	SetLogLevel(level string)
}

// Enumerate durability types
type DurabilityType int

const (
	DurabilityNone DurabilityType = iota
	DurabilityReplica
	DurabilityDispersion
)

// Returns the size of the device
type DeviceInfo struct {
	// Size in KB
	Size       uint64
	ExtentSize uint64
}

// Brick description
type BrickRequest struct {
	VgId             string
	Name             string
	TpSize           uint64
	Size             uint64
	PoolMetadataSize uint64
	Gid              int64
}

// Returns information about the location of the brick
type BrickInfo struct {
	Path string
	Host string
}

type VolumeRequest struct {
	Bricks []BrickInfo
	Name   string
	Type   DurabilityType

	// Dispersion
	Data       int
	Redundancy int

	// Replica
	Replica int
}

type VolumeInfo struct {
}
