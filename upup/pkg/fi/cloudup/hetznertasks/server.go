/*
Copyright 2022 The Kubernetes Authors.

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

package hetznertasks

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hetznercloud/hcloud-go/hcloud"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/hetzner"
)

// +kops:fitask
type Server struct {
	Name      *string
	Lifecycle fi.Lifecycle
	SSHKey    *SSHKey
	Network   *Network

	ID       *int
	Location string
	Size     string
	Image    string

	EnableIPv4 bool
	EnableIPv6 bool

	UserData fi.Resource

	Labels map[string]string
}

var _ fi.CompareWithID = &Server{}

func (v *Server) CompareWithID() *string {
	return fi.String(strconv.Itoa(fi.IntValue(v.ID)))
}

func (v *Server) Find(c *fi.Context) (*Server, error) {
	cloud := c.Cloud.(hetzner.HetznerCloud)
	client := cloud.ServerClient()

	// TODO(hakman): Find using label selector
	servers, err := client.All(context.TODO())
	if err != nil {
		return nil, err
	}

	for _, server := range servers {
		if server.Name == fi.StringValue(v.Name) {
			matches := &Server{
				Lifecycle: v.Lifecycle,
				Name:      fi.String(server.Name),
				ID:        fi.Int(server.ID),
				Labels:    server.Labels,
			}

			if server.Datacenter != nil && server.Datacenter.Location != nil {
				matches.Location = server.Datacenter.Location.Name
			}
			if server.ServerType != nil {
				matches.Size = server.ServerType.Name
			}
			if server.Image != nil {
				matches.Image = server.Image.Name
			}
			if server.PublicNet.IPv4.IP != nil {
				matches.EnableIPv4 = true
			}
			if server.PublicNet.IPv6.IP != nil {
				matches.EnableIPv4 = true
			}

			// Ignore fields that are not returned by the Hetzner Cloud API
			matches.SSHKey = v.SSHKey
			matches.UserData = v.UserData

			// TODO: The API only returns the network ID, a new API call is required to get the network name
			matches.Network = v.Network

			v.ID = matches.ID
			return matches, nil
		}
	}

	return nil, nil
}

func (v *Server) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(v, c)
}

func (_ *Server) CheckChanges(a, e, changes *Server) error {
	if a != nil {
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
		if changes.ID != nil {
			return fi.CannotChangeField("ID")
		}
		if changes.Location != "" {
			return fi.CannotChangeField("Location")
		}
		if changes.Size != "" {
			return fi.CannotChangeField("Size")
		}
		if changes.Image != "" {
			return fi.CannotChangeField("Image")
		}
		if changes.UserData != nil {
			return fi.CannotChangeField("UserData")
		}
	} else {
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
		if e.Location == "" {
			return fi.RequiredField("Location")
		}
		if e.Size == "" {
			return fi.RequiredField("Size")
		}
		if e.Image == "" {
			return fi.RequiredField("Image")
		}
		if e.UserData == nil {
			return fi.RequiredField("UserData")
		}
	}
	return nil
}

func (_ *Server) RenderHetzner(t *hetzner.HetznerAPITarget, a, e, changes *Server) error {
	client := t.Cloud.ServerClient()
	if a == nil {
		if e.SSHKey == nil {
			return fmt.Errorf("failed to find ssh key for server %q", fi.StringValue(e.Name))
		}
		if e.Network == nil {
			return fmt.Errorf("failed to find network for server %q", fi.StringValue(e.Name))
		}

		userData, err := fi.ResourceAsString(e.UserData)
		if err != nil {
			return err
		}

		opts := hcloud.ServerCreateOpts{
			Name:             fi.StringValue(e.Name),
			StartAfterCreate: fi.Bool(true),
			SSHKeys: []*hcloud.SSHKey{
				{
					ID: fi.IntValue(e.SSHKey.ID),
				},
			},
			Networks: []*hcloud.Network{
				{
					ID: fi.IntValue(e.Network.ID),
				},
			},
			Location: &hcloud.Location{
				Name: e.Location,
			},
			ServerType: &hcloud.ServerType{
				Name: e.Size,
			},
			Image: &hcloud.Image{
				Name: e.Image,
			},
			UserData: userData,
			Labels:   e.Labels,
			PublicNet: &hcloud.ServerCreatePublicNet{
				EnableIPv4: e.EnableIPv4,
				EnableIPv6: e.EnableIPv6,
			},
		}

		_, _, err = client.Create(context.TODO(), opts)
		if err != nil {
			return err
		}

	} else {
		server, _, err := client.Get(context.TODO(), strconv.Itoa(fi.IntValue(a.ID)))
		if err != nil {
			return err
		}

		// Update the labels
		if changes.Name != nil || len(changes.Labels) != 0 {
			_, _, err := client.Update(context.TODO(), server, hcloud.ServerUpdateOpts{
				Name:   fi.StringValue(e.Name),
				Labels: e.Labels,
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}
