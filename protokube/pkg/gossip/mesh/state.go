package mesh

import (
	"bytes"
	"sync"

	"encoding/gob"

	"github.com/weaveworks/mesh"
	"k8s.io/kops/protokube/pkg/gossip"
	"time"
)

// state is an implementation of a LWW map
type state struct {
	mtx      sync.RWMutex
	valueMap map[string]record
	self     mesh.PeerName

	lastSnapshot *gossip.GossipStateSnapshot
	version      uint64
}

type record struct {
	//Key       string
	Data      []byte
	Tombstone bool

	// TODO: Rename to timestamp?
	Version uint64
}

// state implements GossipData.
var _ mesh.GossipData = &state{}

// Construct an empty state object, ready to receive updates.
// This is suitable to use at program start.
// Other peers will populate us with data.
func newState(self mesh.PeerName) *state {
	return &state{
		valueMap: make(map[string]record),
		self:     self,
	}
}

func (st *state) get(key string) []byte {
	st.mtx.RLock()
	defer st.mtx.RUnlock()

	v, found := st.valueMap[key]
	if !found {
		return nil
	}
	if v.Tombstone {
		return nil
	}
	return v.Data
}

func (st *state) now() uint64 {
	// This relies on NTP.  We could have a g-counter or something, but this is probably good enough for our purposes
	return uint64(time.Now().Unix())
}

func (s *state) snapshot() *gossip.GossipStateSnapshot {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if s.lastSnapshot != nil && s.lastSnapshot.Version == s.version {
		// TODO: We could put this branch under a read-lock
		return s.lastSnapshot
	}

	values := make(map[string]string)
	for k, v := range s.valueMap {
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
func (st *state) put(key string, data []byte) {
	st.mtx.Lock()
	defer st.mtx.Unlock()

	now := st.now()

	v := record{
		//Key:     key,
		Data:    data,
		Version: now,
	}
	st.valueMap[key] = v
	st.version++
}

func (s *state) updateValues(removeKeys []string, putEntries map[string]string) {
	if len(removeKeys) == 0 && len(putEntries) == 0 {
		return
	}

	s.mtx.Lock()
	defer s.mtx.Unlock()

	now := s.now()

	for _, k := range removeKeys {
		v := record{
			//Key:       k,
			Tombstone: true,
			Version:   now,
		}

		s.valueMap[k] = v
	}

	for k, v := range putEntries {
		// TODO: Check that now > existing version?
		s.valueMap[k] = record{
			//Key:     k,
			Data:    []byte(v),
			Version: now,
		}
	}

	s.version++
}

func (st *state) copy() *state {
	st.mtx.RLock()
	defer st.mtx.RUnlock()
	return &state{
		version: st.version,

		// TODO: This isn't immutable...
		valueMap: st.valueMap,
	}
}

// Encode serializes our complete state to a slice of byte-slices.
// In this simple example, we use a single gob-encoded
// buffer: see https://golang.org/pkg/encoding/gob/
func (st *state) Encode() [][]byte {
	st.mtx.RLock()
	defer st.mtx.RUnlock()
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(st.valueMap); err != nil {
		panic(err)
	}
	return [][]byte{buf.Bytes()}
}

// Merge merges the other GossipData into this one,
// and returns our resulting, complete state.
func (st *state) Merge(other mesh.GossipData) (complete mesh.GossipData) {
	otherState := other.(*state)

	st.merge(otherState.valueMap, nil)

	return st
}

func (st *state) merge(updates map[string]record, changes *map[string]record) {
	st.mtx.Lock()
	defer st.mtx.Unlock()

	changed := false

	for k := range updates {
		update := updates[k]

		existing, found := st.valueMap[k]
		if found && existing.Version >= update.Version {
			continue
		}
		st.valueMap[k] = update

		// this is required for deltas
		if changes != nil {
			(*changes)[k] = update
		}

		changed = true
	}

	if changed {
		st.version++
	}
}
