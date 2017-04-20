package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/coreos/go-iptables/iptables"
	"github.com/j-keck/arping"
	"github.com/vishvananda/netlink"

	weavenet "github.com/weaveworks/weave/net"
)

// checkIface returns an error if the given interface cannot be found.
func checkIface(args []string) error {
	if len(args) != 1 {
		cmdUsage("check-iface", "<iface-name>")
	}
	ifaceName := args[0]

	if _, err := netlink.LinkByName(ifaceName); err != nil {
		return err
	}

	return nil
}

func delIface(args []string) error {
	if len(args) != 1 {
		cmdUsage("del-iface", "<iface-name>")
	}
	ifName := args[0]

	link, err := netlink.LinkByName(ifName)
	if err != nil {
		return err
	}
	return netlink.LinkDel(link)
}

// setupIface renames the given iface and configures its ARP cache settings.
func setupIface(args []string) error {
	if len(args) != 2 {
		cmdUsage("setup-iface", "<iface-name> <new-iface-name>")
	}
	ifaceName := args[0]
	newIfName := args[1]

	ipt, err := iptables.New()
	if err != nil {
		return err
	}

	link, err := netlink.LinkByName(ifaceName)
	if err != nil {
		return err
	}
	if err := netlink.LinkSetName(link, newIfName); err != nil {
		return err
	}
	if err := weavenet.ConfigureARPCache(newIfName); err != nil {
		return err
	}
	if err := ipt.Append("filter", "INPUT", "-i", newIfName, "-d", "224.0.0.0/4", "-j", "DROP"); err != nil {
		return err
	}

	return nil
}

func configureARP(args []string) error {
	if len(args) != 1 {
		cmdUsage("configure-arp", "<iface-name-prefix>")
	}
	prefix := args[0]

	links, err := netlink.LinkList()
	if err != nil {
		return err
	}
	for _, link := range links {
		ifName := link.Attrs().Name
		if strings.HasPrefix(ifName, prefix) {
			weavenet.ConfigureARPCache(ifName)
			if addrs, err := netlink.AddrList(link, netlink.FAMILY_V4); err == nil {
				for _, addr := range addrs {
					arping.GratuitousArpOverIfaceByName(addr.IPNet.IP, ifName)
				}
			}
		}
	}

	return nil
}

// listNetDevs outputs network ifaces identified by the given indexes
// in the format of weavenet.Dev.
func listNetDevs(args []string) error {
	if len(args) == 0 {
		cmdUsage("list-netdevs", "<iface-index>[ <iface-index>]")
	}

	indexes := make(map[int]struct{})
	for _, index := range args {
		if index != "" {
			id, err := strconv.Atoi(index)
			if err != nil {
				return err
			}
			indexes[id] = struct{}{}
		}
	}

	links, err := netlink.LinkList()
	if err != nil {
		return err
	}

	var netdevs []weavenet.Dev

	for _, link := range links {
		if _, found := indexes[link.Attrs().Index]; found {
			netdev, err := weavenet.LinkToNetDev(link)
			if err != nil {
				return err
			}
			netdevs = append(netdevs, netdev)
		}
	}

	nds, err := json.Marshal(netdevs)
	if err != nil {
		return err
	}
	fmt.Println(string(nds))

	return nil
}
