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

package awstasks

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	elb "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/types"
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
)

// ClassicLoadBalancer represents an AWS Classic Load Balancer (ELB).
//
// kOps no longer creates Classic Load Balancers; this task only supports
// attaching instance groups to external load balancers managed outside of kOps,
// via the instance group's spec.externalLoadBalancers[].loadBalancerName.

// +kops:fitask
type ClassicLoadBalancer struct {
	Name      *string
	Lifecycle fi.Lifecycle

	// LoadBalancerName is the name in ELB, possibly different from our name
	// (ELB is restricted as to names, so we have limited choices!)
	LoadBalancerName *string

	// Shared is set if this is an external LB (one we don't create or own)
	Shared *bool
}

var _ fi.CompareWithID = (*ClassicLoadBalancer)(nil)

func (e *ClassicLoadBalancer) CompareWithID() *string {
	return e.Name
}

func findLoadBalancerByLoadBalancerName(ctx context.Context, cloud awsup.AWSCloud, loadBalancerName string) (*elbtypes.LoadBalancerDescription, error) {
	request := &elb.DescribeLoadBalancersInput{
		LoadBalancerNames: []string{loadBalancerName},
	}
	found, err := describeLoadBalancers(ctx, cloud, request, func(lb elbtypes.LoadBalancerDescription) bool {
		if aws.ToString(lb.LoadBalancerName) == loadBalancerName {
			return true
		}

		klog.Warningf("Got ELB with unexpected name: %q", aws.ToString(lb.LoadBalancerName))
		return false
	})
	if err != nil {
		if awsup.AWSErrorCode(err) == "LoadBalancerNotFound" {
			return nil, nil
		}

		return nil, fmt.Errorf("error listing ELBs: %v", err)
	}

	if len(found) == 0 {
		return nil, nil
	}

	if len(found) != 1 {
		return nil, fmt.Errorf("Found multiple ELBs with name %q", loadBalancerName)
	}

	return &found[0], nil
}

func describeLoadBalancers(ctx context.Context, cloud awsup.AWSCloud, request *elb.DescribeLoadBalancersInput, filter func(elbtypes.LoadBalancerDescription) bool) ([]elbtypes.LoadBalancerDescription, error) {
	var found []elbtypes.LoadBalancerDescription
	paginator := elb.NewDescribeLoadBalancersPaginator(cloud.ELB(), request)
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("error listing ELBs: %w", err)
		}

		for _, lb := range output.LoadBalancerDescriptions {
			if filter(lb) {
				found = append(found, lb)
			}
		}
	}
	return found, nil
}

func (e *ClassicLoadBalancer) Find(c *fi.CloudupContext) (*ClassicLoadBalancer, error) {
	ctx := c.Context()
	cloud := awsup.GetCloud(c)

	if e.LoadBalancerName == nil {
		return nil, nil
	}

	lb, err := findLoadBalancerByLoadBalancerName(ctx, cloud, fi.ValueOf(e.LoadBalancerName))
	if err != nil {
		return nil, err
	}
	if lb == nil {
		return nil, nil
	}

	actual := &ClassicLoadBalancer{}
	actual.Name = e.Name
	actual.LoadBalancerName = lb.LoadBalancerName

	// Ignore system fields
	actual.Lifecycle = e.Lifecycle
	actual.Shared = e.Shared

	return actual, nil
}

func (e *ClassicLoadBalancer) Run(c *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(e, c)
}

func (_ *ClassicLoadBalancer) ShouldCreate(a, e, changes *ClassicLoadBalancer) (bool, error) {
	if fi.ValueOf(e.Shared) {
		return false, nil
	}
	return true, nil
}

func (s *ClassicLoadBalancer) CheckChanges(a, e, changes *ClassicLoadBalancer) error {
	if a == nil {
		if fi.ValueOf(e.Name) == "" {
			return fi.RequiredField("Name")
		}
		if !fi.ValueOf(e.Shared) {
			return fmt.Errorf("creation of AWS Classic Load Balancers is no longer supported")
		}
	}

	return nil
}

func (_ *ClassicLoadBalancer) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *ClassicLoadBalancer) error {
	if fi.ValueOf(e.Shared) {
		return nil
	}

	return fmt.Errorf("creation of AWS Classic Load Balancers is no longer supported")
}

// OrderLoadBalancersByName implements sort.Interface for []OrderLoadBalancersByName, based on name
type OrderLoadBalancersByName []*ClassicLoadBalancer

func (a OrderLoadBalancersByName) Len() int      { return len(a) }
func (a OrderLoadBalancersByName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a OrderLoadBalancersByName) Less(i, j int) bool {
	return fi.ValueOf(a[i].Name) < fi.ValueOf(a[j].Name)
}

func (_ *ClassicLoadBalancer) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *ClassicLoadBalancer) error {
	if fi.ValueOf(e.Shared) {
		return nil
	}

	return fmt.Errorf("creation of AWS Classic Load Balancers is no longer supported")
}

func (e *ClassicLoadBalancer) TerraformLink(params ...string) *terraformWriter.Literal {
	shared := fi.ValueOf(e.Shared)
	if shared {
		if e.LoadBalancerName == nil {
			klog.Fatalf("Name must be set, if LB is shared: %s", e)
		}

		klog.V(4).Infof("reusing existing LB with name %q", *e.LoadBalancerName)
		return terraformWriter.LiteralFromStringValue(*e.LoadBalancerName)
	}

	prop := "id"
	if len(params) > 0 {
		prop = params[0]
	}
	return terraformWriter.LiteralProperty("aws_elb", *e.Name, prop)
}
