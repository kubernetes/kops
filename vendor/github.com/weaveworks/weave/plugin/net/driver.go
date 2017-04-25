package plugin

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/docker/libnetwork/drivers/remote/api"
	"github.com/docker/libnetwork/netlabel"
	"github.com/docker/libnetwork/types"
	"golang.org/x/sys/unix"

	"github.com/vishvananda/netlink"
	weaveapi "github.com/weaveworks/weave/api"
	"github.com/weaveworks/weave/common"
	"github.com/weaveworks/weave/common/docker"
	weavenet "github.com/weaveworks/weave/net"
	"github.com/weaveworks/weave/plugin/skel"
)

const (
	MulticastOption = "works.weave.multicast"
)

type network struct {
	isOurs            bool
	hasMulticastRoute bool
}

type driver struct {
	sync.RWMutex
	scope        string
	docker       *docker.Client
	dns          bool
	isPluginV2   bool
	networks     map[string]network
	isNetworkOur func(driverName string) bool
}

func New(client *docker.Client, weave *weaveapi.Client, scope string, dns, isPluginV2 bool, isNetworkOur func(string) bool) (skel.Driver, error) {
	driver := &driver{
		scope:        scope,
		docker:       client,
		dns:          dns,
		isPluginV2:   isPluginV2,
		networks:     make(map[string]network),
		isNetworkOur: isNetworkOur,
	}

	_, err := NewWatcher(client, weave, driver)
	if err != nil {
		return nil, err
	}
	return driver, nil
}

// === protocol handlers

func (driver *driver) GetCapabilities() (*api.GetCapabilityResponse, error) {
	driver.logReq("GetCapabilities", nil, "")
	var caps = &api.GetCapabilityResponse{
		Scope: driver.scope,
	}
	driver.logRes("GetCapabilities", caps)
	return caps, nil
}

// In Swarm mode, CreateNetwork is called on each Swarm node when a new service
// is created.
func (driver *driver) CreateNetwork(create *api.CreateNetworkRequest) error {
	driver.logReq("CreateNetwork", create, create.NetworkID)
	_, err := driver.setupNetworkInfo(create.NetworkID, true, stringOptions(create))
	return err
}

// NetworkAllocate is called on a Swarm node (master) which creates the network.
// The returned options are passed to CreateNetwork.
func (driver *driver) NetworkAllocate(alloc *api.AllocateNetworkRequest) (*api.AllocateNetworkResponse, error) {
	driver.logReq("NetworkAllocate", alloc, alloc.NetworkID)
	return &api.AllocateNetworkResponse{Options: alloc.Options}, nil
}

// NetworkFree is called on a Swarm master node which created the network.
func (driver *driver) NetworkFree(free *api.FreeNetworkRequest) (*api.FreeNetworkResponse, error) {
	driver.logReq("NetworkFree", free, free.NetworkID)
	return nil, nil
}

// Deal with excessively-generic way the options get decoded from JSON
func stringOptions(create *api.CreateNetworkRequest) map[string]string {
	if create.Options != nil {
		if data, found := create.Options[netlabel.GenericData]; found {
			if options, ok := data.(map[string]interface{}); ok {
				out := make(map[string]string, len(options))
				for key, value := range options {
					if str, ok := value.(string); ok {
						out[key] = str
					}
				}
				return out
			}
		}
	}
	return nil
}

// In Swarm mode, DeleteNetwork is called after a service has been removed.
func (driver *driver) DeleteNetwork(delreq *api.DeleteNetworkRequest) error {
	driver.logReq("DeleteNetwork", delreq, delreq.NetworkID)
	driver.Lock()
	delete(driver.networks, delreq.NetworkID)
	driver.Unlock()
	return nil
}

func (driver *driver) CreateEndpoint(create *api.CreateEndpointRequest) (*api.CreateEndpointResponse, error) {
	driver.logReq("CreateEndpoint", create, create.EndpointID)
	common.Log.Debugf("interface %+v", create.Interface)

	if create.Interface == nil {
		return nil, driver.error("CreateEndpoint", "Not supported: creating an interface from within CreateEndpoint")
	}

	// create veths. note we assume endpoint IDs are unique in the first 9 chars
	name, peerName := vethPair(create.EndpointID)
	if _, err := weavenet.CreateAndAttachVeth(name, peerName, weavenet.WeaveBridgeName, 0, false, nil); err != nil {
		return nil, driver.error("JoinEndpoint", "%s", err)
	}

	// Send back the MAC address
	link, _ := netlink.LinkByName(peerName)
	resp := &api.CreateEndpointResponse{Interface: &api.EndpointInterface{MacAddress: link.Attrs().HardwareAddr.String()}}

	driver.logRes("CreateEndpoint", resp)
	return resp, nil
}

func (driver *driver) DeleteEndpoint(deleteReq *api.DeleteEndpointRequest) error {
	driver.logReq("DeleteEndpoint", deleteReq, deleteReq.EndpointID)
	name, _ := vethPair(deleteReq.EndpointID)
	veth := &netlink.Veth{LinkAttrs: netlink.LinkAttrs{Name: name}}
	if err := netlink.LinkDel(veth); err != nil {
		// Try again using the name construction from earlier plugin version,
		// in case user has upgraded with endpoints still extant
		veth.Name = "vethwl" + deleteReq.EndpointID[:5]
		if err2 := netlink.LinkDel(veth); err2 != nil {
			// Note we report the first error
			driver.warn("LeaveEndpoint", "unable to delete veth %q: %s", name, err)
		}
	}
	return nil
}

func (driver *driver) EndpointInfo(req *api.EndpointInfoRequest) (*api.EndpointInfoResponse, error) {
	driver.logReq("EndpointInfo", req, req.EndpointID)
	return &api.EndpointInfoResponse{Value: map[string]interface{}{}}, nil
}

func (driver *driver) JoinEndpoint(j *api.JoinRequest) (*api.JoinResponse, error) {
	driver.logReq("JoinEndpoint", j, fmt.Sprintf("%s:%s to %s", j.NetworkID, j.EndpointID, j.SandboxKey))

	network, err := driver.findNetworkInfo(j.NetworkID)
	if err != nil {
		return nil, driver.error("JoinEndpoint", "unable to find network info: %s", err)
	}

	_, peerName := vethPair(j.EndpointID)
	response := &api.JoinResponse{
		InterfaceName: &api.InterfaceName{
			SrcName:   peerName,
			DstPrefix: weavenet.VethName,
		},
	}
	if network.hasMulticastRoute {
		multicastRoute := api.StaticRoute{
			Destination: "224.0.0.0/4",
			RouteType:   types.CONNECTED,
		}
		response.StaticRoutes = append(response.StaticRoutes, multicastRoute)
	}
	driver.logRes("JoinEndpoint", response)
	return response, nil
}

func (driver *driver) findNetworkInfo(id string) (network, error) {
	driver.Lock()
	network, found := driver.networks[id]
	driver.Unlock()
	if found {
		return network, nil
	}
	info, err := driver.docker.NetworkInfo(id)
	if err != nil {
		return network, err
	}
	return driver.setupNetworkInfo(id, driver.isNetworkOur(info.Driver), info.Options)
}

func (driver *driver) setupNetworkInfo(id string, isOurs bool, options map[string]string) (network, error) {
	network := network{isOurs: isOurs}
	if isOurs {
		for key, value := range options {
			switch key {
			case MulticastOption:
				if value == "" { // interpret "--opt works.weave.multicast" as "turn it on"
					network.hasMulticastRoute = true
				} else {
					var err error
					if network.hasMulticastRoute, err = strconv.ParseBool(value); err != nil {
						return network, fmt.Errorf("unrecognized value %q for option %s", value, key)
					}

				}
			default:
				driver.warn("setupNetworkInfo", "unrecognized option: %s", key)
			}
		}
	}
	driver.Lock()
	driver.networks[id] = network
	driver.Unlock()
	return network, nil
}

func (driver *driver) LeaveEndpoint(leave *api.LeaveRequest) error {
	driver.logReq("LeaveEndpoint", leave, fmt.Sprintf("%s:%s", leave.NetworkID, leave.EndpointID))
	return nil
}

func (driver *driver) DiscoverNew(disco *api.DiscoveryNotification) error {
	driver.logReq("DiscoverNew", disco, "")
	return nil
}

func (driver *driver) DiscoverDelete(disco *api.DiscoveryNotification) error {
	driver.logReq("DiscoverDelete", disco, "")
	return nil
}

func vethPair(id string) (string, string) {
	// IFNAMSIZ is buffer length; subtract 6 for "vethwl" and 1 for terminating nul
	return "vethwl" + id[:unix.IFNAMSIZ-7], "vethwg" + id[:unix.IFNAMSIZ-7]
}

// logging

func (driver *driver) logReq(fun string, req interface{}, short string) {
	driver.log(common.Log.Debugf, " %+v", fun, req)
	common.Log.Infof("[net] %s %s", fun, short)
}

func (driver *driver) logRes(fun string, res interface{}) {
	driver.log(common.Log.Debugf, " %+v", fun, res)
}

func (driver *driver) warn(fun string, format string, a ...interface{}) {
	driver.log(common.Log.Warnf, ": "+format, fun, a...)
}

func (driver *driver) debug(fun string, format string, a ...interface{}) {
	driver.log(common.Log.Debugf, ": "+format, fun, a...)
}

func (driver *driver) error(fun string, format string, a ...interface{}) error {
	driver.log(common.Log.Errorf, ": "+format, fun, a...)
	return fmt.Errorf(format, a...)
}

func (driver *driver) log(f func(string, ...interface{}), format string, fun string, a ...interface{}) {
	f("[net] %s"+format, append([]interface{}{fun}, a...)...)
}
