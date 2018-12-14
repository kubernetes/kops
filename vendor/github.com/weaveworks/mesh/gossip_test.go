package mesh

import (
	"fmt"
	"io/ioutil"
	"log"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

// TODO test gossip unicast; atm we only test topology gossip and
// surrogates, neither of which employ unicast.

type mockGossipConnection struct {
	remoteConnection
	dest    *Router
	senders *gossipSenders
	start   chan struct{}
}

var _ gossipConnection = &mockGossipConnection{}

func newTestRouter(name string) *Router {
	peerName, _ := PeerNameFromString(name)
	router := NewRouter(Config{}, peerName, "nick", nil, log.New(ioutil.Discard, "", 0))
	router.Start()
	return router
}

func (conn *mockGossipConnection) breakTie(dupConn ourConnection) connectionTieBreak {
	return tieBreakTied
}

func (conn *mockGossipConnection) shutdown(err error) {
}

func (conn *mockGossipConnection) logf(format string, args ...interface{}) {
	format = "->[" + conn.remoteTCPAddr + "|" + conn.remote.String() + "]: " + format
	if len(format) == 0 || format[len(format)-1] != '\n' {
		format += "\n"
	}
	fmt.Printf(format, args...)
}

func (conn *mockGossipConnection) SendProtocolMsg(pm protocolMsg) error {
	<-conn.start
	return conn.dest.handleGossip(pm.tag, pm.msg)
}

func (conn *mockGossipConnection) gossipSenders() *gossipSenders {
	return conn.senders
}

func (conn *mockGossipConnection) Start() {
	close(conn.start)
}

func sendPendingGossip(routers ...*Router) {
	// Loop until all routers report they didn't send anything
	for sentSomething := true; sentSomething; {
		sentSomething = false
		for _, router := range routers {
			sentSomething = router.sendPendingGossip() || sentSomething
		}
	}
}

func addTestGossipConnection(r1, r2 *Router) {
	c1 := r1.newTestGossipConnection(r2)
	c2 := r2.newTestGossipConnection(r1)
	c1.Start()
	c2.Start()
}

func (router *Router) newTestGossipConnection(r *Router) *mockGossipConnection {
	to := r.Ourself.Peer
	toPeer := newPeer(to.Name, to.NickName, to.UID, 0, to.ShortID)
	toPeer = router.Peers.fetchWithDefault(toPeer) // Has side-effect of incrementing refcount

	conn := &mockGossipConnection{
		remoteConnection: *newRemoteConnection(router.Ourself.Peer, toPeer, "", false, true),
		dest:             r,
		start:            make(chan struct{}),
	}
	conn.senders = newGossipSenders(conn, make(chan struct{}))
	router.Ourself.handleAddConnection(conn, false)
	router.Ourself.handleConnectionEstablished(conn)
	return conn
}

func (router *Router) DeleteTestGossipConnection(r *Router) {
	toName := r.Ourself.Peer.Name
	conn, _ := router.Ourself.ConnectionTo(toName)
	router.Peers.dereference(conn.Remote())
	router.Ourself.handleDeleteConnection(conn.(ourConnection))
}

// Create a Peer representing the receiver router, with connections to
// the routers supplied as arguments, carrying across all UID and
// version information.
func (router *Router) tp(routers ...*Router) *Peer {
	peer := newPeerFrom(router.Ourself.Peer)
	connections := make(map[PeerName]Connection)
	for _, r := range routers {
		p := newPeerFrom(r.Ourself.Peer)
		connections[r.Ourself.Peer.Name] = newMockConnection(peer, p)
	}
	peer.Version = router.Ourself.Peer.Version
	peer.connections = connections
	return peer
}

// Check that the topology of router matches the peers and all of their connections
func checkTopology(t *testing.T, router *Router, wantedPeers ...*Peer) {
	router.Peers.RLock()
	checkTopologyPeers(t, true, router.Peers.allPeers(), wantedPeers...)
	router.Peers.RUnlock()
}

func flushAndCheckTopology(t *testing.T, routers []*Router, wantedPeers ...*Peer) {
	sendPendingGossip(routers...)
	for _, r := range routers {
		checkTopology(t, r, wantedPeers...)
	}
}

func TestGossipTopology(t *testing.T) {
	// Create some peers that will talk to each other
	r1 := newTestRouter("01:00:00:01:00:00")
	r2 := newTestRouter("02:00:00:02:00:00")
	r3 := newTestRouter("03:00:00:03:00:00")
	routers := []*Router{r1, r2, r3}
	// Check state when they have no connections
	checkTopology(t, r1, r1.tp())
	checkTopology(t, r2, r2.tp())

	// Now try adding some connections
	addTestGossipConnection(r1, r2)
	sendPendingGossip(r1, r2)
	checkTopology(t, r1, r1.tp(r2), r2.tp(r1))
	checkTopology(t, r2, r1.tp(r2), r2.tp(r1))

	addTestGossipConnection(r2, r3)
	flushAndCheckTopology(t, routers, r1.tp(r2), r2.tp(r1, r3), r3.tp(r2))

	addTestGossipConnection(r3, r1)
	flushAndCheckTopology(t, routers, r1.tp(r2, r3), r2.tp(r1, r3), r3.tp(r1, r2))

	// Drop the connection from 2 to 3
	r2.DeleteTestGossipConnection(r3)
	flushAndCheckTopology(t, routers, r1.tp(r2, r3), r2.tp(r1), r3.tp(r1, r2))

	// Drop the connection from 1 to 3
	r1.DeleteTestGossipConnection(r3)
	sendPendingGossip(r1, r2, r3)
	checkTopology(t, r1, r1.tp(r2), r2.tp(r1))
	checkTopology(t, r2, r1.tp(r2), r2.tp(r1))
	// r3 still thinks r1 has a connection to it
	checkTopology(t, r3, r1.tp(r2, r3), r2.tp(r1), r3.tp(r1, r2))
}

func TestGossipSurrogate(t *testing.T) {
	// create the topology r1 <-> r2 <-> r3
	r1 := newTestRouter("01:00:00:01:00:00")
	r2 := newTestRouter("02:00:00:02:00:00")
	r3 := newTestRouter("03:00:00:03:00:00")
	routers := []*Router{r1, r2, r3}
	addTestGossipConnection(r1, r2)
	addTestGossipConnection(r3, r2)
	flushAndCheckTopology(t, routers, r1.tp(r2), r2.tp(r1, r3), r3.tp(r2))

	// create a gossiper at either end, but not the middle
	g1 := newTestGossiper()
	g3 := newTestGossiper()
	s1 := r1.NewGossip("Test", g1)
	s3 := r3.NewGossip("Test", g3)

	// broadcast a message from each end, check it reaches the other
	broadcast(s1, 1)
	broadcast(s3, 2)
	sendPendingGossip(r1, r2, r3)
	g1.checkHas(t, 2)
	g3.checkHas(t, 1)

	// check that each end gets their message back through periodic
	// gossip
	r1.sendAllGossip()
	r3.sendAllGossip()
	sendPendingGossip(r1, r2, r3)
	g1.checkHas(t, 1, 2)
	g3.checkHas(t, 1, 2)
}

type testGossiper struct {
	sync.RWMutex
	state map[byte]struct{}
}

func newTestGossiper() *testGossiper {
	return &testGossiper{state: make(map[byte]struct{})}
}

func (g *testGossiper) OnGossipUnicast(sender PeerName, msg []byte) error {
	return nil
}

func (g *testGossiper) OnGossipBroadcast(_ PeerName, update []byte) (GossipData, error) {
	g.Lock()
	defer g.Unlock()
	for _, v := range update {
		g.state[v] = struct{}{}
	}
	return newSurrogateGossipData(update), nil
}

func (g *testGossiper) Gossip() GossipData {
	g.RLock()
	defer g.RUnlock()
	state := make([]byte, len(g.state))
	for v := range g.state {
		state = append(state, v)
	}
	return newSurrogateGossipData(state)
}

func (g *testGossiper) OnGossip(update []byte) (GossipData, error) {
	g.Lock()
	defer g.Unlock()
	var delta []byte
	for _, v := range update {
		if _, found := g.state[v]; !found {
			delta = append(delta, v)
			g.state[v] = struct{}{}
		}
	}
	if len(delta) == 0 {
		return nil, nil
	}
	return newSurrogateGossipData(delta), nil
}

func (g *testGossiper) checkHas(t *testing.T, vs ...byte) {
	g.RLock()
	defer g.RUnlock()
	for _, v := range vs {
		if _, found := g.state[v]; !found {
			require.FailNow(t, fmt.Sprintf("%d is missing", v))
		}
	}
}

func broadcast(s Gossip, v byte) {
	s.GossipBroadcast(newSurrogateGossipData([]byte{v}))
}
