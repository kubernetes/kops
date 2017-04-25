// No mocks are tested by this file.
//
// It supplies some mock implementations to other unit tests, and is
// named "...test.go" so it is only compiled under `go test`.

package mesh

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

// Add to peers a connection from peers.ourself to p
func (peers *Peers) AddTestConnection(p *Peer) {
	summary := p.peerSummary
	summary.Version = 0
	toPeer := newPeerFromSummary(summary)
	toPeer = peers.fetchWithDefault(toPeer) // Has side-effect of incrementing refcount
	conn := newMockConnection(peers.ourself.Peer, toPeer)
	peers.ourself.addConnection(conn)
	peers.ourself.connectionEstablished(conn)
}

// Add to peers a connection from p1 to p2
func (peers *Peers) AddTestRemoteConnection(p1, p2 *Peer) {
	fromPeer := newPeerFrom(p1)
	fromPeer = peers.fetchWithDefault(fromPeer)
	toPeer := newPeerFrom(p2)
	toPeer = peers.fetchWithDefault(toPeer)
	peers.ourself.addConnection(newRemoteConnection(fromPeer, toPeer, "", false, false))
}

func (peers *Peers) DeleteTestConnection(p *Peer) {
	toName := p.Name
	toPeer := peers.Fetch(toName)
	peers.dereference(toPeer)
	conn, _ := peers.ourself.ConnectionTo(toName)
	peers.ourself.deleteConnection(conn)
}

// mockConnection used in testing is very similar to a
// RemoteConnection, without the RemoteTCPAddr(). We are making it a
// separate type in order to distinguish what is created by the test
// from what is created by the real code.
func newMockConnection(from, to *Peer) Connection {
	type mockConnection struct{ *remoteConnection }
	return &mockConnection{newRemoteConnection(from, to, "", false, false)}
}

func checkEqualConns(t *testing.T, ourName PeerName, got, wanted map[PeerName]Connection) {
	checkConns := make(peerNameSet)
	for _, conn := range wanted {
		checkConns[conn.Remote().Name] = struct{}{}
	}
	for _, conn := range got {
		remoteName := conn.Remote().Name
		if _, found := checkConns[remoteName]; found {
			delete(checkConns, remoteName)
		} else {
			require.FailNow(t, fmt.Sprintf("Unexpected connection from %s to %s", ourName, remoteName))
		}
	}
	if len(checkConns) > 0 {
		require.FailNow(t, fmt.Sprintf("Expected connections not found: from %s to %v", ourName, checkConns))
	}
}

// Get all the peers from a Peers in a slice
func (peers *Peers) allPeers() []*Peer {
	var res []*Peer
	for _, peer := range peers.byName {
		res = append(res, peer)
	}
	return res
}

func (peers *Peers) allPeersExcept(excludeName PeerName) []*Peer {
	res := peers.allPeers()
	for i, peer := range res {
		if peer.Name == excludeName {
			return append(res[:i], res[i+1:]...)
		}
	}
	return res
}

// Check that the peers slice matches the wanted peers
func checkPeerArray(t *testing.T, peers []*Peer, wantedPeers ...*Peer) {
	checkTopologyPeers(t, false, peers, wantedPeers...)
}

// Check that the peers slice matches the wanted peers and optionally
// all of their connections
func checkTopologyPeers(t *testing.T, checkConns bool, peers []*Peer, wantedPeers ...*Peer) {
	check := make(map[PeerName]*Peer)
	for _, peer := range wantedPeers {
		check[peer.Name] = peer
	}
	for _, peer := range peers {
		name := peer.Name
		if wantedPeer, found := check[name]; found {
			if checkConns {
				checkEqualConns(t, name, peer.connections, wantedPeer.connections)
			}
			delete(check, name)
		} else {
			require.FailNow(t, fmt.Sprintf("Unexpected peer: %s", name))
		}
	}
	if len(check) > 0 {
		require.FailNow(t, fmt.Sprintf("Expected peers not found: %v", check))
	}
}
