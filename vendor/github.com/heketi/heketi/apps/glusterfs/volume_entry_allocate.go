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
)

func (v *VolumeEntry) allocBricksInCluster(db *bolt.DB,
	allocator Allocator,
	cluster string,
	gbsize int) ([]*BrickEntry, error) {

	size := uint64(gbsize) * GB

	// Setup a brick size generator
	// Note: subsequent calls to gen need to return decreasing
	//       brick sizes in order for the following code to work!
	gen := v.Durability.BrickSizeGenerator(size)

	// Try decreasing possible brick sizes until space is found
	for {
		// Determine next possible brick size
		sets, brick_size, err := gen()
		if err != nil {
			logger.Err(err)
			return nil, err
		}

		num_bricks := sets * v.Durability.BricksInSet()

		logger.Debug("brick_size = %v", brick_size)
		logger.Debug("sets = %v", sets)
		logger.Debug("num_bricks = %v", num_bricks)

		// Check that the volume would not have too many bricks
		if (num_bricks + len(v.Bricks)) > BrickMaxNum {
			logger.Debug("Maximum number of bricks reached")
			return nil, ErrMaxBricks
		}

		// Allocate bricks in the cluster
		brick_entries, err := v.allocBricks(db, allocator, cluster, sets, brick_size)
		if err == ErrNoSpace {
			logger.Debug("No space, re-trying with smaller brick size")
			continue
		}
		if err != nil {
			logger.Err(err)
			return nil, err
		}

		// We were able to allocate bricks
		return brick_entries, nil
	}
}

func (v *VolumeEntry) allocBricks(
	db *bolt.DB,
	allocator Allocator,
	cluster string,
	bricksets int,
	brick_size uint64) (brick_entries []*BrickEntry, e error) {

	// Setup garbage collector function in case of error
	defer func() {

		// Check the named return value 'err'
		if e != nil {
			logger.Debug("Error detected.  Cleaning up volume %v: Len(%v) ", v.Info.Id, len(brick_entries))
			db.Update(func(tx *bolt.Tx) error {
				for _, brick := range brick_entries {
					v.removeBrickFromDb(tx, brick)
				}
				return nil
			})
		}
	}()

	// Initialize brick_entries
	brick_entries = make([]*BrickEntry, 0)

	// Determine allocation for each brick required for this volume
	for brick_num := 0; brick_num < bricksets; brick_num++ {
		logger.Info("brick_num: %v", brick_num)

		// Create a brick set list to later make sure that the
		// proposed bricks and devices are acceptable
		setlist := make([]*BrickEntry, 0)

		// Generate an id for the brick
		brickId := utils.GenUUID()

		// Get allocator generator
		// The same generator should be used for the brick and its replicas
		deviceCh, done, errc := allocator.GetNodes(cluster, brickId)
		defer func() {
			close(done)
		}()

		// Check location has space for each brick and its replicas
		for i := 0; i < v.Durability.BricksInSet(); i++ {
			logger.Debug("%v / %v", i, v.Durability.BricksInSet())

			// Do the work in the database context so that the cluster
			// data does not change while determining brick location
			err := db.Update(func(tx *bolt.Tx) error {

				// Check the ring for devices to place the brick
				for deviceId := range deviceCh {

					// Get device entry
					device, err := NewDeviceEntryFromId(tx, deviceId)
					if err != nil {
						return err
					}

					// Do not allow a device from the same node to be
					// in the set
					deviceOk := true
					for _, brickInSet := range setlist {
						if brickInSet.Info.NodeId == device.NodeId {
							deviceOk = false
						}
					}

					if !deviceOk {
						continue
					}

					// Try to allocate a brick on this device
					brick := device.NewBrickEntry(brick_size,
						float64(v.Info.Snapshot.Factor),
						v.gidRequested)

					// Determine if it was successful
					if brick != nil {

						// If the first in the set, the reset the id
						if i == 0 {
							brick.SetId(brickId)
						}

						// Save the brick entry to create later
						brick_entries = append(brick_entries, brick)

						// Add to set list
						setlist = append(setlist, brick)

						// Add brick to device
						device.BrickAdd(brick.Id())

						// Add brick to volume
						v.BrickAdd(brick.Id())

						// Save values
						err := device.Save(tx)
						if err != nil {
							return err
						}
						return nil
					}
				}

				// Check if allocator returned an error
				if err := <-errc; err != nil {
					return err
				}

				// No devices found
				return ErrNoSpace

			})
			if err != nil {
				return brick_entries, err
			}
		}
	}

	return brick_entries, nil

}

func (v *VolumeEntry) removeBrickFromDb(tx *bolt.Tx, brick *BrickEntry) error {

	// Access device
	device, err := NewDeviceEntryFromId(tx, brick.Info.DeviceId)
	if err != nil {
		logger.Err(err)
		return err
	}

	// Deallocate space on device
	device.StorageFree(brick.TotalSize())

	// Delete brick from device
	device.BrickDelete(brick.Info.Id)

	// Save device
	err = device.Save(tx)
	if err != nil {
		logger.Err(err)
		return err
	}

	// Delete brick entryfrom db
	err = brick.Delete(tx)
	if err != nil {
		logger.Err(err)
		return err
	}

	// Delete brick from volume db
	v.BrickDelete(brick.Info.Id)
	if err != nil {
		logger.Err(err)
		return err
	}

	return nil
}
