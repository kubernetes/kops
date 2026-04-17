/*
Copyright 2026 The Kubernetes Authors.

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

package linodetasks

import (
	"context"
	"fmt"
	"net"
	"slices"
	"strconv"
	"strings"

	"github.com/linode/linodego"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/wellknownports"
	"k8s.io/kops/pkg/wellknownservices"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/linode"
)

// +kops:fitask
type LoadBalancer struct {
	Name      *string
	Lifecycle fi.Lifecycle

	ID     *int
	Region *string
	Tags   []string

	// WellKnownServices indicates which services are supported by this resource.
	// This field is internal and is not rendered to the cloud.
	WellKnownServices []wellknownservices.WellKnownService
}

var _ fi.CloudupTask = &LoadBalancer{}
var _ fi.CompareWithID = &LoadBalancer{}
var _ fi.HasLifecycle = &LoadBalancer{}
var _ fi.HasName = &LoadBalancer{}
var _ fi.HasAddress = &LoadBalancer{}

func (l *LoadBalancer) CompareWithID() *string {
	if l.ID == nil {
		return nil
	}
	id := strconv.Itoa(fi.ValueOf(l.ID))
	return fi.PtrTo(id)
}

func (l *LoadBalancer) GetLifecycle() fi.Lifecycle {
	return l.Lifecycle
}

func (l *LoadBalancer) SetLifecycle(lifecycle fi.Lifecycle) {
	l.Lifecycle = lifecycle
}

func (l *LoadBalancer) GetName() *string {
	return l.Name
}

func (l *LoadBalancer) String() string {
	return fi.CloudupTaskAsString(l)
}

// GetWellKnownServices returns the well-known services this load balancer provides.
func (l *LoadBalancer) GetWellKnownServices() []wellknownservices.WellKnownService {
	return l.WellKnownServices
}

func (l *LoadBalancer) FindAddresses(c *fi.CloudupContext) ([]string, error) {
	actual, err := l.Find(c)
	if err != nil {
		return nil, err
	}
	if actual == nil {
		return nil, nil
	}

	cloud := c.T.Cloud.(linode.LinodeCloud)
	nodebalancer, err := cloud.Client().GetNodeBalancer(c.Context(), fi.ValueOf(actual.ID))
	if err != nil {
		if linodego.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting Linode (Akamai) load balancer %q: %w", fi.ValueOf(l.Name), err)
	}
	if nodebalancer == nil || nodebalancer.IPv4 == nil || *nodebalancer.IPv4 == "" {
		return nil, nil
	}

	return []string{*nodebalancer.IPv4}, nil
}

func (l *LoadBalancer) Find(c *fi.CloudupContext) (*LoadBalancer, error) {
	cloud := c.T.Cloud.(linode.LinodeCloud)
	nodebalancers, err := cloud.Client().ListNodeBalancers(c.Context(), nil)
	if err != nil {
		return nil, fmt.Errorf("error listing Linode (Akamai) load balancers: %w", err)
	}

	taskLabel := linode.NormalizedLoadBalancerLabel(fi.ValueOf(l.Name))
	for i := range nodebalancers {
		nb := nodebalancers[i]
		if fi.ValueOf(nb.Label) != taskLabel {
			continue
		}

		actual := &LoadBalancer{
			// Preserve desired task identity to avoid a synthetic Name change when
			// the cloud label is normalized from the desired DNS-like name.
			Name:              l.Name,
			Lifecycle:         l.Lifecycle,
			ID:                fi.PtrTo(nb.ID),
			Region:            fi.PtrTo(nb.Region),
			Tags:              slices.Clone(nb.Tags),
			WellKnownServices: slices.Clone(l.WellKnownServices),
		}
		l.ID = actual.ID
		return actual, nil
	}

	return nil, nil
}

func (l *LoadBalancer) Run(c *fi.CloudupContext) error {
	if err := fi.CloudupDefaultDeltaRunMethod(l, c); err != nil {
		return err
	}

	if l.ID == nil {
		// LB does not yet exist (e.g., dry-run); nothing to reconcile.
		return nil
	}

	cloud := c.T.Cloud.(linode.LinodeCloud)
	backends, err := linodeDiscoverControlPlaneBackends(cloud.Client(), l.Tags)
	if err != nil {
		return err
	}
	if len(backends) == 0 {
		return nil
	}

	return ensureLoadBalancerConfigs(cloud.Client(), fi.ValueOf(l.ID), fi.ValueOf(l.Name), backends)
}

func (_ *LoadBalancer) CheckChanges(a, e, changes *LoadBalancer) error {
	if a != nil {
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
		if changes.ID != nil {
			return fi.CannotChangeField("ID")
		}
		if changes.Region != nil {
			return fi.CannotChangeField("Region")
		}
	} else {
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
		if e.Region == nil {
			return fi.RequiredField("Region")
		}
	}

	return nil
}

func (_ *LoadBalancer) RenderLinode(t *linode.APITarget, a, e, changes *LoadBalancer) error {
	backends, err := linodeDiscoverControlPlaneBackends(t.Cloud.Client(), e.Tags)
	if err != nil {
		return err
	}

	label := linode.NormalizedLoadBalancerLabel(fi.ValueOf(e.Name))

	if a == nil {
		nb, err := t.Cloud.Client().CreateNodeBalancer(context.Background(), linodego.NodeBalancerCreateOptions{
			Label:   fi.PtrTo(label),
			Region:  fi.ValueOf(e.Region),
			Tags:    slices.Clone(e.Tags),
			Type:    linodego.NBTypeCommon,
			Configs: nil,
		})
		if err != nil {
			return fmt.Errorf("error creating Linode (Akamai) load balancer %q: %w", fi.ValueOf(e.Name), err)
		}

		e.ID = fi.PtrTo(nb.ID)
		if len(backends) == 0 {
			return nil
		}

		return ensureLoadBalancerConfigs(t.Cloud.Client(), nb.ID, fi.ValueOf(e.Name), backends)
	}

	if len(backends) == 0 {
		return nil
	}

	return ensureLoadBalancerConfigs(t.Cloud.Client(), fi.ValueOf(a.ID), fi.ValueOf(e.Name), backends)
}

func ensureLoadBalancerConfigs(client linode.LinodeClient, nodebalancerID int, name string, backends []string) error {
	configs, err := client.ListNodeBalancerConfigs(context.Background(), nodebalancerID, nil)
	if err != nil {
		return fmt.Errorf("error listing Linode (Akamai) load balancer configs for %q: %w", name, err)
	}

	configByPort := map[int]linodego.NodeBalancerConfig{}
	for _, cfg := range configs {
		configByPort[cfg.Port] = cfg
	}

	for _, port := range []int{wellknownports.KubeAPIServer, wellknownports.KopsControllerPort} {
		nodes := linodeCreateNodeOptions(backends, port)
		var configID int
		if cfg, found := configByPort[port]; found {
			configID = cfg.ID
			_, err := client.RebuildNodeBalancerConfig(context.Background(), nodebalancerID, cfg.ID, linodego.NodeBalancerConfigRebuildOptions{
				Port:      port,
				Protocol:  linodego.ProtocolTCP,
				Check:     linodego.CheckConnection,
				Algorithm: linodego.AlgorithmRoundRobin,
				Nodes:     linodeCreateRebuildNodeOptions(nodes),
			})
			if err != nil {
				return fmt.Errorf("error rebuilding Linode (Akamai) load balancer config for port %d: %w", port, err)
			}
		} else {
			createdConfig, err := client.CreateNodeBalancerConfig(context.Background(), nodebalancerID, *linodeCreateTCPConfig(port, backends))
			if err != nil {
				return fmt.Errorf("error creating Linode (Akamai) load balancer config for port %d: %w", port, err)
			}
			configID = createdConfig.ID
		}

		if err := ensureLoadBalancerConfigNodes(client, nodebalancerID, configID, nodes); err != nil {
			return fmt.Errorf("error reconciling Linode (Akamai) load balancer nodes for port %d: %w", port, err)
		}
	}

	return nil
}

func ensureLoadBalancerConfigNodes(client linode.LinodeClient, nodebalancerID int, configID int, desiredNodes []linodego.NodeBalancerNodeCreateOptions) error {
	existingNodes, err := client.ListNodeBalancerNodes(context.Background(), nodebalancerID, configID, nil)
	if err != nil {
		return fmt.Errorf("error listing Linode (Akamai) load balancer config nodes: %w", err)
	}

	desiredByAddress := make(map[string]linodego.NodeBalancerNodeCreateOptions, len(desiredNodes))
	for _, desired := range desiredNodes {
		desiredByAddress[desired.Address] = desired
	}

	for _, existing := range existingNodes {
		desired, found := desiredByAddress[existing.Address]
		if !found {
			continue
		}

		if existing.Label != desired.Label || existing.Mode != desired.Mode || existing.Weight != desired.Weight {
			_, err := client.UpdateNodeBalancerNode(context.Background(), nodebalancerID, configID, existing.ID, linodego.NodeBalancerNodeUpdateOptions{
				Address: desired.Address,
				Label:   desired.Label,
				Mode:    desired.Mode,
				Weight:  desired.Weight,
			})
			if err != nil {
				return fmt.Errorf("error updating Linode (Akamai) load balancer node %q: %w", existing.Address, err)
			}
		}

		delete(desiredByAddress, existing.Address)
	}

	for _, desired := range desiredByAddress {
		if _, err := client.CreateNodeBalancerNode(context.Background(), nodebalancerID, configID, desired); err != nil {
			return fmt.Errorf("error creating Linode (Akamai) load balancer node %q: %w", desired.Address, err)
		}
	}

	return nil
}

func linodeDiscoverControlPlaneBackends(client linode.LinodeClient, tags []string) ([]string, error) {
	instances, err := client.ListInstances(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("error listing Linode (Akamai) instances for load balancer backends: %w", err)
	}

	clusterTag := extractClusterTag(tags)

	controlPlaneTag := linode.BuildLinodeTag(linode.TagKubernetesInstanceRole, string(kops.InstanceGroupRoleControlPlane))
	apiServerTag := linode.BuildLinodeTag(linode.TagKubernetesInstanceRole, string(kops.InstanceGroupRoleAPIServer))

	var backends []string
	for _, instance := range instances {
		if clusterTag != "" && !slices.Contains(instance.Tags, clusterTag) {
			continue
		}
		if !slices.Contains(instance.Tags, controlPlaneTag) && !slices.Contains(instance.Tags, apiServerTag) {
			continue
		}

		if ip := selectPrivateIPv4(instance.IPv4); ip != "" {
			backends = append(backends, ip)
		}
	}

	return backends, nil
}

func extractClusterTag(tags []string) string {
	for _, tag := range tags {
		if strings.HasPrefix(tag, kops.LabelClusterName+":") {
			return tag
		}
	}

	return ""
}

func linodeCreateTCPConfig(port int, backends []string) *linodego.NodeBalancerConfigCreateOptions {
	return &linodego.NodeBalancerConfigCreateOptions{
		Port:      port,
		Protocol:  linodego.ProtocolTCP,
		Check:     linodego.CheckConnection,
		Algorithm: linodego.AlgorithmRoundRobin,
		Nodes:     linodeCreateNodeOptions(backends, port),
	}
}

func linodeCreateNodeOptions(backends []string, port int) []linodego.NodeBalancerNodeCreateOptions {
	nodes := make([]linodego.NodeBalancerNodeCreateOptions, 0, len(backends))
	for _, ip := range backends {
		nodes = append(nodes, linodego.NodeBalancerNodeCreateOptions{
			Address: net.JoinHostPort(ip, strconv.Itoa(port)),
			Label:   fmt.Sprintf("cp-%s-%d", ip, port),
			Mode:    linodego.ModeAccept,
			Weight:  100,
		})
	}
	return nodes
}

func linodeCreateRebuildNodeOptions(nodes []linodego.NodeBalancerNodeCreateOptions) []linodego.NodeBalancerConfigRebuildNodeOptions {
	out := make([]linodego.NodeBalancerConfigRebuildNodeOptions, 0, len(nodes))
	for _, n := range nodes {
		out = append(out, linodego.NodeBalancerConfigRebuildNodeOptions{NodeBalancerNodeCreateOptions: n})
	}
	return out
}

func selectPrivateIPv4(ips []*net.IP) string {
	return linode.SelectIPv4(ips, true)
}
