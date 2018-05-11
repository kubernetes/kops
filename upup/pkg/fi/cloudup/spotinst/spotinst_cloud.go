/*
Copyright 2017 The Kubernetes Authors.

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

package spotinst

import (
	"context"
	"fmt"
	"strings"

	"github.com/blang/semver"
	"github.com/golang/glog"
	"github.com/spotinst/spotinst-sdk-go/service/elastigroup"
	"github.com/spotinst/spotinst-sdk-go/service/elastigroup/providers/aws"
	"github.com/spotinst/spotinst-sdk-go/spotinst"
	"github.com/spotinst/spotinst-sdk-go/spotinst/session"
	"k8s.io/api/core/v1"
	kopsv "k8s.io/kops"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
)

// Cloud represents a Spotinst cloud instance.
type Cloud interface {
	fi.Cloud

	Cloud() fi.Cloud
	Service() elastigroup.Service

	MachineType(cluster *kops.Cluster, ig *kops.InstanceGroup) (string, error)
	Image(cluster *kops.Cluster, channel *kops.Channel) string
	Region() string
	Tags() []string

	ListResources(clusterName string) (map[string]*Resource, error)
	DeleteResource(resource interface{}) error

	FindClusterStatus(cluster *kops.Cluster) (*kops.ClusterStatus, error)
	GetApiIngressStatus(cluster *kops.Cluster) ([]kops.ApiIngressStatus, error)
}

type cloud struct {
	cloud   fi.Cloud
	service elastigroup.Service
}

var _ Cloud = &cloud{}

// NewCloud returns Cloud instance for given ClusterSpec.
func NewCloud(cluster *kops.Cluster) (*cloud, error) {
	glog.V(2).Info("Creating Spotinst cloud")

	cloudProviderID := GuessCloudFromClusterSpec(&cluster.Spec)
	if cloudProviderID == "" {
		return nil, fmt.Errorf("spotinst: unable to infer cloud provider from zones")
	}

	var cloudProvider fi.Cloud
	var err error

	switch cloudProviderID {
	case kops.CloudProviderAWS, kops.CloudProviderGCE:
		glog.V(2).Infof("Cloud provider detected: %s", cloudProviderID)
		cloudProvider, err = buildCloud(cloudProviderID, cluster)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("spotinst: unknown cloud provider: %s", cloudProviderID)
	}

	config := spotinst.DefaultConfig()
	config.WithCredentials(NewCredentials())
	config.WithUserAgent("Kubernetes-Kops/" + kopsv.Version)
	config.WithLogger(newStdLogger())

	return &cloud{
		cloud:   cloudProvider,
		service: elastigroup.New(session.New(config)),
	}, nil
}

func (c *cloud) ProviderID() kops.CloudProviderID {
	return kops.CloudProviderSpotinst
}

func (c *cloud) DNS() (dnsprovider.Interface, error) {
	return c.cloud.DNS()
}

func (c *cloud) FindVPCInfo(id string) (*fi.VPCInfo, error) {
	return c.cloud.FindVPCInfo(id)
}

func (c *cloud) DeleteInstance(instance *cloudinstances.CloudInstanceGroupMember) error {
	instanceID := instance.ID
	if instanceID == "" {
		return fmt.Errorf("spotinst: unexpected instance id: %v", instanceID)
	}

	var nodeName string
	if instance.Node != nil {
		nodeName = instance.Node.Name
	}

	var groupID string
	if instance.CloudInstanceGroup != nil {
		groupID = fi.StringValue(instance.CloudInstanceGroup.Raw.(*aws.Group).ID)
	}

	glog.V(2).Infof("Stopping instance %q, node %q, in group %q", instanceID, nodeName, groupID)
	input := &aws.DetachGroupInput{
		GroupID:                       fi.String(groupID),
		InstanceIDs:                   []string{instanceID},
		ShouldDecrementTargetCapacity: fi.Bool(false),
		ShouldTerminateInstances:      fi.Bool(true),
	}
	if _, err := c.service.CloudProviderAWS().Detach(context.Background(), input); err != nil {
		if nodeName != "" {
			return fmt.Errorf("spotinst: failed to delete instance %q, node %q: %v", instanceID, nodeName, err)
		}
		return fmt.Errorf("spotinst: failed to delete instance %q: %v", instanceID, err)
	}

	return nil
}

func (c *cloud) DeleteGroup(group *cloudinstances.CloudInstanceGroup) error {
	groupID := fi.StringValue(group.Raw.(*aws.Group).ID)

	glog.V(2).Infof("Deleting group %q", groupID)
	input := &aws.DeleteGroupInput{
		GroupID: fi.String(groupID),
	}
	_, err := c.service.CloudProviderAWS().Delete(context.Background(), input)
	if err != nil {
		return fmt.Errorf("spotinst: failed to delete group %q: %v", groupID, err)
	}

	return nil
}

func (c *cloud) GetCloudGroups(cluster *kops.Cluster, instancegroups []*kops.InstanceGroup, warnUnmatched bool, nodes []v1.Node) (map[string]*cloudinstances.CloudInstanceGroup, error) {
	groups := make(map[string]*cloudinstances.CloudInstanceGroup)
	nodeMap := cloudinstances.GetNodeMap(nodes)

	resources, err := c.ListResources(cluster.Name)
	if err != nil {
		return nil, fmt.Errorf("unable to find groups: %v", err)
	}

	for _, resource := range resources {
		group, ok := resource.Raw.(*aws.Group)
		if !ok {
			continue
		}
		var instancegroup *kops.InstanceGroup
		for _, ig := range instancegroups {
			name := getGroupNameByRole(cluster, ig)
			if name == "" {
				continue
			}
			if name == resource.Name {
				if instancegroup != nil {
					return nil, fmt.Errorf("spotinst: found multiple instance groups matching group %q", name)
				}
				instancegroup = ig
			}
		}
		if instancegroup == nil {
			if warnUnmatched {
				glog.Warningf("Found group with no corresponding instance group %q", resource.Name)
			}
			continue
		}
		input := &aws.StatusGroupInput{
			GroupID: group.ID,
		}
		output, err := c.service.CloudProviderAWS().Status(context.Background(), input)
		if err != nil {
			return nil, err
		}
		ig, err := buildInstanceGroup(instancegroup, group, output.Instances, nodeMap)
		if err != nil {
			return nil, fmt.Errorf("spotinst: failed to build instance group: %v", err)
		}
		groups[instancegroup.ObjectMeta.Name] = ig
	}

	return groups, nil
}

func (c *cloud) Cloud() fi.Cloud {
	return c.cloud
}

func (c *cloud) Service() elastigroup.Service {
	return c.service
}

// Default machine types for various types of instance group machine.
const (
	defaultMasterMachineTypeGCE = "n1-standard-1"
	defaultMasterMachineTypeAWS = "m3.medium"

	defaultNodeMachineTypeGCE = "n1-standard-2"
	defaultNodeMachineTypeAWS = "m3.medium"

	defaultBastionMachineTypeGCE = "f1-micro"
	defaultBastionMachineTypeAWS = "m3.medium"
)

func (c *cloud) MachineType(cluster *kops.Cluster, ig *kops.InstanceGroup) (string, error) {
	var machineType string

	switch ig.Spec.Role {
	case kops.InstanceGroupRoleMaster:
		switch c.cloud.ProviderID() {
		case kops.CloudProviderAWS:
			machineType = defaultMasterMachineTypeAWS
		case kops.CloudProviderGCE:
			machineType = defaultMasterMachineTypeGCE
		}

	case kops.InstanceGroupRoleNode:
		switch c.cloud.ProviderID() {
		case kops.CloudProviderAWS:
			machineType = defaultNodeMachineTypeAWS
		case kops.CloudProviderGCE:
			machineType = defaultNodeMachineTypeGCE
		}

	case kops.InstanceGroupRoleBastion:
		switch c.cloud.ProviderID() {
		case kops.CloudProviderAWS:
			machineType = defaultBastionMachineTypeAWS
		case kops.CloudProviderGCE:
			machineType = defaultBastionMachineTypeGCE
		}

	default:
		return "", fmt.Errorf("spotinst: unknown instance group role: %s", ig.Spec.Role)
	}

	return machineType, nil
}

func (c *cloud) Image(cluster *kops.Cluster, channel *kops.Channel) string {
	var image string

	if channel != nil {
		var kubernetesVersion *semver.Version
		if cluster.Spec.KubernetesVersion != "" {
			var err error
			kubernetesVersion, err = util.ParseKubernetesVersion(cluster.Spec.KubernetesVersion)
			if err != nil {
				glog.Warningf("spotinst: cannot parse KubernetesVersion %q in cluster", cluster.Spec.KubernetesVersion)
			}
		}
		if kubernetesVersion != nil {
			imageSpec := channel.FindImage(c.cloud.ProviderID(), *kubernetesVersion)
			if imageSpec != nil {
				image = imageSpec.Name
			}
		}
	}

	return image
}

func (c *cloud) Region() string {
	var region string

	switch c.cloud.ProviderID() {
	case kops.CloudProviderAWS:
		cloud := c.cloud.(awsup.AWSCloud)
		region = cloud.Region()
	case kops.CloudProviderGCE:
		cloud := c.cloud.(gce.GCECloud)
		region = cloud.Region()
	}

	return region
}

func (c *cloud) Tags() []string {
	tags := []string{"_spotinst"}

	switch c.cloud.ProviderID() {
	case kops.CloudProviderAWS:
		tags = append(tags, "_aws")
	case kops.CloudProviderGCE:
		tags = append(tags, "_gce")
	}

	return tags
}

const ResourceTypeGroup = "Group"

type Resource struct {
	ID   string
	Name string
	Type string
	Raw  interface{}
}

func (c *cloud) ListResources(clusterName string) (map[string]*Resource, error) {
	ctx := context.Background()
	resources := make(map[string]*Resource)

	glog.V(2).Info("Listing resources")
	switch c.cloud.ProviderID() {
	case kops.CloudProviderAWS:
		{
			out, err := c.service.CloudProviderAWS().List(ctx, nil)
			if err != nil {
				return nil, err
			}
			for _, group := range out.Groups {
				groupID := spotinst.StringValue(group.ID)
				groupName := spotinst.StringValue(group.Name)
				if strings.HasSuffix(groupName, clusterName) {
					resource := &Resource{
						Type: ResourceTypeGroup,
						ID:   groupID,
						Name: groupName,
						Raw:  group,
					}
					resourceKey := resource.Type + ":" + resource.ID
					resources[resourceKey] = resource
					glog.V(2).Infof("Discovered group: %s (%s)", groupID, groupName)
				}
			}
		}
	case kops.CloudProviderGCE:
		{
			// TODO(liran): Not implemented yet.
		}
	}

	return resources, nil
}

func (c *cloud) DeleteResource(resource interface{}) error {
	rs, ok := resource.(*Resource)
	if !ok {
		return fmt.Errorf("spotinst: unknown resource: %T", resource)
	}
	switch c.cloud.ProviderID() {
	case kops.CloudProviderAWS:
		{
			switch rs.Type {
			case ResourceTypeGroup:
				return c.DeleteGroup(&cloudinstances.CloudInstanceGroup{Raw: rs.Raw})
			}
		}
	case kops.CloudProviderGCE:
		{
			// TODO(liran): Not implemented yet.
		}
	}
	return nil
}

func (c *cloud) FindClusterStatus(cluster *kops.Cluster) (*kops.ClusterStatus, error) {
	var status *kops.ClusterStatus
	var err error

	switch c.cloud.ProviderID() {
	case kops.CloudProviderAWS:
		status, err = findEtcdStatusAWS(c.cloud.(awsup.AWSCloud), cluster)
		if err != nil {
			return nil, err
		}
	case kops.CloudProviderGCE:
		status, err = findEtcdStatusGCE(c.cloud.(gce.GCECloud), cluster)
		if err != nil {
			return nil, err
		}
	}

	return status, nil
}

func (c *cloud) GetApiIngressStatus(cluster *kops.Cluster) ([]kops.ApiIngressStatus, error) {
	var status []kops.ApiIngressStatus
	var err error

	switch c.cloud.ProviderID() {
	case kops.CloudProviderAWS:
		status, err = getApiIngressStatusAWS(c.cloud.(awsup.AWSCloud), cluster)
		if err != nil {
			return nil, err
		}
	case kops.CloudProviderGCE:
		status, err = getApiIngressStatusGCE(c.cloud.(gce.GCECloud), cluster)
		if err != nil {
			return nil, err
		}
	}

	return status, nil
}
