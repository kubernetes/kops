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

type NodeEntry struct {
	Entry

	Info    api.NodeInfo
	Devices sort.StringSlice
}

func NewNodeEntry() *NodeEntry {
	entry := &NodeEntry{}
	entry.Devices = make(sort.StringSlice, 0)
	entry.SetOnline()

	return entry
}

func NewNodeEntryFromRequest(req *api.NodeAddRequest) *NodeEntry {
	godbc.Require(req != nil)

	node := NewNodeEntry()
	node.Info.Id = utils.GenUUID()
	node.Info.ClusterId = req.ClusterId
	node.Info.Hostnames = req.Hostnames
	node.Info.Zone = req.Zone

	return node
}

func NewNodeEntryFromId(tx *bolt.Tx, id string) (*NodeEntry, error) {
	godbc.Require(tx != nil)

	entry := NewNodeEntry()
	err := EntryLoad(tx, entry, id)
	if err != nil {
		return nil, err
	}

	return entry, nil
}

func (n *NodeEntry) registerManageKey(host string) string {
	return "MANAGE" + host
}

func (n *NodeEntry) registerStorageKey(host string) string {
	return "STORAGE" + host
}

func (n *NodeEntry) Register(tx *bolt.Tx) error {

	// Save manage hostnames
	for _, h := range n.Info.Hostnames.Manage {
		val, err := EntryRegister(tx, n, n.registerManageKey(h), []byte(n.Info.Id))
		if err == ErrKeyExists {
			// Now check if the node actually exists.  This only happens
			// when the application crashes and it doesn't clean up stale
			// registrations.
			conflictId := string(val)
			_, err := NewNodeEntryFromId(tx, conflictId)
			if err == ErrNotFound {
				// (stale) There is actually no conflict, we can allow
				// the registration
				return nil
			} else if err != nil {
				return logger.Err(err)
			}

			// Return that we found a conflict
			return fmt.Errorf("Hostname %v already used by node with id %v\n",
				h, conflictId)
		} else if err != nil {
			return err
		}
	}

	// Save storage hostnames
	for _, h := range n.Info.Hostnames.Storage {
		val, err := EntryRegister(tx, n, n.registerStorageKey(h), []byte(n.Info.Id))
		if err == ErrKeyExists {

			// Check if it exists
			conflictId := string(val)
			_, err := NewNodeEntryFromId(tx, conflictId)
			if err == ErrNotFound {
				// (stale) There is actually no conflict, we can allow
				// the registration
				return nil
			} else if err != nil {
				return logger.Err(err)
			}

			// Return that we found a conflict
			return fmt.Errorf("Hostname %v already used by node with id %v\n",
				h, conflictId)
		} else if err != nil {
			return err
		}
	}

	return nil

}

func (n *NodeEntry) Deregister(tx *bolt.Tx) error {

	// Remove manage hostnames from Db
	for _, h := range n.Info.Hostnames.Manage {
		err := EntryDelete(tx, n, n.registerManageKey(h))
		if err != nil {
			return err
		}
	}

	// Remove storage hostnames
	for _, h := range n.Info.Hostnames.Storage {
		err := EntryDelete(tx, n, n.registerStorageKey(h))
		if err != nil {
			return err
		}
	}

	return nil
}

func (n *NodeEntry) BucketName() string {
	return BOLTDB_BUCKET_NODE
}

func (n *NodeEntry) Save(tx *bolt.Tx) error {
	godbc.Require(tx != nil)
	godbc.Require(len(n.Info.Id) > 0)

	return EntrySave(tx, n, n.Info.Id)

}

func (n *NodeEntry) ManageHostName() string {
	godbc.Require(n.Info.Hostnames.Manage != nil)
	godbc.Require(len(n.Info.Hostnames.Manage) > 0)

	return n.Info.Hostnames.Manage[0]
}

func (n *NodeEntry) StorageHostName() string {
	godbc.Require(n.Info.Hostnames.Storage != nil)
	godbc.Require(len(n.Info.Hostnames.Storage) > 0)

	return n.Info.Hostnames.Storage[0]
}

func (n *NodeEntry) IsDeleteOk() bool {
	// Check if the nodes still has drives
	if len(n.Devices) > 0 {
		return false
	}
	return true
}

func (n *NodeEntry) ConflictString() string {
	return fmt.Sprintf("Unable to delete node [%v] because it contains devices", n.Info.Id)
}

func (n *NodeEntry) Delete(tx *bolt.Tx) error {
	godbc.Require(tx != nil)

	// Check if the nodes still has drives
	if !n.IsDeleteOk() {
		logger.Warning(n.ConflictString())
		return ErrConflict
	}

	return EntryDelete(tx, n, n.Info.Id)
}

func (n *NodeEntry) removeAllDisksFromRing(tx *bolt.Tx,
	a Allocator) error {

	cluster, err := NewClusterEntryFromId(tx, n.Info.ClusterId)
	if err != nil {
		return err
	}

	for _, deviceId := range n.Devices {
		device, err := NewDeviceEntryFromId(tx, deviceId)
		if err != nil {
			return err
		}

		// Remove device
		err = a.RemoveDevice(cluster, n, device)
		if err != nil {
			return err
		}
	}

	return nil
}

func (n *NodeEntry) addAllDisksToRing(tx *bolt.Tx,
	a Allocator) error {

	cluster, err := NewClusterEntryFromId(tx, n.Info.ClusterId)
	if err != nil {
		return err
	}

	// Add all devices
	for _, deviceId := range n.Devices {
		device, err := NewDeviceEntryFromId(tx, deviceId)
		if err != nil {
			return err
		}

		// Add device
		if device.isOnline() {
			err = a.AddDevice(cluster, n, device)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (n *NodeEntry) SetState(tx *bolt.Tx,
	a Allocator,
	s api.EntryState) error {

	// Check current state
	switch n.State {
	case api.EntryStateFailed:
		if s == api.EntryStateFailed {
			return nil
		}
		return fmt.Errorf("Cannot reuse a failed node")

	case api.EntryStateOnline:
		switch s {
		case api.EntryStateOnline:
			return nil
		case api.EntryStateFailed:
		case api.EntryStateOffline:
		default:
			return fmt.Errorf("Unknown state type: %v", s)
		}

		// Remove all disks from Ring
		err := n.removeAllDisksFromRing(tx, a)
		if err != nil {
			return err
		}

		// Save state
		n.State = s

	case api.EntryStateOffline:
		switch s {
		case api.EntryStateOffline:
			return nil
		case api.EntryStateOnline:
			// Add disks back
			err := n.addAllDisksToRing(tx, a)
			if err != nil {
				return err
			}
		case api.EntryStateFailed:
			// Only thing to do here is to set the state
		default:
			return fmt.Errorf("Unknown state type: %v", s)
		}
		n.State = s
	}

	return nil
}

func (n *NodeEntry) NewInfoReponse(tx *bolt.Tx) (*api.NodeInfoResponse, error) {

	godbc.Require(tx != nil)

	info := &api.NodeInfoResponse{}
	info.ClusterId = n.Info.ClusterId
	info.Hostnames = n.Info.Hostnames
	info.Id = n.Info.Id
	info.Zone = n.Info.Zone
	info.State = n.State
	info.DevicesInfo = make([]api.DeviceInfoResponse, 0)

	// Add each drive information
	for _, deviceid := range n.Devices {
		device, err := NewDeviceEntryFromId(tx, deviceid)
		if err != nil {
			return nil, err
		}

		driveinfo, err := device.NewInfoResponse(tx)
		if err != nil {
			return nil, err
		}
		info.DevicesInfo = append(info.DevicesInfo, *driveinfo)
	}

	return info, nil
}

func (n *NodeEntry) Marshal() ([]byte, error) {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(*n)

	return buffer.Bytes(), err
}

func (n *NodeEntry) Unmarshal(buffer []byte) error {
	dec := gob.NewDecoder(bytes.NewReader(buffer))
	err := dec.Decode(n)
	if err != nil {
		return err
	}

	// Make sure to setup arrays if nil
	if n.Devices == nil {
		n.Devices = make(sort.StringSlice, 0)
	}

	return nil
}

func (n *NodeEntry) DeviceAdd(id string) {
	godbc.Require(!utils.SortedStringHas(n.Devices, id))

	n.Devices = append(n.Devices, id)
	n.Devices.Sort()
}

func (n *NodeEntry) DeviceDelete(id string) {
	n.Devices = utils.SortedStringsDelete(n.Devices, id)
}
