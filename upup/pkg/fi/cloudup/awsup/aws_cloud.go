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

package awsup

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/route53/route53iface"
	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kubernetes/federation/pkg/dnsprovider"
	dnsproviderroute53 "k8s.io/kubernetes/federation/pkg/dnsprovider/providers/aws/route53"
	"strings"
	"time"
)

// By default, aws-sdk-go only retries 3 times, which doesn't give
// much time for exponential backoff to work for serious issues. At 13
// retries, we'll try a given request for up to ~6m with exponential
// backoff along the way.
const ClientMaxRetries = 13

const DescribeTagsMaxAttempts = 120
const DescribeTagsRetryInterval = 2 * time.Second
const DescribeTagsLogInterval = 10 // this is in "retry intervals"

const CreateTagsMaxAttempts = 120
const CreateTagsRetryInterval = 2 * time.Second
const CreateTagsLogInterval = 10 // this is in "retry intervals"

const DeleteTagsMaxAttempts = 120
const DeleteTagsRetryInterval = 2 * time.Second
const DeleteTagsLogInterval = 10 // this is in "retry intervals"

const TagClusterName = "KubernetesCluster"
const TagNameRolePrefix = "k8s.io/role/"
const TagNameEtcdClusterPrefix = "k8s.io/etcd/"

const TagRoleMaster = "master"

const (
	WellKnownAccountKopeio = "383156758163"
	WellKnownAccountRedhat = "309956199498"
	WellKnownAccountCoreOS = "595879546273"
)

type AWSCloud interface {
	fi.Cloud

	Region() string

	CloudFormation() *cloudformation.CloudFormation
	EC2() ec2iface.EC2API
	IAM() *iam.IAM
	ELB() *elb.ELB
	Autoscaling() autoscalingiface.AutoScalingAPI
	Route53() route53iface.Route53API

	// TODO: Document and rationalize these tags/filters methods
	AddTags(name *string, tags map[string]string)
	BuildFilters(name *string) []*ec2.Filter
	BuildTags(name *string) map[string]string
	Tags() map[string]string

	// GetTags will fetch the tags for the specified resource, retrying (up to MaxDescribeTagsAttempts) if it hits an eventual-consistency type error
	GetTags(resourceId string) (map[string]string, error)

	// CreateTags will add tags to the specified resource, retrying up to MaxCreateTagsAttempts times if it hits an eventual-consistency type error
	CreateTags(resourceId string, tags map[string]string) error

	AddAWSTags(id string, expected map[string]string) error
	GetELBTags(loadBalancerName string) (map[string]string, error)

	// CreateELBTags will add tags to the specified loadBalancer, retrying up to MaxCreateTagsAttempts times if it hits an eventual-consistency type error
	CreateELBTags(loadBalancerName string, tags map[string]string) error

	// DescribeInstance is a helper that queries for the specified instance by id
	DescribeInstance(instanceID string) (*ec2.Instance, error)

	// DescribeVPC is a helper that queries for the specified vpc by id
	DescribeVPC(vpcID string) (*ec2.Vpc, error)

	DescribeAvailabilityZones() ([]*ec2.AvailabilityZone, error)

	// ResolveImage finds an AMI image based on the given name.
	// The name can be one of:
	// `ami-...` in which case it is presumed to be an id
	// owner/name in which case we find the image with the specified name, owned by owner
	// name in which case we find the image with the specified name, with the current owner
	ResolveImage(name string) (*ec2.Image, error)

	// WithTags created a copy of AWSCloud with the specified default-tags bound
	WithTags(tags map[string]string) AWSCloud

	// DefaultInstanceType determines a suitable instance type for the specified instance group
	DefaultInstanceType(cluster *kops.Cluster, ig *kops.InstanceGroup) (string, error)
}

type awsCloudImplementation struct {
	cf          *cloudformation.CloudFormation
	ec2         *ec2.EC2
	iam         *iam.IAM
	elb         *elb.ELB
	autoscaling *autoscaling.AutoScaling
	route53     *route53.Route53

	region string

	tags map[string]string
}

var _ fi.Cloud = &awsCloudImplementation{}

func (c *awsCloudImplementation) ProviderID() kops.CloudProviderID {
	return kops.CloudProviderAWS
}

func (c *awsCloudImplementation) Region() string {
	return c.region
}

var awsCloudInstances map[string]AWSCloud = make(map[string]AWSCloud)

func NewAWSCloud(region string, tags map[string]string) (AWSCloud, error) {
	raw := awsCloudInstances[region]
	if raw == nil {
		c := &awsCloudImplementation{region: region}

		config := aws.NewConfig().WithRegion(region)

		// Add some logging of retries
		config.Retryer = newLoggingRetryer(ClientMaxRetries)

		// This avoids a confusing error message when we fail to get credentials
		// e.g. https://github.com/kubernetes/kops/issues/605
		config = config.WithCredentialsChainVerboseErrors(true)

		requestLogger := newRequestLogger(2)

		c.cf = cloudformation.New(session.New(), config)
		c.cf.Handlers.Send.PushFront(requestLogger)

		c.ec2 = ec2.New(session.New(), config)
		c.ec2.Handlers.Send.PushFront(requestLogger)

		c.iam = iam.New(session.New(), config)
		c.iam.Handlers.Send.PushFront(requestLogger)

		c.elb = elb.New(session.New(), config)
		c.elb.Handlers.Send.PushFront(requestLogger)

		c.autoscaling = autoscaling.New(session.New(), config)
		c.autoscaling.Handlers.Send.PushFront(requestLogger)

		c.route53 = route53.New(session.New(), config)
		c.route53.Handlers.Send.PushFront(requestLogger)

		awsCloudInstances[region] = c
		raw = c
	}

	i := raw.WithTags(tags)

	return i, nil
}

func NewEC2Filter(name string, values ...string) *ec2.Filter {
	awsValues := []*string{}
	for _, value := range values {
		awsValues = append(awsValues, aws.String(value))
	}
	filter := &ec2.Filter{
		Name:   aws.String(name),
		Values: awsValues,
	}
	return filter
}

func (c *awsCloudImplementation) Tags() map[string]string {
	// Defensive copy
	tags := make(map[string]string)
	for k, v := range c.tags {
		tags[k] = v
	}
	return tags
}

func (c *awsCloudImplementation) WithTags(tags map[string]string) AWSCloud {
	i := &awsCloudImplementation{}
	*i = *c
	i.tags = tags
	return i
}

var tagsEventualConsistencyErrors = map[string]bool{
	"InvalidInstanceID.NotFound":        true,
	"InvalidRouteTableID.NotFound":      true,
	"InvalidVpcID.NotFound":             true,
	"InvalidGroup.NotFound":             true,
	"InvalidSubnetID.NotFound":          true,
	"InvalidDhcpOptionsID.NotFound":     true,
	"InvalidInternetGatewayID.NotFound": true,
}

// isTagsEventualConsistencyError checks if the error is one of the errors encountered when we try to create/get tags before the resource has fully 'propagated' in EC2
func isTagsEventualConsistencyError(err error) bool {
	if awsErr, ok := err.(awserr.Error); ok {
		isEventualConsistency, found := tagsEventualConsistencyErrors[awsErr.Code()]
		if found {
			return isEventualConsistency
		}

		glog.Warningf("Uncategorized error in isTagsEventualConsistencyError: %v", awsErr.Code())
	}
	return false
}

// GetTags will fetch the tags for the specified resource, retrying (up to MaxDescribeTagsAttempts) if it hits an eventual-consistency type error
func (c *awsCloudImplementation) GetTags(resourceID string) (map[string]string, error) {
	return getTags(c, resourceID)
}

func getTags(c AWSCloud, resourceId string) (map[string]string, error) {
	if resourceId == "" {
		return nil, fmt.Errorf("resourceId not provided to getTags")
	}

	tags := map[string]string{}

	request := &ec2.DescribeTagsInput{
		Filters: []*ec2.Filter{
			NewEC2Filter("resource-id", resourceId),
		},
	}

	attempt := 0
	for {
		attempt++

		response, err := c.EC2().DescribeTags(request)
		if err != nil {
			if isTagsEventualConsistencyError(err) {
				if attempt > DescribeTagsMaxAttempts {
					return nil, fmt.Errorf("Got retryable error while getting tags on %q, but retried too many times without success: %v", resourceId, err)
				}

				if (attempt % DescribeTagsLogInterval) == 0 {
					glog.Infof("waiting for eventual consistency while describing tags on %q", resourceId)
				}

				glog.V(2).Infof("will retry after encountering error getting tags on %q: %v", resourceId, err)
				time.Sleep(DescribeTagsRetryInterval)
				continue
			}

			return nil, fmt.Errorf("error listing tags on %v: %v", resourceId, err)
		}

		for _, tag := range response.Tags {
			if tag == nil {
				glog.Warning("unexpected nil tag")
				continue
			}
			tags[aws.StringValue(tag.Key)] = aws.StringValue(tag.Value)
		}

		return tags, nil
	}
}

// CreateTags will add tags to the specified resource, retrying up to MaxCreateTagsAttempts times if it hits an eventual-consistency type error
func (c *awsCloudImplementation) CreateTags(resourceId string, tags map[string]string) error {
	return createTags(c, resourceId, tags)
}

func createTags(c AWSCloud, resourceId string, tags map[string]string) error {
	if len(tags) == 0 {
		return nil
	}

	ec2Tags := []*ec2.Tag{}
	for k, v := range tags {
		ec2Tags = append(ec2Tags, &ec2.Tag{Key: aws.String(k), Value: aws.String(v)})
	}

	attempt := 0
	for {
		attempt++

		request := &ec2.CreateTagsInput{
			Tags:      ec2Tags,
			Resources: []*string{&resourceId},
		}

		_, err := c.EC2().CreateTags(request)
		if err != nil {
			if isTagsEventualConsistencyError(err) {
				if attempt > CreateTagsMaxAttempts {
					return fmt.Errorf("Got retryable error while creating tags on %q, but retried too many times without success: %v", resourceId, err)
				}

				if (attempt % CreateTagsLogInterval) == 0 {
					glog.Infof("waiting for eventual consistency while creating tags on %q", resourceId)
				}

				glog.V(2).Infof("will retry after encountering error creating tags on %q: %v", resourceId, err)
				time.Sleep(CreateTagsRetryInterval)
				continue
			}

			return fmt.Errorf("error creating tags on %v: %v", resourceId, err)
		}

		return nil
	}
}

// DeleteTags will remove tags from the specified resource, retrying up to MaxCreateTagsAttempts times if it hits an eventual-consistency type error
func (c *awsCloudImplementation) DeleteTags(resourceId string, tags map[string]string) error {
	if len(tags) == 0 {
		return nil
	}

	ec2Tags := []*ec2.Tag{}
	for k, v := range tags {
		ec2Tags = append(ec2Tags, &ec2.Tag{Key: aws.String(k), Value: aws.String(v)})
	}

	attempt := 0
	for {
		attempt++

		request := &ec2.DeleteTagsInput{
			Tags:      ec2Tags,
			Resources: []*string{&resourceId},
		}

		_, err := c.EC2().DeleteTags(request)
		if err != nil {
			if isTagsEventualConsistencyError(err) {
				if attempt > DeleteTagsMaxAttempts {
					return fmt.Errorf("Got retryable error while deleting tags on %q, but retried too many times without success: %v", resourceId, err)
				}

				if (attempt % DeleteTagsLogInterval) == 0 {
					glog.Infof("waiting for eventual consistency while deleting tags on %q", resourceId)
				}

				glog.V(2).Infof("will retry after encountering error deleting tags on %q: %v", resourceId, err)
				time.Sleep(DeleteTagsRetryInterval)
				continue
			}

			return fmt.Errorf("error deleting tags on %v: %v", resourceId, err)
		}

		return nil
	}
}

func (c *awsCloudImplementation) AddAWSTags(id string, expected map[string]string) error {
	return addAWSTags(c, id, expected)
}

func addAWSTags(c AWSCloud, id string, expected map[string]string) error {
	actual, err := c.GetTags(id)
	if err != nil {
		return fmt.Errorf("unexpected error fetching tags for resource: %v", err)
	}

	missing := map[string]string{}
	for k, v := range expected {
		actualValue, found := actual[k]
		if found && actualValue == v {
			continue
		}
		missing[k] = v
	}

	if len(missing) != 0 {
		glog.V(4).Infof("adding tags to %q: %v", id, missing)

		err := c.CreateTags(id, missing)
		if err != nil {
			return fmt.Errorf("error adding tags to resource %q: %v", id, err)
		}
	}

	return nil
}

func (c *awsCloudImplementation) GetELBTags(loadBalancerName string) (map[string]string, error) {
	tags := map[string]string{}

	request := &elb.DescribeTagsInput{
		LoadBalancerNames: []*string{&loadBalancerName},
	}

	attempt := 0
	for {
		attempt++

		response, err := c.ELB().DescribeTags(request)
		if err != nil {
			return nil, fmt.Errorf("error listing tags on %v: %v", loadBalancerName, err)
		}

		for _, tagset := range response.TagDescriptions {
			for _, tag := range tagset.Tags {
				tags[aws.StringValue(tag.Key)] = aws.StringValue(tag.Value)
			}
		}

		return tags, nil
	}
}

// CreateELBTags will add tags to the specified loadBalancer, retrying up to MaxCreateTagsAttempts times if it hits an eventual-consistency type error
func (c *awsCloudImplementation) CreateELBTags(loadBalancerName string, tags map[string]string) error {
	if len(tags) == 0 {
		return nil
	}

	elbTags := []*elb.Tag{}
	for k, v := range tags {
		elbTags = append(elbTags, &elb.Tag{Key: aws.String(k), Value: aws.String(v)})
	}

	attempt := 0
	for {
		attempt++

		request := &elb.AddTagsInput{
			Tags:              elbTags,
			LoadBalancerNames: []*string{&loadBalancerName},
		}

		_, err := c.ELB().AddTags(request)
		if err != nil {
			return fmt.Errorf("error creating tags on %v: %v", loadBalancerName, err)
		}

		return nil
	}
}

func (c *awsCloudImplementation) BuildTags(name *string) map[string]string {
	return buildTags(c.tags, name)
}

func buildTags(commonTags map[string]string, name *string) map[string]string {
	tags := make(map[string]string)
	if name != nil {
		tags["Name"] = *name
	} else {
		glog.Warningf("Name not set when filtering by name")
	}
	for k, v := range commonTags {
		tags[k] = v
	}
	return tags
}

func (c *awsCloudImplementation) AddTags(name *string, tags map[string]string) {
	if name != nil {
		tags["Name"] = *name
	}
	for k, v := range c.tags {
		tags[k] = v
	}
}

func (c *awsCloudImplementation) BuildFilters(name *string) []*ec2.Filter {
	return buildFilters(c.tags, name)
}

func buildFilters(commonTags map[string]string, name *string) []*ec2.Filter {
	filters := []*ec2.Filter{}

	merged := make(map[string]string)
	if name != nil {
		merged["Name"] = *name
	} else {
		glog.Warningf("Name not set when filtering by name")
	}
	for k, v := range commonTags {
		merged[k] = v
	}

	for k, v := range merged {
		filter := NewEC2Filter("tag:"+k, v)
		filters = append(filters, filter)
	}
	return filters
}

// DescribeInstance is a helper that queries for the specified instance by id
func (t *awsCloudImplementation) DescribeInstance(instanceID string) (*ec2.Instance, error) {
	glog.V(2).Infof("Calling DescribeInstances for instance %q", instanceID)
	request := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{&instanceID},
	}

	response, err := t.EC2().DescribeInstances(request)
	if err != nil {
		return nil, fmt.Errorf("error listing Instances: %v", err)
	}
	if response == nil || len(response.Reservations) == 0 {
		return nil, nil
	}
	if len(response.Reservations) != 1 {
		glog.Fatalf("found multiple Reservations for %q", instanceID)
	}

	reservation := response.Reservations[0]
	if len(reservation.Instances) == 0 {
		return nil, nil
	}

	if len(reservation.Instances) != 1 {
		return nil, fmt.Errorf("found multiple Instances for %q", instanceID)
	}

	instance := reservation.Instances[0]
	return instance, nil
}

// DescribeVPC is a helper that queries for the specified vpc by id
func (t *awsCloudImplementation) DescribeVPC(vpcID string) (*ec2.Vpc, error) {
	glog.V(2).Infof("Calling DescribeVPC for VPC %q", vpcID)
	request := &ec2.DescribeVpcsInput{
		VpcIds: []*string{&vpcID},
	}

	response, err := t.EC2().DescribeVpcs(request)
	if err != nil {
		return nil, fmt.Errorf("error listing VPCs: %v", err)
	}
	if response == nil || len(response.Vpcs) == 0 {
		return nil, nil
	}
	if len(response.Vpcs) != 1 {
		return nil, fmt.Errorf("found multiple VPCs for %q", vpcID)
	}

	vpc := response.Vpcs[0]
	return vpc, nil
}

// ResolveImage finds an AMI image based on the given name.
// The name can be one of:
// `ami-...` in which case it is presumed to be an id
// owner/name in which case we find the image with the specified name, owned by owner
// name in which case we find the image with the specified name, with the current owner
func (c *awsCloudImplementation) ResolveImage(name string) (*ec2.Image, error) {
	return resolveImage(c.ec2, name)
}

func resolveImage(ec2Client ec2iface.EC2API, name string) (*ec2.Image, error) {
	// TODO: Cache this result during a single execution (we get called multiple times)
	glog.V(2).Infof("Calling DescribeImages to resolve name %q", name)
	request := &ec2.DescribeImagesInput{}

	if strings.HasPrefix(name, "ami-") {
		// ami-xxxxxxxx
		request.ImageIds = []*string{&name}
	} else {
		// Either <imagename> or <owner>/<imagename>
		tokens := strings.SplitN(name, "/", 2)
		if len(tokens) == 1 {
			// self is a well-known value in the DescribeImages call
			request.Owners = aws.StringSlice([]string{"self"})
			request.Filters = append(request.Filters, NewEC2Filter("name", name))
		} else if len(tokens) == 2 {
			owner := tokens[0]

			// Check for well known owner aliases
			switch owner {
			case "kope.io":
				owner = WellKnownAccountKopeio
			case "coreos.com":
				owner = WellKnownAccountCoreOS
			case "redhat.com":
				owner = WellKnownAccountRedhat
			}

			request.Owners = []*string{&owner}
			request.Filters = append(request.Filters, NewEC2Filter("name", tokens[1]))
		} else {
			return nil, fmt.Errorf("image name specification not recognized: %q", name)
		}
	}

	response, err := ec2Client.DescribeImages(request)
	if err != nil {
		return nil, fmt.Errorf("error listing images: %v", err)
	}
	if response == nil || len(response.Images) == 0 {
		return nil, fmt.Errorf("could not find Image for %q", name)
	}
	if len(response.Images) != 1 {
		return nil, fmt.Errorf("found multiple Images for %q", name)
	}

	image := response.Images[0]
	glog.V(4).Infof("Resolved image %q", aws.StringValue(image.ImageId))
	return image, nil
}

func (c *awsCloudImplementation) DescribeAvailabilityZones() ([]*ec2.AvailabilityZone, error) {
	glog.V(2).Infof("Querying EC2 for all valid zones in region %q", c.region)

	request := &ec2.DescribeAvailabilityZonesInput{}
	response, err := c.EC2().DescribeAvailabilityZones(request)
	if err != nil {
		return nil, fmt.Errorf("error querying for valid AZs in %q - verify your AWS credentials.  Error: %v", c.region, err)
	}

	return response.AvailabilityZones, nil
}

// ValidateZones checks that every zone in the sliced passed is recognized
func ValidateZones(zones []string, cloud AWSCloud) error {
	azs, err := cloud.DescribeAvailabilityZones()
	if err != nil {
		return err
	}

	zoneMap := make(map[string]*ec2.AvailabilityZone)
	for _, z := range azs {
		name := aws.StringValue(z.ZoneName)
		zoneMap[name] = z
	}

	for _, zone := range zones {
		z := zoneMap[zone]
		if z == nil {
			var knownZones []string
			for z := range zoneMap {
				knownZones = append(knownZones, z)
			}

			glog.Infof("Known zones: %q", strings.Join(knownZones, ","))
			return fmt.Errorf("Zone is not a recognized AZ: %q (check you have specified a valid zone?)", zone)
		}

		for _, message := range z.Messages {
			glog.Warningf("Zone %q has message: %q", aws.StringValue(message.Message))
		}

		if aws.StringValue(z.State) != "available" {
			glog.Warningf("Zone %q has state %q", aws.StringValue(z.State))
		}
	}

	return nil
}

func (c *awsCloudImplementation) DNS() (dnsprovider.Interface, error) {
	provider, err := dnsprovider.GetDnsProvider(dnsproviderroute53.ProviderName, nil)
	if err != nil {
		return nil, fmt.Errorf("Error building (k8s) DNS provider: %v", err)
	}
	return provider, nil
}

func (c *awsCloudImplementation) CloudFormation() *cloudformation.CloudFormation {
	return c.cf
}

func (c *awsCloudImplementation) EC2() ec2iface.EC2API {
	return c.ec2
}

func (c *awsCloudImplementation) IAM() *iam.IAM {
	return c.iam
}

func (c *awsCloudImplementation) ELB() *elb.ELB {
	return c.elb
}

func (c *awsCloudImplementation) Autoscaling() autoscalingiface.AutoScalingAPI {
	return c.autoscaling
}

func (c *awsCloudImplementation) Route53() route53iface.Route53API {
	return c.route53
}

func (c *awsCloudImplementation) FindVPCInfo(vpcID string) (*fi.VPCInfo, error) {
	vpc, err := c.DescribeVPC(vpcID)
	if err != nil {
		return nil, err
	}
	if vpc == nil {
		return nil, nil
	}

	vpcInfo := &fi.VPCInfo{
		CIDR: aws.StringValue(vpc.CidrBlock),
	}

	// Find subnets in the VPC
	{
		glog.V(2).Infof("Calling DescribeSubnets for subnets in VPC %q", vpcID)
		request := &ec2.DescribeSubnetsInput{
			Filters: []*ec2.Filter{NewEC2Filter("vpc-id", vpcID)},
		}

		response, err := c.ec2.DescribeSubnets(request)
		if err != nil {
			return nil, fmt.Errorf("error listing subnets in VPC %q: %v", vpcID, err)
		}
		if response != nil {
			for _, subnet := range response.Subnets {
				subnetInfo := &fi.SubnetInfo{
					ID:   aws.StringValue(subnet.SubnetId),
					CIDR: aws.StringValue(subnet.CidrBlock),
					Zone: aws.StringValue(subnet.AvailabilityZone),
				}

				vpcInfo.Subnets = append(vpcInfo.Subnets, subnetInfo)
			}
		}
	}

	return vpcInfo, nil
}

// DefaultInstanceType determines an instance type for the specified cluster & instance group
func (c *awsCloudImplementation) DefaultInstanceType(cluster *kops.Cluster, ig *kops.InstanceGroup) (string, error) {
	var candidates []string

	switch ig.Spec.Role {
	case kops.InstanceGroupRoleMaster:
		// Some regions do not (currently) support the m3 family; the c4 large is the cheapest non-burstable instance
		// (us-east-2, ca-central-1, eu-west-2, ap-northeast-2).
		// Also some accounts are no longer supporting m3 in us-east-1 zones
		candidates = []string{"m3.medium", "c4.large"}

	case kops.InstanceGroupRoleNode:
		candidates = []string{"t2.medium"}

	case kops.InstanceGroupRoleBastion:
		candidates = []string{"t2.micro"}

	default:
		return "", fmt.Errorf("unhandled role %q", ig.Spec.Role)
	}

	// Find the AZs the InstanceGroup targets
	igZones := sets.NewString()
	for _, subnetName := range ig.Spec.Subnets {
		var subnet *kops.ClusterSubnetSpec
		for i := range cluster.Spec.Subnets {
			if cluster.Spec.Subnets[i].Name == subnetName {
				subnet = &cluster.Spec.Subnets[i]
			}
		}
		if subnet == nil {
			return "", fmt.Errorf("subnet %q is not defined in cluster", subnetName)
		}
		igZones.Insert(subnet.Zone)
	}

	// TODO: Validate that instance type exists in all AZs, but skip AZs that don't support any VPC stuff
	for _, instanceType := range candidates {
		zones, err := c.zonesWithInstanceType(instanceType)
		if err != nil {
			return "", err
		}
		if zones.IsSuperset(igZones) {
			return instanceType, nil
		} else {
			glog.V(2).Infof("can't use instance type %q, available in zones %v but need %v", instanceType, zones, igZones)
		}
	}

	return "", fmt.Errorf("could not find a suitable supported instance type for the instance group %q (type %q) in region %q", ig.Name, ig.Spec.Role, c.region)
}

// supportsInstanceType uses the DescribeReservedInstancesOfferings API call to determine if an instance type is supported in a region
func (c *awsCloudImplementation) zonesWithInstanceType(instanceType string) (sets.String, error) {
	glog.V(4).Infof("checking if instance type %q is supported in region %q", instanceType, c.region)
	request := &ec2.DescribeReservedInstancesOfferingsInput{}
	request.InstanceTenancy = aws.String("default")
	request.IncludeMarketplace = aws.Bool(false)
	request.OfferingClass = aws.String("standard")
	request.OfferingType = aws.String("No Upfront")
	request.ProductDescription = aws.String("Linux/UNIX (Amazon VPC)")
	request.InstanceType = aws.String(instanceType)

	zones := sets.NewString()

	response, err := c.ec2.DescribeReservedInstancesOfferings(request)
	if err != nil {
		return zones, fmt.Errorf("error checking if instance type %q is supported in region %q: %v", instanceType, c.region, err)
	}

	for _, item := range response.ReservedInstancesOfferings {
		if aws.StringValue(item.InstanceType) == instanceType {
			zones.Insert(aws.StringValue(item.AvailabilityZone))
		} else {
			glog.Warningf("skipping non-matching instance type offering: %v", item)
		}
	}

	return zones, nil
}
