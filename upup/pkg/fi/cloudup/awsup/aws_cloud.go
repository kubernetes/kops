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
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elb/elbiface"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/aws/aws-sdk-go/service/elbv2/elbv2iface"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/route53/route53iface"
	"k8s.io/klog"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	dnsproviderroute53 "k8s.io/kops/dnsprovider/pkg/dnsprovider/providers/aws/route53"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/model"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/resources/spotinst"
	"k8s.io/kops/upup/pkg/fi"
	k8s_aws "k8s.io/kubernetes/pkg/cloudprovider/providers/aws"
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

// TagNameKopsRole is the AWS tag used to identify the role an object plays for a cluster
const TagNameKopsRole = "kubernetes.io/kops/role"

// TagNameClusterOwnershipPrefix is the AWS tag used for ownership
const TagNameClusterOwnershipPrefix = "kubernetes.io/cluster/"

const (
	WellKnownAccountKopeio             = "383156758163"
	WellKnownAccountRedhat             = "309956199498"
	WellKnownAccountCoreOS             = "595879546273"
	WellKnownAccountAmazonSystemLinux2 = "137112412989"
	WellKnownAccountUbuntu             = "099720109477"
)

type AWSCloud interface {
	fi.Cloud

	Region() string

	CloudFormation() *cloudformation.CloudFormation
	EC2() ec2iface.EC2API
	IAM() iamiface.IAMAPI
	ELB() elbiface.ELBAPI
	ELBV2() elbv2iface.ELBV2API
	Autoscaling() autoscalingiface.AutoScalingAPI
	Route53() route53iface.Route53API
	Spotinst() spotinst.Service

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
	// RemoveELBTags will remove tags from the specified loadBalancer, retrying up to MaxCreateTagsAttempts times if it hits an eventual-consistency type error
	RemoveELBTags(loadBalancerName string, tags map[string]string) error

	// DeleteTags will delete tags from the specified resource, retrying up to MaxCreateTagsAttempts times if it hits an eventual-consistency type error
	DeleteTags(id string, tags map[string]string) error

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

	// FindClusterStatus gets the status of the cluster as it exists in AWS, inferred from volumes
	FindClusterStatus(cluster *kops.Cluster) (*kops.ClusterStatus, error)
}

type awsCloudImplementation struct {
	cf          *cloudformation.CloudFormation
	ec2         *ec2.EC2
	iam         *iam.IAM
	elb         *elb.ELB
	elbv2       *elbv2.ELBV2
	autoscaling *autoscaling.AutoScaling
	route53     *route53.Route53
	spotinst    spotinst.Service

	region string

	tags map[string]string

	regionDelayers *RegionDelayers
}

type RegionDelayers struct {
	mutex      sync.Mutex
	delayerMap map[string]*k8s_aws.CrossRequestRetryDelay
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
		c := &awsCloudImplementation{
			region: region,
			regionDelayers: &RegionDelayers{
				delayerMap: make(map[string]*k8s_aws.CrossRequestRetryDelay),
			},
		}

		config := aws.NewConfig().WithRegion(region)

		// This avoids a confusing error message when we fail to get credentials
		// e.g. https://github.com/kubernetes/kops/issues/605
		config = config.WithCredentialsChainVerboseErrors(true)
		config = request.WithRetryer(config, newLoggingRetryer(ClientMaxRetries))

		// We have the updated aws sdk from 1.9, but don't have https://github.com/kubernetes/kubernetes/pull/55307
		// Set the SleepDelay function to work around this
		// TODO: Remove once we update to k8s >= 1.9 (or a version of the retry delayer than includes this)
		config.SleepDelay = func(d time.Duration) {
			klog.V(6).Infof("aws request sleeping for %v", d)
			time.Sleep(d)
		}

		requestLogger := newRequestLogger(2)

		sess, err := session.NewSession(config)
		if err != nil {
			return c, err
		}
		c.cf = cloudformation.New(sess, config)
		c.cf.Handlers.Send.PushFront(requestLogger)
		c.addHandlers(region, &c.cf.Handlers)

		sess, err = session.NewSession(config)
		if err != nil {
			return c, err
		}
		c.ec2 = ec2.New(sess, config)
		c.ec2.Handlers.Send.PushFront(requestLogger)
		c.addHandlers(region, &c.ec2.Handlers)

		sess, err = session.NewSession(config)
		if err != nil {
			return c, err
		}
		c.iam = iam.New(sess, config)
		c.iam.Handlers.Send.PushFront(requestLogger)
		c.addHandlers(region, &c.iam.Handlers)

		sess, err = session.NewSession(config)
		if err != nil {
			return c, err
		}
		c.elb = elb.New(sess, config)
		c.elb.Handlers.Send.PushFront(requestLogger)
		c.addHandlers(region, &c.elb.Handlers)

		sess, err = session.NewSession(config)
		if err != nil {
			return c, err
		}
		c.elbv2 = elbv2.New(sess, config)
		c.elbv2.Handlers.Send.PushFront(requestLogger)
		c.addHandlers(region, &c.elbv2.Handlers)

		sess, err = session.NewSession(config)
		if err != nil {
			return c, err
		}
		c.autoscaling = autoscaling.New(sess, config)
		c.autoscaling.Handlers.Send.PushFront(requestLogger)
		c.addHandlers(region, &c.autoscaling.Handlers)

		sess, err = session.NewSession(config)
		if err != nil {
			return c, err
		}
		c.route53 = route53.New(sess, config)
		c.route53.Handlers.Send.PushFront(requestLogger)
		c.addHandlers(region, &c.route53.Handlers)

		if featureflag.Spotinst.Enabled() {
			c.spotinst, err = spotinst.NewService(kops.CloudProviderAWS)
			if err != nil {
				return c, err
			}
		}

		awsCloudInstances[region] = c
		raw = c
	}

	i := raw.WithTags(tags)

	return i, nil
}

func (c *awsCloudImplementation) addHandlers(regionName string, h *request.Handlers) {

	delayer := c.getCrossRequestRetryDelay(regionName)
	if delayer != nil {
		h.Sign.PushFrontNamed(request.NamedHandler{
			Name: "kops/delay-presign",
			Fn:   delayer.BeforeSign,
		})

		h.AfterRetry.PushFrontNamed(request.NamedHandler{
			Name: "kops/delay-afterretry",
			Fn:   delayer.AfterRetry,
		})
	}
}

// Get a CrossRequestRetryDelay, scoped to the region, not to the request.
// This means that when we hit a limit on a call, we will delay _all_ calls to the API.
// We do this to protect the AWS account from becoming overloaded and effectively locked.
// We also log when we hit request limits.
// Note that this delays the current goroutine; this is bad behaviour and will
// likely cause kops to become slow or unresponsive for cloud operations.
// However, this throttle is intended only as a last resort.  When we observe
// this throttling, we need to address the root cause (e.g. add a delay to a
// controller retry loop)
func (c *awsCloudImplementation) getCrossRequestRetryDelay(regionName string) *k8s_aws.CrossRequestRetryDelay {
	c.regionDelayers.mutex.Lock()
	defer c.regionDelayers.mutex.Unlock()

	delayer, found := c.regionDelayers.delayerMap[regionName]
	if !found {
		delayer = k8s_aws.NewCrossRequestRetryDelay()
		c.regionDelayers.delayerMap[regionName] = delayer
	}
	return delayer
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

// DeleteGroup deletes an aws autoscaling group
func (c *awsCloudImplementation) DeleteGroup(g *cloudinstances.CloudInstanceGroup) error {
	if c.spotinst != nil {
		return spotinst.DeleteGroup(c.spotinst, g)
	}

	return deleteGroup(c, g)
}

func deleteGroup(c AWSCloud, g *cloudinstances.CloudInstanceGroup) error {
	asg := g.Raw.(*autoscaling.Group)

	name := aws.StringValue(asg.AutoScalingGroupName)
	template := aws.StringValue(asg.LaunchConfigurationName)
	launchTemplate := ""
	if asg.LaunchTemplate != nil {
		launchTemplate = aws.StringValue(asg.LaunchTemplate.LaunchTemplateName)
	}

	// Delete ASG
	{
		klog.V(2).Infof("Deleting autoscaling group %q", name)
		request := &autoscaling.DeleteAutoScalingGroupInput{
			AutoScalingGroupName: aws.String(name),
			ForceDelete:          aws.Bool(true),
		}
		_, err := c.Autoscaling().DeleteAutoScalingGroup(request)
		if err != nil {
			return fmt.Errorf("error deleting autoscaling group %q: %v", name, err)
		}
	}

	// Delete LaunchConfig
	if launchTemplate != "" {
		// Delete launchTemplate
		{
			klog.V(2).Infof("Deleting autoscaling launch template %q", launchTemplate)
			req := &ec2.DeleteLaunchTemplateInput{
				LaunchTemplateName: aws.String(launchTemplate),
			}
			_, err := c.EC2().DeleteLaunchTemplate(req)
			if err != nil {
				return fmt.Errorf("error deleting autoscaling launch template %q: %v", launchTemplate, err)
			}
		}
	} else if template != "" {
		// Delete LaunchConfig
		{
			klog.V(2).Infof("Deleting autoscaling launch configuration %q", template)
			request := &autoscaling.DeleteLaunchConfigurationInput{
				LaunchConfigurationName: aws.String(template),
			}
			_, err := c.Autoscaling().DeleteLaunchConfiguration(request)
			if err != nil {
				return fmt.Errorf("error deleting autoscaling launch configuration %q: %v", template, err)
			}
		}
	}

	klog.V(8).Infof("deleted aws autoscaling group: %q", name)

	return nil
}

// DeleteInstance deletes an aws instance
func (c *awsCloudImplementation) DeleteInstance(i *cloudinstances.CloudInstanceGroupMember) error {
	if c.spotinst != nil {
		return spotinst.DeleteInstance(c.spotinst, i)
	}

	return deleteInstance(c, i)
}

func deleteInstance(c AWSCloud, i *cloudinstances.CloudInstanceGroupMember) error {
	id := i.ID
	if id == "" {
		return fmt.Errorf("id was not set on CloudInstanceGroupMember: %v", i)
	}

	request := &autoscaling.TerminateInstanceInAutoScalingGroupInput{
		InstanceId:                     aws.String(id),
		ShouldDecrementDesiredCapacity: aws.Bool(false),
	}

	if _, err := c.Autoscaling().TerminateInstanceInAutoScalingGroup(request); err != nil {
		return fmt.Errorf("error deleting instance %q: %v", id, err)
	}

	klog.V(8).Infof("deleted aws ec2 instance %q", id)

	return nil
}

// TODO not used yet, as this requires a major refactor of rolling-update code, slowly but surely

// GetCloudGroups returns a groups of instances that back a kops instance groups
func (c *awsCloudImplementation) GetCloudGroups(cluster *kops.Cluster, instancegroups []*kops.InstanceGroup, warnUnmatched bool, nodes []v1.Node) (map[string]*cloudinstances.CloudInstanceGroup, error) {
	if c.spotinst != nil {
		return spotinst.GetCloudGroups(c.spotinst, cluster,
			instancegroups, warnUnmatched, nodes)
	}

	return getCloudGroups(c, cluster, instancegroups, warnUnmatched, nodes)
}

func getCloudGroups(c AWSCloud, cluster *kops.Cluster, instancegroups []*kops.InstanceGroup, warnUnmatched bool, nodes []v1.Node) (map[string]*cloudinstances.CloudInstanceGroup, error) {
	nodeMap := cloudinstances.GetNodeMap(nodes, cluster)

	groups := make(map[string]*cloudinstances.CloudInstanceGroup)
	asgs, err := FindAutoscalingGroups(c, c.Tags())
	if err != nil {
		return nil, fmt.Errorf("unable to find autoscale groups: %v", err)
	}

	for _, asg := range asgs {
		name := aws.StringValue(asg.AutoScalingGroupName)

		instancegroup, err := matchInstanceGroup(name, cluster.ObjectMeta.Name, instancegroups)
		if err != nil {
			return nil, fmt.Errorf("error getting instance group for ASG %q", name)
		}
		if instancegroup == nil {
			if warnUnmatched {
				klog.Warningf("Found ASG with no corresponding instance group %q", name)
			}
			continue
		}

		groups[instancegroup.ObjectMeta.Name], err = awsBuildCloudInstanceGroup(c, instancegroup, asg, nodeMap)
		if err != nil {
			return nil, fmt.Errorf("error getting cloud instance group %q: %v", instancegroup.ObjectMeta.Name, err)
		}
	}

	return groups, nil

}

// FindAutoscalingGroups finds autoscaling groups matching the specified tags
// This isn't entirely trivial because autoscaling doesn't let us filter with as much precision as we would like
func FindAutoscalingGroups(c AWSCloud, tags map[string]string) ([]*autoscaling.Group, error) {
	var asgs []*autoscaling.Group

	klog.V(2).Infof("Listing all Autoscaling groups matching cluster tags")
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

		err := c.Autoscaling().DescribeTagsPages(request, func(p *autoscaling.DescribeTagsOutput, lastPage bool) bool {
			for _, t := range p.Tags {
				switch *t.ResourceType {
				case "auto-scaling-group":
					asgNames = append(asgNames, t.ResourceId)
				default:
					klog.Warningf("Unknown resource type: %v", *t.ResourceType)

				}
			}
			return true
		})
		if err != nil {
			return nil, fmt.Errorf("error listing autoscaling cluster tags: %v", err)
		}
	}

	if len(asgNames) != 0 {
		for i := 0; i < len(asgNames); i += 50 {
			batch := asgNames[i:minInt(i+50, len(asgNames))]
			request := &autoscaling.DescribeAutoScalingGroupsInput{
				AutoScalingGroupNames: batch,
			}
			err := c.Autoscaling().DescribeAutoScalingGroupsPages(request, func(p *autoscaling.DescribeAutoScalingGroupsOutput, lastPage bool) bool {
				for _, asg := range p.AutoScalingGroups {
					if !matchesAsgTags(tags, asg.Tags) {
						// We used an inexact filter above
						continue
					}
					// Check for "Delete in progress" (the only use of .Status)
					if asg.Status != nil {
						klog.Warningf("Skipping ASG %v (which matches tags): %v", *asg.AutoScalingGroupARN, *asg.Status)
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

	}

	return asgs, nil
}

// Returns the minimum of two ints
func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

// matchesAsgTags is used to filter an asg by tags
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

// findAutoscalingGroupLaunchConfiguration is responsible for finding the launch - which could be a launchconfiguration, a template or a mixed instance policy template
func findAutoscalingGroupLaunchConfiguration(c AWSCloud, g *autoscaling.Group) (string, error) {
	name := aws.StringValue(g.LaunchConfigurationName)
	if name != "" {
		return name, nil
	}

	// @check the launch template then
	if g.LaunchTemplate != nil {
		name = aws.StringValue(g.LaunchTemplate.LaunchTemplateName)
		version := aws.StringValue(g.LaunchTemplate.Version)
		if name != "" {
			launchTemplate := name + ":" + version
			return launchTemplate, nil
		}
	}

	// @check: ok, lets check the mixed instance policy
	if g.MixedInstancesPolicy != nil {
		if g.MixedInstancesPolicy.LaunchTemplate != nil {
			if g.MixedInstancesPolicy.LaunchTemplate.LaunchTemplateSpecification != nil {
				var version string
				name = aws.StringValue(g.MixedInstancesPolicy.LaunchTemplate.LaunchTemplateSpecification.LaunchTemplateName)
				//See what version the ASG is set to use
				mixedVersion := aws.StringValue(g.MixedInstancesPolicy.LaunchTemplate.LaunchTemplateSpecification.Version)
				//Correctly Handle Default and Latest Versions
				if mixedVersion == "$Default" || mixedVersion == "$Latest" {
					request := &ec2.DescribeLaunchTemplatesInput{
						LaunchTemplateNames: []*string{&name},
					}
					dltResponse, err := c.EC2().DescribeLaunchTemplates(request)
					if err != nil {
						return "", fmt.Errorf("error describing launch templates: %v", err)
					}
					launchTemplate := dltResponse.LaunchTemplates[0]
					if mixedVersion == "$Default" {
						version = strconv.FormatInt(*launchTemplate.DefaultVersionNumber, 10)
					} else {
						version = strconv.FormatInt(*launchTemplate.LatestVersionNumber, 10)
					}
				} else {
					version = mixedVersion
				}
				klog.V(4).Infof("Launch Template Version Specified By ASG: %v", mixedVersion)
				klog.V(4).Infof("Luanch Template Version we are using for compare: %v", version)
				if name != "" {
					launchTemplate := name + ":" + version
					return launchTemplate, nil
				}
			}
		}
	}

	return "", fmt.Errorf("error finding launch template or configuration for autoscaling group: %s", aws.StringValue(g.AutoScalingGroupName))
}

// findInstanceLaunchConfiguration is responsible for discoverying the launch configuration for an instance
func findInstanceLaunchConfiguration(i *autoscaling.Instance) string {
	name := aws.StringValue(i.LaunchConfigurationName)
	if name != "" {
		return name
	}

	// else we need to check the launch template
	if i.LaunchTemplate != nil {
		name = aws.StringValue(i.LaunchTemplate.LaunchTemplateName)
		version := aws.StringValue(i.LaunchTemplate.Version)
		if name != "" {
			launchTemplate := name + ":" + version
			return launchTemplate
		}
	}

	return ""
}

func awsBuildCloudInstanceGroup(c AWSCloud, ig *kops.InstanceGroup, g *autoscaling.Group, nodeMap map[string]*v1.Node) (*cloudinstances.CloudInstanceGroup, error) {
	newConfigName, err := findAutoscalingGroupLaunchConfiguration(c, g)
	if err != nil {
		return nil, err
	}

	cg := &cloudinstances.CloudInstanceGroup{
		HumanName:     aws.StringValue(g.AutoScalingGroupName),
		InstanceGroup: ig,
		MinSize:       int(aws.Int64Value(g.MinSize)),
		MaxSize:       int(aws.Int64Value(g.MaxSize)),
		Raw:           g,
	}

	for _, i := range g.Instances {
		id := aws.StringValue(i.InstanceId)
		if id == "" {
			klog.Warningf("ignoring instance with no instance id: %s in autoscaling group: %s", id, cg.HumanName)
			continue
		}
		// @step: check if the instance is terminating
		if aws.StringValue(i.LifecycleState) == autoscaling.LifecycleStateTerminating {
			klog.Warningf("ignoring instance  as it is terminating: %s in autoscaling group: %s", id, cg.HumanName)
			continue
		}
		currentConfigName := findInstanceLaunchConfiguration(i)

		if err := cg.NewCloudInstanceGroupMember(id, newConfigName, currentConfigName, nodeMap); err != nil {
			return nil, fmt.Errorf("error creating cloud instance group member: %v", err)
		}
	}

	return cg, nil
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

		klog.Warningf("Uncategorized error in isTagsEventualConsistencyError: %v", awsErr.Code())
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
					klog.Infof("waiting for eventual consistency while describing tags on %q", resourceId)
				}

				klog.V(2).Infof("will retry after encountering error getting tags on %q: %v", resourceId, err)
				time.Sleep(DescribeTagsRetryInterval)
				continue
			}

			return nil, fmt.Errorf("error listing tags on %v: %v", resourceId, err)
		}

		for _, tag := range response.Tags {
			if tag == nil {
				klog.Warning("unexpected nil tag")
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
					klog.Infof("waiting for eventual consistency while creating tags on %q", resourceId)
				}

				klog.V(2).Infof("will retry after encountering error creating tags on %q: %v", resourceId, err)
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
	return deleteTags(c, resourceId, tags)
}

func deleteTags(c AWSCloud, resourceId string, tags map[string]string) error {
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
					klog.Infof("waiting for eventual consistency while deleting tags on %q", resourceId)
				}

				klog.V(2).Infof("will retry after encountering error deleting tags on %q: %v", resourceId, err)
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
		klog.V(4).Infof("adding tags to %q: %v", id, missing)

		err := c.CreateTags(id, missing)
		if err != nil {
			return fmt.Errorf("error adding tags to resource %q: %v", id, err)
		}
	}

	return nil
}

func (c *awsCloudImplementation) GetELBTags(loadBalancerName string) (map[string]string, error) {
	return getELBTags(c, loadBalancerName)
}

func getELBTags(c AWSCloud, loadBalancerName string) (map[string]string, error) {
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
	return createELBTags(c, loadBalancerName, tags)
}

func createELBTags(c AWSCloud, loadBalancerName string, tags map[string]string) error {
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

// RemoveELBTags will remove tags to the specified loadBalancer, retrying up to MaxCreateTagsAttempts times if it hits an eventual-consistency type error
func (c *awsCloudImplementation) RemoveELBTags(loadBalancerName string, tags map[string]string) error {
	return removeELBTags(c, loadBalancerName, tags)
}

func removeELBTags(c AWSCloud, loadBalancerName string, tags map[string]string) error {
	if len(tags) == 0 {
		return nil
	}

	elbTagKeysOnly := []*elb.TagKeyOnly{}
	for k := range tags {
		elbTagKeysOnly = append(elbTagKeysOnly, &elb.TagKeyOnly{Key: aws.String(k)})
	}

	attempt := 0
	for {
		attempt++

		request := &elb.RemoveTagsInput{
			Tags:              elbTagKeysOnly,
			LoadBalancerNames: []*string{&loadBalancerName},
		}

		_, err := c.ELB().RemoveTags(request)
		if err != nil {
			return fmt.Errorf("error creating tags on %v: %v", loadBalancerName, err)
		}

		return nil
	}
}

func (c *awsCloudImplementation) GetELBV2Tags(ResourceArn string) (map[string]string, error) {
	return getELBV2Tags(c, ResourceArn)
}

func getELBV2Tags(c AWSCloud, ResourceArn string) (map[string]string, error) {
	tags := map[string]string{}

	request := &elbv2.DescribeTagsInput{
		ResourceArns: []*string{&ResourceArn},
	}

	attempt := 0
	for {
		attempt++

		response, err := c.ELBV2().DescribeTags(request)
		if err != nil {
			return nil, fmt.Errorf("error listing tags on %v: %v", ResourceArn, err)
		}

		for _, tagset := range response.TagDescriptions {
			for _, tag := range tagset.Tags {
				tags[aws.StringValue(tag.Key)] = aws.StringValue(tag.Value)
			}
		}

		return tags, nil
	}
}

func (c *awsCloudImplementation) CreateELBV2Tags(ResourceArn string, tags map[string]string) error {
	return createELBV2Tags(c, ResourceArn, tags)
}

func createELBV2Tags(c AWSCloud, ResourceArn string, tags map[string]string) error {
	if len(tags) == 0 {
		return nil
	}

	elbv2Tags := []*elbv2.Tag{}
	for k, v := range tags {
		elbv2Tags = append(elbv2Tags, &elbv2.Tag{Key: aws.String(k), Value: aws.String(v)})
	}

	attempt := 0
	for {
		attempt++

		request := &elbv2.AddTagsInput{
			Tags:         elbv2Tags,
			ResourceArns: []*string{&ResourceArn},
		}

		_, err := c.ELBV2().AddTags(request)
		if err != nil {
			return fmt.Errorf("error creating tags on %v: %v", ResourceArn, err)
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
		klog.Warningf("Name not set when filtering by name")
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
		klog.Warningf("Name not set when filtering by name")
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
func (c *awsCloudImplementation) DescribeInstance(instanceID string) (*ec2.Instance, error) {
	klog.V(2).Infof("Calling DescribeInstances for instance %q", instanceID)
	request := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{&instanceID},
	}

	response, err := c.EC2().DescribeInstances(request)
	if err != nil {
		return nil, fmt.Errorf("error listing Instances: %v", err)
	}
	if response == nil || len(response.Reservations) == 0 {
		return nil, nil
	}
	if len(response.Reservations) != 1 {
		klog.Fatalf("found multiple Reservations for %q", instanceID)
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
func (c *awsCloudImplementation) DescribeVPC(vpcID string) (*ec2.Vpc, error) {
	return describeVPC(c, vpcID)
}

func describeVPC(c AWSCloud, vpcID string) (*ec2.Vpc, error) {
	klog.V(2).Infof("Calling DescribeVPC for VPC %q", vpcID)
	request := &ec2.DescribeVpcsInput{
		VpcIds: []*string{&vpcID},
	}

	response, err := c.EC2().DescribeVpcs(request)
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
	klog.V(2).Infof("Calling DescribeImages to resolve name %q", name)
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
			case "amazon.com":
				owner = WellKnownAccountAmazonSystemLinux2
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

	image := response.Images[0]
	for _, v := range response.Images {
		itime, _ := time.Parse(time.RFC3339, *image.CreationDate)
		vtime, _ := time.Parse(time.RFC3339, *v.CreationDate)
		if vtime.After(itime) {
			image = v
		}
	}

	klog.V(4).Infof("Resolved image %q", aws.StringValue(image.ImageId))
	return image, nil
}

func (c *awsCloudImplementation) DescribeAvailabilityZones() ([]*ec2.AvailabilityZone, error) {
	klog.V(2).Infof("Querying EC2 for all valid zones in region %q", c.region)

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

			klog.Infof("Known zones: %q", strings.Join(knownZones, ","))
			return fmt.Errorf("Zone is not a recognized AZ: %q (check you have specified a valid zone?)", zone)
		}

		for _, message := range z.Messages {
			klog.Warningf("Zone %q has message: %q", zone, aws.StringValue(message.Message))
		}

		if aws.StringValue(z.State) != "available" {
			klog.Warningf("Zone %q has state %q", zone, aws.StringValue(z.State))
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

func (c *awsCloudImplementation) IAM() iamiface.IAMAPI {
	return c.iam
}

func (c *awsCloudImplementation) ELB() elbiface.ELBAPI {
	return c.elb
}

func (c *awsCloudImplementation) ELBV2() elbv2iface.ELBV2API {
	return c.elbv2
}

func (c *awsCloudImplementation) Autoscaling() autoscalingiface.AutoScalingAPI {
	return c.autoscaling
}

func (c *awsCloudImplementation) Route53() route53iface.Route53API {
	return c.route53
}

func (c *awsCloudImplementation) Spotinst() spotinst.Service {
	return c.spotinst
}

func (c *awsCloudImplementation) FindVPCInfo(vpcID string) (*fi.VPCInfo, error) {
	return findVPCInfo(c, vpcID)
}

func findVPCInfo(c AWSCloud, vpcID string) (*fi.VPCInfo, error) {
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
		klog.V(2).Infof("Calling DescribeSubnets for subnets in VPC %q", vpcID)
		request := &ec2.DescribeSubnetsInput{
			Filters: []*ec2.Filter{NewEC2Filter("vpc-id", vpcID)},
		}

		response, err := c.EC2().DescribeSubnets(request)
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
	igZones, err := model.FindZonesForInstanceGroup(cluster, ig)
	if err != nil {
		return "", err
	}
	igZonesSet := sets.NewString(igZones...)

	// TODO: Validate that instance type exists in all AZs, but skip AZs that don't support any VPC stuff
	for _, instanceType := range candidates {
		zones, err := c.zonesWithInstanceType(instanceType)
		if err != nil {
			return "", err
		}
		if zones.IsSuperset(igZonesSet) {
			return instanceType, nil
		} else {
			klog.V(2).Infof("can't use instance type %q, available in zones %v but need %v", instanceType, zones, igZones)
		}
	}

	return "", fmt.Errorf("could not find a suitable supported instance type for the instance group %q (type %q) in region %q", ig.Name, ig.Spec.Role, c.region)
}

// supportsInstanceType uses the DescribeReservedInstancesOfferings API call to determine if an instance type is supported in a region
func (c *awsCloudImplementation) zonesWithInstanceType(instanceType string) (sets.String, error) {
	klog.V(4).Infof("checking if instance type %q is supported in region %q", instanceType, c.region)
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
			klog.Warningf("skipping non-matching instance type offering: %v", item)
		}
	}

	return zones, nil
}
