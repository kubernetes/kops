/*
Copyright 2016 The Kubernetes Authors.

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

package resources

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kops/pkg/dns"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

const (
	TypeAutoscalingLaunchConfig = "autoscaling-config"
	TypeNatGateway              = "nat-gateway"
	TypeElasticIp               = "elastic-ip"
	TypeLoadBalancer            = "load-balancer"
)

type listFn func(fi.Cloud, string) ([]*ResourceTracker, error)

func (c *ClusterResources) listResourcesAWS() (map[string]*ResourceTracker, error) {
	cloud := c.Cloud.(awsup.AWSCloud)

	resources := make(map[string]*ResourceTracker)

	// These are the functions that are used for looking up
	// cluster resources by their tags.
	listFunctions := []listFn{

		// CloudFormation
		//ListCloudFormationStacks,

		// EC2
		ListInstances,
		ListKeypairs,
		ListSecurityGroups,
		ListVolumes,
		// EC2 VPC
		ListDhcpOptions,
		ListInternetGateways,
		ListRouteTables,
		ListSubnets,
		ListVPCs,
		// ELBs
		ListELBs,
		// ASG
		ListAutoScalingGroups,

		// Route 53
		ListRoute53Records,
		// IAM
		ListIAMInstanceProfiles,
		ListIAMRoles,
	}
	for _, fn := range listFunctions {
		trackers, err := fn(cloud, c.ClusterName)
		if err != nil {
			return nil, err
		}
		for _, t := range trackers {
			resources[t.Type+":"+t.ID] = t
		}
	}

	{
		// Gateways weren't tagged in kube-up
		// If we are deleting the VPC, we should delete the attached gateway
		// (no real reason not to; easy to recreate; no real state etc)

		gateways, err := DescribeInternetGatewaysIgnoreTags(cloud)
		if err != nil {
			return nil, err
		}

		for _, igw := range gateways {
			for _, attachment := range igw.Attachments {
				vpcID := aws.StringValue(attachment.VpcId)
				igwID := aws.StringValue(igw.InternetGatewayId)
				if vpcID == "" || igwID == "" {
					continue
				}
				vpc := resources["vpc:"+vpcID]
				if vpc != nil && resources["internet-gateway:"+igwID] == nil {
					resources["internet-gateway:"+igwID] = &ResourceTracker{
						Name:    FindName(igw.Tags),
						ID:      igwID,
						Type:    "internet-gateway",
						deleter: DeleteInternetGateway,
						Shared:  vpc.Shared, // Shared iff the VPC is shared
					}
				}
			}
		}
	}

	{
		// We delete a launch configuration if it is bound to one of the tagged security groups
		securityGroups := sets.NewString()
		for k := range resources {
			if !strings.HasPrefix(k, "security-group:") {
				continue
			}
			id := strings.TrimPrefix(k, "security-group:")
			securityGroups.Insert(id)
		}
		lcs, err := FindAutoScalingLaunchConfigurations(cloud, securityGroups)
		if err != nil {
			return nil, err
		}

		for _, t := range lcs {
			resources[t.Type+":"+t.ID] = t
		}
	}

	if err := addUntaggedRouteTables(cloud, c.ClusterName, resources); err != nil {
		return nil, err
	}

	{
		// We delete a NAT gateway if it is linked to our route table
		routeTableIds := make(map[string]*ResourceTracker)
		for _, resource := range resources {
			if resource.Type != ec2.ResourceTypeRouteTable {
				continue
			}
			id := resource.ID
			routeTableIds[id] = resource
		}
		natGateways, err := FindNatGateways(cloud, routeTableIds)
		if err != nil {
			return nil, err
		}

		for _, t := range natGateways {
			resources[t.Type+":"+t.ID] = t
		}
	}

	for k, t := range resources {
		if t.done {
			delete(resources, k)
		}
	}
	return resources, nil
}

func BuildEC2Filters(cloud fi.Cloud) []*ec2.Filter {
	awsCloud := cloud.(awsup.AWSCloud)
	tags := awsCloud.Tags()

	var filters []*ec2.Filter
	for k, v := range tags {
		filter := awsup.NewEC2Filter("tag:"+k, v)
		filters = append(filters, filter)
	}
	return filters
}

func addUntaggedRouteTables(cloud awsup.AWSCloud, clusterName string, resources map[string]*ResourceTracker) error {
	// We sometimes have trouble tagging the route table (eventual consistency, e.g. #597)
	// If we are deleting the VPC, we should delete the route table
	// (no real reason not to; easy to recreate; no real state etc)
	routeTables, err := DescribeRouteTablesIgnoreTags(cloud)
	if err != nil {
		return err
	}

	for _, rt := range routeTables {
		rtID := aws.StringValue(rt.RouteTableId)
		vpcID := aws.StringValue(rt.VpcId)
		if vpcID == "" || rtID == "" {
			continue
		}

		if resources["vpc:"+vpcID] == nil {
			// Not deleting this VPC; ignore
			continue
		}

		clusterTag, _ := awsup.FindEC2Tag(rt.Tags, awsup.TagClusterName)
		if clusterTag != "" && clusterTag != clusterName {
			glog.Infof("Skipping route table in VPC, but with wrong cluster tag (%q)", clusterTag)
			continue
		}

		isMain := false
		for _, a := range rt.Associations {
			if aws.BoolValue(a.Main) == true {
				isMain = true
			}
		}
		if isMain {
			glog.V(4).Infof("ignoring main routetable %q", rtID)
			continue
		}

		t := buildTrackerForRouteTable(rt)
		if resources[t.Type+":"+t.ID] == nil {
			resources[t.Type+":"+t.ID] = t
		}
	}

	return nil
}

// FindAutoscalingGroups finds autoscaling groups matching the specified tags
// This isn't entirely trivial because autoscaling doesn't let us filter with as much precision as we would like
func FindAutoscalingGroups(cloud awsup.AWSCloud, tags map[string]string) ([]*autoscaling.Group, error) {
	var asgs []*autoscaling.Group

	glog.V(2).Infof("Listing all Autoscaling groups matching cluster tags")
	var asgNames []*string
	{
		var asFilters []*autoscaling.Filter
		for _, v := range tags {
			// Not an exact match, but likely the best we can do
			asFilters = append(asFilters, &autoscaling.Filter{
				Name:   aws.String("value"),
				Values: []*string{aws.String(v)},
			})
		}
		request := &autoscaling.DescribeTagsInput{
			Filters: asFilters,
		}

		err := cloud.Autoscaling().DescribeTagsPages(request, func(p *autoscaling.DescribeTagsOutput, lastPage bool) bool {
			for _, t := range p.Tags {
				switch *t.ResourceType {
				case "auto-scaling-group":
					asgNames = append(asgNames, t.ResourceId)
				default:
					glog.Warningf("Unknown resource type: %v", *t.ResourceType)

				}
			}
			return true
		})
		if err != nil {
			return nil, fmt.Errorf("error listing autoscaling cluster tags: %v", err)
		}
	}

	if len(asgNames) != 0 {
		request := &autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: asgNames,
		}
		err := cloud.Autoscaling().DescribeAutoScalingGroupsPages(request, func(p *autoscaling.DescribeAutoScalingGroupsOutput, lastPage bool) bool {
			for _, asg := range p.AutoScalingGroups {
				if !MatchesAsgTags(tags, asg.Tags) {
					// We used an inexact filter above
					continue
				}
				// Check for "Delete in progress" (the only use of .Status)
				if asg.Status != nil {
					glog.Warningf("Skipping ASG %v (which matches tags): %v", *asg.AutoScalingGroupARN, *asg.Status)
					continue
				}
				asgs = append(asgs, asg)
			}
			return true
		})
		if err != nil {
			return nil, fmt.Errorf("error listing autoscaling groups: %v", err)
		}

	}

	return asgs, nil
}

// FindAutoscalingLaunchConfiguration finds an AWS launch configuration given its name
func FindAutoscalingLaunchConfiguration(cloud awsup.AWSCloud, name string) (*autoscaling.LaunchConfiguration, error) {
	glog.V(2).Infof("Retrieving Autoscaling LaunchConfigurations %q", name)

	var results []*autoscaling.LaunchConfiguration

	request := &autoscaling.DescribeLaunchConfigurationsInput{
		LaunchConfigurationNames: []*string{&name},
	}
	err := cloud.Autoscaling().DescribeLaunchConfigurationsPages(request, func(p *autoscaling.DescribeLaunchConfigurationsOutput, lastPage bool) bool {
		for _, t := range p.LaunchConfigurations {
			results = append(results, t)
		}
		return true
	})
	if err != nil {
		return nil, fmt.Errorf("error listing autoscaling LaunchConfigurations: %v", err)
	}

	if len(results) == 0 {
		return nil, nil
	}
	if len(results) != 1 {
		return nil, fmt.Errorf("Found multiple LaunchConfigurations with name %q", name)
	}
	return results[0], nil
}

func MatchesAsgTags(tags map[string]string, actual []*autoscaling.TagDescription) bool {
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

//type DeletableResource interface {
//	Delete(cloud fi.Cloud) error
//}

func DeleteInstance(cloud fi.Cloud, t *ResourceTracker) error {
	c := cloud.(awsup.AWSCloud)

	id := t.ID
	glog.V(2).Infof("Deleting EC2 instance %q", id)
	request := &ec2.TerminateInstancesInput{
		InstanceIds: []*string{&id},
	}
	_, err := c.EC2().TerminateInstances(request)
	if err != nil {
		if awsup.AWSErrorCode(err) == "InvalidInstanceID.NotFound" {
			glog.V(2).Infof("Got InvalidInstanceID.NotFound error deleting instance %q; will treat as already-deleted", id)
		} else {
			return fmt.Errorf("error deleting Instance %q: %v", id, err)
		}
	}
	return nil
}

func DeleteCloudFormationStack(cloud fi.Cloud, t *ResourceTracker) error {
	c := cloud.(awsup.AWSCloud)

	id := t.ID
	glog.V(2).Infof("deleting CloudFormation stack %q %q", t.Name, id)

	request := &cloudformation.DeleteStackInput{}
	request.StackName = &t.Name

	_, err := c.CloudFormation().DeleteStack(request)
	if err != nil {
		return fmt.Errorf("error deleting CloudFormation stack %q: %v", id, err)
	}
	return nil
}

func DumpCloudFormationStack(r *ResourceTracker) (interface{}, error) {
	data := make(map[string]interface{})
	data["id"] = r.ID
	data["type"] = r.Type
	data["raw"] = r.obj
	return data, nil
}

func ListCloudFormationStacks(cloud fi.Cloud, clusterName string) ([]*ResourceTracker, error) {
	var trackers []*ResourceTracker
	request := &cloudformation.ListStacksInput{}
	c := cloud.(awsup.AWSCloud)
	response, err := c.CloudFormation().ListStacks(request)
	if err != nil {
		return nil, fmt.Errorf("Unable to list CloudFormation stacks: %v", err)
	}
	for _, stack := range response.StackSummaries {
		if *stack.StackName == clusterName {
			tracker := &ResourceTracker{
				Name:    *stack.StackName,
				ID:      *stack.StackId,
				Type:    "cloud-formation",
				deleter: DeleteCloudFormationStack,
				Dumper:  DumpCloudFormationStack,
				obj:     stack,
			}
			trackers = append(trackers, tracker)
		}
	}

	return trackers, nil
}

func ListInstances(cloud fi.Cloud, clusterName string) ([]*ResourceTracker, error) {
	c := cloud.(awsup.AWSCloud)

	glog.V(2).Infof("Querying EC2 instances")
	request := &ec2.DescribeInstancesInput{
		Filters: BuildEC2Filters(cloud),
	}

	var trackers []*ResourceTracker

	err := c.EC2().DescribeInstancesPages(request, func(p *ec2.DescribeInstancesOutput, lastPage bool) bool {
		for _, reservation := range p.Reservations {
			for _, instance := range reservation.Instances {
				id := aws.StringValue(instance.InstanceId)

				if instance.State != nil {
					stateName := aws.StringValue(instance.State.Name)
					switch stateName {
					case "terminated", "shutting-down":
						continue

					case "running", "stopped":
						// We need to delete
						glog.V(4).Infof("instance %q has state=%q", id, stateName)

					default:
						glog.Infof("unknown instance state for %q: %q", id, stateName)
					}
				}

				tracker := &ResourceTracker{
					Name:    FindName(instance.Tags),
					ID:      id,
					Type:    ec2.ResourceTypeInstance,
					deleter: DeleteInstance,
					Dumper:  DumpInstance,
					obj:     instance,
				}

				var blocks []string
				blocks = append(blocks, "vpc:"+aws.StringValue(instance.VpcId))

				for _, volume := range instance.BlockDeviceMappings {
					if volume.Ebs == nil {
						continue
					}
					blocks = append(blocks, "volume:"+aws.StringValue(volume.Ebs.VolumeId))
				}
				for _, sg := range instance.SecurityGroups {
					blocks = append(blocks, "security-group:"+aws.StringValue(sg.GroupId))
				}
				blocks = append(blocks, "subnet:"+aws.StringValue(instance.SubnetId))
				blocks = append(blocks, "vpc:"+aws.StringValue(instance.VpcId))

				tracker.blocks = blocks

				trackers = append(trackers, tracker)

			}
		}
		return true
	})
	if err != nil {
		return nil, fmt.Errorf("error describing instances: %v", err)
	}

	return trackers, nil
}

func DumpInstance(r *ResourceTracker) (interface{}, error) {
	data := make(map[string]interface{})
	data["id"] = r.ID
	data["type"] = ec2.ResourceTypeInstance
	data["raw"] = r.obj
	return data, nil
}

func DeleteSecurityGroup(cloud fi.Cloud, t *ResourceTracker) error {
	c := cloud.(awsup.AWSCloud)

	id := t.ID
	// First clear all inter-dependent rules
	// TODO: Move to a "pre-execute" phase?
	{
		request := &ec2.DescribeSecurityGroupsInput{
			GroupIds: []*string{&id},
		}
		response, err := c.EC2().DescribeSecurityGroups(request)
		if err != nil {
			if awsup.AWSErrorCode(err) == "InvalidGroup.NotFound" {
				glog.V(2).Infof("Got InvalidGroup.NotFound error describing SecurityGroup %q; will treat as already-deleted", id)
				return nil
			}
			return fmt.Errorf("error describing SecurityGroup %q: %v", id, err)
		}

		if len(response.SecurityGroups) == 0 {
			return nil
		}
		if len(response.SecurityGroups) != 1 {
			return fmt.Errorf("found mutiple SecurityGroups with ID %q", id)
		}
		sg := response.SecurityGroups[0]

		if len(sg.IpPermissions) != 0 {
			revoke := &ec2.RevokeSecurityGroupIngressInput{
				GroupId:       &id,
				IpPermissions: sg.IpPermissions,
			}
			_, err = c.EC2().RevokeSecurityGroupIngress(revoke)
			if err != nil {
				return fmt.Errorf("cannot revoke ingress for ID %q: %v", id, err)
			}
		}
	}

	{
		glog.V(2).Infof("Deleting EC2 SecurityGroup %q", id)
		request := &ec2.DeleteSecurityGroupInput{
			GroupId: &id,
		}
		_, err := c.EC2().DeleteSecurityGroup(request)
		if err != nil {
			if IsDependencyViolation(err) {
				return err
			}
			return fmt.Errorf("error deleting SecurityGroup %q: %v", id, err)
		}
	}
	return nil
}

func DumpSecurityGroup(r *ResourceTracker) (interface{}, error) {
	data := make(map[string]interface{})
	data["id"] = r.ID
	data["type"] = ec2.ResourceTypeSecurityGroup
	data["raw"] = r.obj
	return data, nil
}

func ListSecurityGroups(cloud fi.Cloud, clusterName string) ([]*ResourceTracker, error) {
	groups, err := DescribeSecurityGroups(cloud)
	if err != nil {
		return nil, err
	}

	var trackers []*ResourceTracker

	for _, sg := range groups {
		tracker := &ResourceTracker{
			Name:    FindName(sg.Tags),
			ID:      aws.StringValue(sg.GroupId),
			Type:    "security-group",
			deleter: DeleteSecurityGroup,
			Dumper:  DumpSecurityGroup,
			obj:     sg,
		}

		var blocks []string
		blocks = append(blocks, "vpc:"+aws.StringValue(sg.VpcId))

		tracker.blocks = blocks

		trackers = append(trackers, tracker)
	}

	return trackers, nil
}

func DescribeSecurityGroups(cloud fi.Cloud) ([]*ec2.SecurityGroup, error) {
	c := cloud.(awsup.AWSCloud)

	glog.V(2).Infof("Listing EC2 SecurityGroups")
	request := &ec2.DescribeSecurityGroupsInput{
		Filters: BuildEC2Filters(cloud),
	}
	response, err := c.EC2().DescribeSecurityGroups(request)
	if err != nil {
		return nil, fmt.Errorf("error listing SecurityGroups: %v", err)
	}

	return response.SecurityGroups, nil
}

func DeleteVolume(cloud fi.Cloud, r *ResourceTracker) error {
	c := cloud.(awsup.AWSCloud)

	id := r.ID

	glog.V(2).Infof("Deleting EC2 Volume %q", id)
	request := &ec2.DeleteVolumeInput{
		VolumeId: &id,
	}
	_, err := c.EC2().DeleteVolume(request)
	if err != nil {
		if IsDependencyViolation(err) {
			return err
		}
		if awsup.AWSErrorCode(err) == "InvalidVolume.NotFound" {
			// Concurrently deleted
			return nil
		}
		return fmt.Errorf("error deleting Volume %q: %v", id, err)
	}
	return nil
}

func ListVolumes(cloud fi.Cloud, clusterName string) ([]*ResourceTracker, error) {
	c := cloud.(awsup.AWSCloud)

	volumes, err := DescribeVolumes(cloud)
	if err != nil {
		return nil, err
	}
	var trackers []*ResourceTracker

	elasticIPs := make(map[string]bool)
	for _, volume := range volumes {
		id := aws.StringValue(volume.VolumeId)

		tracker := &ResourceTracker{
			Name:    FindName(volume.Tags),
			ID:      id,
			Type:    "volume",
			deleter: DeleteVolume,
		}

		var blocks []string
		//blocks = append(blocks, "vpc:" + aws.StringValue(rt.VpcId))

		tracker.blocks = blocks

		trackers = append(trackers, tracker)

		// Check for an elastic IP tag
		for _, tag := range volume.Tags {
			name := aws.StringValue(tag.Key)
			ip := ""
			if name == "kubernetes.io/master-ip" {
				ip = aws.StringValue(tag.Value)
			}
			if ip != "" {
				elasticIPs[ip] = true
			}
		}

	}

	if len(elasticIPs) != 0 {
		glog.V(2).Infof("Querying EC2 Elastic IPs")
		request := &ec2.DescribeAddressesInput{}
		response, err := c.EC2().DescribeAddresses(request)
		if err != nil {
			return nil, fmt.Errorf("error describing addresses: %v", err)
		}

		for _, address := range response.Addresses {
			ip := aws.StringValue(address.PublicIp)
			if !elasticIPs[ip] {
				continue
			}

			tracker := &ResourceTracker{
				Name:    ip,
				ID:      aws.StringValue(address.AllocationId),
				Type:    TypeElasticIp,
				deleter: DeleteElasticIP,
			}

			trackers = append(trackers, tracker)

		}
	}

	return trackers, nil
}

func DescribeVolumes(cloud fi.Cloud) ([]*ec2.Volume, error) {
	c := cloud.(awsup.AWSCloud)

	var volumes []*ec2.Volume

	glog.V(2).Infof("Listing EC2 Volumes")
	request := &ec2.DescribeVolumesInput{
		Filters: BuildEC2Filters(c),
	}

	err := c.EC2().DescribeVolumesPages(request, func(p *ec2.DescribeVolumesOutput, lastPage bool) bool {
		for _, volume := range p.Volumes {
			volumes = append(volumes, volume)
		}
		return true
	})
	if err != nil {
		return nil, fmt.Errorf("error describing volumes: %v", err)
	}

	return volumes, nil
}

func DeleteKeypair(cloud fi.Cloud, r *ResourceTracker) error {
	c := cloud.(awsup.AWSCloud)

	name := r.Name

	glog.V(2).Infof("Deleting EC2 Keypair %q", name)
	request := &ec2.DeleteKeyPairInput{
		KeyName: &name,
	}
	_, err := c.EC2().DeleteKeyPair(request)
	if err != nil {
		return fmt.Errorf("error deleting KeyPair %q: %v", name, err)
	}
	return nil
}

func ListKeypairs(cloud fi.Cloud, clusterName string) ([]*ResourceTracker, error) {
	if !strings.Contains(clusterName, ".") {
		glog.Infof("cluster %q is legacy (kube-up) cluster; won't delete keypairs", clusterName)
		return nil, nil
	}

	c := cloud.(awsup.AWSCloud)

	keypairName := "kubernetes." + clusterName

	glog.V(2).Infof("Listing EC2 Keypairs")
	request := &ec2.DescribeKeyPairsInput{
	// We need to match both the name and a prefix
	//Filters: []*ec2.Filter{awsup.NewEC2Filter("key-name", keypairName)},
	}
	response, err := c.EC2().DescribeKeyPairs(request)
	if err != nil {
		return nil, fmt.Errorf("error listing KeyPairs: %v", err)
	}

	var trackers []*ResourceTracker

	for _, keypair := range response.KeyPairs {
		name := aws.StringValue(keypair.KeyName)
		if name != keypairName && !strings.HasPrefix(name, keypairName+"-") {
			continue
		}
		tracker := &ResourceTracker{
			Name:    name,
			ID:      name,
			Type:    "keypair",
			deleter: DeleteKeypair,
		}

		trackers = append(trackers, tracker)
	}

	return trackers, nil
}

func IsDependencyViolation(err error) bool {
	code := awsup.AWSErrorCode(err)
	switch code {
	case "":
		return false
	case "DependencyViolation", "VolumeInUse", "InvalidIPAddress.InUse":
		return true
	default:
		glog.Infof("unexpected aws error code: %q", code)
		return false
	}
}

func DeleteSubnet(cloud fi.Cloud, tracker *ResourceTracker) error {
	c := cloud.(awsup.AWSCloud)

	id := tracker.ID

	glog.V(2).Infof("Deleting EC2 Subnet %q", id)
	request := &ec2.DeleteSubnetInput{
		SubnetId: &id,
	}
	_, err := c.EC2().DeleteSubnet(request)
	if err != nil {
		if awsup.AWSErrorCode(err) == "InvalidSubnetID.NotFound" {
			glog.V(2).Infof("Got InvalidSubnetID.NotFound error deleting subnet %q; will treat as already-deleted", id)
			return nil
		} else if IsDependencyViolation(err) {
			return err
		}
		return fmt.Errorf("error deleting Subnet %q: %v", id, err)
	}
	return nil
}

func ListSubnets(cloud fi.Cloud, clusterName string) ([]*ResourceTracker, error) {
	c := cloud.(awsup.AWSCloud)
	subnets, err := DescribeSubnets(cloud)
	if err != nil {
		return nil, fmt.Errorf("error listing subnets: %v", err)
	}

	var trackers []*ResourceTracker
	elasticIPs := sets.NewString()
	ownedElasticIPs := sets.NewString()
	natGatewayIds := sets.NewString()
	ownedNatGatewayIds := sets.NewString()
	for _, subnet := range subnets {
		subnetID := aws.StringValue(subnet.SubnetId)

		shared := HasSharedTag("subnet:"+subnetID, subnet.Tags, clusterName)
		tracker := &ResourceTracker{
			Name:    FindName(subnet.Tags),
			ID:      subnetID,
			Type:    "subnet",
			deleter: DeleteSubnet,
			Shared:  shared,
		}
		tracker.blocks = append(tracker.blocks, "vpc:"+aws.StringValue(subnet.VpcId))
		trackers = append(trackers, tracker)

		// Get tags and append with EIPs/NGWs as needed
		for _, tag := range subnet.Tags {
			name := aws.StringValue(tag.Key)
			if name == "AssociatedElasticIp" {
				eip := aws.StringValue(tag.Value)
				if eip != "" {
					elasticIPs.Insert(eip)
					// A shared subnet means the EIP is not owned
					if !shared {
						ownedElasticIPs.Insert(eip)
					}
				}
			}
			if name == "AssociatedNatgateway" {
				ngwID := aws.StringValue(tag.Value)
				if ngwID != "" {
					natGatewayIds.Insert(ngwID)
					// A shared subnet means the NAT gateway is not owned
					if !shared {
						ownedNatGatewayIds.Insert(ngwID)
					}
				}
			}
		}
	}

	// Associated Elastic IPs
	if elasticIPs.Len() != 0 {
		glog.V(2).Infof("Querying EC2 Elastic IPs")
		request := &ec2.DescribeAddressesInput{}
		response, err := c.EC2().DescribeAddresses(request)
		if err != nil {
			return nil, fmt.Errorf("error describing addresses: %v", err)
		}

		for _, address := range response.Addresses {
			ip := aws.StringValue(address.PublicIp)
			if !elasticIPs.Has(ip) {
				continue
			}

			tracker := &ResourceTracker{
				Name:    ip,
				ID:      aws.StringValue(address.AllocationId),
				Type:    TypeElasticIp,
				deleter: DeleteElasticIP,
				Shared:  !ownedElasticIPs.Has(ip),
			}
			trackers = append(trackers, tracker)
		}
	}

	// Associated Nat Gateways
	// Note: we must not delete any shared NAT Gateways here.
	// Since we don't have tagging on the NGWs, we have to read the route tables
	if natGatewayIds.Len() != 0 {

		rtRequest := &ec2.DescribeRouteTablesInput{}
		rtResponse, err := c.EC2().DescribeRouteTables(rtRequest)
		if err != nil {
			return nil, fmt.Errorf("error describing RouteTables: %v", err)
		}
		// sharedNgwIds is the set of IDs for shared NGWs, that we should not delete
		sharedNgwIds := sets.NewString()
		{
			for _, rt := range rtResponse.RouteTables {
				for _, t := range rt.Tags {
					k := aws.StringValue(t.Key)
					v := aws.StringValue(t.Value)

					if k == "AssociatedNatgateway" {
						sharedNgwIds.Insert(v)
					}
				}
			}

		}

		glog.V(2).Infof("Querying Nat Gateways")
		request := &ec2.DescribeNatGatewaysInput{}
		response, err := c.EC2().DescribeNatGateways(request)
		if err != nil {
			return nil, fmt.Errorf("error describing NatGateways: %v", err)
		}

		for _, ngw := range response.NatGateways {
			id := aws.StringValue(ngw.NatGatewayId)
			if !natGatewayIds.Has(id) {
				continue
			}

			tracker := &ResourceTracker{
				Name:    id,
				ID:      id,
				Type:    TypeNatGateway,
				deleter: DeleteNatGateway,
				Shared:  sharedNgwIds.Has(id) || !ownedNatGatewayIds.Has(id),
			}

			// The NAT gateway blocks deletion of any associated Elastic IPs
			for _, address := range ngw.NatGatewayAddresses {
				if address.AllocationId != nil {
					tracker.blocks = append(tracker.blocks, TypeElasticIp+":"+aws.StringValue(address.AllocationId))
				}
			}

			trackers = append(trackers, tracker)
		}
	}

	return trackers, nil
}

func DescribeSubnets(cloud fi.Cloud) ([]*ec2.Subnet, error) {
	c := cloud.(awsup.AWSCloud)

	glog.V(2).Infof("Listing EC2 subnets")
	request := &ec2.DescribeSubnetsInput{
		Filters: BuildEC2Filters(cloud),
	}
	response, err := c.EC2().DescribeSubnets(request)
	if err != nil {
		return nil, fmt.Errorf("error listing subnets: %v", err)
	}

	return response.Subnets, nil
}

func DeleteRouteTable(cloud fi.Cloud, r *ResourceTracker) error {
	c := cloud.(awsup.AWSCloud)

	id := r.ID

	glog.V(2).Infof("Deleting EC2 RouteTable %q", id)
	request := &ec2.DeleteRouteTableInput{
		RouteTableId: &id,
	}
	_, err := c.EC2().DeleteRouteTable(request)
	if err != nil {
		if awsup.AWSErrorCode(err) == "InvalidRouteTableID.NotFound" {
			glog.V(2).Infof("Got InvalidRouteTableID.NotFound error describing RouteTable %q; will treat as already-deleted", id)
			return nil
		}

		if IsDependencyViolation(err) {
			return err
		}
		return fmt.Errorf("error deleting RouteTable %q: %v", id, err)
	}
	return nil
}

// DescribeRouteTablesIgnoreTags returns all ec2.RouteTable, ignoring tags
func DescribeRouteTablesIgnoreTags(cloud fi.Cloud) ([]*ec2.RouteTable, error) {
	c := cloud.(awsup.AWSCloud)

	glog.V(2).Infof("Listing all RouteTables")
	request := &ec2.DescribeRouteTablesInput{}
	response, err := c.EC2().DescribeRouteTables(request)
	if err != nil {
		return nil, fmt.Errorf("error listing RouteTables: %v", err)
	}

	return response.RouteTables, nil
}

// DescribeRouteTables lists route-tables tagged for the cloud
func DescribeRouteTables(cloud fi.Cloud) ([]*ec2.RouteTable, error) {
	c := cloud.(awsup.AWSCloud)

	glog.V(2).Infof("Listing EC2 RouteTables")
	request := &ec2.DescribeRouteTablesInput{
		Filters: BuildEC2Filters(cloud),
	}
	response, err := c.EC2().DescribeRouteTables(request)
	if err != nil {
		return nil, fmt.Errorf("error listing RouteTables: %v", err)
	}

	return response.RouteTables, nil
}

func ListRouteTables(cloud fi.Cloud, clusterName string) ([]*ResourceTracker, error) {
	routeTables, err := DescribeRouteTables(cloud)
	if err != nil {
		return nil, err
	}

	var trackers []*ResourceTracker

	for _, rt := range routeTables {
		tracker := buildTrackerForRouteTable(rt)
		trackers = append(trackers, tracker)
	}

	return trackers, nil
}

func buildTrackerForRouteTable(rt *ec2.RouteTable) *ResourceTracker {
	tracker := &ResourceTracker{
		Name:    FindName(rt.Tags),
		ID:      aws.StringValue(rt.RouteTableId),
		Type:    ec2.ResourceTypeRouteTable,
		deleter: DeleteRouteTable,
	}

	var blocks []string
	var blocked []string

	blocks = append(blocks, "vpc:"+aws.StringValue(rt.VpcId))

	for _, a := range rt.Associations {
		blocked = append(blocked, "subnet:"+aws.StringValue(a.SubnetId))
	}

	tracker.blocks = blocks
	tracker.blocked = blocked

	return tracker
}

func DeleteDhcpOptions(cloud fi.Cloud, r *ResourceTracker) error {
	c := cloud.(awsup.AWSCloud)

	id := r.ID

	glog.V(2).Infof("Deleting EC2 DhcpOptions %q", id)
	request := &ec2.DeleteDhcpOptionsInput{
		DhcpOptionsId: &id,
	}
	_, err := c.EC2().DeleteDhcpOptions(request)
	if err != nil {
		if IsDependencyViolation(err) {
			return err
		}
		return fmt.Errorf("error deleting DhcpOptions %q: %v", id, err)
	}
	return nil
}

func ListDhcpOptions(cloud fi.Cloud, clusterName string) ([]*ResourceTracker, error) {
	dhcpOptions, err := DescribeDhcpOptions(cloud)
	if err != nil {
		return nil, err
	}

	var trackers []*ResourceTracker

	for _, o := range dhcpOptions {
		tracker := &ResourceTracker{
			Name:    FindName(o.Tags),
			ID:      aws.StringValue(o.DhcpOptionsId),
			Type:    "dhcp-options",
			deleter: DeleteDhcpOptions,
		}

		var blocks []string

		tracker.blocks = blocks

		trackers = append(trackers, tracker)
	}

	return trackers, nil
}

func DescribeDhcpOptions(cloud fi.Cloud) ([]*ec2.DhcpOptions, error) {
	c := cloud.(awsup.AWSCloud)

	glog.V(2).Infof("Listing EC2 DhcpOptions")
	request := &ec2.DescribeDhcpOptionsInput{
		Filters: BuildEC2Filters(cloud),
	}
	response, err := c.EC2().DescribeDhcpOptions(request)
	if err != nil {
		return nil, fmt.Errorf("error listing DhcpOptions: %v", err)
	}

	return response.DhcpOptions, nil
}

func DeleteInternetGateway(cloud fi.Cloud, r *ResourceTracker) error {
	c := cloud.(awsup.AWSCloud)

	id := r.ID

	var igw *ec2.InternetGateway
	{
		request := &ec2.DescribeInternetGatewaysInput{
			InternetGatewayIds: []*string{&id},
		}
		response, err := c.EC2().DescribeInternetGateways(request)
		if err != nil {
			if awsup.AWSErrorCode(err) == "InvalidInternetGatewayID.NotFound" {
				glog.Infof("Internet gateway %q not found; assuming already deleted", id)
				return nil
			}

			return fmt.Errorf("error describing InternetGateway %q: %v", id, err)
		}
		if response == nil || len(response.InternetGateways) == 0 {
			return nil
		}
		if len(response.InternetGateways) != 1 {
			return fmt.Errorf("found multiple InternetGateways with id %q", id)
		}
		igw = response.InternetGateways[0]
	}

	for _, a := range igw.Attachments {
		glog.V(2).Infof("Detaching EC2 InternetGateway %q", id)
		request := &ec2.DetachInternetGatewayInput{
			InternetGatewayId: &id,
			VpcId:             a.VpcId,
		}
		_, err := c.EC2().DetachInternetGateway(request)
		if err != nil {
			if IsDependencyViolation(err) {
				return err
			}
			return fmt.Errorf("error detaching InternetGateway %q: %v", id, err)
		}
	}

	{
		glog.V(2).Infof("Deleting EC2 InternetGateway %q", id)
		request := &ec2.DeleteInternetGatewayInput{
			InternetGatewayId: &id,
		}
		_, err := c.EC2().DeleteInternetGateway(request)
		if err != nil {
			if IsDependencyViolation(err) {
				return err
			}
			if awsup.AWSErrorCode(err) == "InvalidInternetGatewayID.NotFound" {
				glog.Infof("Internet gateway %q not found; assuming already deleted", id)
				return nil
			}
			return fmt.Errorf("error deleting InternetGateway %q: %v", id, err)
		}
	}

	return nil
}

func ListInternetGateways(cloud fi.Cloud, clusterName string) ([]*ResourceTracker, error) {
	gateways, err := DescribeInternetGateways(cloud)
	if err != nil {
		return nil, err
	}

	var trackers []*ResourceTracker

	for _, o := range gateways {
		tracker := &ResourceTracker{
			Name:    FindName(o.Tags),
			ID:      aws.StringValue(o.InternetGatewayId),
			Type:    "internet-gateway",
			deleter: DeleteInternetGateway,
		}

		var blocks []string
		for _, a := range o.Attachments {
			if aws.StringValue(a.VpcId) != "" {
				blocks = append(blocks, "vpc:"+aws.StringValue(a.VpcId))
			}
		}
		tracker.blocks = blocks

		trackers = append(trackers, tracker)
	}

	return trackers, nil
}

func DescribeInternetGateways(cloud fi.Cloud) ([]*ec2.InternetGateway, error) {
	c := cloud.(awsup.AWSCloud)

	glog.V(2).Infof("Listing EC2 InternetGateways")
	request := &ec2.DescribeInternetGatewaysInput{
		Filters: BuildEC2Filters(cloud),
	}
	response, err := c.EC2().DescribeInternetGateways(request)
	if err != nil {
		return nil, fmt.Errorf("error listing InternetGateway: %v", err)
	}

	var gateways []*ec2.InternetGateway
	for _, o := range response.InternetGateways {
		gateways = append(gateways, o)
	}

	return gateways, nil
}

// DescribeInternetGatewaysIgnoreTags returns all ec2.InternetGateways, ignoring tags
// (gateways were not always tagged in kube-up)
func DescribeInternetGatewaysIgnoreTags(cloud fi.Cloud) ([]*ec2.InternetGateway, error) {
	c := cloud.(awsup.AWSCloud)

	glog.V(2).Infof("Listing all Internet Gateways")

	request := &ec2.DescribeInternetGatewaysInput{}
	response, err := c.EC2().DescribeInternetGateways(request)
	if err != nil {
		return nil, fmt.Errorf("error listing (all) InternetGateways: %v", err)
	}

	var gateways []*ec2.InternetGateway

	for _, igw := range response.InternetGateways {
		gateways = append(gateways, igw)
	}

	return gateways, nil
}

func DeleteVPC(cloud fi.Cloud, r *ResourceTracker) error {
	c := cloud.(awsup.AWSCloud)

	id := r.ID

	glog.V(2).Infof("Deleting EC2 VPC %q", id)
	request := &ec2.DeleteVpcInput{
		VpcId: &id,
	}
	_, err := c.EC2().DeleteVpc(request)
	if err != nil {
		if IsDependencyViolation(err) {
			return err
		}
		return fmt.Errorf("error deleting VPC %q: %v", id, err)
	}
	return nil
}

func DumpVPC(r *ResourceTracker) (interface{}, error) {
	data := make(map[string]interface{})
	data["id"] = r.ID
	data["type"] = ec2.ResourceTypeVpc
	data["raw"] = r.obj
	return data, nil
}

func DescribeVPCs(cloud fi.Cloud) ([]*ec2.Vpc, error) {
	c := cloud.(awsup.AWSCloud)

	glog.V(2).Infof("Listing EC2 VPC")
	request := &ec2.DescribeVpcsInput{
		Filters: BuildEC2Filters(cloud),
	}
	response, err := c.EC2().DescribeVpcs(request)
	if err != nil {
		return nil, fmt.Errorf("error listing VPCs: %v", err)
	}

	return response.Vpcs, nil
}

func ListVPCs(cloud fi.Cloud, clusterName string) ([]*ResourceTracker, error) {
	vpcs, err := DescribeVPCs(cloud)
	if err != nil {
		return nil, err
	}

	var trackers []*ResourceTracker
	for _, v := range vpcs {
		vpcID := aws.StringValue(v.VpcId)

		tracker := &ResourceTracker{
			Name:    FindName(v.Tags),
			ID:      vpcID,
			Type:    ec2.ResourceTypeVpc,
			deleter: DeleteVPC,
			Dumper:  DumpVPC,
			obj:     v,
			Shared:  HasSharedTag(ec2.ResourceTypeVpc+":"+vpcID, v.Tags, clusterName),
		}

		var blocks []string
		blocks = append(blocks, "dhcp-options:"+aws.StringValue(v.DhcpOptionsId))

		tracker.blocks = blocks

		trackers = append(trackers, tracker)
	}

	return trackers, nil
}

func DeleteAutoScalingGroup(cloud fi.Cloud, r *ResourceTracker) error {
	c := cloud.(awsup.AWSCloud)

	id := r.ID

	glog.V(2).Infof("Deleting autoscaling group %q", id)
	request := &autoscaling.DeleteAutoScalingGroupInput{
		AutoScalingGroupName: &id,
		ForceDelete:          aws.Bool(true),
	}
	_, err := c.Autoscaling().DeleteAutoScalingGroup(request)
	if err != nil {
		if IsDependencyViolation(err) {
			return err
		}
		return fmt.Errorf("error deleting autoscaling group %q: %v", id, err)
	}
	return nil
}

func ListAutoScalingGroups(cloud fi.Cloud, clusterName string) ([]*ResourceTracker, error) {
	c := cloud.(awsup.AWSCloud)

	tags := c.Tags()

	asgs, err := FindAutoscalingGroups(c, tags)
	if err != nil {
		return nil, err
	}

	var trackers []*ResourceTracker

	for _, asg := range asgs {
		tracker := &ResourceTracker{
			Name:    FindASGName(asg.Tags),
			ID:      aws.StringValue(asg.AutoScalingGroupName),
			Type:    "autoscaling-group",
			deleter: DeleteAutoScalingGroup,
		}

		var blocks []string
		subnets := aws.StringValue(asg.VPCZoneIdentifier)
		for _, subnet := range strings.Split(subnets, ",") {
			if subnet == "" {
				continue
			}
			blocks = append(blocks, "subnet:"+subnet)
		}
		blocks = append(blocks, TypeAutoscalingLaunchConfig+":"+aws.StringValue(asg.LaunchConfigurationName))

		tracker.blocks = blocks

		trackers = append(trackers, tracker)
	}

	return trackers, nil
}

func FindAutoScalingLaunchConfigurations(cloud fi.Cloud, securityGroups sets.String) ([]*ResourceTracker, error) {
	c := cloud.(awsup.AWSCloud)

	glog.V(2).Infof("Finding all Autoscaling LaunchConfigurations by security group")

	var trackers []*ResourceTracker

	request := &autoscaling.DescribeLaunchConfigurationsInput{}
	err := c.Autoscaling().DescribeLaunchConfigurationsPages(request, func(p *autoscaling.DescribeLaunchConfigurationsOutput, lastPage bool) bool {
		for _, t := range p.LaunchConfigurations {
			found := false
			for _, sg := range t.SecurityGroups {
				if securityGroups.Has(aws.StringValue(sg)) {
					found = true
					break
				}
			}
			if !found {
				continue
			}

			tracker := &ResourceTracker{
				Name:    aws.StringValue(t.LaunchConfigurationName),
				ID:      aws.StringValue(t.LaunchConfigurationName),
				Type:    TypeAutoscalingLaunchConfig,
				deleter: DeleteAutoscalingLaunchConfiguration,
			}

			var blocks []string
			//blocks = append(blocks, TypeAutoscalingLaunchConfig + ":" + aws.StringValue(asg.LaunchConfigurationName))

			tracker.blocks = blocks

			trackers = append(trackers, tracker)
		}
		return true
	})
	if err != nil {
		return nil, fmt.Errorf("error listing autoscaling LaunchConfigurations: %v", err)
	}

	return trackers, nil
}

func FindNatGateways(cloud fi.Cloud, routeTables map[string]*ResourceTracker) ([]*ResourceTracker, error) {
	if len(routeTables) == 0 {
		return nil, nil
	}

	c := cloud.(awsup.AWSCloud)

	natGatewayIds := sets.NewString()
	ownedNatGatewayIds := sets.NewString()
	{
		request := &ec2.DescribeRouteTablesInput{}
		for _, routeTable := range routeTables {
			request.RouteTableIds = append(request.RouteTableIds, aws.String(routeTable.ID))
		}
		response, err := c.EC2().DescribeRouteTables(request)
		if err != nil {
			return nil, fmt.Errorf("error from DescribeRouteTables: %v", err)
		}
		for _, rt := range response.RouteTables {
			routeTableID := aws.StringValue(rt.RouteTableId)
			resource := routeTables[routeTableID]
			if resource == nil {
				// We somehow got a route table that we didn't ask for
				glog.Warningf("unable to find resource for route table %s", routeTableID)
				continue
			}

			shared := resource.Shared
			for _, t := range rt.Tags {
				k := *t.Key
				// v := *t.Value
				if k == "AssociatedNatgateway" {
					shared = true
				}
			}

			for _, route := range rt.Routes {
				if route.NatGatewayId != nil {
					natGatewayIds.Insert(*route.NatGatewayId)
					if !shared {
						ownedNatGatewayIds.Insert(*route.NatGatewayId)
					}
				}
			}
		}
	}

	var trackers []*ResourceTracker
	if len(natGatewayIds) != 0 {
		request := &ec2.DescribeNatGatewaysInput{}
		for natGatewayId := range natGatewayIds {
			request.NatGatewayIds = append(request.NatGatewayIds, aws.String(natGatewayId))
		}
		response, err := c.EC2().DescribeNatGateways(request)
		if err != nil {
			return nil, fmt.Errorf("error from DescribeNatGateways: %v", err)
		}

		if response.NextToken != nil {
			return nil, fmt.Errorf("NextToken set from DescribeNatGateways, but pagination not implemented")
		}

		for _, t := range response.NatGateways {
			natGatewayId := aws.StringValue(t.NatGatewayId)
			ngwTracker := &ResourceTracker{
				Name:    natGatewayId,
				ID:      natGatewayId,
				Type:    TypeNatGateway,
				deleter: DeleteNatGateway,
				Shared:  !ownedNatGatewayIds.Has(natGatewayId),
			}
			trackers = append(trackers, ngwTracker)

			// If we're deleting the NatGateway, we should delete the ElasticIP also
			for _, address := range t.NatGatewayAddresses {
				if address.AllocationId != nil {
					name := aws.StringValue(address.PublicIp)
					if name == "" {
						name = aws.StringValue(address.PrivateIp)
					}
					if name == "" {
						name = aws.StringValue(address.AllocationId)
					}

					eipTracker := &ResourceTracker{
						Name:    name,
						ID:      aws.StringValue(address.AllocationId),
						Type:    TypeElasticIp,
						deleter: DeleteElasticIP,
						Shared:  !ownedNatGatewayIds.Has(natGatewayId),
					}
					trackers = append(trackers, eipTracker)

					ngwTracker.blocks = append(ngwTracker.blocks, eipTracker.Type+":"+eipTracker.ID)
				}
			}
		}
	}

	return trackers, nil
}

// extractClusterName performs string-matching / parsing to determine the ClusterName in some instance-data
// It returns "" if it could not be (uniquely) determined
func extractClusterName(userData string) string {
	clusterName := ""

	scanner := bufio.NewScanner(bytes.NewReader([]byte(userData)))
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "INSTANCE_PREFIX:") {
			// kube-up
			// Match:
			// INSTANCE_PREFIX: 'clustername'
			// INSTANCE_PREFIX: "clustername"
			// INSTANCE_PREFIX: clustername
			line = strings.TrimPrefix(line, "INSTANCE_PREFIX:")
		} else if strings.HasPrefix(line, "ClusterName:") {
			// kops
			// Match:
			// ClusterName: 'clustername'
			// ClusterName: "clustername"
			// ClusterName: clustername
			line = strings.TrimPrefix(line, "ClusterName:")
		} else {
			continue
		}

		line = strings.TrimSpace(line)
		line = strings.Trim(line, "'\"")
		if clusterName != "" && clusterName != line {
			glog.Warning("cannot uniquely determine cluster-name, found %q and %q", line, clusterName)
			return ""
		}
		clusterName = line

	}
	if err := scanner.Err(); err != nil {
		glog.Warning("error scanning UserData: %v", err)
		return ""
	}

	return clusterName

}
func DeleteAutoscalingLaunchConfiguration(cloud fi.Cloud, r *ResourceTracker) error {
	c := cloud.(awsup.AWSCloud)

	id := r.ID
	glog.V(2).Infof("Deleting autoscaling LaunchConfiguration %q", id)
	request := &autoscaling.DeleteLaunchConfigurationInput{
		LaunchConfigurationName: &id,
	}
	_, err := c.Autoscaling().DeleteLaunchConfiguration(request)
	if err != nil {
		return fmt.Errorf("error deleting autoscaling LaunchConfiguration %q: %v", id, err)
	}
	return nil
}

func DeleteELB(cloud fi.Cloud, r *ResourceTracker) error {
	c := cloud.(awsup.AWSCloud)

	id := r.ID

	glog.V(2).Infof("Deleting ELB %q", id)
	request := &elb.DeleteLoadBalancerInput{
		LoadBalancerName: &id,
	}
	_, err := c.ELB().DeleteLoadBalancer(request)
	if err != nil {
		if IsDependencyViolation(err) {
			return err
		}
		return fmt.Errorf("error deleting LoadBalancer %q: %v", id, err)
	}
	return nil
}

func DumpELB(r *ResourceTracker) (interface{}, error) {
	data := make(map[string]interface{})
	data["id"] = r.ID
	data["type"] = TypeLoadBalancer
	data["raw"] = r.obj
	return data, nil
}

func ListELBs(cloud fi.Cloud, clusterName string) ([]*ResourceTracker, error) {
	elbs, elbTags, err := DescribeELBs(cloud)
	if err != nil {
		return nil, err
	}

	var trackers []*ResourceTracker
	for _, elb := range elbs {
		id := aws.StringValue(elb.LoadBalancerName)
		tracker := &ResourceTracker{
			Name:    FindELBName(elbTags[id]),
			ID:      id,
			Type:    TypeLoadBalancer,
			deleter: DeleteELB,
			Dumper:  DumpELB,
			obj:     elb,
		}

		var blocks []string
		for _, sg := range elb.SecurityGroups {
			blocks = append(blocks, "security-group:"+aws.StringValue(sg))
		}
		for _, s := range elb.Subnets {
			blocks = append(blocks, "subnet:"+aws.StringValue(s))
		}
		blocks = append(blocks, "vpc:"+aws.StringValue(elb.VPCId))

		tracker.blocks = blocks

		trackers = append(trackers, tracker)
	}

	return trackers, nil
}

func DescribeELBs(cloud fi.Cloud) ([]*elb.LoadBalancerDescription, map[string][]*elb.Tag, error) {
	c := cloud.(awsup.AWSCloud)
	tags := c.Tags()

	glog.V(2).Infof("Listing all ELBs")

	request := &elb.DescribeLoadBalancersInput{}
	// ELB DescribeTags has a limit of 20 names, so we set the page size here to 20 also
	request.PageSize = aws.Int64(20)

	var elbs []*elb.LoadBalancerDescription
	elbTags := make(map[string][]*elb.Tag)

	var innerError error
	err := c.ELB().DescribeLoadBalancersPages(request, func(p *elb.DescribeLoadBalancersOutput, lastPage bool) bool {
		if len(p.LoadBalancerDescriptions) == 0 {
			return true
		}

		tagRequest := &elb.DescribeTagsInput{}

		nameToELB := make(map[string]*elb.LoadBalancerDescription)
		for _, elb := range p.LoadBalancerDescriptions {
			name := aws.StringValue(elb.LoadBalancerName)
			nameToELB[name] = elb

			tagRequest.LoadBalancerNames = append(tagRequest.LoadBalancerNames, elb.LoadBalancerName)
		}

		tagResponse, err := c.ELB().DescribeTags(tagRequest)
		if err != nil {
			innerError = fmt.Errorf("error listing elb Tags: %v", err)
			return false
		}

		for _, t := range tagResponse.TagDescriptions {
			elbName := aws.StringValue(t.LoadBalancerName)

			if !matchesElbTags(tags, t.Tags) {
				continue
			}

			elbTags[elbName] = t.Tags

			elb := nameToELB[elbName]
			elbs = append(elbs, elb)
		}
		return true
	})
	if err != nil {
		return nil, nil, fmt.Errorf("error describing LoadBalancers: %v", err)
	}
	if innerError != nil {
		return nil, nil, fmt.Errorf("error describing LoadBalancers: %v", innerError)
	}

	return elbs, elbTags, nil
}

func DeleteElasticIP(cloud fi.Cloud, t *ResourceTracker) error {
	c := cloud.(awsup.AWSCloud)

	id := t.ID

	glog.V(2).Infof("Releasing IP %s", t.Name)
	request := &ec2.ReleaseAddressInput{
		AllocationId: &id,
	}
	_, err := c.EC2().ReleaseAddress(request)
	if err != nil {
		if awsup.AWSErrorCode(err) == "InvalidAllocationID.NotFound" {
			glog.V(2).Infof("Got InvalidAllocationID.NotFound error describing ElasticIP %q; will treat as already-deleted", id)
			return nil
		}

		if IsDependencyViolation(err) {
			return err
		}
		return fmt.Errorf("error deleting elastic ip %q: %v", t.Name, err)
	}
	return nil
}

func DeleteNatGateway(cloud fi.Cloud, t *ResourceTracker) error {
	c := cloud.(awsup.AWSCloud)

	id := t.ID

	glog.V(2).Infof("Removing NatGateway %s", t.Name)
	request := &ec2.DeleteNatGatewayInput{
		NatGatewayId: &id,
	}
	_, err := c.EC2().DeleteNatGateway(request)
	if err != nil {
		if IsDependencyViolation(err) {
			return err
		}
		return fmt.Errorf("error deleting ngw %q: %v", t.Name, err)
	}
	return nil
}

func deleteRoute53Records(cloud fi.Cloud, zone *route53.HostedZone, trackers []*ResourceTracker) error {
	c := cloud.(awsup.AWSCloud)

	var changes []*route53.Change
	var names []string
	for _, tracker := range trackers {
		names = append(names, tracker.Name)
		changes = append(changes, &route53.Change{
			Action:            aws.String("DELETE"),
			ResourceRecordSet: tracker.obj.(*route53.ResourceRecordSet),
		})
	}
	human := strings.Join(names, ", ")
	glog.V(2).Infof("Deleting route53 records %q", human)

	changeBatch := &route53.ChangeBatch{
		Changes: changes,
	}
	request := &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: zone.Id,
		ChangeBatch:  changeBatch,
	}
	_, err := c.Route53().ChangeResourceRecordSets(request)
	if err != nil {
		return fmt.Errorf("error deleting route53 record %q: %v", human, err)
	}
	return nil
}

func ListRoute53Records(cloud fi.Cloud, clusterName string) ([]*ResourceTracker, error) {
	var trackers []*ResourceTracker

	if dns.IsGossipHostname(clusterName) {
		return trackers, nil
	}

	c := cloud.(awsup.AWSCloud)

	// Normalize cluster name, with leading "."
	clusterName = "." + strings.TrimSuffix(clusterName, ".")

	// TODO: If we have the zone id in the cluster spec, use it!
	var zones []*route53.HostedZone
	{
		glog.V(2).Infof("Querying for all route53 zones")

		request := &route53.ListHostedZonesInput{}
		err := c.Route53().ListHostedZonesPages(request, func(p *route53.ListHostedZonesOutput, lastPage bool) bool {
			for _, zone := range p.HostedZones {
				zoneName := aws.StringValue(zone.Name)
				zoneName = "." + strings.TrimSuffix(zoneName, ".")

				if strings.HasSuffix(clusterName, zoneName) {
					zones = append(zones, zone)
				}
			}
			return true
		})
		if err != nil {
			return nil, fmt.Errorf("error querying for route53 zones: %v", err)
		}
	}

	for i := range zones {
		// Be super careful because we close over this later (in groupDeleter)
		zone := zones[i]

		hostedZoneID := strings.TrimPrefix(aws.StringValue(zone.Id), "/hostedzone/")

		glog.V(2).Infof("Querying for records in zone: %q", aws.StringValue(zone.Name))
		request := &route53.ListResourceRecordSetsInput{
			HostedZoneId: zone.Id,
		}
		err := c.Route53().ListResourceRecordSetsPages(request, func(p *route53.ListResourceRecordSetsOutput, lastPage bool) bool {
			for _, rrs := range p.ResourceRecordSets {
				if aws.StringValue(rrs.Type) != "A" {
					continue
				}

				name := aws.StringValue(rrs.Name)
				name = "." + strings.TrimSuffix(name, ".")

				if !strings.HasSuffix(name, clusterName) {
					continue
				}
				prefix := strings.TrimSuffix(name, clusterName)

				remove := false
				// TODO: Compute the actual set of names?
				if prefix == ".api" || prefix == ".api.internal" || prefix == ".bastion" {
					remove = true
				} else if strings.HasPrefix(prefix, ".etcd-") {
					remove = true
				}

				if !remove {
					continue
				}

				tracker := &ResourceTracker{
					Name:     aws.StringValue(rrs.Name),
					ID:       hostedZoneID + "/" + aws.StringValue(rrs.Name),
					Type:     "route53-record",
					groupKey: hostedZoneID,
					groupDeleter: func(cloud fi.Cloud, trackers []*ResourceTracker) error {
						return deleteRoute53Records(cloud, zone, trackers)
					},
					obj: rrs,
				}
				trackers = append(trackers, tracker)
			}
			return true
		})
		if err != nil {
			return nil, fmt.Errorf("error querying for route53 records for zone %q: %v", aws.StringValue(zone.Name), err)
		}
	}

	return trackers, nil
}

func DeleteIAMRole(cloud fi.Cloud, r *ResourceTracker) error {
	c := cloud.(awsup.AWSCloud)

	roleName := r.Name

	var policyNames []string
	{
		request := &iam.ListRolePoliciesInput{
			RoleName: aws.String(roleName),
		}
		err := c.IAM().ListRolePoliciesPages(request, func(page *iam.ListRolePoliciesOutput, lastPage bool) bool {
			for _, policy := range page.PolicyNames {
				policyNames = append(policyNames, aws.StringValue(policy))
			}
			return true
		})
		if err != nil {
			if awsup.AWSErrorCode(err) == "NoSuchEntity" {
				glog.V(2).Infof("Got NoSuchEntity describing IAM RolePolicy %q; will treat as already-deleted", roleName)
				return nil
			}

			return fmt.Errorf("error listing IAM role policies for %q: %v", roleName, err)
		}
	}

	for _, policyName := range policyNames {
		glog.V(2).Infof("Deleting IAM role policy %q %q", roleName, policyName)
		request := &iam.DeleteRolePolicyInput{
			RoleName:   aws.String(r.Name),
			PolicyName: aws.String(policyName),
		}
		_, err := c.IAM().DeleteRolePolicy(request)
		if err != nil {
			return fmt.Errorf("error deleting IAM role policy %q %q: %v", roleName, policyName, err)
		}
	}

	{
		glog.V(2).Infof("Deleting IAM role %q", r.Name)
		request := &iam.DeleteRoleInput{
			RoleName: aws.String(r.Name),
		}
		_, err := c.IAM().DeleteRole(request)
		if err != nil {
			return fmt.Errorf("error deleting IAM role %q: %v", r.Name, err)
		}
	}

	return nil
}

func ListIAMRoles(cloud fi.Cloud, clusterName string) ([]*ResourceTracker, error) {
	c := cloud.(awsup.AWSCloud)

	remove := make(map[string]bool)
	remove["masters."+clusterName] = true
	remove["nodes."+clusterName] = true
	remove["bastions."+clusterName] = true

	var roles []*iam.Role
	// Find roles matching remove map
	{
		request := &iam.ListRolesInput{}
		err := c.IAM().ListRolesPages(request, func(p *iam.ListRolesOutput, lastPage bool) bool {
			for _, r := range p.Roles {
				name := aws.StringValue(r.RoleName)
				if remove[name] {
					roles = append(roles, r)
				}
			}
			return true
		})
		if err != nil {
			return nil, fmt.Errorf("error listing IAM roles: %v", err)
		}
	}

	var trackers []*ResourceTracker

	for _, role := range roles {
		name := aws.StringValue(role.RoleName)
		tracker := &ResourceTracker{
			Name:    name,
			ID:      name,
			Type:    "iam-role",
			deleter: DeleteIAMRole,
		}
		trackers = append(trackers, tracker)
	}

	return trackers, nil
}

func DeleteIAMInstanceProfile(cloud fi.Cloud, r *ResourceTracker) error {
	c := cloud.(awsup.AWSCloud)

	profile := r.obj.(*iam.InstanceProfile)
	name := aws.StringValue(profile.InstanceProfileName)

	// Remove roles
	{
		for _, role := range profile.Roles {
			glog.V(2).Infof("Removing role %q from IAM instance profile %q", aws.StringValue(role.RoleName), name)
			request := &iam.RemoveRoleFromInstanceProfileInput{
				InstanceProfileName: profile.InstanceProfileName,
				RoleName:            role.RoleName,
			}
			_, err := c.IAM().RemoveRoleFromInstanceProfile(request)
			if err != nil {
				return fmt.Errorf("error removing role %q from IAM instance profile %q: %v", aws.StringValue(role.RoleName), name, err)
			}
		}
	}

	// Delete the instance profile
	{
		glog.V(2).Infof("Deleting IAM instance profile %q", name)
		request := &iam.DeleteInstanceProfileInput{
			InstanceProfileName: profile.InstanceProfileName,
		}
		_, err := c.IAM().DeleteInstanceProfile(request)
		if err != nil {
			return fmt.Errorf("error deleting IAM instance profile %q: %v", name, err)
		}
	}

	return nil
}

func ListIAMInstanceProfiles(cloud fi.Cloud, clusterName string) ([]*ResourceTracker, error) {
	c := cloud.(awsup.AWSCloud)

	remove := make(map[string]bool)
	remove["masters."+clusterName] = true
	remove["nodes."+clusterName] = true
	remove["bastions."+clusterName] = true

	var profiles []*iam.InstanceProfile

	request := &iam.ListInstanceProfilesInput{}
	err := c.IAM().ListInstanceProfilesPages(request, func(p *iam.ListInstanceProfilesOutput, lastPage bool) bool {
		for _, p := range p.InstanceProfiles {
			name := aws.StringValue(p.InstanceProfileName)
			if remove[name] {
				profiles = append(profiles, p)
			}
		}
		return true
	})
	if err != nil {
		return nil, fmt.Errorf("error listing IAM instance profiles: %v", err)
	}

	var trackers []*ResourceTracker

	for _, profile := range profiles {
		name := aws.StringValue(profile.InstanceProfileName)
		tracker := &ResourceTracker{
			Name:    name,
			ID:      name,
			Type:    "iam-instance-profile",
			deleter: DeleteIAMInstanceProfile,
			obj:     profile,
		}

		trackers = append(trackers, tracker)
	}

	return trackers, nil
}

func FindName(tags []*ec2.Tag) string {
	if name, found := awsup.FindEC2Tag(tags, "Name"); found {
		return name
	}
	return ""
}

func FindASGName(tags []*autoscaling.TagDescription) string {
	if name, found := awsup.FindASGTag(tags, "Name"); found {
		return name
	}
	return ""
}

func FindELBName(tags []*elb.Tag) string {
	if name, found := awsup.FindELBTag(tags, "Name"); found {
		return name
	}
	return ""
}

// HasSharedTag looks for the shared tag indicating that the cluster does not own the resource
func HasSharedTag(description string, tags []*ec2.Tag, clusterName string) bool {
	tagKey := "kubernetes.io/cluster/" + clusterName

	var found *ec2.Tag
	for _, tag := range tags {
		if aws.StringValue(tag.Key) != tagKey {
			continue
		}

		found = tag
	}

	if found == nil {
		glog.Warningf("(new) cluster tag not found on %s", description)
		return false
	}

	tagValue := aws.StringValue(found.Value)
	switch tagValue {
	case "owned":
		return false
	case "shared":
		return true

	default:
		glog.Warningf("unknown cluster tag on %s: %q=%q", description, tagKey, tagValue)
		return false
	}
}
