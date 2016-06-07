package awstasks

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/golang/glog"
	"k8s.io/kube-deploy/upup/pkg/fi"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/awsup"
)

type LoadBalancerHealthChecks struct {
	LoadBalancer *LoadBalancer

	Target *string

	HealthyThreshold   *int64
	UnhealthyThreshold *int64

	Interval *int64
	Timeout  *int64
}

func (e *LoadBalancerHealthChecks) String() string {
	return fi.TaskAsString(e)
}

func (e *LoadBalancerHealthChecks) Find(c *fi.Context) (*LoadBalancerHealthChecks, error) {
	cloud := c.Cloud.(*awsup.AWSCloud)

	elbName := fi.StringValue(e.LoadBalancer.ID)

	lb, err := findELB(cloud, elbName)
	if err != nil {
		return nil, err
	}
	if lb == nil {
		return nil, nil
	}

	actual := &LoadBalancerHealthChecks{}
	actual.LoadBalancer = e.LoadBalancer

	if lb.HealthCheck != nil {
		actual.Target = lb.HealthCheck.Target
		actual.HealthyThreshold = lb.HealthCheck.HealthyThreshold
		actual.UnhealthyThreshold = lb.HealthCheck.UnhealthyThreshold
		actual.Interval = lb.HealthCheck.Interval
		actual.Timeout = lb.HealthCheck.Timeout
	}
	return actual, nil

}

func (e *LoadBalancerHealthChecks) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *LoadBalancerHealthChecks) CheckChanges(a, e, changes *LoadBalancerHealthChecks) error {
	if a == nil {
		if e.LoadBalancer == nil {
			return fi.RequiredField("LoadBalancer")
		}
		if e.Target == nil {
			return fi.RequiredField("Target")
		}
	}
	return nil
}

func (_ *LoadBalancerHealthChecks) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *LoadBalancerHealthChecks) error {
	request := &elb.ConfigureHealthCheckInput{}
	request.LoadBalancerName = e.LoadBalancer.ID
	request.HealthCheck = &elb.HealthCheck{
		Target:             e.Target,
		HealthyThreshold:   e.HealthyThreshold,
		UnhealthyThreshold: e.UnhealthyThreshold,
		Interval:           e.Interval,
		Timeout:            e.Timeout,
	}

	glog.V(2).Infof("Configuring health checks on ELB %q", *e.LoadBalancer.ID)

	_, err := t.Cloud.ELB.ConfigureHealthCheck(request)
	if err != nil {
		return fmt.Errorf("error attaching autoscaling group to ELB: %v", err)
	}

	return nil
}
