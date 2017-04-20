package net

import (
	"fmt"
	"net"
	"syscall"

	"github.com/vishvananda/netlink"
)

// Wait for an interface to come up.
func EnsureInterface(ifaceName string) (*net.Interface, error) {
	iface, err := ensureInterface(ifaceName)
	if err != nil {
		return nil, err
	}
	return iface, err
}

func ensureInterface(ifaceName string) (*net.Interface, error) {
	ch := make(chan netlink.LinkUpdate)
	// NB: We do not supply (and eventually close) a 'done' channel
	// here since that can cause incorrect file descriptor
	// re-usage. See https://github.com/weaveworks/weave/issues/2120
	if err := netlink.LinkSubscribe(ch, nil); err != nil {
		return nil, err
	}
	// check for currently-existing interface after subscribing, to avoid race
	if iface, err := findInterface(ifaceName); err == nil {
		return iface, nil
	}
	for update := range ch {
		if ifaceName == update.Link.Attrs().Name && update.IfInfomsg.Flags&syscall.IFF_UP != 0 {
			break
		}
	}
	iface, err := findInterface(ifaceName)
	return iface, err
}

// Wait for an interface to come up and have a route added to the multicast subnet.
// This matches the behaviour in 'weave attach', which is the only context in which
// we expect this to be called.  If you change one, change the other to match.
func EnsureInterfaceAndMcastRoute(ifaceName string) (*net.Interface, error) {
	iface, err := ensureInterface(ifaceName)
	if err != nil {
		return nil, err
	}
	ch := make(chan netlink.RouteUpdate)
	if err := netlink.RouteSubscribe(ch, nil); err != nil {
		return nil, err
	}
	dest := net.IPv4(224, 0, 0, 0)
	check := func(route netlink.Route) bool {
		return route.Dst != nil && route.Dst.IP.Equal(dest)
	}
	// check for currently-existing route after subscribing, to avoid race
	routes, err := netlink.RouteList(nil, netlink.FAMILY_V4)
	if err != nil {
		return nil, err
	}
	for _, route := range routes {
		if check(route) {
			return iface, nil
		}
	}
	for update := range ch {
		if check(update.Route) {
			return iface, nil
		}
	}
	// should never get here
	return iface, nil
}

func findInterface(ifaceName string) (iface *net.Interface, err error) {
	if iface, err = net.InterfaceByName(ifaceName); err != nil {
		return iface, fmt.Errorf("Unable to find interface %s", ifaceName)
	}
	if 0 == (net.FlagUp & iface.Flags) {
		return iface, fmt.Errorf("Interface %s is not up", ifaceName)
	}
	return
}

func LocalAddresses() ([]*net.IPNet, error) {
	links, err := netlink.LinkList()
	if err != nil {
		return nil, err
	}

	retval := []*net.IPNet{}
	for _, link := range links {
		if _, isBridge := link.(*netlink.Bridge); isBridge || (link.Attrs().Flags&net.FlagLoopback) != 0 {
			continue
		}
		addrs, err := netlink.AddrList(link, netlink.FAMILY_V4)
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			retval = append(retval, addr.IPNet)
		}
	}
	return retval, nil
}
