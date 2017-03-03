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
	"testing"

	"github.com/heketi/heketi/executors"
	"github.com/heketi/tests"
)

func TestNoneDurabilityDefaults(t *testing.T) {
	r := &NoneDurability{}
	tests.Assert(t, r.Replica == 0)

	r.SetDurability()
	tests.Assert(t, r.Replica == 1)
}

func TestDisperseDurabilityDefaults(t *testing.T) {
	r := &VolumeDisperseDurability{}
	tests.Assert(t, r.Data == 0)
	tests.Assert(t, r.Redundancy == 0)

	r.SetDurability()
	tests.Assert(t, r.Data == DEFAULT_EC_DATA)
	tests.Assert(t, r.Redundancy == DEFAULT_EC_REDUNDANCY)
}

func TestReplicaDurabilityDefaults(t *testing.T) {
	r := &VolumeReplicaDurability{}
	tests.Assert(t, r.Replica == 0)

	r.SetDurability()
	tests.Assert(t, r.Replica == DEFAULT_REPLICA)
}

func TestNoneDurabilitySetExecutorRequest(t *testing.T) {
	r := &NoneDurability{}
	r.SetDurability()

	v := &executors.VolumeRequest{}
	r.SetExecutorVolumeRequest(v)
	tests.Assert(t, v.Replica == 1)
	tests.Assert(t, v.Type == executors.DurabilityNone)
}

func TestDisperseDurabilitySetExecutorRequest(t *testing.T) {
	r := &VolumeDisperseDurability{}
	r.SetDurability()

	v := &executors.VolumeRequest{}
	r.SetExecutorVolumeRequest(v)
	tests.Assert(t, v.Data == r.Data)
	tests.Assert(t, v.Redundancy == r.Redundancy)
	tests.Assert(t, v.Type == executors.DurabilityDispersion)
}

func TestReplicaDurabilitySetExecutorRequest(t *testing.T) {
	r := &VolumeReplicaDurability{}
	r.SetDurability()

	v := &executors.VolumeRequest{}
	r.SetExecutorVolumeRequest(v)
	tests.Assert(t, v.Replica == r.Replica)
	tests.Assert(t, v.Type == executors.DurabilityReplica)
}

func TestNoneDurability(t *testing.T) {
	r := &NoneDurability{}
	r.SetDurability()

	gen := r.BrickSizeGenerator(100 * GB)

	// Gen 1
	sets, brick_size, err := gen()
	tests.Assert(t, err == nil)
	tests.Assert(t, sets == 1)
	tests.Assert(t, brick_size == 100*GB)
	tests.Assert(t, 1 == r.BricksInSet())

	// Gen 2
	sets, brick_size, err = gen()
	tests.Assert(t, err == nil)
	tests.Assert(t, sets == 2)
	tests.Assert(t, brick_size == 50*GB)
	tests.Assert(t, 1 == r.BricksInSet())

	// Gen 3
	sets, brick_size, err = gen()
	tests.Assert(t, err == nil)
	tests.Assert(t, sets == 4)
	tests.Assert(t, brick_size == 25*GB)
	tests.Assert(t, 1 == r.BricksInSet())

	// Gen 4
	sets, brick_size, err = gen()
	tests.Assert(t, err == nil)
	tests.Assert(t, sets == 8)
	tests.Assert(t, brick_size == 12800*1024)
	tests.Assert(t, 1 == r.BricksInSet())

	// Gen 5
	sets, brick_size, err = gen()
	tests.Assert(t, err == nil)
	tests.Assert(t, sets == 16)
	tests.Assert(t, brick_size == 6400*1024)
	tests.Assert(t, 1 == r.BricksInSet())

	// Gen 6
	sets, brick_size, err = gen()
	tests.Assert(t, err == ErrMinimumBrickSize)
	tests.Assert(t, sets == 0)
	tests.Assert(t, brick_size == 0)
	tests.Assert(t, 1 == r.BricksInSet())
}

func TestDisperseDurability(t *testing.T) {

	r := &VolumeDisperseDurability{}
	r.Data = 8
	r.Redundancy = 3

	gen := r.BrickSizeGenerator(200 * GB)

	// Gen 1
	sets, brick_size, err := gen()
	tests.Assert(t, err == nil)
	tests.Assert(t, sets == 1)
	tests.Assert(t, brick_size == uint64(200*GB/8))
	tests.Assert(t, 8+3 == r.BricksInSet())

	// Gen 2
	sets, brick_size, err = gen()
	tests.Assert(t, err == nil)
	tests.Assert(t, sets == 2)
	tests.Assert(t, brick_size == uint64(100*GB/8))
	tests.Assert(t, 8+3 == r.BricksInSet())

	// Gen 3
	sets, brick_size, err = gen()
	tests.Assert(t, err == nil)
	tests.Assert(t, sets == 4)
	tests.Assert(t, brick_size == uint64(50*GB/8))
	tests.Assert(t, 8+3 == r.BricksInSet())

	// Gen 4
	sets, brick_size, err = gen()
	tests.Assert(t, err == ErrMinimumBrickSize)
	tests.Assert(t, 8+3 == r.BricksInSet())
}

func TestDisperseDurabilityLargeBrickGenerator(t *testing.T) {
	r := &VolumeDisperseDurability{}
	r.Data = 8
	r.Redundancy = 3

	gen := r.BrickSizeGenerator(800 * TB)

	// Gen 1
	sets, brick_size, err := gen()
	tests.Assert(t, err == nil)
	tests.Assert(t, sets == 32)
	tests.Assert(t, brick_size == 3200*GB)
	tests.Assert(t, 8+3 == r.BricksInSet())
}

func TestReplicaDurabilityGenerator(t *testing.T) {
	r := &VolumeReplicaDurability{}
	r.Replica = 2

	gen := r.BrickSizeGenerator(100 * GB)

	// Gen 1
	sets, brick_size, err := gen()
	tests.Assert(t, err == nil)
	tests.Assert(t, sets == 1)
	tests.Assert(t, brick_size == 100*GB)
	tests.Assert(t, 2 == r.BricksInSet())

	// Gen 2
	sets, brick_size, err = gen()
	tests.Assert(t, err == nil)
	tests.Assert(t, sets == 2, "sets we got:", sets)
	tests.Assert(t, brick_size == 50*GB)
	tests.Assert(t, 2 == r.BricksInSet())

	// Gen 3
	sets, brick_size, err = gen()
	tests.Assert(t, err == nil)
	tests.Assert(t, sets == 4)
	tests.Assert(t, brick_size == 25*GB)
	tests.Assert(t, 2 == r.BricksInSet())

	// Gen 4
	sets, brick_size, err = gen()
	tests.Assert(t, err == nil)
	tests.Assert(t, sets == 8)
	tests.Assert(t, brick_size == 12800*1024)
	tests.Assert(t, 2 == r.BricksInSet())

	// Gen 5
	sets, brick_size, err = gen()
	tests.Assert(t, err == nil)
	tests.Assert(t, sets == 16)
	tests.Assert(t, brick_size == 6400*1024)
	tests.Assert(t, 2 == r.BricksInSet())

	// Gen 6
	sets, brick_size, err = gen()
	tests.Assert(t, err == ErrMinimumBrickSize)
	tests.Assert(t, sets == 0)
	tests.Assert(t, brick_size == 0)
	tests.Assert(t, 2 == r.BricksInSet())
}

func TestReplicaDurabilityLargeBrickGenerator(t *testing.T) {
	r := &VolumeReplicaDurability{}
	r.Replica = 2

	gen := r.BrickSizeGenerator(100 * TB)

	// Gen 1
	sets, brick_size, err := gen()
	tests.Assert(t, err == nil)
	tests.Assert(t, sets == 32)
	tests.Assert(t, brick_size == 3200*GB)
	tests.Assert(t, 2 == r.BricksInSet())
}

func TestNoneDurabilityMinVolumeSize(t *testing.T) {
	r := &NoneDurability{}
	r.SetDurability()

	minvolsize := r.MinVolumeSize()

	tests.Assert(t, minvolsize == BrickMinSize)
}

func TestReplicaDurabilityMinVolumeSize(t *testing.T) {
	r := &VolumeReplicaDurability{}
	r.Replica = 3

	minvolsize := r.MinVolumeSize()

	tests.Assert(t, minvolsize == BrickMinSize)
}

func TestDisperseDurabilityMinVolumeSize(t *testing.T) {
	r := &VolumeDisperseDurability{}
	r.Data = 8
	r.Redundancy = 3

	minvolsize := r.MinVolumeSize()

	tests.Assert(t, minvolsize == BrickMinSize*8)
}
