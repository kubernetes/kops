package ipam

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/weaveworks/weave/common"
	"github.com/weaveworks/weave/common/docker"
	"github.com/weaveworks/weave/net/address"
)

func badRequest(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusBadRequest)
	common.Log.Warningln("[allocator]:", err.Error())
}

func parseCIDR(w http.ResponseWriter, cidrStr string, net bool) (address.CIDR, bool) {
	var cidr address.CIDR
	var err error
	if net {
		cidr, err = ParseCIDRSubnet(cidrStr)
	} else {
		cidr, err = address.ParseCIDR(cidrStr)
	}
	if err != nil {
		badRequest(w, err)
		return address.CIDR{}, false
	}
	return cidr, true
}

func writeAddresses(w http.ResponseWriter, cidrs []address.CIDR) {
	for i, cidr := range cidrs {
		fmt.Fprint(w, cidr)
		if i < len(cidrs)-1 {
			w.Write([]byte{' '})
		}
	}
}

func hasBeenCancelled(dockerCli *docker.Client, closedChan <-chan bool, ident string, checkAlive bool) func() bool {
	return func() bool {
		select {
		case <-closedChan:
			return true
		default:
			res := checkAlive && dockerCli != nil && dockerCli.IsContainerNotRunning(ident)
			checkAlive = false // we check only once; if the container dies later we learn about that through events
			return res
		}
	}
}

func cancellationErr(w http.ResponseWriter, err error) bool {
	if _, ok := err.(*errorCancelled); ok {
		common.Log.Infoln("[allocator]:", err.Error())
		fmt.Fprint(w, "cancelled")
		return true
	}
	return false
}

func (alloc *Allocator) handleHTTPAllocate(dockerCli *docker.Client, w http.ResponseWriter, ident string, checkAlive bool, subnet address.CIDR) {
	addr, err := alloc.Allocate(ident, subnet, checkAlive,
		hasBeenCancelled(dockerCli, w.(http.CloseNotifier).CloseNotify(), ident, checkAlive))
	if err != nil {
		if !cancellationErr(w, err) {
			badRequest(w, err)
		}
		return
	}
	fmt.Fprintf(w, "%s/%d", addr, subnet.PrefixLen)
}

func (alloc *Allocator) handleHTTPClaim(dockerCli *docker.Client, w http.ResponseWriter, ident string, cidr address.CIDR, checkAlive, noErrorOnUnknown bool) {
	err := alloc.Claim(ident, cidr, checkAlive, noErrorOnUnknown,
		hasBeenCancelled(dockerCli, w.(http.CloseNotifier).CloseNotify(), ident, checkAlive))
	if err != nil {
		if !cancellationErr(w, err) {
			badRequest(w, fmt.Errorf("Unable to claim: %s", err))
		}
		return
	}
	w.WriteHeader(204)
}

// HandleHTTP wires up ipams HTTP endpoints to the provided mux.
func (alloc *Allocator) HandleHTTP(router *mux.Router, defaultSubnet address.CIDR, tracker string, dockerCli *docker.Client) {
	router.Methods("GET").Path("/ipinfo/defaultsubnet").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s", defaultSubnet)
	})

	router.Methods("PUT").Path("/ip/{id}/{ip}/{prefixlen}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		if cidr, ok := parseCIDR(w, vars["ip"]+"/"+vars["prefixlen"], false); ok {
			ident := vars["id"]
			checkAlive := r.FormValue("check-alive") == "true"
			noErrorOnUnknown := r.FormValue("noErrorOnUnknown") == "true"
			alloc.handleHTTPClaim(dockerCli, w, ident, cidr, checkAlive, noErrorOnUnknown)
		}
	})

	router.Methods("GET").Path("/ring").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		alloc.Prime()
	})

	router.Methods("GET").Path("/ip/{id}/{ip}/{prefixlen}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		if subnet, ok := parseCIDR(w, vars["ip"]+"/"+vars["prefixlen"], true); ok {
			cidrs, err := alloc.Lookup(vars["id"], subnet.HostRange())
			if err != nil {
				http.NotFound(w, r)
				return
			}
			writeAddresses(w, cidrs)
		}
	})

	router.Methods("GET").Path("/ip/{id}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		addrs, err := alloc.Lookup(mux.Vars(r)["id"], defaultSubnet.HostRange())
		if err != nil {
			http.NotFound(w, r)
			return
		}
		writeAddresses(w, addrs)
	})

	router.Methods("GET").Path("/ip").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type mapping struct {
			ContainerID string   `json:"containerid"`
			Addrs       []string `json:"addrs"`
		}

		type mappings struct {
			Owned []mapping `json:"owned"`
		}

		resultChan := make(chan mappings)
		alloc.actionChan <- func() {
			ms := mappings{}
			for containerid, d := range alloc.owned {
				m := mapping{
					ContainerID: containerid,
					Addrs:       []string{},
				}
				for _, addr := range d.Cidrs {
					m.Addrs = append(m.Addrs, addr.String())
				}
				ms.Owned = append(ms.Owned, m)
			}
			resultChan <- ms
		}
		ms := <-resultChan

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(ms); err != nil {
			common.Log.Warningln("[allocator]:", err.Error())
		}
	})

	router.Methods("POST").Path("/ip/{id}/{ip}/{prefixlen}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		if subnet, ok := parseCIDR(w, vars["ip"]+"/"+vars["prefixlen"], true); ok {
			alloc.handleHTTPAllocate(dockerCli, w, vars["id"], r.FormValue("check-alive") == "true", subnet)
		}
	})

	router.Methods("POST").Path("/ip/{id}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		alloc.handleHTTPAllocate(dockerCli, w, vars["id"], r.FormValue("check-alive") == "true", defaultSubnet)
	})

	router.Methods("DELETE").Path("/ip/{id}/{ip}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		ident := vars["id"]
		ipStr := vars["ip"]
		if ip, err := address.ParseIP(ipStr); err != nil {
			badRequest(w, err)
			return
		} else if err := alloc.Free(ident, ip); err != nil {
			badRequest(w, fmt.Errorf("Unable to free: %s", err))
			return
		}

		w.WriteHeader(204)
	})

	router.Methods("DELETE").Path("/ip/{id}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ident := mux.Vars(r)["id"]
		if err := alloc.Delete(ident); err != nil {
			badRequest(w, err)
			return
		}

		w.WriteHeader(204)
	})

	router.Methods("GET").Path("/ipinfo/tracker").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, tracker)
	})
}
