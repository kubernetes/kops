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

package scaleway

import (
	"fmt"
	"os"
	"strings"

	domain "github.com/scaleway/scaleway-sdk-go/api/domain/v2beta1"
	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/api/vpc/v1"
	"github.com/scaleway/scaleway-sdk-go/api/vpcgw/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/klog/v2"
	kopsv "k8s.io/kops"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	dns "k8s.io/kops/dnsprovider/pkg/dnsprovider/providers/scaleway"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/upup/pkg/fi"
)

const (
	TagClusterName           = "kops.k8s.io/cluster"
	KopsUserAgentPrefix      = "kubernetes-kops/"
	TagInstanceGroup         = "instance-group"
	TagNameRolePrefix        = "k8s.io/role"
	TagNameEtcdClusterPrefix = "k8s.io/etcd"
	TagRoleControlPlane      = "ControlPlane" // changed from 'control-plane' to match kops.InstanceGroupRoleControlPlane
	TagRoleWorker            = "worker"
	TagRoleLoadBalancer      = "load-balancer"
	TagNeedsUpdate           = "kops.k8s.io-needs-update"
)

// ScwCloud exposes all the interfaces required to operate on Scaleway resources
type ScwCloud interface {
	fi.Cloud

	ClusterName(tags []string) string
	DNS() (dnsprovider.Interface, error)
	ProviderID() kops.CloudProviderID
	Region() string
	Zone() string

	DomainService() *domain.API
	GatewayService() *vpcgw.API
	IamService() *iam.API
	InstanceService() *instance.API
	LBService() *lb.ZonedAPI
	VPCService() *vpc.API

	DeleteGroup(group *cloudinstances.CloudInstanceGroup) error
	DeleteInstance(i *cloudinstances.CloudInstance) error
	DeregisterInstance(instance *cloudinstances.CloudInstance) error
	DetachInstance(instance *cloudinstances.CloudInstance) error
	FindClusterStatus(cluster *kops.Cluster) (*kops.ClusterStatus, error)
	FindVPCInfo(id string) (*fi.VPCInfo, error)
	GetApiIngressStatus(cluster *kops.Cluster) ([]fi.ApiIngressStatus, error)
	GetCloudGroups(cluster *kops.Cluster, instancegroups []*kops.InstanceGroup, warnUnmatched bool, nodes []v1.Node) (map[string]*cloudinstances.CloudInstanceGroup, error)

	GetClusterGatewayNetworks(clusterName string) ([]*vpcgw.GatewayNetwork, error)
	GetClusterGateways(clusterName string) ([]*vpcgw.Gateway, error)
	GetClusterLoadBalancers(clusterName string) ([]*lb.LB, error)
	GetClusterServers(clusterName string, instanceGroupName *string) ([]*instance.Server, error)
	GetClusterSSHKeys(clusterName string) ([]*iam.SSHKey, error)
	GetClusterVolumes(clusterName string) ([]*instance.Volume, error)
	GetClusterVPCs(clusterName string) ([]*vpc.PrivateNetwork, error)

	DeleteDNSRecord(record *domain.Record, domainName string) error
	DeleteGateway(gateway *vpcgw.Gateway) error
	DeleteLoadBalancer(loadBalancer *lb.LB) error
	DeleteServer(server *instance.Server) error
	DeleteSSHKey(sshkey *iam.SSHKey) error
	DeleteVolume(volume *instance.Volume) error
	DeleteVPC(vpc *vpc.PrivateNetwork) error
}

// static compile time check to validate ScwCloud's fi.Cloud Interface.
var _ fi.Cloud = &scwCloudImplementation{}

// scwCloudImplementation holds the scw.Client object to interact with Scaleway resources.
type scwCloudImplementation struct {
	client *scw.Client
	region scw.Region
	zone   scw.Zone
	dns    dnsprovider.Interface
	tags   map[string]string

	domainAPI   *domain.API
	gatewayAPI  *vpcgw.API
	iamAPI      *iam.API
	instanceAPI *instance.API
	lbAPI       *lb.ZonedAPI
	vpcAPI      *vpc.API
}

// NewScwCloud returns a Cloud with a Scaleway Client using the env vars SCW_ACCESS_KEY, SCW_SECRET_KEY and SCW_DEFAULT_PROJECT_ID
func NewScwCloud(tags map[string]string) (ScwCloud, error) {
	displayEnv()

	errList := []error(nil)

	region, err := scw.ParseRegion(tags["region"])
	if err != nil {
		errList = append(errList, fmt.Errorf("error parsing Scaleway region: %w", err))
	}
	zone, err := scw.ParseZone(tags["zone"])
	if err != nil {
		errList = append(errList, fmt.Errorf("error parsing Scaleway zone: %w", err))
	}

	// We make sure that the credentials env vars are defined
	scwAccessKey := os.Getenv("SCW_ACCESS_KEY")
	if scwAccessKey == "" {
		errList = append(errList, fmt.Errorf("SCW_ACCESS_KEY has to be set as an environment variable"))
	}
	scwSecretKey := os.Getenv("SCW_SECRET_KEY")
	if scwSecretKey == "" {
		errList = append(errList, fmt.Errorf("SCW_SECRET_KEY has to be set as an environment variable"))
	}
	scwProjectID := os.Getenv("SCW_DEFAULT_PROJECT_ID")
	if scwProjectID == "" {
		errList = append(errList, fmt.Errorf("SCW_DEFAULT_PROJECT_ID has to be set as an environment variable"))
	}

	if len(errList) != 0 {
		return nil, errors.NewAggregate(errList)
	}

	scwClient, err := scw.NewClient(
		scw.WithUserAgent(KopsUserAgentPrefix+kopsv.Version),
		scw.WithEnv(),
	)
	if err != nil {
		return nil, fmt.Errorf("error building client for Scaleway Cloud: %w", err)
	}

	return &scwCloudImplementation{
		client:      scwClient,
		region:      region,
		zone:        zone,
		dns:         dns.NewProvider(scwClient),
		tags:        tags,
		domainAPI:   domain.NewAPI(scwClient),
		gatewayAPI:  vpcgw.NewAPI(scwClient),
		iamAPI:      iam.NewAPI(scwClient),
		instanceAPI: instance.NewAPI(scwClient),
		lbAPI:       lb.NewZonedAPI(scwClient),
		vpcAPI:      vpc.NewAPI(scwClient),
	}, nil
}

func (s *scwCloudImplementation) ClusterName(tags []string) string {
	for _, tag := range tags {
		if strings.HasPrefix(tag, TagClusterName) {
			return strings.TrimPrefix(tag, TagClusterName+"=")
		}
	}
	return ""
}

func (s *scwCloudImplementation) DNS() (dnsprovider.Interface, error) {
	provider, err := dnsprovider.GetDnsProvider(dns.ProviderName, nil)
	if err != nil {
		return nil, fmt.Errorf("error building DNS provider: %w", err)
	}
	return provider, nil
}

func (s *scwCloudImplementation) ProviderID() kops.CloudProviderID {
	return kops.CloudProviderScaleway
}

func (s *scwCloudImplementation) Region() string {
	return string(s.region)
}

func (s *scwCloudImplementation) Zone() string {
	return string(s.zone)
}

func (s *scwCloudImplementation) DomainService() *domain.API {
	return s.domainAPI
}

func (s *scwCloudImplementation) GatewayService() *vpcgw.API {
	return s.gatewayAPI
}

func (s *scwCloudImplementation) IamService() *iam.API {
	return s.iamAPI
}

func (s *scwCloudImplementation) InstanceService() *instance.API {
	return s.instanceAPI
}

func (s *scwCloudImplementation) LBService() *lb.ZonedAPI {
	return s.lbAPI
}

func (s *scwCloudImplementation) VPCService() *vpc.API {
	return s.vpcAPI
}

// DeleteGroup deletes the cloud resources that make up a CloudInstanceGroup, including the instances.
func (s *scwCloudImplementation) DeleteGroup(group *cloudinstances.CloudInstanceGroup) error {
	toDelete := append(group.NeedUpdate, group.Ready...)
	for _, cloudInstance := range toDelete {
		err := s.DeleteInstance(cloudInstance)
		if err != nil {
			return fmt.Errorf("error deleting group %q: %w", group.HumanName, err)
		}
	}
	return nil
}

func (s *scwCloudImplementation) DeleteInstance(i *cloudinstances.CloudInstance) error {
	server, err := s.instanceAPI.GetServer(&instance.GetServerRequest{
		Zone:     s.zone,
		ServerID: i.ID,
	})
	if err != nil {
		if is404Error(err) {
			klog.V(4).Infof("error deleting cloud instance %s of group %s : instance was already deleted", i.ID, i.CloudInstanceGroup.HumanName)
			return nil
		}
		return fmt.Errorf("deleting cloud instance %s of group %s: %w", i.ID, i.CloudInstanceGroup.HumanName, err)
	}

	err = s.DeleteServer(server.Server)
	if err != nil {
		return fmt.Errorf("deleting cloud instance %s of group %s: %w", i.ID, i.CloudInstanceGroup.HumanName, err)
	}

	return nil
}

func (s *scwCloudImplementation) DeregisterInstance(i *cloudinstances.CloudInstance) error {
	server, err := s.instanceAPI.GetServer(&instance.GetServerRequest{
		Zone:     s.zone,
		ServerID: i.ID,
	})
	if err != nil {
		return fmt.Errorf("deregistering cloud instance %s of group %q: %w", i.ID, i.CloudInstanceGroup.HumanName, err)
	}

	// We remove the instance's IP from load-balancers
	lbs, err := s.GetClusterLoadBalancers(s.ClusterName(server.Server.Tags))
	if err != nil {
		return fmt.Errorf("deregistering cloud instance %s of group %q: %w", i.ID, i.CloudInstanceGroup.HumanName, err)
	}
	for _, loadBalancer := range lbs {
		backEnds, err := s.lbAPI.ListBackends(&lb.ZonedAPIListBackendsRequest{
			Zone: s.zone,
			LBID: loadBalancer.ID,
		}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("deregistering cloud instance %s of group %q: listing load-balancer's back-ends for instance creation: %w", i.ID, i.CloudInstanceGroup.HumanName, err)
		}
		for _, backEnd := range backEnds.Backends {
			for _, serverIP := range backEnd.Pool {
				if serverIP == fi.ValueOf(server.Server.PrivateIP) {
					_, err := s.lbAPI.RemoveBackendServers(&lb.ZonedAPIRemoveBackendServersRequest{
						Zone:      s.zone,
						BackendID: backEnd.ID,
						ServerIP:  []string{serverIP},
					})
					if err != nil {
						return fmt.Errorf("deregistering cloud instance %s of group %q: removing IP from lb: %w", i.ID, i.CloudInstanceGroup.HumanName, err)
					}
				}
			}
		}
	}

	return nil
}

func (s *scwCloudImplementation) DetachInstance(i *cloudinstances.CloudInstance) error {
	klog.V(8).Infof("Scaleway DetachInstance is not implemented yet")
	return fmt.Errorf("DetachInstance is not implemented yet for Scaleway")
}

// FindClusterStatus was used before etcd-manager to check the etcd cluster status and prevent unsupported changes.
func (s *scwCloudImplementation) FindClusterStatus(cluster *kops.Cluster) (*kops.ClusterStatus, error) {
	klog.V(8).Info("Scaleway FindClusterStatus is not implemented")
	return nil, nil
}

// FindVPCInfo is not implemented yet, it's only here to satisfy the fi.Cloud interface
func (s *scwCloudImplementation) FindVPCInfo(id string) (*fi.VPCInfo, error) {
	klog.V(8).Info("Scaleway clusters don't have a VPC yet so FindVPCInfo is not implemented")
	return nil, fmt.Errorf("FindVPCInfo is not implemented yet for Scaleway")
}

func (s *scwCloudImplementation) GetApiIngressStatus(cluster *kops.Cluster) ([]fi.ApiIngressStatus, error) {
	var ingresses []fi.ApiIngressStatus
	name := "api." + cluster.Name

	responseLoadBalancers, err := s.lbAPI.ListLBs(&lb.ZonedAPIListLBsRequest{
		Zone: s.zone,
		Name: &name,
	}, scw.WithAllPages())
	if err != nil {
		return nil, fmt.Errorf("finding load-balancers: %w", err)
	}
	if len(responseLoadBalancers.LBs) == 0 {
		klog.V(8).Infof("Could not find any load-balancers for cluster %s", cluster.Name)
		return nil, nil
	}
	if len(responseLoadBalancers.LBs) > 1 {
		klog.V(4).Infof("More than 1 load-balancer with the name %s was found", name)
	}

	for _, loadBalancer := range responseLoadBalancers.LBs {
		for _, lbIP := range loadBalancer.IP {
			ingresses = append(ingresses, fi.ApiIngressStatus{IP: lbIP.IPAddress})
		}
	}

	return ingresses, nil
}

func (s *scwCloudImplementation) GetCloudGroups(cluster *kops.Cluster, instancegroups []*kops.InstanceGroup, warnUnmatched bool, nodes []v1.Node) (map[string]*cloudinstances.CloudInstanceGroup, error) {
	groups := make(map[string]*cloudinstances.CloudInstanceGroup)

	nodeMap := cloudinstances.GetNodeMap(nodes, cluster)

	serverGroups, err := findServerGroups(s, cluster.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to find server groups: %w", err)
	}

	for _, ig := range instancegroups {
		serverGroup, ok := serverGroups[ig.Name]
		if !ok {
			if warnUnmatched {
				klog.Warningf("Server group %q has no corresponding instance group", ig.Name)
			}
			continue
		}

		groups[ig.Name], err = buildCloudGroup(ig, serverGroup, nodeMap)
		if err != nil {
			return nil, fmt.Errorf("failed to build cloud group for instance group %q: %w", ig.Name, err)
		}
	}

	return groups, nil
}

func findServerGroups(s *scwCloudImplementation, clusterName string) (map[string][]*instance.Server, error) {
	servers, err := s.GetClusterServers(clusterName, nil)
	if err != nil {
		return nil, err
	}

	serverGroups := make(map[string][]*instance.Server)
	for _, server := range servers {
		igName := ""
		for _, tag := range server.Tags {
			if strings.HasPrefix(tag, TagInstanceGroup) {
				igName = strings.TrimPrefix(tag, TagInstanceGroup+"=")
				break
			}
		}
		serverGroups[igName] = append(serverGroups[igName], server)
	}

	return serverGroups, nil
}

func buildCloudGroup(ig *kops.InstanceGroup, sg []*instance.Server, nodeMap map[string]*v1.Node) (*cloudinstances.CloudInstanceGroup, error) {
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
		for _, tag := range server.Tags {
			if tag == TagNeedsUpdate {
				status = cloudinstances.CloudInstanceStatusNeedsUpdate
			}
		}

		cloudInstance, err := cloudInstanceGroup.NewCloudInstance(server.ID, status, nodeMap[server.ID])
		if err != nil {
			return nil, fmt.Errorf("failed to create cloud instance for server %s(%s): %w", server.Name, server.ID, err)
		}
		cloudInstance.State = cloudinstances.State(server.State)
		cloudInstance.MachineType = server.CommercialType
		for _, tag := range server.Tags {
			if strings.HasPrefix(tag, TagNameRolePrefix) {
				cloudInstance.Roles = append(cloudInstance.Roles, strings.TrimPrefix(tag, TagNameRolePrefix+"="))
			}
		}
		if server.PrivateIP != nil {
			cloudInstance.PrivateIP = *server.PrivateIP
		}
	}

	return cloudInstanceGroup, nil
}

func (s *scwCloudImplementation) GetClusterGatewayNetworks(privateNetworkID string) ([]*vpcgw.GatewayNetwork, error) {
	gwNetworks, err := s.gatewayAPI.ListGatewayNetworks(&vpcgw.ListGatewayNetworksRequest{
		Zone:             s.zone,
		PrivateNetworkID: scw.StringPtr(privateNetworkID),
	}, scw.WithAllPages())
	if err != nil {
		return nil, fmt.Errorf("failed to list gateway networks: %w", err)
	}
	return gwNetworks.GatewayNetworks, nil
}

func (s *scwCloudImplementation) GetClusterGateways(clusterName string) ([]*vpcgw.Gateway, error) {
	gws, err := s.gatewayAPI.ListGateways(&vpcgw.ListGatewaysRequest{
		Zone: s.zone,
		Tags: []string{TagClusterName + "=" + clusterName},
	}, scw.WithAllPages())
	if err != nil {
		return nil, fmt.Errorf("failed to list gateway networks: %w", err)
	}
	return gws.Gateways, nil
}

func (s *scwCloudImplementation) GetClusterLoadBalancers(clusterName string) ([]*lb.LB, error) {
	loadBalancerName := "api." + clusterName
	lbs, err := s.lbAPI.ListLBs(&lb.ZonedAPIListLBsRequest{
		Zone: s.zone,
		Name: &loadBalancerName,
	}, scw.WithAllPages())
	if err != nil {
		return nil, fmt.Errorf("listing cluster load-balancers: %w", err)
	}
	return lbs.LBs, nil
}

func (s *scwCloudImplementation) GetClusterServers(clusterName string, instanceGroupName *string) ([]*instance.Server, error) {
	tags := []string{TagClusterName + "=" + clusterName}
	if instanceGroupName != nil {
		tags = append(tags, fmt.Sprintf("%s=%s", TagInstanceGroup, *instanceGroupName))
	}
	request := &instance.ListServersRequest{
		Zone: s.zone,
		Tags: tags,
	}
	servers, err := s.instanceAPI.ListServers(request, scw.WithAllPages())
	if err != nil {
		if instanceGroupName != nil {
			return nil, fmt.Errorf("failed to list cluster servers named %q: %w", *instanceGroupName, err)
		}
		return nil, fmt.Errorf("failed to list cluster servers: %w", err)
	}
	return servers.Servers, nil
}

func (s *scwCloudImplementation) GetClusterSSHKeys(clusterName string) ([]*iam.SSHKey, error) {
	clusterSSHKeys := []*iam.SSHKey(nil)
	allSSHKeys, err := s.iamAPI.ListSSHKeys(&iam.ListSSHKeysRequest{}, scw.WithAllPages())
	for _, sshkey := range allSSHKeys.SSHKeys {
		if strings.HasPrefix(sshkey.Name, fmt.Sprintf("kubernetes.%s-", clusterName)) {
			clusterSSHKeys = append(clusterSSHKeys, sshkey)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to list cluster ssh keys: %w", err)
	}
	return clusterSSHKeys, nil
}

func (s *scwCloudImplementation) GetClusterVolumes(clusterName string) ([]*instance.Volume, error) {
	volumes, err := s.instanceAPI.ListVolumes(&instance.ListVolumesRequest{
		Zone: s.zone,
		Tags: []string{TagClusterName + "=" + clusterName},
	}, scw.WithAllPages())
	if err != nil {
		return nil, fmt.Errorf("failed to list cluster volumes: %w", err)
	}
	return volumes.Volumes, nil
}

func (s *scwCloudImplementation) GetClusterVPCs(clusterName string) ([]*vpc.PrivateNetwork, error) {
	vpcs, err := s.vpcAPI.ListPrivateNetworks(&vpc.ListPrivateNetworksRequest{
		Zone: s.zone,
		Tags: []string{TagClusterName + "=" + clusterName},
	}, scw.WithAllPages())
	if err != nil {
		return nil, fmt.Errorf("failed to list cluster VPCs: %w", err)
	}
	return vpcs.PrivateNetworks, nil
}

func (s *scwCloudImplementation) DeleteGateway(gateway *vpcgw.Gateway) error {
	// We look for gateway connexions to private networks and detach them before deleting the gateway
	connexions, err := s.GetClusterGatewayNetworks(gateway.ID)
	if err != nil {
		if is404Error(err) {
			klog.V(8).Infof("Gateway %q (%s) was already deleted", gateway.Name, gateway.ID)
			return nil
		}
		return fmt.Errorf("error listing gateway networks: %w", err)
	}
	for _, connexion := range connexions {
		err := s.gatewayAPI.DeleteGatewayNetwork(&vpcgw.DeleteGatewayNetworkRequest{
			Zone:             s.zone,
			GatewayNetworkID: connexion.ID,
			CleanupDHCP:      true,
		})
		if err != nil {
			return fmt.Errorf("failed to detach gateway %s from private network: %w", gateway.ID, err)
		}
	}

	// We detach the IP of the gateway
	_, err = s.gatewayAPI.WaitForGateway(&vpcgw.WaitForGatewayRequest{
		GatewayID: gateway.ID,
		Zone:      s.zone,
	})
	if err != nil {
		if is404Error(err) {
			klog.V(8).Infof("Gateway %q (%s) was already deleted", gateway.Name, gateway.ID)
			return nil
		}
		return fmt.Errorf("error waiting for gateway: %w", err)
	}

	_, err = s.gatewayAPI.UpdateIP(&vpcgw.UpdateIPRequest{
		Zone:      s.zone,
		IPID:      gateway.IP.ID,
		GatewayID: scw.StringPtr(""),
	})
	if err != nil {
		return fmt.Errorf("failed to detach gateway IP: %w", err)
	}

	// We delete the IP of the gateway
	_, err = s.gatewayAPI.WaitForGateway(&vpcgw.WaitForGatewayRequest{
		GatewayID: gateway.ID,
		Zone:      s.zone,
	})
	if err != nil {
		return fmt.Errorf("error waiting for gateway: %w", err)
	}

	err = s.gatewayAPI.DeleteIP(&vpcgw.DeleteIPRequest{
		Zone: s.zone,
		IPID: gateway.IP.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete gateway IP: %w", err)
	}

	// We delete the gateway once it's in a stable state
	_, err = s.gatewayAPI.WaitForGateway(&vpcgw.WaitForGatewayRequest{
		GatewayID: gateway.ID,
		Zone:      s.zone,
	})
	if err != nil {
		return fmt.Errorf("error waiting for gateway: %w", err)
	}
	err = s.gatewayAPI.DeleteGateway(&vpcgw.DeleteGatewayRequest{
		Zone:        s.zone,
		GatewayID:   gateway.ID,
		CleanupDHCP: true,
	})
	if err != nil {
		return fmt.Errorf("failed to delete gateway %s: %w", gateway.ID, err)
	}

	return nil
}

func (s *scwCloudImplementation) DeleteLoadBalancer(loadBalancer *lb.LB) error {
	ipsToRelease := loadBalancer.IP

	// We delete the load-balancer once it's in a stable state
	_, err := s.lbAPI.WaitForLb(&lb.ZonedAPIWaitForLBRequest{
		LBID: loadBalancer.ID,
		Zone: s.zone,
	})
	if err != nil {
		if is404Error(err) {
			klog.V(8).Infof("Load-balancer %q (%s) was already deleted", loadBalancer.Name, loadBalancer.ID)
			return nil
		}
		return fmt.Errorf("waiting for load-balancer: %w", err)
	}
	err = s.lbAPI.DeleteLB(&lb.ZonedAPIDeleteLBRequest{
		Zone: s.zone,
		LBID: loadBalancer.ID,
	})
	if err != nil {
		return fmt.Errorf("deleting load-balancer %s: %w", loadBalancer.ID, err)
	}

	// We wait for the load-balancer to be deleted, then we detach its IPs
	_, err = s.lbAPI.WaitForLb(&lb.ZonedAPIWaitForLBRequest{
		LBID: loadBalancer.ID,
		Zone: s.zone,
	})
	if !is404Error(err) {
		return fmt.Errorf("waiting for load-balancer %s after deletion: %w", loadBalancer.ID, err)
	}
	for _, ip := range ipsToRelease {
		err := s.lbAPI.ReleaseIP(&lb.ZonedAPIReleaseIPRequest{
			Zone: s.zone,
			IPID: ip.ID,
		})
		if err != nil {
			return fmt.Errorf("deleting load-balancer IP: %w", err)
		}
	}
	return nil
}

func (s *scwCloudImplementation) DeleteDNSRecord(record *domain.Record, domainName string) error {
	recordDeleteRequest := &domain.UpdateDNSZoneRecordsRequest{
		DNSZone: domainName,
		Changes: []*domain.RecordChange{
			{
				Delete: &domain.RecordChangeDelete{
					ID: scw.StringPtr(record.ID),
				},
			},
		},
	}
	_, err := s.domainAPI.UpdateDNSZoneRecords(recordDeleteRequest)
	if err != nil {
		if is404Error(err) {
			klog.V(8).Infof("DNS record %q (%s) was already deleted", record.Name, record.ID)
			return nil
		}
		return fmt.Errorf("failed to delete record %s: %w", record.Name, err)
	}
	return nil
}

func (s *scwCloudImplementation) DeleteServer(server *instance.Server) error {
	srv, err := s.instanceAPI.GetServer(&instance.GetServerRequest{
		Zone:     s.zone,
		ServerID: server.ID,
	})
	if err != nil {
		if is404Error(err) {
			klog.V(8).Infof("Instance server %q (%s) was already deleted", server.Name, server.ID)
			return nil
		}
		return err
	}

	// We detach the private network
	if len(srv.Server.PrivateNics) > 0 {
		err = s.instanceAPI.DeletePrivateNIC(&instance.DeletePrivateNICRequest{
			Zone:         s.zone,
			ServerID:     server.ID,
			PrivateNicID: srv.Server.PrivateNics[0].ID,
		})
		if err != nil {
			return fmt.Errorf("delete instance %s: error detaching private network: %w", server.ID, err)
		}
		return err
	}

	// If the server is running, we turn it off and wait before deleting it
	if srv.Server.State == instance.ServerStateRunning {
		_, err := s.instanceAPI.ServerAction(&instance.ServerActionRequest{
			Zone:     s.zone,
			ServerID: server.ID,
			Action:   instance.ServerActionPoweroff,
		})
		if err != nil {
			return fmt.Errorf("delete server %s: error powering off instance: %w", server.ID, err)
		}
	}
	_, err = s.instanceAPI.WaitForServer(&instance.WaitForServerRequest{
		ServerID: server.ID,
		Zone:     s.zone,
	})
	if err != nil {
		return fmt.Errorf("delete server %s: error waiting for instance after power-off: %w", server.ID, err)
	}

	// We delete the server and wait before deleting its volumes
	err = s.instanceAPI.DeleteServer(&instance.DeleteServerRequest{
		ServerID: server.ID,
		Zone:     s.zone,
	})
	if err != nil {
		return fmt.Errorf("delete server %s: error deleting instance: %w", server.ID, err)
	}
	_, err = s.instanceAPI.WaitForServer(&instance.WaitForServerRequest{
		ServerID: server.ID,
		Zone:     s.zone,
	})
	if !is404Error(err) {
		return fmt.Errorf("delete server %s: error waiting for instance after deletion: %w", server.ID, err)
	}

	// We delete the volumes that were attached to the server (including etcd volumes)
	for i := range server.Volumes {
		err = s.instanceAPI.DeleteVolume(&instance.DeleteVolumeRequest{
			Zone:     s.zone,
			VolumeID: server.Volumes[i].ID,
		})
		if err != nil {
			return fmt.Errorf("delete server %s: error deleting volume %s: %w", server.ID, server.Volumes[i].Name, err)
		}
	}

	return nil
}

func (s *scwCloudImplementation) DeleteSSHKey(sshkey *iam.SSHKey) error {
	err := s.iamAPI.DeleteSSHKey(&iam.DeleteSSHKeyRequest{
		SSHKeyID: sshkey.ID,
	})
	if err != nil {
		if is404Error(err) {
			klog.V(8).Infof("SSH key %q (%s) was already deleted", sshkey.Name, sshkey.ID)
			return nil
		}
		return fmt.Errorf("failed to delete ssh key %s: %w", sshkey.ID, err)
	}
	return nil
}

func (s *scwCloudImplementation) DeleteVolume(volume *instance.Volume) error {
	err := s.instanceAPI.DeleteVolume(&instance.DeleteVolumeRequest{
		VolumeID: volume.ID,
		Zone:     s.zone,
	})
	if err != nil {
		if is404Error(err) {
			klog.V(8).Infof("Volume %q (%s) was already deleted", volume.Name, volume.ID)
			return nil
		}
		return fmt.Errorf("failed to delete volume %s: %w", volume.ID, err)
	}

	_, err = s.instanceAPI.WaitForVolume(&instance.WaitForVolumeRequest{
		VolumeID: volume.ID,
		Zone:     s.zone,
	})
	if !is404Error(err) {
		return fmt.Errorf("delete volume %s: error waiting for volume after deletion: %w", volume.ID, err)
	}

	return nil
}

func (s *scwCloudImplementation) DeleteVPC(privateNetwork *vpc.PrivateNetwork) error {
	err := s.vpcAPI.DeletePrivateNetwork(&vpc.DeletePrivateNetworkRequest{
		PrivateNetworkID: privateNetwork.ID,
		Zone:             s.zone,
	})
	if err != nil {
		if is404Error(err) {
			klog.V(8).Infof("Private network %q (%s) was already deleted", privateNetwork.Name, privateNetwork.ID)
			return nil
		}
		return fmt.Errorf("failed to delete VPC %s: %w", privateNetwork.ID, err)
	}
	return nil
}
