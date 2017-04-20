package ipam

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"sort"
	"time"

	"github.com/weaveworks/mesh"

	"github.com/weaveworks/weave/common"
	"github.com/weaveworks/weave/db"
	"github.com/weaveworks/weave/ipam/paxos"
	"github.com/weaveworks/weave/ipam/ring"
	"github.com/weaveworks/weave/ipam/space"
	"github.com/weaveworks/weave/ipam/tracker"
	"github.com/weaveworks/weave/net/address"
)

// Kinds of message we can unicast to other peers
const (
	msgSpaceRequest = iota
	msgRingUpdate
	msgSpaceRequestDenied

	tickInterval         = time.Second * 5
	MinSubnetSize        = 4 // first and last addresses are excluded, so 2 would be too small
	containerDiedTimeout = time.Second * 30
)

// operation represents something which Allocator wants to do, but
// which may need to wait until some other message arrives.
type operation interface {
	// Try attempts this operations and returns false if needs to be tried again.
	Try(alloc *Allocator) bool

	Cancel()

	// Does this operation pertain to the given container id?
	// Used for tidying up pending operations when containers die.
	ForContainer(ident string) bool
}

// This type is persisted hence all fields exported
type ownedData struct {
	IsContainer bool
	Cidrs       []address.CIDR
}

// Allocator brings together Ring and space.Set, and does the
// necessary plumbing.  Runs as a single-threaded Actor, so no locks
// are used around data structures.
type Allocator struct {
	actionChan        chan<- func()
	stopChan          chan<- struct{}
	ourName           mesh.PeerName
	seed              []mesh.PeerName          // optional user supplied ring seed
	universe          address.CIDR             // superset of all ranges
	ring              *ring.Ring               // information on ranges owned by all peers
	space             space.Space              // more detail on ranges owned by us
	owned             map[string]ownedData     // who owns what addresses, indexed by container-ID
	nicknames         map[mesh.PeerName]string // so we can map nicknames for rmpeer
	pendingAllocates  []operation              // held until we get some free space
	pendingClaims     []operation              // held until we know who owns the space
	pendingPrimes     []operation              // held while our ring is empty
	dead              map[string]time.Time     // containers we heard were dead, and when
	db                db.DB                    // persistence
	gossip            mesh.Gossip              // our link to the outside world for sending messages
	paxos             paxos.Participant
	awaitingConsensus bool
	ticker            *time.Ticker
	shuttingDown      bool // to avoid doing any requests while trying to shut down
	isKnownPeer       func(mesh.PeerName) bool
	quorum            func() uint
	now               func() time.Time
}

// PreClaims are IP addresses discovered before we could initialize IPAM
type PreClaim struct {
	Ident       string // a container ID, something like "weave:expose", or api.NoContainerID
	IsContainer bool   // true if Ident is a container ID
	Cidr        address.CIDR
}

type Config struct {
	OurName     mesh.PeerName
	OurUID      mesh.PeerUID
	OurNickname string
	Seed        []mesh.PeerName
	Universe    address.CIDR
	IsObserver  bool
	PreClaims   []PreClaim
	Quorum      func() uint
	Db          db.DB
	IsKnownPeer func(name mesh.PeerName) bool
	Tracker     tracker.LocalRangeTracker
}

// NewAllocator creates and initialises a new Allocator
func NewAllocator(config Config) *Allocator {
	var participant paxos.Participant
	var alloc *Allocator
	var onUpdate ring.OnUpdate

	if config.IsObserver {
		participant = paxos.NewObserver()
	} else {
		participant = paxos.NewNode(config.OurName, config.OurUID, 0)
	}

	if config.Tracker != nil {
		onUpdate = func(prev []address.Range, curr []address.Range, local bool) {
			if err := config.Tracker.HandleUpdate(prev, curr, local); err != nil {
				alloc.errorf("HandleUpdate failed: %s", err)
			}
		}
	}

	alloc = &Allocator{
		ourName:     config.OurName,
		seed:        config.Seed,
		universe:    config.Universe,
		ring:        ring.New(config.Universe.Range().Start, config.Universe.Range().End, config.OurName, onUpdate),
		owned:       make(map[string]ownedData),
		db:          config.Db,
		paxos:       participant,
		nicknames:   map[mesh.PeerName]string{config.OurName: config.OurNickname},
		isKnownPeer: config.IsKnownPeer,
		quorum:      config.Quorum,
		dead:        make(map[string]time.Time),
		now:         time.Now,
	}

	alloc.pendingClaims = make([]operation, len(config.PreClaims))
	for i, c := range config.PreClaims {
		alloc.pendingClaims[i] = &claim{ident: c.Ident, cidr: c.Cidr}
	}

	return alloc
}

func ParseCIDRSubnet(cidrStr string) (cidr address.CIDR, err error) {
	cidr, err = address.ParseCIDR(cidrStr)
	if err != nil {
		return
	}
	if !cidr.IsSubnet() {
		err = fmt.Errorf("invalid subnet - bits after network prefix are not all zero: %s", cidrStr)
	}
	if cidr.Size() < MinSubnetSize {
		err = fmt.Errorf("invalid subnet - smaller than minimum size %d: %s", MinSubnetSize, cidrStr)
	}
	return
}

// Start runs the allocator goroutine
func (alloc *Allocator) Start() {
	loadedPersistedData := alloc.loadPersistedData()
	switch {
	case loadedPersistedData && len(alloc.seed) != 0:
		alloc.infof("Found persisted IPAM data, ignoring supplied IPAM seed")
	case loadedPersistedData:
		alloc.infof("Initialising with persisted data")
	case len(alloc.seed) != 0:
		alloc.infof("Initialising with supplied IPAM seed")
		alloc.createRing(alloc.seed)
	case alloc.paxos.IsElector():
		alloc.infof("Initialising via deferred consensus")
	default:
		alloc.infof("Initialising as observer - awaiting IPAM data from another peer")
	}
	if loadedPersistedData { // do any pre-claims right away
		alloc.tryOps(&alloc.pendingClaims)
	}
	actionChan := make(chan func(), mesh.ChannelSize)
	stopChan := make(chan struct{})
	alloc.actionChan = actionChan
	alloc.stopChan = stopChan
	alloc.ticker = time.NewTicker(tickInterval)
	go alloc.actorLoop(actionChan, stopChan)
}

// Stop makes the actor routine exit, for test purposes ONLY because any
// calls after this is processed will hang. Async.
func (alloc *Allocator) Stop() {
	select {
	case alloc.stopChan <- struct{}{}:
	default:
	}
}

// Operation life cycle

// Given an operation, try it, and add it to the pending queue if it didn't succeed
func (alloc *Allocator) doOperation(op operation, ops *[]operation) {
	alloc.actionChan <- func() {
		if alloc.shuttingDown {
			op.Cancel()
			return
		}
		if !op.Try(alloc) {
			*ops = append(*ops, op)
		}
	}
}

// Given an operation, remove it from the pending queue
//  Note the op may not be on the queue; it may have
//  already succeeded.  If it is on the queue, we call
//  cancel on it, allowing callers waiting for the resultChans
//  to unblock.
func (alloc *Allocator) cancelOp(opToCancel operation, ops *[]operation) {
	for i, op := range *ops {
		if op == opToCancel {
			*ops = append((*ops)[:i], (*ops)[i+1:]...)
			op.Cancel()
			break
		}
	}
}

// Cancel all operations in a queue
func (alloc *Allocator) cancelOps(ops *[]operation) {
	for _, op := range *ops {
		op.Cancel()
	}
	*ops = []operation{}
}

// Cancel all operations for a given container id, returns true
// if we found any.
func (alloc *Allocator) cancelOpsFor(ops *[]operation, ident string) bool {
	var found bool
	for i := 0; i < len(*ops); {
		if op := (*ops)[i]; op.ForContainer(ident) {
			found = true
			op.Cancel()
			*ops = append((*ops)[:i], (*ops)[i+1:]...)
		} else {
			i++
		}
	}
	return found
}

// Try all operations in a queue
func (alloc *Allocator) tryOps(ops *[]operation) {
	for i := 0; i < len(*ops); {
		op := (*ops)[i]
		if !op.Try(alloc) {
			i++
			continue
		}
		*ops = append((*ops)[:i], (*ops)[i+1:]...)
	}
}

// Try all pending operations
func (alloc *Allocator) tryPendingOps() {
	// Unblock pending primes first
	alloc.tryOps(&alloc.pendingPrimes)
	// Process existing claims before servicing new allocations
	alloc.tryOps(&alloc.pendingClaims)
	alloc.tryOps(&alloc.pendingAllocates)
}

func (alloc *Allocator) havePendingOps() bool {
	return len(alloc.pendingPrimes)+len(alloc.pendingClaims)+len(alloc.pendingAllocates) > 0
}

func (alloc *Allocator) spaceRequestDenied(sender mesh.PeerName, r address.Range) {
	for i := 0; i < len(alloc.pendingClaims); {
		claim := alloc.pendingClaims[i].(*claim)
		if r.Contains(claim.cidr.Addr) {
			claim.deniedBy(alloc, sender)
			alloc.pendingClaims = append(alloc.pendingClaims[:i], alloc.pendingClaims[i+1:]...)
			continue
		}
		i++
	}
}

type errorCancelled struct {
	kind  string
	ident string
}

func (e *errorCancelled) Error() string {
	return fmt.Sprintf("%s request for %s cancelled", e.kind, e.ident)
}

// Actor client API

// Prime (Sync) - wait for consensus
func (alloc *Allocator) Prime() {
	resultChan := make(chan struct{})
	op := &prime{resultChan: resultChan}
	alloc.doOperation(op, &alloc.pendingPrimes)
	<-resultChan
}

// Allocate (Sync) - get new IP address for container with given name in range
// if there isn't any space in that range we block indefinitely
func (alloc *Allocator) Allocate(ident string, r address.CIDR, isContainer bool, hasBeenCancelled func() bool) (address.Address, error) {
	resultChan := make(chan allocateResult)
	op := &allocate{
		resultChan:       resultChan,
		ident:            ident,
		r:                r,
		isContainer:      isContainer,
		hasBeenCancelled: hasBeenCancelled,
	}
	alloc.doOperation(op, &alloc.pendingAllocates)
	result := <-resultChan
	return result.addr, result.err
}

// Lookup (Sync) - get existing IP addresses for container with given name in range
func (alloc *Allocator) Lookup(ident string, r address.Range) ([]address.CIDR, error) {
	resultChan := make(chan []address.CIDR)
	alloc.actionChan <- func() {
		resultChan <- alloc.ownedInRange(ident, r)
	}
	return <-resultChan, nil
}

// Claim an address that we think we should own (Sync)
func (alloc *Allocator) Claim(ident string, cidr address.CIDR, isContainer, noErrorOnUnknown bool, hasBeenCancelled func() bool) error {
	resultChan := make(chan error)
	op := &claim{
		resultChan:       resultChan,
		ident:            ident,
		cidr:             cidr,
		isContainer:      isContainer,
		noErrorOnUnknown: noErrorOnUnknown,
		hasBeenCancelled: hasBeenCancelled,
	}
	alloc.doOperation(op, &alloc.pendingClaims)
	return <-resultChan
}

// ContainerDied called from the updater interface.  Async.
func (alloc *Allocator) ContainerDied(ident string) {
	alloc.actionChan <- func() {
		if alloc.hasOwnedByContainer(ident) {
			alloc.debugln("Container", ident, "died; noting to remove later")
			alloc.dead[ident] = alloc.now()
		}
		// Also remove any pending ops
		alloc.cancelOpsFor(&alloc.pendingAllocates, ident)
		alloc.cancelOpsFor(&alloc.pendingClaims, ident)
	}
}

// ContainerDestroyed called from the updater interface.  Async.
func (alloc *Allocator) ContainerDestroyed(ident string) {
	alloc.actionChan <- func() {
		if alloc.hasOwnedByContainer(ident) {
			alloc.debugln("Container", ident, "destroyed; removing addresses")
			alloc.delete(ident)
			delete(alloc.dead, ident)
		}
	}
}

func (alloc *Allocator) removeDeadContainers() {
	cutoff := alloc.now().Add(-containerDiedTimeout)
	for ident, timeOfDeath := range alloc.dead {
		if timeOfDeath.Before(cutoff) {
			if err := alloc.delete(ident); err == nil {
				alloc.debugln("Removed addresses for container", ident)
			}
			delete(alloc.dead, ident)
		}
	}
}

func (alloc *Allocator) ContainerStarted(ident string) {
	alloc.actionChan <- func() {
		delete(alloc.dead, ident) // delete is no-op if key not in map
	}
}

func (alloc *Allocator) PruneOwned(ids []string) {
	idmap := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		idmap[id] = struct{}{}
	}
	alloc.actionChan <- func() {
		alloc.pruneOwned(idmap)
	}
}

// Delete (Sync) - release all IP addresses for container with given name
func (alloc *Allocator) Delete(ident string) error {
	errChan := make(chan error)
	alloc.actionChan <- func() {
		errChan <- alloc.delete(ident)
	}
	return <-errChan
}

func (alloc *Allocator) delete(ident string) error {
	cidrs := alloc.removeAllOwned(ident)
	if len(cidrs) == 0 {
		return fmt.Errorf("Delete: no addresses for %s", ident)
	}
	for _, cidr := range cidrs {
		alloc.space.Free(cidr.Addr)
	}
	return nil
}

// Free (Sync) - release single IP address for container
func (alloc *Allocator) Free(ident string, addrToFree address.Address) error {
	errChan := make(chan error)
	alloc.actionChan <- func() {
		if alloc.removeOwned(ident, addrToFree) {
			alloc.debugln("Freed", addrToFree, "for", ident)
			alloc.space.Free(addrToFree)
			errChan <- nil
			return
		}

		errChan <- fmt.Errorf("Free: address %s not found for %s", addrToFree, ident)
	}
	return <-errChan
}

func (alloc *Allocator) pickPeerFromNicknames(isValid func(mesh.PeerName) bool) mesh.PeerName {
	for name := range alloc.nicknames {
		if name != alloc.ourName && isValid(name) {
			return name
		}
	}
	return mesh.UnknownPeerName
}

func (alloc *Allocator) pickPeerForTransfer() mesh.PeerName {
	// first try alive peers that actively participate in IPAM (i.e. have entries)
	if heir := alloc.ring.PickPeerForTransfer(alloc.isKnownPeer); heir != mesh.UnknownPeerName {
		return heir
	}
	// next try alive peers that have IPAM enabled but have no entries
	if heir := alloc.pickPeerFromNicknames(alloc.isKnownPeer); heir != mesh.UnknownPeerName {
		return heir
	}
	// next try disappeared peers that still have entries
	t := func(mesh.PeerName) bool { return true }
	if heir := alloc.ring.PickPeerForTransfer(t); heir != mesh.UnknownPeerName {
		return heir
	}
	// finally, disappeared peers that passively participated in IPAM
	return alloc.pickPeerFromNicknames(t)
}

// Shutdown (Sync)
func (alloc *Allocator) Shutdown() {
	alloc.infof("Shutdown")
	doneChan := make(chan struct{})
	alloc.actionChan <- func() {
		alloc.shuttingDown = true
		alloc.cancelOps(&alloc.pendingClaims)
		alloc.cancelOps(&alloc.pendingAllocates)
		alloc.cancelOps(&alloc.pendingPrimes)
		heir := alloc.pickPeerForTransfer()
		alloc.ring.Transfer(alloc.ourName, heir)
		alloc.space.Clear()
		if heir != mesh.UnknownPeerName {
			alloc.persistRing()
			alloc.gossip.GossipBroadcast(alloc.Gossip())
		}
		doneChan <- struct{}{}
	}
	<-doneChan
}

// AdminTakeoverRanges (Sync) - take over the ranges owned by a given
// peer, and return how much space was transferred in the process.
// Only done on administrator command.
func (alloc *Allocator) AdminTakeoverRanges(peerNameOrNickname string) address.Count {
	resultChan := make(chan address.Count)
	alloc.actionChan <- func() {
		peername, err := alloc.lookupPeername(peerNameOrNickname)
		if err != nil {
			alloc.warnf("attempt to take over range from unknown peer '%s'", peerNameOrNickname)
			resultChan <- address.Count(0)
			return
		}

		alloc.debugln("AdminTakeoverRanges:", peername)
		if peername == alloc.ourName {
			alloc.warnf("attempt to take over range from ourself")
			resultChan <- address.Count(0)
			return
		}

		newRanges := alloc.ring.Transfer(peername, alloc.ourName)

		if len(newRanges) == 0 {
			resultChan <- address.Count(0)
			return
		}

		before := alloc.space.NumFreeAddresses()
		alloc.ringUpdated()
		after := alloc.space.NumFreeAddresses()

		alloc.gossip.GossipBroadcast(alloc.Gossip())

		resultChan <- after - before
	}
	return <-resultChan
}

// Lookup a PeerName by nickname or stringified PeerName.  We can't
// call into the router for this because we are interested in peers
// that have gone away but are still in the ring, which is why we
// maintain our own nicknames map.
func (alloc *Allocator) lookupPeername(name string) (mesh.PeerName, error) {
	for peername, nickname := range alloc.nicknames {
		if nickname == name {
			return peername, nil
		}
	}

	return mesh.PeerNameFromString(name)
}

// Restrict the peers in "nicknames" to those in the ring plus peers known to the router
func (alloc *Allocator) pruneNicknames() {
	ringPeers := alloc.ring.PeerNames()
	for name := range alloc.nicknames {
		if _, ok := ringPeers[name]; !ok && !alloc.isKnownPeer(name) {
			delete(alloc.nicknames, name)
		}
	}
}

func (alloc *Allocator) annotatePeernames(names []mesh.PeerName) []string {
	var res []string
	for _, name := range names {
		if nickname, found := alloc.nicknames[name]; found {
			res = append(res, fmt.Sprint(name, "(", nickname, ")"))
		} else {
			res = append(res, name.String())
		}
	}
	return res
}

// PeerGone removes nicknames of peers which are no longer mentioned
// in the ring. Async.
//
// NB: the function is invoked by the gossip library routines and should be
//     registered manually.
func (alloc *Allocator) PeerGone(peerName mesh.PeerName) {
	alloc.debugf("PeerGone: peer %s", peerName)

	alloc.actionChan <- func() {
		ringPeers := alloc.ring.PeerNames()
		if _, ok := ringPeers[peerName]; !ok {
			delete(alloc.nicknames, peerName)
		}
	}
}

func decodeRange(msg []byte) (r address.Range, err error) {
	decoder := gob.NewDecoder(bytes.NewReader(msg))
	return r, decoder.Decode(&r)
}

// OnGossipUnicast (Sync)
func (alloc *Allocator) OnGossipUnicast(sender mesh.PeerName, msg []byte) error {
	alloc.debugln("OnGossipUnicast from", sender, ": ", len(msg), "bytes")
	resultChan := make(chan error)
	alloc.actionChan <- func() {
		switch msg[0] {
		case msgSpaceRequest:
			alloc.debugln("Peer", sender, "asked me for space")
			r, err := decodeRange(msg[1:])
			// If we don't have a ring, just ignore a request for space.
			// They'll probably ask again later.
			if err == nil && !alloc.ring.Empty() {
				alloc.donateSpace(r, sender)
			}
			resultChan <- err
		case msgSpaceRequestDenied:
			r, err := decodeRange(msg[1:])
			if err == nil {
				alloc.spaceRequestDenied(sender, r)
			}
			resultChan <- err
		case msgRingUpdate:
			resultChan <- alloc.update(sender, msg[1:])
		}
	}
	return <-resultChan
}

// OnGossipBroadcast (Sync)
func (alloc *Allocator) OnGossipBroadcast(sender mesh.PeerName, msg []byte) (mesh.GossipData, error) {
	alloc.debugln("OnGossipBroadcast from", sender, ":", len(msg), "bytes")
	resultChan := make(chan error)
	alloc.actionChan <- func() {
		resultChan <- alloc.update(sender, msg)
	}
	return alloc.Gossip(), <-resultChan
}

type gossipState struct {
	// We send a timstamp along with the information to be
	// gossipped for backwards-compatibility (previously to detect skewed clocks)
	Now       int64
	Nicknames map[mesh.PeerName]string

	Paxos paxos.GossipState
	Ring  *ring.Ring
}

func (alloc *Allocator) encode() []byte {
	data := gossipState{
		Now:       alloc.now().Unix(),
		Nicknames: alloc.nicknames,
	}

	// We're only interested in Paxos until we have a Ring.
	// Non-electing participants (e.g. observers) return
	// a nil gossip state in order to provoke a unicast ring
	// update from electing peers who have reached consensus.
	if alloc.ring.Empty() {
		data.Paxos = alloc.paxos.GossipState()
	} else {
		data.Ring = alloc.ring
	}
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(data); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

// Encode (Sync)
func (alloc *Allocator) Encode() []byte {
	resultChan := make(chan []byte)
	alloc.actionChan <- func() {
		resultChan <- alloc.encode()
	}
	return <-resultChan
}

// OnGossip (Sync)
func (alloc *Allocator) OnGossip(msg []byte) (mesh.GossipData, error) {
	alloc.debugln("Allocator.OnGossip:", len(msg), "bytes")
	resultChan := make(chan error)
	alloc.actionChan <- func() {
		resultChan <- alloc.update(mesh.UnknownPeerName, msg)
	}
	return nil, <-resultChan // for now, we never propagate updates. TBD
}

// GossipData implementation is trivial - we always gossip the latest
// data we have at time of sending
type ipamGossipData struct {
	alloc *Allocator
}

func (d *ipamGossipData) Merge(other mesh.GossipData) mesh.GossipData {
	return d // no-op
}

func (d *ipamGossipData) Encode() [][]byte {
	return [][]byte{d.alloc.Encode()}
}

// Gossip returns a GossipData implementation, which in this case always
// returns the latest ring state (and does nothing on merge)
func (alloc *Allocator) Gossip() mesh.GossipData {
	return &ipamGossipData{alloc}
}

// SetInterfaces gives the allocator two interfaces for talking to the outside world
func (alloc *Allocator) SetInterfaces(gossip mesh.Gossip) {
	alloc.gossip = gossip
}

// ACTOR server

func (alloc *Allocator) actorLoop(actionChan <-chan func(), stopChan <-chan struct{}) {
	defer alloc.ticker.Stop()
	for {
		select {
		case action := <-actionChan:
			action()
		case <-stopChan:
			return
		case <-alloc.ticker.C:
			// Retry things in case messages got lost between here and recipients
			if alloc.awaitingConsensus {
				alloc.propose()
			} else if alloc.havePendingOps() {
				if alloc.ring.Empty() {
					alloc.establishRing()
				} else {
					alloc.tryPendingOps()
				}
			}
			alloc.removeDeadContainers()
		}

		alloc.assertInvariants()
		alloc.reportFreeSpace()
	}
}

// Helper functions

// Ensure we are making progress towards an established ring
func (alloc *Allocator) establishRing() {
	if !alloc.ring.Empty() || alloc.awaitingConsensus {
		return
	}

	alloc.awaitingConsensus = true
	alloc.paxos.SetQuorum(alloc.quorum())
	alloc.propose()
	if ok, cons := alloc.paxos.Consensus(); ok {
		// If the quorum was 1, then proposing immediately
		// leads to consensus
		alloc.createRing(cons.Value)
	}
}

func (alloc *Allocator) createRing(peers []mesh.PeerName) {
	alloc.debugln("Paxos consensus:", peers)
	alloc.ring.ClaimForPeers(normalizeConsensus(peers))
	alloc.ringUpdated()
	alloc.gossip.GossipBroadcast(alloc.Gossip())
}

func (alloc *Allocator) ringUpdated() {
	// When we have a ring, we don't need paxos any more
	if alloc.awaitingConsensus {
		alloc.awaitingConsensus = false
		alloc.paxos = nil
	}

	alloc.persistRing()
	alloc.space.UpdateRanges(alloc.ring.OwnedRanges())
	alloc.tryPendingOps()
}

// For compatibility with sort.Interface
type peerNames []mesh.PeerName

func (a peerNames) Len() int           { return len(a) }
func (a peerNames) Less(i, j int) bool { return a[i] < a[j] }
func (a peerNames) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

// When we get a consensus from Paxos, the peer names are not in a
// defined order and may contain duplicates.  This function sorts them
// and de-dupes.
func normalizeConsensus(consensus []mesh.PeerName) []mesh.PeerName {
	if len(consensus) == 0 {
		return nil
	}

	peers := make(peerNames, len(consensus))
	copy(peers, consensus)
	sort.Sort(peers)

	dst := 0
	for src := 1; src < len(peers); src++ {
		if peers[dst] != peers[src] {
			dst++
			peers[dst] = peers[src]
		}
	}

	return peers[:dst+1]
}

func (alloc *Allocator) propose() {
	alloc.debugf("Paxos proposing")
	alloc.paxos.Propose()
	alloc.gossip.GossipBroadcast(alloc.Gossip())
}

func encodeRange(r address.Range) []byte {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(r); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func (alloc *Allocator) sendSpaceRequest(dest mesh.PeerName, r address.Range) error {
	msg := append([]byte{msgSpaceRequest}, encodeRange(r)...)
	return alloc.gossip.GossipUnicast(dest, msg)
}

func (alloc *Allocator) sendSpaceRequestDenied(dest mesh.PeerName, r address.Range) error {
	msg := append([]byte{msgSpaceRequestDenied}, encodeRange(r)...)
	return alloc.gossip.GossipUnicast(dest, msg)
}

func (alloc *Allocator) sendRingUpdate(dest mesh.PeerName) {
	msg := append([]byte{msgRingUpdate}, alloc.encode()...)
	alloc.gossip.GossipUnicast(dest, msg)
}

func (alloc *Allocator) update(sender mesh.PeerName, msg []byte) error {
	reader := bytes.NewReader(msg)
	decoder := gob.NewDecoder(reader)
	var data gossipState

	if err := decoder.Decode(&data); err != nil {
		return err
	}

	// Merge nicknames
	for peer, nickname := range data.Nicknames {
		alloc.nicknames[peer] = nickname
	}

	switch {
	// If someone sent us a ring, merge it into ours. Note this will move us
	// out of the awaiting-consensus state if we didn't have a ring already.
	case data.Ring != nil:
		updated, err := alloc.ring.Merge(*data.Ring)
		switch err {
		case nil:
			if updated {
				alloc.pruneNicknames()
				alloc.ringUpdated()
			}
		case ring.ErrDifferentSeeds:
			return fmt.Errorf("IP allocation was seeded by different peers (received: %v, ours: %v)",
				alloc.annotatePeernames(data.Ring.Seeds), alloc.annotatePeernames(alloc.ring.Seeds))
		case ring.ErrDifferentRange:
			return fmt.Errorf("Incompatible IP allocation ranges (received: %s, ours: %s)",
				data.Ring.Range().AsCIDRString(), alloc.ring.Range().AsCIDRString())
		default:
			return err
		}

	// If we reach this point we know the sender is either an elector
	// broadcasting a paxos proposal to form a ring or a non-elector
	// broadcasting a ring request. If we have a ring already we can just send
	// it back regardless.
	case !alloc.ring.Empty():
		if sender != mesh.UnknownPeerName {
			alloc.sendRingUpdate(sender)
		}

	// Otherwise, we need to react according to whether or not we received a
	// paxos proposal.
	case data.Paxos != nil:
		// Process the proposal (this is a no-op if we're an observer)
		if alloc.paxos.Update(data.Paxos) {
			if alloc.paxos.Think() {
				// If something important changed, broadcast
				alloc.gossip.GossipBroadcast(alloc.Gossip())
			}

			if ok, cons := alloc.paxos.Consensus(); ok {
				alloc.createRing(cons.Value)
			}
		}

	// No paxos proposal present, so sender is a non-elector. We don't have a
	// ring to send, so attempt to establish one on their behalf. NB we only do
	// this:
	//
	// * On an explicit broadcast request triggered by a remote allocation attempt
	//   (if we did so on periodic gossip we would force consensus unnecessarily)
	// * If we are an elector (to avoid a broadcast storm of ring request messages)
	default:
		if alloc.paxos.IsElector() && sender != mesh.UnknownPeerName {
			alloc.establishRing()
		}
	}

	return nil
}

func (alloc *Allocator) donateSpace(r address.Range, to mesh.PeerName) {
	// No matter what we do, we'll send a unicast gossip
	// of our ring back to the chap who asked for space.
	// This serves to both tell him of any space we might
	// have given him, or tell him where he might find some
	// more.
	defer alloc.sendRingUpdate(to)

	chunk, ok := alloc.space.Donate(r)
	if !ok {
		free := alloc.space.NumFreeAddressesInRange(r)
		common.Assert(free == 0)
		alloc.debugln("No space to give to peer", to)
		// separate message maintains backwards-compatibility:
		// down-level peers will ignore this and still get the ring update.
		alloc.sendSpaceRequestDenied(to, r)
		return
	}
	alloc.debugln("Giving range", chunk, "to", to)
	alloc.ring.GrantRangeToHost(chunk.Start, chunk.End, to)
	alloc.persistRing()
}

func (alloc *Allocator) assertInvariants() {
	// We need to ensure all ranges the ring thinks we own have
	// a corresponding space in the space set, and vice versa
	checkSpace := space.New()
	checkSpace.AddRanges(alloc.ring.OwnedRanges())
	ranges := checkSpace.OwnedRanges()
	spaces := alloc.space.OwnedRanges()

	common.Assert(len(ranges) == len(spaces))

	for i := 0; i < len(ranges); i++ {
		r := ranges[i]
		s := spaces[i]
		common.Assert(s.Start == r.Start && s.End == r.End)
	}
}

func (alloc *Allocator) reportFreeSpace() {
	ranges := alloc.ring.OwnedRanges()
	if len(ranges) == 0 {
		return
	}

	freespace := make(map[address.Address]address.Count)
	for _, r := range ranges {
		freespace[r.Start] = alloc.space.NumFreeAddressesInRange(r)
	}
	if alloc.ring.ReportFree(freespace) {
		alloc.persistRing()
	}
}

// Persistent data
const (
	ringIdent  = "ring"
	ownedIdent = "ownedAddresses"
)

func (alloc *Allocator) persistRing() {
	// It would be better if these two Save operations happened in the same transaction
	if err := alloc.db.Save(db.NameIdent, alloc.ourName); err != nil {
		alloc.fatalf("Error persisting ring data: %s", err)
		return
	}
	if err := alloc.db.Save(ringIdent, alloc.ring); err != nil {
		alloc.fatalf("Error persisting ring data: %s", err)
	}
}

// Returns true if persisted data is to be used, otherwise false
func (alloc *Allocator) loadPersistedData() bool {
	var checkPeerName mesh.PeerName
	nameFound, err := alloc.db.Load(db.NameIdent, &checkPeerName)
	if err != nil {
		alloc.fatalf("Error loading persisted peer name: %s", err)
	}
	var persistedRing *ring.Ring
	ringFound, err := alloc.db.Load(ringIdent, &persistedRing)
	if err != nil {
		alloc.fatalf("Error loading persisted IPAM data: %s", err)
	}
	var persistedOwned map[string]ownedData
	ownedFound, err := alloc.db.Load(ownedIdent, &persistedOwned)
	if err != nil {
		alloc.fatalf("Error loading persisted address data: %s", err)
	}

	overwritePersisted := func(fmt string, args ...interface{}) {
		alloc.infof(fmt, args...)
		alloc.persistRing()
		alloc.persistOwned()
	}

	if !nameFound || !ringFound {
		overwritePersisted("No valid persisted data")
		return false
	}

	if checkPeerName != alloc.ourName {
		overwritePersisted("Deleting persisted data for peername %s", checkPeerName)
		return false
	}

	if persistedRing.Range() != alloc.universe.Range() {
		overwritePersisted("Deleting persisted data for IPAM range %s; our range is %s", persistedRing.Range(), alloc.universe)
		return false
	}

	alloc.ring.Restore(persistedRing)
	alloc.space.UpdateRanges(alloc.ring.OwnedRanges())

	if ownedFound {
		alloc.owned = persistedOwned
		for _, d := range alloc.owned {
			for _, cidr := range d.Cidrs {
				alloc.space.Claim(cidr.Addr)
			}
		}
	}
	return true
}

func (alloc *Allocator) persistOwned() {
	if err := alloc.db.Save(ownedIdent, alloc.owned); err != nil {
		alloc.fatalf("Error persisting address data: %s", err)
	}
}

// Owned addresses

func (alloc *Allocator) hasOwnedByContainer(ident string) bool {
	d, b := alloc.owned[ident]
	return b && d.IsContainer
}

// NB: addr must not be owned by ident already
func (alloc *Allocator) addOwned(ident string, cidr address.CIDR, isContainer bool) {
	d := alloc.owned[ident]
	d.IsContainer = isContainer
	d.Cidrs = append(d.Cidrs, cidr)
	alloc.owned[ident] = d
	alloc.persistOwned()
}

func (alloc *Allocator) removeAllOwned(ident string) []address.CIDR {
	a := alloc.owned[ident]
	delete(alloc.owned, ident)
	alloc.persistOwned()
	return a.Cidrs
}

func (alloc *Allocator) removeOwned(ident string, addrToFree address.Address) bool {
	d := alloc.owned[ident]
	for i, ownedCidr := range d.Cidrs {
		if ownedCidr.Addr == addrToFree {
			if len(d.Cidrs) == 1 {
				delete(alloc.owned, ident)
			} else {
				d.Cidrs = append(d.Cidrs[:i], d.Cidrs[i+1:]...)
				alloc.owned[ident] = d
			}
			alloc.persistOwned()
			return true
		}
	}
	return false
}

func (alloc *Allocator) ownedInRange(ident string, r address.Range) []address.CIDR {
	var c []address.CIDR
	for _, cidr := range alloc.owned[ident].Cidrs {
		if r.Contains(cidr.Addr) {
			c = append(c, cidr)
		}
	}
	return c
}

func (alloc *Allocator) findOwner(addr address.Address) string {
	for ident, d := range alloc.owned {
		for _, candidate := range d.Cidrs {
			if candidate.Addr == addr {
				return ident
			}
		}
	}
	return ""
}

// For each ID in the 'owned' map, remove the entry if it isn't in the map
func (alloc *Allocator) pruneOwned(ids map[string]struct{}) {
	changed := false
	for ident, d := range alloc.owned {
		if !d.IsContainer {
			continue
		}
		if _, found := ids[ident]; !found {
			for _, cidr := range d.Cidrs {
				alloc.space.Free(cidr.Addr)
			}
			alloc.debugf("Deleting old entry %s: %v", ident, d.Cidrs)
			delete(alloc.owned, ident)
			changed = true
		}
	}
	if changed {
		alloc.persistOwned()
	}
}

// Logging

func (alloc *Allocator) fatalf(fmt string, args ...interface{}) {
	alloc.logf(common.Log.Fatalf, fmt, args...)
}
func (alloc *Allocator) warnf(fmt string, args ...interface{}) {
	alloc.logf(common.Log.Warnf, fmt, args...)
}
func (alloc *Allocator) errorf(fmt string, args ...interface{}) {
	common.Log.Errorf("[allocator %s] "+fmt, append([]interface{}{alloc.ourName}, args...)...)
}
func (alloc *Allocator) infof(fmt string, args ...interface{}) {
	alloc.logf(common.Log.Infof, fmt, args...)
}
func (alloc *Allocator) debugf(fmt string, args ...interface{}) {
	alloc.logf(common.Log.Debugf, fmt, args...)
}
func (alloc *Allocator) logf(f func(string, ...interface{}), fmt string, args ...interface{}) {
	f("[allocator %s] "+fmt, append([]interface{}{alloc.ourName}, args...)...)
}
func (alloc *Allocator) debugln(args ...interface{}) {
	common.Log.Debugln(append([]interface{}{fmt.Sprintf("[allocator %s]:", alloc.ourName)}, args...)...)
}
