package memberlist

import (
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"
	"k8s.io/klog"
	"k8s.io/kops/protokube/pkg/gossip"
	"k8s.io/kops/protokube/pkg/gossip/mesh"
)

type state struct {
	mtx  sync.RWMutex
	data mesh.KVState

	lastSnapshot *gossip.GossipStateSnapshot
	version      uint64
}

func (s *state) MarshalBinary() ([]byte, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	klog.V(4).Infof("Gossip => %v", s.data)
	return proto.Marshal(&s.data)
}

func (s *state) Merge(b []byte) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	var other mesh.KVState
	if err := proto.Unmarshal(b, &other); err != nil {
		return err
	}

	if s.data.Records == nil {
		s.data.Records = make(map[string]*mesh.KVStateRecord)
	}

	if other.Records == nil {
		other.Records = make(map[string]*mesh.KVStateRecord)
	}

	deltas := mesh.KVState{Records: make(map[string]*mesh.KVStateRecord)}
	for k, update := range other.Records {
		existing, found := s.data.Records[k]
		if found && existing.Version >= update.Version {
			continue
		}
		s.data.Records[k] = update
		deltas.Records[k] = update
	}

	if len(deltas.Records) == 0 {
		// per OnGossip requirements
		klog.V(4).Infof("MergeGossip %v => delta empty", other)
	} else {
		s.version++
		klog.V(4).Infof("MergeGossip %v => delta %v", other, deltas)
	}

	return nil
}

func (s *state) get(key string) []byte {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	v, found := s.data.Records[key]
	if !found {
		return nil
	}
	if v.Tombstone {
		return nil
	}
	return v.Data
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

func (s *state) put(key string, data []byte) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	now := s.now()

	v := &mesh.KVStateRecord{
		Data:    data,
		Version: now,
	}

	if s.data.Records == nil {
		s.data.Records = make(map[string]*mesh.KVStateRecord)
	}

	s.data.Records[key] = v
	s.version++
}

func (s *state) updateValues(removeKeys []string, putEntries map[string]string) {
	if len(removeKeys) == 0 && len(putEntries) == 0 {
		return
	}

	s.mtx.Lock()
	defer s.mtx.Unlock()

	now := s.now()

	if s.data.Records == nil {
		s.data.Records = make(map[string]*mesh.KVStateRecord)
	}

	for _, k := range removeKeys {
		v := &mesh.KVStateRecord{
			Tombstone: true,
			Version:   now,
		}

		s.data.Records[k] = v
	}

	for k, v := range putEntries {
		// TODO: Check that now > existing version?
		s.data.Records[k] = &mesh.KVStateRecord{
			Data:    []byte(v),
			Version: now,
		}
	}

	s.version++
}
