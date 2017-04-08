package mesh

import "testing"

func newPeerFrom(peer *Peer) *Peer {
	return newPeerFromSummary(peer.peerSummary)
}

func TestPeerRoutes(t *testing.T) {
	t.Skip("TODO")
}

func TestPeerForEachConnectedPeer(t *testing.T) {
	t.Skip("TODO")
}
