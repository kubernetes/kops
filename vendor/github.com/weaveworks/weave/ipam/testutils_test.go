package ipam

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/weaveworks/mesh"

	"github.com/stretchr/testify/require"
	"github.com/weaveworks/weave/net/address"
	"github.com/weaveworks/weave/testing/gossip"
)

type mockMessage struct {
	dst     mesh.PeerName
	msgType byte
	buf     []byte
}

func (m *mockMessage) String() string {
	return fmt.Sprintf("-> %s [%x]", m.dst, m.buf)
}

func toStringArray(messages []mockMessage) []string {
	out := make([]string, len(messages))
	for i := range out {
		out[i] = messages[i].String()
	}
	return out
}

type mockGossipComms struct {
	sync.RWMutex
	*testing.T
	name     string
	messages []mockMessage
}

func (m *mockGossipComms) String() string {
	m.RLock()
	defer m.RUnlock()
	return fmt.Sprintf("[mockGossipComms %s]", m.name)
}

// Note: this style of verification, using equalByteBuffer, requires
// that the contents of messages are never re-ordered.  Which, for instance,
// requires they are not based off iterating through a map.

func (m *mockGossipComms) GossipBroadcast(update mesh.GossipData) {
	m.Lock()
	defer m.Unlock()
	buf := []byte{}
	if len(m.messages) == 0 {
		require.FailNow(m, fmt.Sprintf("%s: Gossip broadcast message unexpected: \n%x", m.name, buf))
	} else if msg := m.messages[0]; msg.dst != mesh.UnknownPeerName {
		require.FailNow(m, fmt.Sprintf("%s: Expected Gossip message to %s but got broadcast", m.name, msg.dst))
	} else if msg.buf != nil && !equalByteBuffer(msg.buf, buf) {
		require.FailNow(m, fmt.Sprintf("%s: Gossip message not sent as expected: \nwant: %x\ngot : %x", m.name, msg.buf, buf))
	} else {
		// Swallow this message
		m.messages = m.messages[1:]
	}
}

func equalByteBuffer(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func (m *mockGossipComms) GossipUnicast(dstPeerName mesh.PeerName, buf []byte) error {
	m.Lock()
	defer m.Unlock()
	if len(m.messages) == 0 {
		require.FailNow(m, fmt.Sprintf("%s: Gossip message to %s unexpected: \n%s", m.name, dstPeerName, buf))
	} else if msg := m.messages[0]; msg.dst == mesh.UnknownPeerName {
		require.FailNow(m, fmt.Sprintf("%s: Expected Gossip broadcast message but got dest %s", m.name, dstPeerName))
	} else if msg.dst != dstPeerName {
		require.FailNow(m, fmt.Sprintf("%s: Expected Gossip message to %s but got dest %s", m.name, msg.dst, dstPeerName))
	} else if buf[0] != msg.msgType {
		require.FailNow(m, fmt.Sprintf("%s: Expected Gossip message of type %d but got type %d", m.name, msg.msgType, buf[0]))
	} else if msg.buf != nil && !equalByteBuffer(msg.buf, buf[1:]) {
		require.FailNow(m, fmt.Sprintf("%s: Gossip message not sent as expected: \nwant: %x\ngot : %x", m.name, msg.buf, buf[1:]))
	} else {
		// Swallow this message
		m.messages = m.messages[1:]
	}
	return nil
}

func ExpectMessage(alloc *Allocator, dst string, msgType byte, buf []byte) {
	m := alloc.gossip.(*mockGossipComms)
	dstPeerName, _ := mesh.PeerNameFromString(dst)
	m.Lock()
	m.messages = append(m.messages, mockMessage{dstPeerName, msgType, buf})
	m.Unlock()
}

func ExpectBroadcastMessage(alloc *Allocator, buf []byte) {
	m := alloc.gossip.(*mockGossipComms)
	m.Lock()
	m.messages = append(m.messages, mockMessage{mesh.UnknownPeerName, 0, buf})
	m.Unlock()
}

func CheckAllExpectedMessagesSent(allocs ...*Allocator) {
	for _, alloc := range allocs {
		m := alloc.gossip.(*mockGossipComms)
		m.RLock()
		if len(m.messages) > 0 {
			require.FailNow(m, fmt.Sprintf("%s: Gossip message(s) not sent as expected: \n%x", m.name, m.messages))
		}
		m.RUnlock()
	}
}

type mockDB struct{}

func (d *mockDB) Load(_ string, _ interface{}) (bool, error) { return false, nil }
func (d *mockDB) Save(_ string, _ interface{}) error         { return nil }

func makeAllocator(name string, cidrStr string, quorum uint, preClaims ...PreClaim) (*Allocator, address.CIDR) {
	peername, err := mesh.PeerNameFromString(name)
	if err != nil {
		panic(err)
	}

	cidr, err := address.ParseCIDR(cidrStr)
	if err != nil {
		panic(err)
	}

	return NewAllocator(Config{
		OurName:     peername,
		OurUID:      mesh.PeerUID(rand.Int63()),
		OurNickname: "nick-" + name,
		Universe:    cidr,
		IsObserver:  quorum == 0,
		PreClaims:   preClaims,
		Quorum:      func() uint { return quorum },
		Db:          new(mockDB),
		IsKnownPeer: func(mesh.PeerName) bool { return true },
	}), cidr
}

func makeAllocatorWithMockGossip(t *testing.T, name string, universeCIDR string, quorum uint) (*Allocator, address.CIDR) {
	alloc, subnet := makeAllocator(name, universeCIDR, quorum)
	gossip := &mockGossipComms{T: t, name: name}
	alloc.SetInterfaces(gossip)
	alloc.Start()
	return alloc, subnet
}

func (alloc *Allocator) claimRingForTesting(allocs ...*Allocator) {
	peers := []mesh.PeerName{alloc.ourName}
	for _, alloc2 := range allocs {
		peers = append(peers, alloc2.ourName)
	}
	alloc.ring.ClaimForPeers(normalizeConsensus(peers))
	alloc.space.AddRanges(alloc.ring.OwnedRanges())
}

func (alloc *Allocator) NumFreeAddresses(r address.Range) address.Count {
	resultChan := make(chan address.Count)
	alloc.actionChan <- func() {
		resultChan <- alloc.space.NumFreeAddressesInRange(r)
	}
	return <-resultChan
}

func (alloc *Allocator) OwnedRanges() (result []address.Range) {
	resultChan := make(chan []address.Range)
	alloc.actionChan <- func() {
		resultChan <- alloc.ring.OwnedRanges()
	}
	return <-resultChan
}

// Check whether or not something was sent on a channel
func AssertSent(t *testing.T, ch <-chan bool) {
	timeout := time.After(10 * time.Second)
	select {
	case <-ch:
		// This case is ok
	case <-timeout:
		require.FailNow(t, "Nothing sent on channel")
	}
}

func AssertNothingSent(t *testing.T, ch <-chan bool) {
	select {
	case val := <-ch:
		require.FailNow(t, fmt.Sprintf("Unexpected value on channel: %v", val))
	default:
		// no message received
	}
}

func AssertNothingSentErr(t *testing.T, ch <-chan error) {
	select {
	case val := <-ch:
		require.FailNow(t, fmt.Sprintf("Unexpected value on channel: %v", val))
	default:
		// no message received
	}
}

func makeNetworkOfAllocators(size int, cidr string, preClaims ...[]PreClaim) ([]*Allocator, *gossip.TestRouter, address.CIDR) {
	gossipRouter := gossip.NewTestRouter(0.0)
	allocs := make([]*Allocator, size)
	var subnet address.CIDR

	for i := 0; i < size; i++ {
		var alloc *Allocator
		preClaim := []PreClaim{}
		if i < len(preClaims) {
			preClaim = preClaims[i]
		}
		alloc, subnet = makeAllocator(fmt.Sprintf("%02d:00:00:02:00:00", i),
			cidr, uint(size/2+1), preClaim...)
		alloc.SetInterfaces(gossipRouter.Connect(alloc.ourName, alloc))
		alloc.Start()
		allocs[i] = alloc
	}

	allocs[size-1].gossip.GossipBroadcast(allocs[size-1].Gossip())
	gossipRouter.Flush()
	return allocs, gossipRouter, subnet
}

func stopNetworkOfAllocators(allocs []*Allocator, gossipRouter *gossip.TestRouter) {
	// NB: We must stop the router first since ipam gossip makes
	// synchronous calls into the allocator.
	gossipRouter.Stop()
	for _, alloc := range allocs {
		alloc.Stop()
	}
}
