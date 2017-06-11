package mesh

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TODO we should also test:
//
// - applying an incremental update, including the case where that
//   leads to an UnknownPeerError
//
// - the "improved update" calculation
//
// - non-gc of peers that are only referenced locally

func newNode(name PeerName) (*Peer, *Peers) {
	peer := newLocalPeer(name, "", nil)
	peers := newPeers(peer)
	return peer.Peer, peers
}

// Check that ApplyUpdate copies the whole topology from peers
func checkApplyUpdate(t *testing.T, peers *Peers) {
	dummyName, _ := PeerNameFromString("99:00:00:01:00:00")
	// We need a new node outside of the network, with a connection
	// into it.
	_, testBedPeers := newNode(dummyName)
	testBedPeers.AddTestConnection(peers.ourself.Peer)
	testBedPeers.applyUpdate(peers.encodePeers(peers.names()))

	checkTopologyPeers(t, true, testBedPeers.allPeersExcept(dummyName), peers.allPeers()...)
}

func TestPeersEncoding(t *testing.T) {
	const numNodes = 20
	const numIters = 1000
	var peer [numNodes]*Peer
	var ps [numNodes]*Peers
	for i := 0; i < numNodes; i++ {
		name, _ := PeerNameFromString(fmt.Sprintf("%02d:00:00:01:00:00", i))
		peer[i], ps[i] = newNode(name)
	}

	var conns []struct{ from, to int }
	for i := 0; i < numIters; i++ {
		oper := rand.Intn(2)
		switch oper {
		case 0:
			from, to := rand.Intn(numNodes), rand.Intn(numNodes)
			if from != to {
				if _, found := peer[from].connections[peer[to].Name]; !found {
					ps[from].AddTestConnection(peer[to])
					conns = append(conns, struct{ from, to int }{from, to})
					checkApplyUpdate(t, ps[from])
				}
			}
		case 1:
			if len(conns) > 0 {
				n := rand.Intn(len(conns))
				c := conns[n]
				ps[c.from].DeleteTestConnection(peer[c.to])
				ps[c.from].GarbageCollect()
				checkApplyUpdate(t, ps[c.from])
				conns = append(conns[:n], conns[n+1:]...)
			}
		}
	}
}

func garbageCollect(peers *Peers) []*Peer {
	var removed []*Peer
	peers.OnGC(func(peer *Peer) { removed = append(removed, peer) })
	peers.GarbageCollect()
	return removed
}

func TestPeersGarbageCollection(t *testing.T) {
	const (
		peer1NameString = "01:00:00:01:00:00"
		peer2NameString = "02:00:00:02:00:00"
		peer3NameString = "03:00:00:03:00:00"
	)
	var (
		peer1Name, _ = PeerNameFromString(peer1NameString)
		peer2Name, _ = PeerNameFromString(peer2NameString)
		peer3Name, _ = PeerNameFromString(peer3NameString)
	)

	// Create some peers with some connections to each other
	p1, ps1 := newNode(peer1Name)
	p2, ps2 := newNode(peer2Name)
	p3, ps3 := newNode(peer3Name)
	ps1.AddTestConnection(p2)
	ps2.AddTestRemoteConnection(p1, p2)
	ps2.AddTestConnection(p1)
	ps2.AddTestConnection(p3)
	ps3.AddTestConnection(p1)
	ps1.AddTestConnection(p3)
	ps2.AddTestRemoteConnection(p1, p3)
	ps2.AddTestRemoteConnection(p3, p1)

	// Every peer is referenced, so nothing should be dropped
	require.Empty(t, garbageCollect(ps1), "peers removed")
	require.Empty(t, garbageCollect(ps2), "peers removed")
	require.Empty(t, garbageCollect(ps3), "peers removed")

	// Drop the connection from 2 to 3, and 3 isn't garbage-collected
	// because 1 has a connection to 3
	ps2.DeleteTestConnection(p3)
	require.Empty(t, garbageCollect(ps2), "peers removed")

	// Drop the connection from 1 to 3, and 3 will get removed by
	// garbage-collection
	ps1.DeleteTestConnection(p3)
	checkPeerArray(t, garbageCollect(ps1), p3)
}

func TestShortIDCollisions(t *testing.T) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	_, peers := newNode(PeerName(1 << peerShortIDBits))

	// Make enough peers that short id collisions are
	// overwhelmingly likely
	ps := make([]*Peer, 1<<peerShortIDBits)
	for i := 0; i < 1<<peerShortIDBits; i++ {
		ps[i] = newPeer(PeerName(i), "", PeerUID(i), 0,
			PeerShortID(rng.Intn(1<<peerShortIDBits)))
	}

	shuffle := func() {
		for i := range ps {
			j := rng.Intn(i + 1)
			ps[i], ps[j] = ps[j], ps[i]
		}
	}

	// Fill peers
	shuffle()
	var pending peersPendingNotifications
	for _, p := range ps {
		peers.addByShortID(p, &pending)
	}

	// Check invariants
	counts := make([]int, 1<<peerShortIDBits)
	saw := func(p *Peer) {
		if p != peers.ourself.Peer {
			counts[p.UID]++
		}
	}

	for shortID, entry := range peers.byShortID {
		if entry.peer == nil {
			// no principal peer for this short id, so
			// others must be empty
			require.Empty(t, entry.others)
			continue
		}

		require.Equal(t, shortID, entry.peer.ShortID)
		saw(entry.peer)

		for _, p := range entry.others {
			saw(p)
			require.Equal(t, shortID, p.ShortID)

			// the principal peer should have the lowest name
			require.True(t, p.Name > entry.peer.Name)
		}
	}

	// Check that every peer was seen
	for _, n := range counts {
		require.Equal(t, 1, n)
	}

	// Delete all the peers
	shuffle()
	for _, p := range ps {
		peers.deleteByShortID(p, &pending)
	}

	for _, entry := range peers.byShortID {
		if entry.peer != peers.ourself.Peer {
			require.Nil(t, entry.peer)
		}

		require.Empty(t, entry.others)
	}
}

// Test the easy case of short id reassignment, when few short ids are taken
func TestShortIDReassignmentEasy(t *testing.T) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	_, peers := newNode(PeerName(0))

	for i := 1; i <= 10; i++ {
		peers.fetchWithDefault(newPeer(PeerName(i), "", PeerUID(i), 0,
			PeerShortID(rng.Intn(1<<peerShortIDBits))))
	}

	checkShortIDReassignment(t, peers)
}

// Test the hard case of short id reassignment, when most short ids are taken
func TestShortIDReassignmentHard(t *testing.T) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	_, peers := newNode(PeerName(1 << peerShortIDBits))

	// Take all short ids
	ps := make([]*Peer, 1<<peerShortIDBits)
	var pending peersPendingNotifications
	for i := 0; i < 1<<peerShortIDBits; i++ {
		ps[i] = newPeer(PeerName(i), "", PeerUID(i), 0,
			PeerShortID(i))
		peers.addByShortID(ps[i], &pending)
	}

	// As all short ids are taken, an attempted reassigment won't
	// do anything
	oldShortID := peers.ourself.ShortID
	require.False(t, peers.reassignLocalShortID(&pending))
	require.Equal(t, oldShortID, peers.ourself.ShortID)

	// Free up a few ids
	for i := 0; i < 10; i++ {
		x := rng.Intn(len(ps))
		if ps[x] != nil {
			peers.deleteByShortID(ps[x], &pending)
			ps[x] = nil
		}
	}

	checkShortIDReassignment(t, peers)
}

func checkShortIDReassignment(t *testing.T, peers *Peers) {
	oldShortID := peers.ourself.ShortID
	peers.reassignLocalShortID(&peersPendingNotifications{})
	require.NotEqual(t, oldShortID, peers.ourself.ShortID)
	require.Equal(t, peers.ourself.Peer, peers.byShortID[peers.ourself.ShortID].peer)
}

func TestShortIDInvalidation(t *testing.T) {
	_, peers := newNode(PeerName(1 << peerShortIDBits))

	// need to use a short id that is not the local peer's
	shortID := peers.ourself.ShortID + 1

	var pending peersPendingNotifications

	requireInvalidateShortIDs := func(expect bool) {
		require.Equal(t, expect, pending.invalidateShortIDs)
		pending.invalidateShortIDs = false
	}

	// The use of a fresh short id does not cause invalidation
	a := newPeer(PeerName(1), "", PeerUID(1), 0, shortID)
	peers.addByShortID(a, &pending)
	requireInvalidateShortIDs(false)

	// An addition which does not change the mapping
	b := newPeer(PeerName(2), "", PeerUID(2), 0, shortID)
	peers.addByShortID(b, &pending)
	requireInvalidateShortIDs(false)

	// An addition which does change the mapping
	c := newPeer(PeerName(0), "", PeerUID(0), 0, shortID)
	peers.addByShortID(c, &pending)
	requireInvalidateShortIDs(true)

	// A deletion which does not change the mapping
	peers.deleteByShortID(b, &pending)
	requireInvalidateShortIDs(false)

	// A deletion which does change the mapping
	peers.deleteByShortID(c, &pending)
	requireInvalidateShortIDs(true)

	// Deleting the last peer with a short id does not cause invalidation
	peers.deleteByShortID(a, &pending)
	requireInvalidateShortIDs(false)

	// .. but subsequent reuse of that short id does cause invalidation
	peers.addByShortID(a, &pending)
	requireInvalidateShortIDs(true)
}

func TestShortIDPropagation(t *testing.T) {
	_, peers1 := newNode(PeerName(1))
	_, peers2 := newNode(PeerName(2))

	peers1.AddTestConnection(peers2.ourself.Peer)
	peers1.applyUpdate(peers2.encodePeers(peers2.names()))
	peers12 := peers1.Fetch(PeerName(2))
	old := peers12.peerSummary

	require.True(t,
		peers2.reassignLocalShortID(&peersPendingNotifications{}))
	peers1.applyUpdate(peers2.encodePeers(peers2.names()))
	require.NotEqual(t, old.Version, peers12.Version)
	require.NotEqual(t, old.ShortID, peers12.ShortID)
}

func TestShortIDCollision(t *testing.T) {
	// Create 3 peers
	_, peers1 := newNode(PeerName(1))
	_, peers2 := newNode(PeerName(2))
	_, peers3 := newNode(PeerName(3))

	var pending peersPendingNotifications
	peers1.setLocalShortID(1, &pending)
	peers2.setLocalShortID(2, &pending)
	peers3.setLocalShortID(3, &pending)

	peers2.AddTestConnection(peers1.ourself.Peer)
	peers3.AddTestConnection(peers2.ourself.Peer)

	// Propogate from 1 to 2 to 3
	peers2.applyUpdate(peers1.encodePeers(peers1.names()))
	peers3.applyUpdate(peers2.encodePeers(peers2.names()))

	// Force the short id of peer 1 to collide with peer 2.  Peer
	// 1 has the lowest name, so it gets to keep the short id
	peers1.setLocalShortID(2, &pending)

	oldShortID := peers2.ourself.ShortID
	_, updated, _ := peers2.applyUpdate(peers1.encodePeers(peers1.names()))

	// peer 2 should have noticed the collision and resolved it
	require.NotEqual(t, oldShortID, peers2.ourself.ShortID)

	// The Peers do not have a Router, so broadcastPeerUpdate does
	// nothing in the context of this test.  So we fake what it
	// would do.
	updated[PeerName(2)] = struct{}{}

	// the update from peer 2 should include its short id change
	peers3.applyUpdate(peers2.encodePeers(updated))
	require.Equal(t, peers2.ourself.ShortID,
		peers3.Fetch(PeerName(2)).ShortID)
}

// Test the case where all short ids are taken, but then some peers go
// away, so the local peer reassigns
func TestDeferredShortIDReassignment(t *testing.T) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	_, us := newNode(PeerName(1 << peerShortIDBits))

	// Connect us to other peers occupying all short ids
	others := make([]*Peers, 1<<peerShortIDBits)
	var pending peersPendingNotifications
	for i := range others {
		_, others[i] = newNode(PeerName(i))
		others[i].setLocalShortID(PeerShortID(i), &pending)
		us.AddTestConnection(others[i].ourself.Peer)
	}

	// Check that, as expected, the local peer does not own its
	// short id
	require.NotEqual(t, us.ourself.Peer,
		us.byShortID[us.ourself.ShortID].peer)

	// Disconnect one peer, and we should now be able to claim its
	// short id
	other := others[rng.Intn(1<<peerShortIDBits)]
	us.DeleteTestConnection(other.ourself.Peer)
	us.GarbageCollect()

	require.Equal(t, us.ourself.Peer, us.byShortID[us.ourself.ShortID].peer)
}
