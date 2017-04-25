package ring

import (
	"bytes"
	"fmt"
	"math/rand"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/weaveworks/mesh"

	"github.com/weaveworks/weave/common"
	"github.com/weaveworks/weave/net/address"
)

var (
	peer1name, _ = mesh.PeerNameFromString("01:00:00:00:02:00")
	peer2name, _ = mesh.PeerNameFromString("02:00:00:00:02:00")
	peer3name, _ = mesh.PeerNameFromString("03:00:00:00:02:00")

	start, end    = ParseIP("10.0.0.0"), ParseIP("10.0.1.0")
	dot8          = ParseIP("10.0.0.8")
	dot10, dot245 = ParseIP("10.0.0.10"), ParseIP("10.0.0.245")
	dot250        = ParseIP("10.0.0.250")
	middle        = ParseIP("10.0.0.128")
)

func ParseIP(s string) address.Address {
	addr, _ := address.ParseIP(s)
	return addr
}

func merge(r1, r2 *Ring) error {
	_, err := r1.Merge(*r2)
	return err
}

func NewRing(start, end address.Address, peer mesh.PeerName) *Ring {
	return New(start, end, peer, nil)
}

func TestInvariants(t *testing.T) {
	ring := NewRing(start, end, peer1name)

	// Check ring is sorted
	ring.Entries = []*entry{{Token: dot245, Peer: peer1name}, {Token: dot10, Peer: peer2name}}
	require.True(t, ring.checkInvariants() == ErrNotSorted, "Expected error")

	// Check tokens don't appear twice
	ring.Entries = []*entry{{Token: dot245, Peer: peer1name}, {Token: dot245, Peer: peer2name}}
	require.True(t, ring.checkInvariants() == ErrTokenRepeated, "Expected error")

	// Check tokens are in bounds
	ring = NewRing(dot10, dot245, peer1name)
	ring.Entries = []*entry{{Token: start, Peer: peer1name}}
	require.True(t, ring.checkInvariants() == ErrTokenOutOfRange, "Expected error")

	ring.Entries = []*entry{{Token: end, Peer: peer1name}}
	require.True(t, ring.checkInvariants() == ErrTokenOutOfRange, "Expected error")
}

func TestInsert(t *testing.T) {
	ring := NewRing(start, end, peer1name)
	ring.Entries = []*entry{{Token: start, Peer: peer1name, Free: 255}}

	require.Panics(t, func() {
		ring.Entries.insert(entry{Token: start, Peer: peer1name})
	})

	ring.Entries.entry(0).Free = 0
	ring.Entries.insert(entry{Token: dot245, Peer: peer1name})
	check := []RangeInfo{
		{Peer: peer1name, Range: address.Range{Start: start, End: dot245}},
		{Peer: peer1name, Range: address.Range{Start: dot245, End: end}},
	}
	require.Equal(t, check, ring.AllRangeInfo())

	ring.Entries.insert(entry{Token: dot10, Peer: peer1name})
	check2 := []RangeInfo{
		{Peer: peer1name, Range: address.Range{Start: start, End: dot10}},
		{Peer: peer1name, Range: address.Range{Start: dot10, End: dot245}},
		{Peer: peer1name, Range: address.Range{Start: dot245, End: end}},
	}
	require.Equal(t, check2, ring.AllRangeInfo())
}

func TestBetween(t *testing.T) {
	ring1 := NewRing(start, end, peer1name)
	ring1.Entries = []*entry{{Token: start, Peer: peer1name, Free: 255}}

	// First off, in a ring where everything is owned by the peer
	// between should return true for everything
	for i := 1; i <= 255; i++ {
		ip := ParseIP(fmt.Sprintf("10.0.0.%d", i))
		require.True(t, ring1.Entries.between(ip, 0, 1), "between should be true!")
	}

	// Now, construct a ring with entries at +10 and -10
	// And check the correct behaviour

	ring1.Entries = []*entry{{Token: dot10, Peer: peer1name}, {Token: dot245, Peer: peer2name}}
	ring1.assertInvariants()
	for i := 10; i <= 244; i++ {
		ipStr := fmt.Sprintf("10.0.0.%d", i)
		ip := ParseIP(ipStr)
		require.True(t, ring1.Entries.between(ip, 0, 1),
			fmt.Sprintf("Between should be true for %s!", ipStr))
		require.False(t, ring1.Entries.between(ip, 1, 2),
			fmt.Sprintf("Between should be false for %s!", ipStr))
	}
	for i := 0; i <= 9; i++ {
		ipStr := fmt.Sprintf("10.0.0.%d", i)
		ip := ParseIP(ipStr)
		require.False(t, ring1.Entries.between(ip, 0, 1),
			fmt.Sprintf("Between should be false for %s!", ipStr))
		require.True(t, ring1.Entries.between(ip, 1, 2),
			fmt.Sprintf("Between should be true for %s!", ipStr))
	}
	for i := 245; i <= 255; i++ {
		ipStr := fmt.Sprintf("10.0.0.%d", i)
		ip := ParseIP(ipStr)
		require.False(t, ring1.Entries.between(ip, 0, 1),
			fmt.Sprintf("Between should be false for %s!", ipStr))
		require.True(t, ring1.Entries.between(ip, 1, 2),
			fmt.Sprintf("Between should be true for %s!", ipStr))
	}
}

func TestGrantSimple(t *testing.T) {
	ring1 := NewRing(start, end, peer1name)
	ring2 := NewRing(start, end, peer2name)

	// Claim everything for peer1
	ring1.ClaimItAll()
	require.Equal(t, entries{{Token: start, Peer: peer1name, Free: 256}}, ring1.Entries)

	// Now grant everything to peer2
	ring1.GrantRangeToHost(start, end, peer2name)
	ring2.Entries = []*entry{{Token: start, Peer: peer2name, Free: 256, Version: 1}}
	require.Equal(t, ring2.Entries, ring1.Entries)

	// Now spint back to peer 1
	ring2.GrantRangeToHost(dot10, end, peer1name)
	ring1.Entries = []*entry{{Token: start, Peer: peer2name, Free: 10, Version: 2},
		{Token: dot10, Peer: peer1name, Free: 246}}
	require.Equal(t, ring2.Entries, ring1.Entries)

	// And spint back to peer 2 again
	ring1.GrantRangeToHost(dot245, end, peer2name)
	require.Equal(t, entries{{Token: start, Peer: peer2name, Free: 10, Version: 2},
		{Token: dot10, Peer: peer1name, Free: 235, Version: 1},
		{Token: dot245, Peer: peer2name, Free: 11}}, ring1.Entries)

	// Grant range spanning a live token
	ring1.Entries = []*entry{{Token: start, Peer: peer1name, Free: 10, Version: 2},
		{Token: dot10, Peer: peer1name, Free: 235}, {Token: dot245, Peer: peer1name, Free: 10}}
	ring1.GrantRangeToHost(dot10, end, peer2name)
	require.Equal(t, entries{{Token: start, Peer: peer1name, Free: 10, Version: 2},
		{Token: dot10, Peer: peer2name, Free: 235, Version: 1},
		{Token: dot245, Peer: peer2name, Free: 10, Version: 1}}, ring1.Entries)

}

func TestGrantSplit(t *testing.T) {
	ring1 := NewRing(start, end, peer1name)
	ring2 := NewRing(start, end, peer2name)

	// Claim everything for peer1
	ring1.Entries = []*entry{{Token: start, Peer: peer1name, Free: 256}}
	merge(ring2, ring1)
	require.Equal(t, ring2.Entries, ring1.Entries)

	// Now grant a split range to peer2
	ring1.GrantRangeToHost(dot10, dot245, peer2name)
	require.Equal(t, entries{{Token: start, Peer: peer1name, Free: 10, Version: 1},
		{Token: dot10, Peer: peer2name, Free: 235},
		{Token: dot245, Peer: peer1name, Free: 11}}, ring1.Entries)

	ring1.assertInvariants()

	// Grant range spanning a live token, and inserting a new token
	ring1.Entries = []*entry{{Token: start, Peer: peer1name, Free: 10, Version: 2},
		{Token: dot10, Peer: peer1name, Free: 118}, {Token: middle, Peer: peer1name, Free: 127}}
	ring1.GrantRangeToHost(dot10, dot245, peer2name)
	require.Equal(t, entries{{Token: start, Peer: peer1name, Free: 10, Version: 2},
		{Token: dot10, Peer: peer2name, Free: 118, Version: 1},
		{Token: middle, Peer: peer2name, Free: 117, Version: 1},
		{Token: dot245, Peer: peer1name, Free: 11, Version: 0}}, ring1.Entries)

	ring1.assertInvariants()
}

func TestMergeSimple(t *testing.T) {
	ring1 := NewRing(start, end, peer1name)
	ring2 := NewRing(start, end, peer2name)

	// Claim everything for peer1
	ring1.ClaimItAll()
	ring1.GrantRangeToHost(middle, end, peer2name)
	require.NoError(t, merge(ring2, ring1))

	require.Equal(t, entries{{Token: start, Peer: peer1name, Free: 128, Version: 1},
		{Token: middle, Peer: peer2name, Free: 128}}, ring1.Entries)

	require.Equal(t, ring2.Entries, ring1.Entries)

	// Now to two different operations on either side,
	// check we can Merge again
	ring1.GrantRangeToHost(start, middle, peer2name)
	ring2.GrantRangeToHost(middle, end, peer1name)

	require.NoError(t, merge(ring2, ring1))
	require.NoError(t, merge(ring1, ring2))

	require.Equal(t, entries{{Token: start, Peer: peer2name, Free: 128, Version: 2},
		{Token: middle, Peer: peer1name, Free: 128, Version: 1}}, ring1.Entries)

	require.Equal(t, ring2.Entries, ring1.Entries)
}

func TestMergeErrors(t *testing.T) {
	// Cannot Merge in an invalid ring
	ring1 := NewRing(start, end, peer1name)
	ring2 := NewRing(start, end, peer2name)
	ring2.Entries = []*entry{{Token: middle, Peer: peer2name}, {Token: start, Peer: peer2name}}
	require.True(t, merge(ring1, ring2) == ErrNotSorted, "Expected ErrNotSorted")

	// Should Merge two rings for different ranges
	ring2 = NewRing(start, middle, peer2name)
	ring2.Entries = []*entry{}
	require.True(t, merge(ring1, ring2) == ErrDifferentRange, "Expected ErrDifferentRange")

	// Cannot Merge newer version of entry I own
	ring2 = NewRing(start, end, peer2name)
	ring1.Entries = []*entry{{Token: start, Peer: peer1name}}
	ring2.Entries = []*entry{{Token: start, Peer: peer1name, Version: 1}}
	fmt.Println(merge(ring1, ring2))
	require.Error(t, merge(ring1, ring2), "Expected error")

	// Cannot Merge two entries with same version but different hosts
	ring1.Entries = []*entry{{Token: start, Peer: peer1name}}
	ring2.Entries = []*entry{{Token: start, Peer: peer2name}}
	require.Error(t, merge(ring1, ring2), "Expected error")

	// Cannot Merge an entry into a range I own
	ring1.Entries = []*entry{{Token: start, Peer: peer1name}}
	ring2.Entries = []*entry{{Token: middle, Peer: peer2name}}
	require.Error(t, merge(ring1, ring2), "Expected error")
}

func TestMergeMore(t *testing.T) {
	ring1 := NewRing(start, end, peer1name)
	ring2 := NewRing(start, end, peer2name)

	assertRing := func(ring *Ring, entries entries) {
		require.Equal(t, entries, ring.Entries)
	}

	assertRing(ring1, []*entry{})
	assertRing(ring2, []*entry{})

	// Claim everything for peer1
	ring1.ClaimItAll()
	assertRing(ring1, []*entry{{Token: start, Peer: peer1name, Free: 256}})
	assertRing(ring2, []*entry{})

	// Check the Merge sends it to the other ring
	require.NoError(t, merge(ring2, ring1))
	assertRing(ring1, []*entry{{Token: start, Peer: peer1name, Free: 256}})
	assertRing(ring2, []*entry{{Token: start, Peer: peer1name, Free: 256}})

	// Give everything to peer2
	ring1.GrantRangeToHost(start, end, peer2name)
	assertRing(ring1, []*entry{{Token: start, Peer: peer2name, Free: 256, Version: 1}})
	assertRing(ring2, []*entry{{Token: start, Peer: peer1name, Free: 256}})

	require.NoError(t, merge(ring2, ring1))
	assertRing(ring1, []*entry{{Token: start, Peer: peer2name, Free: 256, Version: 1}})
	assertRing(ring2, []*entry{{Token: start, Peer: peer2name, Free: 256, Version: 1}})

	// And carve off some space
	ring2.GrantRangeToHost(middle, end, peer1name)
	assertRing(ring2, []*entry{{Token: start, Peer: peer2name, Free: 128, Version: 2},
		{Token: middle, Peer: peer1name, Free: 128}})
	assertRing(ring1, []*entry{{Token: start, Peer: peer2name, Free: 256, Version: 1}})

	// And Merge back
	require.NoError(t, merge(ring1, ring2))
	assertRing(ring1, []*entry{{Token: start, Peer: peer2name, Free: 128, Version: 2},
		{Token: middle, Peer: peer1name, Free: 128}})
	assertRing(ring2, []*entry{{Token: start, Peer: peer2name, Free: 128, Version: 2},
		{Token: middle, Peer: peer1name, Free: 128}})

	// This should be a no-op
	require.NoError(t, merge(ring2, ring1))
	assertRing(ring1, []*entry{{Token: start, Peer: peer2name, Free: 128, Version: 2},
		{Token: middle, Peer: peer1name, Free: 128}})
	assertRing(ring2, []*entry{{Token: start, Peer: peer2name, Free: 128, Version: 2},
		{Token: middle, Peer: peer1name, Free: 128}})
}

func TestMergeSplit(t *testing.T) {
	ring1 := NewRing(start, end, peer1name)
	ring2 := NewRing(start, end, peer2name)

	// Claim everything for peer2
	ring1.Entries = []*entry{{Token: start, Peer: peer2name, Free: 256}}
	require.NoError(t, merge(ring2, ring1))
	require.Equal(t, ring2.Entries, ring1.Entries)

	// Now grant a split range to peer1
	ring2.GrantRangeToHost(dot10, dot245, peer1name)
	require.Equal(t, entries{{Token: start, Peer: peer2name, Free: 10, Version: 1},
		{Token: dot10, Peer: peer1name, Free: 235},
		{Token: dot245, Peer: peer2name, Free: 11}}, ring2.Entries)

	require.NoError(t, merge(ring1, ring2))
	require.Equal(t, entries{{Token: start, Peer: peer2name, Free: 10, Version: 1},
		{Token: dot10, Peer: peer1name, Free: 235},
		{Token: dot245, Peer: peer2name, Free: 11}}, ring1.Entries)

	require.Equal(t, ring2.Entries, ring1.Entries)
}

func TestMergeSplit2(t *testing.T) {
	ring1 := NewRing(start, end, peer1name)
	ring2 := NewRing(start, end, peer2name)

	// Claim everything for peer2
	ring1.Entries = []*entry{{Token: start, Peer: peer2name, Free: 250}, {Token: dot250, Peer: peer2name, Free: 5}}
	require.NoError(t, merge(ring2, ring1))
	require.Equal(t, ring2.Entries, ring1.Entries)

	// Now grant a split range to peer1
	ring2.GrantRangeToHost(dot10, dot245, peer1name)
	require.Equal(t, entries{{Token: start, Peer: peer2name, Free: 10, Version: 1},
		{Token: dot10, Peer: peer1name, Free: 235},
		{Token: dot245, Peer: peer2name, Free: 5}, {Token: dot250, Peer: peer2name, Free: 5}}, ring2.Entries)

	require.NoError(t, merge(ring1, ring2))
	require.Equal(t, entries{{Token: start, Peer: peer2name, Free: 10, Version: 1},
		{Token: dot10, Peer: peer1name, Free: 235},
		{Token: dot245, Peer: peer2name, Free: 5}, {Token: dot250, Peer: peer2name, Free: 5}}, ring1.Entries)

	require.Equal(t, ring2.Entries, ring1.Entries)
}

// A simple test, very similar to above, but using the marshalling to byte[]s
func TestGossip(t *testing.T) {
	ring1 := NewRing(start, end, peer1name)
	ring2 := NewRing(start, end, peer2name)

	assertRing := func(ring *Ring, entries entries) {
		require.Equal(t, entries, ring.Entries)
	}

	assertRing(ring1, []*entry{})
	assertRing(ring2, []*entry{})

	// Claim everything for peer1
	ring1.ClaimItAll()
	assertRing(ring1, []*entry{{Token: start, Peer: peer1name, Free: 256}})
	assertRing(ring2, []*entry{})

	// Check the Merge sends it to the other ring
	require.NoError(t, merge(ring2, ring1))
	assertRing(ring1, []*entry{{Token: start, Peer: peer1name, Free: 256}})
	assertRing(ring2, []*entry{{Token: start, Peer: peer1name, Free: 256}})
}

func assertPeersWithSpace(t *testing.T, ring *Ring, start, end address.Address, expected int) []mesh.PeerName {
	peers := ring.ChoosePeersToAskForSpace(start, end)
	require.Equal(t, expected, len(peers))
	return peers
}

func TestFindFree(t *testing.T) {
	ring1 := NewRing(start, end, peer1name)

	assertPeersWithSpace(t, ring1, start, end, 0)

	ring1.Entries = []*entry{{Token: start, Peer: peer1name}}
	assertPeersWithSpace(t, ring1, start, end, 0)

	// We shouldn't return outselves
	ring1.ReportFree(map[address.Address]address.Count{start: 10})
	assertPeersWithSpace(t, ring1, start, end, 0)

	ring1.Entries = []*entry{{Token: start, Peer: peer1name, Free: 1},
		{Token: middle, Peer: peer1name, Free: 1}}
	assertPeersWithSpace(t, ring1, start, end, 0)
	ring1.assertInvariants()

	// We should return others
	var peers []mesh.PeerName

	ring1.Entries = []*entry{{Token: start, Peer: peer2name, Free: 1}}
	peers = assertPeersWithSpace(t, ring1, start, end, 1)
	require.Equal(t, peer2name, peers[0])

	ring1.Entries = []*entry{{Token: start, Peer: peer2name, Free: 1},
		{Token: middle, Peer: peer3name, Free: 1}}
	peers = assertPeersWithSpace(t, ring1, start, middle, 1)
	require.Equal(t, peer2name, peers[0])

	peers = assertPeersWithSpace(t, ring1, middle, end, 1)
	require.Equal(t, peer3name, peers[0])

	assertPeersWithSpace(t, ring1, start, end, 2)
	ring1.assertInvariants()
}

func TestReportFree(t *testing.T) {
	ring1 := NewRing(start, end, peer1name)
	ring2 := NewRing(start, end, peer2name)

	ring1.ClaimItAll()
	ring1.GrantRangeToHost(middle, end, peer2name)
	require.NoError(t, merge(ring2, ring1))

	freespace := make(map[address.Address]address.Count)
	for _, r := range ring2.OwnedRanges() {
		freespace[r.Start] = 0
	}
	ring2.ReportFree(freespace)
}

func TestMisc(t *testing.T) {
	ring := NewRing(start, end, peer1name)

	require.True(t, ring.Empty(), "empty")

	ring.ClaimItAll()
	println(ring.String())
}

func TestEmptyGossip(t *testing.T) {
	ring1 := NewRing(start, end, peer1name)
	ring2 := NewRing(start, end, peer2name)

	ring1.ClaimItAll()
	// This used to panic, and it shouldn't
	require.NoError(t, merge(ring1, ring2))
}

func TestMergeOldMessage(t *testing.T) {
	ring1 := NewRing(start, end, peer1name)
	ring2 := NewRing(start, end, peer2name)

	ring1.ClaimItAll()
	require.NoError(t, merge(ring2, ring1))

	ring1.GrantRangeToHost(middle, end, peer1name)
	require.NoError(t, merge(ring1, ring2))
}

func TestSplitRangeAtBeginning(t *testing.T) {
	ring1 := NewRing(start, end, peer1name)
	ring2 := NewRing(start, end, peer2name)

	ring1.ClaimItAll()
	require.NoError(t, merge(ring2, ring1))

	ring1.GrantRangeToHost(start, middle, peer2name)
	require.NoError(t, merge(ring2, ring1))
}

func TestOwnedRange(t *testing.T) {
	ring1 := NewRing(start, end, peer1name)
	ring1.ClaimItAll()

	require.Equal(t, []address.Range{{Start: start, End: end}}, ring1.OwnedRanges())

	ring1.GrantRangeToHost(middle, end, peer2name)
	require.Equal(t, []address.Range{{Start: start, End: middle}}, ring1.OwnedRanges())

	ring2 := NewRing(start, end, peer2name)
	merge(ring2, ring1)
	require.Equal(t, []address.Range{{Start: middle, End: end}}, ring2.OwnedRanges())

	ring2.Entries = []*entry{{Token: middle, Peer: peer2name}}
	require.Equal(t, []address.Range{{Start: start, End: middle}, {Start: middle, End: end}}, ring2.OwnedRanges())

	ring2.Entries = []*entry{{Token: dot10, Peer: peer2name}, {Token: middle, Peer: peer2name}}
	require.Equal(t, []address.Range{{Start: start, End: dot10}, {Start: dot10, End: middle},
		{Start: middle, End: end}}, ring2.OwnedRanges())

}

func TestTransfer(t *testing.T) {
	// First test just checks if we can grant some range to a host, when we transfer it, we get it back
	ring1 := NewRing(start, end, peer1name)
	ring1.ClaimItAll()
	ring1.GrantRangeToHost(middle, end, peer2name)
	ring1.Transfer(peer2name, peer1name)
	require.Equal(t, []address.Range{{Start: start, End: middle}, {Start: middle, End: end}}, ring1.OwnedRanges())

	// Second test is what happens when a token exists at the end of a range but is transferred
	// - does it get resurrected correctly?
	ring1 = NewRing(start, end, peer1name)
	ring1.ClaimItAll()
	ring1.GrantRangeToHost(middle, end, peer2name)
	ring1.Transfer(peer2name, peer1name)
	ring1.GrantRangeToHost(dot10, middle, peer2name)
	require.Equal(t, []address.Range{{Start: start, End: dot10}, {Start: middle, End: end}}, ring1.OwnedRanges())
}

func TestOwner(t *testing.T) {
	ring1 := NewRing(start, end, peer1name)
	require.True(t, ring1.Contains(start), "start should be in ring")
	require.False(t, ring1.Contains(end), "end should not be in ring")

	require.Equal(t, mesh.UnknownPeerName, ring1.Owner(start))

	ring1.ClaimItAll()
	ring1.GrantRangeToHost(middle, end, peer2name)
	require.Equal(t, peer1name, ring1.Owner(start))
	require.Equal(t, peer2name, ring1.Owner(middle))
	require.Panics(t, func() {
		ring1.Owner(end)
	})

}

func makePeerName(i int) mesh.PeerName {
	if i >= 10000 {
		panic("makePeerName: invalid value")
	}
	peer, _ := mesh.PeerNameFromString(fmt.Sprintf("%02d:%02d:00:00:00:ff", i/100, i%100))
	return peer
}

func makePeers(numPeers int) []mesh.PeerName {
	peers := make([]mesh.PeerName, numPeers)
	for i := 0; i < numPeers; i++ {
		peers[i] = makePeerName(i)
	}
	return peers
}

func TestClaimForPeers(t *testing.T) {
	const numPeers = 12
	// Different end to usual so we get a number of addresses that a)
	// is smaller than the max number of peers, and b) is divisible by
	// some number of peers. This maximises coverage of edge cases.
	end := dot8
	peers := makePeers(numPeers)
	// Test for a range of peer counts
	for i := 0; i < numPeers; i++ {
		ring := NewRing(start, end, peers[0])
		ring.ClaimForPeers(peers[:i+1])
	}
}

type addressSlice []address.Address

func (s addressSlice) Len() int           { return len(s) }
func (s addressSlice) Less(i, j int) bool { return s[i] < s[j] }
func (s addressSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func TestFuzzRing(t *testing.T) {
	var (
		numPeers   = 25
		iterations = 1000
	)

	peers := make([]mesh.PeerName, numPeers)
	for i := 0; i < numPeers; i++ {
		peers[i] = makePeerName(i)
	}

	// Make a valid, random ring
	makeGoodRandomRing := func() *Ring {
		addressSpace := end - start
		numTokens := rand.Intn(int(addressSpace))

		tokenMap := make(map[address.Address]bool)
		for i := 0; i < numTokens; i++ {
			tokenMap[address.Address(rand.Intn(int(addressSpace)))] = true
		}
		var tokens []address.Address
		for token := range tokenMap {
			tokens = append(tokens, token)
		}
		sort.Sort(addressSlice(tokens))

		peer := peers[rand.Intn(len(peers))]
		ring := NewRing(start, end, peer)
		for _, token := range tokens {
			peer = peers[rand.Intn(len(peers))]
			ring.Entries = append(ring.Entries, &entry{Token: start + token, Peer: peer})
		}

		ring.assertInvariants()
		return ring
	}

	for i := 0; i < iterations; i++ {
		// make 2 random rings
		ring1 := makeGoodRandomRing()
		ring2 := makeGoodRandomRing()

		// Merge them - this might fail, we don't care
		// We just want to make sure it doesn't panic
		merge(ring1, ring2)

		// Check whats left still passes assertions
		ring1.assertInvariants()
		ring2.assertInvariants()
	}

	// Make an invalid, random ring
	makeBadRandomRing := func() *Ring {
		addressSpace := end - start
		numTokens := rand.Intn(int(addressSpace))
		tokens := make([]address.Address, numTokens)
		for i := 0; i < numTokens; i++ {
			tokens[i] = address.Address(rand.Intn(int(addressSpace)))
		}

		peer := peers[rand.Intn(len(peers))]
		ring := NewRing(start, end, peer)
		for _, token := range tokens {
			peer = peers[rand.Intn(len(peers))]
			ring.Entries = append(ring.Entries, &entry{Token: start + token, Peer: peer})
		}

		return ring
	}

	for i := 0; i < iterations; i++ {
		// make 2 random rings
		ring1 := makeGoodRandomRing()
		ring2 := makeBadRandomRing()

		// Merge them - this might fail, we don't care
		// We just want to make sure it doesn't panic
		merge(ring1, ring2)

		// Check whats left still passes assertions
		ring1.assertInvariants()
	}
}

func TestFuzzRingHard(t *testing.T) {
	//common.SetLogLevel("debug")
	var (
		numPeers   = 100
		iterations = 3000
		peers      []mesh.PeerName
		rings      []*Ring
		nextPeerID = 0
	)

	addPeer := func() {
		peer := makePeerName(nextPeerID)
		common.Log.Debugf("%s: Adding peer", peer)
		nextPeerID++
		peers = append(peers, peer)
		rings = append(rings, NewRing(start, end, peer))
	}

	for i := 0; i < numPeers; i++ {
		addPeer()
	}

	rings[0].ClaimItAll()

	randomPeer := func(exclude int) (int, mesh.PeerName, *Ring) {
		var peerIndex int
		if exclude >= 0 {
			peerIndex = rand.Intn(len(peers) - 1)
			if peerIndex == exclude {
				peerIndex++
			}
		} else {
			peerIndex = rand.Intn(len(peers))
		}
		return peerIndex, peers[peerIndex], rings[peerIndex]
	}

	// Keep a map of index -> ranges, as these are a little expensive to
	// calculate for every ring on every iteration.
	var theRanges = make(map[int][]address.Range)
	theRanges[0] = rings[0].OwnedRanges()

	addOrRmPeer := func() {
		if len(peers) < numPeers {
			addPeer()
			return
		}

		// Pick one peer to remove, and a different one to transfer to
		peerIndex, peername, _ := randomPeer(-1)
		_, otherPeername, otherRing := randomPeer(peerIndex)

		// We need to be in a ~converged ring to rmpeer
		for _, ring := range rings {
			require.NoError(t, merge(otherRing, ring))
		}

		common.Log.Debugf("%s: transferring from peer %s", otherPeername, peername)
		otherRing.Transfer(peername, otherPeername)

		// Remove peer from our state
		peers = append(peers[:peerIndex], peers[peerIndex+1:]...)
		rings = append(rings[:peerIndex], rings[peerIndex+1:]...)
		theRanges = make(map[int][]address.Range)

		// And now tell everyone about the transfer - rmpeer is
		// not partition safe
		for i, ring := range rings {
			require.NoError(t, merge(ring, otherRing))
			theRanges[i] = ring.OwnedRanges()
		}
	}

	doGrantOrGossip := func() {
		var ringsWithRanges = make([]int, 0, len(rings))
		for index, ranges := range theRanges {
			if len(ranges) > 0 {
				ringsWithRanges = append(ringsWithRanges, index)
			}
		}

		if len(ringsWithRanges) > 0 {
			// Produce a random split in a random owned range, given to a random peer
			indexWithRanges := ringsWithRanges[rand.Intn(len(ringsWithRanges))]
			ownedRanges := theRanges[indexWithRanges]
			ring := rings[indexWithRanges]

			rangeToSplit := ownedRanges[rand.Intn(len(ownedRanges))]
			size := address.Subtract(rangeToSplit.End, rangeToSplit.Start)
			ipInRange := address.Add(rangeToSplit.Start, address.Offset(rand.Intn(int(size))))
			_, peerToGiveTo, _ := randomPeer(-1)
			common.Log.Debugf("%s: Granting [%v, %v) to %s", ring.Peer, ipInRange, rangeToSplit.End, peerToGiveTo)
			ring.GrantRangeToHost(ipInRange, rangeToSplit.End, peerToGiveTo)

			// Now 'gossip' this to a random host (note, note could be same host as above)
			otherIndex, _, otherRing := randomPeer(-1)
			common.Log.Debugf("%s: 'Gossiping' to %s", ring.Peer, otherRing.Peer)
			require.NoError(t, merge(otherRing, ring))

			theRanges[indexWithRanges] = ring.OwnedRanges()
			theRanges[otherIndex] = otherRing.OwnedRanges()
			return
		}

		// No rings think they own anything (as gossip might be behind)
		// We're going to pick a random host (which has entries) and gossip
		// it to a random host (which may or may not have entries).
		var ringsWithEntries = make([]*Ring, 0, len(rings))
		for _, ring := range rings {
			if len(ring.Entries) > 0 {
				ringsWithEntries = append(ringsWithEntries, ring)
			}
		}
		ring1 := ringsWithEntries[rand.Intn(len(ringsWithEntries))]
		ring2index, _, ring2 := randomPeer(-1)
		common.Log.Debugf("%s: 'Gossiping' to %s", ring1.Peer, ring2.Peer)
		require.NoError(t, merge(ring2, ring1))
		theRanges[ring2index] = ring2.OwnedRanges()
	}

	for i := 0; i < iterations; i++ {
		// about 1 in 10 times, rmpeer or add host
		n := rand.Intn(10)
		switch {
		case n < 1:
			addOrRmPeer()
		default:
			doGrantOrGossip()
		}
	}
}

func (r *Ring) ClaimItAll() {
	r.ClaimForPeers([]mesh.PeerName{r.Peer})
}

func (es entries) String() string {
	var buffer bytes.Buffer
	fmt.Fprintf(&buffer, "[")
	for i, entry := range es {
		fmt.Fprintf(&buffer, "%+v", *entry)
		if i+1 < len(es) {
			fmt.Fprintf(&buffer, " ")
		}
	}
	fmt.Fprintf(&buffer, "]")
	return buffer.String()
}
