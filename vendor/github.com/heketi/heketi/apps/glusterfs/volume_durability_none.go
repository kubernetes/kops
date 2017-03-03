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
)

type NoneDurability struct {
	VolumeReplicaDurability
}

func NewNoneDurability() *NoneDurability {
	n := &NoneDurability{}
	n.Replica = 1

	return n
}

func (n *NoneDurability) SetDurability() {
	n.Replica = 1
}

func (n *NoneDurability) BricksInSet() int {
	return 1
}

func (n *NoneDurability) SetExecutorVolumeRequest(v *executors.VolumeRequest) {
	v.Type = executors.DurabilityNone
	v.Replica = n.Replica
}
