package tracker

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/weaveworks/weave/net/address"
)

var (
	r0to127     = cidr("10.0.0.0", "10.0.0.127")
	r128to255   = cidr("10.0.0.128", "10.0.0.255")
	r0to255     = cidr("10.0.0.0", "10.0.0.255")
	r1dot0to255 = cidr("10.0.1.0", "10.0.1.255")
	r2dot0to255 = cidr("10.0.2.0", "10.0.2.255")
)

func TestRemoveCommon(t *testing.T) {
	a := []address.CIDR{r0to127, r1dot0to255}
	b := []address.CIDR{r1dot0to255, r2dot0to255}
	newA, newB := removeCommon(a, b)
	require.Equal(t, []address.CIDR{r0to127}, newA)
	require.Equal(t, []address.CIDR{r2dot0to255}, newB)
}

func TestMerge(t *testing.T) {
	ranges := []address.Range{
		r0to127.Range(),
		r128to255.Range(),
		r2dot0to255.Range(),
	}
	require.Equal(t, []address.Range{r0to255.Range(), r2dot0to255.Range()}, merge(ranges))
}

// Helper

// TODO(mp) DRY with helpers of other tests.

func ip(s string) address.Address {
	addr, _ := address.ParseIP(s)
	return addr
}

// [start; end]
func cidr(start, end string) address.CIDR {
	c := address.Range{Start: ip(start), End: ip(end) + 1}.CIDRs()
	if len(c) != 1 {
		panic("invalid cidr")
	}
	return c[0]
}
