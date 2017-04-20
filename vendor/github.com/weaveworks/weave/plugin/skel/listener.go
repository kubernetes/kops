package skel

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/docker/libnetwork/drivers/remote/api"
	"github.com/docker/libnetwork/ipamapi"
	iapi "github.com/docker/libnetwork/ipams/remote/api"
)

const (
	networkReceiver = "NetworkDriver"
	ipamReceiver    = ipamapi.PluginEndpointType
)

type Driver interface {
	GetCapabilities() (*api.GetCapabilityResponse, error)
	CreateNetwork(create *api.CreateNetworkRequest) error
	DeleteNetwork(delete *api.DeleteNetworkRequest) error
	CreateEndpoint(create *api.CreateEndpointRequest) (*api.CreateEndpointResponse, error)
	DeleteEndpoint(delete *api.DeleteEndpointRequest) error
	EndpointInfo(req *api.EndpointInfoRequest) (*api.EndpointInfoResponse, error)
	JoinEndpoint(j *api.JoinRequest) (response *api.JoinResponse, error error)
	NetworkAllocate(*api.AllocateNetworkRequest) (*api.AllocateNetworkResponse, error)
	NetworkFree(*api.FreeNetworkRequest) (*api.FreeNetworkResponse, error)
	LeaveEndpoint(leave *api.LeaveRequest) error
	DiscoverNew(discover *api.DiscoveryNotification) error
	DiscoverDelete(delete *api.DiscoveryNotification) error
}

type IpamCapability interface {
	GetCapabilities() (RequiresMACAddress bool, RequiresRequestReplay bool, err error)
}

type listener struct {
	d Driver
	i ipamapi.Ipam
}

func Listen(socket net.Listener, driver Driver, ipamDriver ipamapi.Ipam) error {
	router := mux.NewRouter()
	router.NotFoundHandler = http.HandlerFunc(notFound)

	listener := &listener{driver, ipamDriver}

	router.Methods("POST").Path("/Plugin.Activate").HandlerFunc(listener.handshake)

	handleMethod := func(receiver, method string, h http.HandlerFunc) {
		router.Methods("POST").Path(fmt.Sprintf("/%s.%s", receiver, method)).HandlerFunc(h)
	}

	handleMethod(networkReceiver, "GetCapabilities", listener.getCapabilities)

	if driver != nil {
		handleMethod(networkReceiver, "CreateNetwork", listener.createNetwork)
		handleMethod(networkReceiver, "DeleteNetwork", listener.deleteNetwork)
		handleMethod(networkReceiver, "CreateEndpoint", listener.createEndpoint)
		handleMethod(networkReceiver, "DeleteEndpoint", listener.deleteEndpoint)
		handleMethod(networkReceiver, "EndpointOperInfo", listener.infoEndpoint)
		handleMethod(networkReceiver, "Join", listener.joinEndpoint)
		handleMethod(networkReceiver, "Leave", listener.leaveEndpoint)
		handleMethod(networkReceiver, "AllocateNetwork", listener.networkAllocate)
		handleMethod(networkReceiver, "FreeNetwork", listener.networkFree)
	}

	if ipamDriver != nil {
		handleMethod(ipamReceiver, "GetCapabilities", listener.getIpamCapabilities)
		handleMethod(ipamReceiver, "GetDefaultAddressSpaces", listener.getDefaultAddressSpaces)
		handleMethod(ipamReceiver, "RequestPool", listener.requestPool)
		handleMethod(ipamReceiver, "ReleasePool", listener.releasePool)
		handleMethod(ipamReceiver, "RequestAddress", listener.requestAddress)
		handleMethod(ipamReceiver, "ReleaseAddress", listener.releaseAddress)
	}

	return http.Serve(socket, router)
}

func decode(w http.ResponseWriter, r *http.Request, v interface{}) error {
	err := json.NewDecoder(r.Body).Decode(v)
	if err != nil {
		sendError(w, "Unable to decode JSON payload: "+err.Error(), http.StatusBadRequest)
	}
	return err
}

// === protocol handlers

type handshakeResp struct {
	Implements []string
}

func (listener *listener) handshake(w http.ResponseWriter, r *http.Request) {
	var resp handshakeResp
	if listener.d != nil {
		resp.Implements = append(resp.Implements, "NetworkDriver")
	}
	if listener.i != nil {
		resp.Implements = append(resp.Implements, "IpamDriver")
	}
	err := json.NewEncoder(w).Encode(&resp)
	if err != nil {
		sendError(w, "encode error", http.StatusInternalServerError)
		return
	}
}

func (listener *listener) getCapabilities(w http.ResponseWriter, r *http.Request) {
	caps, err := listener.d.GetCapabilities()
	objectOrErrorResponse(w, caps, err)
}

func (listener *listener) createNetwork(w http.ResponseWriter, r *http.Request) {
	var create api.CreateNetworkRequest
	err := json.NewDecoder(r.Body).Decode(&create)
	if err != nil {
		sendError(w, "Unable to decode JSON payload: "+err.Error(), http.StatusBadRequest)
		return
	}
	emptyOrErrorResponse(w, listener.d.CreateNetwork(&create))
}

func (listener *listener) deleteNetwork(w http.ResponseWriter, r *http.Request) {
	var delete api.DeleteNetworkRequest
	if err := json.NewDecoder(r.Body).Decode(&delete); err != nil {
		sendError(w, "Unable to decode JSON payload: "+err.Error(), http.StatusBadRequest)
		return
	}
	emptyOrErrorResponse(w, listener.d.DeleteNetwork(&delete))
}

func (listener *listener) createEndpoint(w http.ResponseWriter, r *http.Request) {
	var create api.CreateEndpointRequest
	if err := json.NewDecoder(r.Body).Decode(&create); err != nil {
		sendError(w, "unable to decode JSON payload: "+err.Error(), http.StatusBadRequest)
		return
	}
	res, err := listener.d.CreateEndpoint(&create)
	objectOrErrorResponse(w, res, err)
}

func (listener *listener) deleteEndpoint(w http.ResponseWriter, r *http.Request) {
	var delete api.DeleteEndpointRequest
	if err := json.NewDecoder(r.Body).Decode(&delete); err != nil {
		sendError(w, "Could not decode JSON encoded payload", http.StatusBadRequest)
		return
	}
	emptyOrErrorResponse(w, listener.d.DeleteEndpoint(&delete))
}

func (listener *listener) infoEndpoint(w http.ResponseWriter, r *http.Request) {
	var req api.EndpointInfoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Could not decode JSON encoded payload", http.StatusBadRequest)
		return
	}
	info, err := listener.d.EndpointInfo(&req)
	objectOrErrorResponse(w, info, err)
}

func (listener *listener) joinEndpoint(w http.ResponseWriter, r *http.Request) {
	var join api.JoinRequest
	if err := json.NewDecoder(r.Body).Decode(&join); err != nil {
		sendError(w, "Could not decode JSON encoded payload", http.StatusBadRequest)
		return
	}
	res, err := listener.d.JoinEndpoint(&join)
	objectOrErrorResponse(w, res, err)
}

func (listener *listener) networkAllocate(w http.ResponseWriter, r *http.Request) {
	var alloc api.AllocateNetworkRequest
	if err := json.NewDecoder(r.Body).Decode(&alloc); err != nil {
		sendError(w, "Could not decode JSON encode payload", http.StatusBadRequest)
		return
	}
	res, err := listener.d.NetworkAllocate(&alloc)
	objectOrErrorResponse(w, res, err)
}

func (listener *listener) networkFree(w http.ResponseWriter, r *http.Request) {
	var free api.FreeNetworkRequest
	if err := json.NewDecoder(r.Body).Decode(&free); err != nil {
		sendError(w, "Could not decode JSON encode payload", http.StatusBadRequest)
		return
	}
	res, err := listener.d.NetworkFree(&free)
	objectOrErrorResponse(w, res, err)
}

func (listener *listener) leaveEndpoint(w http.ResponseWriter, r *http.Request) {
	var l api.LeaveRequest
	if err := json.NewDecoder(r.Body).Decode(&l); err != nil {
		sendError(w, "Could not decode JSON encoded payload", http.StatusBadRequest)
		return
	}
	emptyOrErrorResponse(w, listener.d.LeaveEndpoint(&l))
}

func (listener *listener) discoverNew(w http.ResponseWriter, r *http.Request) {
	var disco api.DiscoveryNotification
	if err := json.NewDecoder(r.Body).Decode(&disco); err != nil {
		sendError(w, "Could not decode JSON encoded payload", http.StatusBadRequest)
		return
	}
	emptyOrErrorResponse(w, listener.d.DiscoverNew(&disco))
}

func (listener *listener) discoverDelete(w http.ResponseWriter, r *http.Request) {
	var disco api.DiscoveryNotification
	if err := json.NewDecoder(r.Body).Decode(&disco); err != nil {
		sendError(w, "Could not decode JSON encoded payload", http.StatusBadRequest)
		return
	}
	emptyOrErrorResponse(w, listener.d.DiscoverDelete(&disco))
}

// ===

func (listener *listener) getIpamCapabilities(w http.ResponseWriter, r *http.Request) {
	c, ok := listener.i.(IpamCapability)
	if !ok {
		http.NotFound(w, r)
		return
	}
	requiresMACAddress, requiresRequestReplay, err := c.GetCapabilities()
	response := &ipamapi.Capability{
		RequiresMACAddress:    requiresMACAddress,
		RequiresRequestReplay: requiresRequestReplay,
	}
	objectOrErrorResponse(w, response, err)
}

func (listener *listener) getDefaultAddressSpaces(w http.ResponseWriter, r *http.Request) {
	local, global, err := listener.i.GetDefaultAddressSpaces()
	response := &iapi.GetAddressSpacesResponse{
		LocalDefaultAddressSpace:  local,
		GlobalDefaultAddressSpace: global,
	}
	objectOrErrorResponse(w, response, err)
}

func (listener *listener) requestPool(w http.ResponseWriter, r *http.Request) {
	var rq iapi.RequestPoolRequest
	if err := decode(w, r, &rq); err != nil {
		return
	}
	poolID, pool, data, err := listener.i.RequestPool(rq.AddressSpace, rq.Pool, rq.SubPool, rq.Options, rq.V6)
	if err != nil {
		errorResponse(w, err.Error())
		return
	}
	response := &iapi.RequestPoolResponse{
		PoolID: poolID,
		Pool:   pool.String(),
		Data:   data,
	}
	objectResponse(w, response)
}

func (listener *listener) releasePool(w http.ResponseWriter, r *http.Request) {
	var rq iapi.ReleasePoolRequest
	if err := decode(w, r, &rq); err != nil {
		return
	}
	err := listener.i.ReleasePool(rq.PoolID)
	emptyOrErrorResponse(w, err)
}

func (listener *listener) requestAddress(w http.ResponseWriter, r *http.Request) {
	var rq iapi.RequestAddressRequest
	if err := decode(w, r, &rq); err != nil {
		return
	}
	address, data, err := listener.i.RequestAddress(rq.PoolID, net.ParseIP(rq.Address), rq.Options)
	if err != nil {
		errorResponse(w, err.Error())
		return
	}
	response := &iapi.RequestAddressResponse{
		Address: address.String(),
		Data:    data,
	}
	objectResponse(w, response)
}

func (listener *listener) releaseAddress(w http.ResponseWriter, r *http.Request) {
	var rq iapi.ReleaseAddressRequest
	if err := decode(w, r, &rq); err != nil {
		return
	}
	err := listener.i.ReleaseAddress(rq.PoolID, net.ParseIP(rq.Address))
	emptyOrErrorResponse(w, err)
}

// ===

func notFound(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}

func sendError(w http.ResponseWriter, msg string, code int) {
	http.Error(w, msg, code)
}

func errorResponse(w http.ResponseWriter, fmtString string, item ...interface{}) {
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(map[string]string{
		"Err": fmt.Sprintf(fmtString, item...),
	})
}

func objectResponse(w http.ResponseWriter, obj interface{}) {
	if err := json.NewEncoder(w).Encode(obj); err != nil {
		sendError(w, "Could not JSON encode response", http.StatusInternalServerError)
		return
	}
}

func emptyResponse(w http.ResponseWriter) {
	json.NewEncoder(w).Encode(map[string]string{})
}

func objectOrErrorResponse(w http.ResponseWriter, obj interface{}, err error) {
	if err != nil {
		errorResponse(w, err.Error())
		return
	}
	objectResponse(w, obj)
}

func emptyOrErrorResponse(w http.ResponseWriter, err error) {
	if err != nil {
		errorResponse(w, err.Error())
		return
	}
	emptyResponse(w)
}
