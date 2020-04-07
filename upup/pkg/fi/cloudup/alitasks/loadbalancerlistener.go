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
	"strconv"

	"k8s.io/klog"

	"github.com/denverdino/aliyungo/slb"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/aliup"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

const ListenerRunningStatus = "running"

//go:generate fitask -type=LoadBalancerListener
type LoadBalancerListener struct {
	LoadBalancer      *LoadBalancer
	Name              *string
	ListenerPort      *int
	BackendServerPort *int
	Lifecycle         *fi.Lifecycle
	ListenerStatus    *string
	Bandwidth         *int
}

var _ fi.CompareWithID = &LoadBalancerListener{}

func (l *LoadBalancerListener) CompareWithID() *string {
	listenerPort := strconv.Itoa(fi.IntValue(l.ListenerPort))
	return fi.String(listenerPort)
}

func (l *LoadBalancerListener) Find(c *fi.Context) (*LoadBalancerListener, error) {
	if l.LoadBalancer == nil || l.LoadBalancer.LoadbalancerId == nil {
		klog.V(4).Infof("LoadBalancer / LoadbalancerId not found for %s, skipping Find", fi.StringValue(l.Name))
		return nil, nil
	}
	cloud := c.Cloud.(aliup.ALICloud)
	loadBalancerId := fi.StringValue(l.LoadBalancer.LoadbalancerId)
	listenerPort := fi.IntValue(l.ListenerPort)

	//TODO: should sort errors?
	response, err := cloud.SlbClient().DescribeLoadBalancerTCPListenerAttribute(loadBalancerId, listenerPort)
	if err != nil {
		return nil, nil
	}

	klog.V(2).Infof("found matching LoadBalancerListener with ListenerPort: %v", listenerPort)

	actual := &LoadBalancerListener{}
	actual.BackendServerPort = fi.Int(response.BackendServerPort)
	actual.ListenerPort = fi.Int(response.ListenerPort)
	actual.ListenerStatus = fi.String(string(response.Status))
	actual.Bandwidth = fi.Int(response.Bandwidth)
	// Ignore "system" fields
	actual.LoadBalancer = l.LoadBalancer
	actual.Lifecycle = l.Lifecycle
	actual.Name = l.Name
	return actual, nil
}

func (l *LoadBalancerListener) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(l, c)
}

func (_ *LoadBalancerListener) CheckChanges(a, e, changes *LoadBalancerListener) error {
	if a == nil {
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
		if e.ListenerPort == nil {
			return fi.RequiredField("ListenerPort")
		}
		if e.BackendServerPort == nil {
			return fi.RequiredField("BackendServerPort")
		}
	} else {
		if changes.BackendServerPort != nil {
			return fi.CannotChangeField("BackendServerPort")
		}
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
	}
	return nil
}

//LoadBalancer can only modify tags.
func (_ *LoadBalancerListener) RenderALI(t *aliup.ALIAPITarget, a, e, changes *LoadBalancerListener) error {
	loadBalancerId := fi.StringValue(e.LoadBalancer.LoadbalancerId)
	listenerPort := fi.IntValue(e.ListenerPort)
	if a == nil {
		klog.V(2).Infof("Creating LoadBalancerListener with ListenerPort: %q", listenerPort)

		createLoadBalancerTCPListenerArgs := &slb.CreateLoadBalancerTCPListenerArgs{
			LoadBalancerId:    loadBalancerId,
			ListenerPort:      listenerPort,
			BackendServerPort: fi.IntValue(e.BackendServerPort),
			Bandwidth:         fi.IntValue(e.Bandwidth),
		}
		err := t.Cloud.SlbClient().CreateLoadBalancerTCPListener(createLoadBalancerTCPListenerArgs)
		if err != nil {
			return fmt.Errorf("error creating LoadBalancerlistener: %v", err)
		}
	}

	if fi.StringValue(e.ListenerStatus) == ListenerRunningStatus {
		klog.V(2).Infof("Starting LoadBalancerListener with ListenerPort: %q", listenerPort)

		err := t.Cloud.SlbClient().StartLoadBalancerListener(loadBalancerId, listenerPort)
		if err != nil {
			return fmt.Errorf("error starting LoadBalancerListener: %v", err)
		}
	} else {
		klog.V(2).Infof("Stopping  LoadBalancerListener with ListenerPort: %q", listenerPort)

		err := t.Cloud.SlbClient().StopLoadBalancerListener(loadBalancerId, listenerPort)
		if err != nil {
			return fmt.Errorf("error stopping LoadBalancerListener: %v", err)
		}
	}

	klog.V(2).Infof("Waiting LoadBalancerListener with ListenerPort: %q", listenerPort)

	_, err := t.Cloud.SlbClient().WaitForListener(loadBalancerId, listenerPort, slb.TCP)
	if err != nil {
		return fmt.Errorf("error waitting LoadBalancerListener: %v", err)
	}

	return nil
}

type terraformLoadBalancerListener struct {
	ListenerPort      *int               `json:"frontend_port,omitempty" cty:"frontend_port"`
	BackendServerPort *int               `json:"backend_port,omitempty" cty:"backend_port"`
	Protocol          *string            `json:"protocol,omitempty" cty:"protocol"`
	LoadBalancerId    *terraform.Literal `json:"load_balancer_id,omitempty" cty:"load_balancer_id"`
}

func (_ *LoadBalancerListener) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *LoadBalancerListener) error {
	protocol := "tcp"
	tf := &terraformLoadBalancerListener{
		ListenerPort:      e.ListenerPort,
		BackendServerPort: e.BackendServerPort,
		Protocol:          &protocol,
		LoadBalancerId:    e.LoadBalancer.TerraformLink(),
	}

	return t.RenderResource("alicloud_slb_listener", *e.Name, tf)
}

func (s *LoadBalancerListener) TerraformLink() *terraform.Literal {
	return terraform.LiteralProperty("alicloud_slb_listener", *s.Name, "frontend_port")
}
