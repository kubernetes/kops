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
	"github.com/heketi/heketi/executors"
	"github.com/heketi/heketi/pkg/glusterfs/api"
)

type VolumeDisperseDurability struct {
	api.DisperseDurability
}

func NewVolumeDisperseDurability(d *api.DisperseDurability) *VolumeDisperseDurability {
	v := &VolumeDisperseDurability{}
	v.Data = d.Data
	v.Redundancy = d.Redundancy

	return v
}

func (d *VolumeDisperseDurability) SetDurability() {
	if d.Data == 0 {
		d.Data = DEFAULT_EC_DATA
	}
	if d.Redundancy == 0 {
		d.Redundancy = DEFAULT_EC_REDUNDANCY
	}
}

func (d *VolumeDisperseDurability) BrickSizeGenerator(size uint64) func() (int, uint64, error) {

	sets := 1
	return func() (int, uint64, error) {

		var brick_size uint64
		var num_sets int

		for {
			num_sets = sets
			sets *= 2
			brick_size = size / uint64(num_sets)

			// Divide what would be the brick size for replica by the
			// number of data drives in the disperse request
			brick_size /= uint64(d.Data)

			if brick_size < BrickMinSize {
				return 0, 0, ErrMinimumBrickSize
			} else if brick_size <= BrickMaxSize {
				break
			}
		}

		return num_sets, brick_size, nil
	}
}

func (d *VolumeDisperseDurability) MinVolumeSize() uint64 {
	return BrickMinSize * uint64(d.Data)
}

func (d *VolumeDisperseDurability) BricksInSet() int {
	return d.Data + d.Redundancy
}

func (d *VolumeDisperseDurability) SetExecutorVolumeRequest(v *executors.VolumeRequest) {
	v.Type = executors.DurabilityDispersion
	v.Data = d.Data
	v.Redundancy = d.Redundancy
}
