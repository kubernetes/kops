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
	"sort"

	"k8s.io/klog"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	slbnew "github.com/aliyun/alibaba-cloud-sdk-go/services/slb"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/aliup"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

// +kops:fitask
type LoadBalancerACL struct {
	ID                   *string
	Name                 *string
	LoadBalancer         *LoadBalancer
	LoadBalancerListener *LoadBalancerListener
	SourceItems          []*string
	Lifecycle            *fi.Lifecycle
}

type AclEntry struct {
	Entry   string `json:"entry"`
	Comment string `json:"comment"`
}

var _ fi.CompareWithID = &LoadBalancerACL{}

func (l *LoadBalancerACL) CompareWithID() *string {
	return l.Name
}

func (l *LoadBalancerACL) Find(c *fi.Context) (*LoadBalancerACL, error) {
	if l.LoadBalancer == nil || l.LoadBalancer.LoadbalancerId == nil {
		klog.V(4).Infof("LoadBalancer / LoadbalancerId not found for %s, skipping Find", fi.StringValue(l.Name))
		return nil, nil
	}
	if l.LoadBalancerListener == nil || l.LoadBalancerListener.ListenerPort == nil {
		klog.V(4).Infof("LoadBalancerListener / LoadbalancerListenerPort not found for %s, skipping Find", fi.StringValue(l.Name))
		return nil, nil
	}

	cloud := c.Cloud.(aliup.ALICloud)

	describeAclReq := slbnew.CreateDescribeAccessControlListsRequest()
	describeAclReq.AclName = fi.StringValue(l.Name)

	describeAclResp, err := cloud.SLB().DescribeAccessControlLists(describeAclReq)
	if err != nil {
		return nil, fmt.Errorf("error listing LoadBalancerAccessControlList: %v", err)
	}
	acls := describeAclResp.Acls.Acl
	if len(acls) == 0 {
		return nil, nil
	}

	if len(acls) > 1 {
		return nil, fmt.Errorf("found multiple LoadBalancerAccessControlList with name %s", fi.StringValue(l.Name))
	}

	acl := acls[0]

	klog.V(2).Infof("found matching LoadBalancerAccessControlList: %s", acl.AclId)

	describeAclAttrReq := slbnew.CreateDescribeAccessControlListAttributeRequest()
	describeAclAttrReq.AclId = acl.AclId

	describeAclAttrResp, err := cloud.SLB().DescribeAccessControlListAttribute(describeAclAttrReq)
	if err != nil {
		return nil, fmt.Errorf("error describing LoadBalancerAccessControlListAttribute: %v", err)
	}

	var sourceItems []*string

	for _, entry := range describeAclAttrResp.AclEntrys.AclEntry {
		ip := entry.AclEntryIP
		sourceItems = append(sourceItems, &ip)
	}

	actual := &LoadBalancerACL{
		ID:          fi.String(acl.AclId),
		Name:        fi.String(describeAclAttrResp.AclName),
		SourceItems: sourceItems,
	}

	listeners := describeAclAttrResp.RelatedListeners.RelatedListener

	if len(listeners) != 1 {
		actual.LoadBalancerListener = nil
		actual.LoadBalancer = nil
	} else {
		listener := listeners[0]
		lb := &LoadBalancer{LoadbalancerId: fi.String(listener.LoadBalancerId)}
		actual.LoadBalancer = lb
		actual.LoadBalancerListener = &LoadBalancerListener{
			LoadBalancer: lb,
			ListenerPort: fi.Int(listener.ListenerPort),
		}
	}

	// Ignore "system" fields
	l.ID = actual.ID
	actual.Lifecycle = l.Lifecycle

	return actual, nil
}

func (l *LoadBalancerACL) Run(c *fi.Context) error {
	l.Normalize()
	return fi.DefaultDeltaRunMethod(l, c)
}

func (l *LoadBalancerACL) Normalize() {
	// We need to sort our arrays consistently, so we don't get spurious changes
	sort.Stable(StringPointers(l.SourceItems))
}

// StringPointers implements sort.Interface for []*string
type StringPointers []*string

func (s StringPointers) Len() int      { return len(s) }
func (s StringPointers) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s StringPointers) Less(i, j int) bool {
	return fi.StringValue(s[i]) < fi.StringValue(s[j])
}

func (_ *LoadBalancerACL) CheckChanges(a, e, changes *LoadBalancerACL) error {
	if a == nil {
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
	}
	return nil
}

func (_ *LoadBalancerACL) RenderALI(t *aliup.ALIAPITarget, a, e, changes *LoadBalancerACL) error {
	if a == nil {
		klog.V(2).Infof("Creating LoadBalancerAccessControlList with name: %q for SLB %q", *e.Name, *e.LoadBalancer.LoadbalancerId)
		if err := createAcl(t.Cloud, e); err != nil {
			return err
		}
		return e.on(t.Cloud)
	} else {
		if changes.SourceItems != nil {
			klog.V(2).Info("Turning off LoadBalancerACL for SLB")
			if err := e.off(t.Cloud); err != nil {
				return err
			}

			klog.V(2).Infof("Deleting LoadBalancerAccessControlList %q", *a.Name)
			err := deleteAcl(t.Cloud, a)
			if err != nil {
				return fmt.Errorf("error deleting LoadBalancerAccessControlList: %v", err)
			}

			klog.V(2).Infof("Creating LoadBalancerAccessControlList with name: %q for SLB %q", *e.Name, *e.LoadBalancer.LoadbalancerId)
			if err := createAcl(t.Cloud, e); err != nil {
				return err
			}

			return e.on(t.Cloud)
		}

		if changes.LoadBalancer != nil || changes.LoadBalancerListener != nil {
			klog.V(2).Info("Turning on LoadBalancerACL for SLB")
			return e.on(t.Cloud)
		}
	}

	return nil
}

func (l *LoadBalancerACL) on(alicloud aliup.ALICloud) error {
	setLBTCPlistenerAttrReq := slbnew.CreateSetLoadBalancerTCPListenerAttributeRequest()
	setLBTCPlistenerAttrReq.AclId = fi.StringValue(l.ID)
	setLBTCPlistenerAttrReq.AclType = "white"
	setLBTCPlistenerAttrReq.AclStatus = "on"
	setLBTCPlistenerAttrReq.LoadBalancerId = fi.StringValue(l.LoadBalancer.LoadbalancerId)
	setLBTCPlistenerAttrReq.ListenerPort = requests.NewInteger(fi.IntValue(l.LoadBalancerListener.ListenerPort))

	_, err := alicloud.SLB().SetLoadBalancerTCPListenerAttribute(setLBTCPlistenerAttrReq)
	if err != nil {
		return fmt.Errorf("error turning on LoadBalancerACL %v", err)
	}

	return nil
}

func (l *LoadBalancerACL) off(alicloud aliup.ALICloud) error {
	setLBTCPlistenerAttrReq := slbnew.CreateSetLoadBalancerTCPListenerAttributeRequest()
	setLBTCPlistenerAttrReq.AclStatus = "off"
	setLBTCPlistenerAttrReq.LoadBalancerId = fi.StringValue(l.LoadBalancer.LoadbalancerId)
	setLBTCPlistenerAttrReq.ListenerPort = requests.NewInteger(fi.IntValue(l.LoadBalancerListener.ListenerPort))

	_, err := alicloud.SLB().SetLoadBalancerTCPListenerAttribute(setLBTCPlistenerAttrReq)
	if err != nil {
		return fmt.Errorf("error turning off LoadBalancerACL %v", err)
	}

	return nil
}

func createAcl(alicloud aliup.ALICloud, acl *LoadBalancerACL) error {
	createAclReq := slbnew.CreateCreateAccessControlListRequest()
	createAclReq.AclName = fi.StringValue(acl.Name)

	resp, err := alicloud.SLB().CreateAccessControlList(createAclReq)
	if err != nil {
		return fmt.Errorf("error creating LoadBalancerAccessControlList: %v", err)
	}

	aclID := resp.AclId
	acl.ID = fi.String(aclID)

	var aclEntries []AclEntry
	for _, each := range acl.SourceItems {
		aclEntries = append(aclEntries, AclEntry{Entry: *each})
	}

	aclEntriesBytes, err := json.Marshal(aclEntries)
	if err != nil {
		return fmt.Errorf("error marshalling %v : %v", aclEntries, err)
	}

	addAclEntryReq := slbnew.CreateAddAccessControlListEntryRequest()
	addAclEntryReq.AclId = aclID
	addAclEntryReq.AclEntrys = string(aclEntriesBytes)

	_, err = alicloud.SLB().AddAccessControlListEntry(addAclEntryReq)
	if err != nil {
		return fmt.Errorf("error adding AclEntries %v: %v", addAclEntryReq, err)
	}

	return nil
}

func deleteAcl(alicloud aliup.ALICloud, acl *LoadBalancerACL) error {
	deleteAclReq := slbnew.CreateDeleteAccessControlListRequest()
	deleteAclReq.AclId = fi.StringValue(acl.ID)

	_, err := alicloud.SLB().DeleteAccessControlList(deleteAclReq)
	if err != nil {
		return fmt.Errorf("error deleting LoadBalancerAccessControlList %q, %v", fi.StringValue(acl.ID), err)
	}

	return nil
}

func (_ *LoadBalancerACL) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *LoadBalancerACL) error {
	klog.Warningf("terraform does not support LoadBalancerAccessControlList on ALI cloud")
	return nil
}
