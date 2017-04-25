package router

import (
	"github.com/weaveworks/mesh"
)

// Interface to overlay network packet handling
type NetworkOverlay interface {
	mesh.Overlay

	// The routes have changed, so any cached information should
	// be discarded.
	InvalidateRoutes()

	// A mapping of a short id to a peer has changed
	InvalidateShortIDs()

	// Start consuming forwarded packets.
	StartConsumingPackets(*mesh.Peer, *mesh.Peers, OverlayConsumer) error
}

// When a consumer is called, the decoder will already have been used
// to decode the frame.
type OverlayConsumer func(ForwardPacketKey) FlowOp

// All of the machinery to forward packets to a particular peer
type OverlayForwarder interface {
	mesh.OverlayConnection
	// Forward a packet across the connection.  May be called as soon
	// as the overlay connection is created, in particular before
	// Confirm().  The return value nil means the key could not be
	// handled by this forwarder.
	Forward(ForwardPacketKey) FlowOp
}

type NullNetworkOverlay struct{ mesh.NullOverlay }

func (NullNetworkOverlay) InvalidateRoutes() {
}

func (NullNetworkOverlay) InvalidateShortIDs() {
}

func (NullNetworkOverlay) StartConsumingPackets(*mesh.Peer, *mesh.Peers, OverlayConsumer) error {
	return nil
}

func (NullNetworkOverlay) Forward(ForwardPacketKey) FlowOp {
	return DiscardingFlowOp{}
}
