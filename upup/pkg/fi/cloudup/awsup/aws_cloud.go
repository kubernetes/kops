package awsup

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/fi"
	"strings"
	"time"
)

const MaxDescribeTagsAttempts = 60
const MaxCreateTagsAttempts = 60

const TagClusterName = "KubernetesCluster"

type AWSCloud struct {
	EC2         *ec2.EC2
	IAM         *iam.IAM
	ELB         *elb.ELB
	Autoscaling *autoscaling.AutoScaling
	Route53     *route53.Route53

	Region string

	tags map[string]string
}

var _ fi.Cloud = &AWSCloud{}

func (c *AWSCloud) ProviderID() fi.CloudProviderID {
	return fi.CloudProviderAWS
}

func NewAWSCloud(region string, tags map[string]string) (*AWSCloud, error) {
	c := &AWSCloud{Region: region}

	config := aws.NewConfig().WithRegion(region)
	c.EC2 = ec2.New(session.New(), config)
	c.IAM = iam.New(session.New(), config)
	c.ELB = elb.New(session.New(), config)
	c.Autoscaling = autoscaling.New(session.New(), config)
	c.Route53 = route53.New(session.New(), config)

	c.tags = tags
	return c, nil
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

func (c *AWSCloud) Tags() map[string]string {
	// Defensive copy
	tags := make(map[string]string)
	for k, v := range c.tags {
		tags[k] = v
	}
	return tags
}

// isTagsEventualConsistencyError checks if the error is one of the errors encountered when we try to create/get tags before the resource has fully 'propagated' in EC2
func isTagsEventualConsistencyError(err error) bool {
	if awsErr, ok := err.(awserr.Error); ok {
		switch awsErr.Code() {
		case "InvalidInstanceID.NotFound", "InvalidRouteTableID.NotFound", "InvalidVpcID.NotFound", "InvalidGroup.NotFound", "InvalidSubnetID.NotFound", "InvalidInternetGatewayID.NotFound", "InvalidDhcpOptionsID.NotFound":
			return true

		default:
			glog.Warningf("Uncategorized error in isTagsEventualConsistencyError: %v", awsErr.Code())
		}
	}
	return false
}

// GetTags will fetch the tags for the specified resource, retrying (up to MaxDescribeTagsAttempts) if it hits an eventual-consistency type error
func (c *AWSCloud) GetTags(resourceId string) (map[string]string, error) {
	tags := map[string]string{}

	request := &ec2.DescribeTagsInput{
		Filters: []*ec2.Filter{
			NewEC2Filter("resource-id", resourceId),
		},
	}

	attempt := 0
	for {
		attempt++

		response, err := c.EC2.DescribeTags(request)
		if err != nil {
			if isTagsEventualConsistencyError(err) {
				if attempt > MaxDescribeTagsAttempts {
					return nil, fmt.Errorf("Got retryable error while getting tags on %q, but retried too many times without success: %v", resourceId, err)
				}

				glog.V(2).Infof("will retry after encountering error getting tags on %q: %v", resourceId, err)
				time.Sleep(2 * time.Second)
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
func (c *AWSCloud) CreateTags(resourceId string, tags map[string]string) error {
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

		_, err := c.EC2.CreateTags(request)
		if err != nil {
			if isTagsEventualConsistencyError(err) {
				if attempt > MaxCreateTagsAttempts {
					return fmt.Errorf("Got retryable error while creating tags on %q, but retried too many times without success: %v", resourceId, err)
				}

				glog.V(2).Infof("will retry after encountering error creating tags on %q: %v", resourceId, err)
				time.Sleep(2 * time.Second)
				continue
			}

			return fmt.Errorf("error creating tags on %v: %v", resourceId, err)
		}

		return nil
	}
}

// DeleteTags will remove tags from the specified resource, retrying up to MaxCreateTagsAttempts times if it hits an eventual-consistency type error
func (c *AWSCloud) DeleteTags(resourceId string, tags map[string]string) error {
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

		_, err := c.EC2.DeleteTags(request)
		if err != nil {
			if isTagsEventualConsistencyError(err) {
				if attempt > MaxCreateTagsAttempts {
					return fmt.Errorf("Got retryable error while deleting tags on %q, but retried too many times without success: %v", resourceId, err)
				}

				glog.V(2).Infof("will retry after encountering error deleting tags on %q: %v", resourceId, err)
				time.Sleep(2 * time.Second)
				continue
			}

			return fmt.Errorf("error deleting tags on %v: %v", resourceId, err)
		}

		return nil
	}
}

func (c *AWSCloud) AddAWSTags(id string, expected map[string]string) error {
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

func (c *AWSCloud) GetELBTags(loadBalancerName string) (map[string]string, error) {
	tags := map[string]string{}

	request := &elb.DescribeTagsInput{
		LoadBalancerNames: []*string{&loadBalancerName},
	}

	attempt := 0
	for {
		attempt++

		response, err := c.ELB.DescribeTags(request)
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
func (c *AWSCloud) CreateELBTags(loadBalancerName string, tags map[string]string) error {
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

		_, err := c.ELB.AddTags(request)
		if err != nil {
			return fmt.Errorf("error creating tags on %v: %v", loadBalancerName, err)
		}

		return nil
	}
}

func (c *AWSCloud) BuildTags(name *string) map[string]string {
	tags := make(map[string]string)
	if name != nil {
		tags["Name"] = *name
	} else {
		glog.Warningf("Name not set when filtering by name")
	}
	for k, v := range c.tags {
		tags[k] = v
	}
	return tags
}

func (c *AWSCloud) AddTags(name *string, tags map[string]string) {
	if name != nil {
		tags["Name"] = *name
	}
	for k, v := range c.tags {
		tags[k] = v
	}
}

func (c *AWSCloud) BuildFilters(name *string) []*ec2.Filter {
	filters := []*ec2.Filter{}

	merged := make(map[string]string)
	if name != nil {
		merged["Name"] = *name
	} else {
		glog.Warningf("Name not set when filtering by name")
	}
	for k, v := range c.tags {
		merged[k] = v
	}

	for k, v := range merged {
		filter := NewEC2Filter("tag:"+k, v)
		filters = append(filters, filter)
	}
	return filters
}

// DescribeInstance is a helper that queries for the specified instance by id
func (t *AWSCloud) DescribeInstance(instanceID string) (*ec2.Instance, error) {
	glog.V(2).Infof("Calling DescribeInstances for instance %q", instanceID)
	request := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{&instanceID},
	}

	response, err := t.EC2.DescribeInstances(request)
	if err != nil {
		return nil, fmt.Errorf("error listing Instances: %v", err)
	}
	if response == nil || len(response.Reservations) == 0 {
		return nil, nil
	}
	if len(response.Reservations) != 1 {
		glog.Fatalf("found multiple Reservations for instance id")
	}

	reservation := response.Reservations[0]
	if len(reservation.Instances) == 0 {
		return nil, nil
	}

	if len(reservation.Instances) != 1 {
		return nil, fmt.Errorf("found multiple Instances for instance id")
	}

	instance := reservation.Instances[0]
	return instance, nil
}

// DescribeVPC is a helper that queries for the specified vpc by id
func (t *AWSCloud) DescribeVPC(vpcID string) (*ec2.Vpc, error) {
	glog.V(2).Infof("Calling DescribeVPC for VPC %q", vpcID)
	request := &ec2.DescribeVpcsInput{
		VpcIds: []*string{&vpcID},
	}

	response, err := t.EC2.DescribeVpcs(request)
	if err != nil {
		return nil, fmt.Errorf("error listing VPCs: %v", err)
	}
	if response == nil || len(response.Vpcs) == 0 {
		return nil, nil
	}
	if len(response.Vpcs) != 1 {
		return nil, fmt.Errorf("found multiple VPCs for instance id")
	}

	vpc := response.Vpcs[0]
	return vpc, nil
}

// ResolveImage finds an AMI image based on the given name.
// The name can be one of:
// `ami-...` in which case it is presumed to be an id
// owner/name in which case we find the image with the specified name, owned by owner
// name in which case we find the image with the specified name, with the current owner
func (c *AWSCloud) ResolveImage(name string) (*ec2.Image, error) {
	// TODO: Cache this result during a single execution (we get called multiple times)
	glog.V(2).Infof("Calling DescribeImages to resolve name %q", name)
	request := &ec2.DescribeImagesInput{}

	if strings.HasPrefix(name, "ami-") {
		// ami-xxxxxxxx
		request.ImageIds = []*string{&name}
	} else {
		// Either <imagename> or <owner>/<imagename>
		tokens := strings.Split(name, "/")
		if len(tokens) == 1 {
			// self is a well-known value in the DescribeImages call
			request.Owners = aws.StringSlice([]string{"self"})
			request.Filters = append(request.Filters, NewEC2Filter("name", name))
		} else if len(tokens) == 2 {
			request.Owners = []*string{&tokens[0]}
			request.Filters = append(request.Filters, NewEC2Filter("name", tokens[1]))
		} else {
			return nil, fmt.Errorf("image name specification not recognized: %q", name)
		}
	}

	response, err := c.EC2.DescribeImages(request)
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

// ValidateZones checks that every zone in the sliced passed is recognized
func (c *AWSCloud) ValidateZones(zones []string) error {
	glog.V(2).Infof("Querying EC2 for all valid zones in region %q", c.Region)

	request := &ec2.DescribeAvailabilityZonesInput{}
	response, err := c.EC2.DescribeAvailabilityZones(request)

	if err != nil {
		return fmt.Errorf("Got an error while querying for valid AZs in %q (verify your AWS credentials?)", c.Region)
	}

	zoneMap := make(map[string]*ec2.AvailabilityZone)
	var knownZones []string
	for _, z := range response.AvailabilityZones {
		name := aws.StringValue(z.ZoneName)
		zoneMap[name] = z
		knownZones = append(knownZones, name)
	}

	for _, zone := range zones {
		z := zoneMap[zone]
		if z == nil {
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

func (c *AWSCloud) FindDNSHostedZone(clusterDNSName string) (string, error) {
	glog.V(2).Infof("Querying for all route53 zones to find match for %q", clusterDNSName)

	clusterDNSName = "." + strings.TrimSuffix(clusterDNSName, ".")

	var zones []*route53.HostedZone
	request := &route53.ListHostedZonesInput{}
	err := c.Route53.ListHostedZonesPages(request, func(p *route53.ListHostedZonesOutput, lastPage bool) bool {
		for _, zone := range p.HostedZones {
			zoneName := aws.StringValue(zone.Name)
			zoneName = "." + strings.TrimSuffix(zoneName, ".")

			if strings.HasSuffix(clusterDNSName, zoneName) {
				zones = append(zones, zone)
			}
		}
		return true
	})
	if err != nil {
		return "", fmt.Errorf("error querying for route53 zones: %v", err)
	}

	// Find the longest zones
	maxLength := -1
	maxLengthZones := []*route53.HostedZone{}
	for _, z := range zones {
		n := len(aws.StringValue(z.Name))
		if n < maxLength {
			continue
		}

		if n > maxLength {
			maxLength = n
			maxLengthZones = []*route53.HostedZone{}
		}

		maxLengthZones = append(maxLengthZones, z)
	}

	if len(maxLengthZones) == 0 {
		// We make this an error because you have to set up DNS delegation anyway
		tokens := strings.Split(clusterDNSName, ".")
		suffix := strings.Join(tokens[len(tokens)-2:], ".")
		//glog.Warningf("No matching hosted zones found; will created %q", suffix)
		//return suffix, nil
		return "", fmt.Errorf("No matching hosted zones found for %q; please create one (e.g. %q) first", clusterDNSName, suffix)
	}

	if len(maxLengthZones) == 1 {
		id := aws.StringValue(maxLengthZones[0].Id)
		id = strings.TrimPrefix(id, "/hostedzone/")
		return id, nil
	}

	return "", fmt.Errorf("Found multiple hosted zones matching cluster %q; please specify the ID of the zone to use")
}
