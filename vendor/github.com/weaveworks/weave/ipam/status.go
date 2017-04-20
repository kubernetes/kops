package ipam

import (
	"fmt"
	"github.com/weaveworks/weave/ipam/paxos"
	"github.com/weaveworks/weave/net/address"
)

type Status struct {
	Paxos            *paxos.Status
	Range            string
	RangeNumIPs      int
	ActiveIPs        int
	DefaultSubnet    string
	Entries          []EntryStatus
	PendingClaims    []ClaimStatus
	PendingAllocates []string
}

type EntryStatus struct {
	Token       string
	Size        uint32
	Peer        string
	Nickname    string
	IsKnownPeer bool
	Version     uint32
}

type ClaimStatus struct {
	Ident string
	CIDR  address.CIDR
}

func NewStatus(allocator *Allocator, defaultSubnet address.CIDR) *Status {
	if allocator == nil {
		return nil
	}

	var paxosStatus *paxos.Status
	if allocator.awaitingConsensus && allocator.paxos != nil {
		if allocator.paxos.IsElector() {
			if node, ok := allocator.paxos.(*paxos.Node); ok {
				paxosStatus = paxos.NewStatus(node)
			}
		} else {
			paxosStatus = &paxos.Status{Elector: false}
		}
	}

	resultChan := make(chan *Status)
	allocator.actionChan <- func() {
		resultChan <- &Status{
			paxosStatus,
			allocator.universe.String(),
			int(allocator.universe.Size()),
			int(allocator.space.NumOwnedAddresses()),
			defaultSubnet.String(),
			newEntryStatusSlice(allocator),
			newClaimStatusSlice(allocator),
			newAllocateIdentSlice(allocator)}
	}

	return <-resultChan
}

func newEntryStatusSlice(allocator *Allocator) []EntryStatus {
	var slice []EntryStatus

	if allocator.ring.Empty() {
		return slice
	}

	for _, r := range allocator.ring.AllRangeInfo() {
		slice = append(slice, EntryStatus{
			Token:       r.Start.String(),
			Size:        uint32(r.Size()),
			Peer:        r.Peer.String(),
			Nickname:    allocator.nicknames[r.Peer],
			IsKnownPeer: allocator.isKnownPeer(r.Peer),
			Version:     r.Version,
		})
	}

	return slice
}

func newClaimStatusSlice(allocator *Allocator) []ClaimStatus {
	var slice []ClaimStatus
	for _, op := range allocator.pendingClaims {
		claim := op.(*claim)
		slice = append(slice, ClaimStatus{claim.ident, claim.cidr})
	}
	return slice
}

func newAllocateIdentSlice(allocator *Allocator) []string {
	var slice []string
	for _, op := range allocator.pendingAllocates {
		allocate := op.(*allocate)
		slice = append(slice, fmt.Sprintf("%s %s", allocate.ident, allocate.r.String()))
	}
	return slice
}
