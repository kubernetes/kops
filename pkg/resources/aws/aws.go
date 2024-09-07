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
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	autoscalingtypes "github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	elb "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/types"
	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	elbv2types "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	route53types "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
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
	TypeEventBridgeRule         = "eventbridge-rule"
	TypeLoadBalancer            = "load-balancer"
	TypeTargetGroup             = "target-group"
)

type listFn func(fi.Cloud, string, string) ([]*resources.Resource, error)

func ListResourcesAWS(cloud awsup.AWSCloud, clusterInfo resources.ClusterInfo) (map[string]*resources.Resource, error) {
	clusterName := clusterInfo.Name
	clusterUsesNoneDNS := clusterInfo.UsesNoneDNS

	resourceTrackers := make(map[string]*resources.Resource)

	// These are the functions that are used for looking up
	// cluster resources by their tags.
	listFunctions := []listFn{
		// EC2
		ListAutoScalingGroups,
		ListInstances,
		ListKeypairs,
		ListSecurityGroups,
		ListVolumes,
		// EC2 VPC
		ListDhcpOptions,
		ListInternetGateways,
		ListEgressOnlyInternetGateways,
		ListRouteTables,
		ListSubnets,
		ListENIs,
		// ELBs
		ListELBs,
		ListELBV2s,
		ListTargetGroups,
		// IAM
		ListIAMInstanceProfiles,
		ListIAMRoles,
		ListIAMOIDCProviders,
		// SQS
		ListSQSQueues,
		// EventBridge
		ListEventBridgeRules,
	}

	if !dns.IsGossipClusterName(clusterName) && !clusterUsesNoneDNS {
		// Route 53
		listFunctions = append(listFunctions, ListRoute53Records)
	}

	if featureflag.Spotinst.Enabled() {
		// Spotinst resources
		listFunctions = append(listFunctions, ListSpotinstResources)
	}

	var vpcID string
	{
		r, err := ListVPCs(cloud, clusterName)
		if err != nil {
			return nil, err
		}

		if len(r) > 0 {
			vpcID = r[0].ID
			resourceTrackers[r[0].Type+":"+r[0].ID] = r[0]
		}
	}

	for _, fn := range listFunctions {
		rt, err := fn(cloud, vpcID, clusterName)
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
				vpcID := aws.ToString(attachment.VpcId)
				igwID := aws.ToString(igw.InternetGatewayId)
				if vpcID == "" || igwID == "" {
					continue
				}
				vpc := resourceTrackers["vpc:"+vpcID]
				if vpc != nil && resourceTrackers["internet-gateway:"+igwID] == nil {
					resourceTrackers["internet-gateway:"+igwID] = &resources.Resource{
						Name:    FindName(igw.Tags),
						ID:      igwID,
						Obj:     igw,
						Type:    "internet-gateway",
						Dumper:  DumpInternetGateway,
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
		lts, err := FindAutoScalingLaunchTemplates(cloud, clusterName)
		if err != nil {
			return nil, err
		}
		for _, t := range lts {
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
			if resource.Type != string(ec2types.ResourceTypeRouteTable) {
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

func BuildEC2Filters(cloud fi.Cloud) []ec2types.Filter {
	awsCloud := cloud.(awsup.AWSCloud)
	tags := awsCloud.Tags()

	var filters []ec2types.Filter
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
		rtID := aws.ToString(rt.RouteTableId)
		vpcID := aws.ToString(rt.VpcId)
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
			if aws.ToBool(a.Main) {
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

func matchesElbTags(tags map[string]string, actual []elbtypes.Tag) bool {
	for k, v := range tags {
		found := false
		for _, a := range actual {
			if aws.ToString(a.Key) == k {
				if aws.ToString(a.Value) == v {
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

func matchesElbV2Tags(tags map[string]string, actual []elbv2types.Tag) bool {
	for k, v := range tags {
		found := false
		for _, a := range actual {
			if aws.ToString(a.Key) == k {
				if aws.ToString(a.Value) == v {
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

func matchesIAMTags(tags map[string]string, actual []iamtypes.Tag) bool {
	for k, v := range tags {
		found := false
		for _, a := range actual {
			if aws.ToString(a.Key) == k {
				if aws.ToString(a.Value) == v {
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

func DeleteInstances(cloud fi.Cloud, t []*resources.Resource) error {
	ctx := context.TODO()
	c := cloud.(awsup.AWSCloud)

	var ids []string
	for i, instance := range t {
		ids = append(ids, instance.ID)
		if len(ids) < 100 && i < len(t)-1 {
			continue
		}

		klog.Infof("Terminating %d EC2 instances", len(ids))
		request := &ec2.TerminateInstancesInput{
			InstanceIds: ids,
		}
		ids = []string{}
		_, err := c.EC2().TerminateInstances(ctx, request)
		if err != nil {
			if awsup.AWSErrorCode(err) == "InvalidInstanceID.NotFound" {
				klog.V(2).Infof("Got InvalidInstanceID.NotFound error terminating instances; will treat as already terminated")
			} else {
				return fmt.Errorf("error terminating instances: %v", err)
			}
		}
	}
	return nil
}

func ListInstances(cloud fi.Cloud, vpcID, clusterName string) ([]*resources.Resource, error) {
	ctx := context.TODO()
	c := cloud.(awsup.AWSCloud)

	klog.V(2).Infof("Querying EC2 instances")
	filters := BuildEC2Filters(cloud)
	filters = append(filters, awsup.NewEC2Filter("vpc-id", vpcID))
	filters = append(filters, awsup.NewEC2Filter("instance-state-name", string(ec2types.InstanceStateNameRunning)))
	request := &ec2.DescribeInstancesInput{
		Filters: filters,
	}

	var resourceTrackers []*resources.Resource

	paginator := ec2.NewDescribeInstancesPaginator(c.EC2(), request)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("error describing instances: %v", err)
		}
		for _, reservation := range page.Reservations {
			for _, instance := range reservation.Instances {
				id := aws.ToString(instance.InstanceId)

				resourceTracker := &resources.Resource{
					Name:         FindName(instance.Tags),
					ID:           id,
					Type:         string(ec2types.ResourceTypeInstance),
					GroupDeleter: DeleteInstances,
					GroupKey:     fi.ValueOf(instance.SubnetId),
					Dumper:       DumpInstance,
					Obj:          instance,
				}

				var blocks []string
				blocks = append(blocks, "subnet:"+aws.ToString(instance.SubnetId))
				blocks = append(blocks, "vpc:"+aws.ToString(instance.VpcId))

				for _, volume := range instance.BlockDeviceMappings {
					if volume.Ebs == nil || fi.ValueOf(volume.Ebs.DeleteOnTermination) {
						continue
					}
					blocks = append(blocks, "volume:"+aws.ToString(volume.Ebs.VolumeId))
				}
				for _, sg := range instance.SecurityGroups {
					blocks = append(blocks, "security-group:"+aws.ToString(sg.GroupId))
				}

				resourceTracker.Blocks = blocks

				resourceTrackers = append(resourceTrackers, resourceTracker)
			}
		}
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

func guessSSHUser(image *ec2types.Image) string {
	owner := aws.ToString(image.OwnerId)
	switch owner {
	case awsup.WellKnownAccountAmazonLinux2, awsup.WellKnownAccountRedhat:
		return "ec2-user"
	case awsup.WellKnownAccountDebian:
		return "admin"
	case awsup.WellKnownAccountUbuntu:
		return "ubuntu"
	case awsup.WellKnownAccountFlatcar:
		return "core"
	}

	name := aws.ToString(image.Name)
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
	data["type"] = ec2types.ResourceTypeInstance
	data["raw"] = r.Obj
	op.Dump.Resources = append(op.Dump.Resources, data)

	ec2Instance := r.Obj.(ec2types.Instance)
	i := &resources.Instance{
		Name: r.ID,
	}
	for _, networkInterface := range ec2Instance.NetworkInterfaces {
		if networkInterface.Association != nil {
			publicIP := aws.ToString(networkInterface.Association.PublicIp)
			if publicIP != "" {
				i.PublicAddresses = append(i.PublicAddresses, publicIP)
			}
		}
	}
	if len(i.PublicAddresses) == 0 {
		if ec2Instance.Ipv6Address != nil {
			i.PrivateAddresses = append(i.PrivateAddresses, *ec2Instance.Ipv6Address)
		} else if ec2Instance.PrivateIpAddress != nil {
			i.PrivateAddresses = append(i.PrivateAddresses, *ec2Instance.PrivateIpAddress)
		}
	}
	isControlPlane := false
	for _, tag := range ec2Instance.Tags {
		key := aws.ToString(tag.Key)
		if !strings.HasPrefix(key, awsup.TagNameRolePrefix) {
			continue
		}
		role := strings.TrimPrefix(key, awsup.TagNameRolePrefix)
		if role == "master" || role == "control-plane" {
			isControlPlane = true
		} else {
			i.Roles = append(i.Roles, role)
		}
	}
	if isControlPlane {
		i.Roles = append(i.Roles, "control-plane")
	}

	imageID := aws.ToString(ec2Instance.ImageId)
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
	ctx := context.TODO()
	c := cloud.(awsup.AWSCloud)

	id := r.ID

	klog.V(2).Infof("Deleting EC2 Volume %q", id)
	request := &ec2.DeleteVolumeInput{
		VolumeId: &id,
	}
	_, err := c.EC2().DeleteVolume(ctx, request)
	if err != nil {
		if awsup.AWSErrorCode(err) == "InvalidVolume.NotFound" {
			klog.V(2).Infof("Got InvalidVolume.NotFound error deleting Volume %q; will treat as already-deleted", id)
			return nil
		}
		if IsDependencyViolation(err) {
			return err
		}
		return fmt.Errorf("error deleting Volume %q: %v", id, err)
	}
	return nil
}

func ListVolumes(cloud fi.Cloud, vpcID, clusterName string) ([]*resources.Resource, error) {
	ctx := context.TODO()
	c := cloud.(awsup.AWSCloud)

	volumes, err := DescribeVolumes(cloud)
	if err != nil {
		return nil, err
	}
	var resourceTrackers []*resources.Resource

	elasticIPs := make(map[string]bool)
	for _, volume := range volumes {
		id := aws.ToString(volume.VolumeId)

		deleteOnTermination := false
		for _, attachment := range volume.Attachments {
			if aws.ToBool(attachment.DeleteOnTermination) {
				deleteOnTermination = true
				break
			}
		}
		if deleteOnTermination {
			continue
		}

		resourceTracker := &resources.Resource{
			Name:    FindName(volume.Tags),
			ID:      id,
			Type:    "volume",
			Deleter: DeleteVolume,
			Shared:  HasSharedTag(string(ec2types.ResourceTypeVolume)+":"+id, volume.Tags, clusterName),
		}

		var blocks []string
		// blocks = append(blocks, "vpc:" + aws.ValueOf(rt.VpcId))

		resourceTracker.Blocks = blocks

		resourceTrackers = append(resourceTrackers, resourceTracker)

		// Check for an elastic IP tag
		for _, tag := range volume.Tags {
			name := aws.ToString(tag.Key)
			ip := ""
			if name == "kubernetes.io/master-ip" {
				ip = aws.ToString(tag.Value)
			}
			if ip != "" {
				elasticIPs[ip] = true
			}
		}

	}

	if len(elasticIPs) != 0 {
		klog.V(2).Infof("Querying EC2 Elastic IPs")
		request := &ec2.DescribeAddressesInput{}
		response, err := c.EC2().DescribeAddresses(ctx, request)
		if err != nil {
			return nil, fmt.Errorf("error describing addresses: %v", err)
		}

		for _, address := range response.Addresses {
			ip := aws.ToString(address.PublicIp)
			if !elasticIPs[ip] {
				continue
			}

			resourceTrackers = append(resourceTrackers, buildElasticIPResource(address, false, clusterName))
		}
	}

	return resourceTrackers, nil
}

func DescribeVolumes(cloud fi.Cloud) ([]ec2types.Volume, error) {
	ctx := context.TODO()
	c := cloud.(awsup.AWSCloud)

	var volumes []ec2types.Volume

	klog.V(2).Infof("Listing EC2 Volumes")
	request := &ec2.DescribeVolumesInput{
		Filters: BuildEC2Filters(c),
	}

	paginator := ec2.NewDescribeVolumesPaginator(c.EC2(), request)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("error describing volumes: %v", err)
		}
		volumes = append(volumes, page.Volumes...)
	}

	return volumes, nil
}

func DeleteKeypair(cloud fi.Cloud, r *resources.Resource) error {
	ctx := context.TODO()
	c := cloud.(awsup.AWSCloud)

	id := r.ID

	klog.V(2).Infof("Deleting EC2 Keypair %q", id)
	request := &ec2.DeleteKeyPairInput{
		KeyPairId: &id,
	}
	_, err := c.EC2().DeleteKeyPair(ctx, request)
	if err != nil {
		return fmt.Errorf("error deleting KeyPair %q: %v", id, err)
	}
	return nil
}

func ListKeypairs(cloud fi.Cloud, vpcID, clusterName string) ([]*resources.Resource, error) {
	ctx := context.TODO()
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
	response, err := c.EC2().DescribeKeyPairs(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("error listing KeyPairs: %v", err)
	}

	var resourceTrackers []*resources.Resource

	for _, keypair := range response.KeyPairs {
		name := aws.ToString(keypair.KeyName)
		id := aws.ToString(keypair.KeyPairId)
		if name != keypairName && !strings.HasPrefix(name, keypairName+"-") {
			continue
		}
		resourceTracker := &resources.Resource{
			Name:    name,
			ID:      id,
			Type:    "keypair",
			Deleter: DeleteKeypair,
		}

		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}

func DeleteSubnet(cloud fi.Cloud, tracker *resources.Resource) error {
	ctx := context.TODO()
	c := cloud.(awsup.AWSCloud)

	id := tracker.ID

	klog.V(2).Infof("Deleting EC2 Subnet %q", id)
	request := &ec2.DeleteSubnetInput{
		SubnetId: &id,
	}
	_, err := c.EC2().DeleteSubnet(ctx, request)
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

func ListSubnets(cloud fi.Cloud, vpcID, clusterName string) ([]*resources.Resource, error) {
	ctx := context.TODO()
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
		subnetID := aws.ToString(subnet.SubnetId)

		shared := HasSharedTag("subnet:"+subnetID, subnet.Tags, clusterName)
		resourceTracker := &resources.Resource{
			Name:    FindName(subnet.Tags),
			ID:      subnetID,
			Type:    string(ec2types.ResourceTypeSubnet),
			Deleter: DeleteSubnet,
			Dumper:  DumpSubnet,
			Shared:  shared,
			Obj:     subnet,
		}
		resourceTracker.Blocks = append(resourceTracker.Blocks, "vpc:"+aws.ToString(subnet.VpcId))
		resourceTrackers = append(resourceTrackers, resourceTracker)

		// Get tags and append with EIPs/NGWs as needed
		for _, tag := range subnet.Tags {
			name := aws.ToString(tag.Key)
			if name == "AssociatedElasticIp" {
				eip := aws.ToString(tag.Value)
				if eip != "" {
					elasticIPs.Insert(eip)
					// A shared subnet means the EIP is not owned
					if !shared {
						ownedElasticIPs.Insert(eip)
					}
				}
			}
			if name == "AssociatedNatgateway" {
				ngwID := aws.ToString(tag.Value)
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
		response, err := c.EC2().DescribeAddresses(ctx, request)
		if err != nil {
			return nil, fmt.Errorf("error describing addresses: %v", err)
		}

		for _, address := range response.Addresses {
			ip := aws.ToString(address.PublicIp)
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
		rtResponse, err := c.EC2().DescribeRouteTables(ctx, rtRequest)
		if err != nil && awsup.AWSErrorCode(err) != "InvalidRouteTableID.NotFound" {
			return nil, fmt.Errorf("error describing RouteTables: %v", err)
		}
		// sharedNgwIds is the set of IDs for shared NGWs, that we should not delete
		sharedNgwIds := sets.NewString()
		if rtResponse != nil {
			for _, rt := range rtResponse.RouteTables {
				for _, t := range rt.Tags {
					k := aws.ToString(t.Key)
					v := aws.ToString(t.Value)

					if k == "AssociatedNatgateway" {
						sharedNgwIds.Insert(v)
					}
				}
			}
		}

		klog.V(2).Infof("Querying Nat Gateways")
		request := &ec2.DescribeNatGatewaysInput{}
		response, err := c.EC2().DescribeNatGateways(ctx, request)
		if err != nil {
			return nil, fmt.Errorf("error describing NatGateways: %v", err)
		}

		for _, ngw := range response.NatGateways {
			id := aws.ToString(ngw.NatGatewayId)
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

func DescribeSubnets(cloud fi.Cloud) ([]ec2types.Subnet, error) {
	ctx := context.TODO()
	c := cloud.(awsup.AWSCloud)

	klog.V(2).Infof("Listing EC2 subnets")
	request := &ec2.DescribeSubnetsInput{
		Filters: BuildEC2Filters(cloud),
	}
	response, err := c.EC2().DescribeSubnets(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("error listing subnets: %v", err)
	}

	return response.Subnets, nil
}

func DeleteRouteTable(cloud fi.Cloud, r *resources.Resource) error {
	ctx := context.TODO()
	c := cloud.(awsup.AWSCloud)

	id := r.ID

	klog.V(2).Infof("Deleting EC2 RouteTable %q", id)
	request := &ec2.DeleteRouteTableInput{
		RouteTableId: &id,
	}
	_, err := c.EC2().DeleteRouteTable(ctx, request)
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
func DescribeRouteTablesIgnoreTags(cloud fi.Cloud) ([]ec2types.RouteTable, error) {
	ctx := context.TODO()
	c := cloud.(awsup.AWSCloud)

	klog.V(2).Infof("Listing all RouteTables")
	request := &ec2.DescribeRouteTablesInput{}
	response, err := c.EC2().DescribeRouteTables(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("error listing RouteTables: %v", err)
	}

	return response.RouteTables, nil
}

func DeleteDhcpOptions(cloud fi.Cloud, r *resources.Resource) error {
	ctx := context.TODO()
	c := cloud.(awsup.AWSCloud)

	id := r.ID

	klog.V(2).Infof("Deleting EC2 DhcpOptions %q", id)
	request := &ec2.DeleteDhcpOptionsInput{
		DhcpOptionsId: &id,
	}
	_, err := c.EC2().DeleteDhcpOptions(ctx, request)
	if err != nil {
		if awsup.AWSErrorCode(err) == "InvalidDhcpOptionsID.NotFound" {
			klog.V(2).Infof("Got InvalidDhcpOptionsID.NotFound error deleting DhcpOptions %q; will treat as already-deleted", id)
			return nil
		} else if IsDependencyViolation(err) {
			return err
		}
		return fmt.Errorf("error deleting DhcpOptions %q: %v", id, err)
	}
	return nil
}

func ListDhcpOptions(cloud fi.Cloud, vpcID, clusterName string) ([]*resources.Resource, error) {
	dhcpOptions, err := DescribeDhcpOptions(cloud)
	if err != nil {
		return nil, err
	}

	var resourceTrackers []*resources.Resource

	for _, o := range dhcpOptions {
		resourceTracker := &resources.Resource{
			Name:    FindName(o.Tags),
			ID:      aws.ToString(o.DhcpOptionsId),
			Type:    "dhcp-options",
			Deleter: DeleteDhcpOptions,
			Shared:  HasSharedTag(string(ec2types.ResourceTypeDhcpOptions)+":"+aws.ToString(o.DhcpOptionsId), o.Tags, clusterName),
		}

		var blocks []string

		resourceTracker.Blocks = blocks

		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}

func DescribeDhcpOptions(cloud fi.Cloud) ([]ec2types.DhcpOptions, error) {
	ctx := context.TODO()
	c := cloud.(awsup.AWSCloud)

	klog.V(2).Infof("Listing EC2 DhcpOptions")
	request := &ec2.DescribeDhcpOptionsInput{
		Filters: BuildEC2Filters(cloud),
	}
	response, err := c.EC2().DescribeDhcpOptions(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("error listing DhcpOptions: %v", err)
	}

	return response.DhcpOptions, nil
}

func DeleteInternetGateway(cloud fi.Cloud, r *resources.Resource) error {
	ctx := context.TODO()
	c := cloud.(awsup.AWSCloud)

	id := r.ID

	var igw *ec2types.InternetGateway
	{
		request := &ec2.DescribeInternetGatewaysInput{
			InternetGatewayIds: []string{id},
		}
		response, err := c.EC2().DescribeInternetGateways(ctx, request)
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
		igw = &response.InternetGateways[0]
	}

	for _, a := range igw.Attachments {
		klog.V(2).Infof("Detaching EC2 InternetGateway %q", id)
		request := &ec2.DetachInternetGatewayInput{
			InternetGatewayId: &id,
			VpcId:             a.VpcId,
		}
		_, err := c.EC2().DetachInternetGateway(ctx, request)
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
		_, err := c.EC2().DeleteInternetGateway(ctx, request)
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

func DumpInternetGateway(op *resources.DumpOperation, r *resources.Resource) error {
	data := make(map[string]interface{})
	data["id"] = r.ID
	data["type"] = r.Type
	data["raw"] = r.Obj
	op.Dump.Resources = append(op.Dump.Resources, data)
	return nil
}

func ListInternetGateways(cloud fi.Cloud, vpcID, clusterName string) ([]*resources.Resource, error) {
	gateways, err := DescribeInternetGateways(cloud)
	if err != nil {
		return nil, err
	}

	var resourceTrackers []*resources.Resource

	for _, o := range gateways {
		resourceTracker := &resources.Resource{
			Name:    FindName(o.Tags),
			ID:      aws.ToString(o.InternetGatewayId),
			Type:    "internet-gateway",
			Deleter: DeleteInternetGateway,
			Shared:  HasSharedTag(string(ec2types.ResourceTypeInternetGateway)+":"+aws.ToString(o.InternetGatewayId), o.Tags, clusterName),
		}

		var blocks []string
		for _, a := range o.Attachments {
			if aws.ToString(a.VpcId) != "" {
				blocks = append(blocks, "vpc:"+aws.ToString(a.VpcId))
			}
		}
		resourceTracker.Blocks = blocks

		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}

func DescribeInternetGateways(cloud fi.Cloud) ([]ec2types.InternetGateway, error) {
	ctx := context.TODO()
	c := cloud.(awsup.AWSCloud)

	klog.V(2).Infof("Listing EC2 InternetGateways")
	request := &ec2.DescribeInternetGatewaysInput{
		Filters: BuildEC2Filters(cloud),
	}
	response, err := c.EC2().DescribeInternetGateways(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("error listing InternetGateway: %v", err)
	}

	var gateways []ec2types.InternetGateway
	gateways = append(gateways, response.InternetGateways...)

	return gateways, nil
}

// DescribeInternetGatewaysIgnoreTags returns all ec2.InternetGateways, ignoring tags
// (gateways were not always tagged in kube-up)
func DescribeInternetGatewaysIgnoreTags(cloud fi.Cloud) ([]ec2types.InternetGateway, error) {
	ctx := context.TODO()
	c := cloud.(awsup.AWSCloud)

	klog.V(2).Infof("Listing all Internet Gateways")

	request := &ec2.DescribeInternetGatewaysInput{}
	response, err := c.EC2().DescribeInternetGateways(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("error listing (all) InternetGateways: %v", err)
	}

	var gateways []ec2types.InternetGateway

	gateways = append(gateways, response.InternetGateways...)

	return gateways, nil
}

func DumpEgressOnlyInternetGateway(op *resources.DumpOperation, r *resources.Resource) error {
	data := make(map[string]interface{})
	data["id"] = r.ID
	data["type"] = r.Type
	data["raw"] = r.Obj
	op.Dump.Resources = append(op.Dump.Resources, data)
	return nil
}

func DeleteEgressOnlyInternetGateway(cloud fi.Cloud, r *resources.Resource) error {
	ctx := context.TODO()
	c := cloud.(awsup.AWSCloud)

	id := r.ID

	{
		klog.V(2).Infof("Deleting EC2 EgressOnlyInternetGateway %q", id)
		request := &ec2.DeleteEgressOnlyInternetGatewayInput{
			EgressOnlyInternetGatewayId: &id,
		}
		_, err := c.EC2().DeleteEgressOnlyInternetGateway(ctx, request)
		if err != nil {
			if IsDependencyViolation(err) {
				return err
			}
			if awsup.AWSErrorCode(err) == "InvalidEgressOnlyInternetGatewayID.NotFound" {
				klog.Infof("Egress-only internet gateway %q not found; assuming already deleted", id)
				return nil
			}
			return fmt.Errorf("error deleting EgressOnlyInternetGateway %q: %v", id, err)
		}
	}

	return nil
}

func ListEgressOnlyInternetGateways(cloud fi.Cloud, vpcID, clusterName string) ([]*resources.Resource, error) {
	gateways, err := DescribeEgressOnlyInternetGateways(cloud)
	if err != nil {
		return nil, err
	}

	var resourceTrackers []*resources.Resource

	for _, o := range gateways {
		resourceTracker := &resources.Resource{
			Name:    FindName(o.Tags),
			ID:      aws.ToString(o.EgressOnlyInternetGatewayId),
			Type:    "egress-only-internet-gateway",
			Obj:     o,
			Dumper:  DumpEgressOnlyInternetGateway,
			Deleter: DeleteEgressOnlyInternetGateway,
			Shared:  HasSharedTag(string(ec2types.ResourceTypeEgressOnlyInternetGateway)+":"+aws.ToString(o.EgressOnlyInternetGatewayId), o.Tags, clusterName),
		}

		var blocks []string
		for _, a := range o.Attachments {
			if aws.ToString(a.VpcId) != "" {
				blocks = append(blocks, "vpc:"+aws.ToString(a.VpcId))
			}
		}
		resourceTracker.Blocks = blocks

		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}

func DescribeEgressOnlyInternetGateways(cloud fi.Cloud) ([]ec2types.EgressOnlyInternetGateway, error) {
	ctx := context.TODO()
	c := cloud.(awsup.AWSCloud)

	klog.V(2).Infof("Listing EC2 EgressOnlyInternetGateways")
	request := &ec2.DescribeEgressOnlyInternetGatewaysInput{
		Filters: BuildEC2Filters(cloud),
	}
	response, err := c.EC2().DescribeEgressOnlyInternetGateways(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("error listing EgressOnlyInternetGateway: %v", err)
	}

	var gateways []ec2types.EgressOnlyInternetGateway
	gateways = append(gateways, response.EgressOnlyInternetGateways...)

	return gateways, nil
}

func DeleteAutoScalingGroup(cloud fi.Cloud, r *resources.Resource) error {
	ctx := context.TODO()

	c := cloud.(awsup.AWSCloud)

	id := r.ID

	klog.V(2).Infof("Deleting autoscaling group %q", id)
	request := &autoscaling.DeleteAutoScalingGroupInput{
		AutoScalingGroupName: &id,
		ForceDelete:          aws.Bool(true),
	}
	_, err := c.Autoscaling().DeleteAutoScalingGroup(ctx, request)
	if err != nil {
		if IsDependencyViolation(err) {
			return err
		}
		return fmt.Errorf("error deleting autoscaling group %q: %v", id, err)
	}
	return nil
}

func ListAutoScalingGroups(cloud fi.Cloud, vpcID, clusterName string) ([]*resources.Resource, error) {
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
			ID:      aws.ToString(asg.AutoScalingGroupName),
			Type:    "autoscaling-group",
			Deleter: DeleteAutoScalingGroup,
		}

		var blocks []string
		subnets := aws.ToString(asg.VPCZoneIdentifier)
		for _, subnet := range strings.Split(subnets, ",") {
			if subnet == "" {
				continue
			}
			blocks = append(blocks, "subnet:"+subnet)
		}
		if asg.LaunchConfigurationName != nil {
			blocks = append(blocks, TypeAutoscalingLaunchConfig+":"+aws.ToString(asg.LaunchConfigurationName))
		}
		if asg.LaunchTemplate != nil {
			blocks = append(blocks, TypeAutoscalingLaunchConfig+":"+aws.ToString(asg.LaunchTemplate.LaunchTemplateName))
		}

		resourceTracker.Blocks = blocks

		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}

// FindAutoScalingLaunchTemplates finds any launch templates owned by the cluster (by tag).
func FindAutoScalingLaunchTemplates(cloud fi.Cloud, clusterName string) ([]*resources.Resource, error) {
	ctx := context.TODO()
	c := cloud.(awsup.AWSCloud)

	klog.V(2).Infof("Finding all AutoScaling LaunchTemplates owned by the cluster")

	input := &ec2.DescribeLaunchTemplatesInput{
		Filters: []ec2types.Filter{
			{
				Name:   aws.String("tag:kubernetes.io/cluster/" + clusterName),
				Values: []string{"owned"},
			},
		},
	}

	var list []*resources.Resource
	paginator := ec2.NewDescribeLaunchTemplatesPaginator(c.EC2(), input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("error listing AutoScaling LaunchTemplates: %v", err)
		}
		for _, lt := range page.LaunchTemplates {
			list = append(list, &resources.Resource{
				Name:    aws.ToString(lt.LaunchTemplateName),
				ID:      aws.ToString(lt.LaunchTemplateId),
				Type:    TypeAutoscalingLaunchConfig,
				Deleter: DeleteAutoScalingGroupLaunchTemplate,
			})
		}
	}

	return list, nil
}

func FindNatGateways(cloud fi.Cloud, routeTables map[string]*resources.Resource, clusterName string) ([]*resources.Resource, error) {
	if len(routeTables) == 0 {
		return nil, nil
	}

	ctx := context.TODO()
	c := cloud.(awsup.AWSCloud)

	natGatewayIds := sets.NewString()
	ownedNatGatewayIds := sets.NewString()
	{
		request := &ec2.DescribeRouteTablesInput{}
		for _, routeTable := range routeTables {
			request.RouteTableIds = append(request.RouteTableIds, routeTable.ID)
		}
		response, err := c.EC2().DescribeRouteTables(ctx, request)
		if err != nil && awsup.AWSErrorCode(err) != "InvalidRouteTableID.NotFound" {
			return nil, fmt.Errorf("error from DescribeRouteTables: %v", err)
		}
		if response != nil {
			for _, rt := range response.RouteTables {
				routeTableID := aws.ToString(rt.RouteTableId)
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
	}

	var resourceTrackers []*resources.Resource
	for natGatewayId := range natGatewayIds {
		request := &ec2.DescribeNatGatewaysInput{
			NatGatewayIds: []string{natGatewayId},
		}
		response, err := c.EC2().DescribeNatGateways(ctx, request)
		if err != nil {
			if awsup.AWSErrorCode(err) == "NatGatewayNotFound" {
				klog.V(2).Infof("Got NatGatewayNotFound describing NatGateway %s; will treat as already-deleted", natGatewayId)
				continue
			}
			return nil, fmt.Errorf("error from DescribeNatGateways: %v", err)
		}

		if response.NextToken != nil {
			return nil, fmt.Errorf("NextToken set from DescribeNatGateways, but pagination not implemented")
		}

		for _, ngw := range response.NatGateways {
			natGatewayId := aws.ToString(ngw.NatGatewayId)

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
					request.AllocationIds = []string{aws.ToString(address.AllocationId)}
					response, err := c.EC2().DescribeAddresses(ctx, request)
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
	ctx := context.TODO()
	c, ok := cloud.(awsup.AWSCloud)
	if !ok {
		return errors.New("expected a aws.Cloud provider")
	}
	klog.V(2).Infof("Deleting EC2 LaunchTemplate %q", r.ID)

	if _, err := c.EC2().DeleteLaunchTemplate(ctx, &ec2.DeleteLaunchTemplateInput{
		LaunchTemplateId: fi.PtrTo(r.ID),
	}); err != nil {
		return fmt.Errorf("error deleting ec2 LaunchTemplate %q: %v", r.ID, err)
	}

	return nil
}

func DeleteELB(cloud fi.Cloud, r *resources.Resource) error {
	ctx := context.TODO()
	c := cloud.(awsup.AWSCloud)

	id := r.ID

	klog.V(2).Infof("Deleting ELB %q", id)
	request := &elb.DeleteLoadBalancerInput{
		LoadBalancerName: &id,
	}
	_, err := c.ELB().DeleteLoadBalancer(ctx, request)
	if err != nil {
		if IsDependencyViolation(err) {
			return err
		}
		return fmt.Errorf("error deleting LoadBalancer %q: %v", id, err)
	}
	return nil
}

func DeleteELBV2(cloud fi.Cloud, r *resources.Resource) error {
	ctx := context.TODO()
	c := cloud.(awsup.AWSCloud)
	id := r.ID

	klog.V(2).Infof("Deleting ELBV2 %q", id)
	request := &elbv2.DeleteLoadBalancerInput{
		LoadBalancerArn: aws.String(id),
	}
	_, err := c.ELBV2().DeleteLoadBalancer(ctx, request)
	if err != nil {
		if IsDependencyViolation(err) {
			return err
		}
		return fmt.Errorf("error deleting V2 LoadBalancer %q: %v", id, err)
	}
	return nil
}

func DeleteTargetGroup(cloud fi.Cloud, r *resources.Resource) error {
	ctx := context.TODO()
	c := cloud.(awsup.AWSCloud)
	id := r.ID

	klog.V(2).Infof("Deleting TargetGroup %q", id)
	request := &elbv2.DeleteTargetGroupInput{
		TargetGroupArn: aws.String(id),
	}
	_, err := c.ELBV2().DeleteTargetGroup(ctx, request)
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

	if lb, ok := r.Obj.(elbv2types.LoadBalancer); ok {
		op.Dump.LoadBalancers = append(op.Dump.LoadBalancers, &resources.LoadBalancer{
			Name:    fi.ValueOf(lb.LoadBalancerName),
			DNSName: fi.ValueOf(lb.DNSName),
		})

	}
	return nil
}

func ListELBs(cloud fi.Cloud, vpcID, clusterName string) ([]*resources.Resource, error) {
	elbs, elbTags, err := DescribeELBs(cloud)
	if err != nil {
		return nil, err
	}

	var resourceTrackers []*resources.Resource
	for _, elb := range elbs {
		id := aws.ToString(elb.LoadBalancerName)
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
			blocks = append(blocks, "security-group:"+sg)
		}
		for _, s := range elb.Subnets {
			blocks = append(blocks, "subnet:"+s)
		}
		blocks = append(blocks, "vpc:"+aws.ToString(elb.VPCId))

		resourceTracker.Blocks = blocks

		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}

func DescribeELBs(cloud fi.Cloud) ([]elbtypes.LoadBalancerDescription, map[string][]elbtypes.Tag, error) {
	ctx := context.TODO()
	c := cloud.(awsup.AWSCloud)
	tags := c.Tags()

	klog.V(2).Infof("Listing all ELBs")

	request := &elb.DescribeLoadBalancersInput{}
	// ELB DescribeTags has a limit of 20 names, so we set the page size here to 20 also
	request.PageSize = aws.Int32(20)

	var elbs []elbtypes.LoadBalancerDescription
	elbTags := make(map[string][]elbtypes.Tag)

	paginator := elb.NewDescribeLoadBalancersPaginator(c.ELB(), request)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("error listing elbs: %v", err)
		}
		if len(page.LoadBalancerDescriptions) == 0 {
			continue
		}
		tagRequest := &elb.DescribeTagsInput{}

		nameToELB := make(map[string]elbtypes.LoadBalancerDescription)
		for _, elb := range page.LoadBalancerDescriptions {
			name := aws.ToString(elb.LoadBalancerName)
			nameToELB[name] = elb

			tagRequest.LoadBalancerNames = append(tagRequest.LoadBalancerNames, aws.ToString(elb.LoadBalancerName))
		}

		tagResponse, err := c.ELB().DescribeTags(ctx, tagRequest)
		if err != nil {
			return nil, nil, fmt.Errorf("error listing elb Tags: %v", err)
		}

		for _, t := range tagResponse.TagDescriptions {
			elbName := aws.ToString(t.LoadBalancerName)

			if !matchesElbTags(tags, t.Tags) {
				continue
			}

			elbTags[elbName] = t.Tags

			elb := nameToELB[elbName]
			elbs = append(elbs, elb)
		}
	}
	return elbs, elbTags, nil
}

// For NLBs and ALBs
func ListELBV2s(cloud fi.Cloud, vpcID, clusterName string) ([]*resources.Resource, error) {
	ctx := context.TODO()

	loadBalancers, err := awsup.ListELBV2LoadBalancers(ctx, cloud.(awsup.AWSCloud))
	if err != nil {
		return nil, err
	}

	var resourceTrackers []*resources.Resource
	for _, loadBalancer := range loadBalancers {
		elb := loadBalancer.LoadBalancer
		id := aws.ToString(elb.LoadBalancerName)
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
			blocks = append(blocks, "security-group:"+sg)
		}

		blocks = append(blocks, "vpc:"+aws.ToString(elb.VpcId))

		resourceTracker.Blocks = blocks

		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}

func DumpTargetGroup(op *resources.DumpOperation, r *resources.Resource) error {
	data := make(map[string]interface{})
	data["id"] = r.ID
	data["type"] = TypeTargetGroup
	data["raw"] = r.Obj
	op.Dump.Resources = append(op.Dump.Resources, data)
	return nil
}

func ListTargetGroups(cloud fi.Cloud, vpcID, clusterName string) ([]*resources.Resource, error) {
	targetGroups, err := listMatchingTargetGroups(cloud)
	if err != nil {
		return nil, err
	}

	var resourceTrackers []*resources.Resource
	for _, targetGroup := range targetGroups {
		tg := targetGroup.TargetGroup
		id := aws.ToString(tg.TargetGroupName)
		resourceTracker := &resources.Resource{
			Name:    id,
			ID:      targetGroup.ARN,
			Type:    TypeTargetGroup,
			Deleter: DeleteTargetGroup,
			Dumper:  DumpTargetGroup,
			Obj:     tg,
		}

		resourceTrackers = append(resourceTrackers, resourceTracker)
	}
	return resourceTrackers, nil
}

func listMatchingTargetGroups(cloud fi.Cloud) ([]*awsup.TargetGroupInfo, error) {
	ctx := context.TODO()

	c := cloud.(awsup.AWSCloud)
	tags := c.Tags()

	klog.V(2).Infof("Listing all TargetGroups")

	targetGroups, err := awsup.ListELBV2TargetGroups(ctx, c)
	if err != nil {
		return nil, err
	}

	var matches []*awsup.TargetGroupInfo
	for _, tg := range targetGroups {
		if matchesElbV2Tags(tags, tg.Tags) {
			matches = append(matches, tg)
		}
	}

	return matches, nil
}

func DeleteElasticIP(cloud fi.Cloud, t *resources.Resource) error {
	ctx := context.TODO()
	c := cloud.(awsup.AWSCloud)

	id := t.ID

	klog.V(2).Infof("Releasing IP %s", t.Name)
	request := &ec2.ReleaseAddressInput{
		AllocationId: &id,
	}
	_, err := c.EC2().ReleaseAddress(ctx, request)
	if err != nil {
		if awsup.AWSErrorCode(err) == "InvalidAllocationID.NotFound" {
			klog.V(2).Infof("Got InvalidAllocationID.NotFound error deleting ElasticIP %q; will treat as already-deleted", id)
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
	ctx := context.TODO()
	c := cloud.(awsup.AWSCloud)

	id := t.ID

	klog.V(2).Infof("Removing NatGateway %s", t.Name)
	request := &ec2.DeleteNatGatewayInput{
		NatGatewayId: &id,
	}
	_, err := c.EC2().DeleteNatGateway(ctx, request)
	if err != nil {
		if IsDependencyViolation(err) {
			return err
		}
		return fmt.Errorf("error deleting ngw %q: %v", t.Name, err)
	}
	return nil
}

func deleteRoute53Records(ctx context.Context, cloud fi.Cloud, zone route53types.HostedZone, resourceTrackers []*resources.Resource) error {
	c := cloud.(awsup.AWSCloud)

	var changes []route53types.Change
	var names []string
	for _, resourceTracker := range resourceTrackers {
		names = append(names, resourceTracker.Name)
		changes = append(changes, route53types.Change{
			Action:            route53types.ChangeActionDelete,
			ResourceRecordSet: resourceTracker.Obj.(*route53types.ResourceRecordSet),
		})
	}
	human := strings.Join(names, ", ")
	klog.V(2).Infof("Deleting route53 records %q", human)

	changeBatch := &route53types.ChangeBatch{
		Changes: changes,
	}
	request := &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: zone.Id,
		ChangeBatch:  changeBatch,
	}
	_, err := c.Route53().ChangeResourceRecordSets(ctx, request)
	if err != nil {
		return fmt.Errorf("error deleting route53 record %q: %v", human, err)
	}
	return nil
}

func ListRoute53Records(cloud fi.Cloud, vpcID, clusterName string) ([]*resources.Resource, error) {
	ctx := context.TODO()
	var resourceTrackers []*resources.Resource

	c := cloud.(awsup.AWSCloud)

	// Normalize cluster name, with leading "."
	clusterName = "." + strings.TrimSuffix(clusterName, ".")

	// TODO: If we have the zone id in the cluster spec, use it!
	var zones []route53types.HostedZone
	{
		klog.V(2).Infof("Querying for all route53 zones")

		request := &route53.ListHostedZonesInput{}
		paginator := route53.NewListHostedZonesPaginator(c.Route53(), request)
		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				return nil, fmt.Errorf("error querying for route53 zones: %w", err)
			}
			for _, zone := range page.HostedZones {
				zoneName := aws.ToString(zone.Name)
				zoneName = "." + strings.TrimSuffix(zoneName, ".")

				if strings.HasSuffix(clusterName, zoneName) {
					zones = append(zones, zone)
				}
			}
		}
	}

	for i := range zones {
		// Be super careful because we close over this later (in groupDeleter)
		zone := zones[i]

		hostedZoneID := strings.TrimPrefix(aws.ToString(zone.Id), "/hostedzone/")

		klog.V(2).Infof("Querying for records in zone: %q", aws.ToString(zone.Name))
		request := &route53.ListResourceRecordSetsInput{
			HostedZoneId: zone.Id,
		}
		paginator := route53.NewListResourceRecordSetsPaginator(c.Route53(), request)
		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				return nil, fmt.Errorf("error querying for route53 records for zone %q: %v", aws.ToString(zone.Name), err)
			}
			for _, rrs := range page.ResourceRecordSets {
				if rrs.Type != route53types.RRTypeA &&
					rrs.Type != route53types.RRTypeAaaa &&
					rrs.Type != route53types.RRTypeTxt {
					continue
				}

				name := aws.ToString(rrs.Name)
				name = "." + strings.TrimSuffix(name, ".")

				if !strings.HasSuffix(name, clusterName) {
					continue
				}
				prefix := strings.TrimSuffix(name, clusterName)

				// Also trim ownership records for AAAA records
				if rrs.Type == route53types.RRTypeTxt && strings.HasPrefix(prefix, ".aaaa-") {
					prefix = "." + strings.TrimPrefix(prefix, ".aaaa-")
				}

				remove := false
				// TODO: Compute the actual set of names?
				if prefix == ".api" || prefix == ".api.internal" || prefix == ".bastion" || prefix == ".kops-controller.internal" {
					remove = true
				} else if strings.HasPrefix(prefix, ".etcd-") {
					remove = true
				}

				if !remove {
					continue
				}

				resourceTracker := &resources.Resource{
					Name:     aws.ToString(rrs.Name),
					ID:       hostedZoneID + "/" + string(rrs.Type) + "/" + aws.ToString(rrs.Name),
					Type:     "route53-record",
					GroupKey: hostedZoneID,
					GroupDeleter: func(cloud fi.Cloud, resourceTrackers []*resources.Resource) error {
						return deleteRoute53Records(ctx, cloud, zone, resourceTrackers)
					},
					Obj: &rrs,
				}
				resourceTrackers = append(resourceTrackers, resourceTracker)
			}
		}
	}

	return resourceTrackers, nil
}

func DeleteIAMRole(cloud fi.Cloud, r *resources.Resource) error {
	ctx := context.TODO()
	var attachedPolicies []iamtypes.AttachedPolicy
	var policyNames []string

	c := cloud.(awsup.AWSCloud)
	roleName := r.Name

	// List Inline policies
	{
		request := &iam.ListRolePoliciesInput{
			RoleName: aws.String(roleName),
		}
		paginator := iam.NewListRolePoliciesPaginator(c.IAM(), request)
		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				if awsup.IsIAMNoSuchEntityException(err) {
					klog.V(2).Infof("Got NoSuchEntity describing IAM RolePolicy %q; will treat as already-deleted", roleName)
					return nil
				}
				return fmt.Errorf("error listing IAM role policies for %q: %v", roleName, err)
			}
			policyNames = append(policyNames, page.PolicyNames...)
		}
	}

	// List Attached Policies
	{
		request := &iam.ListAttachedRolePoliciesInput{
			RoleName: aws.String(roleName),
		}
		paginator := iam.NewListAttachedRolePoliciesPaginator(c.IAM(), request)
		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				if awsup.IsIAMNoSuchEntityException(err) {
					klog.V(2).Infof("Got NoSuchEntity describing IAM RolePolicy %q; will treat as already-deleted", roleName)
					return nil
				}
				return fmt.Errorf("error listing IAM role policies for %q: %v", roleName, err)
			}
			attachedPolicies = append(attachedPolicies, page.AttachedPolicies...)
		}
	}

	// Delete inline policies
	for _, policyName := range policyNames {
		klog.V(2).Infof("Deleting IAM role policy %q %q", roleName, policyName)
		request := &iam.DeleteRolePolicyInput{
			RoleName:   aws.String(r.Name),
			PolicyName: aws.String(policyName),
		}
		_, err := c.IAM().DeleteRolePolicy(ctx, request)
		if err != nil {
			return fmt.Errorf("error deleting IAM role policy %q %q: %v", roleName, policyName, err)
		}
	}

	// Detach Managed Policies
	for _, policy := range attachedPolicies {
		klog.V(2).Infof("Detaching IAM role policy %q %v", roleName, policy)
		request := &iam.DetachRolePolicyInput{
			RoleName:  aws.String(r.Name),
			PolicyArn: policy.PolicyArn,
		}
		_, err := c.IAM().DetachRolePolicy(ctx, request)
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
		_, err := c.IAM().DeleteRole(ctx, request)
		if err != nil {
			return fmt.Errorf("error deleting IAM role %q: %v", r.Name, err)
		}
	}

	return nil
}

func ListIAMRoles(cloud fi.Cloud, vpcID, clusterName string) ([]*resources.Resource, error) {
	ctx := context.TODO()
	c := cloud.(awsup.AWSCloud)

	var resourceTrackers []*resources.Resource
	// Find roles owned by the cluster
	{
		ownershipTag := "kubernetes.io/cluster/" + clusterName
		request := &iam.ListRolesInput{}
		paginator := iam.NewListRolesPaginator(c.IAM(), request)
		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				return nil, fmt.Errorf("error listing IAM roles: %v", err)
			}
			for _, r := range page.Roles {
				name := aws.ToString(r.RoleName)

				getRequest := &iam.GetRoleInput{RoleName: r.RoleName}
				roleOutput, err := c.IAM().GetRole(ctx, getRequest)
				if err != nil {
					if awsup.IsIAMNoSuchEntityException(err) {
						klog.Warningf("could not find role %q. Resource may already have been deleted: %v", name, err)
						continue
					} else if awsup.AWSErrorCode(err) == "403" {
						klog.Warningf("failed to determine ownership of %q: %v", name, err)
						continue
					}
					return nil, fmt.Errorf("calling IAM GetRole on %s: %w", name, err)
				}
				for _, tag := range roleOutput.Role.Tags {
					if fi.ValueOf(tag.Key) == ownershipTag && fi.ValueOf(tag.Value) == "owned" {
						resourceTracker := &resources.Resource{
							Name:    name,
							ID:      name,
							Type:    "iam-role",
							Deleter: DeleteIAMRole,
						}
						resourceTrackers = append(resourceTrackers, resourceTracker)
					}
				}
			}
		}
	}

	return resourceTrackers, nil
}

func DeleteIAMInstanceProfile(cloud fi.Cloud, r *resources.Resource) error {
	ctx := context.TODO()
	c := cloud.(awsup.AWSCloud)

	profile := r.Obj.(iamtypes.InstanceProfile)
	name := aws.ToString(profile.InstanceProfileName)

	// Remove roles
	{
		for _, role := range profile.Roles {
			klog.V(2).Infof("Removing role %q from IAM instance profile %q", aws.ToString(role.RoleName), name)
			request := &iam.RemoveRoleFromInstanceProfileInput{
				InstanceProfileName: profile.InstanceProfileName,
				RoleName:            role.RoleName,
			}
			_, err := c.IAM().RemoveRoleFromInstanceProfile(ctx, request)
			if err != nil {
				return fmt.Errorf("error removing role %q from IAM instance profile %q: %v", aws.ToString(role.RoleName), name, err)
			}
		}
	}

	// Delete the instance profile
	{
		klog.V(2).Infof("Deleting IAM instance profile %q", name)
		request := &iam.DeleteInstanceProfileInput{
			InstanceProfileName: profile.InstanceProfileName,
		}
		_, err := c.IAM().DeleteInstanceProfile(ctx, request)
		if err != nil {
			return fmt.Errorf("error deleting IAM instance profile %q: %v", name, err)
		}
	}

	return nil
}

func ListIAMInstanceProfiles(cloud fi.Cloud, vpcID, clusterName string) ([]*resources.Resource, error) {
	ctx := context.TODO()
	c := cloud.(awsup.AWSCloud)

	var profiles []iamtypes.InstanceProfile
	ownershipTag := "kubernetes.io/cluster/" + clusterName

	request := &iam.ListInstanceProfilesInput{}
	paginator := iam.NewListInstanceProfilesPaginator(c.IAM(), request)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("error listing IAM instance profiles: %v", err)
		}
		for _, p := range page.InstanceProfiles {
			name := aws.ToString(p.InstanceProfileName)

			getRequest := &iam.GetInstanceProfileInput{InstanceProfileName: p.InstanceProfileName}
			profileOutput, err := c.IAM().GetInstanceProfile(ctx, getRequest)
			if err != nil {
				if awsup.IsIAMNoSuchEntityException(err) {
					klog.Warningf("could not find role %q. Resource may already have been deleted: %v", name, err)
					continue
				} else if awsup.AWSErrorCode(err) == "403" {
					klog.Warningf("failed to determine ownership of %q: %v", *p.InstanceProfileName, err)
					continue
				}
				return nil, fmt.Errorf("calling IAM GetInstanceProfile on %s: %w", name, err)
			}
			for _, tag := range profileOutput.InstanceProfile.Tags {
				if fi.ValueOf(tag.Key) == ownershipTag && fi.ValueOf(tag.Value) == "owned" {
					profiles = append(profiles, p)
				}
			}
		}
	}

	var resourceTrackers []*resources.Resource

	for _, profile := range profiles {
		name := aws.ToString(profile.InstanceProfileName)
		resourceTracker := &resources.Resource{
			Name:    name,
			ID:      name,
			Type:    "iam-instance-profile",
			Deleter: DeleteIAMInstanceProfile,
			Obj:     profile,
		}
		resourceTracker.Blocks = append(resourceTracker.Blocks, "iam-role:"+name)

		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}

func ListIAMOIDCProviders(cloud fi.Cloud, vpcID, clusterName string) ([]*resources.Resource, error) {
	ctx := context.TODO()
	c := cloud.(awsup.AWSCloud)
	tags := c.Tags()

	var providers []*string
	{
		request := &iam.ListOpenIDConnectProvidersInput{}
		response, err := c.IAM().ListOpenIDConnectProviders(ctx, request)
		if err != nil {
			return nil, fmt.Errorf("error listing IAM OIDC Providers: %v", err)
		}
		for _, provider := range response.OpenIDConnectProviderList {
			arn := provider.Arn
			descReq := &iam.GetOpenIDConnectProviderInput{
				OpenIDConnectProviderArn: arn,
			}
			resp, err := c.IAM().GetOpenIDConnectProvider(ctx, descReq)
			if err != nil {
				if awsup.IsIAMNoSuchEntityException(err) {
					klog.Warningf("could not find IAM OIDC Provider %q. Resource may already have been deleted: %v", aws.ToString(arn), err)
					continue
				} else if awsup.AWSErrorCode(err) == "403" {
					klog.Warningf("failed to determine ownership of %q: %v", aws.ToString(arn), err)
					continue
				}
				return nil, fmt.Errorf("error getting IAM OIDC Provider %q: %w", aws.ToString(arn), err)
			}
			if !matchesIAMTags(tags, resp.Tags) {
				continue
			}
			providers = append(providers, arn)
		}
	}

	var resourceTrackers []*resources.Resource

	for _, arn := range providers {
		resourceTracker := &resources.Resource{
			Name:    aws.ToString(arn),
			ID:      aws.ToString(arn),
			Type:    "oidc-provider",
			Deleter: DeleteIAMOIDCProvider,
		}
		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}

func DeleteIAMOIDCProvider(cloud fi.Cloud, r *resources.Resource) error {
	ctx := context.TODO()
	c := cloud.(awsup.AWSCloud)
	arn := fi.PtrTo(r.ID)
	{
		klog.V(2).Infof("Deleting IAM OIDC Provider %v", arn)
		request := &iam.DeleteOpenIDConnectProviderInput{
			OpenIDConnectProviderArn: arn,
		}
		_, err := c.IAM().DeleteOpenIDConnectProvider(ctx, request)
		if err != nil {
			if awsup.IsIAMNoSuchEntityException(err) {
				klog.V(2).Infof("Got NoSuchEntity deleting IAM OIDC Provider %v; will treat as already-deleted", arn)
				return nil
			}
			return fmt.Errorf("error deleting IAM OIDC Provider %v: %v", arn, err)
		}
	}

	return nil
}

func ListSpotinstResources(cloud fi.Cloud, vpcID, clusterName string) ([]*resources.Resource, error) {
	return spotinst.ListResources(cloud.(awsup.AWSCloud).Spotinst(), clusterName)
}

func FindName(tags []ec2types.Tag) string {
	if name, found := awsup.FindEC2Tag(tags, "Name"); found {
		return name
	}
	return ""
}

func FindASGName(tags []autoscalingtypes.TagDescription) string {
	if name, found := awsup.FindASGTag(tags, "Name"); found {
		return name
	}
	return ""
}

func FindELBName(tags []elbtypes.Tag) string {
	if name, found := awsup.FindELBTag(tags, "Name"); found {
		return name
	}
	return ""
}

func FindELBV2Name(tags []elbv2types.Tag) string {
	if name, found := awsup.FindELBV2Tag(tags, "Name"); found {
		return name
	}
	return ""
}

// HasSharedTag looks for the shared tag indicating that the cluster does not own the resource
func HasSharedTag(description string, tags []ec2types.Tag, clusterName string) bool {
	tagKey := "kubernetes.io/cluster/" + clusterName

	var found *ec2types.Tag
	for _, tag := range tags {
		if aws.ToString(tag.Key) != tagKey {
			continue
		}

		found = &tag
	}

	if found == nil {
		klog.Warningf("(new) cluster tag not found on %s", description)
		return false
	}

	tagValue := aws.ToString(found.Value)
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
