package metcd

import (
	"time"

	"github.com/coreos/etcd/raft/raftpb"

	"github.com/weaveworks/mesh"
)

// membership regularly polls the mesh.Router for peers in the mesh.
// New peer UIDs are sent on addc. Removed peer UIDs are sent on remc.
// If the membership set gets smaller than minCount, membership will
// close shrunkc and stop, and the caller should terminate.
type membership struct {
	router   *mesh.Router
	minCount int
	addc     chan<- uint64   // to configurator
	remc     chan<- uint64   // to configurator
	shrunkc  chan<- struct{} // to calling context
	quitc    chan struct{}
	logger   mesh.Logger
}

func newMembership(router *mesh.Router, initial uint64set, minCount int, addc, remc chan<- uint64, shrunkc chan<- struct{}, logger mesh.Logger) *membership {
	m := &membership{
		router:   router,
		minCount: minCount,
		addc:     addc,
		remc:     remc,
		shrunkc:  shrunkc,
		quitc:    make(chan struct{}),
		logger:   logger,
	}
	go m.loop(initial)
	return m
}

func (m *membership) stop() {
	close(m.quitc)
}

func (m *membership) loop(members uint64set) {
	defer m.logger.Printf("membership: loop exit")

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	var add, rem uint64set

	for {
		select {
		case <-ticker.C:
			add, rem, members = diff(members, membershipSet(m.router))
			if len(members) < m.minCount {
				m.logger.Printf("membership: member count (%d) shrunk beneath minimum (%d)", len(members), m.minCount)
				close(m.shrunkc)
				return
			}
			for id := range add {
				m.addc <- id
			}
			for id := range rem {
				m.remc <- id
			}
		case <-m.quitc:
			return
		}
	}
}

func membershipSet(router *mesh.Router) uint64set {
	descriptions := router.Peers.Descriptions()
	members := make(uint64set, len(descriptions))
	for _, description := range descriptions {
		members.add(uint64(description.UID))
	}
	return members
}

func diff(prev, curr uint64set) (add, rem, next uint64set) {
	add, rem, next = uint64set{}, uint64set{}, uint64set{}
	for i := range prev {
		prev.del(i)
		if curr.has(i) { // was in previous, still in current
			curr.del(i) // prevent it from being interpreted as new
			next.add(i) // promoted to next
		} else { // was in previous, no longer in current
			rem.add(i) // marked as removed
		}
	}
	for i := range curr {
		curr.del(i)
		add.add(i)
		next.add(i)
	}
	return add, rem, next
}

// configurator sits between the mesh membership subsystem and the raft.Node.
// When the mesh tells us that a peer is removed, the configurator adds that
// peer ID to a pending-remove set. Every tick, the configurator sends a
// ConfChange Remove proposal to the raft.Node for each peer in the
// pending-remove set. And when the configurator receives a committed ConfChange
// Remove entry for the peer, it removes the peer from the pending-remove set.
//
// We do the same thing for the add flow, for symmetry.
//
// Why is this necessary? Well, due to what looks like a bug in the raft.Node,
// ConfChange Remove proposals can get lost when the target node disappears. It
// is especially acute when the killed node is the leader. The current (or new)
// leader ends up spamming Heartbeats to the terminated node forever. So,
// lacking any obvious way to track the state of individual proposals, I've
// elected to continuously re-propose ConfChanges until they are confirmed i.e.
// committed.
type configurator struct {
	addc        <-chan uint64            // from membership
	remc        <-chan uint64            // from membership
	confchangec chan<- raftpb.ConfChange // to raft.Node
	entryc      <-chan raftpb.Entry      // from raft.Node
	quitc       chan struct{}
	logger      mesh.Logger
}

func newConfigurator(addc, remc <-chan uint64, confchangec chan<- raftpb.ConfChange, entryc <-chan raftpb.Entry, logger mesh.Logger) *configurator {
	c := &configurator{
		addc:        addc,
		remc:        remc,
		confchangec: confchangec,
		entryc:      entryc,
		quitc:       make(chan struct{}),
		logger:      logger,
	}
	go c.loop()
	return c
}

func (c *configurator) stop() {
	close(c.quitc)
}

func (c *configurator) loop() {
	defer c.logger.Printf("configurator: loop exit")

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	var (
		pendingAdd = uint64set{}
		pendingRem = uint64set{}
	)

	for {
		select {
		case id := <-c.addc:
			if pendingAdd.has(id) {
				c.logger.Printf("configurator: recv add %x, was pending add already", id)
			} else {
				c.logger.Printf("configurator: recv add %x, now pending add", id)
				pendingAdd.add(id)
				// We *must* wait before emitting a ConfChange.
				// https://github.com/coreos/etcd/issues/4759
			}

		case id := <-c.remc:
			if pendingRem.has(id) {
				c.logger.Printf("configurator: recv rem %x, was pending rem already", id)
			} else {
				c.logger.Printf("configurator: recv rem %x, now pending rem", id)
				pendingRem.add(id)
				// We *must* wait before emitting a ConfChange.
				// https://github.com/coreos/etcd/issues/4759
			}

		case <-ticker.C:
			for id := range pendingAdd {
				c.logger.Printf("configurator: send ConfChangeAddNode %x", id)
				c.confchangec <- raftpb.ConfChange{
					Type:   raftpb.ConfChangeAddNode,
					NodeID: id,
				}
			}
			for id := range pendingRem {
				c.logger.Printf("configurator: send ConfChangeRemoveNode %x", id)
				c.confchangec <- raftpb.ConfChange{
					Type:   raftpb.ConfChangeRemoveNode,
					NodeID: id,
				}
			}

		case entry := <-c.entryc:
			if entry.Type != raftpb.EntryConfChange {
				c.logger.Printf("configurator: ignoring %s", entry.Type)
				continue
			}
			var cc raftpb.ConfChange
			if err := cc.Unmarshal(entry.Data); err != nil {
				c.logger.Printf("configurator: got invalid ConfChange (%v); ignoring", err)
				continue
			}
			switch cc.Type {
			case raftpb.ConfChangeAddNode:
				if _, ok := pendingAdd[cc.NodeID]; ok {
					c.logger.Printf("configurator: recv %s %x: was pending add, deleting", cc.Type, cc.NodeID)
					delete(pendingAdd, cc.NodeID)
				} else {
					c.logger.Printf("configurator: recv %s %x: not pending add, ignoring", cc.Type, cc.NodeID)
				}
			case raftpb.ConfChangeRemoveNode:
				if _, ok := pendingRem[cc.NodeID]; ok {
					c.logger.Printf("configurator: recv %s %x: was pending rem, deleting", cc.Type, cc.NodeID)
					delete(pendingRem, cc.NodeID)
				} else {
					c.logger.Printf("configurator: recv %s %x: not pending rem, ignoring", cc.Type, cc.NodeID)
				}
			}

		case <-c.quitc:
			return
		}
	}
}

type uint64set map[uint64]struct{}

func (s uint64set) add(i uint64)      { s[i] = struct{}{} }
func (s uint64set) has(i uint64) bool { _, ok := s[i]; return ok }
func (s uint64set) del(i uint64)      { delete(s, i) }
