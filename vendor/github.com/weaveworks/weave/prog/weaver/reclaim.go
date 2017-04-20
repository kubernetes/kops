package main

import (
	"net"

	docker "github.com/fsouza/go-dockerclient"

	"github.com/weaveworks/weave/api"
	"github.com/weaveworks/weave/common"
	weavedocker "github.com/weaveworks/weave/common/docker"
	"github.com/weaveworks/weave/ipam"
	weavenet "github.com/weaveworks/weave/net"
	"github.com/weaveworks/weave/net/address"
)

func a(cidr *net.IPNet) address.CIDR {
	prefixLength, _ := cidr.Mask.Size()
	return address.CIDR{Addr: address.FromIP4(cidr.IP), PrefixLen: prefixLength}
}

// Get all the existing Weave IPs at startup, so we can stop IPAM
// giving out any as duplicates
func findExistingAddresses(dockerCli *weavedocker.Client, bridgeName string) (addrs []ipam.PreClaim, err error) {
	Log.Infof("Checking for pre-existing addresses on %s bridge", bridgeName)
	// First get the address for the bridge
	bridgeNetDev, err := weavenet.GetBridgeNetDev(bridgeName)
	if err != nil {
		return nil, err
	}
	for _, cidr := range bridgeNetDev.CIDRs {
		Log.Infof("%s bridge has address %v", bridgeName, cidr)
		addrs = append(addrs, ipam.PreClaim{Ident: "weave:expose", Cidr: a(cidr)})
	}

	add := func(cid string, isContainer bool, netDevs []weavenet.Dev) {
		for _, netDev := range netDevs {
			for _, cidr := range netDev.CIDRs {
				Log.Infof("Found address %v for ID %s", cidr, cid)
				addrs = append(addrs, ipam.PreClaim{Ident: cid, IsContainer: isContainer, Cidr: a(cidr)})
			}
		}
	}

	// Then find all veths connected to the bridge
	peerIDs, err := weavenet.ConnectedToBridgeVethPeerIds(bridgeName)
	if err != nil {
		return nil, err
	}

	// Now iterate over all containers to see if they have a network
	// namespace with an attached interface
	if dockerCli != nil {
		containerIDs, err := dockerCli.RunningContainerIDs()
		if err != nil {
			return nil, err
		}

		for _, cid := range containerIDs {
			container, err := dockerCli.InspectContainer(cid)
			if err != nil {
				if _, ok := err.(*docker.NoSuchContainer); ok {
					continue
				}
				return nil, err
			}
			if container.State.Pid != 0 {
				netDevs, err := weavenet.GetNetDevsByVethPeerIds(container.State.Pid, peerIDs)
				if err != nil {
					return nil, err
				}
				add(cid, true, netDevs)
			}
		}
	} else {
		// If we don't have a Docker connection, iterate over all processes
		pids, err := common.AllPids("/proc")
		if err != nil {
			return nil, err
		}
		for _, pid := range pids {
			netDevs, err := weavenet.GetNetDevsByVethPeerIds(pid, peerIDs)
			if err != nil {
				return nil, err
			}
			add(api.NoContainerID, false, netDevs)
		}
	}
	return addrs, nil
}
