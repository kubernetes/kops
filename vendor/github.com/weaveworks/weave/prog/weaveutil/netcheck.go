/* netcheck: check whether a given network or address overlaps with any existing routes */
package main

import (
	"net"

	weavenet "github.com/weaveworks/weave/net"
)

func netcheck(args []string) error {
	if len(args) < 1 {
		cmdUsage("netcheck", "<cidr> [<interface-to-ignore> ...]")
	}
	addr, ipnet, err := net.ParseCIDR(args[0])
	if err != nil {
		return err
	}
	ignoreIfaceNames := make(map[string]struct{})
	for _, ifName := range args[1:] {
		ignoreIfaceNames[ifName] = struct{}{}
	}
	if ipnet.IP.Equal(addr) {
		err = weavenet.CheckNetworkFree(ipnet, ignoreIfaceNames)
	} else {
		err = weavenet.CheckAddressOverlap(addr, ignoreIfaceNames)
	}
	return err
}
