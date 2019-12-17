/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package mesh

import (
	"fmt"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/weaveworks/mesh"
	"k8s.io/kops/protokube/pkg/gossip"
)

// state is an implementation of a LWW map
type state struct {
	mtx  sync.RWMutex
	data KVState
	self mesh.PeerName

	lastSnapshot *gossip.GossipStateSnapshot
	version      uint64
}

//// state implements GossipData.
//var _ mesh.GossipData = &state{}

// Construct an empty state object, ready to receive updates.
// This is suitable to use at program start.
// Other peers will populate us with data.
func newState(self mesh.PeerName) *state {
	return &state{
		self: self,
	}
}

func (s *state) now() uint64 {
	// TODO: This relies on NTP.  We could have a g-counter or something, but this is probably good enough for V1
	// It's good enough for weave :-)
	return uint64(time.Now().Unix())
}

func (s *state) snapshot() *gossip.GossipStateSnapshot {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if s.lastSnapshot != nil && s.lastSnapshot.Version == s.version {
		// Potential optimization - this common path only needs a read-lock
		return s.lastSnapshot
	}

	values := make(map[string]string)
	for k, v := range s.data.Records {
		if v.Tombstone {
			continue
		}
		values[k] = string(v.Data)
	}

	snapshot := &gossip.GossipStateSnapshot{
		Values:  values,
		Version: s.version,
	}
	s.lastSnapshot = snapshot
	return snapshot
}

func (s *state) updateValues(removeKeys []string, putEntries map[string]string) {
	if len(removeKeys) == 0 && len(putEntries) == 0 {
		return
	}

	s.mtx.Lock()
	defer s.mtx.Unlock()

	now := s.now()

	if s.data.Records == nil {
		s.data.Records = make(map[string]*KVStateRecord)
	}

	for _, k := range removeKeys {
		v := &KVStateRecord{
			Tombstone: true,
			Version:   now,
		}

		s.data.Records[k] = v
	}

	for k, v := range putEntries {
		// TODO: Check that now > existing version?
		s.data.Records[k] = &KVStateRecord{
			Data:    []byte(v),
			Version: now,
		}
	}

	s.version++
}

// getData returns a copy of our data, suitable for gossiping
func (s *state) getData() *KVState {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	// make a deep-copy. To avoid a bunch of reflection etc. this simply marshals and unmarshals
	b, _ := proto.Marshal(&s.data)
	d, _ := DecodeKVState(b)
	return d
}

func (s *state) merge(message *KVState, changes *KVState) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	var c KVState
	if s.data.Records != nil {
		c.Records = make(map[string]*KVStateRecord)
		for k, v := range s.data.Records {
			c.Records[k] = v
		}
	}

	changed := mergeKVState(&c, message, changes)

	if changed {
		s.version++
		s.data = c
	}
}

var _ mesh.GossipData = &KVState{}

func mergeKVState(dest *KVState, src *KVState, changes *KVState) bool {
	changed := false

	if dest.Records == nil {
		dest.Records = make(map[string]*KVStateRecord)
	}

	if changes != nil && changes.Records == nil {
		changes.Records = make(map[string]*KVStateRecord)
	}

	for k, update := range src.Records {
		existing, found := dest.Records[k]
		if found && existing.Version >= update.Version {
			continue
		}
		dest.Records[k] = update

		// this is required for deltas
		if changes != nil {
			changes.Records[k] = update
		}

		changed = true
	}

	return changed
}

// Encode serializes our complete state to a slice of byte-slices.
// In this simple example, we use a single gob-encoded
// buffer: see https://golang.org/pkg/encoding/gob/
func (s *KVState) Encode() [][]byte {
	data, err := proto.Marshal(s)
	if err != nil {
		panic(fmt.Sprintf("error encoding gossip state: %v", err))
	}
	return [][]byte{data}
}

func DecodeKVState(data []byte) (*KVState, error) {
	state := &KVState{}
	err := proto.Unmarshal(data, state)
	if err != nil {
		return nil, fmt.Errorf("error decoding gossip state: %v", err)
	}
	return state, nil
}

// Merge merges the other GossipData into this one,
// and returns our resulting, complete state.
func (s *KVState) Merge(other mesh.GossipData) (complete mesh.GossipData) {
	otherState := other.(*KVState)

	mergeKVState(s, otherState, nil)

	return s
}
