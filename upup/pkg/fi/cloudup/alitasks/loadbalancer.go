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

package alitasks

import (
	"encoding/json"
	"fmt"

	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/slb"
	"k8s.io/klog"

	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/aliup"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

// LoadBalancer represents a ALI Cloud LoadBalancer
//go:generate fitask -type=LoadBalancer

type LoadBalancer struct {
	Name                *string
	LoadbalancerId      *string
	AddressType         *string
	VSwitchId           *string
	LoadBalancerAddress *string
	Lifecycle           *fi.Lifecycle
	Tags                map[string]string
}

var _ fi.CompareWithID = &LoadBalancer{}
var _ fi.HasAddress = &LoadBalancer{}

func (l *LoadBalancer) CompareWithID() *string {
	return l.LoadbalancerId
}

func (l *LoadBalancer) Find(c *fi.Context) (*LoadBalancer, error) {
	cloud := c.Cloud.(aliup.ALICloud)

	// TODO:Get loadbalancer with LoadBalancerName, hope to support finding with tags
	describeLoadBalancersArgs := &slb.DescribeLoadBalancersArgs{
		RegionId:         common.Region(cloud.Region()),
		LoadBalancerName: fi.StringValue(l.Name),
		AddressType:      slb.AddressType(fi.StringValue(l.AddressType)),
	}

	responseLoadBalancers, err := cloud.SlbClient().DescribeLoadBalancers(describeLoadBalancersArgs)
	if err != nil {
		return nil, fmt.Errorf("error finding LoadBalancers: %v", err)
	}

	// There's no loadbalancer with specified ClusterTags or Name.
	if len(responseLoadBalancers) == 0 {
		klog.V(4).Infof("can't find loadbalancer with name: %q", *l.Name)
		return nil, nil
	}
	if len(responseLoadBalancers) > 1 {
		return nil, fmt.Errorf("more than 1 loadbalancer is found with name: %q", *l.Name)
	}

	klog.V(2).Infof("found matching LoadBalancer: %q", *l.Name)
	lb := responseLoadBalancers[0]

	actual := &LoadBalancer{
		Name:                fi.String(lb.LoadBalancerName),
		AddressType:         fi.String(string(lb.AddressType)),
		LoadbalancerId:      fi.String(lb.LoadBalancerId),
		LoadBalancerAddress: fi.String(lb.Address),
		VSwitchId:           fi.String(lb.VSwitchId),
	}

	describeTagsArgs := &slb.DescribeTagsArgs{
		RegionId:       common.Region(cloud.Region()),
		LoadBalancerID: fi.StringValue(actual.LoadbalancerId),
	}
	tags, _, err := cloud.SlbClient().DescribeTags(describeTagsArgs)
	if err != nil {
		return nil, fmt.Errorf("error getting tags on loadbalancer: %v", err)
	}

	if len(tags) != 0 {
		actual.Tags = make(map[string]string)
		for _, tag := range tags {
			key := tag.TagKey
			value := tag.TagValue
			actual.Tags[key] = value
		}
	}
	// Ignore "system" fields
	l.LoadbalancerId = actual.LoadbalancerId
	actual.Lifecycle = l.Lifecycle
	return actual, nil
}

func (l *LoadBalancer) FindIPAddress(context *fi.Context) (*string, error) {
	cloud := context.Cloud.(aliup.ALICloud)

	// TODO:Get loadbalancer with LoadBalancerName, hope to support finding with tags
	describeLoadBalancersArgs := &slb.DescribeLoadBalancersArgs{
		RegionId:         common.Region(cloud.Region()),
		LoadBalancerName: fi.StringValue(l.Name),
		AddressType:      slb.AddressType(fi.StringValue(l.AddressType)),
	}

	responseLoadBalancers, err := cloud.SlbClient().DescribeLoadBalancers(describeLoadBalancersArgs)
	if err != nil {
		return nil, fmt.Errorf("error finding LoadBalancers: %v", err)
	}

	// Don't exist loadbalancer with specified ClusterTags or Name.
	if len(responseLoadBalancers) == 0 {
		return nil, nil
	}
	if len(responseLoadBalancers) > 1 {
		klog.V(4).Infof("The number of specified loadbalancer with the same name exceeds 1, loadbalancerName:%q", *l.Name)
	}

	address := responseLoadBalancers[0].Address
	return &address, nil
}

func (l *LoadBalancer) Run(c *fi.Context) error {
	if l.Tags == nil {
		l.Tags = make(map[string]string)
	}
	c.Cloud.(aliup.ALICloud).AddClusterTags(l.Tags)
	return fi.DefaultDeltaRunMethod(l, c)
}

func (_ *LoadBalancer) CheckChanges(a, e, changes *LoadBalancer) error {
	if a == nil {
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
		if e.AddressType == nil {
			return fi.RequiredField("AddressType")
		}
	} else {
		if changes.AddressType != nil {
			return fi.CannotChangeField("AddressType")
		}
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
	}
	return nil
}

//LoadBalancer can only modify tags.
func (_ *LoadBalancer) RenderALI(t *aliup.ALIAPITarget, a, e, changes *LoadBalancer) error {
	if a == nil {
		klog.V(2).Infof("Creating LoadBalancer with Name:%q", fi.StringValue(e.Name))

		createLoadBalancerArgs := &slb.CreateLoadBalancerArgs{
			RegionId:         common.Region(t.Cloud.Region()),
			LoadBalancerName: fi.StringValue(e.Name),
			AddressType:      slb.AddressType(fi.StringValue(e.AddressType)),
			VSwitchId:        fi.StringValue(e.VSwitchId),
		}
		response, err := t.Cloud.SlbClient().CreateLoadBalancer(createLoadBalancerArgs)
		if err != nil {
			return fmt.Errorf("error creating loadbalancer: %v", err)
		}
		e.LoadbalancerId = fi.String(response.LoadBalancerId)
		e.LoadBalancerAddress = fi.String(response.Address)
	}

	if changes != nil && changes.Tags != nil {
		tagItems := e.jsonMarshalTags(e.Tags)
		addTagsArgs := &slb.AddTagsArgs{
			RegionId:       common.Region(t.Cloud.Region()),
			LoadBalancerID: fi.StringValue(e.LoadbalancerId),
			Tags:           string(tagItems),
		}
		err := t.Cloud.SlbClient().AddTags(addTagsArgs)
		if err != nil {
			return fmt.Errorf("error adding Tags to Loadbalancer: %v", err)
		}
	}

	if a != nil && (len(a.Tags) > 0) {
		klog.V(2).Infof("Modifying LoadBalancer with Name:%q, update LoadBalancer tags", fi.StringValue(e.Name))

		tagsToDelete := e.getLoadBalancerTagsToDelete(a.Tags)
		if len(tagsToDelete) > 0 {
			tagItems := e.jsonMarshalTags(tagsToDelete)
			removeTagsArgs := &slb.RemoveTagsArgs{
				RegionId:       common.Beijing,
				LoadBalancerID: fi.StringValue(a.LoadbalancerId),
				Tags:           string(tagItems),
			}
			if err := t.Cloud.SlbClient().RemoveTags(removeTagsArgs); err != nil {
				return fmt.Errorf("error removing Tags from LoadBalancer: %v", err)
			}
		}
	}

	return nil
}

// getDiskTagsToDelete loops through the currently set tags and builds a list of tags to be deleted from the specified disk
func (l *LoadBalancer) getLoadBalancerTagsToDelete(currentTags map[string]string) map[string]string {
	tagsToDelete := map[string]string{}
	for k, v := range currentTags {
		if _, ok := l.Tags[k]; !ok {
			tagsToDelete[k] = v
		}
	}

	return tagsToDelete
}

func (l *LoadBalancer) jsonMarshalTags(tags map[string]string) string {
	tagItemArr := []slb.TagItem{}
	tagItem := slb.TagItem{}
	for key, value := range tags {
		tagItem.TagKey = key
		tagItem.TagValue = value
		tagItemArr = append(tagItemArr, tagItem)
	}
	tagItems, _ := json.Marshal(tagItemArr)

	return string(tagItems)
}

type terraformLoadBalancer struct {
	Name     *string `json:"name,omitempty" cty:"name"`
	Internet *bool   `json:"internet,omitempty" cty:"internet"`
}

func (_ *LoadBalancer) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *LoadBalancer) error {
	tf := &terraformLoadBalancer{
		Name: e.Name,
	}

	if slb.AddressType(fi.StringValue(e.AddressType)) == slb.InternetAddressType {
		internet := true
		tf.Internet = &internet
	} else {
		internet := false
		tf.Internet = &internet
	}

	return t.RenderResource("alicloud_slb", *e.Name, tf)
}

func (l *LoadBalancer) TerraformLink() *terraform.Literal {
	return terraform.LiteralProperty("alicloud_slb", *l.Name, "id")
}
