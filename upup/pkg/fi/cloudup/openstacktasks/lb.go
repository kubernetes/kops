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

package openstacktasks

import (
	"fmt"
	"time"

	"github.com/gophercloud/gophercloud/openstack/networking/v2/ports"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/loadbalancer/v2/loadbalancers"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/subnets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
)

//go:generate fitask -type=LB
type LB struct {
	ID            *string
	Name          *string
	Subnet        *string
	VipSubnet     *string
	Lifecycle     *fi.Lifecycle
	PortID        *string
	SecurityGroup *SecurityGroup
}

const (
	// loadbalancerActive* is configuration of exponential backoff for
	// going into ACTIVE loadbalancer provisioning status. Starting with 1
	// seconds, multiplying by 1.2 with each step and taking 22 steps at maximum
	// it will time out after 326s, which roughly corresponds to about 5 minutes
	loadbalancerActiveInitDelay = 1 * time.Second
	loadbalancerActiveFactor    = 1.2
	loadbalancerActiveSteps     = 22

	activeStatus = "ACTIVE"
	errorStatus  = "ERROR"
)

func waitLoadbalancerActiveProvisioningStatus(client *gophercloud.ServiceClient, loadbalancerID string) (string, error) {
	backoff := wait.Backoff{
		Duration: loadbalancerActiveInitDelay,
		Factor:   loadbalancerActiveFactor,
		Steps:    loadbalancerActiveSteps,
	}

	var provisioningStatus string
	err := wait.ExponentialBackoff(backoff, func() (bool, error) {
		loadbalancer, err := loadbalancers.Get(client, loadbalancerID).Extract()
		if err != nil {
			return false, err
		}
		provisioningStatus = loadbalancer.ProvisioningStatus
		if loadbalancer.ProvisioningStatus == activeStatus {
			return true, nil
		} else if loadbalancer.ProvisioningStatus == errorStatus {
			return true, fmt.Errorf("loadbalancer has gone into ERROR state")
		} else {
			klog.Infof("Waiting for Loadbalancer to be ACTIVE...")
			return false, nil
		}

	})

	if err == wait.ErrWaitTimeout {
		err = fmt.Errorf("loadbalancer failed to go into ACTIVE provisioning status within allotted time")
	}
	return provisioningStatus, err
}

// GetDependencies returns the dependencies of the Instance task
func (e *LB) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	var deps []fi.Task
	for _, task := range tasks {
		if _, ok := task.(*Subnet); ok {
			deps = append(deps, task)
		}
		if _, ok := task.(*ServerGroup); ok {
			deps = append(deps, task)
		}
		if _, ok := task.(*Instance); ok {
			deps = append(deps, task)
		}
		if _, ok := task.(*SecurityGroup); ok {
			deps = append(deps, task)
		}
	}
	return deps
}

var _ fi.CompareWithID = &LB{}

func (s *LB) CompareWithID() *string {
	return s.ID
}

func NewLBTaskFromCloud(cloud openstack.OpenstackCloud, lifecycle *fi.Lifecycle, lb *loadbalancers.LoadBalancer, find *LB) (*LB, error) {
	osCloud := cloud.(openstack.OpenstackCloud)
	sub, err := subnets.Get(osCloud.NetworkingClient(), lb.VipSubnetID).Extract()
	if err != nil {
		return nil, err
	}

	sg, err := getSecurityGroupByName(&SecurityGroup{Name: fi.String(lb.Name)}, osCloud)
	if err != nil {
		return nil, err
	}

	actual := &LB{
		ID:            fi.String(lb.ID),
		Name:          fi.String(lb.Name),
		Lifecycle:     lifecycle,
		PortID:        fi.String(lb.VipPortID),
		Subnet:        fi.String(sub.Name),
		VipSubnet:     fi.String(lb.VipSubnetID),
		SecurityGroup: sg,
	}

	if find != nil {
		find.ID = actual.ID
		find.PortID = actual.PortID
		find.VipSubnet = actual.VipSubnet
	}
	return actual, nil
}

func (s *LB) Find(context *fi.Context) (*LB, error) {
	if s.Name == nil {
		return nil, nil
	}

	cloud := context.Cloud.(openstack.OpenstackCloud)
	lbPage, err := loadbalancers.List(cloud.LoadBalancerClient(), loadbalancers.ListOpts{
		Name: fi.StringValue(s.Name),
	}).AllPages()
	if err != nil {
		return nil, fmt.Errorf("Failed to retrieve loadbalancers for name %s: %v", fi.StringValue(s.Name), err)
	}
	lbs, err := loadbalancers.ExtractLoadBalancers(lbPage)
	if err != nil {
		return nil, fmt.Errorf("Failed to extract loadbalancers : %v", err)
	}
	if len(lbs) == 0 {
		return nil, nil
	}
	if len(lbs) > 1 {
		return nil, fmt.Errorf("Multiple load balancers for name %s", fi.StringValue(s.Name))
	}

	return NewLBTaskFromCloud(cloud, s.Lifecycle, &lbs[0], s)
}

func (s *LB) Run(context *fi.Context) error {
	return fi.DefaultDeltaRunMethod(s, context)
}

func (_ *LB) CheckChanges(a, e, changes *LB) error {
	if a == nil {
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
	} else {
		if changes.ID != nil {
			return fi.CannotChangeField("ID")
		}
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
	}
	return nil
}

func (_ *LB) RenderOpenstack(t *openstack.OpenstackAPITarget, a, e, changes *LB) error {
	if a == nil {
		klog.V(2).Infof("Creating LB with Name: %q", fi.StringValue(e.Name))

		subnets, err := t.Cloud.ListSubnets(subnets.ListOpts{
			Name: fi.StringValue(e.Subnet),
		})
		if err != nil {
			return fmt.Errorf("Failed to retrieve subnet `%s` in loadbalancer creation: %v", fi.StringValue(e.Subnet), err)
		}
		if len(subnets) != 1 {
			return fmt.Errorf("Unexpected desired subnets for `%s`.  Expected 1, got %d", fi.StringValue(e.Subnet), len(subnets))
		}

		lbopts := loadbalancers.CreateOpts{
			Name:        fi.StringValue(e.Name),
			VipSubnetID: subnets[0].ID,
		}
		lb, err := t.Cloud.CreateLB(lbopts)
		if err != nil {
			return fmt.Errorf("error creating LB: %v", err)
		}
		e.ID = fi.String(lb.ID)
		e.PortID = fi.String(lb.VipPortID)
		e.VipSubnet = fi.String(lb.VipSubnetID)

		opts := ports.UpdateOpts{
			SecurityGroups: &[]string{fi.StringValue(e.SecurityGroup.ID)},
		}
		_, err = ports.Update(t.Cloud.NetworkingClient(), lb.VipPortID, opts).Extract()
		if err != nil {
			return fmt.Errorf("Failed to update security group for port %s: %v", lb.VipPortID, err)
		}
		return nil
	}
	// We may have failed to update the security groups on the load balancer
	port, err := t.Cloud.GetPort(fi.StringValue(a.PortID))
	if err != nil {
		return fmt.Errorf("Failed to get port with id %s: %v", fi.StringValue(a.PortID), err)
	}
	// Ensure the loadbalancer port has one security group and it is the one specified,
	if e.SecurityGroup != nil &&
		(len(port.SecurityGroups) < 1 || port.SecurityGroups[0] != fi.StringValue(e.SecurityGroup.ID)) {

		opts := ports.UpdateOpts{
			SecurityGroups: &[]string{fi.StringValue(e.SecurityGroup.ID)},
		}
		_, err = ports.Update(t.Cloud.NetworkingClient(), fi.StringValue(a.PortID), opts).Extract()
		if err != nil {
			return fmt.Errorf("Failed to update security group for port %s: %v", fi.StringValue(a.PortID), err)
		}
		return nil
	}

	klog.V(2).Infof("Openstack task LB::RenderOpenstack did nothing")
	return nil
}
