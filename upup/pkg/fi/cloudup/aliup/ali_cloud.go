/*
Copyright 2018 The Kubernetes Authors.

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
	"strings"

	common "github.com/denverdino/aliyungo/common"
	ecs "github.com/denverdino/aliyungo/ecs"

	"k8s.io/api/core/v1"
	prj "k8s.io/kops"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/upup/pkg/fi"
)

const TagClusterName = "KubernetesCluster"

// This is for statistic purpose.
var KubernetesKopsIdentity = fmt.Sprintf("Kubernetes.Kops/%s", prj.Version)

type ALICloud interface {
	fi.Cloud

	EcsClient() *ecs.Client
	Region() string
}

type aliCloudImplementation struct {
	ecsClient *ecs.Client

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
	c.tags = tags

	return c, nil
}

func (c *aliCloudImplementation) EcsClient() *ecs.Client {
	return c.ecsClient
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
	return errors.New("DeleteInstance not implemented on aliCloud")
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
	vswitcheList, _, err := c.EcsClient().DescribeVSwitches(describeVSwitchesArgs)
	if err != nil {
		return nil, fmt.Errorf("error listing VSwitchs: %v", err)
	}

	for _, vswitch := range vswitcheList {
		s := &fi.SubnetInfo{
			ID:   vswitch.VSwitchId,
			Zone: vswitch.ZoneId,
			CIDR: vswitch.CidrBlock,
		}
		vpcInfo.Subnets = append(vpcInfo.Subnets, s)
	}

	return vpcInfo, nil

}

func (c *aliCloudImplementation) GetCloudGroups(cluster *kops.Cluster, instancegroups []*kops.InstanceGroup, warnUnmatched bool, nodes []v1.Node) (map[string]*cloudinstances.CloudInstanceGroup, error) {
	return nil, fmt.Errorf("GetCloudGroups not implemented on aliCloud")
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

	if vpc == nil || len(vpc) == 0 {
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

		vswitcheList, _, err := aliCloud.EcsClient().DescribeVSwitches(describeVSwitchesArgs)
		if err != nil {
			return nil, fmt.Errorf("error listing VSwitchs: %v", err)
		}

		if len(vswitcheList) == 0 {
			return nil, fmt.Errorf("VSwitch %q not found", VSwitchId)
		}

		if len(vswitcheList) != 1 {
			return nil, fmt.Errorf("found multiple VSwitchs for %q", VSwitchId)
		}

		zone := vswitcheList[0].ZoneId
		if res[zone] != "" {
			return res, fmt.Errorf("vswitch %s and %s have the same zone", vswitcheList[0].VSwitchId, zone)
		}
		res[zone] = vswitcheList[0].VSwitchId

	}
	return res, nil
}

func getRegionByZones(zones []string) (string, error) {
	region := ""

	for _, zone := range zones {
		zoneSplit := strings.Split(zone, "-")
		zoneRegion := ""
		if len(zoneSplit) != 3 {
			return "", fmt.Errorf("invalid ALI zone: %q ", zone)
		}

		if len(zoneSplit[2]) == 1 {
			zoneRegion = zoneSplit[0] + "-" + zoneSplit[1]
		} else if len(zoneSplit[2]) == 2 {
			zoneRegion = zone[:len(zone)-1]
		} else {
			return "", fmt.Errorf("invalid ALI zone: %q ", zone)
		}

		if region != "" && zoneRegion != region {
			return "", fmt.Errorf("clusters cannot span multiple regions (found zone %q, but region is %q)", zone, region)
		}
		region = zoneRegion
	}

	return region, nil
}
