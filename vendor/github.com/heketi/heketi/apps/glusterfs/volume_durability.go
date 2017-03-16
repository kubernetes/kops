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

type VolumeDurability interface {
	BrickSizeGenerator(size uint64) func() (int, uint64, error)
	MinVolumeSize() uint64
	BricksInSet() int
	SetDurability()
	SetExecutorVolumeRequest(v *executors.VolumeRequest)
}
