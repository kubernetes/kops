package net

import (
	"fmt"
	"net"
	"os"

	"github.com/vishvananda/netlink"
)

// A network is considered free if it does not overlap any existing
// routes on this host. This is the same approach taken by Docker.
func CheckNetworkFree(subnet *net.IPNet, ignoreIfaceNames map[string]struct{}) error {
	return forEachRoute(ignoreIfaceNames, func(route netlink.Route) error {
		if route.Dst != nil && overlaps(route.Dst, subnet) {
			return fmt.Errorf("Network %s overlaps with existing route %s on host", subnet, route.Dst)
		}
		return nil
	})
}

// Two networks overlap if the start-point of one is inside the other.
func overlaps(n1, n2 *net.IPNet) bool {
	return n1.Contains(n2.IP) || n2.Contains(n1.IP)
}

// For a specific address, we only care if it is actually *inside* an
// existing route, because weave-local traffic never hits IP routing.
func CheckAddressOverlap(addr net.IP, ignoreIfaceNames map[string]struct{}) error {
	return forEachRoute(ignoreIfaceNames, func(route netlink.Route) error {
		if route.Dst != nil && route.Dst.Contains(addr) {
			return fmt.Errorf("Address %s overlaps with existing route %s on host", addr, route.Dst)
		}
		return nil
	})
}

func forEachRoute(ignoreIfaceNames map[string]struct{}, check func(r netlink.Route) error) error {
	ignoreIfaceIndices := make(map[int]struct{})
	for ifaceName := range ignoreIfaceNames {
		if iface, err := net.InterfaceByName(ifaceName); err == nil {
			ignoreIfaceIndices[iface.Index] = struct{}{}
		}
	}
	routes, err := netlink.RouteList(nil, netlink.FAMILY_V4)
	if err != nil {
		return err
	}
	for _, route := range routes {
		if _, found := ignoreIfaceIndices[route.LinkIndex]; found {
			continue
		}
		if err := check(route); err != nil {
			return err
		}
	}
	return nil
}

func AddRoute(link netlink.Link, scope netlink.Scope, dst *net.IPNet, gw net.IP) error {
	err := netlink.RouteAdd(&netlink.Route{
		LinkIndex: link.Attrs().Index,
		Scope:     scope,
		Dst:       dst,
		Gw:        gw,
	})
	if os.IsExist(err) { // squash duplicate route errors
		err = nil
	}
	return err
}
