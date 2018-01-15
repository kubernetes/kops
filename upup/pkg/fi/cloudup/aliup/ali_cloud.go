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

	"github.com/golang/glog"

	common "github.com/denverdino/aliyungo/common"
	ecs "github.com/denverdino/aliyungo/ecs"
	"k8s.io/api/core/v1"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/upup/pkg/fi"
)

const TagClusterName = "KubernetesCluster"
const TagNameRolePrefix = "k8s.io/role/"
const TagNameEtcdClusterPrefix = "k8s.io/etcd/"

type ALICloud interface {
	fi.Cloud

	EcsClient() *ecs.Client
	Region() string
	AddClusterTags(tags map[string]string)
	GetTags(resourceId string, resourceType string) (map[string]string, error)
	CreateTags(resourceId string, resourceType string, tags map[string]string) error
	RemoveTags(resourceId string, resourceType string, tags map[string]string) error
	GetClusterTags() map[string]string
}

type aliCloudImplementation struct {
	ecsClient *ecs.Client
	region    string
	tags      map[string]string
}

var _ fi.Cloud = &aliCloudImplementation{}

// NewALICloud returns a Cloud, expecting the env vars ALIYUN_ACCESS_KEY_ID && ALIYUN_ACCESS_KET_SECRET
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

	escclient := ecs.NewClient(accessKeyId, accessKeySecret)
	c.ecsClient = escclient
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
	return nil, fmt.Errorf("DNS not implemented on aliCloud")
}

func (c *aliCloudImplementation) DeleteGroup(g *cloudinstances.CloudInstanceGroup) error {
	return fmt.Errorf("DeleteGroup not implemented on aliCloud")
}

func (c *aliCloudImplementation) DeleteInstance(i *cloudinstances.CloudInstanceGroupMember) error {
	return fmt.Errorf("DeleteInstance not implemented on aliCloud")
}

func (c *aliCloudImplementation) FindVPCInfo(id string) (*fi.VPCInfo, error) {
	return nil, fmt.Errorf("FindVPCInfo not implemented on aliCloud")
}

func (c *aliCloudImplementation) GetCloudGroups(cluster *kops.Cluster, instancegroups []*kops.InstanceGroup, warnUnmatched bool, nodes []v1.Node) (map[string]*cloudinstances.CloudInstanceGroup, error) {
	return nil, fmt.Errorf("GetCloudGroups not implemented on aliCloud")
}

// GetTags will get the specified resource's tags.
func (c *aliCloudImplementation) GetTags(resourceId string, resourceType string) (map[string]string, error) {
	if resourceId == "" {
		return nil, fmt.Errorf("resourceId not provided to GetTags")
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
	for k, v := range c.tags {
		tags[k] = v
	}
}

// CreateTags will add tags to the specified resource.
func (c *aliCloudImplementation) CreateTags(resourceId string, resourceType string, tags map[string]string) error {
	if len(tags) == 0 {
		return nil
	} else if len(tags) > 10 {
		glog.V(4).Info("The number of specified resource's tags exceeds 10, resourceId:%q", resourceId)
	}
	if resourceId == "" {
		return fmt.Errorf("resourceId not provided to CreateTags")
	}
	if resourceType == "" {
		return fmt.Errorf("resourceType not provided to CreateTags")
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
		return fmt.Errorf("resourceId not provided to RemoveTags")
	}
	if resourceType == "" {
		return fmt.Errorf("resourceType not provided to RemoveTags")
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
