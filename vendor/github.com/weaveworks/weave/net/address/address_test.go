package address

import (
	"testing"
	"testing/quick"

	"github.com/stretchr/testify/require"
)

func isPower2(x Count) bool {
	if x == 0 {
		return false
	}
	for ; x > 1; x /= 2 {
		if x&1 != 0 {
			return false
		}
	}
	return true
}

func TestBiggestPow2AlignedRange(t *testing.T) {
	require.Equal(t, NewRange(0, 1), NewRange(0, 1).BiggestCIDRRange())
	require.Equal(t, NewRange(1, 1), NewRange(1, 2).BiggestCIDRRange())
	require.Equal(t, NewRange(2, 2), NewRange(1, 3).BiggestCIDRRange())
	require.Equal(t, NewRange(0, 0x40000000), NewRange(0, 0x7fffffff).BiggestCIDRRange())
	require.Equal(t, NewRange(0xfffffffe, 1), NewRange(0xfffffffe, 1).BiggestCIDRRange())
	prop := func(start Address, size Offset) bool {
		if size > Offset(0xffffffff)-Offset(start) { // out of range
			return true
		}
		r := NewRange(start, size)
		result := r.BiggestCIDRRange()
		return r.Contains(result.Start) &&
			r.Contains(result.End-1) &&
			isPower2(result.Size()) &&
			result.Size() > r.Size()/4 &&
			Count(result.Start)%result.Size() == 0
	}
	require.NoError(t, quick.Check(prop, &quick.Config{MaxCount: 1000000}))
}

func ip(s string) Address {
	addr, _ := ParseIP(s)
	return addr
}

func cidr(s string) CIDR {
	c, err := ParseCIDR(s)
	if err != nil {
		panic(err)
	}
	return c
}

func TestCIDRs(t *testing.T) {
	start := ip("10.0.0.1")
	end := ip("10.0.0.9")
	r := NewRange(start, Subtract(end, start))
	require.Equal(t,
		[]CIDR{cidr("10.0.0.1/32"), cidr("10.0.0.2/31"), cidr("10.0.0.4/30"), cidr("10.0.0.8/32")},
		r.CIDRs())
}

func TestCIDRStartAndEnd(t *testing.T) {
	cidr, _ := ParseCIDR("10.0.0.0/24")
	require.Equal(t, ip("10.0.0.0"), cidr.Start(), "")
	require.Equal(t, ip("10.0.1.0"), cidr.End(), "")
}
