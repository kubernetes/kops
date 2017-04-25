package main

import (
	"strconv"

	"github.com/weaveworks/weave/common"
	weavenet "github.com/weaveworks/weave/net"
)

func processAddrs(args []string) error {
	if len(args) < 1 {
		cmdUsage("process-addrs", "<bridgeName>")
	}
	bridgeName := args[0]

	peerIDs, err := weavenet.ConnectedToBridgeVethPeerIds(bridgeName)
	if err != nil {
		if err == weavenet.ErrLinkNotFound {
			return nil
		}
		return err
	}

	pids, err := common.AllPids("/proc")
	if err != nil {
		return err
	}

	for _, pid := range pids {
		netDevs, err := weavenet.GetNetDevsByVethPeerIds(pid, peerIDs)
		if err != nil {
			return err
		}
		printNetDevs(strconv.Itoa(pid), netDevs)
	}
	return nil
}
