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
	"github.com/heketi/heketi/executors"
	"github.com/heketi/heketi/pkg/utils"
)

type CreateType int

const (
	CREATOR_CREATE CreateType = iota
	CREATOR_DESTROY
)

func createDestroyConcurrently(db *bolt.DB,
	executor executors.Executor,
	brick_entries []*BrickEntry,
	create_type CreateType) error {

	sg := utils.NewStatusGroup()

	// Create a goroutine for each brick
	for _, brick := range brick_entries {
		sg.Add(1)
		go func(b *BrickEntry) {
			defer sg.Done()
			if create_type == CREATOR_CREATE {
				sg.Err(b.Create(db, executor))
			} else {
				sg.Err(b.Destroy(db, executor))
			}
		}(brick)
	}

	// Wait here until all goroutines have returned.  If
	// any of errored, it would be cought here
	err := sg.Result()
	if err != nil {
		logger.Err(err)

		// Destroy all bricks and cleanup
		if create_type == CREATOR_CREATE {
			createDestroyConcurrently(db, executor, brick_entries, CREATOR_DESTROY)
		}
	}
	return err
}

func CreateBricks(db *bolt.DB, executor executors.Executor, brick_entries []*BrickEntry) error {
	return createDestroyConcurrently(db, executor, brick_entries, CREATOR_CREATE)
}

func DestroyBricks(db *bolt.DB, executor executors.Executor, brick_entries []*BrickEntry) error {
	return createDestroyConcurrently(db, executor, brick_entries, CREATOR_DESTROY)
}
