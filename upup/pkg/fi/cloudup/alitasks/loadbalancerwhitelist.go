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
	"fmt"
	"strings"

	"k8s.io/klog"

	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/aliup"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

//go:generate fitask -type=LoadBalancerWhiteList
type LoadBalancerWhiteList struct {
	LoadBalancer         *LoadBalancer
	LoadBalancerListener *LoadBalancerListener
	Name                 *string
	SourceItems          *string
	Lifecycle            *fi.Lifecycle
}

var _ fi.CompareWithID = &LoadBalancerWhiteList{}

func (l *LoadBalancerWhiteList) CompareWithID() *string {
	return l.Name
}

func (l *LoadBalancerWhiteList) Find(c *fi.Context) (*LoadBalancerWhiteList, error) {
	if l.LoadBalancer == nil || l.LoadBalancer.LoadbalancerId == nil {
		klog.V(4).Infof("LoadBalancer / LoadbalancerId not found for %s, skipping Find", fi.StringValue(l.Name))
		return nil, nil
	}
	if l.LoadBalancerListener == nil || l.LoadBalancerListener.ListenerPort == nil {
		klog.V(4).Infof("LoadBalancerListener / LoadbalancerListenerPort not found for %s, skipping Find", fi.StringValue(l.Name))
		return nil, nil
	}

	cloud := c.Cloud.(aliup.ALICloud)
	loadBalancerId := fi.StringValue(l.LoadBalancer.LoadbalancerId)
	listenertPort := fi.IntValue(l.LoadBalancerListener.ListenerPort)

	response, err := cloud.SlbClient().DescribeListenerAccessControlAttribute(loadBalancerId, listenertPort)
	if err != nil {
		return nil, fmt.Errorf("error finding LoadBalancerWhiteList: %v", err)
	}

	if response.SourceItems == "" {
		klog.V(2).Infof("can't find matching LoadBalancerWhiteList of ListenerPort: %v", listenertPort)
		return nil, nil
	}

	klog.V(2).Infof("found matching LoadBalancerWhiteList of ListenerPort: %v", listenertPort)
	actual := &LoadBalancerWhiteList{}
	actual.SourceItems = fi.String(response.SourceItems)

	// Ignore "system" fields
	actual.Name = l.Name
	actual.LoadBalancer = l.LoadBalancer
	actual.LoadBalancerListener = l.LoadBalancerListener
	actual.Lifecycle = l.Lifecycle

	return actual, nil
}

func (l *LoadBalancerWhiteList) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(l, c)
}

func (_ *LoadBalancerWhiteList) CheckChanges(a, e, changes *LoadBalancerWhiteList) error {
	if a == nil {
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
	}
	return nil
}

func (_ *LoadBalancerWhiteList) RenderALI(t *aliup.ALIAPITarget, a, e, changes *LoadBalancerWhiteList) error {

	klog.V(2).Infof("Updating LoadBalancerWhiteList of ListenerPort: %q", *e.LoadBalancerListener.ListenerPort)

	loadBalancerId := fi.StringValue(e.LoadBalancer.LoadbalancerId)
	listenertPort := fi.IntValue(e.LoadBalancerListener.ListenerPort)
	sourceItems := fi.StringValue(e.SourceItems)
	if sourceItems != "" {
		err := t.Cloud.SlbClient().AddListenerWhiteListItem(loadBalancerId, listenertPort, sourceItems)
		if err != nil {
			return fmt.Errorf("error adding LoadBalancerWhiteListItems: %v", err)
		}
	}

	if a != nil && changes.SourceItems != nil {
		itemsToDelete := e.getWhiteItemsToDelete(fi.StringValue(a.SourceItems))
		if itemsToDelete != "" {
			err := t.Cloud.SlbClient().RemoveListenerWhiteListItem(loadBalancerId, listenertPort, itemsToDelete)
			if err != nil {
				return fmt.Errorf("error removing LoadBalancerWhiteListItems: %v", err)
			}
		}
	}
	return nil
}

func (l *LoadBalancerWhiteList) getWhiteItemsToDelete(currentWhiteItems string) string {
	currentWhiteItemsList := strings.Split(currentWhiteItems, ",")
	expectedWhiteItemsList := strings.Split(fi.StringValue(l.SourceItems), ",")
	itemsToDelete := ""
	for _, currentItem := range currentWhiteItemsList {
		expected := false
		if currentItem == "" {
			continue
		}
		for _, expectedItem := range expectedWhiteItemsList {
			if currentItem == expectedItem {
				expected = true
			}
		}
		if !expected {
			itemsToDelete = itemsToDelete + "," + currentItem
		}
	}
	return itemsToDelete
}

func (_ *LoadBalancerWhiteList) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *LoadBalancerWhiteList) error {
	klog.Warningf("terraform does not support LoadBalancerWhiteList on ALI cloud")
	return nil
}
