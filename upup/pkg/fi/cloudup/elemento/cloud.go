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
	"fmt"

	"github.com/Elemento-Modular-Cloud/tesi-paolobeci/ecloud"
	"k8s.io/kops/upup/pkg/fi"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
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

	Region() string
	DNS() (dnsprovider.Interface, error)
	NetworkClient() ecloud.NetworkClient // TODO
	ServerClient() ecloud.ServerClient // TODO
	SSHKeyClient() ecloud.SSHKeyClient // TODO

	// TODO: Detect and add additional fields here
}

var _ fi.Cloud = &elementoCloudImplementation{}

// Interaction with Elemento cloud resources
type elementoCloudImplementation struct {
	Client *ecloud.Client
	region string
	dns    dnsprovider.Interface
	// TODO: Add additional fields here
}

func NewElementoCloud(region string) (ElementoCloud, error) {
	// Elemento does not use an access token, but is previously authenticated with 
	// the CLI and deamons, you must execute the Electros app to authenticate.

	client, err := ecloud.NewClient("kops-client", "0.1") // TODO

	if err != nil {
		return nil, fmt.Errorf("creating client for Elemento Cloud: %w", err)
	}

	return &elementoCloudImplementation{
		Client: client,
		dns:    nil,
		region: region,
	}, nil
}

func (c *elementoCloudImplementation) ProviderID() kops.CloudProviderID {
    return kops.CloudProviderElemento
}

func (c *elementoCloudImplementation) Region() string {
	return c.region
}

func (c *elementoCloudImplementation) DNS() (dnsprovider.Interface, error) {
    if c.dns == nil {
        return nil, fmt.Errorf("DNS provider is not initialized")
    }
    return c.dns, nil
}

func (c *elementoCloudImplementation) NetworkClient() ecloud.NetworkClient {
	return c.Client.Network
}

func (c *elementoCloudImplementation) ServerClient() ecloud.ServerClient {
	return c.Client.Server
}

func (c *elementoCloudImplementation) SSHKeyClient() ecloud.SSHKeyClient {
	return c.Client.SSHKey
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
	klog.V(8).Info("Elemento GetApiIngressStatus is not implemented")
	return nil, fmt.Errorf("GetApiIngressStatus is not implemented yet for Elemento")
}

func (s *elementoCloudImplementation) GetCloudGroups(cluster *kops.Cluster, instancegroups []*kops.InstanceGroup, warnUnmatched bool, nodes []v1.Node) (map[string]*cloudinstances.CloudInstanceGroup, error) {
	klog.V(8).Info("Elemento GetCloudGroups is not implemented")
	return nil, fmt.Errorf("GetCloudGroups is not implemented yet for Elemento")
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