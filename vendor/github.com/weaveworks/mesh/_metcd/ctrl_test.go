package metcd

import (
	"log"
	"net"
	"os"
	"testing"
	"time"

	"github.com/coreos/etcd/raft/raftpb"

	"github.com/weaveworks/mesh"
	"github.com/weaveworks/mesh/meshconn"
)

func TestCtrlTerminates(t *testing.T) {
	var (
		peerName, _  = mesh.PeerNameFromString("01:23:45:67:89:01")
		self         = meshconn.MeshAddr{PeerName: peerName, PeerUID: 123}
		others       = []net.Addr{}
		minPeerCount = 5
		incomingc    = make(chan raftpb.Message)
		outgoingc    = make(chan raftpb.Message, 10000)
		unreachablec = make(chan uint64)
		confchangec  = make(chan raftpb.ConfChange)
		snapshotc    = make(chan raftpb.Snapshot, 10000)
		entryc       = make(chan raftpb.Entry)
		proposalc    = make(chan []byte)
		removedc     = make(chan struct{})
		logger       = log.New(os.Stderr, "", log.LstdFlags)
	)
	c := newCtrl(
		self,
		others,
		minPeerCount,
		incomingc,
		outgoingc,
		unreachablec,
		confchangec,
		snapshotc,
		entryc,
		proposalc,
		removedc,
		logger,
	)
	stopped := make(chan struct{})
	go func() {
		c.stop()
		close(stopped)
	}()
	select {
	case <-stopped:
		t.Log("ctrl terminated")
	case <-time.After(5 * time.Second):
		t.Fatal("ctrl didn't terminate")
	}
}
