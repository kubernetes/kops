package net

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

type Dev struct {
	Name  string           `json:"Name,omitempty"`
	MAC   net.HardwareAddr `json:"MAC,omitempty"`
	CIDRs []*net.IPNet     `json:"CIDRs,omitempty"`
}

func LinkToNetDev(link netlink.Link) (Dev, error) {
	addrs, err := netlink.AddrList(link, netlink.FAMILY_V4)
	if err != nil {
		return Dev{}, err
	}

	netDev := Dev{Name: link.Attrs().Name, MAC: link.Attrs().HardwareAddr}
	for _, addr := range addrs {
		netDev.CIDRs = append(netDev.CIDRs, addr.IPNet)
	}
	return netDev, nil
}

// ConnectedToBridgeVethPeerIds returns peer indexes of veth links connected to
// the given bridge. The peer index is used to query from a container netns
// whether the container is connected to the bridge.
func ConnectedToBridgeVethPeerIds(bridgeName string) ([]int, error) {
	var ids []int

	br, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return nil, err
	}
	links, err := netlink.LinkList()
	if err != nil {
		return nil, err
	}

	for _, link := range links {
		if _, isveth := link.(*netlink.Veth); isveth && link.Attrs().MasterIndex == br.Attrs().Index {
			peerID := link.Attrs().ParentIndex
			if peerID == 0 {
				// perhaps running on an older kernel where ParentIndex doesn't work.
				// as fall-back, assume the peers are consecutive
				peerID = link.Attrs().Index - 1
			}
			ids = append(ids, peerID)
		}
	}

	return ids, nil
}

// Lookup the weave interface of a container
func GetWeaveNetDevs(processID int) ([]Dev, error) {
	peerIDs, err := ConnectedToBridgeVethPeerIds("weave")
	if err != nil {
		return nil, err
	}

	return GetNetDevsByVethPeerIds(processID, peerIDs)
}

func GetNetDevsByVethPeerIds(processID int, peerIDs []int) ([]Dev, error) {
	// Bail out if this container is running in the root namespace
	netnsRoot, err := netns.GetFromPid(1)
	if err != nil {
		return nil, fmt.Errorf("unable to open root namespace: %s", err)
	}
	defer netnsRoot.Close()
	netnsContainer, err := netns.GetFromPid(processID)
	if err != nil {
		// Unable to find a namespace for this process - just return nothing
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("unable to open process %d namespace: %s", processID, err)
	}
	defer netnsContainer.Close()
	if netnsRoot.Equal(netnsContainer) {
		return nil, nil
	}

	var netdevs []Dev
	peersStr := make([]string, len(peerIDs))
	for i, id := range peerIDs {
		peersStr[i] = strconv.Itoa(id)
	}
	nds, err := WithNetNSByPid(processID, "list-netdevs", peersStr...)
	if err != nil {
		return nil, fmt.Errorf("list-netdevs failed: %s", err)
	}
	err = json.Unmarshal(nds, &netdevs)

	return netdevs, err
}

// Get the weave bridge interface.
// NB: Should be called from the root network namespace.
func GetBridgeNetDev(bridgeName string) (Dev, error) {
	link, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return Dev{}, err
	}

	return LinkToNetDev(link)
}
