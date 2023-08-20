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

package awsup

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/service/eventbridge"
	"github.com/aws/aws-sdk-go/service/eventbridge/eventbridgeiface"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"golang.org/x/sync/errgroup"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface"
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
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"github.com/aws/aws-sdk-go/service/sts"
	"k8s.io/klog/v2"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	k8s_aws "k8s.io/cloud-provider-aws/pkg/providers/v1"

	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	dnsproviderroute53 "k8s.io/kops/dnsprovider/pkg/dnsprovider/providers/aws/route53"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/model"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/pkg/featureflag"
	identity_aws "k8s.io/kops/pkg/nodeidentity/aws"
	"k8s.io/kops/pkg/resources/spotinst"
	"k8s.io/kops/upup/pkg/fi"
)

// By default, aws-sdk-go only retries 3 times, which doesn't give
// much time for exponential backoff to work for serious issues. At 13
// retries, we'll try a given request for up to ~6m with exponential
// backoff along the way.
const ClientMaxRetries = 13

const (
	DescribeTagsMaxAttempts   = 120
	DescribeTagsRetryInterval = 2 * time.Second
	DescribeTagsLogInterval   = 10 // this is in "retry intervals"
)

const (
	CreateTagsMaxAttempts   = 120
	CreateTagsRetryInterval = 2 * time.Second
	CreateTagsLogInterval   = 10 // this is in "retry intervals"
)

const (
	DeleteTagsMaxAttempts   = 120
	DeleteTagsRetryInterval = 2 * time.Second
	DeleteTagsLogInterval   = 10 // this is in "retry intervals"
)

const (
	TagClusterName           = "KubernetesCluster"
	TagNameRolePrefix        = "k8s.io/role/"
	TagNameEtcdClusterPrefix = "k8s.io/etcd/"
)

const TagRoleControlPlane = "control-plane"
const TagRoleMaster = "master"

// TagNameKopsRole is the AWS tag used to identify the role an object plays for a cluster
const TagNameKopsRole = "kubernetes.io/kops/role"

// TagNameClusterOwnershipPrefix is the AWS tag used for ownership
const TagNameClusterOwnershipPrefix = "kubernetes.io/cluster/"

const tagNameDetachedInstance = "kops.k8s.io/detached-from-asg"

const (
	WellKnownAccountAmazonLinux2 = "137112412989"
	WellKnownAccountDebian       = "136693071363"
	WellKnownAccountFlatcar      = "075585003325"
	WellKnownAccountRedhat       = "309956199498"
	WellKnownAccountUbuntu       = "099720109477"
)

const instanceInServiceState = "InService"

// AWSErrCodeInvalidAction is returned in AWS partitions that don't support certain actions
const AWSErrCodeInvalidAction = "InvalidAction"

type AWSCloud interface {
	fi.Cloud
	Session() (*session.Session, error)
	EC2() ec2iface.EC2API
	IAM() iamiface.IAMAPI
	ELB() elbiface.ELBAPI
	ELBV2() elbv2iface.ELBV2API
	Autoscaling() autoscalingiface.AutoScalingAPI
	Route53() route53iface.Route53API
	Spotinst() spotinst.Cloud
	SQS() sqsiface.SQSAPI
	EventBridge() eventbridgeiface.EventBridgeAPI
	SSM() ssmiface.SSMAPI

	// TODO: Document and rationalize these tags/filters methods
	AddTags(name *string, tags map[string]string)
	BuildFilters(name *string) []*ec2.Filter
	BuildTags(name *string) map[string]string
	Tags() map[string]string

	// GetTags will fetch the tags for the specified resource, retrying (up to MaxDescribeTagsAttempts) if it hits an eventual-consistency type error
	GetTags(resourceId string) (map[string]string, error)
	// CreateTags will add/modify tags to the specified resource, retrying up to MaxCreateTagsAttempts times if it hits an eventual-consistency type error
	CreateTags(resourceId string, tags map[string]string) error
	// DeleteTags will remove tags from the specified resource, retrying up to MaxCreateTagsAttempts times if it hits an eventual-consistency type error
	DeleteTags(resourceId string, tags map[string]string) error
	// UpdateTags will update tags of the specified resource to match tags, using getTags(), createTags() and deleteTags()
	UpdateTags(resourceId string, tags map[string]string) error
	AddAWSTags(id string, expected map[string]string) error
	GetELBTags(loadBalancerName string) (map[string]string, error)
	GetELBV2Tags(ResourceArn string) (map[string]string, error)

	// CreateELBTags will add tags to the specified loadBalancer, retrying up to MaxCreateTagsAttempts times if it hits an eventual-consistency type error
	CreateELBTags(loadBalancerName string, tags map[string]string) error
	CreateELBV2Tags(ResourceArn string, tags map[string]string) error
	// RemoveELBTags will remove tags from the specified loadBalancer, retrying up to MaxCreateTagsAttempts times if it hits an eventual-consistency type error
	RemoveELBTags(loadBalancerName string, tags map[string]string) error
	RemoveELBV2Tags(ResourceArn string, tags map[string]string) error
	FindELBByNameTag(findNameTag string) (*elb.LoadBalancerDescription, error)
	DescribeELBTags(loadBalancerNames []string) (map[string][]*elb.Tag, error)
	FindELBV2ByNameTag(findNameTag string) (*elbv2.LoadBalancer, error)
	DescribeELBV2Tags(loadBalancerNames []string) (map[string][]*elbv2.Tag, error)
	FindELBV2NetworkInterfacesByName(vpcID string, loadBalancerName string) ([]*ec2.NetworkInterface, error)

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

	// DescribeInstanceType calls ec2.DescribeInstanceType to get information for a particular instance type
	DescribeInstanceType(instanceType string) (*ec2.InstanceTypeInfo, error)

	// AccountInfo returns the AWS account ID and AWS partition that we are deploying into
	AccountInfo() (string, string, error)
}

type awsCloudImplementation struct {
	ec2         *ec2.EC2
	iam         *iam.IAM
	elb         *elb.ELB
	elbv2       *elbv2.ELBV2
	autoscaling *autoscaling.AutoScaling
	route53     *route53.Route53
	spotinst    spotinst.Cloud
	sts         *sts.STS
	sqs         *sqs.SQS
	eventbridge *eventbridge.EventBridge
	ssm         *ssm.SSM

	region string

	tags map[string]string

	regionDelayers *RegionDelayers

	instanceTypes *instanceTypes
}

type RegionDelayers struct {
	mutex      sync.Mutex
	delayerMap map[string]*k8s_aws.CrossRequestRetryDelay
}

type instanceTypes struct {
	mutex   sync.Mutex
	typeMap map[string]*ec2.InstanceTypeInfo
}

var _ fi.Cloud = &awsCloudImplementation{}

func (c *awsCloudImplementation) ProviderID() kops.CloudProviderID {
	return kops.CloudProviderAWS
}

func (c *awsCloudImplementation) Region() string {
	return c.region
}

type awsCloudInstancesRegionMap struct {
	mutex     sync.Mutex
	regionMap map[string]AWSCloud
}

func newAwsCloudInstancesRegionMap() *awsCloudInstancesRegionMap {
	return &awsCloudInstancesRegionMap{
		regionMap: make(map[string]AWSCloud),
	}
}

var awsCloudInstances *awsCloudInstancesRegionMap = newAwsCloudInstancesRegionMap()

func ResetAWSCloudInstances() {
	awsCloudInstances.mutex.Lock()
	awsCloudInstances.regionMap = make(map[string]AWSCloud)
	awsCloudInstances.mutex.Unlock()
}

func setConfig(config *aws.Config) *aws.Config {
	// This avoids a confusing error message when we fail to get credentials
	// e.g. https://github.com/kubernetes/kops/issues/605
	config = config.WithCredentialsChainVerboseErrors(true)
	return request.WithRetryer(config, newLoggingRetryer(ClientMaxRetries))
}

func updateAwsCloudInstances(region string, cloud AWSCloud) {
	awsCloudInstances.mutex.Lock()
	awsCloudInstances.regionMap[region] = cloud
	awsCloudInstances.mutex.Unlock()
}

func getCloudInstancesFromRegion(region string) AWSCloud {
	awsCloudInstances.mutex.Lock()
	defer awsCloudInstances.mutex.Unlock()

	cloud, ok := awsCloudInstances.regionMap[region]
	if !ok {
		return nil
	}

	return cloud
}

func NewAWSCloud(region string, tags map[string]string) (AWSCloud, error) {
	raw := getCloudInstancesFromRegion(region)

	if raw == nil {
		c := &awsCloudImplementation{
			region: region,
			regionDelayers: &RegionDelayers{
				delayerMap: make(map[string]*k8s_aws.CrossRequestRetryDelay),
			},
			instanceTypes: &instanceTypes{
				typeMap: make(map[string]*ec2.InstanceTypeInfo),
			},
		}

		config := aws.NewConfig().WithRegion(region)
		config = setConfig(config)

		requestLogger := newRequestLogger(2)

		sess, err := session.NewSessionWithOptions(session.Options{
			Config:            *config,
			SharedConfigState: session.SharedConfigEnable,
		})
		if err != nil {
			return c, err
		}

		// assumes the role before executing commands
		roleARN := os.Getenv("KOPS_AWS_ROLE_ARN")
		if roleARN != "" {
			creds := stscreds.NewCredentials(sess, roleARN)
			config = &aws.Config{Credentials: creds}
			config = setConfig(config).WithRegion(region)
		}

		c.ec2 = ec2.New(sess, config)
		c.ec2.Handlers.Send.PushFront(requestLogger)
		c.addHandlers(region, &c.ec2.Handlers)

		sess, err = session.NewSessionWithOptions(session.Options{
			Config:            *config,
			SharedConfigState: session.SharedConfigEnable,
		})
		if err != nil {
			return c, err
		}
		c.iam = iam.New(sess, config)
		c.iam.Handlers.Send.PushFront(requestLogger)
		c.addHandlers(region, &c.iam.Handlers)

		sess, err = session.NewSessionWithOptions(session.Options{
			Config:            *config,
			SharedConfigState: session.SharedConfigEnable,
		})
		if err != nil {
			return c, err
		}
		c.elb = elb.New(sess, config)
		c.elb.Handlers.Send.PushFront(requestLogger)
		c.addHandlers(region, &c.elb.Handlers)

		sess, err = session.NewSessionWithOptions(session.Options{
			Config:            *config,
			SharedConfigState: session.SharedConfigEnable,
		})
		if err != nil {
			return c, err
		}
		c.elbv2 = elbv2.New(sess, config)
		c.elbv2.Handlers.Send.PushFront(requestLogger)
		c.addHandlers(region, &c.elbv2.Handlers)

		sess, err = session.NewSessionWithOptions(session.Options{
			Config:            *config,
			SharedConfigState: session.SharedConfigEnable,
		})
		if err != nil {
			return c, err
		}
		c.sts = sts.New(sess, config)
		c.sts.Handlers.Send.PushFront(requestLogger)
		c.addHandlers(region, &c.sts.Handlers)

		sess, err = session.NewSessionWithOptions(session.Options{
			Config:            *config,
			SharedConfigState: session.SharedConfigEnable,
		})
		if err != nil {
			return c, err
		}
		c.autoscaling = autoscaling.New(sess, config)
		c.autoscaling.Handlers.Send.PushFront(requestLogger)
		c.addHandlers(region, &c.autoscaling.Handlers)

		sess, err = session.NewSessionWithOptions(session.Options{
			Config:            *config,
			SharedConfigState: session.SharedConfigEnable,
		})
		if err != nil {
			return c, err
		}
		c.route53 = route53.New(sess, config)
		c.route53.Handlers.Send.PushFront(requestLogger)
		c.addHandlers(region, &c.route53.Handlers)

		if featureflag.Spotinst.Enabled() {
			c.spotinst, err = spotinst.NewCloud(kops.CloudProviderAWS)
			if err != nil {
				return c, err
			}
		}

		sess, err = session.NewSessionWithOptions(session.Options{
			Config:            *config,
			SharedConfigState: session.SharedConfigEnable,
		})
		if err != nil {
			return c, err
		}
		c.sqs = sqs.New(sess, config)
		c.sqs.Handlers.Send.PushFront(requestLogger)
		c.addHandlers(region, &c.sqs.Handlers)

		sess, err = session.NewSessionWithOptions(session.Options{
			Config:            *config,
			SharedConfigState: session.SharedConfigEnable,
		})
		if err != nil {
			return c, err
		}
		c.eventbridge = eventbridge.New(sess, config)
		c.eventbridge.Handlers.Send.PushFront(requestLogger)
		c.addHandlers(region, &c.eventbridge.Handlers)

		sess, err = session.NewSessionWithOptions(session.Options{
			Config:            *config,
			SharedConfigState: session.SharedConfigEnable,
		})
		if err != nil {
			return c, err
		}
		c.ssm = ssm.New(sess, config)
		c.ssm.Handlers.Send.PushFront(requestLogger)
		c.addHandlers(region, &c.ssm.Handlers)

		updateAwsCloudInstances(region, c)

		raw = c
	}

	i := raw.WithTags(tags)

	return i, nil
}

func (c *awsCloudImplementation) Session() (*session.Session, error) {
	config := aws.NewConfig().WithRegion(c.region)
	config = config.WithCredentialsChainVerboseErrors(true)
	config = request.WithRetryer(config, newLoggingRetryer(ClientMaxRetries))

	sess, err := session.NewSessionWithOptions(session.Options{
		Config:            *config,
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return sess, err
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
		if featureflag.SpotinstHybrid.Enabled() {
			if _, ok := g.Raw.(*autoscaling.Group); ok {
				return deleteGroup(c, g)
			}
		}

		return spotinst.DeleteInstanceGroup(c.spotinst, g)
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

	// Delete detached instances
	{
		detached, err := findDetachedInstances(c, asg)
		if err != nil {
			return fmt.Errorf("error searching for detached instances for autoscaling group %q: %v", name, err)
		}
		if len(detached) > 0 {
			klog.V(2).Infof("Deleting detached instances for autoscaling group %q", name)
			req := &ec2.TerminateInstancesInput{
				InstanceIds: detached,
			}
			if _, err := c.EC2().TerminateInstances(req); err != nil {
				return fmt.Errorf("error deleting detached instances for autoscaling group %q: %v", name, err)
			}
		}
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
func (c *awsCloudImplementation) DeleteInstance(i *cloudinstances.CloudInstance) error {
	if c.spotinst != nil {
		if featureflag.SpotinstHybrid.Enabled() {
			if _, ok := i.CloudInstanceGroup.Raw.(*autoscaling.Group); ok {
				return deleteInstance(c, i)
			}
		}

		return spotinst.DeleteInstance(c.spotinst, i)
	}

	return deleteInstance(c, i)
}

// DeregisterInstance drains a cloud instance and load balancers.
func (c *awsCloudImplementation) DeregisterInstance(i *cloudinstances.CloudInstance) error {
	if c.spotinst != nil || i.CloudInstanceGroup.InstanceGroup.Spec.Manager == kops.InstanceManagerKarpenter {
		return nil
	}

	err := deregisterInstance(c, i)
	if err != nil {
		return fmt.Errorf("failed to deregister instance from loadBalancer before terminating: %v", err)
	}

	return nil
}

func deleteInstance(c AWSCloud, i *cloudinstances.CloudInstance) error {
	id := i.ID
	if id == "" {
		return fmt.Errorf("id was not set on CloudInstance: %v", i)
	}

	request := &ec2.TerminateInstancesInput{
		InstanceIds: []*string{aws.String(id)},
	}

	if _, err := c.EC2().TerminateInstances(request); err != nil {
		if AWSErrorCode(err) == "InvalidInstanceID.NotFound" {
			klog.V(2).Infof("Got InvalidInstanceID.NotFound error deleting instance %q; will treat as already-deleted", id)
		} else {
			return fmt.Errorf("error deleting instance %q: %v", id, err)
		}
	}

	klog.V(8).Infof("deleted aws ec2 instance %q", id)

	return nil
}

// deregisterInstance ensures that the instance is fully drained/removed from all associated loadBalancers and targetGroups before termination.
func deregisterInstance(c AWSCloud, i *cloudinstances.CloudInstance) error {
	asg := i.CloudInstanceGroup.Raw.(*autoscaling.Group)

	asgDetails, err := c.Autoscaling().DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{asg.AutoScalingGroupName},
	})
	if err != nil {
		return fmt.Errorf("error describing autoScalingGroups: %v", err)
	}

	if len(asgDetails.AutoScalingGroups) == 0 {
		return nil
	}

	// there will always be only one ASG in the DescribeAutoScalingGroups response.
	loadBalancerNames := aws.StringValueSlice(asgDetails.AutoScalingGroups[0].LoadBalancerNames)
	targetGroupArns := aws.StringValueSlice(asgDetails.AutoScalingGroups[0].TargetGroupARNs)

	eg, _ := errgroup.WithContext(context.Background())

	if len(loadBalancerNames) != 0 {
		eg.Go(func() error {
			return deregisterInstanceFromClassicLoadBalancer(c, loadBalancerNames, i.ID)
		})
	}

	if len(targetGroupArns) != 0 {
		eg.Go(func() error {
			return deregisterInstanceFromTargetGroups(c, targetGroupArns, i.ID)
		})
	}

	if err := eg.Wait(); err != nil {
		return fmt.Errorf("failed to deregister instance from load balancers: %v", err)
	}

	return nil
}

// deregisterInstanceFromClassicLoadBalancer ensures that connectionDraining completes for the associated classic loadBalancer to ensure no dropped connections.
func deregisterInstanceFromClassicLoadBalancer(c AWSCloud, loadBalancerNames []string, instanceId string) error {
	klog.Infof("Deregistering instance from classic loadBalancers: %v", loadBalancerNames)

	for {
		instanceDraining := false
		for _, loadBalancerName := range loadBalancerNames {
			response, err := c.ELB().DescribeInstanceHealth(&elb.DescribeInstanceHealthInput{
				LoadBalancerName: aws.String(loadBalancerName),
				Instances: []*elb.Instance{{
					InstanceId: aws.String(instanceId),
				}},
			})
			if err != nil {
				return fmt.Errorf("error describing instance health: %v", err)
			}

			// describeInstanceHealth can return an empty list if the instance was already terminated.
			if len(response.InstanceStates) == 0 {
				continue
			}

			// there will be only one instance in the DescribeInstanceHealth response.
			if aws.StringValue(response.InstanceStates[0].State) == instanceInServiceState {
				c.ELB().DeregisterInstancesFromLoadBalancer(&elb.DeregisterInstancesFromLoadBalancerInput{
					LoadBalancerName: aws.String(loadBalancerName),
					Instances: []*elb.Instance{{
						InstanceId: aws.String(instanceId),
					}},
				})
				instanceDraining = true
			}
		}

		if !instanceDraining {
			break
		}

		time.Sleep(5 * time.Second)
	}
	return nil
}

// deregisterInstanceFromTargetGroups ensures that instances are fully unused in the corresponding targetGroups before instance termination.
// this ensures that connections are fully drained from the instance before terminating.
func deregisterInstanceFromTargetGroups(c AWSCloud, targetGroupArns []string, instanceId string) error {
	eg, _ := errgroup.WithContext(context.Background())

	for _, targetGroupArn := range targetGroupArns {
		arn := targetGroupArn
		eg.Go(func() error {
			return deregisterInstanceFromTargetGroup(c, arn, instanceId)
		})
	}

	if err := eg.Wait(); err != nil {
		return fmt.Errorf("failed to register instance from targetGroups: %w", err)
	}

	return nil
}

func deregisterInstanceFromTargetGroup(c AWSCloud, targetGroupArn string, instanceId string) error {
	klog.Infof("Deregistering instance from targetGroup: %s", targetGroupArn)

	for {
		instanceDraining := false

		response, err := c.ELBV2().DescribeTargetHealth(&elbv2.DescribeTargetHealthInput{
			TargetGroupArn: aws.String(targetGroupArn),
			Targets: []*elbv2.TargetDescription{{
				Id: aws.String(instanceId),
			}},
		})
		if err != nil {
			return fmt.Errorf("error describing target health: %w", err)
		}

		// there will be only one target in the DescribeTargetHealth response.
		// DescribeTargetHealth response will contain a target even if the targetId doesn't exist.
		// all other states besides TargetHealthStateUnused means that the instance may still be serving traffic.
		if aws.StringValue(response.TargetHealthDescriptions[0].TargetHealth.State) != elbv2.TargetHealthStateEnumUnused {
			_, err = c.ELBV2().DeregisterTargets(&elbv2.DeregisterTargetsInput{
				TargetGroupArn: aws.String(targetGroupArn),
				Targets: []*elbv2.TargetDescription{{
					Id: aws.String(instanceId),
				}},
			})

			if err != nil {
				return fmt.Errorf("error deregistering target: %w", err)
			}

			instanceDraining = true
		}

		if !instanceDraining {
			break
		}

		time.Sleep(5 * time.Second)
	}

	klog.Infof("Successfully drained instance from targetGroup: %s", targetGroupArn)

	return nil
}

// DetachInstance causes an aws instance to no longer be counted against the ASG's size limits.
func (c *awsCloudImplementation) DetachInstance(i *cloudinstances.CloudInstance) error {
	if i.Status == cloudinstances.CloudInstanceStatusDetached {
		return nil
	}
	if c.spotinst != nil {
		return spotinst.DetachInstance(c.spotinst, i)
	}

	return detachInstance(c, i)
}

func detachInstance(c AWSCloud, i *cloudinstances.CloudInstance) error {
	id := i.ID
	if id == "" {
		return fmt.Errorf("id was not set on CloudInstance: %v", i)
	}

	asg := i.CloudInstanceGroup.Raw.(*autoscaling.Group)
	if err := c.CreateTags(id, map[string]string{tagNameDetachedInstance: *asg.AutoScalingGroupName}); err != nil {
		return fmt.Errorf("error tagging instance %q: %v", id, err)
	}

	// TODO this also deregisters the instance from any ELB attached to the ASG. Do we care?

	input := &autoscaling.DetachInstancesInput{
		AutoScalingGroupName:           aws.String(i.CloudInstanceGroup.HumanName),
		InstanceIds:                    []*string{aws.String(id)},
		ShouldDecrementDesiredCapacity: aws.Bool(false),
	}

	if _, err := c.Autoscaling().DetachInstances(input); err != nil {
		return fmt.Errorf("error detaching instance %q: %v", id, err)
	}

	klog.V(8).Infof("detached aws ec2 instance %q", id)

	return nil
}

// GetCloudGroups returns a groups of instances that back a kops instance groups
func (c *awsCloudImplementation) GetCloudGroups(cluster *kops.Cluster, instancegroups []*kops.InstanceGroup, warnUnmatched bool, nodes []v1.Node) (map[string]*cloudinstances.CloudInstanceGroup, error) {
	if c.spotinst != nil {
		sgroups, err := spotinst.GetCloudGroups(c.spotinst, cluster, instancegroups, warnUnmatched, nodes)
		if err != nil {
			return nil, err
		}

		if featureflag.SpotinstHybrid.Enabled() {
			agroups, err := getCloudGroups(c, cluster, instancegroups, warnUnmatched, nodes)
			if err != nil {
				return nil, err
			}

			for name, group := range agroups {
				sgroups[name] = group
			}
		}

		return sgroups, nil
	}

	cloudGroups, err := getCloudGroups(c, cluster, instancegroups, warnUnmatched, nodes)
	if err != nil {
		return nil, err
	}
	karpenterGroups, err := getKarpenterGroups(c, cluster, instancegroups, nodes)
	if err != nil {
		return nil, err
	}

	for name, group := range karpenterGroups {
		cloudGroups[name] = group
	}
	return cloudGroups, nil
}

func getKarpenterGroups(c AWSCloud, cluster *kops.Cluster, instancegroups []*kops.InstanceGroup, nodes []v1.Node) (map[string]*cloudinstances.CloudInstanceGroup, error) {
	cloudGroups := make(map[string]*cloudinstances.CloudInstanceGroup)
	for _, ig := range instancegroups {
		if ig.Spec.Manager == kops.InstanceManagerKarpenter {
			group, err := buildKarpenterGroup(c, cluster, ig, nodes)
			if err != nil {
				return nil, err
			}
			cloudGroups[ig.ObjectMeta.Name] = group
		}
	}
	return cloudGroups, nil
}

func buildKarpenterGroup(c AWSCloud, cluster *kops.Cluster, ig *kops.InstanceGroup, nodes []v1.Node) (*cloudinstances.CloudInstanceGroup, error) {
	nodeMap := cloudinstances.GetNodeMap(nodes, cluster)
	instances := make(map[string]*ec2.Instance)
	updatedInstances := make(map[string]*ec2.Instance)
	clusterName := c.Tags()[TagClusterName]
	var version string

	{
		input := &ec2.DescribeLaunchTemplatesInput{
			Filters: []*ec2.Filter{
				NewEC2Filter("tag:"+identity_aws.CloudTagInstanceGroupName, ig.ObjectMeta.Name),
				NewEC2Filter("tag:"+TagClusterName, clusterName),
			},
		}
		var list []*ec2.LaunchTemplate
		err := c.EC2().DescribeLaunchTemplatesPages(input, func(p *ec2.DescribeLaunchTemplatesOutput, lastPage bool) (shouldContinue bool) {
			list = append(list, p.LaunchTemplates...)
			return true
		})
		if err != nil {
			return nil, err
		}
		lt := list[0]
		versionNumber := *lt.LatestVersionNumber
		version = strconv.Itoa(int(versionNumber))

	}

	karpenterGroup := &cloudinstances.CloudInstanceGroup{
		InstanceGroup: ig,
		HumanName:     ig.ObjectMeta.Name,
	}
	{
		req := &ec2.DescribeInstancesInput{
			Filters: []*ec2.Filter{
				NewEC2Filter("tag:"+identity_aws.CloudTagInstanceGroupName, ig.ObjectMeta.Name),
				NewEC2Filter("tag:"+TagClusterName, clusterName),
				NewEC2Filter("instance-state-name", "pending", "running", "stopping", "stopped"),
			},
		}

		result, err := c.EC2().DescribeInstances(req)
		if err != nil {
			return nil, err
		}

		for _, r := range result.Reservations {
			for _, i := range r.Instances {
				id := aws.StringValue(i.InstanceId)
				instances[id] = i
			}
		}
	}

	klog.V(2).Infof("found %d karpenter instances", len(instances))

	{
		req := &ec2.DescribeInstancesInput{
			Filters: []*ec2.Filter{
				NewEC2Filter("tag:"+identity_aws.CloudTagInstanceGroupName, ig.ObjectMeta.Name),
				NewEC2Filter("tag:"+TagClusterName, clusterName),
				NewEC2Filter("instance-state-name", "pending", "running", "stopping", "stopped"),
				NewEC2Filter("tag:aws:ec2launchtemplate:version", version),
			},
		}

		result, err := c.EC2().DescribeInstances(req)
		if err != nil {
			return nil, err
		}

		for _, r := range result.Reservations {
			for _, i := range r.Instances {
				id := aws.StringValue(i.InstanceId)
				updatedInstances[id] = i
			}
		}
	}
	klog.V(2).Infof("found %d updated instances", len(updatedInstances))

	{
		for _, instance := range instances {
			id := *instance.InstanceId
			_, ready := updatedInstances[id]
			var status string
			if ready {
				status = cloudinstances.CloudInstanceStatusUpToDate
			} else {
				status = cloudinstances.CloudInstanceStatusNeedsUpdate
			}
			cloudInstance, _ := karpenterGroup.NewCloudInstance(id, status, nodeMap[id])
			addCloudInstanceData(cloudInstance, instance)
		}
	}
	return karpenterGroup, nil
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

		groups[instancegroup.ObjectMeta.Name], err = awsBuildCloudInstanceGroup(c, cluster, instancegroup, asg, nodeMap)
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
	var launchTemplate *autoscaling.LaunchTemplateSpecification
	if g.LaunchTemplate != nil {
		launchTemplate = g.LaunchTemplate
	} else if g.MixedInstancesPolicy != nil && g.MixedInstancesPolicy.LaunchTemplate != nil && g.MixedInstancesPolicy.LaunchTemplate.LaunchTemplateSpecification != nil {
		launchTemplate = g.MixedInstancesPolicy.LaunchTemplate.LaunchTemplateSpecification
	} else {
		return "", fmt.Errorf("error finding launch template or configuration for autoscaling group: %s", aws.StringValue(g.AutoScalingGroupName))
	}

	id := aws.StringValue(launchTemplate.LaunchTemplateId)
	if id == "" {
		return "", fmt.Errorf("error finding launch template ID for autoscaling group: %s", aws.StringValue(g.AutoScalingGroupName))
	}

	version := aws.StringValue(launchTemplate.Version)
	// Correctly Handle Default and Latest Versions
	klog.V(4).Infof("Launch Template Version Specified By ASG: %v", version)
	if version == "" || version == "$Default" || version == "$Latest" {
		input := &ec2.DescribeLaunchTemplatesInput{
			LaunchTemplateIds: []*string{&id},
		}
		output, err := c.EC2().DescribeLaunchTemplates(input)
		if err != nil {
			return "", fmt.Errorf("error describing launch templates: %q", err)
		}
		if len(output.LaunchTemplates) == 0 {
			return "", fmt.Errorf("error finding launch template by ID: %q", id)
		}
		launchTemplate := output.LaunchTemplates[0]
		if version == "$Latest" {
			version = strconv.FormatInt(*launchTemplate.LatestVersionNumber, 10)
		} else {
			version = strconv.FormatInt(*launchTemplate.DefaultVersionNumber, 10)
		}
	}
	klog.V(4).Infof("Launch Template Version used for compare: %q", version)

	return fmt.Sprintf("%s:%s", id, version), nil
}

// findInstanceLaunchConfiguration is responsible for discoverying the launch configuration for an instance
func findInstanceLaunchConfiguration(i *autoscaling.Instance) string {
	name := aws.StringValue(i.LaunchConfigurationName)
	if name != "" {
		return name
	}

	// else we need to check the launch template
	if i.LaunchTemplate != nil {
		id := aws.StringValue(i.LaunchTemplate.LaunchTemplateId)
		version := aws.StringValue(i.LaunchTemplate.Version)
		if id != "" {
			launchTemplate := id + ":" + version
			return launchTemplate
		}
	}

	return ""
}

func awsBuildCloudInstanceGroup(c AWSCloud, cluster *kops.Cluster, ig *kops.InstanceGroup, g *autoscaling.Group, nodeMap map[string]*v1.Node) (*cloudinstances.CloudInstanceGroup, error) {
	newConfigName, err := findAutoscalingGroupLaunchConfiguration(c, g)
	if err != nil {
		return nil, err
	}

	instanceSeen := map[string]bool{}
	instances, err := findInstances(c, ig)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch instances: %v", err)
	}

	cg := &cloudinstances.CloudInstanceGroup{
		HumanName:     aws.StringValue(g.AutoScalingGroupName),
		InstanceGroup: ig,
		MinSize:       int(aws.Int64Value(g.MinSize)),
		TargetSize:    int(aws.Int64Value(g.DesiredCapacity)),
		MaxSize:       int(aws.Int64Value(g.MaxSize)),
		Raw:           g,
	}

	for _, i := range g.Instances {
		err := buildCloudInstance(i, instances, instanceSeen, nodeMap, cg, newConfigName)
		if err != nil {
			return nil, err
		}
	}

	result, err := c.Autoscaling().DescribeWarmPool(&autoscaling.DescribeWarmPoolInput{
		AutoScalingGroupName: g.AutoScalingGroupName,
	})
	if err != nil {
		return nil, err
	}
	for _, i := range result.Instances {
		err := buildCloudInstance(i, instances, instanceSeen, nodeMap, cg, newConfigName)
		if err != nil {
			return nil, err
		}
	}
	var detached []*string
	for id, instance := range instances {
		for _, tag := range instance.Tags {
			if aws.StringValue(tag.Key) == tagNameDetachedInstance {
				detached = append(detached, aws.String(id))
			}
		}
	}
	if err != nil {
		return nil, fmt.Errorf("error searching for detached instances: %v", err)
	}
	for _, id := range detached {
		if id != nil && *id != "" && !instanceSeen[*id] {
			cm, err := cg.NewCloudInstance(*id, cloudinstances.CloudInstanceStatusDetached, nodeMap[*id])
			if err != nil {
				return nil, fmt.Errorf("error creating cloud instance group member: %v", err)
			}
			instanceSeen[*id] = true
			addCloudInstanceData(cm, instances[aws.StringValue(id)])
		}
	}

	return cg, nil
}

func buildCloudInstance(i *autoscaling.Instance, instances map[string]*ec2.Instance, instanceSeen map[string]bool, nodeMap map[string]*v1.Node, cg *cloudinstances.CloudInstanceGroup, newConfigName string) error {
	id := aws.StringValue(i.InstanceId)
	if id == "" {
		klog.Warningf("ignoring instance with no instance id: %s in autoscaling group: %s", id, cg.HumanName)
		return nil
	}
	instanceSeen[id] = true
	// @step: check if the instance is terminating
	if aws.StringValue(i.LifecycleState) == autoscaling.LifecycleStateTerminating {
		klog.Warningf("ignoring instance as it is terminating: %s in autoscaling group: %s", id, cg.HumanName)
		return nil
	}
	if instances[id] == nil {
		return nil
	}
	currentConfigName := findInstanceLaunchConfiguration(i)
	status := cloudinstances.CloudInstanceStatusUpToDate
	if newConfigName != currentConfigName {
		status = cloudinstances.CloudInstanceStatusNeedsUpdate
	}
	cm, err := cg.NewCloudInstance(id, status, nodeMap[id])
	if err != nil {
		return fmt.Errorf("error creating cloud instance group member: %v", err)
	}
	if strings.HasPrefix(*i.LifecycleState, "Warmed") {
		cm.State = cloudinstances.WarmPool
	}

	addCloudInstanceData(cm, instances[id])
	return nil
}

func addCloudInstanceData(cm *cloudinstances.CloudInstance, instance *ec2.Instance) {
	cm.MachineType = aws.StringValue(instance.InstanceType)
	isControlPlane := false
	for _, tag := range instance.Tags {
		key := aws.StringValue(tag.Key)
		if !strings.HasPrefix(key, TagNameRolePrefix) {
			continue
		}
		role := strings.TrimPrefix(key, TagNameRolePrefix)
		if role == "master" || role == "control-plane" {
			isControlPlane = true
		} else {
			cm.Roles = append(cm.Roles, role)
			cm.PrivateIP = aws.StringValue(instance.PrivateIpAddress)
		}
	}
	if isControlPlane {
		cm.Roles = append(cm.Roles, "control-plane")
		cm.PrivateIP = aws.StringValue(instance.PrivateIpAddress)
	}
}

func findInstances(c AWSCloud, ig *kops.InstanceGroup) (map[string]*ec2.Instance, error) {
	clusterName := c.Tags()[TagClusterName]
	req := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			NewEC2Filter("tag:"+identity_aws.CloudTagInstanceGroupName, ig.ObjectMeta.Name),
			NewEC2Filter("tag:"+TagClusterName, clusterName),
			NewEC2Filter("instance-state-name", "pending", "running", "stopping", "stopped"),
		},
	}

	result, err := c.EC2().DescribeInstances(req)
	if err != nil {
		return nil, err
	}

	instances := make(map[string]*ec2.Instance)
	for _, r := range result.Reservations {
		for _, i := range r.Instances {
			id := aws.StringValue(i.InstanceId)
			instances[id] = i
		}
	}
	return instances, nil
}

func findDetachedInstances(c AWSCloud, g *autoscaling.Group) ([]*string, error) {
	clusterName := c.Tags()[TagClusterName]
	req := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			NewEC2Filter("tag:"+tagNameDetachedInstance, aws.StringValue(g.AutoScalingGroupName)),
			NewEC2Filter("tag:"+TagClusterName, clusterName),
			NewEC2Filter("instance-state-name", "pending", "running", "stopping", "stopped"),
		},
	}
	result, err := c.EC2().DescribeInstances(req)
	if err != nil {
		return nil, err
	}
	var detached []*string
	for _, r := range result.Reservations {
		for _, i := range r.Instances {
			detached = append(detached, i.InstanceId)
		}
	}
	return detached, nil
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

// isTagsEventualConsistencyError checks if the error is one of the errors encountered
// when we try to create/get tags before the resource has fully 'propagated' in EC2
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

// GetTags will fetch the tags for the specified resource,
// retrying (up to MaxDescribeTagsAttempts) if it hits an eventual-consistency type error
func (c *awsCloudImplementation) GetTags(resourceID string) (map[string]string, error) {
	return getTags(c, resourceID)
}

func getTags(c AWSCloud, resourceID string) (map[string]string, error) {
	if resourceID == "" {
		return nil, fmt.Errorf("resourceID not provided to getTags")
	}

	tags := map[string]string{}

	request := &ec2.DescribeTagsInput{
		Filters: []*ec2.Filter{
			NewEC2Filter("resource-id", resourceID),
		},
	}

	attempt := 0
	for {
		attempt++

		response, err := c.EC2().DescribeTags(request)
		if err != nil {
			if isTagsEventualConsistencyError(err) {
				if attempt > DescribeTagsMaxAttempts {
					return nil, fmt.Errorf("got retryable error while getting tags on %q, but retried too many times without success: %v", resourceID, err)
				}

				if (attempt % DescribeTagsLogInterval) == 0 {
					klog.Infof("waiting for eventual consistency while describing tags on %q", resourceID)
				}

				klog.V(2).Infof("will retry after encountering error getting tags on %q: %v", resourceID, err)
				time.Sleep(DescribeTagsRetryInterval)
				continue
			}

			return nil, fmt.Errorf("error listing tags on %v: %v", resourceID, err)
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
func (c *awsCloudImplementation) CreateTags(resourceID string, tags map[string]string) error {
	return createTags(c, resourceID, tags)
}

func createTags(c AWSCloud, resourceID string, tags map[string]string) error {
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
			Resources: []*string{&resourceID},
		}

		_, err := c.EC2().CreateTags(request)
		if err != nil {
			if isTagsEventualConsistencyError(err) {
				if attempt > CreateTagsMaxAttempts {
					return fmt.Errorf("got retryable error while creating tags on %q, but retried too many times without success: %v", resourceID, err)
				}

				if (attempt % CreateTagsLogInterval) == 0 {
					klog.Infof("waiting for eventual consistency while creating tags on %q", resourceID)
				}

				klog.V(2).Infof("will retry after encountering error creating tags on %q: %v", resourceID, err)
				time.Sleep(CreateTagsRetryInterval)
				continue
			}

			return fmt.Errorf("error creating tags on %v: %v", resourceID, err)
		}

		return nil
	}
}

// DeleteTags will remove tags from the specified resource,
// retrying up to MaxCreateTagsAttempts times if it hits an eventual-consistency type error
func (c *awsCloudImplementation) DeleteTags(resourceID string, tags map[string]string) error {
	return deleteTags(c, resourceID, tags)
}

func deleteTags(c AWSCloud, resourceID string, tags map[string]string) error {
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
			Resources: []*string{&resourceID},
		}

		_, err := c.EC2().DeleteTags(request)
		if err != nil {
			if isTagsEventualConsistencyError(err) {
				if attempt > DeleteTagsMaxAttempts {
					return fmt.Errorf("got retryable error while deleting tags on %q, but retried too many times without success: %v", resourceID, err)
				}

				if (attempt % DeleteTagsLogInterval) == 0 {
					klog.Infof("waiting for eventual consistency while deleting tags on %q", resourceID)
				}

				klog.V(2).Infof("will retry after encountering error deleting tags on %q: %v", resourceID, err)
				time.Sleep(DeleteTagsRetryInterval)
				continue
			}

			return fmt.Errorf("error deleting tags on %v: %v", resourceID, err)
		}

		return nil
	}
}

// UpdateTags will update tags of the specified resource to match tags,
// using getTags(), createTags() and deleteTags()
func (c *awsCloudImplementation) UpdateTags(resourceID string, tags map[string]string) error {
	return updateTags(c, resourceID, tags)
}

func updateTags(c AWSCloud, resourceID string, expectedTags map[string]string) error {
	actual, err := getTags(c, resourceID)
	if err != nil {
		return err
	}

	missing := make(map[string]string)
	for k, v := range expectedTags {
		if actual[k] != v {
			missing[k] = v
		}
	}
	if len(missing) > 0 {
		klog.V(4).Infof("Adding tags to %q: %v", resourceID, missing)
		err = createTags(c, resourceID, missing)
		if err != nil {
			return err
		}
	}

	extra := make(map[string]string)
	for k, v := range actual {
		if _, ok := expectedTags[k]; !ok {
			extra[k] = v
		}
	}
	if len(extra) > 0 {
		klog.V(4).Infof("Removing tags from %q: %v", resourceID, missing)
		err := deleteTags(c, resourceID, extra)
		if err != nil {
			return err
		}
	}

	return nil
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

// CreateELBTags will add tags to the specified loadBalancer,
// retrying up to MaxCreateTagsAttempts times if it hits an eventual-consistency type error
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

func (c *awsCloudImplementation) RemoveELBV2Tags(ResourceArn string, tags map[string]string) error {
	return removeELBV2Tags(c, ResourceArn, tags)
}

func removeELBV2Tags(c AWSCloud, ResourceArn string, tags map[string]string) error {
	if len(tags) == 0 {
		return nil
	}

	elbTagKeysOnly := []*string{}
	for k := range tags {
		elbTagKeysOnly = append(elbTagKeysOnly, aws.String(k))
	}

	request := &elbv2.RemoveTagsInput{
		TagKeys:      elbTagKeysOnly,
		ResourceArns: []*string{&ResourceArn},
	}

	_, err := c.ELBV2().RemoveTags(request)
	if err != nil {
		return fmt.Errorf("error creating tags on %v: %v", ResourceArn, err)
	}

	return nil
}

func (c *awsCloudImplementation) GetELBV2Tags(ResourceArn string) (map[string]string, error) {
	return getELBV2Tags(c, ResourceArn)
}

func getELBV2Tags(c AWSCloud, ResourceArn string) (map[string]string, error) {
	tags := map[string]string{}

	request := &elbv2.DescribeTagsInput{
		ResourceArns: []*string{&ResourceArn},
	}
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

func (c *awsCloudImplementation) FindELBByNameTag(findNameTag string) (*elb.LoadBalancerDescription, error) {
	return findELBByNameTag(c, findNameTag)
}

func findELBByNameTag(c AWSCloud, findNameTag string) (*elb.LoadBalancerDescription, error) {
	// TODO: Any way around this?
	klog.V(2).Infof("Listing all ELBs for findLoadBalancerByNameTag")

	request := &elb.DescribeLoadBalancersInput{}
	// ELB DescribeTags has a limit of 20 names, so we set the page size here to 20 also
	request.PageSize = aws.Int64(20)

	var found []*elb.LoadBalancerDescription

	var innerError error
	err := c.ELB().DescribeLoadBalancersPages(request, func(p *elb.DescribeLoadBalancersOutput, lastPage bool) bool {
		if len(p.LoadBalancerDescriptions) == 0 {
			return true
		}

		// TODO: Filter by cluster?

		var names []string
		nameToELB := make(map[string]*elb.LoadBalancerDescription)
		for _, elb := range p.LoadBalancerDescriptions {
			name := aws.StringValue(elb.LoadBalancerName)
			nameToELB[name] = elb
			names = append(names, name)
		}

		tagMap, err := c.DescribeELBTags(names)
		if err != nil {
			innerError = err
			return false
		}

		for loadBalancerName, tags := range tagMap {
			name, foundNameTag := FindELBTag(tags, "Name")
			if !foundNameTag || name != findNameTag {
				continue
			}

			elb := nameToELB[loadBalancerName]
			found = append(found, elb)
		}
		return true
	})
	if err != nil {
		return nil, fmt.Errorf("error describing LoadBalancers: %v", err)
	}
	if innerError != nil {
		return nil, fmt.Errorf("error describing LoadBalancers: %v", innerError)
	}

	if len(found) == 0 {
		return nil, nil
	}

	if len(found) != 1 {
		return nil, fmt.Errorf("Found multiple ELBs with Name %q", findNameTag)
	}

	return found[0], nil
}

func (c *awsCloudImplementation) DescribeELBTags(loadBalancerNames []string) (map[string][]*elb.Tag, error) {
	return describeELBTags(c, loadBalancerNames)
}

func describeELBTags(c AWSCloud, loadBalancerNames []string) (map[string][]*elb.Tag, error) {
	// TODO: Filter by cluster?

	request := &elb.DescribeTagsInput{}
	request.LoadBalancerNames = aws.StringSlice(loadBalancerNames)

	// TODO: Cache?
	klog.V(2).Infof("Querying ELB tags for %s", loadBalancerNames)
	response, err := c.ELB().DescribeTags(request)
	if err != nil {
		return nil, err
	}

	tagMap := make(map[string][]*elb.Tag)
	for _, tagset := range response.TagDescriptions {
		tagMap[aws.StringValue(tagset.LoadBalancerName)] = tagset.Tags
	}
	return tagMap, nil
}

func (c *awsCloudImplementation) FindELBV2ByNameTag(findNameTag string) (*elbv2.LoadBalancer, error) {
	return findELBV2ByNameTag(c, findNameTag)
}

func findELBV2ByNameTag(c AWSCloud, findNameTag string) (*elbv2.LoadBalancer, error) {
	// TODO: Any way around this?
	klog.V(2).Infof("Listing all NLBs for findNetworkLoadBalancerByNameTag")

	request := &elbv2.DescribeLoadBalancersInput{}
	// ELB DescribeTags has a limit of 20 names, so we set the page size here to 20 also
	request.PageSize = aws.Int64(20)

	var found []*elbv2.LoadBalancer

	var innerError error
	err := c.ELBV2().DescribeLoadBalancersPages(request, func(p *elbv2.DescribeLoadBalancersOutput, lastPage bool) bool {
		if len(p.LoadBalancers) == 0 {
			return true
		}

		// TODO: Filter by cluster?

		var arns []string
		arnToELB := make(map[string]*elbv2.LoadBalancer)
		for _, elb := range p.LoadBalancers {
			arn := aws.StringValue(elb.LoadBalancerArn)
			arnToELB[arn] = elb
			arns = append(arns, arn)
		}

		tagMap, err := c.DescribeELBV2Tags(arns)
		if err != nil {
			innerError = err
			return false
		}

		for loadBalancerArn, tags := range tagMap {
			name, foundNameTag := FindELBV2Tag(tags, "Name")
			if !foundNameTag || name != findNameTag {
				continue
			}
			elb := arnToELB[loadBalancerArn]
			found = append(found, elb)
		}
		return true
	})
	if err != nil {
		return nil, fmt.Errorf("error describing LoadBalancers: %v", err)
	}
	if innerError != nil {
		return nil, fmt.Errorf("error describing LoadBalancers: %v", innerError)
	}

	if len(found) == 0 {
		return nil, nil
	}

	if len(found) != 1 {
		return nil, fmt.Errorf("Found multiple NLBs with Name %q", findNameTag)
	}

	return found[0], nil
}

func (c *awsCloudImplementation) FindELBV2NetworkInterfacesByName(vpcID string, loadBalancerName string) ([]*ec2.NetworkInterface, error) {
	return findELBV2NetworkInterfaces(c, vpcID, loadBalancerName)
}

func findELBV2NetworkInterfaces(c AWSCloud, vpcID, lbName string) ([]*ec2.NetworkInterface, error) {
	klog.V(2).Infof("Listing all NLB network interfaces")

	request := &ec2.DescribeNetworkInterfacesInput{
		Filters: []*ec2.Filter{
			NewEC2Filter("vpc-id", vpcID),
			NewEC2Filter("interface-type", "network_load_balancer"),
		},
	}

	response, err := c.EC2().DescribeNetworkInterfaces(request)
	if err != nil {
		return nil, fmt.Errorf("error describing network interfaces: %w", err)
	}

	var found []*ec2.NetworkInterface
	for _, ni := range response.NetworkInterfaces {
		if strings.HasPrefix(aws.StringValue(ni.Description), "ELB net/"+lbName+"/") {
			found = append(found, ni)
		}
	}

	return found, nil
}

func (c *awsCloudImplementation) DescribeELBV2Tags(loadBalancerArns []string) (map[string][]*elbv2.Tag, error) {
	return describeELBV2Tags(c, loadBalancerArns)
}

func describeELBV2Tags(c AWSCloud, loadBalancerArns []string) (map[string][]*elbv2.Tag, error) {
	// TODO: Filter by cluster?

	request := &elbv2.DescribeTagsInput{}
	request.ResourceArns = aws.StringSlice(loadBalancerArns)

	// TODO: Cache?
	klog.V(2).Infof("Querying ELBV2 api for tags for %s", loadBalancerArns)
	response, err := c.ELBV2().DescribeTags(request)
	if err != nil {
		return nil, err
	}

	tagMap := make(map[string][]*elbv2.Tag)
	for _, tagset := range response.TagDescriptions {
		tagMap[aws.StringValue(tagset.ResourceArn)] = tagset.Tags
	}
	return tagMap, nil
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
	return resolveImage(c.ssm, c.ec2, name)
}

func resolveSSMParameter(ssmClient ssmiface.SSMAPI, name string) (string, error) {
	klog.V(2).Infof("Resolving SSM parameter %q", name)
	request := &ssm.GetParameterInput{
		Name: aws.String(name),
	}

	response, err := ssmClient.GetParameter(request)
	if err != nil {
		return "", fmt.Errorf("failed to get value for SSM Parameter %q", name)
	}

	return aws.StringValue(response.Parameter.Value), nil
}

func resolveImage(ssmClient ssmiface.SSMAPI, ec2Client ec2iface.EC2API, name string) (*ec2.Image, error) {
	// TODO: Cache this result during a single execution (we get called multiple times)
	klog.V(2).Infof("Calling DescribeImages to resolve name %q", name)
	request := &ec2.DescribeImagesInput{}

	if strings.HasPrefix(name, "ami-") {
		// ami-xxxxxxxx
		request.ImageIds = []*string{&name}
	} else if strings.HasPrefix(name, "ssm:") {
		parameter := strings.TrimPrefix(name, "ssm:")

		image, err := resolveSSMParameter(ssmClient, parameter)
		if err != nil {
			return nil, err
		}

		request.ImageIds = []*string{&image}
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
			case "amazon", "amazon.com":
				owner = WellKnownAccountAmazonLinux2
			case "debian10":
				owner = WellKnownAccountDebian
			case "debian11":
				owner = WellKnownAccountDebian
			case "flatcar":
				owner = WellKnownAccountFlatcar
			case "redhat", "redhat.com":
				owner = WellKnownAccountRedhat
			case "ubuntu":
				owner = WellKnownAccountUbuntu
			}

			request.Owners = []*string{&owner}
			request.Filters = append(request.Filters, NewEC2Filter("name", tokens[1]))
		} else {
			return nil, fmt.Errorf("image name specification not recognized: %q", name)
		}
	}

	var image *ec2.Image
	err := ec2Client.DescribeImagesPagesWithContext(context.TODO(), request, func(output *ec2.DescribeImagesOutput, b bool) bool {
		for _, v := range output.Images {
			if image == nil {
				image = v
			} else {
				itime, _ := time.Parse(time.RFC3339, *image.CreationDate)
				vtime, _ := time.Parse(time.RFC3339, *v.CreationDate)
				if vtime.After(itime) {
					image = v
				}
			}
		}
		return true
	})
	if err != nil {
		return nil, fmt.Errorf("error listing images: %v", err)
	}
	if image == nil {
		return nil, fmt.Errorf("could not find Image for %q", name)
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
			return fmt.Errorf("error Zone is not a recognized AZ: %q (check you have specified a valid zone?)", zone)
		}

		for _, message := range z.Messages {
			klog.Warningf("Zone %q has message: %q", zone, aws.StringValue(message.Message))
		}

		if aws.StringValue(z.State) != ec2.AvailabilityZoneStateAvailable {
			klog.Warningf("Zone %q has state %q", zone, aws.StringValue(z.State))
		}
	}

	return nil
}

func (c *awsCloudImplementation) DNS() (dnsprovider.Interface, error) {
	provider, err := dnsprovider.GetDnsProvider(dnsproviderroute53.ProviderName, nil)
	if err != nil {
		return nil, fmt.Errorf("error building (k8s) DNS provider: %v", err)
	}
	return provider, nil
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

func (c *awsCloudImplementation) Spotinst() spotinst.Cloud {
	return c.spotinst
}

func (c *awsCloudImplementation) SQS() sqsiface.SQSAPI {
	return c.sqs
}

func (c *awsCloudImplementation) EventBridge() eventbridgeiface.EventBridgeAPI {
	return c.eventbridge
}

func (c *awsCloudImplementation) SSM() ssmiface.SSMAPI {
	return c.ssm
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

func (c *awsCloudImplementation) GetApiIngressStatus(cluster *kops.Cluster) ([]fi.ApiIngressStatus, error) {
	return getApiIngressStatus(c, cluster)
}

func getApiIngressStatus(c AWSCloud, cluster *kops.Cluster) ([]fi.ApiIngressStatus, error) {
	var ingresses []fi.ApiIngressStatus
	if lbDnsName, err := findDNSName(c, cluster); err != nil {
		return nil, fmt.Errorf("error finding aws DNSName: %v", err)
	} else if lbDnsName != "" {
		ingresses = append(ingresses, fi.ApiIngressStatus{Hostname: lbDnsName})
	}

	return ingresses, nil
}

func findDNSName(c AWSCloud, cluster *kops.Cluster) (string, error) {
	name := "api." + cluster.Name
	if cluster.Spec.API.LoadBalancer == nil {
		return "", nil
	}
	if cluster.Spec.API.LoadBalancer.Class == kops.LoadBalancerClassClassic {
		if lb, err := c.FindELBByNameTag(name); err != nil {
			return "", fmt.Errorf("error looking for AWS ELB: %v", err)
		} else if lb != nil {
			return aws.StringValue(lb.DNSName), nil
		}
	} else if cluster.Spec.API.LoadBalancer.Class == kops.LoadBalancerClassNetwork {
		if lb, err := c.FindELBV2ByNameTag(name); err != nil {
			return "", fmt.Errorf("error looking for AWS NLB: %v", err)
		} else if lb != nil {
			return aws.StringValue(lb.DNSName), nil
		}
	}
	return "", nil
}

// DefaultInstanceType determines an instance type for the specified cluster & instance group
func (c *awsCloudImplementation) DefaultInstanceType(cluster *kops.Cluster, ig *kops.InstanceGroup) (string, error) {
	var candidates []string

	switch ig.Spec.Role {
	case kops.InstanceGroupRoleControlPlane, kops.InstanceGroupRoleNode, kops.InstanceGroupRoleAPIServer:
		// t3.medium is the cheapest instance with 4GB of mem, unlimited by default, fast and has decent network
		// c5.large and c4.large are a good second option in case t3.medium is not available in the AZ
		candidates = []string{"t3.medium", "c5.large", "c4.large", "t4g.medium"}

	case kops.InstanceGroupRoleBastion:
		candidates = []string{"t3.micro", "t2.micro", "t4g.micro"}

	default:
		return "", fmt.Errorf("unhandled role %q", ig.Spec.Role)
	}

	imageArch := "x86_64"
	if imageInfo, err := c.ResolveImage(ig.Spec.Image); err == nil {
		imageArch = fi.ValueOf(imageInfo.Architecture)
	}

	// Find the AZs the InstanceGroup targets
	igZones, err := model.FindZonesForInstanceGroup(cluster, ig)
	if err != nil {
		return "", err
	}
	igZonesSet := sets.NewString(igZones...)

	// TODO: Validate that instance type exists in all AZs, but skip AZs that don't support any VPC stuff
	var reasons []string
	for _, instanceType := range candidates {
		if strings.HasPrefix(instanceType, "t4g") {
			if imageArch != "arm64" {
				reasons = append(reasons, fmt.Sprintf("instance type %q does not match image architecture %q", instanceType, imageArch))
				continue
			}
		} else {
			if imageArch == "arm64" {
				reasons = append(reasons, fmt.Sprintf("instance type %q does not match image architecture %q", instanceType, imageArch))
				continue
			}
		}

		zones, err := c.zonesWithInstanceType(instanceType)
		if err != nil {
			return "", err
		}
		if zones.IsSuperset(igZonesSet) {
			return instanceType, nil
		} else {
			reasons = append(reasons, fmt.Sprintf("instance type %q is not available in all zones (available in zones %v, need %v)", instanceType, zones, igZones))
			klog.V(2).Infof("can't use instance type %q, available in zones %v but need %v", instanceType, zones, igZones)
		}
	}

	// Log the detailed reasons why we can't find an instance type
	klog.Warning("cannot find suitable instance type")
	for _, reason := range reasons {
		klog.Warning("  *  " + reason)
	}
	return "", fmt.Errorf("could not find a suitable supported instance type for the instance group %q (type %q) in region %q", ig.Name, ig.Spec.Role, c.region)
}

// supportsInstanceType uses the DescribeReservedInstancesOfferings API call to determine if an instance type is supported in a region
func (c *awsCloudImplementation) zonesWithInstanceType(instanceType string) (sets.String, error) {
	klog.V(4).Infof("checking if instance type %q is supported in region %q", instanceType, c.region)
	request := &ec2.DescribeReservedInstancesOfferingsInput{}
	request.InstanceTenancy = aws.String("default")
	request.IncludeMarketplace = aws.Bool(false)
	request.OfferingClass = aws.String(ec2.OfferingClassTypeStandard)
	request.OfferingType = aws.String(ec2.OfferingTypeValuesNoUpfront)
	request.ProductDescription = aws.String(ec2.RIProductDescriptionLinuxUnixamazonVpc)
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

// DescribeInstanceType calls ec2.DescribeInstanceType to get information for a particular instance type
func (c *awsCloudImplementation) DescribeInstanceType(instanceType string) (*ec2.InstanceTypeInfo, error) {
	if info, ok := c.instanceTypes.typeMap[instanceType]; ok {
		return info, nil
	}
	c.instanceTypes.mutex.Lock()
	defer c.instanceTypes.mutex.Unlock()

	info, err := describeInstanceType(c, instanceType)
	if err != nil {
		return nil, err
	}
	c.instanceTypes.typeMap[instanceType] = info
	return info, nil
}

func describeInstanceType(c AWSCloud, instanceType string) (*ec2.InstanceTypeInfo, error) {
	req := &ec2.DescribeInstanceTypesInput{
		InstanceTypes: aws.StringSlice([]string{instanceType}),
	}
	resp, err := c.EC2().DescribeInstanceTypes(req)
	if err != nil {
		return nil, fmt.Errorf("describing instance type %q in region %q: %w", instanceType, c.Region(), err)
	}
	if len(resp.InstanceTypes) != 1 {
		return nil, fmt.Errorf("instance type %q not found in region %q", instanceType, c.Region())
	}
	return resp.InstanceTypes[0], nil
}

// AccountInfo returns the AWS account ID and AWS partition that we are deploying into
func (c *awsCloudImplementation) AccountInfo() (string, string, error) {
	request := &sts.GetCallerIdentityInput{}

	response, err := c.sts.GetCallerIdentity(request)
	if err != nil {
		return "", "", fmt.Errorf("error geting AWS account ID: %v", err)
	}

	arn, err := arn.Parse(aws.StringValue(response.Arn))
	if err != nil {
		return "", "", fmt.Errorf("Failed to parse GetCallerIdentity ARN")
	}

	if arn.AccountID == "" {
		return "", "", fmt.Errorf("AWS account id was empty")
	}
	if arn.Partition == "" {
		return "", "", fmt.Errorf("AWS partition was empty")
	}
	return arn.AccountID, arn.Partition, nil
}

// GetRolesInInstanceProfile return role names which are associated with the instance profile specified by profileName.
func GetRolesInInstanceProfile(c AWSCloud, profileName string) ([]string, error) {
	output, err := c.IAM().GetInstanceProfile(&iam.GetInstanceProfileInput{
		InstanceProfileName: aws.String(profileName),
	})
	if err != nil {
		return nil, err
	}
	var roleNames []string
	for _, role := range output.InstanceProfile.Roles {
		roleNames = append(roleNames, *role.RoleName)
	}
	return roleNames, nil
}

// GetInstanceCertificateNames returns the instance hostname and addresses that should go into certificates.
// The first value is the node name and any additional values are the DNS name and IP addresses.
func GetInstanceCertificateNames(instances *ec2.DescribeInstancesOutput, useInstanceIDForNodeName bool) (addrs []string, err error) {
	if len(instances.Reservations) != 1 {
		return nil, fmt.Errorf("too many reservations returned for the single instance-id")
	}

	if len(instances.Reservations[0].Instances) != 1 {
		return nil, fmt.Errorf("too many instances returned for the single instance-id")
	}

	instance := instances.Reservations[0].Instances[0]

	if useInstanceIDForNodeName {
		addrs = append(addrs, *instance.InstanceId)
	}

	if instance.PrivateDnsName != nil {
		addrs = append(addrs, *instance.PrivateDnsName)
	}

	// We only use data for the first interface, and only the first IP
	for _, iface := range instance.NetworkInterfaces {
		if iface.Attachment == nil {
			continue
		}
		if *iface.Attachment.DeviceIndex != 0 {
			continue
		}
		if iface.PrivateIpAddress != nil {
			addrs = append(addrs, *iface.PrivateIpAddress)
		}
		if iface.Ipv6Addresses != nil && len(iface.Ipv6Addresses) > 0 {
			addrs = append(addrs, *iface.Ipv6Addresses[0].Ipv6Address)
		}
		if iface.Association != nil && iface.Association.PublicIp != nil {
			addrs = append(addrs, *iface.Association.PublicIp)
		}
	}
	return addrs, nil
}
