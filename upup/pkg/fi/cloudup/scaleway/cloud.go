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

	account "github.com/scaleway/scaleway-sdk-go/api/account/v2alpha1"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	kopsv "k8s.io/kops"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/upup/pkg/fi"
)

const (
	TagClusterName      = "kops.k8s.io/cluster"
	KopsUserAgentPrefix = "kubernetes-kops/"
	TagInstanceGroup    = "instance-group"
	TagNameRolePrefix   = "k8s.io/role/"
)

// ScwCloud exposes all the interfaces required to operate on Scaleway resources
type ScwCloud interface {
	fi.Cloud

	Region() string
	Zone() string
	ProviderID() kops.CloudProviderID
	DNS() (dnsprovider.Interface, error)
	ClusterName(tags []string) string

	AccountService() *account.API
	InstanceService() *instance.API

	GetApiIngressStatus(cluster *kops.Cluster) ([]fi.ApiIngressStatus, error)
	FindClusterStatus(cluster *kops.Cluster) (*kops.ClusterStatus, error)
	GetCloudGroups(cluster *kops.Cluster, instancegroups []*kops.InstanceGroup, warnUnmatched bool, nodes []v1.Node) (map[string]*cloudinstances.CloudInstanceGroup, error)
	DeleteGroup(group *cloudinstances.CloudInstanceGroup) error
	FindVPCInfo(id string) (*fi.VPCInfo, error)
	DetachInstance(instance *cloudinstances.CloudInstance) error
	DeregisterInstance(instance *cloudinstances.CloudInstance) error
	DeleteInstance(i *cloudinstances.CloudInstance) error

	GetClusterServers(clusterName string, serverName *string) ([]*instance.Server, error)
	GetClusterVolumes(clusterName string) ([]*instance.Volume, error)

	DeleteServer(server *instance.Server) error
	DeleteVolume(volume *instance.Volume) error
}

// static compile time check to validate ScwCloud's fi.Cloud Interface.
var _ fi.Cloud = &scwCloudImplementation{}

// scwCloudImplementation holds the scw.Client object to interact with Scaleway resources.
type scwCloudImplementation struct {
	client *scw.Client
	region scw.Region
	zone   scw.Zone
	tags   map[string]string

	accountAPI  *account.API
	instanceAPI *instance.API
}

// NewScwCloud returns a Cloud with a Scaleway Client using the env vars SCW_ACCESS_KEY, SCW_SECRET_KEY,
// SCW_DEFAULT_PROJECT_ID, SCW_DEFAULT_REGION and SCW_DEFAULT_ZONE
func NewScwCloud(tags map[string]string) (ScwCloud, error) {
	region, err := scw.ParseRegion(os.Getenv("SCW_DEFAULT_REGION"))
	if err != nil {
		return nil, fmt.Errorf("error parsing SCW_DEFAULT_REGION: %w", err)
	}
	zone, err := scw.ParseZone(os.Getenv("SCW_DEFAULT_ZONE"))
	if err != nil {
		return nil, fmt.Errorf("error parsing SCW_DEFAULT_ZONE: %w", err)
	}

	// We make sure that the credentials env vars are defined
	scwAccessKey := os.Getenv("SCW_ACCESS_KEY")
	if scwAccessKey == "" {
		return nil, fmt.Errorf("SCW_ACCESS_KEY has to be set as an environment variable")
	}
	scwSecretKey := os.Getenv("SCW_SECRET_KEY")
	if scwSecretKey == "" {
		return nil, fmt.Errorf("SCW_SECRET_KEY has to be set as an environment variable")
	}
	scwProjectID := os.Getenv("SCW_DEFAULT_PROJECT_ID")
	if scwProjectID == "" {
		return nil, fmt.Errorf("SCW_DEFAULT_PROJECT_ID has to be set as an environment variable")
	}

	scwClient, err := scw.NewClient(
		scw.WithUserAgent("kubernetes-kops/"+kopsv.Version),
		scw.WithEnv(),
	)
	if err != nil {
		return nil, fmt.Errorf("error building client for Scaleway Cloud: %w", err)
	}

	return &scwCloudImplementation{
		client:      scwClient,
		region:      region,
		zone:        zone,
		tags:        tags,
		accountAPI:  account.NewAPI(scwClient),
		instanceAPI: instance.NewAPI(scwClient),
	}, nil
}

func (s *scwCloudImplementation) Region() string {
	return string(s.region)
}

func (s *scwCloudImplementation) Zone() string {
	return string(s.zone)
}

func (s *scwCloudImplementation) ClusterName(tags []string) string {
	for _, tag := range tags {
		if strings.HasPrefix(tag, TagClusterName) {
			return strings.TrimPrefix(tag, TagClusterName+"=")
		}
	}
	return ""
}

func (s *scwCloudImplementation) ProviderID() kops.CloudProviderID {
	return kops.CloudProviderScaleway
}

func (s *scwCloudImplementation) DNS() (dnsprovider.Interface, error) {
	//TODO(Mia-Cross) implement me
	panic("Scaleway doesn't have a DNS yet")
}

func (s *scwCloudImplementation) AccountService() *account.API {
	return s.accountAPI
}

func (s *scwCloudImplementation) InstanceService() *instance.API {
	return s.instanceAPI
}

// FindVPCInfo is not implemented yet, it's only here to satisfy the fi.Cloud interface
func (s *scwCloudImplementation) FindVPCInfo(id string) (*fi.VPCInfo, error) {
	klog.V(8).Info("FindVPCInfo is not implemented yet for Scaleway")
	return nil, fmt.Errorf("scaleway cloud provider does not support VPC at this time")
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
		return fmt.Errorf("error deleting cloud instance %s of group %s: %w", i.ID, i.CloudInstanceGroup.HumanName, err)
	}

	err = s.DeleteServer(server.Server)
	if err != nil {
		return fmt.Errorf("error deleting cloud instance %s of group %s: %w", i.ID, i.CloudInstanceGroup.HumanName, err)
	}

	return nil
}

func (s *scwCloudImplementation) DeregisterInstance(i *cloudinstances.CloudInstance) error {
	//TODO(Mia-Cross) implement me
	panic("implement me")
}

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

func (s *scwCloudImplementation) DetachInstance(i *cloudinstances.CloudInstance) error {
	//TODO(Mia-Cross) implement me
	panic("implement me")
}

func (s *scwCloudImplementation) GetCloudGroups(cluster *kops.Cluster, instancegroups []*kops.InstanceGroup, warnUnmatched bool, nodes []v1.Node) (map[string]*cloudinstances.CloudInstanceGroup, error) {
	groups := make(map[string]*cloudinstances.CloudInstanceGroup)

	nodeMap := cloudinstances.GetNodeMap(nodes, cluster)

	serverGroups, err := findServerGroups(s, cluster.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to find server groups: %w", err)
	}

	for igName, serverGroup := range serverGroups {
		var instanceGroup *kops.InstanceGroup
		for _, ig := range instancegroups {
			if igName == ig.Name {
				instanceGroup = ig
				break
			}
		}
		if instanceGroup == nil {
			if warnUnmatched {
				klog.Warningf("Server group %q has no corresponding instance group", igName)
			}
			continue
		}

		groups[instanceGroup.Name], err = buildCloudGroup(instanceGroup, serverGroup, nodeMap)
		if err != nil {
			return nil, fmt.Errorf("failed to build cloud group for instance group %q: %w", instanceGroup.Name, err)
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
		MinSize:       int(fi.Int32Value(ig.Spec.MinSize)),
		TargetSize:    int(fi.Int32Value(ig.Spec.MinSize)),
		MaxSize:       int(fi.Int32Value(ig.Spec.MaxSize)),
	}

	for _, server := range sg {
		status := cloudinstances.CloudInstanceStatusUpToDate
		cloudInstance, err := cloudInstanceGroup.NewCloudInstance(server.ID, status, nodeMap[server.ID])
		if err != nil {
			return nil, fmt.Errorf("failed to create cloud instance for server %s(%s): %w", server.Name, server.ID, err)
		}
		cloudInstance.State = cloudinstances.State(server.State)
		cloudInstance.MachineType = server.CommercialType
		for _, tag := range server.Tags {
			if strings.HasPrefix(tag, TagNameRolePrefix) {
				cloudInstance.Roles = append(cloudInstance.Roles, strings.TrimPrefix(tag, TagNameRolePrefix))
			}
		}
		if server.PrivateIP != nil {
			cloudInstance.PrivateIP = *server.PrivateIP
		}
	}

	return cloudInstanceGroup, nil
}

func (s *scwCloudImplementation) FindClusterStatus(cluster *kops.Cluster) (*kops.ClusterStatus, error) {
	klog.V(8).Info("FindClusterStatus is not implemented yet for Scaleway")
	return nil, nil
}

func (s *scwCloudImplementation) GetApiIngressStatus(cluster *kops.Cluster) ([]fi.ApiIngressStatus, error) {
	//TODO(Mia-Cross) implement me
	panic("implement me")
}

func (s *scwCloudImplementation) GetClusterServers(clusterName string, serverName *string) ([]*instance.Server, error) {
	request := &instance.ListServersRequest{
		Zone: s.zone,
		Name: serverName,
		Tags: []string{TagClusterName + "=" + clusterName},
	}
	servers, err := s.instanceAPI.ListServers(request, scw.WithAllPages())
	if err != nil {
		if serverName != nil {
			return nil, fmt.Errorf("failed to list cluster servers named %q: %w", *serverName, err)
		}
		return nil, fmt.Errorf("failed to list cluster servers: %w", err)
	}
	return servers.Servers, nil
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

func (s *scwCloudImplementation) DeleteServer(server *instance.Server) error {
	srv, err := s.instanceAPI.GetServer(&instance.GetServerRequest{
		Zone:     s.zone,
		ServerID: server.ID,
	})
	if err != nil {
		if is404Error(err) {
			klog.V(4).Infof("delete server %s: instance was already deleted", server.ID)
			return nil
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
	for {
		_, err := s.instanceAPI.GetServer(&instance.GetServerRequest{
			Zone:     s.zone,
			ServerID: server.ID,
		})
		if is404Error(err) {
			break
		}
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

func (s *scwCloudImplementation) DeleteVolume(volume *instance.Volume) error {
	err := s.instanceAPI.DeleteVolume(&instance.DeleteVolumeRequest{
		VolumeID: volume.ID,
		Zone:     s.zone,
	})
	if err != nil {
		return fmt.Errorf("failed to delete volume %s: %w", volume.ID, err)
	}
	return nil
}
