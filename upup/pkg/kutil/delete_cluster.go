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
	"time"
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

// HasStatus is implemented by resources where we want to hint the dependencies
// (ideally we would implement for everything, but realistically there are only a few where it is worthwhile)
type HasStatus interface {
	Status(cloud fi.Cloud) (exists bool, blocks []string, err error)
}

func (c *DeleteCluster) ListResources() (map[string]DeletableResource, error) {
	cloud := c.Cloud.(*awsup.AWSCloud)

	resources := make(map[string]DeletableResource)

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
				r := &DeletableASG{Name: *t.AutoScalingGroupName}
				resources["autoscaling-group:"+r.Name] = r
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
				r := &DeletableAutoscalingLaunchConfiguration{Name: *t.LaunchConfigurationName}
				resources["autoscaling-launchconfiguration:"+r.Name] = r
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
				r := &DeletableELBLoadBalancer{Name: *t.LoadBalancerName}
				resources["elb:"+r.Name] = r
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

			resources[*t.ResourceType+":"+*t.ResourceId] = resource
		}
	}

	return resources, nil
}

func (c *DeleteCluster) DeleteResources(resources map[string]DeletableResource) error {
	depMap := make(map[string][]string)

	done := make(map[string]DeletableResource)

	// Initial pass to check that resources actually exist
	for k, r := range resources {
		hs, ok := r.(HasStatus)
		if !ok {
			continue
		}

		fmt.Printf("Checking status of resource %s: ", k)
		exists, blocks, err := hs.Status(c.Cloud)
		if err != nil {
			fmt.Printf("error (ignoring): %v\n", err)
		} else if exists {
			fmt.Printf("exists (gathered dependencies)\n")
		} else {
			fmt.Printf("already removed\n")
			done[k] = r
		}

		for _, block := range blocks {
			depMap[block] = append(depMap[block], k)
		}
	}

	glog.Infof("Dependencies")
	for k, v := range depMap {
		glog.Infof("\t%s\t%v", k, v)
	}

	for {
		// TODO: Some form of default ordering based on types?
		// TODO: Give up eventually?

		failed := make(map[string]DeletableResource)

		for {
			phase := make(map[string]DeletableResource)

			for k, r := range resources {
				if _, d := done[k]; d {
					continue
				}

				if _, d := failed[k]; d {
					// Only attempt each resource once per pass
					continue
				}

				ready := true
				for _, dep := range depMap[k] {
					if _, d := done[dep]; !d {
						glog.V(4).Infof("dependency %q of %q not deleted; skipping")
						ready = false
					}
				}
				if !ready {
					continue
				}

				phase[k] = r
			}

			if len(phase) == 0 {
				break
			}

			// TODO: Parallel delete?
			for k, r := range phase {
				fmt.Printf("Deleting resource %s:  ", k)
				err := r.Delete(c.Cloud)
				if err != nil {
					if IsDependencyViolation(err) {
						fmt.Printf("still has dependencies, will retry\n")
					} else {
						fmt.Printf("error deleting resource, will retry: %v\n", err)
					}
					failed[k] = r
				} else {
					fmt.Printf(" ok\n")
					done[k] = r
				}
			}
		}

		if len(resources) == len(done) {
			return nil
		}

		fmt.Printf("Not all resources deleted; waiting before reattempting deletion\n")
		for k := range resources {
			if _, d := done[k]; d {
				continue
			}

			fmt.Printf("\t%s\n", k)
		}
		time.Sleep(10 * time.Second)
	}
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

func (r *DeletableInstance) Status(cloud fi.Cloud) (bool, []string, error) {
	c := cloud.(*awsup.AWSCloud)

	glog.V(2).Infof("Querying EC2 instance %q", r.ID)
	request := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{&r.ID},
	}
	response, err := c.EC2.DescribeInstances(request)
	if err != nil {
		return false, nil, fmt.Errorf("error describing instance %q: %v", r.ID, err)
	}

	var found []*ec2.Instance
	for _, reservation := range response.Reservations {
		for _, instance := range reservation.Instances {
			if aws.StringValue(instance.InstanceId) == r.ID {
				found = append(found, instance)
			}
		}
	}
	if len(found) == 0 {
		return false, nil, nil
	}
	if len(found) != 1 {
		return false, nil, fmt.Errorf("found multiple instances with id: %q", r.ID)
	}
	i := found[0]
	if i.State != nil {
		stateName := aws.StringValue(i.State.Name)
		switch stateName {
		case "terminated":
			return false, nil, nil

		case "running":
			// Fine
			glog.V(4).Infof("instance %q has state=%q", r.ID, stateName)

		default:
			glog.Infof("unknown instance state for %q: %q", r.ID, stateName)
		}
	}

	var blocks []string
	for _, volume := range i.BlockDeviceMappings {
		if volume.Ebs == nil {
			continue
		}
		blocks = append(blocks, "volume:"+aws.StringValue(volume.Ebs.VolumeId))
	}
	for _, sg := range i.SecurityGroups {
		blocks = append(blocks, "security-group:"+aws.StringValue(sg.GroupId))
	}
	blocks = append(blocks, "subnet:"+aws.StringValue(i.SubnetId))
	blocks = append(blocks, "vpc:"+aws.StringValue(i.VpcId))

	return true, blocks, nil
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
		if IsDependencyViolation(err) {
			// Don't wrap
			return err
		}
		if AWSErrorCode(err) == "InvalidVolume.NotFound" {
			// Concurrently deleted
			return nil
		}
		return fmt.Errorf("error deleting volume %q: %v", r.ID, err)
	}
	return nil
}

func (r *DeletableVolume) String() string {
	return "Volume:" + r.ID
}

func (r *DeletableVolume) Status(cloud fi.Cloud) (bool, []string, error) {
	c := cloud.(*awsup.AWSCloud)

	glog.V(2).Infof("Querying EC2 volume %q", r.ID)
	request := &ec2.DescribeVolumesInput{
		VolumeIds: []*string{&r.ID},
	}
	response, err := c.EC2.DescribeVolumes(request)
	if err != nil {
		if AWSErrorCode(err) == "InvalidVolume.NotFound" {
			return false, nil, nil
		}

		return false, nil, fmt.Errorf("error describing volume %q: %v", r.ID, err)
	}

	var found []*ec2.Volume
	for _, v := range response.Volumes {
		if aws.StringValue(v.VolumeId) == r.ID {
			found = append(found, v)
		}
	}
	if len(found) == 0 {
		return false, nil, nil
	}
	if len(found) != 1 {
		return false, nil, fmt.Errorf("found multiple volumes with id: %q", r.ID)
	}
	//v := found[0]

	var blocks []string

	return true, blocks, nil
}

type DeletableSubnet struct {
	ID string
}

// AWSErrorCode extracts the
func AWSErrorCode(err error) string {
	if awsError, ok := err.(awserr.Error); ok {
		return awsError.Code()
	}
	return ""
}

func IsDependencyViolation(err error) bool {
	code := AWSErrorCode(err)
	switch code {
	case "":
		return false
	case "DependencyViolation", "VolumeInUse":
		return true
	default:
		glog.Infof("unexpected aws error code: %q", code)
		return false
	}
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

func (r *DeletableSubnet) Status(cloud fi.Cloud) (bool, []string, error) {
	c := cloud.(*awsup.AWSCloud)

	glog.V(2).Infof("Querying EC2 subnet %q", r.ID)
	request := &ec2.DescribeSubnetsInput{
		SubnetIds: []*string{&r.ID},
	}
	response, err := c.EC2.DescribeSubnets(request)
	if err != nil {
		return false, nil, fmt.Errorf("error describing subnet %q: %v", r.ID, err)
	}

	var found []*ec2.Subnet
	for _, v := range response.Subnets {
		if aws.StringValue(v.SubnetId) == r.ID {
			found = append(found, v)
		}
	}
	if len(found) == 0 {
		return false, nil, nil
	}
	if len(found) != 1 {
		return false, nil, fmt.Errorf("found multiple subnets with id: %q", r.ID)
	}
	n := found[0]

	var blocks []string
	blocks = append(blocks, "vpc:"+aws.StringValue(n.VpcId))

	return true, blocks, nil
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

func (r *DeletableVPC) Status(cloud fi.Cloud) (bool, []string, error) {
	c := cloud.(*awsup.AWSCloud)

	glog.V(2).Infof("Querying EC2 VPC %q", r.ID)
	request := &ec2.DescribeVpcsInput{
		VpcIds: []*string{&r.ID},
	}
	response, err := c.EC2.DescribeVpcs(request)
	if err != nil {
		return false, nil, fmt.Errorf("error describing VPC %q: %v", r.ID, err)
	}

	var found []*ec2.Vpc
	for _, v := range response.Vpcs {
		if aws.StringValue(v.VpcId) == r.ID {
			found = append(found, v)
		}
	}
	if len(found) == 0 {
		return false, nil, nil
	}
	if len(found) != 1 {
		return false, nil, fmt.Errorf("found multiple VPCs with id: %q", r.ID)
	}
	v := found[0]

	var blocks []string
	blocks = append(blocks, "dhcp-options:"+aws.StringValue(v.DhcpOptionsId))
	return true, blocks, nil
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

func (r *DeletableELBLoadBalancer) Status(cloud fi.Cloud) (bool, []string, error) {
	c := cloud.(*awsup.AWSCloud)

	glog.V(2).Infof("Querying LoadBalancer instance %q", r.Name)
	request := &elb.DescribeLoadBalancersInput{
		LoadBalancerNames: []*string{&r.Name},
	}
	response, err := c.ELB.DescribeLoadBalancers(request)
	if err != nil {
		return false, nil, fmt.Errorf("error describing LoadBalancer %q: %v", r.Name, err)
	}

	var found []*elb.LoadBalancerDescription
	for _, l := range response.LoadBalancerDescriptions {
		if aws.StringValue(l.LoadBalancerName) == r.Name {
			found = append(found, l)
		}
	}
	if len(found) == 0 {
		return false, nil, nil
	}
	if len(found) != 1 {
		return false, nil, fmt.Errorf("found multiple LoadBalancers with Name: %q", r.Name)
	}
	l := found[0]

	var blocks []string
	for _, sg := range l.SecurityGroups {
		blocks = append(blocks, "security-group:"+aws.StringValue(sg))
	}
	for _, s := range l.Subnets {
		blocks = append(blocks, "subnet:"+aws.StringValue(s))
	}
	blocks = append(blocks, "vpc:"+aws.StringValue(l.VPCId))

	return true, blocks, nil
}
