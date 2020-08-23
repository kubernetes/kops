/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package mockcompute

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/gophercloud/gophercloud/openstack/networking/v2/ports"
	"k8s.io/kops/upup/pkg/fi"

	"github.com/google/uuid"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
)

type serverGetResponse struct {
	Server servers.Server `json:"server"`
}

type serverListResponse struct {
	Servers []servers.Server `json:"servers"`
}

type serverCreateRequest struct {
	Server Server `json:"server"`
}

// CreateOpts specifies server creation parameters.
type Server struct {
	// Name is the name to assign to the newly launched server.
	Name string `json:"name" required:"true"`

	// ImageRef [optional; required if ImageName is not provided] is the ID or
	// full URL to the image that contains the server's OS and initial state.
	// Also optional if using the boot-from-volume extension.
	ImageRef string `json:"imageRef"`

	// ImageName [optional; required if ImageRef is not provided] is the name of
	// the image that contains the server's OS and initial state.
	// Also optional if using the boot-from-volume extension.
	ImageName string `json:"-"`

	// FlavorRef [optional; required if FlavorName is not provided] is the ID or
	// full URL to the flavor that describes the server's specs.
	FlavorRef string `json:"flavorRef"`

	// FlavorName [optional; required if FlavorRef is not provided] is the name of
	// the flavor that describes the server's specs.
	FlavorName string `json:"-"`

	// SecurityGroups lists the names of the security groups to which this server
	// should belong.
	SecurityGroups []string `json:"-"`

	// UserData contains configuration information or scripts to use upon launch.
	// Create will base64-encode it for you, if it isn't already.
	UserData []byte `json:"-"`

	// AvailabilityZone in which to launch the server.
	AvailabilityZone string `json:"availability_zone,omitempty"`

	// Networks dictates how this server will be attached to available networks.
	// By default, the server will be attached to all isolated networks for the
	// tenant.
	// Starting with microversion 2.37 networks can also be an "auto" or "none"
	// string.
	Networks []Networks `json:"networks"`

	// Metadata contains key-value pairs (up to 255 bytes each) to attach to the
	// server.
	Metadata map[string]string `json:"metadata,omitempty"`

	// ConfigDrive enables metadata injection through a configuration drive.
	ConfigDrive *bool `json:"config_drive,omitempty"`

	// AdminPass sets the root user password. If not set, a randomly-generated
	// password will be created and returned in the response.
	AdminPass string `json:"adminPass,omitempty"`

	// AccessIPv4 specifies an IPv4 address for the instance.
	AccessIPv4 string `json:"accessIPv4,omitempty"`

	// AccessIPv6 specifies an IPv6 address for the instance.
	AccessIPv6 string `json:"accessIPv6,omitempty"`

	// Min specifies Minimum number of servers to launch.
	Min int `json:"min_count,omitempty"`

	// Max specifies Maximum number of servers to launch.
	Max int `json:"max_count,omitempty"`

	// ServiceClient will allow calls to be made to retrieve an image or
	// flavor ID by name.
	ServiceClient *gophercloud.ServiceClient `json:"-"`

	// Tags allows a server to be tagged with single-word metadata.
	// Requires microversion 2.52 or later.
	Tags []string `json:"tags,omitempty"`
}

type Networks struct {
	Port string `json:"port,omitempty"`
}

func (m *MockClient) mockServers() {
	re := regexp.MustCompile(`/servers/?`)

	handler := func(w http.ResponseWriter, r *http.Request) {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		w.Header().Add("Content-Type", "application/json")

		serverID := re.ReplaceAllString(r.URL.Path, "")
		switch r.Method {
		case http.MethodGet:
			if serverID == "detail" {
				r.ParseForm()
				m.listServers(w, r.Form)
			}
		case http.MethodPost:
			m.createServer(w, r)
		case http.MethodDelete:
			m.deleteServer(w, serverID)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}
	m.Mux.HandleFunc("/servers/", handler)
	m.Mux.HandleFunc("/servers", handler)
}

func (m *MockClient) listServers(w http.ResponseWriter, vals url.Values) {
	serverName := strings.Trim(vals.Get("name"), "^$")
	matched := make([]servers.Server, 0)
	for _, server := range m.servers {
		if server.Name == serverName {
			matched = append(matched, server)
		}
	}
	resp := serverListResponse{
		Servers: matched,
	}
	respB, err := json.Marshal(resp)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal %+v", resp))
	}
	_, err = w.Write(respB)
	if err != nil {
		panic("failed to write body")
	}
}

func (m *MockClient) deleteServer(w http.ResponseWriter, serverID string) {
	if _, ok := m.servers[serverID]; ok {
		delete(m.servers, serverID)
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (m *MockClient) createServer(w http.ResponseWriter, r *http.Request) {
	var create serverCreateRequest
	err := json.NewDecoder(r.Body).Decode(&create)
	if err != nil {
		panic("error decoding create server request")
	}

	w.WriteHeader(http.StatusCreated)

	server := servers.Server{
		ID:       uuid.New().String(),
		Name:     create.Server.Name,
		Metadata: create.Server.Metadata,
	}
	securityGroups := make([]map[string]interface{}, len(create.Server.SecurityGroups))
	for i, groupName := range create.Server.SecurityGroups {
		securityGroups[i] = map[string]interface{}{"name": groupName}
	}
	server.SecurityGroups = securityGroups

	portID := create.Server.Networks[0].Port
	ports.Update(m.networkClient, portID, ports.UpdateOpts{
		DeviceID: fi.String(server.ID),
	})

	// Assign an IP address
	private := make([]map[string]string, 1)
	private[0] = make(map[string]string)
	private[0]["OS-EXT-IPS:type"] = "fixed"
	private[0]["addr"] = "192.168.1.1"
	server.Addresses = map[string]interface{}{
		"private": private,
	}

	m.servers[server.ID] = server

	resp := serverGetResponse{
		Server: server,
	}
	respB, err := json.Marshal(resp)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal %+v", resp))
	}
	_, err = w.Write(respB)
	if err != nil {
		panic("failed to write body")
	}
}
