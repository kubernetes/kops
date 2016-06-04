package kutil

import (
	"encoding/base64"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/golang/glog"
	"k8s.io/kube-deploy/upup/pkg/fi"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/awsup"
	"strings"
)

// DeleteCluster implements deletion of cluster cloud resources
// The algorithm is pretty simple: it discovers all the resources it can (primary using tags),
// and then it repeatedly attempts to delete them all until they are all deleted.
// There are a few tweaks to that approach, like choosing a default ordering, but it is not much
// smarter.  Cluster deletion is a fairly rare operation anyway, and also some dependencies are invisible
// (e.g. ELB dependencies).
type DeleteCluster struct {
	ClusterID string
	Region    string
	Cloud     fi.Cloud
}

func (c *DeleteCluster) ListResources() ([]DeletableResource, error) {
	cloud := c.Cloud.(*awsup.AWSCloud)

	var resources []DeletableResource

	filters := cloud.BuildFilters(nil)
	tags := cloud.BuildTags(nil, nil)

	{
		glog.V(2).Infof("Listing all Autoscaling groups matching cluster tags")
		var asgNames []*string
		{
			var asFilters []*autoscaling.Filter
			for _, f := range filters {
				asFilters = append(asFilters, &autoscaling.Filter{
					Name:   aws.String("value"),
					Values: f.Values,
				})
			}
			request := &autoscaling.DescribeTagsInput{
				Filters: asFilters,
			}
			response, err := cloud.Autoscaling.DescribeTags(request)
			if err != nil {
				return nil, fmt.Errorf("error listing autoscaling cluster tags: %v", err)
			}

			for _, t := range response.Tags {
				switch *t.ResourceType {
				case "auto-scaling-group":
					asgNames = append(asgNames, t.ResourceId)
				default:
					glog.Warningf("Unknown resource type: %v", *t.ResourceType)

				}
			}
		}

		if len(asgNames) != 0 {
			request := &autoscaling.DescribeAutoScalingGroupsInput{
				AutoScalingGroupNames: asgNames,
			}
			response, err := cloud.Autoscaling.DescribeAutoScalingGroups(request)
			if err != nil {
				return nil, fmt.Errorf("error listing autoscaling groups: %v", err)
			}

			for _, t := range response.AutoScalingGroups {
				if !matchesAsgTags(tags, t.Tags) {
					continue
				}
				resources = append(resources, &DeletableASG{Name: *t.AutoScalingGroupName})
			}
		}
	}

	{
		glog.V(2).Infof("Listing all Autoscaling LaunchConfigurations")

		request := &autoscaling.DescribeLaunchConfigurationsInput{}
		response, err := cloud.Autoscaling.DescribeLaunchConfigurations(request)
		if err != nil {
			return nil, fmt.Errorf("error listing autoscaling LaunchConfigurations: %v", err)
		}

		for _, t := range response.LaunchConfigurations {
			if t.UserData == nil {
				continue
			}

			userData, err := base64.StdEncoding.DecodeString(*t.UserData)
			if err != nil {
				glog.Infof("Ignoring autoscaling LaunchConfiguration with invalid UserData: %v", *t.LaunchConfigurationName)
				continue
			}

			if strings.Contains(string(userData), "\nINSTANCE_PREFIX: "+c.ClusterID+"\n") {
				resources = append(resources, &DeletableAutoscalingLaunchConfiguration{Name: *t.LaunchConfigurationName})
			}
		}
	}

	{
		glog.V(2).Infof("Listing all ELB tags")

		request := &elb.DescribeLoadBalancersInput{}
		response, err := cloud.ELB.DescribeLoadBalancers(request)
		if err != nil {
			return nil, fmt.Errorf("error listing elb LoadBalancers: %v", err)
		}

		for _, lb := range response.LoadBalancerDescriptions {
			// TODO: batch?
			request := &elb.DescribeTagsInput{
				LoadBalancerNames: []*string{lb.LoadBalancerName},
			}
			response, err := cloud.ELB.DescribeTags(request)
			if err != nil {
				return nil, fmt.Errorf("error listing elb Tags: %v", err)
			}

			for _, t := range response.TagDescriptions {
				if !matchesElbTags(tags, t.Tags) {
					continue
				}
				resources = append(resources, &DeletableELBLoadBalancer{Name: *t.LoadBalancerName})
			}
		}
	}

	{

		glog.V(2).Infof("Listing all EC2 tags matching cluster tags")
		request := &ec2.DescribeTagsInput{
			Filters: filters,
		}
		response, err := cloud.EC2.DescribeTags(request)
		if err != nil {
			return nil, fmt.Errorf("error listing cluster tags: %v", err)
		}

		for _, t := range response.Tags {
			var resource DeletableResource
			switch *t.ResourceType {
			case "dhcp-options":
				resource = &DeletableDHCPOptions{ID: *t.ResourceId}
			case "instance":
				resource = &DeletableInstance{ID: *t.ResourceId}
			case "volume":
				resource = &DeletableVolume{ID: *t.ResourceId}
			case "subnet":
				resource = &DeletableSubnet{ID: *t.ResourceId}
			case "security-group":
				resource = &DeletableSecurityGroup{ID: *t.ResourceId}
			case "internet-gateway":
				resource = &DeletableInternetGateway{ID: *t.ResourceId}
			case "route-table":
				resource = &DeletableRouteTable{ID: *t.ResourceId}
			case "vpc":
				resource = &DeletableVPC{ID: *t.ResourceId}
			}

			if resource == nil {
				glog.Warningf("Unknown resource type: %v", *t.ResourceType)
				continue
			}

			resources = append(resources, resource)
		}
	}

	return resources, nil
}

func matchesAsgTags(tags map[string]string, actual []*autoscaling.TagDescription) bool {
	for k, v := range tags {
		found := false
		for _, a := range actual {
			if aws.StringValue(a.Key) == k {
				if aws.StringValue(a.Value) == v {
					found = true
					break
				}
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func matchesElbTags(tags map[string]string, actual []*elb.Tag) bool {
	for k, v := range tags {
		found := false
		for _, a := range actual {
			if aws.StringValue(a.Key) == k {
				if aws.StringValue(a.Value) == v {
					found = true
					break
				}
			}
		}
		if !found {
			return false
		}
	}
	return true
}

type DeletableResource interface {
	Delete(cloud fi.Cloud) error
}

type DeletableInstance struct {
	ID string
}

func (r *DeletableInstance) Delete(cloud fi.Cloud) error {
	c := cloud.(*awsup.AWSCloud)

	glog.V(2).Infof("Deleting EC2 instance %q", r.ID)
	request := &ec2.TerminateInstancesInput{
		InstanceIds: []*string{&r.ID},
	}
	_, err := c.EC2.TerminateInstances(request)
	if err != nil {
		return fmt.Errorf("error deleting instance %q: %v", r.ID, err)
	}
	return nil
}
func (r *DeletableInstance) String() string {
	return "Instance:" + r.ID
}

type DeletableSecurityGroup struct {
	ID string
}

func (r *DeletableSecurityGroup) Delete(cloud fi.Cloud) error {
	c := cloud.(*awsup.AWSCloud)

	// First clear all inter-dependent rules
	// TODO: Move to a "pre-execute" phase?
	{
		request := &ec2.DescribeSecurityGroupsInput{
			GroupIds: []*string{&r.ID},
		}
		response, err := c.EC2.DescribeSecurityGroups(request)
		if err != nil {
			return fmt.Errorf("error describing SecurityGroup %q: %v", r.ID, err)
		}

		if len(response.SecurityGroups) == 0 {
			return nil
		}
		if len(response.SecurityGroups) != 1 {
			return fmt.Errorf("found mutiple SecurityGroups with ID %q", r.ID)
		}
		sg := response.SecurityGroups[0]

		if len(sg.IpPermissions) != 0 {
			revoke := &ec2.RevokeSecurityGroupIngressInput{
				GroupId:       &r.ID,
				IpPermissions: sg.IpPermissions,
			}
			_, err = c.EC2.RevokeSecurityGroupIngress(revoke)
			if err != nil {
				return fmt.Errorf("cannot revoke ingress for ID %q: %v", r.ID, err)
			}
		}
	}

	{
		glog.V(2).Infof("Deleting EC2 SecurityGroup %q", r.ID)
		request := &ec2.DeleteSecurityGroupInput{
			GroupId: &r.ID,
		}
		_, err := c.EC2.DeleteSecurityGroup(request)
		if err != nil {
			if IsDependencyViolation(err) {
				return err
			}
			return fmt.Errorf("error deleting SecurityGroup %q: %v", r.ID, err)
		}
	}
	return nil
}
func (r *DeletableSecurityGroup) String() string {
	return "SecurityGroup:" + r.ID
}

type DeletableVolume struct {
	ID string
}

func (r *DeletableVolume) Delete(cloud fi.Cloud) error {
	c := cloud.(*awsup.AWSCloud)

	glog.V(2).Infof("Deleting EC2 volume %q", r.ID)
	request := &ec2.DeleteVolumeInput{
		VolumeId: &r.ID,
	}
	_, err := c.EC2.DeleteVolume(request)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "InvalidVolume.NotFound" {
				return nil
			}
		}
		return fmt.Errorf("error deleting volume %q: %v", r.ID, err)
	}
	return nil
}
func (r *DeletableVolume) String() string {
	return "Volume:" + r.ID
}

type DeletableSubnet struct {
	ID string
}

func IsDependencyViolation(err error) bool {
	if awsError, ok := err.(awserr.Error); ok {
		if awsError.Code() == "DependencyViolation" {
			return true
		}
	}
	return false
}

func (r *DeletableSubnet) Delete(cloud fi.Cloud) error {
	c := cloud.(*awsup.AWSCloud)

	glog.V(2).Infof("Deleting EC2 Subnet %q", r.ID)
	request := &ec2.DeleteSubnetInput{
		SubnetId: &r.ID,
	}
	_, err := c.EC2.DeleteSubnet(request)
	if err != nil {
		if IsDependencyViolation(err) {
			return err
		}
		return fmt.Errorf("error deleting Subnet %q: %v", r.ID, err)
	}
	return nil
}
func (r *DeletableSubnet) String() string {
	return "Subnet:" + r.ID
}

type DeletableRouteTable struct {
	ID string
}

func (r *DeletableRouteTable) Delete(cloud fi.Cloud) error {
	c := cloud.(*awsup.AWSCloud)

	glog.V(2).Infof("Deleting EC2 RouteTable %q", r.ID)
	request := &ec2.DeleteRouteTableInput{
		RouteTableId: &r.ID,
	}
	_, err := c.EC2.DeleteRouteTable(request)
	if err != nil {
		if IsDependencyViolation(err) {
			return err
		}
		return fmt.Errorf("error deleting RouteTable %q: %v", r.ID, err)
	}
	return nil
}
func (r *DeletableRouteTable) String() string {
	return "RouteTable:" + r.ID
}

type DeletableDHCPOptions struct {
	ID string
}

func (r *DeletableDHCPOptions) Delete(cloud fi.Cloud) error {
	c := cloud.(*awsup.AWSCloud)

	glog.V(2).Infof("Deleting EC2 DHCPOptions %q", r.ID)
	request := &ec2.DeleteDhcpOptionsInput{
		DhcpOptionsId: &r.ID,
	}
	_, err := c.EC2.DeleteDhcpOptions(request)
	if err != nil {
		if IsDependencyViolation(err) {
			return err
		}
		return fmt.Errorf("error deleting %q: %v", r.ID, err)
	}
	return nil
}
func (r *DeletableDHCPOptions) String() string {
	return "DHCPOptions:" + r.ID
}

type DeletableInternetGateway struct {
	ID string
}

func (r *DeletableInternetGateway) Delete(cloud fi.Cloud) error {
	c := cloud.(*awsup.AWSCloud)

	var igw *ec2.InternetGateway
	{
		request := &ec2.DescribeInternetGatewaysInput{
			InternetGatewayIds: []*string{&r.ID},
		}
		response, err := c.EC2.DescribeInternetGateways(request)
		if err != nil {
			return fmt.Errorf("error describing InternetGateway %q: %v", r.ID, err)
		}
		if response == nil || len(response.InternetGateways) == 0 {
			return nil
		}
		if len(response.InternetGateways) != 1 {
			return fmt.Errorf("found multiple InternetGateways with id %q", r.ID)
		}
		igw = response.InternetGateways[0]
	}

	for _, a := range igw.Attachments {
		glog.V(2).Infof("Detaching EC2 InternetGateway %q", r.ID)
		request := &ec2.DetachInternetGatewayInput{
			InternetGatewayId: &r.ID,
			VpcId:             a.VpcId,
		}
		_, err := c.EC2.DetachInternetGateway(request)
		if err != nil {
			return fmt.Errorf("error detaching InternetGateway %q: %v", r.ID, err)
		}
	}

	{
		glog.V(2).Infof("Deleting EC2 InternetGateway %q", r.ID)
		request := &ec2.DeleteInternetGatewayInput{
			InternetGatewayId: &r.ID,
		}
		_, err := c.EC2.DeleteInternetGateway(request)
		if err != nil {
			if IsDependencyViolation(err) {
				return err
			}
			return fmt.Errorf("error deleting InternetGateway %q: %v", r.ID, err)
		}
	}

	return nil
}
func (r *DeletableInternetGateway) String() string {
	return "InternetGateway:" + r.ID
}

type DeletableVPC struct {
	ID string
}

func (r *DeletableVPC) Delete(cloud fi.Cloud) error {
	c := cloud.(*awsup.AWSCloud)

	glog.V(2).Infof("Deleting EC2 VPC %q", r.ID)
	request := &ec2.DeleteVpcInput{
		VpcId: &r.ID,
	}
	_, err := c.EC2.DeleteVpc(request)
	if err != nil {
		if IsDependencyViolation(err) {
			return err
		}
		return fmt.Errorf("error deleting VPC %q: %v", r.ID, err)
	}
	return nil
}
func (r *DeletableVPC) String() string {
	return "VPC:" + r.ID
}

type DeletableASG struct {
	Name string
}

func (r *DeletableASG) Delete(cloud fi.Cloud) error {
	c := cloud.(*awsup.AWSCloud)

	glog.V(2).Infof("Deleting autoscaling group %q", r.Name)
	request := &autoscaling.DeleteAutoScalingGroupInput{
		AutoScalingGroupName: &r.Name,
		ForceDelete:          aws.Bool(true),
	}
	_, err := c.Autoscaling.DeleteAutoScalingGroup(request)
	if err != nil {
		if IsDependencyViolation(err) {
			return err
		}
		return fmt.Errorf("error deleting autoscaling group %q: %v", r.Name, err)
	}
	return nil
}
func (r *DeletableASG) String() string {
	return "autoscaling-group:" + r.Name
}

type DeletableAutoscalingLaunchConfiguration struct {
	Name string
}

func (r *DeletableAutoscalingLaunchConfiguration) Delete(cloud fi.Cloud) error {
	c := cloud.(*awsup.AWSCloud)

	glog.V(2).Infof("Deleting autoscaling LaunchConfiguration %q", r.Name)
	request := &autoscaling.DeleteLaunchConfigurationInput{
		LaunchConfigurationName: &r.Name,
	}
	_, err := c.Autoscaling.DeleteLaunchConfiguration(request)
	if err != nil {
		return fmt.Errorf("error deleting autoscaling LaunchConfiguration %q: %v", r.Name, err)
	}
	return nil
}

func (r *DeletableAutoscalingLaunchConfiguration) String() string {
	return "autoscaling-launchconfiguration:" + r.Name
}

type DeletableELBLoadBalancer struct {
	Name string
}

func (r *DeletableELBLoadBalancer) Delete(cloud fi.Cloud) error {
	c := cloud.(*awsup.AWSCloud)

	glog.V(2).Infof("Deleting LoadBalancer %q", r.Name)
	request := &elb.DeleteLoadBalancerInput{
		LoadBalancerName: &r.Name,
	}
	_, err := c.ELB.DeleteLoadBalancer(request)
	if err != nil {
		if IsDependencyViolation(err) {
			return err
		}
		return fmt.Errorf("error deleting LoadBalancer %q: %v", r.Name, err)
	}
	return nil
}

func (r *DeletableELBLoadBalancer) String() string {
	return "LoadBalancer:" + r.Name
}
