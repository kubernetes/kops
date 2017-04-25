package tracker

import (
	"github.com/weaveworks/weave/net/address"
)

// LocalRangeTracker is an interface for tracking changes in the IPAM ring.
type LocalRangeTracker interface {
	// HandleUpdate is called whenever an address ring gets updated.
	//
	// prevRanges corresponds to ranges which were owned by a peer before
	// a change in the ring, while currRanges to the ones which are currently
	// owned by the peer.
	// Both slices have to be sorted in increasing order.
	// Adjacent ranges within each slice might appear as separate ranges.
	//
	// The local parameter indicates whether the ranges belong to the peer
	// by which the method is called.
	HandleUpdate(prevRanges, currRanges []address.Range, local bool) error
}
