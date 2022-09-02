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

package hetzner

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/hetznercloud/hcloud-go/hcloud"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/upup/pkg/fi"
)

const (
	TagKubernetesClusterName         = "kops.k8s.io/cluster"
	TagKubernetesFirewallRole        = "kops.k8s.io/firewall-role"
	TagKubernetesInstanceGroup       = "kops.k8s.io/instance-group"
	TagKubernetesInstanceRole        = "kops.k8s.io/instance-role"
	TagKubernetesInstanceUserData    = "kops.k8s.io/instance-userdata"
	TagKubernetesInstanceNeedsUpdate = "kops.k8s.io/needs-update"
	TagKubernetesVolumeRole          = "kops.k8s.io/volume-role"
)

// HetznerCloud exposes all the interfaces required to operate on Hetzner Cloud resources
type HetznerCloud interface {
	fi.Cloud
	ActionClient() hcloud.ActionClient
	SSHKeyClient() hcloud.SSHKeyClient
	NetworkClient() hcloud.NetworkClient
	LoadBalancerClient() hcloud.LoadBalancerClient
	FirewallClient() hcloud.FirewallClient
	ServerClient() hcloud.ServerClient
	VolumeClient() hcloud.VolumeClient
	GetSSHKeys(clusterName string) ([]*hcloud.SSHKey, error)
	GetNetworks(clusterName string) ([]*hcloud.Network, error)
	GetFirewalls(clusterName string) ([]*hcloud.Firewall, error)
	GetLoadBalancers(clusterName string) ([]*hcloud.LoadBalancer, error)
	GetServers(clusterName string) ([]*hcloud.Server, error)
	GetVolumes(clusterName string) ([]*hcloud.Volume, error)
}

// static compile time check to validate HetznerCloud's fi.Cloud Interface.
var _ fi.Cloud = &hetznerCloudImplementation{}

// hetznerCloudImplementation holds the godo client object to interact with Hetzner resources.
type hetznerCloudImplementation struct {
	Client *hcloud.Client

	dns dnsprovider.Interface

	region string
}

// NewHetznerCloud returns a Cloud, using the env var HCLOUD_TOKEN
func NewHetznerCloud(region string) (HetznerCloud, error) {
	accessToken := os.Getenv("HCLOUD_TOKEN")
	if accessToken == "" {
		return nil, errors.New("HCLOUD_TOKEN is required")
	}

	opts := []hcloud.ClientOption{
		hcloud.WithToken(accessToken),
	}
	client := hcloud.NewClient(opts...)

	return &hetznerCloudImplementation{
		Client: client,
		dns:    nil,
		region: region,
	}, nil
}

// ActionClient returns an implementation of hetzner.ActionClient
func (c *hetznerCloudImplementation) ActionClient() hcloud.ActionClient {
	return c.Client.Action
}

// SSHKeyClient returns an implementation of hetzner.SSHKeyClient
func (c *hetznerCloudImplementation) SSHKeyClient() hcloud.SSHKeyClient {
	return c.Client.SSHKey
}

// NetworkClient returns an implementation of hetzner.NetworkClient
func (c *hetznerCloudImplementation) NetworkClient() hcloud.NetworkClient {
	return c.Client.Network
}

// FirewallClient returns an implementation of hetzner.FirewallClient
func (c *hetznerCloudImplementation) FirewallClient() hcloud.FirewallClient {
	return c.Client.Firewall
}

// LoadBalancerClient returns an implementation of hetzner.LoadBalancerClient
func (c *hetznerCloudImplementation) LoadBalancerClient() hcloud.LoadBalancerClient {
	return c.Client.LoadBalancer
}

// ServerClient returns an implementation of hetzner.ServerClient
func (c *hetznerCloudImplementation) ServerClient() hcloud.ServerClient {
	return c.Client.Server
}

// VolumeClient returns an implementation of hetzner.VolumeClient
func (c *hetznerCloudImplementation) VolumeClient() hcloud.VolumeClient {
	return c.Client.Volume
}

func (c *hetznerCloudImplementation) GetSSHKeys(clusterName string) ([]*hcloud.SSHKey, error) {
	client := c.SSHKeyClient()

	labelSelector := TagKubernetesClusterName + "=" + clusterName
	listOptions := hcloud.ListOpts{
		PerPage:       50,
		LabelSelector: labelSelector,
	}
	sshKeyListOpts := hcloud.SSHKeyListOpts{ListOpts: listOptions}

	matches, err := client.AllWithOpts(context.TODO(), sshKeyListOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to get SSH keys matching label selector %q: %w", labelSelector, err)
	}

	return matches, nil
}

func (c *hetznerCloudImplementation) GetNetworks(clusterName string) ([]*hcloud.Network, error) {
	client := c.NetworkClient()

	labelSelector := TagKubernetesClusterName + "=" + clusterName
	listOptions := hcloud.ListOpts{
		PerPage:       50,
		LabelSelector: labelSelector,
	}
	networkListOptions := hcloud.NetworkListOpts{ListOpts: listOptions}

	matches, err := client.AllWithOpts(context.TODO(), networkListOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to get networks matching label selector %q: %w", labelSelector, err)
	}

	return matches, nil
}

func (c *hetznerCloudImplementation) GetFirewalls(clusterName string) ([]*hcloud.Firewall, error) {
	client := c.FirewallClient()

	labelSelector := TagKubernetesClusterName + "=" + clusterName
	listOptions := hcloud.ListOpts{
		PerPage:       50,
		LabelSelector: labelSelector,
	}
	firewallListOptions := hcloud.FirewallListOpts{ListOpts: listOptions}

	matches, err := client.AllWithOpts(context.TODO(), firewallListOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to get firewalls matching label selector %q: %w", labelSelector, err)
	}

	return matches, nil
}

func (c *hetznerCloudImplementation) GetLoadBalancers(clusterName string) ([]*hcloud.LoadBalancer, error) {
	client := c.LoadBalancerClient()

	labelSelector := TagKubernetesClusterName + "=" + clusterName
	listOptions := hcloud.ListOpts{
		PerPage:       50,
		LabelSelector: labelSelector,
	}
	loadBalancerListOptions := hcloud.LoadBalancerListOpts{ListOpts: listOptions}

	matches, err := client.AllWithOpts(context.TODO(), loadBalancerListOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to get load balancers matching label selector %q: %w", labelSelector, err)
	}

	return matches, nil
}

func (c *hetznerCloudImplementation) GetServers(clusterName string) ([]*hcloud.Server, error) {
	client := c.ServerClient()

	labelSelector := TagKubernetesClusterName + "=" + clusterName
	listOptions := hcloud.ListOpts{
		PerPage:       50,
		LabelSelector: labelSelector,
	}
	sortOptions := []string{
		"created:desc",
	}
	serverListOptions := hcloud.ServerListOpts{ListOpts: listOptions, Sort: sortOptions}

	matches, err := client.AllWithOpts(context.TODO(), serverListOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to get servers matching label selector %q: %w", labelSelector, err)
	}

	return matches, nil
}

func (c *hetznerCloudImplementation) GetVolumes(clusterName string) ([]*hcloud.Volume, error) {
	client := c.VolumeClient()

	labelSelector := TagKubernetesClusterName + "=" + clusterName
	listOptions := hcloud.ListOpts{
		PerPage:       50,
		LabelSelector: labelSelector,
	}
	volumeListOptions := hcloud.VolumeListOpts{ListOpts: listOptions}

	matches, err := client.AllWithOpts(context.TODO(), volumeListOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to get volumes matching label selector %q: %w", labelSelector, err)
	}

	return matches, nil
}

func (c *hetznerCloudImplementation) DNS() (dnsprovider.Interface, error) {
	// TODO(hakman): implement me
	panic("implement me")
}

func (c *hetznerCloudImplementation) DeleteInstance(instance *cloudinstances.CloudInstance) error {
	serverID, err := strconv.Atoi(instance.ID)
	if err != nil {
		return fmt.Errorf("failed to convert server ID %q to int: %w", instance.ID, err)
	}

	err = deleteServer(c, serverID)

	return err
}

// deleteServer shuts down and deletes the given server
func deleteServer(c *hetznerCloudImplementation, id int) error {
	client := c.ServerClient()
	ctx := context.TODO()

	server := &hcloud.Server{ID: id}
	_, _, err := client.Shutdown(ctx, server)
	if err != nil {
		return fmt.Errorf("failed to stop server %q: %w", strconv.Itoa(id), err)
	}

	for i := 1; i <= 30; i++ {
		server, _, err := client.GetByID(ctx, id)
		if err != nil || server == nil {
			return fmt.Errorf("failed to get info for server %q: %w", strconv.Itoa(id), err)
		}

		if server.Status == hcloud.ServerStatusOff {
			break
		}

		time.Sleep(time.Second * 5)
	}

	_, err = client.Delete(ctx, server)
	if err != nil {
		return fmt.Errorf("failed to delete server %q: %w", strconv.Itoa(id), err)
	}

	return nil
}

func (c *hetznerCloudImplementation) DetachInstance(instance *cloudinstances.CloudInstance) error {
	// Hetzner Cloud API doesn't provide the option of using cloud groups
	return nil
}

// ProviderID returns the kOps API identifier for Hetzner Cloud
func (c *hetznerCloudImplementation) ProviderID() kops.CloudProviderID {
	return kops.CloudProviderHetzner
}

// Region returns the Hetzner Cloud region
func (c *hetznerCloudImplementation) Region() string {
	return c.region
}

func (c *hetznerCloudImplementation) GetCloudGroups(cluster *kops.Cluster, instanceGroups []*kops.InstanceGroup, warnUnmatched bool, nodes []v1.Node) (map[string]*cloudinstances.CloudInstanceGroup, error) {
	nodeMap := cloudinstances.GetNodeMap(nodes, cluster)

	serverGroups, err := findServerGroups(c, cluster.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to find server groups: %v", err)
	}

	cloudInstanceGroups := make(map[string]*cloudinstances.CloudInstanceGroup)
	for name, serverGroup := range serverGroups {
		var instanceGroup *kops.InstanceGroup
		for _, ig := range instanceGroups {
			groupName := fmt.Sprintf("%s-%s", cluster.Name, ig.Name)
			if name == groupName {
				instanceGroup = ig
				break
			}
		}
		if instanceGroup == nil {
			if warnUnmatched {
				klog.Warningf("Server group %q has no corresponding instance group", name)
			}
			continue
		}

		cloudInstanceGroups[instanceGroup.Name], err = buildCloudInstanceGroup(instanceGroup, serverGroup, nodeMap)
		if err != nil {
			return nil, fmt.Errorf("failed to build cloud instance group for instance group %q: %w", instanceGroup.Name, err)
		}
	}

	return cloudInstanceGroups, nil
}

// findServerGroups finds all server groups belonging to the cluster
func findServerGroups(c *hetznerCloudImplementation, clusterName string) (map[string][]*hcloud.Server, error) {
	servers, err := c.GetServers(clusterName)
	if err != nil {
		return nil, fmt.Errorf("failed to list servers: %w", err)
	}

	serverGroups := make(map[string][]*hcloud.Server)
	for _, server := range servers {
		instanceGroupNameLabel, ok := server.Labels[TagKubernetesInstanceGroup]
		if !ok {
			klog.Warning("failed to find instance group name for server %s(%d)", server.Name, server.ID)
			continue
		}

		instanceGroupName := fmt.Sprintf("%s-%s", clusterName, instanceGroupNameLabel)
		serverGroups[instanceGroupName] = append(serverGroups[instanceGroupName], server)
	}

	return serverGroups, nil
}

func buildCloudInstanceGroup(ig *kops.InstanceGroup, sg []*hcloud.Server, nodeMap map[string]*v1.Node) (*cloudinstances.CloudInstanceGroup, error) {
	cloudInstanceGroup := &cloudinstances.CloudInstanceGroup{
		HumanName:     ig.Name,
		InstanceGroup: ig,
		Raw:           sg,
		MinSize:       int(fi.Int32Value(ig.Spec.MinSize)),
		TargetSize:    int(fi.Int32Value(ig.Spec.MinSize)),
		MaxSize:       int(fi.Int32Value(ig.Spec.MaxSize)),
	}

	for _, server := range sg {
		status := cloudinstances.CloudInstanceStatusUpToDate
		if _, ok := server.Labels[TagKubernetesInstanceNeedsUpdate]; ok {
			status = cloudinstances.CloudInstanceStatusNeedsUpdate
		}

		id := strconv.Itoa(server.ID)
		cloudInstance, err := cloudInstanceGroup.NewCloudInstance(id, status, nodeMap[id])
		if err != nil {
			return nil, fmt.Errorf("failed to create cloud group instance for server %s(%d): %w", server.Name, server.ID, err)
		}

		// Add additional instance info
		cloudInstance.State = cloudinstances.State(server.Status)
		if role, ok := server.Labels[TagKubernetesInstanceRole]; ok {
			cloudInstance.Roles = append(cloudInstance.Roles, role)
		}
		if server.ServerType != nil {
			cloudInstance.MachineType = server.ServerType.Name
		}
		if len(server.PrivateNet) > 0 {
			cloudInstance.PrivateIP = server.PrivateNet[0].IP.String()
		}
	}

	return cloudInstanceGroup, nil
}

func (c *hetznerCloudImplementation) DeleteGroup(g *cloudinstances.CloudInstanceGroup) error {
	for _, cloudInstance := range append(g.NeedUpdate, g.Ready...) {
		serverID, err := strconv.Atoi(cloudInstance.ID)
		if err != nil {
			return fmt.Errorf("failed to convert server ID %q to int: %w", cloudInstance.ID, err)
		}

		err = deleteServer(c, serverID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *hetznerCloudImplementation) DeregisterInstance(i *cloudinstances.CloudInstance) error {
	// Hetzner Cloud API doesn't provide the option to drain and remove an instance from the associated load balancers
	return nil
}

// FindVPCInfo is not implemented
func (c *hetznerCloudImplementation) FindVPCInfo(id string) (*fi.VPCInfo, error) {
	// TODO(hakman): Implement me
	return nil, errors.New("hetzner cloud provider does not implement FindVPCInfo at this time")
}

// FindClusterStatus was used before etcd-manager to check the etcd cluster status and prevent unsupported changes.
func (c *hetznerCloudImplementation) FindClusterStatus(cluster *kops.Cluster) (*kops.ClusterStatus, error) {
	return nil, nil
}

func (c *hetznerCloudImplementation) GetApiIngressStatus(cluster *kops.Cluster) ([]fi.ApiIngressStatus, error) {
	if cluster.Spec.MasterPublicName == "" {
		return nil, nil
	}

	lbName := "api." + cluster.Name

	client := c.LoadBalancerClient()
	// TODO(hakman): Get load balancer info using label selector instead instead of name?
	lb, _, err := client.GetByName(context.TODO(), lbName)
	if err != nil {
		return nil, fmt.Errorf("failed to get info for load balancer %q: %w", lbName, err)
	}
	if lb == nil {
		return nil, nil
	}

	if !lb.PublicNet.Enabled {
		return nil, fmt.Errorf("load balancer %s(%d) is not public", lb.Name, lb.ID)
	}

	ingresses := []fi.ApiIngressStatus{
		{
			IP: lb.PublicNet.IPv4.IP.String(),
		},
	}

	return ingresses, nil
}
