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

	ecs "github.com/denverdino/aliyungo/ecs"
	"k8s.io/api/core/v1"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/upup/pkg/fi"
)

const TagClusterName = "KubernetesCluster"

type ALICloud interface {
	fi.Cloud

	EcsClient() *ecs.Client
	Region() string
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
	accessKeySecret := os.Getenv("ALIYUN_ACCESS_KET_SECRET")
	if accessKeySecret == "" {
		return nil, errors.New("ALIYUN_ACCESS_KET_SECRET is required")
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
