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
	"net"
	"strconv"
	"time"

	"github.com/hetznercloud/hcloud-go/hcloud"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/hetzner"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
)

// +kops:fitask
type Network struct {
	Name      *string
	Lifecycle fi.Lifecycle

	ID      *string
	Region  string
	IPRange string
	Subnets []string

	Labels map[string]string
}

var _ fi.CompareWithID = &Network{}

func (v *Network) CompareWithID() *string {
	return v.ID
}

func (v *Network) Find(c *fi.Context) (*Network, error) {
	cloud := c.Cloud.(hetzner.HetznerCloud)
	client := cloud.NetworkClient()

	idOrName := fi.StringValue(v.Name)
	if v.ID != nil {
		idOrName = fi.StringValue(v.ID)
	}

	network, _, err := client.Get(context.TODO(), idOrName)
	if err != nil {
		return nil, fmt.Errorf("failed to find network %q: %w", idOrName, err)
	}
	if network == nil {
		if v.ID != nil {
			return nil, fmt.Errorf("failed to find network %q", idOrName)
		}
		return nil, nil
	}

	matches := &Network{
		Name:      v.Name,
		Lifecycle: v.Lifecycle,
		ID:        fi.String(strconv.Itoa(network.ID)),
	}

	if v.ID == nil {
		matches.IPRange = network.IPRange.String()
		matches.Labels = network.Labels
		matches.Region = v.Region
		for _, subnet := range network.Subnets {
			if subnet.IPRange != nil {
				matches.Region = string(subnet.NetworkZone)
				matches.Subnets = append(matches.Subnets, subnet.IPRange.String())
			}
		}
		// Make sure the ID is set (used by other tasks)
		v.ID = matches.ID
	} else {
		// Make sure the ID is numerical
		v.ID = matches.ID
	}

	return matches, nil
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
		e.ID = fi.String(strconv.Itoa(network.ID))

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

type terraformNetwork struct {
	Name    *string           `cty:"name"`
	IPRange *string           `cty:"ip_range"`
	Labels  map[string]string `cty:"labels"`
}

type terraformNetworkSubnet struct {
	NetworkID   *terraformWriter.Literal `cty:"network_id"`
	Type        *string                  `cty:"type"`
	NetworkZone *string                  `cty:"network_zone"`
	IPRange     *string                  `cty:"ip_range"`
}

func (_ *Network) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *Network) error {
	{
		tf := &terraformNetwork{
			Name:    e.Name,
			IPRange: fi.String(e.IPRange),
			Labels:  e.Labels,
		}

		err := t.RenderResource("hcloud_network", *e.Name, tf)
		if err != nil {
			return err
		}
	}

	for _, subnet := range e.Subnets {
		_, subnetIpRange, err := net.ParseCIDR(subnet)
		if err != nil {
			return err
		}

		tf := &terraformNetworkSubnet{
			NetworkID:   e.TerraformLink(),
			Type:        fi.String(string(hcloud.NetworkSubnetTypeCloud)),
			IPRange:     fi.String(subnetIpRange.String()),
			NetworkZone: fi.String(e.Region),
		}

		err = t.RenderResource("hcloud_network_subnet", *e.Name+"-"+subnet, tf)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *Network) TerraformLink() *terraformWriter.Literal {
	return terraformWriter.LiteralProperty("hcloud_network", *e.Name, "id")
}
