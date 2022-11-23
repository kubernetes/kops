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
	"strings"

	"github.com/spf13/viper"

	account "github.com/scaleway/scaleway-sdk-go/api/account/v2alpha1"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/errors"
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

	ClusterName(tags []string) string
	DNS() (dnsprovider.Interface, error)
	ProviderID() kops.CloudProviderID
	Region() string
	Zone() string

	AccountService() *account.API
	InstanceService() *instance.API

	DeleteGroup(group *cloudinstances.CloudInstanceGroup) error
	DeleteInstance(i *cloudinstances.CloudInstance) error
	DeregisterInstance(instance *cloudinstances.CloudInstance) error
	DetachInstance(instance *cloudinstances.CloudInstance) error
	FindClusterStatus(cluster *kops.Cluster) (*kops.ClusterStatus, error)
	FindVPCInfo(id string) (*fi.VPCInfo, error)
	GetApiIngressStatus(cluster *kops.Cluster) ([]fi.ApiIngressStatus, error)
	GetCloudGroups(cluster *kops.Cluster, instancegroups []*kops.InstanceGroup, warnUnmatched bool, nodes []v1.Node) (map[string]*cloudinstances.CloudInstanceGroup, error)

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
	errList := []error(nil)

	region, err := scw.ParseRegion(viper.GetString("SCW_DEFAULT_REGION"))
	if err != nil {
		errList = append(errList, fmt.Errorf("error parsing SCW_DEFAULT_REGION: %w", err))
	}
	zone, err := scw.ParseZone(viper.GetString("SCW_DEFAULT_ZONE"))
	if err != nil {
		errList = append(errList, fmt.Errorf("error parsing SCW_DEFAULT_ZONE: %w", err))
	}

	// We make sure that the credentials env vars are defined
	scwAccessKey := viper.GetString("SCW_ACCESS_KEY")
	if scwAccessKey == "" {
		errList = append(errList, fmt.Errorf("SCW_ACCESS_KEY has to be set as an environment variable or in kops configuration file"))
	}
	scwSecretKey := viper.GetString("SCW_SECRET_KEY")
	if scwSecretKey == "" {
		errList = append(errList, fmt.Errorf("SCW_SECRET_KEY has to be set as an environment variable or in kops configuration file"))
	}
	scwProjectID := viper.GetString("SCW_DEFAULT_PROJECT_ID")
	if scwProjectID == "" {
		errList = append(errList, fmt.Errorf("SCW_DEFAULT_PROJECT_ID has to be set as an environment variable or in kops configuration file"))
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
		tags:        tags,
		accountAPI:  account.NewAPI(scwClient),
		instanceAPI: instance.NewAPI(scwClient),
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
	klog.V(8).Infof("Scaleway DNS is not implemented yet")
	return nil, fmt.Errorf("DNS is not implemented yet for Scaleway")
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

func (s *scwCloudImplementation) AccountService() *account.API {
	return s.accountAPI
}

func (s *scwCloudImplementation) InstanceService() *instance.API {
	return s.instanceAPI
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
	klog.V(8).Infof("Scaleway DeregisterInstance is not implemented yet")
	return fmt.Errorf("DeregisterInstance is not implemented yet for Scaleway")
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
	klog.V(8).Info("Scaleway doesn't have a VPC yet so FindVPCInfo is not implemented")
	return nil, fmt.Errorf("FindVPCInfo is not implemented yet for Scaleway")
}

func (s *scwCloudImplementation) GetApiIngressStatus(cluster *kops.Cluster) ([]fi.ApiIngressStatus, error) {
	klog.V(8).Info("Scaleway doesn't have load-balancers yet so GetApiIngressStatus is not implemented")
	return nil, fmt.Errorf("GetApiIngressStatus is not implemented yet for Scaleway")
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

func (s *scwCloudImplementation) DeleteVolume(volume *instance.Volume) error {
	err := s.instanceAPI.DeleteVolume(&instance.DeleteVolumeRequest{
		VolumeID: volume.ID,
		Zone:     s.zone,
	})
	if err != nil {
		return fmt.Errorf("failed to delete volume %s: %w", volume.ID, err)
	}

	_, err = s.instanceAPI.WaitForVolume(&instance.WaitForVolumeRequest{
		VolumeID: volume.ID,
		Zone:     s.zone,
	})
	if !is404Error(err) {
		return fmt.Errorf("delete server %s: error waiting for volume after deletion: %w", volume.ID, err)
	}

	return nil
}
