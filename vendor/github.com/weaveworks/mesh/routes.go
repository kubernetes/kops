package mesh

import (
	"math"
	"math/rand"
	"sync"
	"time"
)

type unicastRoutes map[PeerName]PeerName
type broadcastRoutes map[PeerName][]PeerName

// routes aggregates unicast and broadcast routes for our peer.
type routes struct {
	sync.RWMutex
	ourself       *localPeer
	peers         *Peers
	onChange      []func()
	unicast       unicastRoutes
	unicastAll    unicastRoutes // [1]
	broadcast     broadcastRoutes
	broadcastAll  broadcastRoutes // [1]
	recalcTimer   *time.Timer
	pendingRecalc bool
	wait          chan chan struct{}
	action        chan<- func()
	// [1] based on *all* connections, not just established &
	// symmetric ones
}

const (
	// We defer recalculation requests by up to 100ms, in order to
	// coalesce multiple recalcs together.
	recalcDeferTime = 100 * time.Millisecond
)

// newRoutes returns a usable Routes based on the LocalPeer and existing Peers.
func newRoutes(ourself *localPeer, peers *Peers) *routes {
	wait := make(chan chan struct{})
	action := make(chan func())
	r := &routes{
		ourself:      ourself,
		peers:        peers,
		unicast:      unicastRoutes{ourself.Name: UnknownPeerName},
		unicastAll:   unicastRoutes{ourself.Name: UnknownPeerName},
		broadcast:    broadcastRoutes{ourself.Name: []PeerName{}},
		broadcastAll: broadcastRoutes{ourself.Name: []PeerName{}},
		recalcTimer:  time.NewTimer(time.Hour),
		wait:         wait,
		action:       action,
	}
	r.recalcTimer.Stop()
	go r.run(wait, action)
	return r
}

// OnChange appends callback to the functions that will be called whenever the
// routes are recalculated.
func (r *routes) OnChange(callback func()) {
	r.Lock()
	defer r.Unlock()
	r.onChange = append(r.onChange, callback)
}

// PeerNames returns the peers that are accountd for in the r.
func (r *routes) PeerNames() peerNameSet {
	return r.peers.names()
}

// Unicast returns the next hop on the unicast route to the named peer,
// based on established and symmetric connections.
func (r *routes) Unicast(name PeerName) (PeerName, bool) {
	r.RLock()
	defer r.RUnlock()
	hop, found := r.unicast[name]
	return hop, found
}

// UnicastAll returns the next hop on the unicast route to the named peer,
// based on all connections.
func (r *routes) UnicastAll(name PeerName) (PeerName, bool) {
	r.RLock()
	defer r.RUnlock()
	hop, found := r.unicastAll[name]
	return hop, found
}

// Broadcast returns the set of peer names that should be notified
// when we receive a broadcast message originating from the named peer
// based on established and symmetric connections.
func (r *routes) Broadcast(name PeerName) []PeerName {
	return r.lookupOrCalculate(name, &r.broadcast, true)
}

// BroadcastAll returns the set of peer names that should be notified
// when we receive a broadcast message originating from the named peer
// based on all connections.
func (r *routes) BroadcastAll(name PeerName) []PeerName {
	return r.lookupOrCalculate(name, &r.broadcastAll, false)
}

func (r *routes) lookupOrCalculate(name PeerName, broadcast *broadcastRoutes, establishedAndSymmetric bool) []PeerName {
	r.RLock()
	hops, found := (*broadcast)[name]
	r.RUnlock()
	if found {
		return hops
	}
	res := make(chan []PeerName)
	r.action <- func() {
		r.RLock()
		hops, found := (*broadcast)[name]
		r.RUnlock()
		if found {
			res <- hops
			return
		}
		r.peers.RLock()
		r.ourself.RLock()
		hops = r.calculateBroadcast(name, establishedAndSymmetric)
		r.ourself.RUnlock()
		r.peers.RUnlock()
		res <- hops
		r.Lock()
		(*broadcast)[name] = hops
		r.Unlock()
	}
	return <-res
}

// RandomNeighbours chooses min(2 log2(n_peers), n_neighbouring_peers)
// neighbours, with a random distribution that is topology-sensitive,
// favouring neighbours at the end of "bottleneck links". We determine the
// latter based on the unicast routing table. If a neighbour appears as the
// value more frequently than others - meaning that we reach a higher
// proportion of peers via that neighbour than other neighbours - then it is
// chosen with a higher probability.
//
// Note that we choose 2log2(n_peers) *neighbours*, not peers. Consequently, on
// sparsely connected peers this function returns a higher proportion of
// neighbours than elsewhere. In extremis, on peers with fewer than
// log2(n_peers) neighbours, all neighbours are returned.
func (r *routes) randomNeighbours(except PeerName) []PeerName {
	r.RLock()
	defer r.RUnlock()
	var total int64 = 0
	weights := make(map[PeerName]int64)
	// First iterate the whole set, counting how often each neighbour appears
	for _, dst := range r.unicastAll {
		if dst != UnknownPeerName && dst != except {
			total++
			weights[dst]++
		}
	}
	needed := int(math.Min(2*math.Log2(float64(len(r.unicastAll))), float64(len(weights))))
	destinations := make([]PeerName, 0, needed)
	for len(destinations) < needed {
		// Pick a random point on the distribution and linear search for it
		rnd := rand.Int63n(total)
		for dst, count := range weights {
			if rnd < count {
				destinations = append(destinations, dst)
				// Remove the one we selected from consideration
				delete(weights, dst)
				total -= count
				break
			}
			rnd -= count
		}
	}
	return destinations
}

// Recalculate requests recalculation of the routing table. This is async but
// can effectively be made synchronous with a subsequent call to
// EnsureRecalculated.
func (r *routes) recalculate() {
	r.Lock()
	if !r.pendingRecalc {
		r.recalcTimer.Reset(recalcDeferTime)
		r.pendingRecalc = true
	}
	r.Unlock()
}

func (r *routes) clearPendingRecalcFlag() {
	r.Lock()
	r.pendingRecalc = false
	r.Unlock()
}

// EnsureRecalculated waits for any preceding Recalculate requests to finish.
func (r *routes) ensureRecalculated() {
	var done chan struct{}
	// If another call is already waiting, wait on the same chan, otherwise make a new one
	select {
	case done = <-r.wait:
	default:
		done = make(chan struct{})
	}
	r.wait <- done
	<-done
}

func (r *routes) run(wait <-chan chan struct{}, action <-chan func()) {
	for {
		select {
		case <-r.recalcTimer.C:
			r.clearPendingRecalcFlag()
			r.calculate()
		case done := <-wait:
			r.Lock()
			pending := r.pendingRecalc
			r.Unlock()
			if pending {
				<-r.recalcTimer.C
				r.clearPendingRecalcFlag()
				r.calculate()
			}
			close(done)
		case f := <-action:
			f()
		}
	}
}

// Calculate unicast and broadcast routes from r.ourself, and reset
// the broadcast route cache.
func (r *routes) calculate() {
	r.peers.RLock()
	r.ourself.RLock()
	var (
		unicast      = r.calculateUnicast(true)
		unicastAll   = r.calculateUnicast(false)
		broadcast    = make(broadcastRoutes)
		broadcastAll = make(broadcastRoutes)
	)
	broadcast[r.ourself.Name] = r.calculateBroadcast(r.ourself.Name, true)
	broadcastAll[r.ourself.Name] = r.calculateBroadcast(r.ourself.Name, false)
	r.ourself.RUnlock()
	r.peers.RUnlock()

	r.Lock()
	r.unicast = unicast
	r.unicastAll = unicastAll
	r.broadcast = broadcast
	r.broadcastAll = broadcastAll
	onChange := r.onChange
	r.Unlock()

	for _, callback := range onChange {
		callback()
	}
}

// Calculate all the routes for the question: if *we* want to send a
// packet to Peer X, what is the next hop?
//
// When we sniff a packet, we determine the destination peer
// ourself. Consequently, we can relay the packet via any
// arbitrary peers - the intermediate peers do not have to have
// any knowledge of the MAC address at all. Thus there's no need
// to exchange knowledge of MAC addresses, nor any constraints on
// the routes that we construct.
func (r *routes) calculateUnicast(establishedAndSymmetric bool) unicastRoutes {
	_, unicast := r.ourself.routes(nil, establishedAndSymmetric)
	return unicast
}

// Calculate the route to answer the question: if we receive a
// broadcast originally from Peer X, which peers should we pass the
// frames on to?
//
// When the topology is stable, and thus all peers perform route
// calculations based on the same data, the algorithm ensures that
// broadcasts reach every peer exactly once.
//
// This is largely due to properties of the Peer.Routes algorithm. In
// particular:
//
// ForAll X,Y,Z in Peers.
//     X.Routes(Y) <= X.Routes(Z) \/
//     X.Routes(Z) <= X.Routes(Y)
// ForAll X,Y,Z in Peers.
//     Y =/= Z /\ X.Routes(Y) <= X.Routes(Z) =>
//     X.Routes(Y) u [P | Y.HasSymmetricConnectionTo(P)] <= X.Routes(Z)
// where <= is the subset relationship on keys of the returned map.
func (r *routes) calculateBroadcast(name PeerName, establishedAndSymmetric bool) []PeerName {
	hops := []PeerName{}
	peer, found := r.peers.byName[name]
	if !found {
		return hops
	}
	if found, reached := peer.routes(r.ourself.Peer, establishedAndSymmetric); found {
		r.ourself.forEachConnectedPeer(establishedAndSymmetric, reached,
			func(remotePeer *Peer) { hops = append(hops, remotePeer.Name) })
	}
	return hops
}
