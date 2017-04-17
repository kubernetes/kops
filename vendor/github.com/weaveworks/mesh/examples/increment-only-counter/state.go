package main

import (
	"bytes"
	"sync"

	"encoding/gob"

	"github.com/weaveworks/mesh"
)

// state is an implementation of a G-counter.
type state struct {
	mtx  sync.RWMutex
	set  map[mesh.PeerName]int
	self mesh.PeerName
}

// state implements GossipData.
var _ mesh.GossipData = &state{}

// Construct an empty state object, ready to receive updates.
// This is suitable to use at program start.
// Other peers will populate us with data.
func newState(self mesh.PeerName) *state {
	return &state{
		set:  map[mesh.PeerName]int{},
		self: self,
	}
}

func (st *state) get() (result int) {
	st.mtx.RLock()
	defer st.mtx.RUnlock()
	for _, v := range st.set {
		result += v
	}
	return result
}

func (st *state) incr() (complete *state) {
	st.mtx.Lock()
	defer st.mtx.Unlock()
	st.set[st.self]++
	return &state{
		set: st.set,
	}
}

func (st *state) copy() *state {
	st.mtx.RLock()
	defer st.mtx.RUnlock()
	return &state{
		set: st.set,
	}
}

// Encode serializes our complete state to a slice of byte-slices.
// In this simple example, we use a single gob-encoded
// buffer: see https://golang.org/pkg/encoding/gob/
func (st *state) Encode() [][]byte {
	st.mtx.RLock()
	defer st.mtx.RUnlock()
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(st.set); err != nil {
		panic(err)
	}
	return [][]byte{buf.Bytes()}
}

// Merge merges the other GossipData into this one,
// and returns our resulting, complete state.
func (st *state) Merge(other mesh.GossipData) (complete mesh.GossipData) {
	return st.mergeComplete(other.(*state).copy().set)
}

// Merge the set into our state, abiding increment-only semantics.
// Return a non-nil mesh.GossipData representation of the received set.
func (st *state) mergeReceived(set map[mesh.PeerName]int) (received mesh.GossipData) {
	st.mtx.Lock()
	defer st.mtx.Unlock()

	for peer, v := range set {
		if v <= st.set[peer] {
			delete(set, peer) // optimization: make the forwarded data smaller
			continue
		}
		st.set[peer] = v
	}

	return &state{
		set: set, // all remaining elements were novel to us
	}
}

// Merge the set into our state, abiding increment-only semantics.
// Return any key/values that have been mutated, or nil if nothing changed.
func (st *state) mergeDelta(set map[mesh.PeerName]int) (delta mesh.GossipData) {
	st.mtx.Lock()
	defer st.mtx.Unlock()

	for peer, v := range set {
		if v <= st.set[peer] {
			delete(set, peer) // requirement: it's not part of a delta
			continue
		}
		st.set[peer] = v
	}

	if len(set) <= 0 {
		return nil // per OnGossip requirements
	}
	return &state{
		set: set, // all remaining elements were novel to us
	}
}

// Merge the set into our state, abiding increment-only semantics.
// Return our resulting, complete state.
func (st *state) mergeComplete(set map[mesh.PeerName]int) (complete mesh.GossipData) {
	st.mtx.Lock()
	defer st.mtx.Unlock()

	for peer, v := range set {
		if v > st.set[peer] {
			st.set[peer] = v
		}
	}

	return &state{
		set: st.set, // n.b. can't .copy() due to lock contention
	}
}
