package ipam

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/weaveworks/weave/common"
	"github.com/weaveworks/weave/net/address"
	"github.com/weaveworks/weave/testing/gossip"
)

const (
	testStart1 = "10.0.1.0"
	testStart2 = "10.0.2.0"
	testStart3 = "10.0.3.0"
)

func returnFalse() bool { return false }

func (alloc *Allocator) SimplyAllocate(ident string, cidr address.CIDR) (address.Address, error) {
	return alloc.Allocate(ident, cidr, true, returnFalse)
}

func (alloc *Allocator) SimplyClaim(ident string, cidr address.CIDR) error {
	return alloc.Claim(ident, cidr, true, true, returnFalse)
}

func TestAllocFree(t *testing.T) {
	const (
		container1 = "abcdef"
		container2 = "baddf00d"
		container3 = "b01df00d"
		universe   = "10.0.3.0/26"
		subnet1    = "10.0.3.0/28"
		subnet2    = "10.0.3.32/28"
		testAddr1  = "10.0.3.1"
		testAddr2  = "10.0.3.33"
		spaceSize  = 62 // 64 IP addresses in /26, minus .0 and .63
	)

	alloc, subnet := makeAllocatorWithMockGossip(t, "01:00:00:01:00:00", universe, 1)
	defer alloc.Stop()
	cidr1, _ := address.ParseCIDR(subnet1)
	cidr2, _ := address.ParseCIDR(subnet2)

	alloc.claimRingForTesting()
	addr1, err := alloc.SimplyAllocate(container1, cidr1)
	require.NoError(t, err)
	require.Equal(t, testAddr1, addr1.String(), "address")

	addr2, err := alloc.SimplyAllocate(container1, cidr2)
	require.NoError(t, err)
	require.Equal(t, testAddr2, addr2.String(), "address")

	addrs, err := alloc.Lookup(container1, subnet.Range())
	require.NoError(t, err)
	require.Equal(t, []address.CIDR{address.MakeCIDR(cidr1, addr1), address.MakeCIDR(cidr2, addr2)}, addrs)

	// Ask for another address for a different container and check it's different
	addr1b, _ := alloc.SimplyAllocate(container2, cidr1)
	if addr1b.String() == testAddr1 {
		t.Fatalf("Expected different address but got %s", addr1b.String())
	}

	// Ask for the first container again and we should get the same addresses again
	addr1a, _ := alloc.SimplyAllocate(container1, cidr1)
	require.Equal(t, testAddr1, addr1a.String(), "address")
	addr2a, _ := alloc.SimplyAllocate(container1, cidr2)
	require.Equal(t, testAddr2, addr2a.String(), "address")

	// Now delete the first container, and we should get its addresses back
	require.NoError(t, alloc.Delete(container1))
	addr3, _ := alloc.SimplyAllocate(container3, cidr1)
	require.Equal(t, testAddr1, addr3.String(), "address")
	addr4, _ := alloc.SimplyAllocate(container3, cidr2)
	require.Equal(t, testAddr2, addr4.String(), "address")

	alloc.ContainerDied(container2)

	// Resurrect
	addr1c, err := alloc.SimplyAllocate(container2, cidr1)
	require.NoError(t, err)
	require.Equal(t, addr1b, addr1c, "address")

	alloc.ContainerDied(container3)
	alloc.Encode() // sync up
	// Move the clock forward and clear out the dead container
	alloc.actionChan <- func() { alloc.now = func() time.Time { return time.Now().Add(containerDiedTimeout * 2) } }
	alloc.actionChan <- func() { alloc.removeDeadContainers() }
	require.Equal(t, address.Count(spaceSize+1), alloc.NumFreeAddresses(subnet.Range()))
}

func TestBootstrap(t *testing.T) {
	const (
		donateSize     = 5
		donateStart    = "10.0.1.7"
		ourNameString  = "01:00:00:01:00:00"
		peerNameString = "02:00:00:02:00:00"
	)

	alloc1, subnet := makeAllocatorWithMockGossip(t, ourNameString, testStart1+"/22", 2)
	defer alloc1.Stop()

	// Simulate another peer on the gossip network
	alloc2, _ := makeAllocatorWithMockGossip(t, peerNameString, testStart1+"/22", 2)
	defer alloc2.Stop()

	alloc1.OnGossipBroadcast(alloc2.ourName, alloc2.Encode())

	alloc1.actionChan <- func() { alloc1.tryPendingOps() }

	ExpectBroadcastMessage(alloc1, nil) // alloc1 will try to form consensus
	done := make(chan bool)
	go func() {
		alloc1.Allocate("somecontainer", subnet, true, returnFalse)
		done <- true
	}()
	time.Sleep(100 * time.Millisecond)
	AssertNothingSent(t, done)

	CheckAllExpectedMessagesSent(alloc1, alloc2)

	// Check that tryPendingOps doesn't cause anything to happen here
	alloc1.actionChan <- func() { alloc1.tryPendingOps() }
	AssertNothingSent(t, done)

	CheckAllExpectedMessagesSent(alloc1, alloc2)

	// alloc2 receives paxos update and broadcasts its reply
	ExpectBroadcastMessage(alloc2, nil)
	alloc2.OnGossipBroadcast(alloc1.ourName, alloc1.Encode())

	ExpectBroadcastMessage(alloc1, nil)
	alloc1.OnGossipBroadcast(alloc2.ourName, alloc2.Encode())

	CheckAllExpectedMessagesSent(alloc1, alloc2)

	// at this point, both nodes have seen each other, but don't have a valid proposal
	// so there is another exchange of messages
	ExpectBroadcastMessage(alloc2, nil)
	alloc2.OnGossipBroadcast(alloc1.ourName, alloc1.Encode())

	// we have consensus now so alloc1 will initialize the ring
	ExpectBroadcastMessage(alloc1, nil)
	alloc1.OnGossipBroadcast(alloc2.ourName, alloc2.Encode())

	CheckAllExpectedMessagesSent(alloc1, alloc2)

	// alloc2 receives the ring and does not reply
	alloc2.OnGossipBroadcast(alloc1.ourName, alloc1.Encode())

	CheckAllExpectedMessagesSent(alloc1, alloc2)

	// now alloc1 should have space

	AssertSent(t, done)

	CheckAllExpectedMessagesSent(alloc1, alloc2)
}

func TestAllocatorClaim(t *testing.T) {
	const (
		container1 = "abcdef"
		container3 = "b01df00d"
		universe   = "10.0.3.0/24"
		testAddr1  = "10.0.3.5/24"
		testAddr2  = "10.0.4.2/24"
		testPre    = "10.0.3.1/24"
	)

	preAddr, _ := address.ParseCIDR(testPre)
	allocs, router, subnet := makeNetworkOfAllocators(2, universe, []PreClaim{{container1, true, preAddr}})
	defer stopNetworkOfAllocators(allocs, router)
	alloc := allocs[1]
	addr1, _ := address.ParseCIDR(testAddr1)

	// First claim should trigger "dunno, I'm going to wait"
	err := alloc.SimplyClaim(container3, addr1)
	require.NoError(t, err)

	alloc.Prime()
	// Do an allocate on the other peer, which we will try to claim later
	addrx, err := allocs[0].Allocate(container1, subnet, true, returnFalse)
	require.NoError(t, err)
	// Should not get the address we pre-claimed
	require.NotEqual(t, addrx, preAddr)
	router.Flush()

	// Now try the claim again
	err = alloc.SimplyClaim(container3, addr1)
	require.NoError(t, err)
	// Check we get this address back if we try an allocate
	addr3, _ := alloc.SimplyAllocate(container3, subnet)
	require.Equal(t, testAddr1, address.MakeCIDR(subnet, addr3).String(), "address")
	// one more claim should still work
	err = alloc.SimplyClaim(container3, addr1)
	require.NoError(t, err)
	// claim for a different container should fail
	err = alloc.SimplyClaim(container1, addr1)
	require.Error(t, err)
	// claiming the address allocated on the other peer should fail
	err = alloc.SimplyClaim(container1, address.MakeCIDR(subnet, addrx))
	require.Error(t, err, "claiming address allocated on other peer should fail")
	// claiming the pre-claimed address should fail on both peers
	err = alloc.SimplyClaim(container3, preAddr)
	require.Error(t, err, "claiming address allocated on other peer should fail")
	err = allocs[0].SimplyClaim(container3, preAddr)
	require.Error(t, err, "claiming address allocated on other peer should fail")
	// Check an address outside of our universe
	addr2, _ := address.ParseCIDR(testAddr2)
	err = alloc.SimplyClaim(container1, addr2)
	require.NoError(t, err)
}

func (alloc *Allocator) pause() func() {
	paused := make(chan struct{})
	alloc.actionChan <- func() {
		paused <- struct{}{}
		<-paused
	}
	<-paused
	return func() {
		close(paused)
	}
}

func TestCancel(t *testing.T) {
	const cidr = "10.0.4.0/26"
	router := gossip.NewTestRouter(0.0)

	alloc1, subnet := makeAllocator("01:00:00:02:00:00", cidr, 2)
	alloc1.SetInterfaces(router.Connect(alloc1.ourName, alloc1))

	alloc2, _ := makeAllocator("02:00:00:02:00:00", cidr, 2)
	alloc2.SetInterfaces(router.Connect(alloc2.ourName, alloc2))
	alloc1.claimRingForTesting(alloc1, alloc2)
	alloc2.claimRingForTesting(alloc1, alloc2)

	alloc1.Start()
	alloc2.Start()

	// tell peers about each other
	alloc1.OnGossipBroadcast(alloc2.ourName, alloc2.Encode())

	// Get some IPs, so each allocator has some space
	res1, _ := alloc1.Allocate("foo", subnet, true, returnFalse)
	common.Log.Debugf("res1 = %s", res1.String())
	res2, _ := alloc2.Allocate("bar", subnet, true, returnFalse)
	common.Log.Debugf("res2 = %s", res2.String())
	if res1 == res2 {
		require.FailNow(t, "Error: got same ips!")
	}

	// Now we're going to pause alloc2 and ask alloc1
	// for an allocation
	unpause := alloc2.pause()

	// Use up all the IPs that alloc1 owns, so the allocation after this will prompt a request to alloc2
	for i := 0; alloc1.NumFreeAddresses(subnet.HostRange()) > 0; i++ {
		alloc1.Allocate(fmt.Sprintf("tmp%d", i), subnet, true, returnFalse)
	}
	cancelChan := make(chan bool, 1)
	doneChan := make(chan bool)
	go func() {
		_, ok := alloc1.Allocate("baz", subnet, true,
			func() bool {
				select {
				case <-cancelChan:
					return true
				default:
					return false
				}
			})
		doneChan <- ok == nil
	}()

	time.Sleep(100 * time.Millisecond)
	AssertNothingSent(t, doneChan)

	cancelChan <- true
	unpause()
	if <-doneChan {
		require.FailNow(t, "Error: got result from Allocate")
	}
}

func TestCancelOnDied(t *testing.T) {
	const (
		cidr       = "10.0.4.0/26"
		container1 = "abcdef"
	)

	router := gossip.NewTestRouter(0.0)
	alloc1, subnet := makeAllocator("01:00:00:02:00:00", cidr, 2)
	alloc1.SetInterfaces(router.Connect(alloc1.ourName, alloc1))
	alloc1.Start()

	doneChan := make(chan bool)
	f := func() {
		_, ok := alloc1.Allocate(container1, subnet, true, returnFalse)
		doneChan <- ok == nil
	}

	// Attempt two allocations in parallel, to check that this is handled correctly
	go f()
	go f()

	// Nothing should happen, because we declared the quorum as 2
	time.Sleep(100 * time.Millisecond)
	AssertNothingSent(t, doneChan)

	alloc1.ContainerDied(container1)

	// Check that the two allocations both exit with an error
	if <-doneChan || <-doneChan {
		require.FailNow(t, "Error: got result from Allocate")
	}
}

func TestGossipShutdown(t *testing.T) {
	const (
		container1 = "abcdef"
		container2 = "baddf00d"
		universe   = "10.0.3.0/30"
	)

	alloc, subnet := makeAllocatorWithMockGossip(t, "01:00:00:01:00:00", universe, 1)
	defer alloc.Stop()

	alloc.claimRingForTesting()
	alloc.SimplyAllocate(container1, subnet)

	alloc.Shutdown()

	_, err := alloc.SimplyAllocate(container2, subnet) // trying to allocate after shutdown should fail
	require.False(t, err == nil, "no address")

	CheckAllExpectedMessagesSent(alloc)
}

func TestNoFrag(t *testing.T) {
	const cidr = "10.0.4.0/22"
	resultChan := make(chan int)
	for i := 0; i < 100; i++ {
		allocs, router, subnet := makeNetworkOfAllocators(3, cidr)
		allocs[1].Allocate("foo", subnet, true, returnFalse)
		allocs[2].Allocate("bar", subnet, true, returnFalse)
		allocs[2].actionChan <- func() {
			resultChan <- len(allocs[2].ring.Entries)
		}
		require.True(t, <-resultChan < 6, "excessive ring fragmentation")
		stopNetworkOfAllocators(allocs, router)
	}
}

func TestTransfer(t *testing.T) {
	const cidr = "10.0.4.0/22"
	allocs, router, subnet := makeNetworkOfAllocators(3, cidr)
	defer stopNetworkOfAllocators(allocs, router)
	alloc0 := allocs[0]
	alloc1 := allocs[1]
	alloc2 := allocs[2]

	_, err := alloc1.Allocate("foo", subnet, true, returnFalse)
	require.True(t, err == nil, "Failed to get address")

	_, err = alloc2.Allocate("bar", subnet, true, returnFalse)
	require.True(t, err == nil, "Failed to get address")

	// simulation of periodic gossip
	alloc2.gossip.GossipBroadcast(alloc2.Gossip())
	router.Flush()
	alloc1.gossip.GossipBroadcast(alloc1.Gossip())
	router.Flush()

	free1 := alloc1.NumFreeAddresses(subnet.Range())
	free2 := alloc2.NumFreeAddresses(subnet.Range())

	router.RemovePeer(alloc1.ourName)
	router.RemovePeer(alloc2.ourName)
	alloc1.Stop()
	alloc2.Stop()
	router.Flush()

	require.Equal(t, free1+1, alloc0.AdminTakeoverRanges(alloc1.ourName.String()))
	require.Equal(t, free2+1, alloc0.AdminTakeoverRanges(alloc2.ourName.String()))
	router.Flush()

	require.Equal(t, address.Count(1024), alloc0.NumFreeAddresses(subnet.Range()))

	_, err = alloc0.Allocate("foo", subnet, true, returnFalse)
	require.True(t, err == nil, "Failed to get address")
	alloc0.Stop()
}

func TestFakeRouterSimple(t *testing.T) {
	const cidr = "10.0.4.0/22"
	allocs, router, subnet := makeNetworkOfAllocators(2, cidr)
	defer stopNetworkOfAllocators(allocs, router)

	alloc1 := allocs[0]
	//alloc2 := allocs[1]

	_, err := alloc1.Allocate("foo", subnet, true, returnFalse)
	require.NoError(t, err, "Failed to get address")
}

func TestAllocatorFuzz(t *testing.T) {
	const (
		firstpass    = 1000
		secondpass   = 10000
		nodes        = 5
		maxAddresses = 1000
		concurrency  = 5
		cidr         = "10.0.4.0/22"
	)
	allocs, router, subnet := makeNetworkOfAllocators(nodes, cidr)
	defer stopNetworkOfAllocators(allocs, router)

	// Test state
	// For each IP issued we store the allocator
	// that issued it and the name of the container
	// it was issued to.
	type result struct {
		name  string
		alloc int32
		block bool
	}
	stateLock := sync.Mutex{}
	state := make(map[string]result)
	// Keep a list of addresses issued, so we
	// Can pick random ones
	var addrs []string
	numPending := 0

	rand.Seed(0)

	// Remove item from list by swapping it with last
	// and reducing slice length by 1
	rm := func(xs []string, i int32) []string {
		ls := len(xs) - 1
		xs[i] = xs[ls]
		return xs[:ls]
	}

	bumpPending := func() bool {
		stateLock.Lock()
		if len(addrs)+numPending >= maxAddresses {
			stateLock.Unlock()
			return false
		}
		numPending++
		stateLock.Unlock()
		return true
	}

	noteAllocation := func(allocIndex int32, name string, addr address.Address) {
		//common.Log.Infof("Allocate: got address %s for name %s", addr, name)
		addrStr := addr.String()

		stateLock.Lock()
		defer stateLock.Unlock()

		if res, existing := state[addrStr]; existing {
			panic(fmt.Sprintf("Dup found for address %s - %s and %s", addrStr,
				name, res.name))
		}

		state[addrStr] = result{name, allocIndex, false}
		addrs = append(addrs, addrStr)
		numPending--
	}

	// Do a Allocate and check the address
	// is unique.  Needs a unique container
	// name.
	allocate := func(name string) {
		if !bumpPending() {
			return
		}

		allocIndex := rand.Int31n(nodes)
		alloc := allocs[allocIndex]
		//common.Log.Infof("Allocate: asking allocator %d", allocIndex)
		addr, err := alloc.SimplyAllocate(name, subnet)

		if err != nil {
			panic(fmt.Sprintf("Could not allocate addr"))
		}

		noteAllocation(allocIndex, name, addr)
	}

	// Free a random address.
	free := func() {
		stateLock.Lock()
		if len(addrs) == 0 {
			stateLock.Unlock()
			return
		}
		// Delete an existing allocation
		// Pick random addr
		addrIndex := rand.Int31n(int32(len(addrs)))
		addr := addrs[addrIndex]
		res := state[addr]
		if res.block {
			stateLock.Unlock()
			return
		}
		addrs = rm(addrs, addrIndex)
		delete(state, addr)
		stateLock.Unlock()

		alloc := allocs[res.alloc]
		//common.Log.Infof("Freeing %s (%s) on allocator %d", res.name, addr, res.alloc)

		oldAddr, err := address.ParseIP(addr)
		if err != nil {
			panic(err)
		}
		require.NoError(t, alloc.Free(res.name, oldAddr))
	}

	// Do a Allocate on an existing container & allocator
	// and check we get the right answer.
	allocateAgain := func() {
		stateLock.Lock()
		addrIndex := rand.Int31n(int32(len(addrs)))
		addr := addrs[addrIndex]
		res := state[addr]
		if res.block {
			stateLock.Unlock()
			return
		}
		res.block = true
		state[addr] = res
		stateLock.Unlock()
		alloc := allocs[res.alloc]

		//common.Log.Infof("Asking for %s (%s) on allocator %d again", res.name, addr, res.alloc)

		newAddr, _ := alloc.SimplyAllocate(res.name, subnet)
		oldAddr, _ := address.ParseIP(addr)
		if newAddr != oldAddr {
			panic(fmt.Sprintf("Got different address for repeat request for %s: %s != %s", res.name, newAddr, oldAddr))
		}

		stateLock.Lock()
		res.block = false
		state[addr] = res
		stateLock.Unlock()
	}

	// Claim a random address for a unique container name - may not succeed
	claim := func(name string) {
		if !bumpPending() {
			return
		}
		allocIndex := rand.Int31n(nodes)
		addressIndex := rand.Int31n(int32(subnet.Size()))
		alloc := allocs[allocIndex]
		addr := address.Add(subnet.Addr, address.Offset(addressIndex))
		err := alloc.SimplyClaim(name, address.MakeCIDR(subnet, addr))
		if err == nil {
			noteAllocation(allocIndex, name, addr)
		}
	}

	// Run function _f_ _iterations_ times, in _concurrency_
	// number of goroutines
	doConcurrentIterations := func(iterations int, f func(int)) {
		iterationsPerThread := iterations / concurrency

		wg := sync.WaitGroup{}
		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(j int) {
				defer wg.Done()
				for k := 0; k < iterationsPerThread; k++ {
					f((j * iterationsPerThread) + k)
				}
			}(i)
		}
		wg.Wait()
	}

	// First pass, just allocate a bunch of ips
	doConcurrentIterations(firstpass, func(iteration int) {
		name := fmt.Sprintf("first%d", iteration)
		allocate(name)
	})

	// Second pass, random ask for more allocations,
	// or remove existing ones, or ask for allocation
	// again.
	doConcurrentIterations(secondpass, func(iteration int) {
		r := rand.Float32()
		switch {
		case 0.0 <= r && r < 0.4:
			// Ask for a new allocation
			name := fmt.Sprintf("second%d", iteration)
			allocate(name)

		case (0.4 <= r && r < 0.8):
			// free a random addr
			free()

		case 0.8 <= r && r < 0.95:
			// ask for an existing name again, check we get same ip
			allocateAgain()

		case 0.95 <= r && r < 1.0:
			name := fmt.Sprintf("second%d", iteration)
			claim(name)
		}
	})
}

func TestSpaceRequest(t *testing.T) {
	const (
		container1 = "cont-1"
		container2 = "cont-2"
		universe   = "10.32.0.0/12"
	)
	allocs, router, subnet := makeNetworkOfAllocators(1, universe)
	defer stopNetworkOfAllocators(allocs, router)
	alloc1 := allocs[0]

	addr, err := alloc1.Allocate(container1, subnet, true, returnFalse)
	require.Nil(t, err, "")
	// free it again so the donation splits the range neatly
	err = alloc1.Free(container1, addr)
	require.Nil(t, err, "")
	require.Equal(t, cidrRanges(universe), alloc1.OwnedRanges(), "")

	// Start a new peer
	alloc2, _ := makeAllocator("02:00:00:02:00:00", universe, 2)
	alloc2.SetInterfaces(router.Connect(alloc2.ourName, alloc2))
	alloc2.Start()
	defer alloc2.Stop()
	alloc2.Allocate(container2, subnet, true, returnFalse)

	// Test whether the universe has been split into two equal halves (GH #2009)
	require.Equal(t, cidrRanges("10.32.0.0/13"), alloc1.OwnedRanges(), "")
	require.Equal(t, cidrRanges("10.40.0.0/13"), alloc2.OwnedRanges(), "")
}

func cidrRanges(s string) []address.Range {
	c, _ := address.ParseCIDR(s)
	return []address.Range{c.Range()}
}
