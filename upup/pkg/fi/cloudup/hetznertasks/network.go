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
	"net"
	"strconv"
	"time"

	"github.com/hetznercloud/hcloud-go/hcloud"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/hetzner"
)

// +kops:fitask
type Network struct {
	Name      *string
	Lifecycle fi.Lifecycle

	ID      *int
	Region  string
	IPRange string
	Subnets []string

	Labels map[string]string
}

var _ fi.CompareWithID = &Network{}

func (v *Network) CompareWithID() *string {
	return fi.String(strconv.Itoa(fi.IntValue(v.ID)))
}

func (v *Network) Find(c *fi.Context) (*Network, error) {
	cloud := c.Cloud.(hetzner.HetznerCloud)
	client := cloud.NetworkClient()

	// TODO(hakman): Find using label selector
	networks, err := client.All(context.TODO())
	if err != nil {
		return nil, err
	}

	for _, network := range networks {
		if network.Name == fi.StringValue(v.Name) {
			matches := &Network{
				Name:      fi.String(network.Name),
				Lifecycle: v.Lifecycle,
				ID:        fi.Int(network.ID),
				IPRange:   network.IPRange.String(),
				Labels:    network.Labels,
			}
			matches.Region = v.Region
			for _, subnet := range network.Subnets {
				if subnet.IPRange != nil {
					matches.Region = string(subnet.NetworkZone)
					matches.Subnets = append(matches.Subnets, subnet.IPRange.String())
				}
			}
			v.ID = matches.ID
			return matches, nil
		}
	}

	return nil, nil
}

func (v *Network) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(v, c)
}

func (_ *Network) CheckChanges(a, e, changes *Network) error {
	if a != nil {
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
		if changes.ID != nil {
			return fi.CannotChangeField("ID")
		}
		if changes.Region != "" {
			return fi.CannotChangeField("Region")
		}
		if changes.IPRange != "" {
			return fi.CannotChangeField("IPRange")
		}
		if len(changes.Subnets) > 0 && len(a.Subnets) > 0 {
			return fi.CannotChangeField("Subnets")
		}
	} else {
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
		if e.Region == "" {
			return fi.RequiredField("Region")
		}
		if e.IPRange == "" {
			return fi.RequiredField("IPRange")
		}
		if len(e.Subnets) == 0 {
			return fi.RequiredField("Subnets")
		}
	}
	return nil
}

func (_ *Network) RenderHetzner(t *hetzner.HetznerAPITarget, a, e, changes *Network) error {
	client := t.Cloud.NetworkClient()

	var network *hcloud.Network
	if a == nil {
		_, ipRange, err := net.ParseCIDR(e.IPRange)
		if err != nil {
			return err
		}
		opts := hcloud.NetworkCreateOpts{
			Name:    fi.StringValue(e.Name),
			IPRange: ipRange,
			Labels:  e.Labels,
		}
		network, _, err = client.Create(context.TODO(), opts)
		if err != nil {
			return err
		}
		e.ID = fi.Int(network.ID)

	} else {
		var err error
		network, _, err = client.Get(context.TODO(), fi.StringValue(e.Name))
		if err != nil {
			return err
		}

		// Update the labels
		if changes.Name != nil || len(changes.Labels) != 0 {
			_, _, err := client.Update(context.TODO(), network, hcloud.NetworkUpdateOpts{
				Name:   fi.StringValue(e.Name),
				Labels: e.Labels,
			})
			if err != nil {
				return err
			}
		}
	}

	// Add subnets separately and follow the progress
	if a == nil || len(a.Subnets) == 0 {
		for _, subnet := range e.Subnets {
			_, subnetIpRange, err := net.ParseCIDR(subnet)
			if err != nil {
				return err
			}
			action, _, err := client.AddSubnet(context.TODO(), network, hcloud.NetworkAddSubnetOpts{
				Subnet: hcloud.NetworkSubnet{
					Type:        hcloud.NetworkSubnetTypeCloud,
					NetworkZone: hcloud.NetworkZone(e.Region),
					IPRange:     subnetIpRange,
				},
			})
			if err != nil {
				return err
			}
			// Check progress
			for action.Progress < 100 {
				time.Sleep(5 * time.Second)
				actionClient := t.Cloud.ActionClient()
				action, _, err = actionClient.GetByID(context.TODO(), action.ID)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
