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

package elemento

import (
	"context"
	"fmt"
	"os"

	"github.com/Elemento-Modular-Cloud/ecloud-go/ecloud"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
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
	TagKubernetesNodeLabelPrefix     = "node-label.kops.k8s.io."
)

// ElementoCloud exposes all the interfaces required to operate on the Elemento cloud
type ElementoCloud interface {
	fi.Cloud

	NetworkClient() ecloud.NetworkClient
	ServerClient() ecloud.ServerClient
	SSHKeyClient() ecloud.SSHKeyClient
	VolumeClient() ecloud.VolumeClient
	NodeupClient(ctx context.Context) ecloud.NodeupClient

	// Add DNS-zone / DNS-record client accessors here once
	// the Elemento SDK exposes them. The provider-native DNS tasks in
	// elementotasks/dns_record.go are the intended callers.
	DnsClient() ecloud.DnsClient
}

var _ fi.Cloud = &elementoCloudImplementation{}

// Interaction with Elemento cloud resources
type elementoCloudImplementation struct {
	Client *ecloud.Client
	region string
	// TODO: Add additional fields here
}

func NewElementoCloud(region string) (ElementoCloud, error) {
	// Elemento does not use an access token, but is previously authenticated with
	// the CLI and deamons, you must execute the Electros app to authenticate.

	klog.V(2).Infof("Creating ecloud client for region %s", region)

	// Here is the entrypoint of the ecloud-go SDK, executed by building a new ecloud-client
	client, err := ecloud.NewClient("kops-client", "0.1")

	if err != nil {
		klog.Errorf("Failed to create ecloud client: %v", err)
		return nil, fmt.Errorf("creating client for Elemento Cloud: %w", err)
	}

	klog.V(2).Infof("Successfully created ecloud client")

	return &elementoCloudImplementation{
		Client: client,
		region: region,
	}, nil
}

func (c *elementoCloudImplementation) ProviderID() kops.CloudProviderID {
	return kops.CloudProviderElemento
}

func (c *elementoCloudImplementation) Region() string {
	return c.region
}

func (c *elementoCloudImplementation) GetCloudGroups(cluster *kops.Cluster, instanceGroups []*kops.InstanceGroup, warnUnmatched bool, nodes []v1.Node) (map[string]*cloudinstances.CloudInstanceGroup, error) {
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
func findServerGroups(c *elementoCloudImplementation, clusterName string) (map[string][]*ecloud.Server, error) {
	servers, err := c.GetServers(clusterName)
	if err != nil {
		return nil, fmt.Errorf("failed to list servers: %w", err)
	}

	serverGroups := make(map[string][]*ecloud.Server)
	for _, server := range servers {
		instanceGroupNameLabel, ok := server.Labels[TagKubernetesInstanceGroup]
		if !ok {
			klog.Warningf("failed to find instance group name for server %s(%s)", server.Name, server.ID)
			continue
		}

		instanceGroupName := fmt.Sprintf("%s-%s", clusterName, instanceGroupNameLabel)
		serverGroups[instanceGroupName] = append(serverGroups[instanceGroupName], server)
	}

	return serverGroups, nil
}

func (c *elementoCloudImplementation) DNS() ecloud.DnsClient {
	// Elemento DNS is expected to be managed through provider-native cloudup tasks,
	return c.Client.Dns
}

func (c *elementoCloudImplementation) NetworkClient() ecloud.NetworkClient {
	return c.Client.Network
}

func (c *elementoCloudImplementation) ServerClient() ecloud.ServerClient {
	klog.V(2).Infof("ECLOUD_DEBUG: Returning ServerClient instance")
	return c.Client.Server
}

func (c *elementoCloudImplementation) SSHKeyClient() ecloud.SSHKeyClient {
	return c.Client.SSHKey
}

func (c *elementoCloudImplementation) VolumeClient() ecloud.VolumeClient {
	return c.Client.Volume
}

func (c *elementoCloudImplementation) NodeupClient(ctx context.Context) ecloud.NodeupClient {
	return c.Client.Nodeup
}

func buildCloudInstanceGroup(ig *kops.InstanceGroup, sg []*ecloud.Server, nodeMap map[string]*v1.Node) (*cloudinstances.CloudInstanceGroup, error) {
	cloudInstanceGroup := &cloudinstances.CloudInstanceGroup{
		HumanName:     ig.Name,
		InstanceGroup: ig,
		Raw:           sg,
		MinSize:       int(fi.ValueOf(ig.Spec.MinSize)),
		TargetSize:    int(fi.ValueOf(ig.Spec.MinSize)),
		MaxSize:       int(fi.ValueOf(ig.Spec.MaxSize)),
	}

	for _, server := range sg {
		status := cloudinstances.CloudInstanceStatusUpToDate
		if _, ok := server.Labels[TagKubernetesInstanceNeedsUpdate]; ok {
			status = cloudinstances.CloudInstanceStatusNeedsUpdate
		}

		cloudInstance, err := cloudInstanceGroup.NewCloudInstance(server.ID, status, nodeMap[server.ID])
		if err != nil {
			return nil, fmt.Errorf("failed to create cloud group instance for server %s(%s): %w", server.Name, server.ID, err)
		}

		// Add additional instance info
		cloudInstance.State = cloudinstances.State(server.Status)
		if role, ok := server.Labels[TagKubernetesInstanceRole]; ok {
			cloudInstance.Roles = append(cloudInstance.Roles, role)
		}
		if server.ServerType != nil {
			cloudInstance.MachineType = server.ServerType.Name
		}
	}

	return cloudInstanceGroup, nil
}

func (s *elementoCloudImplementation) DeleteGroup(group *cloudinstances.CloudInstanceGroup) error {
	toDelete := append(group.NeedUpdate, group.Ready...)
	for _, cloudInstance := range toDelete {
		err := s.DeleteInstance(cloudInstance)
		if err != nil {
			return fmt.Errorf("error deleting group %q: %w", group.HumanName, err)
		}
	}
	return nil
}

func (c *elementoCloudImplementation) GetServers(clusterName string) ([]*ecloud.Server, error) {
	client := c.ServerClient()

	labelSelector := TagKubernetesClusterName + "=" + clusterName
	listOptions := ecloud.ListOpts{
		PerPage:       50,
		LabelSelector: labelSelector,
	}
	serverListOptions := ecloud.ServerListOpts{ListOpts: listOptions}

	matches, _, err := client.List(context.TODO(), serverListOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to get servers matching label selector %q: %w", labelSelector, err)
	}

	return matches, nil
}

// TODO: All the following functions are not implemented yet

func (s *elementoCloudImplementation) DeleteInstance(i *cloudinstances.CloudInstance) error {
	klog.V(8).Infof("Elemento DeleteInstance is not implemented yet")
	return fmt.Errorf("DeleteInstance is not implemented yet for Elemento")
}

func (s *elementoCloudImplementation) DeregisterInstance(i *cloudinstances.CloudInstance) error {
	klog.V(8).Infof("Elemento DeregisterInstance is not implemented yet")
	return fmt.Errorf("DeregisterInstance is not implemented yet for Elemento")
}

func (s *elementoCloudImplementation) DetachInstance(i *cloudinstances.CloudInstance) error {
	klog.V(8).Infof("Elemento DetachInstance is not implemented yet")
	return fmt.Errorf("DetachInstance is not implemented yet for Elemento")
}

// FindClusterStatus was used before etcd-manager to check the etcd cluster status and prevent unsupported changes.
func (s *elementoCloudImplementation) FindClusterStatus(cluster *kops.Cluster) (*kops.ClusterStatus, error) {
	klog.V(8).Info("Elemento FindClusterStatus is not implemented")
	return nil, nil
}

func (s *elementoCloudImplementation) FindVPCInfo(id string) (*fi.VPCInfo, error) {
	klog.V(8).Info("Elemento clusters don't have a VPC yet so FindVPCInfo is not implemented")
	return nil, fmt.Errorf("FindVPCInfo is not implemented yet for Elemento")
}

func (s *elementoCloudImplementation) GetApiIngressStatus(cluster *kops.Cluster) ([]fi.ApiIngressStatus, error) {
	// klog.V(8).Info("Elemento GetApiIngressStatus is not implemented")
	// return nil, fmt.Errorf("GetApiIngressStatus is not implemented yet for Elemento")
	klog.V(8).Info("Elemento GetApiIngressStatus returning mock data")

	//! Mock response with reasonable API ingress statuses
	return []fi.ApiIngressStatus{
		{
			IP:       os.Getenv("ATOMOS_SERVER"), // Mock internal IP
			Hostname: "api.internal.cluster.local",
		},
	}, nil
}

// func findServerGroups(s *elementoCloudImplementation, clusterName string) (map[string][]*instance.Server, error) {
// 	klog.V(8).Info("Elemento findServerGroups is not implemented")
// 	return nil, fmt.Errorf("findServerGroups is not implemented yet for Elemento")
// }

// func buildCloudGroup(s *elementoCloudImplementation, ig *kops.InstanceGroup, sg []*instance.Server, nodeMap map[string]*v1.Node) (*cloudinstances.CloudInstanceGroup, error) {
// 	klog.V(8).Info("Elemento buildCloudGroup is not implemented")
// 	return nil, fmt.Errorf("buildCloudGroup is not implemented yet for Elemento")
// }

// func (s *elementoCloudImplementation) GetClusterDNSRecords(clusterName string) ([]*domain.Record, error) {
// 	klog.V(8).Info("Elemento GetClusterDNSRecords is not implemented")
// 	return nil, fmt.Errorf("GetClusterDNSRecords is not implemented yet for Elemento")
// }

// func (s *elementoCloudImplementation) GetClusterLoadBalancers(clusterName string) ([]*lb.LB, error) {
// 	klog.V(8).Info("Elemento GetClusterLoadBalancers is not implemented")
// 	return nil, fmt.Errorf("GetClusterLoadBalancers is not implemented yet for Elemento")
// }

// func (s *elementoCloudImplementation) GetClusterServers(clusterName string, instanceGroupName *string) ([]*instance.Server, error) {
// 	klog.V(8).Info("Elemento GetClusterServers is not implemented")
// 	return nil, fmt.Errorf("GetClusterServers is not implemented yet for Elemento")
// }

// func (s *elementoCloudImplementation) GetClusterSSHKeys(clusterName string) ([]*iam.SSHKey, error) {
// 	klog.V(8).Info("Elemento GetClusterSSHKeys is not implemented")
// 	return nil, fmt.Errorf("GetClusterSSHKeys is not implemented yet for Elemento")
// }

// func (s *elementoCloudImplementation) GetClusterVolumes(clusterName string) ([]*instance.Volume, error) {
// 	klog.V(8).Info("Elemento GetClusterVolumes is not implemented")
// 	return nil, fmt.Errorf("GetClusterVolumes is not implemented yet for Elemento")
// }

// func (s *elementoCloudImplementation) GetServerIP(serverID string, zone scw.Zone) (string, error) {
// 	klog.V(8).Info("Elemento GetServerIP is not implemented")
// 	return "", fmt.Errorf("GetServerIP is not implemented yet for Elemento")
// }

// func (s *elementoCloudImplementation) DeleteDNSRecord(record *domain.Record, clusterName string) error {
// 	klog.V(8).Info("Elemento DeleteDNSRecord is not implemented")
// 	return fmt.Errorf("DeleteDNSRecord is not implemented yet for Elemento")
// }

// func (s *elementoCloudImplementation) DeleteLoadBalancer(loadBalancer *lb.LB) error {
// 	klog.V(8).Info("Elemento DeleteLoadBalancer is not implemented")
// 	return fmt.Errorf("DeleteLoadBalancer is not implemented yet for Elemento")
// }

// func (s *elementoCloudImplementation) DeleteServer(server *instance.Server) error {
// 	klog.V(8).Info("Elemento DeleteServer is not implemented")
// 	return fmt.Errorf("DeleteServer is not implemented yet for Elemento")
// }

// func (s *elementoCloudImplementation) DeleteSSHKey(sshkey *iam.SSHKey) error {
// 	klog.V(8).Info("Elemento DeleteSSHKey is not implemented")
// 	return fmt.Errorf("DeleteSSHKey is not implemented yet for Elemento")
// }

// func (s *elementoCloudImplementation) DeleteVolume(volume *instance.Volume) error {
// 	klog.V(8).Info("Elemento DeleteVolume is not implemented")
// 	return fmt.Errorf("DeleteVolume is not implemented yet for Elemento")
// }
