package metcd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"time"

	"github.com/coreos/etcd/raft"
	"github.com/coreos/etcd/raft/raftpb"
	"golang.org/x/net/context"

	"github.com/weaveworks/mesh"
	"github.com/weaveworks/mesh/meshconn"
)

// +-------------+   +-----------------+               +-------------------------+   +-------+
// | mesh.Router |   | packetTransport |               |          ctrl           |   | state |
// |             |   |                 |               |  +-------------------+  |   |       |
// |             |   |  +----------+   |               |  |     raft.Node     |  |   |       |
// |             |   |  | meshconn |   |               |  |                   |  |   |       |
// |             |======|  ReadFrom|-----incomingc------->|Step        Propose|<-----|    API|<---
// |             |   |  |   WriteTo|<--------outgoingc----|                   |  |   |       |
// |             |   |  +----------+   |               |  |                   |  |   |       |
// |             |   +-----------------+               |  |                   |  |   |       |
// |             |                                     |  |                   |  |   +-------+
// |             |   +------------+  +--------------+  |  |                   |  |     ^   ^
// |             |===| membership |->| configurator |---->|ProposeConfChange  |  |     |   |
// +-------------+   +------------+  +--------------+  |  |                   |  |     |   |
//                                          ^          |  +-------------------+  |     |   |
//                                          |          |       |         |       |     |   |
//                                          |          +-------|---------|-------+     |   |
//                H E R E                   |                entryc   snapshotc        |   |
//                  B E                     |                  |         |             |   |
//             D R A G O N S                |                  |         '-------------'   |
//                                          |                  v                           |
//                                          |  ConfChange +---------+ Normal               |
//                                          '-------------| demuxer |----------------------'
//                                                        +---------+

type ctrl struct {
	self         raft.Peer
	minPeerCount int
	incomingc    <-chan raftpb.Message    // from the transport
	outgoingc    chan<- raftpb.Message    // to the transport
	unreachablec <-chan uint64            // from the transport
	confchangec  <-chan raftpb.ConfChange // from the mesh
	snapshotc    chan<- raftpb.Snapshot   // to the state machine
	entryc       chan<- raftpb.Entry      // to the demuxer
	proposalc    <-chan []byte            // from the state machine
	stopc        chan struct{}            // from stop()
	removedc     chan<- struct{}          // to calling context
	terminatedc  chan struct{}
	storage      *raft.MemoryStorage
	node         raft.Node
	logger       mesh.Logger
}

func newCtrl(
	self net.Addr,
	others []net.Addr, // to join existing cluster, pass nil or empty others
	minPeerCount int,
	incomingc <-chan raftpb.Message,
	outgoingc chan<- raftpb.Message,
	unreachablec <-chan uint64,
	confchangec <-chan raftpb.ConfChange,
	snapshotc chan<- raftpb.Snapshot,
	entryc chan<- raftpb.Entry,
	proposalc <-chan []byte,
	removedc chan<- struct{},
	logger mesh.Logger,
) *ctrl {
	storage := raft.NewMemoryStorage()
	raftLogger := &raft.DefaultLogger{Logger: log.New(ioutil.Discard, "", 0)}
	raftLogger.EnableDebug()
	nodeConfig := &raft.Config{
		ID:              makeRaftPeer(self).ID,
		ElectionTick:    10,
		HeartbeatTick:   1,
		Storage:         storage,
		Applied:         0,    // starting fresh
		MaxSizePerMsg:   4096, // TODO(pb): looks like bytes; confirm that
		MaxInflightMsgs: 256,  // TODO(pb): copied from docs; confirm that
		CheckQuorum:     true, // leader steps down if quorum is not active for an electionTimeout
		Logger:          raftLogger,
	}

	startPeers := makeRaftPeers(others)
	if len(startPeers) == 0 {
		startPeers = nil // special case: join existing
	}
	node := raft.StartNode(nodeConfig, startPeers)

	c := &ctrl{
		self:         makeRaftPeer(self),
		minPeerCount: minPeerCount,
		incomingc:    incomingc,
		outgoingc:    outgoingc,
		unreachablec: unreachablec,
		confchangec:  confchangec,
		snapshotc:    snapshotc,
		entryc:       entryc,
		proposalc:    proposalc,
		stopc:        make(chan struct{}),
		removedc:     removedc,
		terminatedc:  make(chan struct{}),
		storage:      storage,
		node:         node,
		logger:       logger,
	}
	go c.driveRaft() // analagous to raftexample serveChannels
	return c
}

// It is a programmer error to call stop more than once.
func (c *ctrl) stop() {
	close(c.stopc)
	<-c.terminatedc
}

func (c *ctrl) driveRaft() {
	defer c.logger.Printf("ctrl: driveRaft loop exit")
	defer close(c.terminatedc)
	defer c.node.Stop()

	// We own driveProposals. We may terminate when the user invokes stop, or when
	// the Raft Node shuts down, which is generally when it receives a ConfChange
	// that removes it from the cluster. In either case, we kill driveProposals,
	// and wait for it to exit before returning.
	cancel := make(chan struct{})
	done := make(chan struct{})
	go func() {
		c.driveProposals(cancel)
		close(done)
	}()
	defer func() { <-done }() // order is important here
	defer close(cancel)       //

	// Now that we are holding a raft.Node we have a few responsibilities.
	// https://godoc.org/github.com/coreos/etcd/raft

	ticker := time.NewTicker(100 * time.Millisecond) // TODO(pb): taken from raftexample; need to validate
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.node.Tick()

		case r := <-c.node.Ready():
			if err := c.handleReady(r); err != nil {
				c.logger.Printf("ctrl: handle ready: %v (aborting)", err)
				close(c.removedc)
				return
			}

		case msg := <-c.incomingc:
			c.node.Step(context.TODO(), msg)

		case id := <-c.unreachablec:
			c.node.ReportUnreachable(id)

		case <-c.stopc:
			c.logger.Printf("ctrl: got stop signal")
			return
		}
	}
}

func (c *ctrl) driveProposals(cancel <-chan struct{}) {
	defer c.logger.Printf("ctrl: driveProposals loop exit")

	// driveProposals is a separate goroutine from driveRaft, to mirror
	// contrib/raftexample. To be honest, it's not clear to me why that should be
	// required; it seems like we should be able to drive these channels in the
	// same for/select loop as the others. But we have strange errors (likely
	// deadlocks) if we structure it that way.

	for c.proposalc != nil && c.confchangec != nil {
		select {
		case data, ok := <-c.proposalc:
			if !ok {
				c.logger.Printf("ctrl: got nil proposal; shutting down proposals")
				c.proposalc = nil
				continue
			}
			c.node.Propose(context.TODO(), data)

		case cc, ok := <-c.confchangec:
			if !ok {
				c.logger.Printf("ctrl: got nil conf change; shutting down conf changes")
				c.confchangec = nil
				continue
			}
			c.logger.Printf("ctrl: ProposeConfChange %s %x", cc.Type, cc.NodeID)
			c.node.ProposeConfChange(context.TODO(), cc)

		case <-cancel:
			return
		}
	}
}

func (c *ctrl) handleReady(r raft.Ready) error {
	// These steps may be performed in parallel, except as noted in step 2.
	//
	// 1. Write HardState, Entries, and Snapshot to persistent storage if they are
	// not empty. Note that when writing an Entry with Index i, any
	// previously-persisted entries with Index >= i must be discarded.
	if err := c.readySave(r.Snapshot, r.HardState, r.Entries); err != nil {
		return fmt.Errorf("save: %v", err)
	}

	// 2. Send all Messages to the nodes named in the To field. It is important
	// that no messages be sent until after the latest HardState has been persisted
	// to disk, and all Entries written by any previous Ready batch (Messages may
	// be sent while entries from the same batch are being persisted). If any
	// Message has type MsgSnap, call Node.ReportSnapshot() after it has been sent
	// (these messages may be large). Note: Marshalling messages is not
	// thread-safe; it is important that you make sure that no new entries are
	// persisted while marshalling. The easiest way to achieve this is to serialise
	// the messages directly inside your main raft loop.
	c.readySend(r.Messages)

	// 3. Apply Snapshot (if any) and CommittedEntries to the state machine. If any
	// committed Entry has Type EntryConfChange, call Node.ApplyConfChange() to
	// apply it to the node. The configuration change may be cancelled at this
	// point by setting the NodeID field to zero before calling ApplyConfChange
	// (but ApplyConfChange must be called one way or the other, and the decision
	// to cancel must be based solely on the state machine and not external
	// information such as the observed health of the node).
	if err := c.readyApply(r.Snapshot, r.CommittedEntries); err != nil {
		return fmt.Errorf("apply: %v", err)
	}

	// 4. Call Node.Advance() to signal readiness for the next batch of updates.
	// This may be done at any time after step 1, although all updates must be
	// processed in the order they were returned by Ready.
	c.readyAdvance()

	return nil
}

func (c *ctrl) readySave(snapshot raftpb.Snapshot, hardState raftpb.HardState, entries []raftpb.Entry) error {
	// For the moment, none of these steps persist to disk. That violates some Raft
	// invariants. But we are ephemeral, and will always boot empty, willingly
	// paying the snapshot cost. I trust that that the etcd Raft implementation
	// permits this.
	if !raft.IsEmptySnap(snapshot) {
		if err := c.storage.ApplySnapshot(snapshot); err != nil {
			return fmt.Errorf("apply snapshot: %v", err)
		}
	}
	if !raft.IsEmptyHardState(hardState) {
		if err := c.storage.SetHardState(hardState); err != nil {
			return fmt.Errorf("set hard state: %v", err)
		}
	}
	if err := c.storage.Append(entries); err != nil {
		return fmt.Errorf("append: %v", err)
	}
	return nil
}

func (c *ctrl) readySend(msgs []raftpb.Message) {
	for _, msg := range msgs {
		// If this fails, the transport will tell us asynchronously via unreachablec.
		c.outgoingc <- msg

		if msg.Type == raftpb.MsgSnap {
			// Assume snapshot sends always succeed.
			// TODO(pb): do we need error reporting?
			c.node.ReportSnapshot(msg.To, raft.SnapshotFinish)
		}
	}
}

func (c *ctrl) readyApply(snapshot raftpb.Snapshot, committedEntries []raftpb.Entry) error {
	c.snapshotc <- snapshot

	for _, committedEntry := range committedEntries {
		c.entryc <- committedEntry

		if committedEntry.Type == raftpb.EntryConfChange {
			// See raftexample raftNode.publishEntries
			var cc raftpb.ConfChange
			if err := cc.Unmarshal(committedEntry.Data); err != nil {
				return fmt.Errorf("unmarshal ConfChange: %v", err)
			}
			c.node.ApplyConfChange(cc)
			if cc.Type == raftpb.ConfChangeRemoveNode && cc.NodeID == c.self.ID {
				return errors.New("got ConfChange that removed me from the cluster; terminating")
			}
		}
	}

	return nil
}

func (c *ctrl) readyAdvance() {
	c.node.Advance()
}

// makeRaftPeer converts a net.Addr into a raft.Peer.
// All peers must perform the Addr-to-Peer mapping in the same way.
//
// The etcd Raft implementation tracks the committed entry for each node ID,
// and panics if it discovers a node has lost previously committed entries.
// In effect, it assumes commitment implies durability. But our storage is
// explicitly non-durable. So, whenever a node restarts, we need to give it
// a brand new ID. That is the peer UID.
func makeRaftPeer(addr net.Addr) raft.Peer {
	return raft.Peer{
		ID:      uint64(addr.(meshconn.MeshAddr).PeerUID),
		Context: nil, // TODO(pb): ??
	}
}

func makeRaftPeers(addrs []net.Addr) []raft.Peer {
	peers := make([]raft.Peer, len(addrs))
	for i, addr := range addrs {
		peers[i] = makeRaftPeer(addr)
	}
	return peers
}
