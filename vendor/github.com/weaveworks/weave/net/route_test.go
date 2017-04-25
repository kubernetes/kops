package net

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOverlap(t *testing.T) {
	_, subnet1, _ := net.ParseCIDR("10.0.3.0/24")
	_, subnet2, _ := net.ParseCIDR("10.0.4.0/24")
	_, subnet3, _ := net.ParseCIDR("10.0.3.128/25")
	_, subnet4, _ := net.ParseCIDR("10.0.3.192/25")
	_, universe, _ := net.ParseCIDR("10.0.0.0/8")
	require.False(t, overlaps(subnet1, subnet2))
	require.False(t, overlaps(subnet2, subnet1))
	require.True(t, overlaps(subnet1, subnet1))
	require.True(t, overlaps(subnet1, subnet3))
	require.True(t, overlaps(subnet1, subnet4))
	require.False(t, overlaps(subnet2, subnet4))
	require.False(t, overlaps(subnet4, subnet2))
	require.True(t, overlaps(universe, subnet1))
	require.True(t, overlaps(subnet1, universe))
}
