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

package awstasks

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/aws/aws-sdk-go/service/route53"
	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/cloudformation"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/util/pkg/slice"
)

// NetworkLoadBalancer manages an NLB.  We find the existing NLB using the Name tag.
var _ DNSTarget = &NetworkLoadBalancer{}

//go:generate fitask -type=NetworkLoadBalancer
type NetworkLoadBalancer struct {
	// We use the Name tag to find the existing NLB, because we are (more or less) unrestricted when
	// it comes to tag values, but the LoadBalancerName is length limited
	Name      *string
	Lifecycle *fi.Lifecycle

	// LoadBalancerName is the name in NLB, possibly different from our name
	// (NLB is restricted as to names, so we have limited choices!)
	// We use the Name tag to find the existing NLB.
	LoadBalancerName *string

	DNSName      *string
	HostedZoneId *string

	Subnets        []*Subnet
	SecurityGroups []*SecurityGroup

	Listeners map[string]*NetworkLoadBalancerListener

	Scheme *string

	HealthCheck            *NetworkLoadBalancerHealthCheck
	AccessLog              *NetworkLoadBalancerAccessLog
	CrossZoneLoadBalancing *NetworkLoadBalancerCrossZoneLoadBalancing
	SSLCertificateID       string

	Tags         map[string]string
	ForAPIServer bool

	Type *string

	VPC                *VPC
	DeletionProtection *NetworkLoadBalancerDeletionProtection
	ProxyProtocolV2    *TargetGroupProxyProtocolV2
	Stickiness         *TargetGroupStickiness
	DeregistationDelay *TargetGroupDeregistrationDelay
}

var _ fi.CompareWithID = &NetworkLoadBalancer{}

func (e *NetworkLoadBalancer) CompareWithID() *string {
	return e.Name
}

type NetworkLoadBalancerListener struct {
	InstancePort     int //TODO: Change this to LoadBalancerPort
	SSLCertificateID string
}

func (e *NetworkLoadBalancerListener) mapToAWS(loadBalancerPort int64, targetGroupArn string, loadBalancerArn string) *elbv2.CreateListenerInput {

	l := &elbv2.CreateListenerInput{
		DefaultActions: []*elbv2.Action{
			{
				TargetGroupArn: aws.String(targetGroupArn),
				Type:           aws.String("forward"),
			},
		},
		LoadBalancerArn: aws.String(loadBalancerArn),
		Port:            aws.Int64(loadBalancerPort),
	}

	if e.SSLCertificateID != "" {
		l.Certificates = []*elbv2.Certificate{}
		l.Certificates = append(l.Certificates, &elbv2.Certificate{
			CertificateArn: aws.String(e.SSLCertificateID),
		})
		l.Protocol = aws.String("SSL")
	} else {
		l.Protocol = aws.String("TCP")
	}

	return l
}

var _ fi.HasDependencies = &NetworkLoadBalancerListener{}

func (e *NetworkLoadBalancerListener) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	return nil
}

func findTargetGroupByLoadBalancerArn(cloud awsup.AWSCloud, loadBalancerArn string) (*elbv2.TargetGroup, error) {
	request := &elbv2.DescribeTargetGroupsInput{
		LoadBalancerArn: aws.String(loadBalancerArn),
	}

	response, err := cloud.ELBV2().DescribeTargetGroups(request)

	if err != nil {
		return nil, fmt.Errorf("Error retrieving target groups for loadBalancerArn %v with err : %v", loadBalancerArn, err)
	}

	if len(response.TargetGroups) != 1 {
		return nil, fmt.Errorf("Wrong # of target groups returned in findTargetGroupByLoadBalancerName for name %v", loadBalancerArn)
	}

	return response.TargetGroups[0], nil
}

func findTargetGroupByLoadBalancerName(cloud awsup.AWSCloud, loadBalancerNameTag string) (*elbv2.TargetGroup, error) {

	lb, err := FindNetworkLoadBalancerByNameTag(cloud, loadBalancerNameTag)
	if err != nil {
		return nil, fmt.Errorf("Can't locate NLB with Name Tag %v in findTargetGroupByLoadBalancerName : %v", loadBalancerNameTag, err)
	}

	if lb == nil {
		return nil, nil
	}

	return findTargetGroupByLoadBalancerArn(cloud, *lb.LoadBalancerArn)
}

//The load balancer name 'api.renamenlbcluster.k8s.local' can only contain characters that are alphanumeric characters and hyphens(-)\n\tstatus code: 400,
func findNetworkLoadBalancerByLoadBalancerName(cloud awsup.AWSCloud, loadBalancerName string) (*elbv2.LoadBalancer, error) {
	request := &elbv2.DescribeLoadBalancersInput{
		Names: []*string{&loadBalancerName},
	}
	found, err := describeNetworkLoadBalancers(cloud, request, func(lb *elbv2.LoadBalancer) bool {
		// TODO: Filter by cluster?

		if aws.StringValue(lb.LoadBalancerName) == loadBalancerName {
			return true
		}

		klog.Warningf("Got NLB with unexpected name: %q", aws.StringValue(lb.LoadBalancerName))
		return false
	})

	if err != nil {
		if awsError, ok := err.(awserr.Error); ok {
			if awsError.Code() == "LoadBalancerNotFound" {
				return nil, nil
			}
		}

		return nil, fmt.Errorf("error listing NLBs: %v", err)
	}

	if len(found) == 0 {
		return nil, nil
	}

	if len(found) != 1 {
		return nil, fmt.Errorf("Found multiple NLBs with name %q", loadBalancerName)
	}

	return found[0], nil
}

func findNetworkLoadBalancerByAlias(cloud awsup.AWSCloud, alias *route53.AliasTarget) (*elbv2.LoadBalancer, error) {
	// TODO: Any way to avoid listing all NLBs?
	request := &elbv2.DescribeLoadBalancersInput{}

	dnsName := aws.StringValue(alias.DNSName)
	matchDnsName := strings.TrimSuffix(dnsName, ".")
	if matchDnsName == "" {
		return nil, fmt.Errorf("DNSName not set on AliasTarget")
	}

	matchHostedZoneId := aws.StringValue(alias.HostedZoneId)

	found, err := describeNetworkLoadBalancers(cloud, request, func(lb *elbv2.LoadBalancer) bool {
		// TODO: Filter by cluster?

		if matchHostedZoneId != aws.StringValue(lb.CanonicalHostedZoneId) {
			return false
		}

		lbDnsName := aws.StringValue(lb.DNSName)
		lbDnsName = strings.TrimSuffix(lbDnsName, ".")
		return lbDnsName == matchDnsName || "dualstack."+lbDnsName == matchDnsName
	})

	if err != nil {
		return nil, fmt.Errorf("error listing NLBs: %v", err)
	}

	if len(found) == 0 {
		return nil, nil
	}

	if len(found) != 1 {
		return nil, fmt.Errorf("Found multiple NLBs with DNSName %q", dnsName)
	}

	return found[0], nil
}

func FindNetworkLoadBalancerByNameTag(cloud awsup.AWSCloud, findNameTag string) (*elbv2.LoadBalancer, error) {
	// TODO: Any way around this?
	klog.V(2).Infof("Listing all NLBs for findNetworkLoadBalancerByNameTag")

	request := &elbv2.DescribeLoadBalancersInput{}
	// ELB DescribeTags has a limit of 20 names, so we set the page size here to 20 also
	request.PageSize = aws.Int64(20)

	var found []*elbv2.LoadBalancer

	var innerError error
	err := cloud.ELBV2().DescribeLoadBalancersPages(request, func(p *elbv2.DescribeLoadBalancersOutput, lastPage bool) bool {
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

		tagMap, err := describeNetworkLoadBalancerTags(cloud, arns)
		if err != nil {
			innerError = err
			return false
		}

		for loadBalancerArn, tags := range tagMap {
			name, foundNameTag := awsup.FindELBV2Tag(tags, "Name")
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

func describeNetworkLoadBalancers(cloud awsup.AWSCloud, request *elbv2.DescribeLoadBalancersInput, filter func(*elbv2.LoadBalancer) bool) ([]*elbv2.LoadBalancer, error) {
	var found []*elbv2.LoadBalancer
	err := cloud.ELBV2().DescribeLoadBalancersPages(request, func(p *elbv2.DescribeLoadBalancersOutput, lastPage bool) (shouldContinue bool) {
		for _, lb := range p.LoadBalancers {
			if filter(lb) {
				found = append(found, lb)
			}
		}

		return true
	})

	if err != nil {
		return nil, fmt.Errorf("error listing NLBs: %v", err)
	}

	return found, nil
}

func describeNetworkLoadBalancerTags(cloud awsup.AWSCloud, loadBalancerArns []string) (map[string][]*elbv2.Tag, error) {
	// TODO: Filter by cluster?

	request := &elbv2.DescribeTagsInput{}
	request.ResourceArns = aws.StringSlice(loadBalancerArns)

	// TODO: Cache?
	klog.V(2).Infof("Querying ELBV2 api for tags for %s", loadBalancerArns)
	response, err := cloud.ELBV2().DescribeTags(request)
	if err != nil {
		return nil, err
	}

	tagMap := make(map[string][]*elbv2.Tag)
	for _, tagset := range response.TagDescriptions {
		tagMap[aws.StringValue(tagset.ResourceArn)] = tagset.Tags
	}
	return tagMap, nil
}

func (e *NetworkLoadBalancer) getDNSName() *string {
	return e.DNSName
}

func (e *NetworkLoadBalancer) getHostedZoneId() *string {
	return e.HostedZoneId
}

func (e *NetworkLoadBalancer) Find(c *fi.Context) (*NetworkLoadBalancer, error) {
	cloud := c.Cloud.(awsup.AWSCloud)

	lb, err := FindNetworkLoadBalancerByNameTag(cloud, fi.StringValue(e.Name))
	if err != nil {
		return nil, err
	}
	if lb == nil {
		return nil, nil
	}

	loadBalancerArn := lb.LoadBalancerArn
	var targetGroupArn *string

	actual := &NetworkLoadBalancer{}
	actual.Name = e.Name
	actual.Lifecycle = e.Lifecycle
	actual.LoadBalancerName = lb.LoadBalancerName
	actual.DNSName = lb.DNSName
	actual.HostedZoneId = lb.CanonicalHostedZoneId //CanonicalHostedZoneNameID
	actual.Scheme = lb.Scheme
	actual.VPC = &VPC{ID: lb.VpcId}
	actual.Type = lb.Type

	tagMap, err := describeNetworkLoadBalancerTags(cloud, []string{*loadBalancerArn})
	if err != nil {
		return nil, err
	}
	actual.Tags = make(map[string]string)
	for _, tag := range tagMap[*loadBalancerArn] {
		actual.Tags[aws.StringValue(tag.Key)] = aws.StringValue(tag.Value)
	}

	for _, az := range lb.AvailabilityZones {
		actual.Subnets = append(actual.Subnets, &Subnet{ID: az.SubnetId})
	}

	/*for _, sg := range lb.SecurityGroups {
		actual.SecurityGroups = append(actual.SecurityGroups, &SecurityGroup{ID: sg})
	}*/

	{
		//What happens if someone manually creates additional target groups for this LB?
		request := &elbv2.DescribeTargetGroupsInput{
			LoadBalancerArn: loadBalancerArn,
		}
		response, err := cloud.ELBV2().DescribeTargetGroups(request)
		if err != nil {
			return nil, fmt.Errorf("error querying for NLB Target groups :%v", err)
		}

		if len(response.TargetGroups) == 0 {
			return nil, fmt.Errorf("Found no Target Groups for NLB - misconfiguration :  %v", loadBalancerArn)
		}

		if len(response.TargetGroups) != 1 {
			return nil, fmt.Errorf("Found multiple Target groups for NLB with arn %v", loadBalancerArn)
		}

		targetGroupArn = response.TargetGroups[0].TargetGroupArn
	}

	{
		request := &elbv2.DescribeListenersInput{
			LoadBalancerArn: loadBalancerArn,
		}
		response, err := cloud.ELBV2().DescribeListeners(request)
		if err != nil {
			return nil, fmt.Errorf("error querying for NLB listeners :%v", err)
		}

		actual.Listeners = make(map[string]*NetworkLoadBalancerListener)

		for _, l := range response.Listeners {
			loadBalancerPort := strconv.FormatInt(aws.Int64Value(l.Port), 10)

			actualListener := &NetworkLoadBalancerListener{}
			actualListener.InstancePort = int(aws.Int64Value(l.Port))
			if len(l.Certificates) != 0 {
				actualListener.SSLCertificateID = aws.StringValue(l.Certificates[0].CertificateArn) // What if there is more then one certificate, can we just grab the default certificate? we don't set it as default, we only set the one.
			}
			actual.Listeners[loadBalancerPort] = actualListener
		}

	}

	healthcheck, err := findNLBHealthCheck(cloud, lb)
	if err != nil {
		return nil, err
	}
	actual.HealthCheck = healthcheck

	{
		lbAttributes, err := findNetworkLoadBalancerAttributes(cloud, aws.StringValue(loadBalancerArn))
		if err != nil {
			return nil, err
		}
		klog.V(4).Infof("NLB Load Balancer attributes: %+v", lbAttributes)

		actual.AccessLog = &NetworkLoadBalancerAccessLog{}
		actual.DeletionProtection = &NetworkLoadBalancerDeletionProtection{}
		actual.CrossZoneLoadBalancing = &NetworkLoadBalancerCrossZoneLoadBalancing{}
		for _, attribute := range lbAttributes {
			if attribute.Value == nil {
				continue
			}
			switch key, value := attribute.Key, attribute.Value; *key {
			case "access_logs.s3.enabled":
				b, err := strconv.ParseBool(*value)
				if err != nil {
					return nil, err
				}
				actual.AccessLog.Enabled = fi.Bool(b)
			case "access_logs.s3.bucket":
				actual.AccessLog.S3BucketName = value
			case "access_logs.s3.prefix":
				actual.AccessLog.S3BucketPrefix = value
			case "deletion_protection.enabled":
				b, err := strconv.ParseBool(*value)
				if err != nil {
					return nil, err
				}
				actual.DeletionProtection.Enabled = fi.Bool(b)
			case "load_balancing.cross_zone.enabled":
				b, err := strconv.ParseBool(*value)
				if err != nil {
					return nil, err
				}
				actual.CrossZoneLoadBalancing.Enabled = fi.Bool(b)
			default:
				klog.V(2).Infof("unsupported key -- ignoring, %v.\n", key)
			}
		}
	}

	{
		tgAttributes, err := findTargetGroupAttributes(cloud, aws.StringValue(targetGroupArn))
		if err != nil {
			return nil, err
		}
		klog.V(4).Infof("Target Group attributes: %+v", tgAttributes)

		actual.ProxyProtocolV2 = &TargetGroupProxyProtocolV2{}
		actual.Stickiness = &TargetGroupStickiness{}
		actual.DeregistationDelay = &TargetGroupDeregistrationDelay{}
		for _, attribute := range tgAttributes {
			if attribute.Value == nil {
				continue
			}
			switch key, value := attribute.Key, attribute.Value; *key {
			case "proxy_protocol_v2.enabled":
				b, err := strconv.ParseBool(*value)
				if err != nil {
					return nil, err
				}
				actual.ProxyProtocolV2.Enabled = fi.Bool(b)
			case "stickiness.type":
				actual.Stickiness.Type = value
			case "stickiness.enabled":
				b, err := strconv.ParseBool(*value)
				if err != nil {
					return nil, err
				}
				actual.Stickiness.Enabled = fi.Bool(b)
			case "deregistration_delay.timeout_seconds":
				if n, err := strconv.Atoi(*value); err == nil {
					m := int64(n)
					actual.DeregistationDelay.TimeoutSeconds = fi.Int64(m)
				} else {
					return nil, err
				}

			default:
				klog.V(2).Infof("unsupported key -- ignoring, %v.\n", key)
			}
		}
	}

	// Avoid spurious mismatches
	if subnetSlicesEqualIgnoreOrder(actual.Subnets, e.Subnets) {
		actual.Subnets = e.Subnets
	}
	if e.DNSName == nil {
		e.DNSName = actual.DNSName
	}
	if e.HostedZoneId == nil {
		e.HostedZoneId = actual.HostedZoneId
	}
	if e.LoadBalancerName == nil {
		e.LoadBalancerName = actual.LoadBalancerName
	}

	// We allow for the LoadBalancerName to be wrong:
	// 1. We don't want to force a rename of the NLB, because that is a destructive operation
	if fi.StringValue(e.LoadBalancerName) != fi.StringValue(actual.LoadBalancerName) {
		klog.V(2).Infof("Reusing existing load balancer with name: %q", aws.StringValue(actual.LoadBalancerName))
		e.LoadBalancerName = actual.LoadBalancerName
	}

	// TODO: Make Normalize a standard method
	actual.Normalize()

	klog.V(4).Infof("Found NLB %+v", actual)

	return actual, nil
}

var _ fi.HasAddress = &NetworkLoadBalancer{}

func (e *NetworkLoadBalancer) IsForAPIServer() bool {
	return e.ForAPIServer
}

func (e *NetworkLoadBalancer) FindIPAddress(context *fi.Context) (*string, error) {
	cloud := context.Cloud.(awsup.AWSCloud)

	lb, err := FindNetworkLoadBalancerByNameTag(cloud, fi.StringValue(e.Name))
	if err != nil {
		return nil, err
	}
	if lb == nil {
		return nil, nil
	}

	lbDnsName := fi.StringValue(lb.DNSName)
	if lbDnsName == "" {
		return nil, nil
	}
	return &lbDnsName, nil
}

func (e *NetworkLoadBalancer) Run(c *fi.Context) error {
	// TODO: Make Normalize a standard method
	e.Normalize()

	return fi.DefaultDeltaRunMethod(e, c)
}

func (e *NetworkLoadBalancer) Normalize() {
	// We need to sort our arrays consistently, so we don't get spurious changes
	sort.Stable(OrderSubnetsById(e.Subnets))
	sort.Stable(OrderSecurityGroupsById(e.SecurityGroups))
}

func (s *NetworkLoadBalancer) CheckChanges(a, e, changes *NetworkLoadBalancer) error {
	if a == nil {
		if fi.StringValue(e.Name) == "" {
			return fi.RequiredField("Name")
		}
		// if len(e.SecurityGroups) == 0 {
		// 	return fi.RequiredField("SecurityGroups")
		// }
		if len(e.Subnets) == 0 {
			return fi.RequiredField("Subnets")
		}

		if e.CrossZoneLoadBalancing != nil {
			if e.CrossZoneLoadBalancing.Enabled == nil {
				return fi.RequiredField("CrossZoneLoadBalancing.Enabled")
			}
		}
	}

	return nil
}

func (_ *NetworkLoadBalancer) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *NetworkLoadBalancer) error {
	var loadBalancerName string
	var loadBalancerArn string
	var targetGroupArn string

	if a == nil {
		if e.LoadBalancerName == nil {
			return fi.RequiredField("LoadBalancerName")
		}
		loadBalancerName = *e.LoadBalancerName

		request := &elbv2.CreateLoadBalancerInput{}
		request.Name = e.LoadBalancerName
		request.Scheme = e.Scheme
		request.Type = e.Type

		for _, subnet := range e.Subnets {
			request.Subnets = append(request.Subnets, subnet.ID)
		}

		//request.SecurityGroups = append(request.SecurityGroups, sg.ID)

		/*for _, sg := range e.SecurityGroups {
			request.SecurityGroups = append(request.SecurityGroups, sg.ID)
		}*/

		{
			klog.V(2).Infof("Creating NLB with Name:%q", loadBalancerName)

			response, err := t.Cloud.ELBV2().CreateLoadBalancer(request)
			if err != nil {
				return fmt.Errorf("error creating NLB: %v", err)
			}

			if len(response.LoadBalancers) != 1 {
				return fmt.Errorf("Either too many or too few NLBs were created, wanted to find %q", loadBalancerName)
			} else {
				lb := response.LoadBalancers[0] //TODO: how to avoid doing this
				e.DNSName = lb.DNSName
				e.HostedZoneId = lb.CanonicalHostedZoneId
				loadBalancerArn = fi.StringValue(lb.LoadBalancerArn)
			}
		}

		{
			prefix := loadBalancerName[:24]
			targetGroupName := prefix + "-targets"
			//TODO: GET 443/TCP FROM e.loadbalancer
			request := &elbv2.CreateTargetGroupInput{
				Name:     aws.String(targetGroupName),
				Port:     aws.Int64(443),
				Protocol: aws.String("TCP"),
				VpcId:    e.VPC.ID,
			}

			klog.V(2).Infof("Creating Target Group for NLB")
			response, err := t.Cloud.ELBV2().CreateTargetGroup(request)
			if err != nil {
				return fmt.Errorf("Error creating target group for NLB : %v", err)
			}

			targetGroupArn = *response.TargetGroups[0].TargetGroupArn

			if err := t.AddELBV2Tags(targetGroupArn, e.Tags); err != nil {
				return err
			}
		}

		{
			for loadBalancerPort, listener := range e.Listeners {
				loadBalancerPortInt, err := strconv.ParseInt(loadBalancerPort, 10, 64)
				if err != nil {
					return fmt.Errorf("error parsing load balancer listener port: %q", loadBalancerPort)
				}
				awsListener := listener.mapToAWS(loadBalancerPortInt, targetGroupArn, loadBalancerArn)

				klog.V(2).Infof("Creating Listener for NLB")
				_, err = t.Cloud.ELBV2().CreateListener(awsListener)
				if err != nil {
					return fmt.Errorf("Error creating listener for NLB: %v", err)
				}
			}
		}
	} else {
		loadBalancerName = fi.StringValue(a.LoadBalancerName)

		lb, err := findNetworkLoadBalancerByLoadBalancerName(t.Cloud, loadBalancerName)
		if err != nil {
			return fmt.Errorf("error getting load balancer by name: %v", err)
		}

		// if lb == nil {
		// 	return fmt.Errorf("error querying nlb: %v", err)
		// }

		loadBalancerArn = *lb.LoadBalancerArn
		tg, err := findTargetGroupByLoadBalancerArn(t.Cloud, loadBalancerArn)
		if err != nil {
			return fmt.Errorf("error getting target group by lb arn %v", loadBalancerArn)
		}

		targetGroupArn = *tg.TargetGroupArn

		if changes.Subnets != nil {
			var expectedSubnets []string
			for _, s := range e.Subnets {
				expectedSubnets = append(expectedSubnets, fi.StringValue(s.ID))
			}

			var actualSubnets []string
			for _, s := range a.Subnets {
				actualSubnets = append(actualSubnets, fi.StringValue(s.ID))
			}

			oldSubnetIDs := slice.GetUniqueStrings(expectedSubnets, actualSubnets)
			if len(oldSubnetIDs) > 0 {
				/*request := &elb.DetachLoadBalancerFromSubnetsInput{}
				request.SetLoadBalancerName(loadBalancerName)
				request.SetSubnets(aws.StringSlice(oldSubnetIDs))

				klog.V(2).Infof("Detaching Load Balancer from old subnets")
				if _, err := t.Cloud.ELB().DetachLoadBalancerFromSubnets(request); err != nil {
					return fmt.Errorf("Error detaching Load Balancer from old subnets: %v", err)
				}*/
				//TODO: Seems likely we would have to delete and recreate in this case.
				return fmt.Errorf("NLB's don't support detaching subnets, perhaps we need to recreate the NLB")
			}

			newSubnetIDs := slice.GetUniqueStrings(actualSubnets, expectedSubnets)
			if len(newSubnetIDs) > 0 {

				request := &elbv2.SetSubnetsInput{}
				request.SetLoadBalancerArn(loadBalancerArn)
				request.SetSubnets(aws.StringSlice(append(actualSubnets, newSubnetIDs...)))

				klog.V(2).Infof("Attaching Load Balancer to new subnets")
				if _, err := t.Cloud.ELBV2().SetSubnets(request); err != nil {
					return fmt.Errorf("Error attaching Load Balancer to new subnets: %v", err)
				}
			}
		}

		//TODO: decide if security groups should be applied to master nodes
		/*if changes.SecurityGroups != nil {
			request := &elb.ApplySecurityGroupsToLoadBalancerInput{}
			request.LoadBalancerName = aws.String(loadBalancerName)
			for _, sg := range e.SecurityGroups {
				request.SecurityGroups = append(request.SecurityGroups, sg.ID)
			}

			klog.V(2).Infof("Updating Load Balancer Security Groups")
			if _, err := t.Cloud.ELB().ApplySecurityGroupsToLoadBalancer(request); err != nil {
				return fmt.Errorf("Error updating security groups on Load Balancer: %v", err)
			}
		}*/

		if changes.Listeners != nil {

			if lb != nil {

				request := &elbv2.DescribeListenersInput{
					LoadBalancerArn: lb.LoadBalancerArn,
				}
				response, err := t.Cloud.ELBV2().DescribeListeners(request)
				if err != nil {
					return fmt.Errorf("error querying for NLB listeners :%v", err)
				}

				for _, l := range response.Listeners {
					// delete the listener before recreating it
					_, err := t.Cloud.ELBV2().DeleteListener(&elbv2.DeleteListenerInput{
						ListenerArn: l.ListenerArn,
					})
					if err != nil {
						return fmt.Errorf("error deleting load balancer listener with arn = : %v : %v", l.ListenerArn, err)
					}
				}
			}

			for loadBalancerPort, listener := range changes.Listeners {
				loadBalancerPortInt, err := strconv.ParseInt(loadBalancerPort, 10, 64)
				if err != nil {
					return fmt.Errorf("error parsing load balancer listener port: %q", loadBalancerPort)
				}

				awsListener := listener.mapToAWS(loadBalancerPortInt, targetGroupArn, loadBalancerArn)

				klog.V(2).Infof("Creating Listener for NLB")
				_, err = t.Cloud.ELBV2().CreateListener(awsListener)
				if err != nil {
					return fmt.Errorf("Error creating listener for NLB: %v", err)
				}
			}
		}
	}

	if err := t.AddELBV2Tags(loadBalancerArn, e.Tags); err != nil {
		return err
	}

	//TODO: why is this used in load_balancer.go seems unnecessary to remove tags right after adding them.
	/*if err := t.RemoveELBV2Tags(loadBalancerArn, e.Tags); err != nil {
		return err
	}*/

	if changes.HealthCheck != nil && e.HealthCheck != nil {
		request := &elbv2.ModifyTargetGroupInput{
			HealthCheckPort:         e.HealthCheck.Port,
			TargetGroupArn:          aws.String(targetGroupArn),
			HealthyThresholdCount:   e.HealthCheck.HealthyThreshold,
			UnhealthyThresholdCount: e.HealthCheck.UnhealthyThreshold,
		}

		klog.V(2).Infof("Configuring health checks on NLB %q", loadBalancerName)
		_, err := t.Cloud.ELBV2().ModifyTargetGroup(request)
		if err != nil {
			return fmt.Errorf("error configuring health checks on NLB: %v's target group", err)
		}
	}

	if err := e.modifyLoadBalancerAttributes(t, a, e, changes, loadBalancerArn); err != nil {
		klog.Infof("error modifying NLB attributes: %v", err)
		return err
	}

	if err := e.modifyTargetGroupAttributes(t, a, e, changes, targetGroupArn); err != nil {
		klog.Infof("error modifying NLB Target Group attributes: %v", err)
		return err
	}

	return nil
}

func (e *NetworkLoadBalancer) TerraformLink(params ...string) *terraform.Literal {
	if true {
		panic("NetworkLoadBalancer support for Terraform TBD")
	}
	return nil
}

func (e *NetworkLoadBalancer) CloudformationAttrCanonicalHostedZoneNameID() *cloudformation.Literal {
	if true {
		panic("NetworkLoadBalancer does support for Cloudformation TBD")
	}
	return nil
}

func (e *NetworkLoadBalancer) CloudformationAttrDNSName() *cloudformation.Literal {
	if true {
		panic("NetworkLoadBalancer does support for Cloudformation TBD")
	}
	return nil
}
