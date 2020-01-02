/*
Copyright 2019 The Kubernetes Authors.

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

package aliup

import (
	"errors"
	"fmt"
	"os"
	"time"

	"k8s.io/klog"

	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	"github.com/denverdino/aliyungo/ess"
	"github.com/denverdino/aliyungo/ram"
	"github.com/denverdino/aliyungo/slb"

	prj "k8s.io/kops"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/upup/pkg/fi"
)

const (
	TagClusterName           = "KubernetesCluster"
	TagNameRolePrefix        = "k8s.io/role/"
	TagNameEtcdClusterPrefix = "k8s.io/etcd/"
	TagRoleMaster            = "master"
)

// This is for statistic purpose.
var KubernetesKopsIdentity = fmt.Sprintf("Kubernetes.Kops/%s", prj.Version)

type ALICloud interface {
	fi.Cloud

	EcsClient() *ecs.Client
	SlbClient() *slb.Client
	RamClient() *ram.RamClient
	EssClient() *ess.Client
	VpcClient() *ecs.Client

	AddClusterTags(tags map[string]string)
	GetTags(resourceId string, resourceType string) (map[string]string, error)
	CreateTags(resourceId string, resourceType string, tags map[string]string) error
	RemoveTags(resourceId string, resourceType string, tags map[string]string) error
	GetClusterTags() map[string]string
	FindClusterStatus(cluster *kops.Cluster) (*kops.ClusterStatus, error)
	GetApiIngressStatus(cluster *kops.Cluster) ([]kops.ApiIngressStatus, error)
}

type aliCloudImplementation struct {
	ecsClient *ecs.Client
	slbClient *slb.Client
	ramClient *ram.RamClient
	essClient *ess.Client
	vpcClient *ecs.Client

	region string
	tags   map[string]string
}

var _ fi.Cloud = &aliCloudImplementation{}

// NewALICloud returns a Cloud, expecting the env vars ALIYUN_ACCESS_KEY_ID && ALIYUN_ACCESS_KEY_SECRET
// NewALICloud will return an err if env vars are not defined
func NewALICloud(region string, tags map[string]string) (ALICloud, error) {

	c := &aliCloudImplementation{region: region}

	accessKeyId := os.Getenv("ALIYUN_ACCESS_KEY_ID")
	if accessKeyId == "" {
		return nil, errors.New("ALIYUN_ACCESS_KEY_ID is required")
	}
	accessKeySecret := os.Getenv("ALIYUN_ACCESS_KEY_SECRET")
	if accessKeySecret == "" {
		return nil, errors.New("ALIYUN_ACCESS_KEY_SECRET is required")
	}

	c.ecsClient = ecs.NewClient(accessKeyId, accessKeySecret)
	c.ecsClient.SetUserAgent(KubernetesKopsIdentity)
	c.slbClient = slb.NewClient(accessKeyId, accessKeySecret)
	ramclient := ram.NewClient(accessKeyId, accessKeySecret)
	c.ramClient = ramclient.(*ram.RamClient)
	c.essClient = ess.NewClient(accessKeyId, accessKeySecret)
	c.vpcClient = ecs.NewVPCClient(accessKeyId, accessKeySecret, common.Region(region))

	c.tags = tags

	return c, nil
}

func (c *aliCloudImplementation) EcsClient() *ecs.Client {
	return c.ecsClient
}

func (c *aliCloudImplementation) SlbClient() *slb.Client {
	return c.slbClient
}

func (c *aliCloudImplementation) RamClient() *ram.RamClient {
	return c.ramClient
}

func (c *aliCloudImplementation) EssClient() *ess.Client {
	return c.essClient
}

func (c *aliCloudImplementation) VpcClient() *ecs.Client {
	return c.vpcClient
}

func (c *aliCloudImplementation) Region() string {
	return c.region
}

func (c *aliCloudImplementation) ProviderID() kops.CloudProviderID {
	return kops.CloudProviderALI
}

func (c *aliCloudImplementation) DNS() (dnsprovider.Interface, error) {
	return nil, errors.New("DNS not implemented on aliCloud")
}

func (c *aliCloudImplementation) DeleteGroup(g *cloudinstances.CloudInstanceGroup) error {
	return errors.New("DeleteGroup not implemented on aliCloud")
}

func (c *aliCloudImplementation) DeleteInstance(i *cloudinstances.CloudInstanceGroupMember) error {
	id := i.ID
	if id == "" {
		return fmt.Errorf("id was not set on CloudInstanceGroupMember: %v", i)
	}

	if err := c.EcsClient().StopInstance(id, false); err != nil {
		return fmt.Errorf("error stopping instance %q: %v", id, err)
	}

	// Wait for 3 min to stop the instance
	for i := 0; i < 36; i++ {
		ins, err := c.EcsClient().DescribeInstanceAttribute(id)
		if err != nil {
			return fmt.Errorf("error describing instance %q: %v", id, err)
		}

		klog.V(8).Infof("stopping Alicloud ecs instance %q, current Status: %q", id, ins.Status)
		time.Sleep(time.Second * 5)

		if ins.Status == ecs.Stopped {
			break
		}

		if i == 35 {
			return fmt.Errorf("fail to stop ecs instance %s in 3 mins", id)
		}
	}

	if err := c.EcsClient().DeleteInstance(id); err != nil {
		return fmt.Errorf("error deleting instance %q: %v", id, err)
	}

	klog.V(8).Infof("deleted Alicloud ecs instance %q", id)

	return nil
}

func (c *aliCloudImplementation) FindVPCInfo(id string) (*fi.VPCInfo, error) {
	request := &ecs.DescribeVpcsArgs{
		RegionId: common.Region(c.Region()),
		VpcId:    id,
	}
	vpcs, _, err := c.EcsClient().DescribeVpcs(request)
	if err != nil {
		return nil, fmt.Errorf("error listing VPCs: %v", err)
	}

	if len(vpcs) != 1 {
		return nil, fmt.Errorf("found multiple VPCs for %q", id)
	}
	vpcInfo := &fi.VPCInfo{
		CIDR: vpcs[0].CidrBlock,
	}

	describeVSwitchesArgs := &ecs.DescribeVSwitchesArgs{
		VpcId:    id,
		RegionId: common.Region(c.Region()),
	}
	vswitchList, _, err := c.EcsClient().DescribeVSwitches(describeVSwitchesArgs)
	if err != nil {
		return nil, fmt.Errorf("error listing VSwitchs: %v", err)
	}

	for _, vswitch := range vswitchList {
		s := &fi.SubnetInfo{
			ID:   vswitch.VSwitchId,
			Zone: vswitch.ZoneId,
			CIDR: vswitch.CidrBlock,
		}
		vpcInfo.Subnets = append(vpcInfo.Subnets, s)
	}

	return vpcInfo, nil

}

// GetTags will get the specified resource's tags.
func (c *aliCloudImplementation) GetTags(resourceId string, resourceType string) (map[string]string, error) {
	if resourceId == "" {
		return nil, errors.New("resourceId not provided to GetTags")
	}
	tags := map[string]string{}

	request := &ecs.DescribeTagsArgs{
		RegionId:     common.Region(c.Region()),
		ResourceType: ecs.TagResourceType(resourceType), //image, instance, snapshot or disk
		ResourceId:   resourceId,
	}
	responseTags, _, err := c.EcsClient().DescribeTags(request)
	if err != nil {
		return tags, fmt.Errorf("error getting tags on %v: %v", resourceId, err)
	}

	for _, tag := range responseTags {
		tags[tag.TagKey] = tag.TagValue
	}
	return tags, nil

}

// AddClusterTags will add ClusterTags to resources (in ALI, only disk, instance, snapshot or image can be tagged )
func (c *aliCloudImplementation) AddClusterTags(tags map[string]string) {

	if c.tags != nil && len(c.tags) != 0 && tags != nil {
		for k, v := range c.tags {
			tags[k] = v
		}
	}
}

// CreateTags will add tags to the specified resource.
func (c *aliCloudImplementation) CreateTags(resourceId string, resourceType string, tags map[string]string) error {
	if len(tags) == 0 {
		return nil
	} else if len(tags) > 10 {
		klog.V(4).Infof("The number of specified resource's tags exceeds 10, resourceId:%q", resourceId)
	}
	if resourceId == "" {
		return errors.New("resourceId not provided to CreateTags")
	}
	if resourceType == "" {
		return errors.New("resourceType not provided to CreateTags")
	}

	request := &ecs.AddTagsArgs{
		ResourceId:   resourceId,
		ResourceType: ecs.TagResourceType(resourceType), //image, instance, snapshot or disk
		RegionId:     common.Region(c.Region()),
		Tag:          tags,
	}
	err := c.EcsClient().AddTags(request)
	if err != nil {
		return fmt.Errorf("error creating tags on %v: %v", resourceId, err)
	}

	return nil
}

// RemoveTags will remove tags from the specified resource.
func (c *aliCloudImplementation) RemoveTags(resourceId string, resourceType string, tags map[string]string) error {
	if len(tags) == 0 {
		return nil
	}
	if resourceId == "" {
		return errors.New("resourceId not provided to RemoveTags")
	}
	if resourceType == "" {
		return errors.New("resourceType not provided to RemoveTags")
	}

	request := &ecs.RemoveTagsArgs{
		ResourceId:   resourceId,
		ResourceType: ecs.TagResourceType(resourceType), //image, instance, snapshot or disk
		RegionId:     common.Region(c.Region()),
		Tag:          tags,
	}
	err := c.EcsClient().RemoveTags(request)
	if err != nil {
		return fmt.Errorf("error removing tags on %v: %v", resourceId, err)
	}

	return nil
}

// GetClusterTags will get the ClusterTags
func (c *aliCloudImplementation) GetClusterTags() map[string]string {
	return c.tags
}

func (c *aliCloudImplementation) GetApiIngressStatus(cluster *kops.Cluster) ([]kops.ApiIngressStatus, error) {
	var ingresses []kops.ApiIngressStatus
	name := "api." + cluster.Name

	describeLoadBalancersArgs := &slb.DescribeLoadBalancersArgs{
		RegionId:         common.Region(c.Region()),
		LoadBalancerName: name,
	}

	responseLoadBalancers, err := c.SlbClient().DescribeLoadBalancers(describeLoadBalancersArgs)
	if err != nil {
		return nil, fmt.Errorf("error finding LoadBalancers: %v", err)
	}
	// Don't exist loadbalancer with specified ClusterTags or Name.
	if len(responseLoadBalancers) == 0 {
		return nil, nil
	}
	if len(responseLoadBalancers) > 1 {
		klog.V(4).Infof("The number of specified loadbalancer with the same name exceeds 1, loadbalancerName:%q", name)
	}

	address := responseLoadBalancers[0].Address
	ingresses = append(ingresses, kops.ApiIngressStatus{IP: address})

	return ingresses, nil
}

func ZoneToVSwitchID(VPCID string, zones []string, vswitchIDs []string) (map[string]string, error) {
	regionId, err := getRegionByZones(zones)
	if err != nil {
		return nil, err
	}

	res := make(map[string]string)
	cloudTags := map[string]string{}
	aliCloud, err := NewALICloud(regionId, cloudTags)
	if err != nil {
		return res, fmt.Errorf("error loading cloud: %v", err)
	}

	describeVpcsArgs := &ecs.DescribeVpcsArgs{
		RegionId: common.Region(regionId),
		VpcId:    VPCID,
	}

	vpc, _, err := aliCloud.EcsClient().DescribeVpcs(describeVpcsArgs)
	if err != nil {
		return res, fmt.Errorf("error describing VPC: %v", err)
	}

	if len(vpc) == 0 {
		return res, fmt.Errorf("VPC %q not found", VPCID)
	}

	if len(vpc) != 1 {
		return nil, fmt.Errorf("found multiple VPCs for %q", VPCID)
	}
	subnetByID := make(map[string]string)
	for _, VSId := range vpc[0].VSwitchIds.VSwitchId {
		subnetByID[VSId] = VSId
	}

	for _, VSwitchId := range vswitchIDs {

		_, ok := subnetByID[VSwitchId]
		if !ok {
			return res, fmt.Errorf("vswitch %s not found in VPC %s", VSwitchId, VPCID)
		}
		describeVSwitchesArgs := &ecs.DescribeVSwitchesArgs{
			VpcId:     vpc[0].VpcId,
			RegionId:  common.Region(regionId),
			VSwitchId: VSwitchId,
		}

		vswitchList, _, err := aliCloud.EcsClient().DescribeVSwitches(describeVSwitchesArgs)
		if err != nil {
			return nil, fmt.Errorf("error listing VSwitchs: %v", err)
		}

		if len(vswitchList) == 0 {
			return nil, fmt.Errorf("VSwitch %q not found", VSwitchId)
		}

		if len(vswitchList) != 1 {
			return nil, fmt.Errorf("found multiple VSwitchs for %q", VSwitchId)
		}

		zone := vswitchList[0].ZoneId
		if res[zone] != "" {
			return res, fmt.Errorf("vswitch %s and %s have the same zone", vswitchList[0].VSwitchId, zone)
		}
		res[zone] = vswitchList[0].VSwitchId

	}
	return res, nil
}
