/*
Package ring implements a simple ring CRDT.
*/
package ring

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"sort"
	"time"

	"github.com/weaveworks/mesh"

	"github.com/weaveworks/weave/common"
	"github.com/weaveworks/weave/net/address"
)

type OnUpdate func(prev, curr []address.Range, local bool)

// Ring represents the ring itself
type Ring struct {
	Start, End address.Address // [min, max) tokens in this ring.  Due to wrapping, min == max (effectively)
	Peer       mesh.PeerName   // name of peer owning this ring instance
	Entries    entries         // list of entries sorted by token
	Seeds      []mesh.PeerName // peers with which the ring was seeded
	onUpdate   OnUpdate
}

func (r *Ring) assertInvariants() {
	err := r.checkInvariants()
	if err != nil {
		panic(err.Error())
	}
}

// Errors returned by Merge
var (
	ErrNotSorted       = errors.New("Ring not sorted")
	ErrTokenRepeated   = errors.New("Token appears twice in ring")
	ErrTokenOutOfRange = errors.New("Token is out of range")
	ErrDifferentSeeds  = errors.New("Received ring was seeded differently from ours")
	ErrDifferentRange  = errors.New("Received range differs from ours")
)

func errInconsistentEntry(mine, theirs *entry) error {
	return fmt.Errorf("Inconsistent entries for %s: owned by %s but incoming message says %s", mine.Token, mine.Peer, theirs.Peer)
}
func errEntryInMyRange(theirs *entry) error {
	return fmt.Errorf("Peer %s says it owns the IP range from %s, which I think I own", theirs.Peer, theirs.Token)
}
func errNewerVersion(mine, theirs *entry) error {
	return fmt.Errorf("Received update for IP range I own at %s v%d: incoming message says owner %s v%d", mine.Token, mine.Version, theirs.Peer, theirs.Version)
}

func (r *Ring) checkInvariants() error {
	return r.checkEntries(r.Entries)
}

func (r *Ring) checkEntries(entries entries) error {
	if !sort.IsSorted(entries) {
		return ErrNotSorted
	}

	// Check no token appears twice
	// We know it's sorted...
	for i := 1; i < len(entries); i++ {
		if entries[i-1].Token == entries[i].Token {
			return ErrTokenRepeated
		}
	}

	if len(entries) == 0 {
		return nil
	}

	// Check tokens are in range
	if entries.entry(0).Token < r.Start {
		return ErrTokenOutOfRange
	}
	if entries.entry(-1).Token >= r.End {
		return ErrTokenOutOfRange
	}

	// Check all the freespaces are in range
	for i, entry := range entries {
		next := entries.entry(i + 1)
		distance := r.distance(entry.Token, next.Token)

		if entry.Free > distance {
			return fmt.Errorf("Entry %s-%s reporting too much free space: %d > %d", entry.Token, next.Token, entry.Free, distance)
		}
	}

	return nil
}

func (r *Ring) trackUpdates() func() {
	return r.trackUpdatesOfPeer(r.Peer)
}

func (r *Ring) trackUpdatesOfPeer(peer mesh.PeerName) func() {
	if r.onUpdate == nil {
		return func() {}
	}
	ranges := r.OwnedRangesOfPeer(peer)
	return func() { r.onUpdate(ranges, r.OwnedRangesOfPeer(peer), peer == r.Peer) }
}

// New creates an empty ring belonging to peer.
func New(start, end address.Address, peer mesh.PeerName, f OnUpdate) *Ring {
	common.Assert(start < end)

	ring := &Ring{Start: start, End: end, Peer: peer, Entries: make([]*entry, 0), onUpdate: f}
	return ring
}

func (r *Ring) Restore(other *Ring) {
	defer r.trackUpdates()()

	onUpdate := r.onUpdate
	*r = *other
	r.onUpdate = onUpdate
}

func (r *Ring) Range() address.Range {
	return address.Range{Start: r.Start, End: r.End}
}

// Returns the distance between two tokens on this ring, dealing
// with ranges which cross the origin
func (r *Ring) distance(start, end address.Address) address.Count {
	if end > start {
		return address.Count(end - start)
	}

	return address.Count((r.End - start) + (end - r.Start))
}

// GrantRangeToHost modifies the ring such that range [start, end)
// is assigned to peer.  This may insert up to two new tokens.
// Preconditions:
// - start < end
// - [start, end) must be owned by the calling peer
func (r *Ring) GrantRangeToHost(start, end address.Address, peer mesh.PeerName) {
	//fmt.Printf("%s GrantRangeToHost [%v,%v) -> %s\n", r.Peer, start, end, peer)

	r.assertInvariants()
	defer r.trackUpdates()()
	defer r.assertInvariants()

	// ----------------- Start of Checks -----------------

	common.Assert(start < end)
	common.Assert(r.Start <= start && start < r.End)
	common.Assert(r.Start < end && end <= r.End)
	common.Assert(len(r.Entries) > 0)

	// Look for the left-most entry greater than start, then go one previous
	// to get the right-most entry less than or equal to start
	preceedingPos := sort.Search(len(r.Entries), func(j int) bool {
		return r.Entries[j].Token > start
	})
	preceedingPos--

	// Check all tokens up to end are owned by us
	for pos := preceedingPos; pos < len(r.Entries) && r.Entries.entry(pos).Token < end; pos++ {
		common.Assert(r.Entries.entry(pos).Peer == r.Peer)
	}

	// ----------------- End of Checks -----------------

	// Free space at start is max(length of range, distance to next token)
	startFree := r.distance(start, r.Entries.entry(preceedingPos+1).Token)
	if length := r.distance(start, end); startFree > length {
		startFree = length
	}
	// Is there already a token at start, update it
	if previousEntry := r.Entries.entry(preceedingPos); previousEntry.Token == start {
		previousEntry.update(peer, startFree)
	} else {
		// Otherwise, these isn't a token here, insert a new one.
		r.Entries.insert(entry{Token: start, Peer: peer, Free: startFree})
		preceedingPos++
		// Reset free space on previous entry, which we own.
		previousEntry.update(r.Peer, r.distance(previousEntry.Token, start))
	}

	// Give all intervening tokens to the other peer
	pos := preceedingPos + 1
	for ; pos < len(r.Entries) && r.Entries.entry(pos).Token < end; pos++ {
		entry := r.Entries.entry(pos)
		entry.update(peer, address.Min(entry.Free, r.distance(entry.Token, end)))
	}

	// There is never an entry with a token of r.End, as the end of
	// the ring is exclusive.
	if end == r.End {
		end = r.Start
	}

	//  If there is a token equal to the end of the range, we don't need to do anything further
	if _, found := r.Entries.get(end); found {
		return
	}

	// If not, we need to insert a token such that we claim this bit on the end.
	endFree := r.distance(end, r.Entries.entry(pos).Token)
	r.Entries.insert(entry{Token: end, Peer: r.Peer, Free: endFree})
}

// Merge the given ring into this ring and indicate whether this ring
// got updated as a result.
func (r *Ring) Merge(gossip Ring) (bool, error) {
	r.assertInvariants()
	defer r.trackUpdates()()

	// Don't panic when checking the gossiped in ring.
	// In this case just return any error found.
	if err := gossip.checkInvariants(); err != nil {
		return false, err
	}

	if len(gossip.Seeds) > 0 && len(r.Seeds) > 0 {
		if len(gossip.Seeds) != len(r.Seeds) {
			return false, ErrDifferentSeeds
		}
		for i, seed := range gossip.Seeds {
			if seed != r.Seeds[i] {
				return false, ErrDifferentSeeds
			}
		}
	}

	if r.Start != gossip.Start || r.End != gossip.End {
		return false, ErrDifferentRange
	}

	result, updated, err := r.Entries.merge(gossip.Entries, r.Peer)

	if err != nil {
		return false, err
	}

	if err := r.checkEntries(result); err != nil {
		return false, fmt.Errorf("Merge of incoming data causes: %s", err)
	}

	if len(r.Seeds) == 0 {
		r.Seeds = gossip.Seeds
	}
	r.Entries = result

	return updated, nil
}

// Merge other entries into ours, and complain when that stomps on
// entries belonging to ourPeer. Returns the merged entries and an
// indication whether the merge resulted in any changes, i.e. the
// result differs from the original.
func (es entries) merge(other entries, ourPeer mesh.PeerName) (result entries, updated bool, err error) {
	var mine, theirs *entry
	var previousOwner *mesh.PeerName
	addToResult := func(e entry) { result = append(result, &e) }
	addTheirs := func(e entry) {
		addToResult(e)
		updated = true
		previousOwner = nil
	}

	// i is index into es; j is index into other
	var i, j int
	for i < len(es) && j < len(other) {
		mine, theirs = es[i], other[j]
		switch {
		case mine.Token < theirs.Token:
			addToResult(*mine)
			previousOwner = &mine.Peer
			i++
		case mine.Token > theirs.Token:
			// insert, checking that a range owned by us hasn't been split
			if previousOwner != nil && *previousOwner == ourPeer && theirs.Peer != ourPeer {
				err = errEntryInMyRange(theirs)
				return
			}
			addTheirs(*theirs)
			j++
		case mine.Token == theirs.Token:
			// merge
			switch {
			case mine.Version >= theirs.Version:
				if mine.Version == theirs.Version && !mine.Equal(theirs) {
					err = errInconsistentEntry(mine, theirs)
					return
				}
				addToResult(*mine)
				previousOwner = &mine.Peer
			case mine.Version < theirs.Version:
				if mine.Peer == ourPeer { // We shouldn't receive updates to our own tokens
					err = errNewerVersion(mine, theirs)
					return
				}
				addTheirs(*theirs)
			}
			i++
			j++
		}
	}

	// At this point, either i is at the end of es or j is at the end
	// of other, so copy over the remaining entries.

	for ; i < len(es); i++ {
		mine = es[i]
		addToResult(*mine)
	}

	for ; j < len(other); j++ {
		theirs = other[j]
		if previousOwner != nil && *previousOwner == ourPeer && theirs.Peer != ourPeer {
			err = errEntryInMyRange(theirs)
			return
		}
		addTheirs(*theirs)
	}

	return
}

// Empty returns true if the ring has no entries
func (r *Ring) Empty() bool {
	return len(r.Entries) == 0
}

// Given a slice of ranges which are all in the right order except
// possibly the last one spans zero, fix that up and return the slice
func (r *Ring) splitRangesOverZero(ranges []address.Range) []address.Range {
	if len(ranges) == 0 {
		return ranges
	}
	lastRange := ranges[len(ranges)-1]
	// if end token == start (ie last) entry on ring, we want to actually use r.End
	if lastRange.End == r.Start {
		ranges[len(ranges)-1].End = r.End
	} else if lastRange.End <= lastRange.Start {
		// We wrapped; want to split around 0
		// First shuffle everything up as we want results to be sorted
		ranges = append(ranges, address.Range{})
		copy(ranges[1:], ranges[:len(ranges)-1])
		ranges[0] = address.Range{Start: r.Start, End: lastRange.End}
		ranges[len(ranges)-1].End = r.End
	}
	return ranges
}

// OwnedRanges returns slice of Ranges, ordered by IP, indicating which
// ranges are owned by this peer.  Will split ranges which
// span 0 in the ring.
func (r *Ring) OwnedRanges() (result []address.Range) {
	return r.OwnedRangesOfPeer(r.Peer)
}

func (r *Ring) OwnedRangesOfPeer(peer mesh.PeerName) (result []address.Range) {
	r.assertInvariants()

	for i, entry := range r.Entries {
		if entry.Peer == peer {
			nextEntry := r.Entries.entry(i + 1)
			result = append(result, address.Range{Start: entry.Token, End: nextEntry.Token})
		}
	}

	return r.splitRangesOverZero(result)
}

// For printing status
type RangeInfo struct {
	Peer mesh.PeerName
	address.Range
	Version uint32
}

func (r *Ring) AllRangeInfo() (result []RangeInfo) {
	for i, entry := range r.Entries {
		nextEntry := r.Entries.entry(i + 1)
		ranges := []address.Range{{Start: entry.Token, End: nextEntry.Token}}
		ranges = r.splitRangesOverZero(ranges)
		for _, r := range ranges {
			result = append(result, RangeInfo{entry.Peer, r, entry.Version})
		}
	}
	return
}

// ClaimForPeers claims the entire ring for the array of peers passed
// in.  Only works for empty rings. Each claimed range is CIDR-aligned.
func (r *Ring) ClaimForPeers(peers []mesh.PeerName) {
	common.Assert(r.Empty())

	defer r.trackUpdates()()
	defer r.assertInvariants()
	defer func() {
		e := r.Entries[len(r.Entries)-1]
		common.Assert(address.Add(e.Token, address.Offset(e.Free)) == r.End)
	}()

	r.subdivide(r.Start, r.End, peers)
	r.Seeds = peers
}

// subdivide subdivides the [from,to) CIDR for the given peers into
// CIDR-aligned subranges.
func (r *Ring) subdivide(from, to address.Address, peers []mesh.PeerName) {
	share := address.Length(to, from)
	if share == 0 {
		return
	}
	if share == 1 || len(peers) == 1 {
		r.Entries.insert(entry{Token: from, Peer: peers[0], Free: share})
		return
	}
	mid := address.Add(from, address.Offset(share/2))
	r.subdivide(from, mid, peers[:len(peers)/2])
	r.subdivide(address.Add(mid, address.Offset(share%2)), to, peers[len(peers)/2:])
}

func (r *Ring) FprintWithNicknames(w io.Writer, m map[mesh.PeerName]string) {
	for _, entry := range r.Entries {
		nickname, found := m[entry.Peer]
		if found {
			nickname = fmt.Sprintf(" (%s)", nickname)
		}

		fmt.Fprintf(w, "\n  %s -> %s%s (v%d)", entry.Token,
			entry.Peer, nickname, entry.Version)
	}
}

func (r *Ring) String() string {
	var buffer bytes.Buffer
	fmt.Fprintf(&buffer, "Ring [%s, %s)", r.Start, r.End)
	r.FprintWithNicknames(&buffer, make(map[mesh.PeerName]string))
	return buffer.String()
}

// ReportFree is used by the allocator to tell the ring how many free
// ips are in a given range, so that ChoosePeersToAskForSpace can make
// more intelligent decisions.  Returns true if any changes made.
func (r *Ring) ReportFree(freespace map[address.Address]address.Count) (updated bool) {
	r.assertInvariants()
	defer r.assertInvariants()

	common.Assert(!r.Empty())
	entries := r.Entries

	// As OwnedRanges splits around the origin, we need to
	// detect that here and fix up freespace
	if free, found := freespace[r.Start]; found && entries.entry(0).Token != r.Start {
		lastToken := entries.entry(-1).Token
		prevFree, found := freespace[lastToken]
		common.Assert(found)
		freespace[lastToken] = prevFree + free
		delete(freespace, r.Start)
	}

	for start, free := range freespace {
		// Look for entry
		i := sort.Search(len(entries), func(j int) bool {
			return entries[j].Token >= start
		})

		// Are you trying to report free on space I don't own?
		common.Assert(i < len(entries) && entries[i].Token == start && entries[i].Peer == r.Peer)

		// Check we're not reporting more space than the range
		entry, next := entries.entry(i), entries.entry(i+1)
		maxSize := r.distance(entry.Token, next.Token)
		common.Assert(free <= address.Count(maxSize))

		if entries[i].Free == free {
			continue
		}

		entries[i].Free = free
		entries[i].Version++
		updated = true
	}
	return
}

type weightedPeer struct {
	weight   float64
	peername mesh.PeerName
}
type weightedPeers []weightedPeer

// Note Less is using > so that bigger weights sort earlier
func (ws weightedPeers) Less(i, j int) bool { return ws[i].weight > ws[j].weight }
func (ws weightedPeers) Len() int           { return len(ws) }
func (ws weightedPeers) Swap(i, j int)      { ws[i], ws[j] = ws[j], ws[i] }

// ChoosePeersToAskForSpace returns all peers we can ask for space in
// the range [start, end), in weighted-random order.  Assumes start<end.
func (r *Ring) ChoosePeersToAskForSpace(start, end address.Address) []mesh.PeerName {
	totalSpacePerPeer := make(map[mesh.PeerName]address.Count)

	// iterate through tokens
	for i, entry := range r.Entries {
		// Ignore entries that don't span the range we want
		if i+1 < len(r.Entries) && r.Entries.entry(i+1).Token-1 < start {
			continue
		}
		if entry.Token >= end {
			break
		}
		// Ignore ranges with no free space
		if entry.Free <= 0 {
			continue
		}

		// Don't talk to yourself
		if entry.Peer == r.Peer {
			continue
		}

		totalSpacePerPeer[entry.Peer] += entry.Free
	}

	// Compute weighted random numbers, then sort.
	// This isn't perfect, e.g. an item with weight 2 will get chosen more than
	// twice as often as an item with weight 1, but it's good enough for our purposes.
	ws := make(weightedPeers, 0, len(totalSpacePerPeer))
	for peername, space := range totalSpacePerPeer {
		ws = append(ws, weightedPeer{weight: float64(space) * rand.Float64(), peername: peername})
	}
	sort.Sort(ws)
	result := make([]mesh.PeerName, len(ws))
	for i, wp := range ws {
		result[i] = wp.peername
	}
	return result
}

func (r *Ring) PickPeerForTransfer(isValid func(mesh.PeerName) bool) mesh.PeerName {
	for _, entry := range r.Entries {
		if entry.Peer != r.Peer && isValid(entry.Peer) {
			return entry.Peer
		}
	}
	return mesh.UnknownPeerName
}

// Transfer will mark all entries associated with 'from' peer as owned by 'to' peer
// and return ranges indicating the new space we picked up
func (r *Ring) Transfer(from, to mesh.PeerName) []address.Range {
	r.assertInvariants()
	defer r.trackUpdates()()
	defer r.trackUpdatesOfPeer(from)()
	defer r.assertInvariants()

	var newRanges []address.Range

	for i, entry := range r.Entries {
		if entry.Peer == from {
			entry.Peer = to
			entry.Version++
			newRanges = append(newRanges, address.Range{Start: entry.Token, End: r.Entries.entry(i + 1).Token})
		}
	}

	return r.splitRangesOverZero(newRanges)
}

// Contains returns true if addr is in this ring
func (r *Ring) Contains(addr address.Address) bool {
	return addr >= r.Start && addr < r.End
}

// Owner returns the peername which owns the range containing addr
func (r *Ring) Owner(token address.Address) mesh.PeerName {
	common.Assert(r.Start <= token && token < r.End)

	r.assertInvariants()
	// There can be no owners on an empty ring
	if r.Empty() {
		return mesh.UnknownPeerName
	}

	// Look for the right-most entry, less than or equal to token
	preceedingEntry := sort.Search(len(r.Entries), func(j int) bool {
		return r.Entries[j].Token > token
	})
	preceedingEntry--
	entry := r.Entries.entry(preceedingEntry)
	return entry.Peer
}

// Get the set of PeerNames mentioned in the ring
func (r *Ring) PeerNames() map[mesh.PeerName]struct{} {
	res := make(map[mesh.PeerName]struct{})

	for _, entry := range r.Entries {
		res[entry.Peer] = struct{}{}
	}

	return res
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}
