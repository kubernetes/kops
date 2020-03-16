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

package aws

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/route53"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog"
	"k8s.io/kops/pkg/dns"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/pkg/resources/spotinst"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

const (
	TypeAutoscalingLaunchConfig = "autoscaling-config"
	TypeNatGateway              = "nat-gateway"
	TypeElasticIp               = "elastic-ip"
	TypeLoadBalancer            = "load-balancer"
	TypeTargetGroup             = "target-group"
)

type listFn func(fi.Cloud, string) ([]*resources.Resource, error)

func ListResourcesAWS(cloud awsup.AWSCloud, clusterName string) (map[string]*resources.Resource, error) {
	resourceTrackers := make(map[string]*resources.Resource)

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
		ListELBV2s,
		ListTargetGroups,

		// Route 53
		ListRoute53Records,
		// IAM
		ListIAMInstanceProfiles,
		ListIAMRoles,
	}

	if featureflag.Spotinst.Enabled() {
		// Spotinst resources
		listFunctions = append(listFunctions, ListSpotinstResources)
	} else {
		// AutoScaling Groups
		listFunctions = append(listFunctions, ListAutoScalingGroups)
	}

	for _, fn := range listFunctions {
		rt, err := fn(cloud, clusterName)
		if err != nil {
			return nil, err
		}
		for _, t := range rt {
			resourceTrackers[t.Type+":"+t.ID] = t
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
				vpc := resourceTrackers["vpc:"+vpcID]
				if vpc != nil && resourceTrackers["internet-gateway:"+igwID] == nil {
					resourceTrackers["internet-gateway:"+igwID] = &resources.Resource{
						Name:    FindName(igw.Tags),
						ID:      igwID,
						Type:    "internet-gateway",
						Deleter: DeleteInternetGateway,
						Shared:  vpc.Shared, // Shared iff the VPC is shared
					}
				}
			}
		}
	}

	{
		// We delete a launch configuration if it is bound to one of the tagged security groups
		securityGroups := sets.NewString()
		for k := range resourceTrackers {
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
		lts, err := FindAutoScalingLaunchTemplateConfigurations(cloud, securityGroups)
		if err != nil {
			return nil, err
		}
		for _, t := range append(lcs, lts...) {
			resourceTrackers[t.Type+":"+t.ID] = t
		}
	}

	if err := addUntaggedRouteTables(cloud, clusterName, resourceTrackers); err != nil {
		return nil, err
	}

	{
		// We delete a NAT gateway if it is linked to our route table
		routeTableIds := make(map[string]*resources.Resource)
		for _, resource := range resourceTrackers {
			if resource.Type != ec2.ResourceTypeRouteTable {
				continue
			}
			id := resource.ID
			routeTableIds[id] = resource
		}
		natGateways, err := FindNatGateways(cloud, routeTableIds, clusterName)
		if err != nil {
			return nil, err
		}

		for _, t := range natGateways {
			resourceTrackers[t.Type+":"+t.ID] = t
		}
	}

	for k, t := range resourceTrackers {
		if t.Done {
			delete(resourceTrackers, k)
		}
	}
	return resourceTrackers, nil
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

func addUntaggedRouteTables(cloud awsup.AWSCloud, clusterName string, resources map[string]*resources.Resource) error {
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

		if resources["vpc:"+vpcID] == nil || resources["vpc:"+vpcID].Shared {
			// Not deleting this VPC; ignore
			continue
		}

		clusterTag, _ := awsup.FindEC2Tag(rt.Tags, awsup.TagClusterName)
		if clusterTag != "" && clusterTag != clusterName {
			klog.Infof("Skipping route table in VPC, but with wrong cluster tag (%q)", clusterTag)
			continue
		}

		isMain := false
		for _, a := range rt.Associations {
			if aws.BoolValue(a.Main) {
				isMain = true
			}
		}
		if isMain {
			klog.V(4).Infof("ignoring main routetable %q", rtID)
			continue
		}

		t := buildTrackerForRouteTable(rt, clusterName)
		if resources[t.Type+":"+t.ID] == nil {
			resources[t.Type+":"+t.ID] = t
		}
	}

	return nil
}

// FindAutoscalingLaunchConfiguration finds an AWS launch configuration given its name
func FindAutoscalingLaunchConfiguration(cloud awsup.AWSCloud, name string) (*autoscaling.LaunchConfiguration, error) {
	klog.V(2).Infof("Retrieving Autoscaling LaunchConfigurations %q", name)

	var results []*autoscaling.LaunchConfiguration

	request := &autoscaling.DescribeLaunchConfigurationsInput{
		LaunchConfigurationNames: []*string{&name},
	}
	err := cloud.Autoscaling().DescribeLaunchConfigurationsPages(request, func(p *autoscaling.DescribeLaunchConfigurationsOutput, lastPage bool) bool {
		results = append(results, p.LaunchConfigurations...)
		return true
	})
	if err != nil {
		return nil, fmt.Errorf("error listing autoscaling LaunchConfigurations: %v", err)
	}

	if len(results) == 0 {
		return nil, nil
	}
	if len(results) != 1 {
		return nil, fmt.Errorf("found multiple LaunchConfigurations with name %q", name)
	}
	return results[0], nil
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

func matchesElbV2Tags(tags map[string]string, actual []*elbv2.Tag) bool {
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

func DeleteInstance(cloud fi.Cloud, t *resources.Resource) error {
	c := cloud.(awsup.AWSCloud)

	id := t.ID
	klog.V(2).Infof("Deleting EC2 instance %q", id)
	request := &ec2.TerminateInstancesInput{
		InstanceIds: []*string{&id},
	}
	_, err := c.EC2().TerminateInstances(request)
	if err != nil {
		if awsup.AWSErrorCode(err) == "InvalidInstanceID.NotFound" {
			klog.V(2).Infof("Got InvalidInstanceID.NotFound error deleting instance %q; will treat as already-deleted", id)
		} else {
			return fmt.Errorf("error deleting Instance %q: %v", id, err)
		}
	}
	return nil
}

func DeleteCloudFormationStack(cloud fi.Cloud, t *resources.Resource) error {
	c := cloud.(awsup.AWSCloud)

	id := t.ID
	klog.V(2).Infof("deleting CloudFormation stack %q %q", t.Name, id)

	request := &cloudformation.DeleteStackInput{}
	request.StackName = &t.Name

	_, err := c.CloudFormation().DeleteStack(request)
	if err != nil {
		return fmt.Errorf("error deleting CloudFormation stack %q: %v", id, err)
	}
	return nil
}

func DumpCloudFormationStack(op *resources.DumpOperation, r *resources.Resource) error {
	data := make(map[string]interface{})
	data["id"] = r.ID
	data["type"] = r.Type
	data["raw"] = r.Obj
	op.Dump.Resources = append(op.Dump.Resources, data)
	return nil
}

func ListCloudFormationStacks(cloud fi.Cloud, clusterName string) ([]*resources.Resource, error) {
	var resourceTrackers []*resources.Resource
	request := &cloudformation.ListStacksInput{}
	c := cloud.(awsup.AWSCloud)
	response, err := c.CloudFormation().ListStacks(request)
	if err != nil {
		return nil, fmt.Errorf("unable to list CloudFormation stacks: %v", err)
	}
	for _, stack := range response.StackSummaries {
		if *stack.StackName == clusterName {
			resourceTracker := &resources.Resource{
				Name:    *stack.StackName,
				ID:      *stack.StackId,
				Type:    "cloud-formation",
				Deleter: DeleteCloudFormationStack,
				Dumper:  DumpCloudFormationStack,
				Obj:     stack,
			}
			resourceTrackers = append(resourceTrackers, resourceTracker)
		}
	}

	return resourceTrackers, nil
}

func ListInstances(cloud fi.Cloud, clusterName string) ([]*resources.Resource, error) {
	c := cloud.(awsup.AWSCloud)

	klog.V(2).Infof("Querying EC2 instances")
	request := &ec2.DescribeInstancesInput{
		Filters: BuildEC2Filters(cloud),
	}

	var resourceTrackers []*resources.Resource

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
						klog.V(4).Infof("instance %q has state=%q", id, stateName)

					default:
						klog.Infof("unknown instance state for %q: %q", id, stateName)
					}
				}

				resourceTracker := &resources.Resource{
					Name:    FindName(instance.Tags),
					ID:      id,
					Type:    ec2.ResourceTypeInstance,
					Deleter: DeleteInstance,
					Dumper:  DumpInstance,
					Obj:     instance,
				}

				var blocks []string
				blocks = append(blocks, "subnet:"+aws.StringValue(instance.SubnetId))
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

				resourceTracker.Blocks = blocks

				resourceTrackers = append(resourceTrackers, resourceTracker)

			}
		}
		return true
	})
	if err != nil {
		return nil, fmt.Errorf("error describing instances: %v", err)
	}

	return resourceTrackers, nil
}

// getDumpState gets the dumpState from the dump context, or creates one if not yet initialized
func getDumpState(dumpContext *resources.DumpOperation) *dumpState {
	if dumpContext.CloudState == nil {
		dumpContext.CloudState = &dumpState{
			cloud: dumpContext.Cloud.(awsup.AWSCloud),
		}
	}
	return dumpContext.CloudState.(*dumpState)
}

type imageInfo struct {
	SSHUser string
}

type dumpState struct {
	cloud  awsup.AWSCloud
	mutex  sync.Mutex
	images map[string]*imageInfo
}

func (s *dumpState) getImageInfo(imageID string) (*imageInfo, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.images == nil {
		s.images = make(map[string]*imageInfo)
	}

	info := s.images[imageID]
	if info == nil {
		image, err := s.cloud.ResolveImage(imageID)
		if err != nil {
			return nil, err
		}
		info = &imageInfo{}

		if image != nil {
			sshUser := guessSSHUser(image)
			if sshUser == "" {
				klog.Warningf("unable to guess SSH user for image: %+v", image)
			}
			info.SSHUser = sshUser
		}

		s.images[imageID] = info
	}

	return info, nil
}

func guessSSHUser(image *ec2.Image) string {
	owner := aws.StringValue(image.OwnerId)
	switch owner {
	case awsup.WellKnownAccountAmazonLinux2, awsup.WellKnownAccountRedhat:
		return "ec2-user"
	case awsup.WellKnownAccountCentOS:
		return "centos"
	case awsup.WellKnownAccountDebian9, awsup.WellKnownAccountDebian10, awsup.WellKnownAccountKopeio:
		return "admin"
	case awsup.WellKnownAccountUbuntu:
		return "ubuntu"
	case awsup.WellKnownAccountCoreOS, awsup.WellKnownAccountFlatcar:
		return "core"
	}

	name := aws.StringValue(image.Name)
	name = strings.ToLower(name)
	if strings.HasPrefix(name, "centos") {
		// We could check the marketplace id, but this is just a guess anyway...
		return "centos"
	}

	return ""
}

func DumpInstance(op *resources.DumpOperation, r *resources.Resource) error {
	data := make(map[string]interface{})
	data["id"] = r.ID
	data["type"] = ec2.ResourceTypeInstance
	data["raw"] = r.Obj
	op.Dump.Resources = append(op.Dump.Resources, data)

	ec2Instance := r.Obj.(*ec2.Instance)
	i := &resources.Instance{
		Name: r.ID,
	}
	for _, networkInterface := range ec2Instance.NetworkInterfaces {
		if networkInterface.Association != nil {
			publicIP := aws.StringValue(networkInterface.Association.PublicIp)
			if publicIP != "" {
				i.PublicAddresses = append(i.PublicAddresses, publicIP)
			}
		}
	}
	for _, tag := range ec2Instance.Tags {
		key := aws.StringValue(tag.Key)
		if !strings.HasPrefix(key, awsup.TagNameRolePrefix) {
			continue
		}
		role := strings.TrimPrefix(key, awsup.TagNameRolePrefix)
		i.Roles = append(i.Roles, role)
	}

	imageID := aws.StringValue(ec2Instance.ImageId)
	imageInfo, err := getDumpState(op).getImageInfo(imageID)
	if err != nil {
		klog.Warningf("unable to fetch image %q: %v", imageID, err)
	} else if imageInfo != nil {
		i.SSHUser = imageInfo.SSHUser
	}

	op.Dump.Instances = append(op.Dump.Instances, i)

	return nil
}

func DeleteVolume(cloud fi.Cloud, r *resources.Resource) error {
	c := cloud.(awsup.AWSCloud)

	id := r.ID

	klog.V(2).Infof("Deleting EC2 Volume %q", id)
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

func ListVolumes(cloud fi.Cloud, clusterName string) ([]*resources.Resource, error) {
	c := cloud.(awsup.AWSCloud)

	volumes, err := DescribeVolumes(cloud)
	if err != nil {
		return nil, err
	}
	var resourceTrackers []*resources.Resource

	elasticIPs := make(map[string]bool)
	for _, volume := range volumes {
		id := aws.StringValue(volume.VolumeId)

		resourceTracker := &resources.Resource{
			Name:    FindName(volume.Tags),
			ID:      id,
			Type:    "volume",
			Deleter: DeleteVolume,
			Shared:  HasSharedTag(ec2.ResourceTypeVolume+":"+id, volume.Tags, clusterName),
		}

		var blocks []string
		//blocks = append(blocks, "vpc:" + aws.StringValue(rt.VpcId))

		resourceTracker.Blocks = blocks

		resourceTrackers = append(resourceTrackers, resourceTracker)

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
		klog.V(2).Infof("Querying EC2 Elastic IPs")
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

			resourceTrackers = append(resourceTrackers, buildElasticIPResource(address, false, clusterName))
		}
	}

	return resourceTrackers, nil
}

func DescribeVolumes(cloud fi.Cloud) ([]*ec2.Volume, error) {
	c := cloud.(awsup.AWSCloud)

	var volumes []*ec2.Volume

	klog.V(2).Infof("Listing EC2 Volumes")
	request := &ec2.DescribeVolumesInput{
		Filters: BuildEC2Filters(c),
	}

	err := c.EC2().DescribeVolumesPages(request, func(p *ec2.DescribeVolumesOutput, lastPage bool) bool {
		volumes = append(volumes, p.Volumes...)
		return true
	})
	if err != nil {
		return nil, fmt.Errorf("error describing volumes: %v", err)
	}

	return volumes, nil
}

func DeleteKeypair(cloud fi.Cloud, r *resources.Resource) error {
	c := cloud.(awsup.AWSCloud)

	name := r.Name

	klog.V(2).Infof("Deleting EC2 Keypair %q", name)
	request := &ec2.DeleteKeyPairInput{
		KeyName: &name,
	}
	_, err := c.EC2().DeleteKeyPair(request)
	if err != nil {
		return fmt.Errorf("error deleting KeyPair %q: %v", name, err)
	}
	return nil
}

func ListKeypairs(cloud fi.Cloud, clusterName string) ([]*resources.Resource, error) {
	if !strings.Contains(clusterName, ".") {
		klog.Infof("cluster %q is legacy (kube-up) cluster; won't delete keypairs", clusterName)
		return nil, nil
	}

	c := cloud.(awsup.AWSCloud)

	keypairName := "kubernetes." + clusterName

	klog.V(2).Infof("Listing EC2 Keypairs")

	// TODO: We need to match both the name and a prefix
	// TODO: usee 'Filters: []*ec2.Filter{awsup.NewEC2Filter("key-name", keypairName)},'
	request := &ec2.DescribeKeyPairsInput{}
	response, err := c.EC2().DescribeKeyPairs(request)
	if err != nil {
		return nil, fmt.Errorf("error listing KeyPairs: %v", err)
	}

	var resourceTrackers []*resources.Resource

	for _, keypair := range response.KeyPairs {
		name := aws.StringValue(keypair.KeyName)
		if name != keypairName && !strings.HasPrefix(name, keypairName+"-") {
			continue
		}
		resourceTracker := &resources.Resource{
			Name:    name,
			ID:      name,
			Type:    "keypair",
			Deleter: DeleteKeypair,
		}

		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}

func DeleteSubnet(cloud fi.Cloud, tracker *resources.Resource) error {
	c := cloud.(awsup.AWSCloud)

	id := tracker.ID

	klog.V(2).Infof("Deleting EC2 Subnet %q", id)
	request := &ec2.DeleteSubnetInput{
		SubnetId: &id,
	}
	_, err := c.EC2().DeleteSubnet(request)
	if err != nil {
		if awsup.AWSErrorCode(err) == "InvalidSubnetID.NotFound" {
			klog.V(2).Infof("Got InvalidSubnetID.NotFound error deleting subnet %q; will treat as already-deleted", id)
			return nil
		} else if IsDependencyViolation(err) {
			return err
		}
		return fmt.Errorf("error deleting Subnet %q: %v", id, err)
	}
	return nil
}

func ListSubnets(cloud fi.Cloud, clusterName string) ([]*resources.Resource, error) {
	c := cloud.(awsup.AWSCloud)
	subnets, err := DescribeSubnets(cloud)
	if err != nil {
		return nil, fmt.Errorf("error listing subnets: %v", err)
	}

	var resourceTrackers []*resources.Resource
	elasticIPs := sets.NewString()
	ownedElasticIPs := sets.NewString()
	natGatewayIds := sets.NewString()
	ownedNatGatewayIds := sets.NewString()
	for _, subnet := range subnets {
		subnetID := aws.StringValue(subnet.SubnetId)

		shared := HasSharedTag("subnet:"+subnetID, subnet.Tags, clusterName)
		resourceTracker := &resources.Resource{
			Name:    FindName(subnet.Tags),
			ID:      subnetID,
			Type:    ec2.ResourceTypeSubnet,
			Deleter: DeleteSubnet,
			Dumper:  DumpSubnet,
			Shared:  shared,
			Obj:     subnet,
		}
		resourceTracker.Blocks = append(resourceTracker.Blocks, "vpc:"+aws.StringValue(subnet.VpcId))
		resourceTrackers = append(resourceTrackers, resourceTracker)

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
		klog.V(2).Infof("Querying EC2 Elastic IPs")
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
			resourceTrackers = append(resourceTrackers, buildElasticIPResource(address, ownedElasticIPs.Has(ip), clusterName))
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

		klog.V(2).Infof("Querying Nat Gateways")
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

			forceShared := sharedNgwIds.Has(id) || !ownedNatGatewayIds.Has(id)
			r := buildNatGatewayResource(ngw, forceShared, clusterName)
			resourceTrackers = append(resourceTrackers, r)
		}
	}

	return resourceTrackers, nil
}

func DescribeSubnets(cloud fi.Cloud) ([]*ec2.Subnet, error) {
	c := cloud.(awsup.AWSCloud)

	klog.V(2).Infof("Listing EC2 subnets")
	request := &ec2.DescribeSubnetsInput{
		Filters: BuildEC2Filters(cloud),
	}
	response, err := c.EC2().DescribeSubnets(request)
	if err != nil {
		return nil, fmt.Errorf("error listing subnets: %v", err)
	}

	return response.Subnets, nil
}

func DeleteRouteTable(cloud fi.Cloud, r *resources.Resource) error {
	c := cloud.(awsup.AWSCloud)

	id := r.ID

	klog.V(2).Infof("Deleting EC2 RouteTable %q", id)
	request := &ec2.DeleteRouteTableInput{
		RouteTableId: &id,
	}
	_, err := c.EC2().DeleteRouteTable(request)
	if err != nil {
		if awsup.AWSErrorCode(err) == "InvalidRouteTableID.NotFound" {
			klog.V(2).Infof("Got InvalidRouteTableID.NotFound error describing RouteTable %q; will treat as already-deleted", id)
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

	klog.V(2).Infof("Listing all RouteTables")
	request := &ec2.DescribeRouteTablesInput{}
	response, err := c.EC2().DescribeRouteTables(request)
	if err != nil {
		return nil, fmt.Errorf("error listing RouteTables: %v", err)
	}

	return response.RouteTables, nil
}

func DeleteDhcpOptions(cloud fi.Cloud, r *resources.Resource) error {
	c := cloud.(awsup.AWSCloud)

	id := r.ID

	klog.V(2).Infof("Deleting EC2 DhcpOptions %q", id)
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

func ListDhcpOptions(cloud fi.Cloud, clusterName string) ([]*resources.Resource, error) {
	dhcpOptions, err := DescribeDhcpOptions(cloud)
	if err != nil {
		return nil, err
	}

	var resourceTrackers []*resources.Resource

	for _, o := range dhcpOptions {
		resourceTracker := &resources.Resource{
			Name:    FindName(o.Tags),
			ID:      aws.StringValue(o.DhcpOptionsId),
			Type:    "dhcp-options",
			Deleter: DeleteDhcpOptions,
			Shared:  HasSharedTag(ec2.ResourceTypeDhcpOptions+":"+aws.StringValue(o.DhcpOptionsId), o.Tags, clusterName),
		}

		var blocks []string

		resourceTracker.Blocks = blocks

		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}

func DescribeDhcpOptions(cloud fi.Cloud) ([]*ec2.DhcpOptions, error) {
	c := cloud.(awsup.AWSCloud)

	klog.V(2).Infof("Listing EC2 DhcpOptions")
	request := &ec2.DescribeDhcpOptionsInput{
		Filters: BuildEC2Filters(cloud),
	}
	response, err := c.EC2().DescribeDhcpOptions(request)
	if err != nil {
		return nil, fmt.Errorf("error listing DhcpOptions: %v", err)
	}

	return response.DhcpOptions, nil
}

func DeleteInternetGateway(cloud fi.Cloud, r *resources.Resource) error {
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
				klog.Infof("Internet gateway %q not found; assuming already deleted", id)
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
		klog.V(2).Infof("Detaching EC2 InternetGateway %q", id)
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
		klog.V(2).Infof("Deleting EC2 InternetGateway %q", id)
		request := &ec2.DeleteInternetGatewayInput{
			InternetGatewayId: &id,
		}
		_, err := c.EC2().DeleteInternetGateway(request)
		if err != nil {
			if IsDependencyViolation(err) {
				return err
			}
			if awsup.AWSErrorCode(err) == "InvalidInternetGatewayID.NotFound" {
				klog.Infof("Internet gateway %q not found; assuming already deleted", id)
				return nil
			}
			return fmt.Errorf("error deleting InternetGateway %q: %v", id, err)
		}
	}

	return nil
}

func ListInternetGateways(cloud fi.Cloud, clusterName string) ([]*resources.Resource, error) {
	gateways, err := DescribeInternetGateways(cloud)
	if err != nil {
		return nil, err
	}

	var resourceTrackers []*resources.Resource

	for _, o := range gateways {
		resourceTracker := &resources.Resource{
			Name:    FindName(o.Tags),
			ID:      aws.StringValue(o.InternetGatewayId),
			Type:    "internet-gateway",
			Deleter: DeleteInternetGateway,
			Shared:  HasSharedTag(ec2.ResourceTypeInternetGateway+":"+aws.StringValue(o.InternetGatewayId), o.Tags, clusterName),
		}

		var blocks []string
		for _, a := range o.Attachments {
			if aws.StringValue(a.VpcId) != "" {
				blocks = append(blocks, "vpc:"+aws.StringValue(a.VpcId))
			}
		}
		resourceTracker.Blocks = blocks

		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}

func DescribeInternetGateways(cloud fi.Cloud) ([]*ec2.InternetGateway, error) {
	c := cloud.(awsup.AWSCloud)

	klog.V(2).Infof("Listing EC2 InternetGateways")
	request := &ec2.DescribeInternetGatewaysInput{
		Filters: BuildEC2Filters(cloud),
	}
	response, err := c.EC2().DescribeInternetGateways(request)
	if err != nil {
		return nil, fmt.Errorf("error listing InternetGateway: %v", err)
	}

	var gateways []*ec2.InternetGateway
	gateways = append(gateways, response.InternetGateways...)

	return gateways, nil
}

// DescribeInternetGatewaysIgnoreTags returns all ec2.InternetGateways, ignoring tags
// (gateways were not always tagged in kube-up)
func DescribeInternetGatewaysIgnoreTags(cloud fi.Cloud) ([]*ec2.InternetGateway, error) {
	c := cloud.(awsup.AWSCloud)

	klog.V(2).Infof("Listing all Internet Gateways")

	request := &ec2.DescribeInternetGatewaysInput{}
	response, err := c.EC2().DescribeInternetGateways(request)
	if err != nil {
		return nil, fmt.Errorf("error listing (all) InternetGateways: %v", err)
	}

	var gateways []*ec2.InternetGateway

	gateways = append(gateways, response.InternetGateways...)

	return gateways, nil
}

func DeleteAutoScalingGroup(cloud fi.Cloud, r *resources.Resource) error {
	c := cloud.(awsup.AWSCloud)

	id := r.ID

	klog.V(2).Infof("Deleting autoscaling group %q", id)
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

func ListAutoScalingGroups(cloud fi.Cloud, clusterName string) ([]*resources.Resource, error) {
	c := cloud.(awsup.AWSCloud)

	tags := c.Tags()

	asgs, err := awsup.FindAutoscalingGroups(c, tags)
	if err != nil {
		return nil, err
	}

	var resourceTrackers []*resources.Resource

	for _, asg := range asgs {
		resourceTracker := &resources.Resource{
			Name:    FindASGName(asg.Tags),
			ID:      aws.StringValue(asg.AutoScalingGroupName),
			Type:    "autoscaling-group",
			Deleter: DeleteAutoScalingGroup,
		}

		var blocks []string
		subnets := aws.StringValue(asg.VPCZoneIdentifier)
		for _, subnet := range strings.Split(subnets, ",") {
			if subnet == "" {
				continue
			}
			blocks = append(blocks, "subnet:"+subnet)
		}
		if asg.LaunchConfigurationName != nil {
			blocks = append(blocks, TypeAutoscalingLaunchConfig+":"+aws.StringValue(asg.LaunchConfigurationName))
		}
		if asg.LaunchTemplate != nil {
			blocks = append(blocks, TypeAutoscalingLaunchConfig+":"+aws.StringValue(asg.LaunchTemplate.LaunchTemplateName))
		}

		resourceTracker.Blocks = blocks

		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}

// FindAutoScalingLaunchTemplateConfigurations finds any launch configurations which reference the security groups
func FindAutoScalingLaunchTemplateConfigurations(cloud fi.Cloud, securityGroups sets.String) ([]*resources.Resource, error) {
	var list []*resources.Resource

	c, ok := cloud.(awsup.AWSCloud)
	if !ok {
		return nil, errors.New("expected a aws cloud provider")
	}
	klog.V(2).Infof("Finding all Autoscaling LaunchTemplates associated to security groups")

	resp, err := c.EC2().DescribeLaunchTemplates(&ec2.DescribeLaunchTemplatesInput{MaxResults: fi.Int64(100)})
	if err != nil {
		return list, nil
	}

	for _, x := range resp.LaunchTemplates {
		// @step: grab the actual launch template
		req, err := c.EC2().DescribeLaunchTemplateVersions(&ec2.DescribeLaunchTemplateVersionsInput{
			LaunchTemplateName: x.LaunchTemplateName,
		})
		if err != nil {
			return list, err
		}
		for _, j := range req.LaunchTemplateVersions {
			// @check if the security group references the security group above
			var s []*string
			for _, ni := range j.LaunchTemplateData.NetworkInterfaces {
				s = append(s, ni.Groups...)
			}
			s = append(s, j.LaunchTemplateData.SecurityGroupIds...)
			for _, y := range s {
				if securityGroups.Has(fi.StringValue(y)) {
					list = append(list, &resources.Resource{
						Name:    aws.StringValue(x.LaunchTemplateName),
						ID:      aws.StringValue(x.LaunchTemplateName),
						Type:    TypeAutoscalingLaunchConfig,
						Deleter: DeleteAutoScalingGroupLaunchTemplate,
					})
				}
			}
		}
	}

	return list, nil
}

// FindAutoScalingLaunchConfigurations finds all launch configurations which has a reference to the security groups
func FindAutoScalingLaunchConfigurations(cloud fi.Cloud, securityGroups sets.String) ([]*resources.Resource, error) {
	c := cloud.(awsup.AWSCloud)

	klog.V(2).Infof("Finding all Autoscaling LaunchConfigurations by security group")
	var resourceTrackers []*resources.Resource

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

			resourceTracker := &resources.Resource{
				Name:    aws.StringValue(t.LaunchConfigurationName),
				ID:      aws.StringValue(t.LaunchConfigurationName),
				Type:    TypeAutoscalingLaunchConfig,
				Deleter: DeleteAutoscalingLaunchConfiguration,
			}

			var blocks []string
			//blocks = append(blocks, TypeAutoscalingLaunchConfig + ":" + aws.StringValue(asg.LaunchConfigurationName))

			resourceTracker.Blocks = blocks

			resourceTrackers = append(resourceTrackers, resourceTracker)
		}
		return true
	})
	if err != nil {
		return nil, fmt.Errorf("error listing autoscaling LaunchConfigurations: %v", err)
	}

	return resourceTrackers, nil
}

func FindNatGateways(cloud fi.Cloud, routeTables map[string]*resources.Resource, clusterName string) ([]*resources.Resource, error) {
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
				klog.Warningf("unable to find resource for route table %s", routeTableID)
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

	var resourceTrackers []*resources.Resource
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

		for _, ngw := range response.NatGateways {
			natGatewayId := aws.StringValue(ngw.NatGatewayId)

			forceShared := !ownedNatGatewayIds.Has(natGatewayId)
			ngwResource := buildNatGatewayResource(ngw, forceShared, clusterName)
			resourceTrackers = append(resourceTrackers, ngwResource)

			// Don't try to remove ElasticIPs if NatGateway is shared
			if ngwResource.Shared {
				continue
			}

			// If we're deleting the NatGateway, we should delete the ElasticIP also
			for _, address := range ngw.NatGatewayAddresses {
				if address.AllocationId != nil {
					request := &ec2.DescribeAddressesInput{}
					request.AllocationIds = []*string{address.AllocationId}
					response, err := c.EC2().DescribeAddresses(request)
					if err != nil {
						return nil, fmt.Errorf("error from DescribeAddresses: %v", err)
					}

					for _, eip := range response.Addresses {
						eipTracker := buildElasticIPResource(eip, !ownedNatGatewayIds.Has(natGatewayId), clusterName)
						resourceTrackers = append(resourceTrackers, eipTracker)
					}
				}
			}
		}
	}

	return resourceTrackers, nil
}

// DeleteAutoScalingGroupLaunchTemplate deletes
func DeleteAutoScalingGroupLaunchTemplate(cloud fi.Cloud, r *resources.Resource) error {
	c, ok := cloud.(awsup.AWSCloud)
	if !ok {
		return errors.New("expected a aws.Cloud provider")
	}
	klog.V(2).Infof("Deleting EC2 LaunchTemplate %q", r.ID)

	if _, err := c.EC2().DeleteLaunchTemplate(&ec2.DeleteLaunchTemplateInput{
		LaunchTemplateName: fi.String(r.ID),
	}); err != nil {
		return fmt.Errorf("error deleting ec2 LaunchTemplate %q: %v", r.ID, err)
	}

	return nil
}

func DeleteAutoscalingLaunchConfiguration(cloud fi.Cloud, r *resources.Resource) error {
	c := cloud.(awsup.AWSCloud)

	id := r.ID
	klog.V(2).Infof("Deleting autoscaling LaunchConfiguration %q", id)
	request := &autoscaling.DeleteLaunchConfigurationInput{
		LaunchConfigurationName: &id,
	}
	_, err := c.Autoscaling().DeleteLaunchConfiguration(request)
	if err != nil {
		return fmt.Errorf("error deleting autoscaling LaunchConfiguration %q: %v", id, err)
	}
	return nil
}

func DeleteELB(cloud fi.Cloud, r *resources.Resource) error {
	c := cloud.(awsup.AWSCloud)

	id := r.ID

	klog.V(2).Infof("Deleting ELB %q", id)
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

func DeleteELBV2(cloud fi.Cloud, r *resources.Resource) error {
	c := cloud.(awsup.AWSCloud)
	id := r.ID

	klog.V(2).Infof("Deleting ELBV2 %q", id)
	request := &elbv2.DeleteLoadBalancerInput{
		LoadBalancerArn: aws.String(id),
	}
	_, err := c.ELBV2().DeleteLoadBalancer(request)
	if err != nil {
		if IsDependencyViolation(err) {
			return err
		}
		return fmt.Errorf("error deleting V2 LoadBalancer %q: %v", id, err)
	}
	return nil
}

func DeleteTargetGroup(cloud fi.Cloud, r *resources.Resource) error {
	c := cloud.(awsup.AWSCloud)
	id := r.ID

	klog.V(2).Infof("Deleting TargetGroup %q", id)
	request := &elbv2.DeleteTargetGroupInput{
		TargetGroupArn: aws.String(id),
	}
	_, err := c.ELBV2().DeleteTargetGroup(request)
	if err != nil {
		if IsDependencyViolation(err) {
			return err
		}
		return fmt.Errorf("error deleting TargetGroup %q: %v", id, err)
	}
	return nil
}

func DumpELB(op *resources.DumpOperation, r *resources.Resource) error {
	data := make(map[string]interface{})
	data["id"] = r.ID
	data["type"] = TypeLoadBalancer
	data["raw"] = r.Obj
	op.Dump.Resources = append(op.Dump.Resources, data)
	return nil
}

func ListELBs(cloud fi.Cloud, clusterName string) ([]*resources.Resource, error) {
	elbs, elbTags, err := DescribeELBs(cloud)
	if err != nil {
		return nil, err
	}

	var resourceTrackers []*resources.Resource
	for _, elb := range elbs {
		id := aws.StringValue(elb.LoadBalancerName)
		resourceTracker := &resources.Resource{
			Name:    FindELBName(elbTags[id]),
			ID:      id,
			Type:    TypeLoadBalancer,
			Deleter: DeleteELB,
			Dumper:  DumpELB,
			Obj:     elb,
		}

		var blocks []string
		for _, sg := range elb.SecurityGroups {
			blocks = append(blocks, "security-group:"+aws.StringValue(sg))
		}
		for _, s := range elb.Subnets {
			blocks = append(blocks, "subnet:"+aws.StringValue(s))
		}
		blocks = append(blocks, "vpc:"+aws.StringValue(elb.VPCId))

		resourceTracker.Blocks = blocks

		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}

func DescribeELBs(cloud fi.Cloud) ([]*elb.LoadBalancerDescription, map[string][]*elb.Tag, error) {
	c := cloud.(awsup.AWSCloud)
	tags := c.Tags()

	klog.V(2).Infof("Listing all ELBs")

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

// For NLBs and ALBs
func ListELBV2s(cloud fi.Cloud, clusterName string) ([]*resources.Resource, error) {
	elbv2s, _, err := DescribeELBV2s(cloud)
	if err != nil {
		return nil, err
	}

	var resourceTrackers []*resources.Resource
	for _, elb := range elbv2s {
		id := aws.StringValue(elb.LoadBalancerName)
		resourceTracker := &resources.Resource{
			Name:    id,
			ID:      string(*elb.LoadBalancerArn),
			Type:    TypeLoadBalancer,
			Deleter: DeleteELBV2,
			Dumper:  DumpELB,
			Obj:     elb,
		}

		var blocks []string
		for _, sg := range elb.SecurityGroups {
			blocks = append(blocks, "security-group:"+aws.StringValue(sg))
		}

		blocks = append(blocks, "vpc:"+aws.StringValue(elb.VpcId))

		resourceTracker.Blocks = blocks

		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}

func DescribeELBV2s(cloud fi.Cloud) ([]*elbv2.LoadBalancer, map[string][]*elbv2.Tag, error) {
	c := cloud.(awsup.AWSCloud)
	tags := c.Tags()

	klog.V(2).Infof("Listing all NLBs and ALBs")

	request := &elbv2.DescribeLoadBalancersInput{}
	// ELBV2 DescribeTags has a limit of 20 names, so we set the page size here to 20 also
	request.PageSize = aws.Int64(20)

	var elbv2s []*elbv2.LoadBalancer
	elbv2Tags := make(map[string][]*elbv2.Tag)

	var innerError error
	err := c.ELBV2().DescribeLoadBalancersPages(request, func(p *elbv2.DescribeLoadBalancersOutput, lastPage bool) bool {
		if len(p.LoadBalancers) == 0 {
			return true
		}

		tagRequest := &elbv2.DescribeTagsInput{}

		nameToELB := make(map[string]*elbv2.LoadBalancer)
		for _, elb := range p.LoadBalancers {
			name := aws.StringValue(elb.LoadBalancerArn)
			nameToELB[name] = elb

			tagRequest.ResourceArns = append(tagRequest.ResourceArns, elb.LoadBalancerArn)
		}

		tagResponse, err := c.ELBV2().DescribeTags(tagRequest)
		if err != nil {
			innerError = fmt.Errorf("error listing elb Tags: %v", err)
			return false
		}

		for _, t := range tagResponse.TagDescriptions {

			elbARN := aws.StringValue(t.ResourceArn)
			if !matchesElbV2Tags(tags, t.Tags) {
				continue
			}

			elbv2Tags[elbARN] = t.Tags
			elb := nameToELB[elbARN]
			elbv2s = append(elbv2s, elb)
		}

		return true
	})
	if err != nil {
		return nil, nil, fmt.Errorf("error describing LoadBalancers: %v", err)
	}
	if innerError != nil {
		return nil, nil, fmt.Errorf("error describing LoadBalancers: %v", innerError)
	}

	return elbv2s, elbv2Tags, nil
}

func ListTargetGroups(cloud fi.Cloud, clusterName string) ([]*resources.Resource, error) {
	targetgroups, _, err := DescribeTargetGroups(cloud)
	if err != nil {
		return nil, err
	}

	var resourceTrackers []*resources.Resource
	for _, tg := range targetgroups {
		id := aws.StringValue(tg.TargetGroupName)
		resourceTracker := &resources.Resource{
			Name:    id,
			ID:      string(*tg.TargetGroupArn),
			Type:    TypeTargetGroup,
			Deleter: DeleteTargetGroup,
			Dumper:  DumpELB,
			Obj:     tg,
		}

		resourceTrackers = append(resourceTrackers, resourceTracker)
	}
	return resourceTrackers, nil
}

func DescribeTargetGroups(cloud fi.Cloud) ([]*elbv2.TargetGroup, map[string][]*elbv2.Tag, error) {
	c := cloud.(awsup.AWSCloud)
	tags := c.Tags()

	klog.V(2).Infof("Listing all TargetGroups")

	request := &elbv2.DescribeTargetGroupsInput{}
	// DescribeTags has a limit of 20 names, so we set the page size here to 20 also
	request.PageSize = aws.Int64(20)

	var targetgroups []*elbv2.TargetGroup
	targetgroupTags := make(map[string][]*elbv2.Tag)

	var innerError error
	err := c.ELBV2().DescribeTargetGroupsPages(request, func(p *elbv2.DescribeTargetGroupsOutput, lastPage bool) bool {
		if len(p.TargetGroups) == 0 {
			return true
		}

		tagRequest := &elbv2.DescribeTagsInput{}

		nameToTargetGroup := make(map[string]*elbv2.TargetGroup)
		for _, tg := range p.TargetGroups {
			name := aws.StringValue(tg.TargetGroupArn)
			nameToTargetGroup[name] = tg

			tagRequest.ResourceArns = append(tagRequest.ResourceArns, tg.TargetGroupArn)
		}

		tagResponse, err := c.ELBV2().DescribeTags(tagRequest)
		if err != nil {
			innerError = fmt.Errorf("error listing TargetGroup Tags: %v", err)
			return false
		}

		for _, t := range tagResponse.TagDescriptions {
			tgARN := aws.StringValue(t.ResourceArn)
			if !matchesElbV2Tags(tags, t.Tags) {
				continue
			}
			targetgroupTags[tgARN] = t.Tags

			tg := nameToTargetGroup[tgARN]
			targetgroups = append(targetgroups, tg)
		}

		return true
	})
	if err != nil {
		return nil, nil, fmt.Errorf("error describing TargetGroups: %v", err)
	}
	if innerError != nil {
		return nil, nil, fmt.Errorf("error describing TargetGroups: %v", innerError)
	}

	return targetgroups, targetgroupTags, nil
}

func DeleteElasticIP(cloud fi.Cloud, t *resources.Resource) error {
	c := cloud.(awsup.AWSCloud)

	id := t.ID

	klog.V(2).Infof("Releasing IP %s", t.Name)
	request := &ec2.ReleaseAddressInput{
		AllocationId: &id,
	}
	_, err := c.EC2().ReleaseAddress(request)
	if err != nil {
		if awsup.AWSErrorCode(err) == "InvalidAllocationID.NotFound" {
			klog.V(2).Infof("Got InvalidAllocationID.NotFound error describing ElasticIP %q; will treat as already-deleted", id)
			return nil
		}

		if IsDependencyViolation(err) {
			return err
		}
		return fmt.Errorf("error deleting elastic ip %q: %v", t.Name, err)
	}
	return nil
}

func DeleteNatGateway(cloud fi.Cloud, t *resources.Resource) error {
	c := cloud.(awsup.AWSCloud)

	id := t.ID

	klog.V(2).Infof("Removing NatGateway %s", t.Name)
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

func deleteRoute53Records(cloud fi.Cloud, zone *route53.HostedZone, resourceTrackers []*resources.Resource) error {
	c := cloud.(awsup.AWSCloud)

	var changes []*route53.Change
	var names []string
	for _, resourceTracker := range resourceTrackers {
		names = append(names, resourceTracker.Name)
		changes = append(changes, &route53.Change{
			Action:            aws.String("DELETE"),
			ResourceRecordSet: resourceTracker.Obj.(*route53.ResourceRecordSet),
		})
	}
	human := strings.Join(names, ", ")
	klog.V(2).Infof("Deleting route53 records %q", human)

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

func ListRoute53Records(cloud fi.Cloud, clusterName string) ([]*resources.Resource, error) {
	var resourceTrackers []*resources.Resource

	if dns.IsGossipHostname(clusterName) {
		return resourceTrackers, nil
	}

	c := cloud.(awsup.AWSCloud)

	// Normalize cluster name, with leading "."
	clusterName = "." + strings.TrimSuffix(clusterName, ".")

	// TODO: If we have the zone id in the cluster spec, use it!
	var zones []*route53.HostedZone
	{
		klog.V(2).Infof("Querying for all route53 zones")

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

		klog.V(2).Infof("Querying for records in zone: %q", aws.StringValue(zone.Name))
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

				resourceTracker := &resources.Resource{
					Name:     aws.StringValue(rrs.Name),
					ID:       hostedZoneID + "/" + aws.StringValue(rrs.Name),
					Type:     "route53-record",
					GroupKey: hostedZoneID,
					GroupDeleter: func(cloud fi.Cloud, resourceTrackers []*resources.Resource) error {
						return deleteRoute53Records(cloud, zone, resourceTrackers)
					},
					Obj: rrs,
				}
				resourceTrackers = append(resourceTrackers, resourceTracker)
			}
			return true
		})
		if err != nil {
			return nil, fmt.Errorf("error querying for route53 records for zone %q: %v", aws.StringValue(zone.Name), err)
		}
	}

	return resourceTrackers, nil
}

func DeleteIAMRole(cloud fi.Cloud, r *resources.Resource) error {
	var attachedPolicies []*iam.AttachedPolicy
	var policyNames []string

	c := cloud.(awsup.AWSCloud)
	roleName := r.Name

	// List Inline policies
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
				klog.V(2).Infof("Got NoSuchEntity describing IAM RolePolicy %q; will treat as already-deleted", roleName)
				return nil
			}

			return fmt.Errorf("error listing IAM role policies for %q: %v", roleName, err)
		}
	}

	// List Attached Policies
	{
		request := &iam.ListAttachedRolePoliciesInput{
			RoleName: aws.String(roleName),
		}
		err := c.IAM().ListAttachedRolePoliciesPages(request, func(page *iam.ListAttachedRolePoliciesOutput, lastPage bool) bool {
			attachedPolicies = append(attachedPolicies, page.AttachedPolicies...)
			return true
		})
		if err != nil {
			if awsup.AWSErrorCode(err) == "NoSuchEntity" {
				klog.V(2).Infof("Got NoSuchEntity describing IAM RolePolicy %q; will treat as already-detached", roleName)
				return nil
			}

			return fmt.Errorf("error listing IAM role policies for %q: %v", roleName, err)
		}
	}

	// Delete inline policies
	for _, policyName := range policyNames {
		klog.V(2).Infof("Deleting IAM role policy %q %q", roleName, policyName)
		request := &iam.DeleteRolePolicyInput{
			RoleName:   aws.String(r.Name),
			PolicyName: aws.String(policyName),
		}
		_, err := c.IAM().DeleteRolePolicy(request)
		if err != nil {
			return fmt.Errorf("error deleting IAM role policy %q %q: %v", roleName, policyName, err)
		}
	}

	// Detach Managed Policies
	for _, policy := range attachedPolicies {
		klog.V(2).Infof("Deleting IAM role policy %q %q", roleName, policy)
		request := &iam.DetachRolePolicyInput{
			RoleName:  aws.String(r.Name),
			PolicyArn: policy.PolicyArn,
		}
		_, err := c.IAM().DetachRolePolicy(request)
		if err != nil {
			return fmt.Errorf("error detaching IAM role policy %q %q: %v", roleName, *policy.PolicyArn, err)
		}
	}

	// Delete Role
	{
		klog.V(2).Infof("Deleting IAM role %q", r.Name)
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

func ListIAMRoles(cloud fi.Cloud, clusterName string) ([]*resources.Resource, error) {
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

	var resourceTrackers []*resources.Resource

	for _, role := range roles {
		name := aws.StringValue(role.RoleName)
		resourceTracker := &resources.Resource{
			Name:    name,
			ID:      name,
			Type:    "iam-role",
			Deleter: DeleteIAMRole,
		}
		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}

func DeleteIAMInstanceProfile(cloud fi.Cloud, r *resources.Resource) error {
	c := cloud.(awsup.AWSCloud)

	profile := r.Obj.(*iam.InstanceProfile)
	name := aws.StringValue(profile.InstanceProfileName)

	// Remove roles
	{
		for _, role := range profile.Roles {
			klog.V(2).Infof("Removing role %q from IAM instance profile %q", aws.StringValue(role.RoleName), name)
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
		klog.V(2).Infof("Deleting IAM instance profile %q", name)
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

func ListIAMInstanceProfiles(cloud fi.Cloud, clusterName string) ([]*resources.Resource, error) {
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

	var resourceTrackers []*resources.Resource

	for _, profile := range profiles {
		name := aws.StringValue(profile.InstanceProfileName)
		resourceTracker := &resources.Resource{
			Name:    name,
			ID:      name,
			Type:    "iam-instance-profile",
			Deleter: DeleteIAMInstanceProfile,
			Obj:     profile,
		}

		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}

func ListSpotinstResources(cloud fi.Cloud, clusterName string) ([]*resources.Resource, error) {
	return spotinst.ListResources(cloud.(awsup.AWSCloud).Spotinst(), clusterName)
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

func FindELBV2Name(tags []*elbv2.Tag) string {
	if name, found := awsup.FindELBV2Tag(tags, "Name"); found {
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
		klog.Warningf("(new) cluster tag not found on %s", description)
		return false
	}

	tagValue := aws.StringValue(found.Value)
	switch tagValue {
	case "owned":
		return false
	case "shared":
		return true

	default:
		klog.Warningf("unknown cluster tag on %s: %q=%q", description, tagKey, tagValue)
		return false
	}
}
