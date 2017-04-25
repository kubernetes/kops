package paxos

import (
	"github.com/weaveworks/mesh"
)

// The node identifier.  The use of the UID here is important: Paxos
// acceptors must not forget their promises, so it's important that a
// node does not restart and lose its Paxos state but claim to have
// the same ID.
type NodeID struct {
	Name mesh.PeerName
	UID  mesh.PeerUID
}

// note all fields exported in structs so we can Gob them
type ProposalID struct {
	// round numbers begin at 1.  round 0 indicates an
	// uninitialized ProposalID, and precedes all other ProposalIDs
	Round    uint
	Proposer NodeID
}

func (a ProposalID) precedes(b ProposalID) bool {
	switch {
	case a.Round != b.Round:
		return a.Round < b.Round
	case a.Proposer.Name != b.Proposer.Name:
		return a.Proposer.Name < b.Proposer.Name
	case a.Proposer.UID != b.Proposer.UID:
		return a.Proposer.UID < b.Proposer.UID
	default:
		return false
	}
}

func (a ProposalID) valid() bool {
	return a.Round > 0
}

// For seeding IPAM, the value we want consensus on is a set of peer names
type Value []mesh.PeerName

// An AcceptedValue is a Value plus the proposal which originated that
// Value.  The origin is not essential, but makes comparing
// AcceptedValues easy even if comparing Values is not.
type AcceptedValue struct {
	Value  Value
	Origin ProposalID
}

type NodeClaims struct {
	// The node promises not to accept a proposal with id less
	// than this.
	Promise ProposalID

	// The accepted proposal, if valid
	Accepted    ProposalID
	AcceptedVal AcceptedValue
}

func (a NodeClaims) equals(b NodeClaims) bool {
	return a.Promise == b.Promise && a.Accepted == b.Accepted &&
		a.AcceptedVal.Origin == b.AcceptedVal.Origin
}

type GossipState map[NodeID]NodeClaims

type Node struct {
	id     NodeID
	quorum uint
	knows  GossipState
}

func NewNode(name mesh.PeerName, uid mesh.PeerUID, quorum uint) *Node {
	return &Node{
		id:     NodeID{name, uid},
		quorum: quorum,
		knows:  map[NodeID]NodeClaims{},
	}
}

func (node *Node) SetQuorum(quorum uint) {
	node.quorum = quorum
}

func (node *Node) GossipState() GossipState {
	return node.knows
}

// Update this node's information about what other nodes know.
// Returns true if we learned something new.
func (node *Node) Update(from GossipState) bool {
	changed := false

	for i, fromClaims := range from {
		claims, ok := node.knows[i]
		if ok {
			if claims.Promise.precedes(fromClaims.Promise) {
				claims.Promise = fromClaims.Promise
				changed = true
			}

			if claims.Accepted.precedes(fromClaims.Accepted) {
				claims.Accepted = fromClaims.Accepted
				claims.AcceptedVal = fromClaims.AcceptedVal
				changed = true
			}
		} else {
			claims = fromClaims
			changed = true
		}

		node.knows[i] = claims
	}

	return changed
}

func max(a uint, b uint) uint {
	if a > b {
		return a
	}
	return b
}

// Initiate a new proposal, i.e. the Paxos "Prepare" step.  This is
// simply a matter of gossipping a new proposal that supersedes all
// others.
func (node *Node) Propose() {
	if node.quorum == 0 {
		panic("Paxos node.Propose() called with no quorum set")
	}
	// Find the highest round number around
	round := uint(0)

	for _, claims := range node.knows {
		round = max(round, claims.Promise.Round)
		round = max(round, claims.Accepted.Round)
	}

	ourClaims := node.knows[node.id]
	ourClaims.Promise = ProposalID{
		Round:    round + 1,
		Proposer: node.id,
	}
	node.knows[node.id] = ourClaims

	// With a quorum of 1, we can immediately accept our proposal
	if node.quorum == 1 {
		node.Think()
	}
}

// The heart of the consensus algorithm. Return true if we have
// changed our claims.
func (node *Node) Think() bool {
	ourClaims := node.knows[node.id]

	// The "Promise" step of Paxos: Copy the highest known
	// promise.
	for _, claims := range node.knows {
		if ourClaims.Promise.precedes(claims.Promise) {
			ourClaims.Promise = claims.Promise
		}
	}

	// The "Accept Request" step of Paxos: Acting as a proposer,
	// do we have a proposal that has been promised by a quorum?
	//
	// In Paxos, the "proposer" and "acceptor" roles are distinct,
	// so in principle a node acting as a proposer could continue
	// trying to get its proposal acccepted even after the same
	// node as an acceptor has superseded that proposal.  But
	// that's pointless in a gossip context: If our promise
	// supersedes our own proposal, then anyone who hears about
	// that promise will not accept that proposal.  So our
	// proposal is only in the running if it is also our promise.
	if ourClaims.Promise.Proposer == node.id {
		// Determine whether a quorum has promised, and the
		// best previously-accepted value if there is one.
		count := uint(0)
		var accepted ProposalID
		var acceptedVal AcceptedValue

		for _, claims := range node.knows {
			if claims.Promise == ourClaims.Promise {
				count++

				if accepted.precedes(claims.Accepted) {
					accepted = claims.Accepted
					acceptedVal = claims.AcceptedVal
				}
			}
		}

		if count >= node.quorum {
			if !accepted.valid() {
				acceptedVal.Value = node.pickValue()
				acceptedVal.Origin = ourClaims.Promise
			}

			// We automatically accept our own proposal,
			// and that's how we communicate the "accept
			// request" to other nodes.
			ourClaims.Accepted = ourClaims.Promise
			ourClaims.AcceptedVal = acceptedVal
		}
	}

	// The "Accepted" step of Paxos: If the proposal we promised
	// on got accepted by some other node, we accept it too.
	for _, claims := range node.knows {
		if claims.Accepted == ourClaims.Promise {
			ourClaims.Accepted = claims.Accepted
			ourClaims.AcceptedVal = claims.AcceptedVal
			break
		}
	}

	if ourClaims.equals(node.knows[node.id]) {
		return false
	}

	node.knows[node.id] = ourClaims
	return true
}

// When we get to pick a value, we use the set of peer names we know
// about.  This is not necessarily all peer names, but it is at least
// a quorum, and so good enough for seeding the ring.
func (node *Node) pickValue() Value {
	val := make([]mesh.PeerName, len(node.knows))
	i := 0
	for id := range node.knows {
		val[i] = id.Name
		i++
	}
	return val
}

// Has a consensus been reached, based on the known claims of other nodes?
func (node *Node) Consensus() (bool, AcceptedValue) {
	if node.quorum == 0 {
		return false, AcceptedValue{}
	}
	counts := map[ProposalID]uint{}

	for _, claims := range node.knows {
		if claims.Accepted.valid() {
			origin := claims.AcceptedVal.Origin
			count := counts[origin] + 1
			counts[origin] = count
			if count >= node.quorum {
				return true, claims.AcceptedVal
			}
		}
	}

	return false, AcceptedValue{}
}

func (node *Node) IsElector() bool {
	return true
}

type Status struct {
	Elector    bool
	KnownNodes int
	Quorum     uint
}

func NewStatus(node *Node) *Status {
	return &Status{true, len(node.knows), node.quorum}
}
