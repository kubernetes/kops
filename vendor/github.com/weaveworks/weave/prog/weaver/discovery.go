package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/weaveworks/weave/ipam"
)

// TODO: move these definitions somewhere more shareable
type PeerUpdateRequest struct {
	Name      string   `json:"peername"`
	Nickname  string   `json:"nickname"`  // optional
	Addresses []string `json:"addresses"` // can be empty
}

type PeerUpdateResponse struct {
	Addresses []string `json:"addresses"`
	PeerCount int      `json:"peercount"`
}

var httpClient = &http.Client{Timeout: 30 * time.Second}

func do(verb string, discoveryEndpoint, token string, request interface{}, response interface{}) error {
	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(request)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(verb, discoveryEndpoint+"/peer", body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Scope-Probe token=%s", token))
	req.Header.Set("X-Weave-Net-Version", version)
	req.Header.Set("Content-Type", "application/json")
	Log.Printf("Calling peer discovery %s with %s", req.URL, request)
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		rbody, _ := ioutil.ReadAll(resp.Body)
		return errors.New(resp.Status + ": " + string(rbody))
	}
	if response == nil {
		return nil
	}
	err = json.NewDecoder(resp.Body).Decode(response)
	Log.Printf("peer discovery result: (%v) %v", err, response)
	return err
}

func peerDiscoveryUpdate(discoveryEndpoint, token, peername, nickname string, addresses []string) ([]string, int, error) {
	request := PeerUpdateRequest{
		Name:      peername,
		Nickname:  nickname,
		Addresses: addresses,
	}
	var updateResponse PeerUpdateResponse
	err := do("POST", discoveryEndpoint, token, request, &updateResponse)
	return updateResponse.Addresses, updateResponse.PeerCount, err
}

func peerDiscoveryDelete(discoveryEndpoint, token, peername string) error {
	request := PeerUpdateRequest{Name: peername}
	return do("DELETE", discoveryEndpoint, token, request, nil)
}

func HandleHTTPPeer(router *mux.Router, alloc *ipam.Allocator, discoveryEndpoint, token, peername string) {
	router.Methods("DELETE").Path("/peer").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if discoveryEndpoint != "" {
			if err := peerDiscoveryDelete(discoveryEndpoint, token, peername); err != nil {
				Log.Errorf("Error while deleting self from peer discovery: %s", err)
			}
		}
		if alloc != nil {
			alloc.Shutdown()
		}
		w.WriteHeader(204)
	})

	router.Methods("DELETE").Path("/peer/{id}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ident := mux.Vars(r)["id"]
		if discoveryEndpoint != "" {
			// TODO: deal with this being either a peername or a nickname
			if err := peerDiscoveryDelete(discoveryEndpoint, token, ident); err != nil {
				Log.Errorf("Error while deleting self from peer discovery: %s", err)
			}
		}
		if alloc != nil {
			transferred := alloc.AdminTakeoverRanges(ident)
			fmt.Fprintf(w, "%d IPs taken over from %s\n", transferred, ident)
		}
	})
}
