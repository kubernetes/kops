package space

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/weaveworks/weave/net/address"
	wt "github.com/weaveworks/weave/testing"
)

func makeSpace(start address.Address, size address.Offset) *Space {
	s := New()
	s.Add(start, size)
	return s
}

func ip(s string) address.Address {
	addr, _ := address.ParseIP(s)
	return addr
}

func TestLowlevel(t *testing.T) {
	a := []address.Address{}
	a = add(a, 100, 200)
	require.Equal(t, []address.Address{100, 200}, a)
	require.True(t, !contains(a, 99), "")
	require.True(t, contains(a, 100), "")
	require.True(t, contains(a, 199), "")
	require.True(t, !contains(a, 200), "")
	a = add(a, 700, 800)
	require.Equal(t, []address.Address{100, 200, 700, 800}, a)
	a = add(a, 300, 400)
	require.Equal(t, []address.Address{100, 200, 300, 400, 700, 800}, a)
	a = add(a, 400, 500)
	require.Equal(t, []address.Address{100, 200, 300, 500, 700, 800}, a)
	a = add(a, 600, 700)
	require.Equal(t, []address.Address{100, 200, 300, 500, 600, 800}, a)
	a = add(a, 500, 600)
	require.Equal(t, []address.Address{100, 200, 300, 800}, a)
	a = subtract(a, 500, 600)
	require.Equal(t, []address.Address{100, 200, 300, 500, 600, 800}, a)
	a = subtract(a, 600, 700)
	require.Equal(t, []address.Address{100, 200, 300, 500, 700, 800}, a)
	a = subtract(a, 400, 500)
	require.Equal(t, []address.Address{100, 200, 300, 400, 700, 800}, a)
	a = subtract(a, 300, 400)
	require.Equal(t, []address.Address{100, 200, 700, 800}, a)
	a = subtract(a, 700, 800)
	require.Equal(t, []address.Address{100, 200}, a)
	a = subtract(a, 100, 200)
	require.Equal(t, []address.Address{}, a)

	s := New()
	require.Equal(t, address.Count(0), s.NumFreeAddresses())
	ok, _ := s.Allocate(address.NewRange(0, 1000))
	require.False(t, ok, "allocate in empty space should fail")

	s.Add(100, 100)
	require.Equal(t, address.Count(100), s.NumFreeAddresses())
	ok, got := s.Allocate(address.NewRange(0, 1000))
	require.True(t, ok && got == 100, "allocate")
	require.Equal(t, address.Count(99), s.NumFreeAddresses())
	require.NoError(t, s.Claim(150))
	require.Equal(t, address.Count(98), s.NumFreeAddresses())
	require.NoError(t, s.Free(100))
	require.Equal(t, address.Count(99), s.NumFreeAddresses())
	wt.AssertErrorInterface(t, (*error)(nil), s.Free(0), "free not allocated")
	wt.AssertErrorInterface(t, (*error)(nil), s.Free(100), "double free")

	r, ok := s.Donate(address.NewRange(0, 1000))
	require.True(t, ok, "donate")
	require.Equal(t, address.NewRange(0xa0, 0x20), r, "donate")

	// test Donate when addresses are scarce
	s = New()
	r, ok = s.Donate(address.NewRange(0, 1000))
	require.True(t, !ok, "donate on empty space should fail")
	s.Add(0, 3)
	require.NoError(t, s.Claim(0))
	require.NoError(t, s.Claim(2))
	r, ok = s.Donate(address.NewRange(0, 1000))
	require.True(t, ok, "donate")
	require.Equal(t, address.NewRange(1, 1), r, "donate")
	r, ok = s.Donate(address.NewRange(0, 1000))
	require.True(t, !ok, "donate should fail")
}

func TestSpaceAllocate(t *testing.T) {
	const (
		testAddr1   = "10.0.3.4"
		testAddr2   = "10.0.3.5"
		testAddrx   = "10.0.3.19"
		testAddry   = "10.0.9.19"
		containerID = "deadbeef"
		size        = 20
	)
	var (
		start = ip(testAddr1)
	)

	space1 := makeSpace(start, size)
	require.Equal(t, address.Count(20), space1.NumFreeAddresses())
	require.Equal(t, address.Count(0), space1.NumOwnedAddresses())
	space1.assertInvariants()

	_, addr1 := space1.Allocate(address.NewRange(start, size))
	require.Equal(t, testAddr1, addr1.String(), "address")
	require.Equal(t, address.Count(19), space1.NumFreeAddresses())
	require.Equal(t, address.Count(1), space1.NumOwnedAddresses())
	space1.assertInvariants()

	_, addr2 := space1.Allocate(address.NewRange(start, size))
	require.False(t, addr2.String() == testAddr1, "address")
	require.Equal(t, address.Count(18), space1.NumFreeAddresses())
	require.Equal(t, address.Count(2), space1.NumOwnedAddresses())
	require.Equal(t, address.Count(13), space1.NumFreeAddressesInRange(address.Range{Start: ip(testAddr1), End: ip(testAddrx)}))
	require.Equal(t, address.Count(18), space1.NumFreeAddressesInRange(address.Range{Start: ip(testAddr1), End: ip(testAddry)}))
	space1.assertInvariants()

	space1.Free(addr2)
	space1.assertInvariants()

	wt.AssertErrorInterface(t, (*error)(nil), space1.Free(addr2), "double free")
	wt.AssertErrorInterface(t, (*error)(nil), space1.Free(ip(testAddrx)), "address not allocated")
	wt.AssertErrorInterface(t, (*error)(nil), space1.Free(ip(testAddry)), "wrong out of range")

	space1.assertInvariants()
}

func TestSpaceFree(t *testing.T) {
	const (
		testAddr1   = "10.0.3.16"
		testAddrx   = "10.0.3.19"
		testAddry   = "10.0.9.19"
		containerID = "deadbeef"
		size        = 16
	)

	entireRange := address.NewRange(ip(testAddr1), size)
	space := makeSpace(ip(testAddr1), size)

	// Check we are prepared to give up the entire space
	r := space.biggestFreeRange(entireRange)
	require.Equal(t, address.NewRange(ip(testAddr1), size), r, "Wrong space")

	for i := 0; i < size; i++ {
		ok, _ := space.Allocate(entireRange)
		require.True(t, ok, "Failed to get address")
	}

	// Check we are full
	ok, _ := space.Allocate(entireRange)
	require.True(t, !ok, "Should have failed to get address")
	r, _ = space.Donate(entireRange)
	require.True(t, r.Size() == 0, "Wrong space")

	// Free in the middle
	require.NoError(t, space.Free(ip("10.0.3.23")))
	r = space.biggestFreeRange(entireRange)
	require.True(t, r.Start == ip("10.0.3.23") && r.Size() == 1, "Wrong space")

	// Free one at the end
	require.NoError(t, space.Free(ip("10.0.3.31")))
	r = space.biggestFreeRange(entireRange)
	require.True(t, r.Start == ip("10.0.3.31") && r.Size() == 1, "Wrong space")

	// Now free a few at the end
	require.NoError(t, space.Free(ip("10.0.3.30")))
	require.NoError(t, space.Free(ip("10.0.3.29")))

	require.Equal(t, address.Count(4), space.NumFreeAddresses())

	// Now get the biggest free space; should be 3.30
	r = space.biggestFreeRange(entireRange)
	require.Equal(t, address.NewRange(ip("10.0.3.30"), 2), r, "Wrong space")

	// Now free a few in the middle
	require.NoError(t, space.Free(ip("10.0.3.24")))
	require.NoError(t, space.Free(ip("10.0.3.22")))
	require.NoError(t, space.Free(ip("10.0.3.21")))

	require.Equal(t, address.Count(7), space.NumFreeAddresses())

	// Now get the biggest free space; should be 3.30
	r = space.biggestFreeRange(entireRange)
	require.Equal(t, address.NewRange(ip("10.0.3.30"), 2), r, "Wrong space")

	require.Equal(t, []address.Range{{Start: ip("10.0.3.16"), End: ip("10.0.3.32")}}, space.OwnedRanges())
}

func TestDonateSimple(t *testing.T) {
	const (
		testAddr1 = "10.0.1.0"
		testAddr2 = "10.0.1.32"
		size      = 64
	)

	var (
		ipAddr1 = ip(testAddr1)
	)

	ps1 := makeSpace(ipAddr1, size)

	// Empty space set should split in two and give me the second half
	r, ok := ps1.Donate(address.NewRange(ip(testAddr1), size))
	numGivenUp := r.Size()
	require.True(t, ok, "Donate result")
	require.Equal(t, "10.0.1.32", r.Start.String(), "Invalid start")
	require.Equal(t, address.Count(size/2), numGivenUp)
	require.Equal(t, address.Count(size/2), ps1.NumFreeAddresses())

	// Now check we can give the rest up.
	count := 0 // count to avoid infinite loop
	for ; count < 1000; count++ {
		r, ok := ps1.Donate(address.NewRange(ip(testAddr1), size))
		if !ok {
			break
		}
		numGivenUp += r.Size()
	}
	require.Equal(t, address.Count(0), ps1.NumFreeAddresses())
	require.Equal(t, address.Count(size), numGivenUp)
}

func TestDonateHard(t *testing.T) {
	//common.InitDefaultLogging(true)
	var (
		start                = ip("10.0.1.0")
		size  address.Offset = 48
	)

	// Fill a fresh space
	spaceset := makeSpace(start, size)
	for i := address.Offset(0); i < size; i++ {
		ok, _ := spaceset.Allocate(address.NewRange(start, size))
		require.True(t, ok, "Failed to get IP!")
	}

	require.Equal(t, address.Count(0), spaceset.NumFreeAddresses())

	// Now free all but the last address
	// this will force us to split the free list
	for i := address.Offset(0); i < size-1; i++ {
		require.NoError(t, spaceset.Free(address.Add(start, i)))
	}

	// Now split
	newRange, ok := spaceset.Donate(address.NewRange(start, size))
	require.True(t, ok, "GiveUpSpace result")
	require.Equal(t, address.NewRange(ip("10.0.1.16"), 16), newRange, "Wrong space")
	require.Equal(t, address.Count(31), spaceset.NumFreeAddresses())

	//Space set should now have 3 spaces
	expected := &Space{
		ours: add(nil, ip("10.0.1.47"), ip("10.0.1.48")),
		free: add(add(nil, ip("10.0.1.0"), ip("10.0.1.16")), ip("10.0.1.32"), ip("10.0.1.47")),
	}
	require.Equal(t, expected, spaceset)
}
