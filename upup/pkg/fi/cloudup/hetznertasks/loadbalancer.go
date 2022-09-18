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
	"strings"

	"github.com/hetznercloud/hcloud-go/hcloud"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/hetzner"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
)

// +kops:fitask
type LoadBalancer struct {
	Name      *string
	Lifecycle fi.Lifecycle
	Network   *Network

	ID       *int
	Location string
	Type     string
	Services []*LoadBalancerService
	Target   string

	Labels map[string]string
}

var _ fi.CompareWithID = &LoadBalancer{}

func (v *LoadBalancer) CompareWithID() *string {
	return fi.String(strconv.Itoa(fi.IntValue(v.ID)))
}

var _ fi.HasAddress = &LoadBalancer{}

func (e *LoadBalancer) IsForAPIServer() bool {
	return true
}

func (v *LoadBalancer) FindAddresses(c *fi.Context) ([]string, error) {
	// TODO(hakman): Use mock to handle this more gracefully
	if strings.HasPrefix(c.ClusterConfigBase.Path(), "memfs://tests/") {
		return nil, nil
	}

	ctx := context.TODO()
	cloud := c.Cloud.(hetzner.HetznerCloud)
	client := cloud.LoadBalancerClient()

	// TODO(hakman): Find using label selector
	loadbalancers, err := client.All(ctx)
	if err != nil {
		return nil, err
	}

	for _, loadbalancer := range loadbalancers {
		if loadbalancer.Name == fi.StringValue(v.Name) {
			var addresses []string
			if loadbalancer.PublicNet.IPv4.IP == nil {
				return nil, fmt.Errorf("failed to find load-balancer %q public address", fi.StringValue(v.Name))
			}
			addresses = append(addresses, loadbalancer.PublicNet.IPv4.IP.String())
			for _, privateNetwork := range loadbalancer.PrivateNet {
				if privateNetwork.IP == nil {
					return nil, fmt.Errorf("failed to find load-balancer %q private address", fi.StringValue(v.Name))
				}
				addresses = append(addresses, privateNetwork.IP.String())
			}
			return addresses, nil
		}
	}

	return nil, nil
}

func (v *LoadBalancer) Find(c *fi.Context) (*LoadBalancer, error) {
	ctx := context.TODO()
	cloud := c.Cloud.(hetzner.HetznerCloud)
	client := cloud.LoadBalancerClient()

	// TODO(hakman): Find using label selector
	loadbalancers, err := client.All(ctx)
	if err != nil {
		return nil, err
	}

	for _, loadbalancer := range loadbalancers {
		if loadbalancer.Name == fi.StringValue(v.Name) {
			matches := &LoadBalancer{
				Lifecycle: v.Lifecycle,
				Name:      fi.String(loadbalancer.Name),
				ID:        fi.Int(loadbalancer.ID),
				Labels:    loadbalancer.Labels,
			}

			if loadbalancer.Location != nil {
				matches.Location = loadbalancer.Location.Name
			}
			if loadbalancer.LoadBalancerType != nil {
				matches.Type = loadbalancer.LoadBalancerType.Name
			}

			for _, service := range loadbalancer.Services {
				loadbalancerService := LoadBalancerService{
					Protocol:        string(service.Protocol),
					ListenerPort:    fi.Int(service.ListenPort),
					DestinationPort: fi.Int(service.DestinationPort),
				}
				matches.Services = append(matches.Services, &loadbalancerService)
			}

			for _, target := range loadbalancer.Targets {
				if target.Type == hcloud.LoadBalancerTargetTypeLabelSelector && target.LabelSelector != nil {
					matches.Target = target.LabelSelector.Selector
				}
			}

			// TODO: The API only returns the network ID, a new API call is required to get the network name
			matches.Network = v.Network

			v.ID = matches.ID
			return matches, nil
		}
	}

	return nil, nil
}

func (v *LoadBalancer) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(v, c)
}

func (_ *LoadBalancer) CheckChanges(a, e, changes *LoadBalancer) error {
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
		if changes.Type != "" {
			return fi.CannotChangeField("Type")
		}
		if len(changes.Services) > 0 && len(a.Services) > 0 {
			return fi.CannotChangeField("Subnets")
		}
		if changes.Target != "" && a.Target != "" {
			return fi.CannotChangeField("Target")
		}
	} else {
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
		if e.Location == "" {
			return fi.RequiredField("Location")
		}
		if e.Type == "" {
			return fi.RequiredField("Type")
		}
		if len(e.Services) == 0 {
			return fi.RequiredField("Services")
		}
		if e.Target == "" {
			return fi.RequiredField("Target")
		}
	}
	return nil
}

func (_ *LoadBalancer) RenderHetzner(t *hetzner.HetznerAPITarget, a, e, changes *LoadBalancer) error {
	ctx := context.TODO()
	actionClient := t.Cloud.ActionClient()
	client := t.Cloud.LoadBalancerClient()

	if a == nil {
		if e.Network == nil {
			return fmt.Errorf("failed to find network for loadbalancer %q", fi.StringValue(e.Name))
		}

		networkID, err := strconv.Atoi(fi.StringValue(e.Network.ID))
		if err != nil {
			return fmt.Errorf("failed to convert network ID %q to int: %w", fi.StringValue(e.Network.ID), err)
		}

		opts := hcloud.LoadBalancerCreateOpts{
			Name: fi.StringValue(e.Name),
			LoadBalancerType: &hcloud.LoadBalancerType{
				Name: e.Type,
			},
			Algorithm: &hcloud.LoadBalancerAlgorithm{
				Type: hcloud.LoadBalancerAlgorithmTypeRoundRobin,
			},
			Location: &hcloud.Location{
				Name: e.Location,
			},
			Labels: e.Labels,
			Targets: []hcloud.LoadBalancerCreateOptsTarget{
				{
					Type: hcloud.LoadBalancerTargetTypeLabelSelector,
					LabelSelector: hcloud.LoadBalancerCreateOptsTargetLabelSelector{
						Selector: e.Target,
					},
					UsePrivateIP: fi.Bool(true),
				},
			},
			Network: &hcloud.Network{
				ID: networkID,
			},
		}

		for _, service := range e.Services {
			opts.Services = append(opts.Services, hcloud.LoadBalancerCreateOptsService{
				Protocol:        hcloud.LoadBalancerServiceProtocol(service.Protocol),
				ListenPort:      service.ListenerPort,
				DestinationPort: service.DestinationPort,
			})
		}

		result, _, err := client.Create(ctx, opts)
		if err != nil {
			return err
		}
		_, errCh := actionClient.WatchProgress(ctx, result.Action)
		if err := <-errCh; err != nil {
			return err
		}

	} else {
		var err error
		loadbalancer, _, err := client.Get(ctx, strconv.Itoa(fi.IntValue(a.ID)))
		if err != nil {
			return err
		}

		// Update the labels
		if changes.Name != nil || len(changes.Labels) != 0 {
			_, _, err := client.Update(ctx, loadbalancer, hcloud.LoadBalancerUpdateOpts{
				Name:   fi.StringValue(e.Name),
				Labels: e.Labels,
			})
			if err != nil {
				return err
			}
		}

		// Update the services
		if len(changes.Services) > 0 {
			for _, service := range e.Services {
				action, _, err := client.AddService(ctx, loadbalancer, hcloud.LoadBalancerAddServiceOpts{
					Protocol:        hcloud.LoadBalancerServiceProtocol(service.Protocol),
					ListenPort:      service.ListenerPort,
					DestinationPort: service.DestinationPort,
				})
				if err != nil {
					if err != nil {
						return err
					}
				}
				_, errCh := actionClient.WatchProgress(ctx, action)
				if err := <-errCh; err != nil {
					return err
				}
			}
		}

		// Update the targets
		if a.Target == "" {
			action, _, err := client.AddLabelSelectorTarget(ctx, loadbalancer, hcloud.LoadBalancerAddLabelSelectorTargetOpts{
				Selector:     e.Target,
				UsePrivateIP: fi.Bool(true),
			})
			if err != nil {
				return err
			}
			_, errCh := actionClient.WatchProgress(ctx, action)
			if err := <-errCh; err != nil {
				return err
			}
		}
	}

	return nil
}

// LoadBalancerService represents a LoadBalancer's service.
type LoadBalancerService struct {
	Protocol        string
	ListenerPort    *int
	DestinationPort *int
}

var _ fi.HasDependencies = &LoadBalancerService{}

func (e *LoadBalancerService) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	return nil
}

type terraformLoadBalancer struct {
	Name     *string                      `cty:"name"`
	Type     *string                      `cty:"load_balancer_type"`
	Location *string                      `cty:"location"`
	Target   *terraformLoadBalancerTarget `cty:"target"`
	Network  *terraformWriter.Literal     `cty:"network"`
	Labels   map[string]string            `cty:"labels"`
}

type terraformLoadBalancerNetwork struct {
	LoadBalancerID *terraformWriter.Literal `cty:"load_balancer_id"`
	NetworkID      *terraformWriter.Literal `cty:"network_id"`
}

type terraformLoadBalancerService struct {
	LoadBalancerID  *terraformWriter.Literal `cty:"load_balancer_id"`
	Protocol        *string                  `cty:"protocol"`
	ListenPort      *int                     `cty:"listen_port"`
	DestinationPort *int                     `cty:"destination_port"`
}

type terraformLoadBalancerTarget struct {
	LoadBalancerID *terraformWriter.Literal `cty:"load_balancer_id"`
	Type           *string                  `cty:"type"`
	LabelSelector  *string                  `cty:"label_selector"`
	UsePrivateIP   *bool                    `cty:"use_private_ip"`
}

func (_ *LoadBalancer) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *LoadBalancer) error {
	{
		tf := &terraformLoadBalancer{
			Name:     e.Name,
			Type:     &e.Type,
			Location: &e.Location,
			Labels:   e.Labels,
		}

		err := t.RenderResource("hcloud_load_balancer", *e.Name, tf)
		if err != nil {
			return err
		}
	}

	{
		tf := &terraformLoadBalancerNetwork{
			LoadBalancerID: e.TerraformLink(),
			NetworkID:      e.Network.TerraformLink(),
		}

		err := t.RenderResource("hcloud_load_balancer_network", *e.Name, tf)
		if err != nil {
			return err
		}
	}

	for _, service := range e.Services {
		tf := &terraformLoadBalancerService{
			LoadBalancerID:  e.TerraformLink(),
			Protocol:        fi.String(service.Protocol),
			ListenPort:      service.ListenerPort,
			DestinationPort: service.DestinationPort,
		}

		err := t.RenderResource("hcloud_load_balancer_service", *e.Name, tf)
		if err != nil {
			return err
		}
	}

	{
		tf := &terraformLoadBalancerTarget{
			LoadBalancerID: e.TerraformLink(),
			Type:           fi.String(string(hcloud.LoadBalancerTargetTypeLabelSelector)),
			LabelSelector:  fi.String(e.Target),
			UsePrivateIP:   fi.Bool(true),
		}

		err := t.RenderResource("hcloud_load_balancer_target", *e.Name, tf)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *LoadBalancer) TerraformLink() *terraformWriter.Literal {
	return terraformWriter.LiteralProperty("hcloud_load_balancer", *e.Name, "id")
}
