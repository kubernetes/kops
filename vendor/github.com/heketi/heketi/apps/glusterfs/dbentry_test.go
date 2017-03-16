//
// Copyright (c) 2016 The heketi Authors
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
	"github.com/heketi/tests"
	"os"
	"testing"
	"time"
)

type testDbEntry struct {
}

func (t *testDbEntry) BucketName() string {
	return "TestBucket"
}

func (t *testDbEntry) Marshal() ([]byte, error) {
	return nil, nil
}

func (t *testDbEntry) Unmarshal(data []byte) error {
	return nil
}

func TestEntryRegister(t *testing.T) {
	tmpfile := tests.Tempfile()

	// Setup BoltDB database
	db, err := bolt.Open(tmpfile, 0600, &bolt.Options{Timeout: 3 * time.Second})
	tests.Assert(t, err == nil)
	defer os.Remove(tmpfile)

	// Create a bucket
	entry := &testDbEntry{}
	err = db.Update(func(tx *bolt.Tx) error {

		// Create Cluster Bucket
		_, err := tx.CreateBucketIfNotExists([]byte(entry.BucketName()))
		tests.Assert(t, err == nil)

		// Register a value
		_, err = EntryRegister(tx, entry, "mykey", []byte("myvalue"))
		tests.Assert(t, err == nil)

		return nil
	})
	tests.Assert(t, err == nil)

	// Try to write key again
	err = db.Update(func(tx *bolt.Tx) error {

		// Save again, it should not work
		val, err := EntryRegister(tx, entry, "mykey", []byte("myvalue"))
		tests.Assert(t, err == ErrKeyExists)
		tests.Assert(t, string(val) == "myvalue")

		// Remove key
		err = EntryDelete(tx, entry, "mykey")
		tests.Assert(t, err == nil)

		// Register again
		_, err = EntryRegister(tx, entry, "mykey", []byte("myvalue"))
		tests.Assert(t, err == nil)

		return nil
	})
	tests.Assert(t, err == nil)

}
