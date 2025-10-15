/*
Copyright 2025 The Kubernetes Authors.

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

package elementotasks

import (
	"context"
	"fmt"
	"net"

	"github.com/Elemento-Modular-Cloud/ecloud-go/ecloud"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/elemento"
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

func (v *Network) Find(c *fi.CloudupContext) (*Network, error) {
	cloud := c.T.Cloud.(elemento.ElementoCloud)
	client := cloud.NetworkClient()

	idOrName := fi.ValueOf(v.Name)
	if v.ID != nil {
		idOrName = fi.ValueOf(v.ID)
		network, _, err := client.GetByID(context.TODO(), idOrName)
		if err == nil && network != nil {
			return &Network{
				Name:      v.Name,
				Lifecycle: v.Lifecycle,
				ID:        fi.PtrTo(network.ID),
			}, nil
		}
	}

	network, _, err := client.GetByName(context.TODO(), idOrName)

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
		ID:        fi.PtrTo(network.ID),
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

func (v *Network) Run(c *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(v, c)
}

func (*Network) CheckChanges(a, e, changes *Network) error {
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

func (*Network) RenderElemento(t *elemento.ElementoAPITarget, a, e, changes *Network) error {
	client := t.Cloud.NetworkClient()

	var network *ecloud.Network
	if a == nil {
		// Network doesn't exist, create it
		_, ipRange, err := net.ParseCIDR(e.IPRange)
		if err != nil {
			return err
		}
		opts := ecloud.NetworkCreateOpts{
			Name:    fi.ValueOf(e.Name),
			IPRange: ipRange,
			Labels:  e.Labels,
		}
		network, _, err = client.Create(context.TODO(), opts)
		if err != nil {
			return err
		}
		e.ID = fi.PtrTo(network.ID)
	} else {
		// Network exists, get it
		var err error
		network, _, err = client.GetByName(context.TODO(), fi.ValueOf(e.Name))
		if err != nil {
			return err
		}

		// Update the labels - NOT SUPPORTED
		// if changes.Name != nil || len(changes.Labels) != 0 {
		// 	_, _, err := client.Update(context.TODO(), network, ecloud.NetworkUpdateOpts{
		// 		Name:   fi.ValueOf(e.Name),
		// 		Labels: e.Labels,
		// 	})
		// 	if err != nil {
		// 		return err
		// 	}
		// }
	}

	// Add subnets separately and follow the progress
	if a == nil || len(a.Subnets) == 0 {
		for _, subnet := range e.Subnets {
			_, subnetIpRange, err := net.ParseCIDR(subnet)
			if err != nil {
				return err
			}
			network.Subnets = append(network.Subnets, ecloud.NetworkSubnet{
				Type:        ecloud.NetworkSubnetTypeCloud,
				NetworkZone: ecloud.NetworkZone(e.Region),
				IPRange:     subnetIpRange,
			})
		}
	}

	return nil
}
