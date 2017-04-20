package ipamplugin

import (
	"fmt"
	"net"
	"strings"

	"github.com/docker/libnetwork/discoverapi"
	"github.com/docker/libnetwork/netlabel"
	"github.com/weaveworks/weave/api"
	"github.com/weaveworks/weave/common"
)

type Ipam struct {
	weave *api.Client
}

func NewIpam(weave *api.Client) *Ipam {
	return &Ipam{weave: weave}
}

func (i *Ipam) GetCapabilities() (RequiresMACAddress bool, RequiresRequestReplay bool, err error) {
	i.logReq("GetCapabilities")
	return false, false, nil
}

func (i *Ipam) GetDefaultAddressSpaces() (string, string, error) {
	i.logReq("GetDefaultAddressSpaces")
	return "weavelocal", "weaveglobal", nil
}

func (i *Ipam) RequestPool(addressSpace, pool, subPool string, options map[string]string, v6 bool) (poolname string, subnet *net.IPNet, data map[string]string, err error) {
	i.logReq("RequestPool", addressSpace, pool, subPool, options)
	defer func() { i.logRes("RequestPool", err, poolname, subnet, data) }()
	if pool == "" {
		subnet, err = i.weave.DefaultSubnet()
	} else {
		_, subnet, err = net.ParseCIDR(pool)
	}
	if err != nil {
		return
	}
	iprange := subnet
	if subPool != "" {
		if _, iprange, err = net.ParseCIDR(subPool); err != nil {
			return
		}
	}
	// Cunningly-constructed pool "name" which gives us what we need later
	poolname = strings.Join([]string{"weave", subnet.String(), iprange.String()}, "-")
	// Pass back a fake "gateway address"; we don't actually use it,
	// so just give the network address.
	data = map[string]string{netlabel.Gateway: subnet.String()}
	return
}

func (i *Ipam) ReleasePool(poolID string) error {
	i.logReq("ReleasePool", poolID)
	return nil
}

func splitPoolID(poolID string) (subnet, iprange *net.IPNet, err error) {
	parts := strings.Split(poolID, "-")
	if len(parts) != 3 || parts[0] != "weave" {
		err = fmt.Errorf("Unrecognized pool ID: %s", poolID)
		return
	}
	if _, subnet, err = net.ParseCIDR(parts[1]); err != nil {
		return
	}
	if _, iprange, err = net.ParseCIDR(parts[2]); err != nil {
		return
	}
	return
}

func (i *Ipam) RequestAddress(poolID string, address net.IP, options map[string]string) (ip *net.IPNet, _ map[string]string, err error) {
	i.logReq("RequestAddress", poolID, address, options)
	defer func() { i.logRes("RequestAddress", err, ip) }()
	if poolID == "weavepool" { // old-style
		ip, err = i.weave.AllocateIP(api.NoContainerID)
		return
	}
	subnet, iprange, err := splitPoolID(poolID)
	if err != nil {
		return
	}
	if address != nil { // try to claim specific address requested
		ip = &net.IPNet{IP: address, Mask: subnet.Mask}
		if err = i.weave.ClaimIP(api.NoContainerID, ip); err != nil {
			return
		}
	} else {
		// We are lying slightly to IPAM here: the range is not a subnet
		if ip, err = i.weave.AllocateIPInSubnet(api.NoContainerID, iprange); err != nil {
			return
		}
		ip.Mask = subnet.Mask // fix up the subnet we lied about
	}
	return
}

func (i *Ipam) ReleaseAddress(poolID string, address net.IP) error {
	i.logReq("ReleaseAddress", poolID, address)
	if subnet, _, err := splitPoolID(poolID); err != nil {
		return err
	} else if address.Equal(subnet.IP) { // is it the gateway address we faked earlier?
		return nil
	}
	return i.weave.ReleaseIPsFor(address.String())
}

// Functions required by ipamapi "contract" but not actually used.

func (i *Ipam) DiscoverNew(discoverapi.DiscoveryType, interface{}) error {
	return nil
}

func (i *Ipam) DiscoverDelete(discoverapi.DiscoveryType, interface{}) error {
	return nil
}

// logging

func (i *Ipam) logReq(fun string, args ...interface{}) {
	common.Log.Infoln(append([]interface{}{fmt.Sprintf("[ipam] %s", fun)}, args...)...)
}

func (i *Ipam) logRes(fun string, err error, args ...interface{}) {
	if err == nil {
		common.Log.Debugln(append([]interface{}{fmt.Sprintf("[ipam] %s result", fun)}, args...)...)
		return
	}
	common.Log.Errorf("[ipam] %s: %s", fun, err)
}
