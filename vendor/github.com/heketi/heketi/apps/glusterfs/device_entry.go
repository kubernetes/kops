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
	"bytes"
	"encoding/gob"
	"fmt"
	"sort"

	"github.com/boltdb/bolt"
	"github.com/heketi/heketi/pkg/glusterfs/api"
	"github.com/heketi/heketi/pkg/utils"
	"github.com/lpabon/godbc"
)

const (
	maxPoolMetadataSizeMb = 16 * GB
)

type DeviceEntry struct {
	Entry

	Info       api.DeviceInfo
	Bricks     sort.StringSlice
	NodeId     string
	ExtentSize uint64
}

func DeviceList(tx *bolt.Tx) ([]string, error) {

	list := EntryKeys(tx, BOLTDB_BUCKET_DEVICE)
	if list == nil {
		return nil, ErrAccessList
	}
	return list, nil
}

func NewDeviceEntry() *DeviceEntry {
	entry := &DeviceEntry{}
	entry.Bricks = make(sort.StringSlice, 0)
	entry.SetOnline()

	// Default to 4096KB
	entry.ExtentSize = 4096

	return entry
}

func NewDeviceEntryFromRequest(req *api.DeviceAddRequest) *DeviceEntry {
	godbc.Require(req != nil)

	device := NewDeviceEntry()
	device.Info.Id = utils.GenUUID()
	device.Info.Name = req.Name
	device.NodeId = req.NodeId

	return device
}

func NewDeviceEntryFromId(tx *bolt.Tx, id string) (*DeviceEntry, error) {
	godbc.Require(tx != nil)

	entry := NewDeviceEntry()
	err := EntryLoad(tx, entry, id)
	if err != nil {
		return nil, err
	}

	return entry, nil
}

func (d *DeviceEntry) registerKey() string {
	return "DEVICE" + d.NodeId + d.Info.Name
}

func (d *DeviceEntry) Register(tx *bolt.Tx) error {
	godbc.Require(tx != nil)

	val, err := EntryRegister(tx,
		d,
		d.registerKey(),
		[]byte(d.Id()))
	if err == ErrKeyExists {

		// Now check if the node actually exists.  This only happens
		// when the application crashes and it doesn't clean up stale
		// registrations.
		conflictId := string(val)
		_, err := NewDeviceEntryFromId(tx, conflictId)
		if err == ErrNotFound {
			// (stale) There is actually no conflict, we can allow
			// the registration
			return nil
		} else if err != nil {
			return logger.Err(err)
		}

		return fmt.Errorf("Device %v is already used on node %v by device %v",
			d.Info.Name,
			d.NodeId,
			conflictId)

	} else if err != nil {
		return err
	}

	return nil
}

func (d *DeviceEntry) Deregister(tx *bolt.Tx) error {
	godbc.Require(tx != nil)

	err := EntryDelete(tx, d, d.registerKey())
	if err != nil {
		return err
	}

	return nil
}

func (d *DeviceEntry) SetId(id string) {
	d.Info.Id = id
}

func (d *DeviceEntry) Id() string {
	return d.Info.Id
}

func (d *DeviceEntry) BucketName() string {
	return BOLTDB_BUCKET_DEVICE
}

func (d *DeviceEntry) Save(tx *bolt.Tx) error {
	godbc.Require(tx != nil)
	godbc.Require(len(d.Info.Id) > 0)

	return EntrySave(tx, d, d.Info.Id)

}

func (d *DeviceEntry) IsDeleteOk() bool {
	// Check if the nodes still has drives
	if len(d.Bricks) > 0 {
		return false
	}
	return true
}

func (d *DeviceEntry) ConflictString() string {
	return fmt.Sprintf("Unable to delete device [%v] because it contains bricks", d.Info.Id)
}

func (d *DeviceEntry) Delete(tx *bolt.Tx) error {
	godbc.Require(tx != nil)

	// Check if the devices still has drives
	if !d.IsDeleteOk() {
		logger.Warning(d.ConflictString())
		return ErrConflict
	}

	return EntryDelete(tx, d, d.Info.Id)
}

func (d *DeviceEntry) removeDeviceFromRing(tx *bolt.Tx,
	a Allocator) error {

	node, err := NewNodeEntryFromId(tx, d.NodeId)
	if err != nil {
		return err
	}

	cluster, err := NewClusterEntryFromId(tx, node.Info.ClusterId)
	if err != nil {
		return err
	}

	return a.RemoveDevice(cluster, node, d)
}

func (d *DeviceEntry) addDeviceToRing(tx *bolt.Tx,
	a Allocator) error {

	node, err := NewNodeEntryFromId(tx, d.NodeId)
	if err != nil {
		return err
	}

	cluster, err := NewClusterEntryFromId(tx, node.Info.ClusterId)
	if err != nil {
		return err
	}

	return a.AddDevice(cluster, node, d)
}

func (d *DeviceEntry) SetState(tx *bolt.Tx,
	a Allocator,
	s api.EntryState) error {

	// Check current state
	switch d.State {
	case api.EntryStateFailed:
		if s == api.EntryStateFailed {
			return nil
		}
		return fmt.Errorf("Cannot reuse a failed device")

	case api.EntryStateOnline:
		switch s {
		case api.EntryStateOnline:
			return nil
		case api.EntryStateFailed:
		case api.EntryStateOffline:
		default:
			return fmt.Errorf("Unknown state type: %v", s)
		}

		// Remove disk from Ring
		err := d.removeDeviceFromRing(tx, a)
		if err != nil {
			return err
		}

		// Save state
		d.State = s

	case api.EntryStateOffline:
		switch s {
		case api.EntryStateOffline:
			return nil
		case api.EntryStateOnline:
			// Add disk back
			err := d.addDeviceToRing(tx, a)
			if err != nil {
				return err
			}
		case api.EntryStateFailed:
			// Only thing to do here is to set the state
		default:
			return fmt.Errorf("Unknown state type: %v", s)
		}
		d.State = s
	}

	return nil
}

func (d *DeviceEntry) NewInfoResponse(tx *bolt.Tx) (*api.DeviceInfoResponse, error) {

	godbc.Require(tx != nil)

	info := &api.DeviceInfoResponse{}
	info.Id = d.Info.Id
	info.Name = d.Info.Name
	info.Storage = d.Info.Storage
	info.State = d.State
	info.Bricks = make([]api.BrickInfo, 0)

	// Add each drive information
	for _, id := range d.Bricks {
		brick, err := NewBrickEntryFromId(tx, id)
		if err != nil {
			return nil, err
		}

		brickinfo, err := brick.NewInfoResponse(tx)
		if err != nil {
			return nil, err
		}
		info.Bricks = append(info.Bricks, *brickinfo)
	}

	return info, nil
}

func (d *DeviceEntry) Marshal() ([]byte, error) {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(*d)

	return buffer.Bytes(), err
}

func (d *DeviceEntry) Unmarshal(buffer []byte) error {
	dec := gob.NewDecoder(bytes.NewReader(buffer))
	err := dec.Decode(d)
	if err != nil {
		return err
	}

	// Make sure to setup arrays if nil
	if d.Bricks == nil {
		d.Bricks = make(sort.StringSlice, 0)
	}

	return nil
}

func (d *DeviceEntry) BrickAdd(id string) {
	godbc.Require(!utils.SortedStringHas(d.Bricks, id))

	d.Bricks = append(d.Bricks, id)
	d.Bricks.Sort()
}

func (d *DeviceEntry) BrickDelete(id string) {
	d.Bricks = utils.SortedStringsDelete(d.Bricks, id)
}

func (d *DeviceEntry) StorageSet(amount uint64) {
	d.Info.Storage.Free = amount
	d.Info.Storage.Total = amount
}

func (d *DeviceEntry) StorageAllocate(amount uint64) {
	d.Info.Storage.Free -= amount
	d.Info.Storage.Used += amount
}

func (d *DeviceEntry) StorageFree(amount uint64) {
	d.Info.Storage.Free += amount
	d.Info.Storage.Used -= amount
}

func (d *DeviceEntry) StorageCheck(amount uint64) bool {
	return d.Info.Storage.Free > amount
}

func (d *DeviceEntry) SetExtentSize(amount uint64) {
	d.ExtentSize = amount
}

// Allocates a new brick if the space is available.  It will automatically reserve
// the storage amount required from the device's used storage, but it will not add
// the brick id to the brick list.  The caller is responsabile for adding the brick
// id to the list.
func (d *DeviceEntry) NewBrickEntry(amount uint64, snapFactor float64, gid int64) *BrickEntry {

	// :TODO: This needs unit test

	// Calculate thinpool size
	tpsize := uint64(float64(amount) * snapFactor)

	// Align tpsize to extent
	alignment := tpsize % d.ExtentSize
	if alignment != 0 {
		tpsize += d.ExtentSize - alignment
	}

	// Determine if we need to allocate space for the metadata
	metadataSize := d.poolMetadataSize(tpsize)

	// Align to extent
	alignment = metadataSize % d.ExtentSize
	if alignment != 0 {
		metadataSize += d.ExtentSize - alignment
	}

	// Total required size
	total := tpsize + metadataSize

	logger.Debug("device %v[%v] > required size [%v] ?",
		d.Id(),
		d.Info.Storage.Free, total)
	if !d.StorageCheck(total) {
		return nil
	}

	// Allocate amount from disk
	d.StorageAllocate(total)

	// Create brick
	return NewBrickEntry(amount, tpsize, metadataSize, d.Info.Id, d.NodeId, gid)
}

// Return poolmetadatasize in KB
func (d *DeviceEntry) poolMetadataSize(tpsize uint64) uint64 {

	// TP size is in KB
	p := uint64(float64(tpsize) * 0.005)
	if p > maxPoolMetadataSizeMb {
		p = maxPoolMetadataSizeMb
	}

	return p
}
