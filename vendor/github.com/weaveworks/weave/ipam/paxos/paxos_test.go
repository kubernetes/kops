package paxos

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/weaveworks/mesh"
)

type TestNode struct {
	*Node

	// Topology
	links    []*Link
	isolated bool

	// The first consensus the Node observed
	firstConsensus AcceptedValue
}

type Link struct {
	from     *TestNode
	to       *TestNode
	converse *Link

	// A link is considered ready if it is worthwhile gossipping
	// along it (i.e. unless we already gossipped along it and the
	// "from" node didn't change since then).  If this is <0, then
	// the link is not ready.  Otherwise, this gives the index in
	// the Model's readyLinks slice.
	ready int
}

type Model struct {
	t          *testing.T
	r          *rand.Rand
	quorum     uint
	nodes      []TestNode
	readyLinks []*Link
	nextID     uint
}

func (m *Model) addLink(a, b *TestNode) {
	if a == b {
		// Never link a node to itself
		return
	}

	ab := Link{from: a, to: b, ready: len(m.readyLinks)}
	ba := Link{from: b, to: a, converse: &ab, ready: len(m.readyLinks) + 1}
	ab.converse = &ba
	a.links = append(a.links, &ab)
	b.links = append(b.links, &ba)
	m.readyLinks = append(m.readyLinks, &ab, &ba)
}

func (m *Model) linkExists(a, b *TestNode) bool {
	for _, l := range a.links {
		if l.to == b {
			return true
		}
	}

	return false
}

type TestParams struct {
	// Number of nodes
	nodeCount uint

	// Probability that two nodes are connected.
	connectedProb float32

	// Probability that some node will re-propose at each
	// step. Setting this too high makes it likely that we'll fail
	// to converge.
	reproposeProb float32

	// Probability that some node will be isolated at each step.
	isolateProb float32

	// Probability that some node willbe restarted at each step.
	restartProb float32
}

// Make a network of nodes with random topology
func makeRandomModel(params *TestParams, r *rand.Rand, t *testing.T) *Model {
	m := Model{
		t:          t,
		r:          r,
		quorum:     params.nodeCount/2 + 1,
		nodes:      make([]TestNode, params.nodeCount),
		readyLinks: []*Link{},
		nextID:     params.nodeCount + 1,
	}

	for i := range m.nodes {
		m.nodes[i].Node = NewNode(mesh.PeerName(i/2+1),
			mesh.PeerUID(r.Int63()), m.quorum)
		m.nodes[i].Propose()
	}

	for i := 1; i < len(m.nodes); i++ {
		// was node i connected to the other nodes yet?
		connected := false

		for j := 0; j < i; j++ {
			if r.Float32() < params.connectedProb {
				connected = true
				m.addLink(&m.nodes[i], &m.nodes[j])
			}
		}

		if !connected {
			// node i must be connected into the graph
			// somewhere.  So if we didn't connect it
			// already, this is a last resort.
			m.addLink(&m.nodes[i], &m.nodes[r.Intn(i)])
		}
	}

	return &m
}

// Mark a link as unready
func (m *Model) unreadyLink(link *Link) {
	i := link.ready
	if i >= 0 {
		m.readyLinks[i] = m.readyLinks[len(m.readyLinks)-1]
		m.readyLinks[i].ready = i
		m.readyLinks = m.readyLinks[:len(m.readyLinks)-1]
		link.ready = -1
	}
}

// Mark a link as ready
func (m *Model) readyLink(link *Link) {
	if link.ready < 0 {
		link.ready = len(m.readyLinks)
		m.readyLinks = append(m.readyLinks, link)
	}
}

// Mark all the outgoing links from a node as ready
func (m *Model) nodeChanged(node *TestNode) {
	for _, l := range node.links {
		m.readyLink(l)
	}
}

// Isolate a node
func (m *Model) isolateNode(node *TestNode) {
	node.isolated = true
	for _, l := range node.links {
		if l.ready >= 0 {
			m.unreadyLink(l)
			m.unreadyLink(l.converse)
		}
	}

	// Isolating a node could partition the network.  We don't
	// want to test such a case (because it could prevent
	// consensus).  Checking for partitions would be more code
	// than its worth, so just add some links to prevent the
	// possibility of partitions.
	for i := 1; i < len(node.links); i++ {
		m.addLink(node.links[i-1].to, node.links[i].to)
	}
}

// Restart a node
func (m *Model) restart(node *TestNode) {
	node.Node = NewNode(mesh.PeerName(m.nextID),
		mesh.PeerUID(m.r.Int63()), m.quorum)
	m.nextID++
	node.Propose()

	// The node is now ignorant, so we need to mark the links into
	// the node as ready.
	for _, l := range node.links {
		if !l.to.isolated {
			m.readyLink(l.converse)
		}
	}

	// If a consensus was just accepted due to this node accepting
	// it, without other nodes hearing of it, and we then restart
	// this node, then a different consensus can occur later on.
	// If a tree falls with no one to hear it, does it make a
	// sound?
	node.firstConsensus = AcceptedValue{}
}

func (m *Model) pickNode() *TestNode {
	for {
		node := &m.nodes[m.r.Intn(len(m.nodes))]
		if !node.isolated {
			return node
		}
	}
}

func (m *Model) simulate(params *TestParams) bool {
	nodesLeft := uint(len(m.nodes))
	restarts := uint(0)

	for step := 0; step < 1000000; step++ {
		if len(m.readyLinks) == 0 {
			// Everything has converged.  This can be
			// because consensus was reached, or because
			// consensus became impossible, e.g. everyone
			// promised on a particular proposal, but then
			// the proposer restarted or was isolated.  So
			// we detect the latter cases and force a new
			// proposal
			for i := range m.nodes {
				ok, _ := m.nodes[i].Consensus()
				if ok {
					return true
				}
			}

			node := m.pickNode()
			node.Propose()
			m.nodeChanged(node)
		}

		// Pick a ready link at random
		i := m.r.Intn(len(m.readyLinks))
		link := m.readyLinks[i]
		if link.ready != i {
			m.t.Fatal("Link in readyLinks was not ready")
		}

		// gossip across link
		node := link.to
		if node.Update(link.from.GossipState()) {
			node.Think()

			if !node.firstConsensus.Origin.valid() {
				ok, val := node.Consensus()
				if ok {
					node.firstConsensus = val
				}
			}

			m.nodeChanged(node)
		}

		m.unreadyLink(link)

		// Re-propose?
		if m.r.Float32() < params.reproposeProb {
			node := m.pickNode()
			node.Propose()
			m.nodeChanged(node)
		}

		// Isolate?
		if nodesLeft > m.quorum && m.r.Float32() < params.isolateProb {
			m.isolateNode(m.pickNode())
			nodesLeft--

			// We isolated a node, so get another node to
			// re-propose.  In reality the lack of
			// consensus would be detected via a timeout
			node := m.pickNode()
			node.Propose()
			m.nodeChanged(node)
		}

		// Restart?
		if restarts < m.quorum && m.r.Float32() < params.restartProb {
			restarts++
			node := m.pickNode()
			m.restart(node)
			m.nodeChanged(node)
		}
	}

	return false
}

func (m *Model) dump() {
	for i := range m.nodes {
		node := &m.nodes[i]
		fmt.Println(node.id)
		for n, claims := range node.knows {
			fmt.Printf("\t%d %v\n", n, claims)
		}
	}
}

// Validate the final model state
func (m *Model) validate() {
	var origin ProposalID

	for i := range m.nodes {
		ok, val := m.nodes[i].Consensus()
		if !ok {
			m.dump()
			m.t.Fatal("Node doesn't know about consensus")
		}

		firstConsensus := m.nodes[i].firstConsensus
		if firstConsensus.Origin.valid() &&
			firstConsensus.Origin != val.Origin {
			m.dump()
			m.t.Fatal("Consensus mismatch")
		}

		if i == 0 {
			origin = val.Origin
		} else if val.Origin != origin {
			m.t.Fatal("Node disagrees about consensus")
		}
	}
}

func TestSingleNode(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Test the single node case
	params := &TestParams{
		nodeCount:     1,
		connectedProb: 0,
		reproposeProb: 0,
		isolateProb:   0,
		restartProb:   0,
	}

	m := makeRandomModel(params, r, t)

	if !m.simulate(params) {
		m.t.Fatal("Failed to converge")
	}

	m.validate()
}

func TestPaxos(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	params := &TestParams{
		nodeCount:     10,
		connectedProb: 0.5,
		reproposeProb: 0.01,
		isolateProb:   0.01,

		// Restarts cause failures due to the case where 1)
		// the gossip ordering is such that there is an
		// effective partition from the start, with a quorum
		// on one side and one node less than a quorum on the
		// other; 2) the quorum reaches a consensus; 3) a node
		// is restarted, and at the same time jumps the
		// partition, so the side that was now just less than
		// a quorum becomes a quorum; 4) the new quorum
		// reaches a contradictory consensus.
		restartProb: 0,
	}

	for i := 0; i < 1000; i++ {
		m := makeRandomModel(params, r, t)

		if !m.simulate(params) {
			m.t.Fatal("Failed to converge")
		}

		m.validate()
	}
}
