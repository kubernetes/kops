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
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/route53"
	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/cloudformation"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/util/pkg/slice"
)

// LoadBalancer manages an ELB.  We find the existing ELB using the Name tag.

//go:generate fitask -type=LoadBalancer
type LoadBalancer struct {
	// We use the Name tag to find the existing ELB, because we are (more or less) unrestricted when
	// it comes to tag values, but the LoadBalancerName is length limited
	Name      *string
	Lifecycle *fi.Lifecycle

	// LoadBalancerName is the name in ELB, possibly different from our name
	// (ELB is restricted as to names, so we have limited choices!)
	// We use the Name tag to find the existing ELB.
	LoadBalancerName *string

	DNSName      *string
	HostedZoneId *string

	Subnets        []*Subnet
	SecurityGroups []*SecurityGroup

	Listeners map[string]*LoadBalancerListener

	Scheme *string

	HealthCheck            *LoadBalancerHealthCheck
	AccessLog              *LoadBalancerAccessLog
	ConnectionDraining     *LoadBalancerConnectionDraining
	ConnectionSettings     *LoadBalancerConnectionSettings
	CrossZoneLoadBalancing *LoadBalancerCrossZoneLoadBalancing
	SSLCertificateID       string

	Tags map[string]string
}

var _ fi.CompareWithID = &LoadBalancer{}

func (e *LoadBalancer) CompareWithID() *string {
	return e.Name
}

type LoadBalancerListener struct {
	InstancePort     int
	SSLCertificateID string
}

func (e *LoadBalancerListener) mapToAWS(loadBalancerPort int64) *elb.Listener {
	l := &elb.Listener{
		LoadBalancerPort: aws.Int64(loadBalancerPort),
		InstancePort:     aws.Int64(int64(e.InstancePort)),
	}

	if e.SSLCertificateID != "" {
		l.Protocol = aws.String("SSL")
		l.InstanceProtocol = aws.String("SSL")
		l.SSLCertificateId = aws.String(e.SSLCertificateID)
	} else {
		l.Protocol = aws.String("TCP")
		l.InstanceProtocol = aws.String("TCP")
	}

	return l
}

var _ fi.HasDependencies = &LoadBalancerListener{}

func (e *LoadBalancerListener) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	return nil
}

func findLoadBalancerByLoadBalancerName(cloud awsup.AWSCloud, loadBalancerName string) (*elb.LoadBalancerDescription, error) {
	request := &elb.DescribeLoadBalancersInput{
		LoadBalancerNames: []*string{&loadBalancerName},
	}
	found, err := describeLoadBalancers(cloud, request, func(lb *elb.LoadBalancerDescription) bool {
		// TODO: Filter by cluster?

		if aws.StringValue(lb.LoadBalancerName) == loadBalancerName {
			return true
		}

		klog.Warningf("Got ELB with unexpected name: %q", aws.StringValue(lb.LoadBalancerName))
		return false
	})

	if err != nil {
		if awsError, ok := err.(awserr.Error); ok {
			if awsError.Code() == "LoadBalancerNotFound" {
				return nil, nil
			}
		}

		return nil, fmt.Errorf("error listing ELBs: %v", err)
	}

	if len(found) == 0 {
		return nil, nil
	}

	if len(found) != 1 {
		return nil, fmt.Errorf("Found multiple ELBs with name %q", loadBalancerName)
	}

	return found[0], nil
}

func findLoadBalancerByAlias(cloud awsup.AWSCloud, alias *route53.AliasTarget) (*elb.LoadBalancerDescription, error) {
	// TODO: Any way to avoid listing all ELBs?
	request := &elb.DescribeLoadBalancersInput{}

	dnsName := aws.StringValue(alias.DNSName)
	matchDnsName := strings.TrimSuffix(dnsName, ".")
	if matchDnsName == "" {
		return nil, fmt.Errorf("DNSName not set on AliasTarget")
	}

	matchHostedZoneId := aws.StringValue(alias.HostedZoneId)

	found, err := describeLoadBalancers(cloud, request, func(lb *elb.LoadBalancerDescription) bool {
		// TODO: Filter by cluster?

		if matchHostedZoneId != aws.StringValue(lb.CanonicalHostedZoneNameID) {
			return false
		}

		lbDnsName := aws.StringValue(lb.DNSName)
		lbDnsName = strings.TrimSuffix(lbDnsName, ".")
		return lbDnsName == matchDnsName || "dualstack."+lbDnsName == matchDnsName
	})

	if err != nil {
		return nil, fmt.Errorf("error listing ELBs: %v", err)
	}

	if len(found) == 0 {
		return nil, nil
	}

	if len(found) != 1 {
		return nil, fmt.Errorf("Found multiple ELBs with DNSName %q", dnsName)
	}

	return found[0], nil
}

func FindLoadBalancerByNameTag(cloud awsup.AWSCloud, findNameTag string) (*elb.LoadBalancerDescription, error) {
	// TODO: Any way around this?
	klog.V(2).Infof("Listing all ELBs for findLoadBalancerByNameTag")

	request := &elb.DescribeLoadBalancersInput{}
	// ELB DescribeTags has a limit of 20 names, so we set the page size here to 20 also
	request.PageSize = aws.Int64(20)

	var found []*elb.LoadBalancerDescription

	var innerError error
	err := cloud.ELB().DescribeLoadBalancersPages(request, func(p *elb.DescribeLoadBalancersOutput, lastPage bool) bool {
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

		tagMap, err := describeLoadBalancerTags(cloud, names)
		if err != nil {
			innerError = err
			return false
		}

		for loadBalancerName, tags := range tagMap {
			name, foundNameTag := awsup.FindELBTag(tags, "Name")
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

func describeLoadBalancers(cloud awsup.AWSCloud, request *elb.DescribeLoadBalancersInput, filter func(*elb.LoadBalancerDescription) bool) ([]*elb.LoadBalancerDescription, error) {
	var found []*elb.LoadBalancerDescription
	err := cloud.ELB().DescribeLoadBalancersPages(request, func(p *elb.DescribeLoadBalancersOutput, lastPage bool) (shouldContinue bool) {
		for _, lb := range p.LoadBalancerDescriptions {
			if filter(lb) {
				found = append(found, lb)
			}
		}

		return true
	})

	if err != nil {
		return nil, fmt.Errorf("error listing ELBs: %v", err)
	}

	return found, nil
}

func describeLoadBalancerTags(cloud awsup.AWSCloud, loadBalancerNames []string) (map[string][]*elb.Tag, error) {
	// TODO: Filter by cluster?

	request := &elb.DescribeTagsInput{}
	request.LoadBalancerNames = aws.StringSlice(loadBalancerNames)

	// TODO: Cache?
	klog.V(2).Infof("Querying ELB tags for %s", loadBalancerNames)
	response, err := cloud.ELB().DescribeTags(request)
	if err != nil {
		return nil, err
	}

	tagMap := make(map[string][]*elb.Tag)
	for _, tagset := range response.TagDescriptions {
		tagMap[aws.StringValue(tagset.LoadBalancerName)] = tagset.Tags
	}
	return tagMap, nil
}

func (e *LoadBalancer) Find(c *fi.Context) (*LoadBalancer, error) {
	cloud := c.Cloud.(awsup.AWSCloud)

	lb, err := FindLoadBalancerByNameTag(cloud, fi.StringValue(e.Name))
	if err != nil {
		return nil, err
	}
	if lb == nil {
		return nil, nil
	}

	actual := &LoadBalancer{}
	actual.Name = e.Name
	actual.LoadBalancerName = lb.LoadBalancerName
	actual.DNSName = lb.DNSName
	actual.HostedZoneId = lb.CanonicalHostedZoneNameID
	actual.Scheme = lb.Scheme
	actual.Lifecycle = e.Lifecycle

	tagMap, err := describeLoadBalancerTags(cloud, []string{*lb.LoadBalancerName})
	if err != nil {
		return nil, err
	}
	actual.Tags = make(map[string]string)
	for _, tag := range tagMap[*e.LoadBalancerName] {
		actual.Tags[aws.StringValue(tag.Key)] = aws.StringValue(tag.Value)
	}

	for _, subnet := range lb.Subnets {
		actual.Subnets = append(actual.Subnets, &Subnet{ID: subnet})
	}

	for _, sg := range lb.SecurityGroups {
		actual.SecurityGroups = append(actual.SecurityGroups, &SecurityGroup{ID: sg})
	}

	actual.Listeners = make(map[string]*LoadBalancerListener)

	for _, ld := range lb.ListenerDescriptions {
		l := ld.Listener
		loadBalancerPort := strconv.FormatInt(aws.Int64Value(l.LoadBalancerPort), 10)

		actualListener := &LoadBalancerListener{}
		actualListener.InstancePort = int(aws.Int64Value(l.InstancePort))
		actualListener.SSLCertificateID = aws.StringValue(l.SSLCertificateId)
		actual.Listeners[loadBalancerPort] = actualListener
	}

	healthcheck, err := findHealthCheck(lb)
	if err != nil {
		return nil, err
	}
	actual.HealthCheck = healthcheck

	// Extract attributes
	lbAttributes, err := findELBAttributes(cloud, aws.StringValue(lb.LoadBalancerName))
	if err != nil {
		return nil, err
	}
	klog.V(4).Infof("ELB attributes: %+v", lbAttributes)

	if lbAttributes != nil {
		actual.AccessLog = &LoadBalancerAccessLog{}
		if lbAttributes.AccessLog.EmitInterval != nil {
			actual.AccessLog.EmitInterval = lbAttributes.AccessLog.EmitInterval
		}
		if lbAttributes.AccessLog.Enabled != nil {
			actual.AccessLog.Enabled = lbAttributes.AccessLog.Enabled
		}
		if lbAttributes.AccessLog.S3BucketName != nil {
			actual.AccessLog.S3BucketName = lbAttributes.AccessLog.S3BucketName
		}
		if lbAttributes.AccessLog.S3BucketPrefix != nil {
			actual.AccessLog.S3BucketPrefix = lbAttributes.AccessLog.S3BucketPrefix
		}

		actual.ConnectionDraining = &LoadBalancerConnectionDraining{}
		if lbAttributes.ConnectionDraining.Enabled != nil {
			actual.ConnectionDraining.Enabled = lbAttributes.ConnectionDraining.Enabled
		}
		if lbAttributes.ConnectionDraining.Timeout != nil {
			actual.ConnectionDraining.Timeout = lbAttributes.ConnectionDraining.Timeout
		}

		actual.ConnectionSettings = &LoadBalancerConnectionSettings{}
		if lbAttributes.ConnectionSettings.IdleTimeout != nil {
			actual.ConnectionSettings.IdleTimeout = lbAttributes.ConnectionSettings.IdleTimeout
		}

		actual.CrossZoneLoadBalancing = &LoadBalancerCrossZoneLoadBalancing{}
		if lbAttributes.CrossZoneLoadBalancing.Enabled != nil {
			actual.CrossZoneLoadBalancing.Enabled = lbAttributes.CrossZoneLoadBalancing.Enabled
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
	// 1. We don't want to force a rename of the ELB, because that is a destructive operation
	// 2. We were creating ELBs with insufficiently qualified names previously
	if fi.StringValue(e.LoadBalancerName) != fi.StringValue(actual.LoadBalancerName) {
		klog.V(2).Infof("Reusing existing load balancer with name: %q", aws.StringValue(actual.LoadBalancerName))
		e.LoadBalancerName = actual.LoadBalancerName
	}

	// TODO: Make Normalize a standard method
	actual.Normalize()

	klog.V(4).Infof("Found ELB %+v", actual)

	return actual, nil
}

var _ fi.HasAddress = &LoadBalancer{}

func (e *LoadBalancer) FindIPAddress(context *fi.Context) (*string, error) {
	cloud := context.Cloud.(awsup.AWSCloud)

	lb, err := FindLoadBalancerByNameTag(cloud, fi.StringValue(e.Name))
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

func (e *LoadBalancer) Run(c *fi.Context) error {
	// TODO: Make Normalize a standard method
	e.Normalize()

	return fi.DefaultDeltaRunMethod(e, c)
}

func (e *LoadBalancer) Normalize() {
	// We need to sort our arrays consistently, so we don't get spurious changes
	sort.Stable(OrderSubnetsById(e.Subnets))
	sort.Stable(OrderSecurityGroupsById(e.SecurityGroups))
}

func (s *LoadBalancer) CheckChanges(a, e, changes *LoadBalancer) error {
	if a == nil {
		if fi.StringValue(e.Name) == "" {
			return fi.RequiredField("Name")
		}
		if len(e.SecurityGroups) == 0 {
			return fi.RequiredField("SecurityGroups")
		}
		if len(e.Subnets) == 0 {
			return fi.RequiredField("Subnets")
		}

		if e.AccessLog != nil {
			if e.AccessLog.Enabled == nil {
				return fi.RequiredField("Acceslog.Enabled")
			}
			if *e.AccessLog.Enabled {
				if e.AccessLog.S3BucketName == nil {
					return fi.RequiredField("Acceslog.S3Bucket")
				}
			}
		}
		if e.ConnectionDraining != nil {
			if e.ConnectionDraining.Enabled == nil {
				return fi.RequiredField("ConnectionDraining.Enabled")
			}
		}

		if e.CrossZoneLoadBalancing != nil {
			if e.CrossZoneLoadBalancing.Enabled == nil {
				return fi.RequiredField("CrossZoneLoadBalancing.Enabled")
			}
		}
	}

	return nil
}

func (_ *LoadBalancer) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *LoadBalancer) error {
	var loadBalancerName string
	if a == nil {
		if e.LoadBalancerName == nil {
			return fi.RequiredField("LoadBalancerName")
		}
		loadBalancerName = *e.LoadBalancerName

		request := &elb.CreateLoadBalancerInput{}
		request.LoadBalancerName = e.LoadBalancerName
		request.Scheme = e.Scheme

		for _, subnet := range e.Subnets {
			request.Subnets = append(request.Subnets, subnet.ID)
		}

		for _, sg := range e.SecurityGroups {
			request.SecurityGroups = append(request.SecurityGroups, sg.ID)
		}

		request.Listeners = []*elb.Listener{}

		for loadBalancerPort, listener := range e.Listeners {
			loadBalancerPortInt, err := strconv.ParseInt(loadBalancerPort, 10, 64)
			if err != nil {
				return fmt.Errorf("error parsing load balancer listener port: %q", loadBalancerPort)
			}
			awsListener := listener.mapToAWS(loadBalancerPortInt)
			request.Listeners = append(request.Listeners, awsListener)
		}

		klog.V(2).Infof("Creating ELB with Name:%q", loadBalancerName)

		response, err := t.Cloud.ELB().CreateLoadBalancer(request)
		if err != nil {
			return fmt.Errorf("error creating ELB: %v", err)
		}

		e.DNSName = response.DNSName

		// Requery to get the CanonicalHostedZoneNameID
		lb, err := findLoadBalancerByLoadBalancerName(t.Cloud, loadBalancerName)
		if err != nil {
			return err
		}
		if lb == nil {
			// TODO: Retry?  Is this async
			return fmt.Errorf("Unable to find newly created ELB %q", loadBalancerName)
		}
		e.HostedZoneId = lb.CanonicalHostedZoneNameID
	} else {
		loadBalancerName = fi.StringValue(a.LoadBalancerName)

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
				request := &elb.DetachLoadBalancerFromSubnetsInput{}
				request.SetLoadBalancerName(loadBalancerName)
				request.SetSubnets(aws.StringSlice(oldSubnetIDs))

				klog.V(2).Infof("Detaching Load Balancer from old subnets")
				if _, err := t.Cloud.ELB().DetachLoadBalancerFromSubnets(request); err != nil {
					return fmt.Errorf("Error detaching Load Balancer from old subnets: %v", err)
				}
			}

			newSubnetIDs := slice.GetUniqueStrings(actualSubnets, expectedSubnets)
			if len(newSubnetIDs) > 0 {
				request := &elb.AttachLoadBalancerToSubnetsInput{}
				request.SetLoadBalancerName(loadBalancerName)
				request.SetSubnets(aws.StringSlice(newSubnetIDs))

				klog.V(2).Infof("Attaching Load Balancer to new subnets")
				if _, err := t.Cloud.ELB().AttachLoadBalancerToSubnets(request); err != nil {
					return fmt.Errorf("Error attaching Load Balancer to new subnets: %v", err)
				}
			}
		}

		if changes.SecurityGroups != nil {
			request := &elb.ApplySecurityGroupsToLoadBalancerInput{}
			request.LoadBalancerName = aws.String(loadBalancerName)
			for _, sg := range e.SecurityGroups {
				request.SecurityGroups = append(request.SecurityGroups, sg.ID)
			}

			klog.V(2).Infof("Updating Load Balancer Security Groups")
			if _, err := t.Cloud.ELB().ApplySecurityGroupsToLoadBalancer(request); err != nil {
				return fmt.Errorf("Error updating security groups on Load Balancer: %v", err)
			}
		}

		if changes.Listeners != nil {

			elbDescription, err := findLoadBalancerByLoadBalancerName(t.Cloud, loadBalancerName)
			if err != nil {
				return fmt.Errorf("error getting load balancer by name: %v", err)
			}

			if elbDescription != nil {
				// deleting the listener before recreating it
				t.Cloud.ELB().DeleteLoadBalancerListeners(&elb.DeleteLoadBalancerListenersInput{
					LoadBalancerName:  aws.String(loadBalancerName),
					LoadBalancerPorts: []*int64{aws.Int64(443)},
				})
			}

			request := &elb.CreateLoadBalancerListenersInput{}
			request.LoadBalancerName = aws.String(loadBalancerName)

			for loadBalancerPort, listener := range changes.Listeners {
				loadBalancerPortInt, err := strconv.ParseInt(loadBalancerPort, 10, 64)
				if err != nil {
					return fmt.Errorf("error parsing load balancer listener port: %q", loadBalancerPort)
				}
				awsListener := listener.mapToAWS(loadBalancerPortInt)
				request.Listeners = append(request.Listeners, awsListener)
			}

			klog.V(2).Infof("Creating LoadBalancer listeners")

			_, err = t.Cloud.ELB().CreateLoadBalancerListeners(request)
			if err != nil {
				return fmt.Errorf("error creating LoadBalancerListeners: %v", err)
			}
		}
	}

	if err := t.AddELBTags(loadBalancerName, e.Tags); err != nil {
		return err
	}

	if err := t.RemoveELBTags(loadBalancerName, e.Tags); err != nil {
		return err
	}

	if changes.HealthCheck != nil && e.HealthCheck != nil {
		request := &elb.ConfigureHealthCheckInput{}
		request.LoadBalancerName = aws.String(loadBalancerName)
		request.HealthCheck = &elb.HealthCheck{
			Target:             e.HealthCheck.Target,
			HealthyThreshold:   e.HealthCheck.HealthyThreshold,
			UnhealthyThreshold: e.HealthCheck.UnhealthyThreshold,
			Interval:           e.HealthCheck.Interval,
			Timeout:            e.HealthCheck.Timeout,
		}

		klog.V(2).Infof("Configuring health checks on ELB %q", loadBalancerName)

		_, err := t.Cloud.ELB().ConfigureHealthCheck(request)
		if err != nil {
			return fmt.Errorf("error configuring health checks on ELB: %v", err)
		}
	}

	if err := e.modifyLoadBalancerAttributes(t, a, e, changes); err != nil {
		klog.Infof("error modifying ELB attributes: %v", err)
		return err
	}

	return nil
}

type terraformLoadBalancer struct {
	LoadBalancerName *string                          `json:"name" cty:"name"`
	Listener         []*terraformLoadBalancerListener `json:"listener" cty:"listener"`
	SecurityGroups   []*terraform.Literal             `json:"security_groups" cty:"security_groups"`
	Subnets          []*terraform.Literal             `json:"subnets" cty:"subnets"`
	Internal         *bool                            `json:"internal,omitempty" cty:"internal"`

	HealthCheck *terraformLoadBalancerHealthCheck `json:"health_check,omitempty" cty:"health_check"`
	AccessLog   *terraformLoadBalancerAccessLog   `json:"access_logs,omitempty" cty:"access_logs"`

	ConnectionDraining        *bool  `json:"connection_draining,omitempty" cty:"connection_draining"`
	ConnectionDrainingTimeout *int64 `json:"connection_draining_timeout,omitempty" cty:"connection_draining_timeout"`

	CrossZoneLoadBalancing *bool `json:"cross_zone_load_balancing,omitempty" cty:"cross_zone_load_balancing"`

	IdleTimeout *int64 `json:"idle_timeout,omitempty" cty:"idle_timeout"`

	Tags map[string]string `json:"tags,omitempty" cty:"tags"`
}

type terraformLoadBalancerListener struct {
	InstancePort     int    `json:"instance_port" cty:"instance_port"`
	InstanceProtocol string `json:"instance_protocol" cty:"instance_protocol"`
	LBPort           int64  `json:"lb_port" cty:"lb_port"`
	LBProtocol       string `json:"lb_protocol" cty:"lb_protocol"`
	SSLCertificateID string `json:"ssl_certificate_id,omitempty" cty:"ssl_certificate_id"`
}

type terraformLoadBalancerHealthCheck struct {
	Target             *string `json:"target" cty:"target"`
	HealthyThreshold   *int64  `json:"healthy_threshold" cty:"healthy_threshold"`
	UnhealthyThreshold *int64  `json:"unhealthy_threshold" cty:"unhealthy_threshold"`
	Interval           *int64  `json:"interval" cty:"interval"`
	Timeout            *int64  `json:"timeout" cty:"timeout"`
}

func (_ *LoadBalancer) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *LoadBalancer) error {
	cloud := t.Cloud.(awsup.AWSCloud)

	if e.LoadBalancerName == nil {
		return fi.RequiredField("LoadBalancerName")
	}

	tf := &terraformLoadBalancer{
		LoadBalancerName: e.LoadBalancerName,
	}
	if fi.StringValue(e.Scheme) == "internal" {
		tf.Internal = fi.Bool(true)
	}

	for _, subnet := range e.Subnets {
		tf.Subnets = append(tf.Subnets, subnet.TerraformLink())
	}
	terraform.SortLiterals(tf.Subnets)

	for _, sg := range e.SecurityGroups {
		tf.SecurityGroups = append(tf.SecurityGroups, sg.TerraformLink())
	}
	terraform.SortLiterals(tf.SecurityGroups)

	for loadBalancerPort, listener := range e.Listeners {
		loadBalancerPortInt, err := strconv.ParseInt(loadBalancerPort, 10, 64)
		if err != nil {
			return fmt.Errorf("error parsing load balancer listener port: %q", loadBalancerPort)
		}

		if listener.SSLCertificateID != "" {
			tf.Listener = append(tf.Listener, &terraformLoadBalancerListener{
				InstanceProtocol: "SSL",
				InstancePort:     listener.InstancePort,
				LBPort:           loadBalancerPortInt,
				LBProtocol:       "SSL",
				SSLCertificateID: listener.SSLCertificateID,
			})
		} else {
			tf.Listener = append(tf.Listener, &terraformLoadBalancerListener{
				InstanceProtocol: "TCP",
				InstancePort:     listener.InstancePort,
				LBPort:           loadBalancerPortInt,
				LBProtocol:       "TCP",
			})
		}

	}

	if e.HealthCheck != nil {
		tf.HealthCheck = &terraformLoadBalancerHealthCheck{
			Target:             e.HealthCheck.Target,
			HealthyThreshold:   e.HealthCheck.HealthyThreshold,
			UnhealthyThreshold: e.HealthCheck.UnhealthyThreshold,
			Interval:           e.HealthCheck.Interval,
			Timeout:            e.HealthCheck.Timeout,
		}
	}

	if e.AccessLog != nil {
		tf.AccessLog = &terraformLoadBalancerAccessLog{
			EmitInterval:   e.AccessLog.EmitInterval,
			Enabled:        e.AccessLog.Enabled,
			S3BucketName:   e.AccessLog.S3BucketName,
			S3BucketPrefix: e.AccessLog.S3BucketPrefix,
		}
	}

	if e.ConnectionDraining != nil {
		tf.ConnectionDraining = e.ConnectionDraining.Enabled
		tf.ConnectionDrainingTimeout = e.ConnectionDraining.Timeout
	}

	if e.ConnectionSettings != nil {
		tf.IdleTimeout = e.ConnectionSettings.IdleTimeout
	}

	if e.CrossZoneLoadBalancing != nil {
		tf.CrossZoneLoadBalancing = e.CrossZoneLoadBalancing.Enabled
	}

	var tags map[string]string = cloud.BuildTags(e.Name)
	for k, v := range e.Tags {
		tags[k] = v
	}
	tf.Tags = tags

	return t.RenderResource("aws_elb", *e.Name, tf)
}

func (e *LoadBalancer) TerraformLink(params ...string) *terraform.Literal {
	prop := "id"
	if len(params) > 0 {
		prop = params[0]
	}
	return terraform.LiteralProperty("aws_elb", *e.Name, prop)
}

type cloudformationLoadBalancer struct {
	LoadBalancerName *string                               `json:"LoadBalancerName,omitempty"`
	Listener         []*cloudformationLoadBalancerListener `json:"Listeners,omitempty"`
	SecurityGroups   []*cloudformation.Literal             `json:"SecurityGroups,omitempty"`
	Subnets          []*cloudformation.Literal             `json:"Subnets,omitempty"`
	Scheme           *string                               `json:"Scheme,omitempty"`

	HealthCheck *cloudformationLoadBalancerHealthCheck `json:"HealthCheck,omitempty"`
	AccessLog   *cloudformationLoadBalancerAccessLog   `json:"AccessLoggingPolicy,omitempty"`

	ConnectionDrainingPolicy *cloudformationConnectionDrainingPolicy `json:"ConnectionDrainingPolicy,omitempty"`
	ConnectionSettings       *cloudformationConnectionSettings       `json:"ConnectionSettings,omitempty"`

	CrossZoneLoadBalancing *bool `json:"CrossZone,omitempty"`

	Tags []cloudformationTag `json:"Tags,omitempty"`
}

type cloudformationLoadBalancerListener struct {
	InstancePort         int    `json:"InstancePort"`
	InstanceProtocol     string `json:"InstanceProtocol"`
	LoadBalancerPort     int64  `json:"LoadBalancerPort"`
	LoadBalancerProtocol string `json:"Protocol"`
}

type cloudformationLoadBalancerHealthCheck struct {
	Target             *string `json:"Target"`
	HealthyThreshold   *int64  `json:"HealthyThreshold"`
	UnhealthyThreshold *int64  `json:"UnhealthyThreshold"`
	Interval           *int64  `json:"Interval"`
	Timeout            *int64  `json:"Timeout"`
}

type cloudformationConnectionDrainingPolicy struct {
	Enabled *bool  `json:"Enabled,omitempty"`
	Timeout *int64 `json:"Timeout,omitempty"`
}

type cloudformationConnectionSettings struct {
	IdleTimeout *int64 `json:"IdleTimeout,omitempty"`
}

func (_ *LoadBalancer) RenderCloudformation(t *cloudformation.CloudformationTarget, a, e, changes *LoadBalancer) error {
	// TODO: From http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ec2-elb.html:
	// If this resource has a public IP address and is also in a VPC that is defined in the same template,
	// you must use the DependsOn attribute to declare a dependency on the VPC-gateway attachment.

	cloud := t.Cloud.(awsup.AWSCloud)

	if e.LoadBalancerName == nil {
		return fi.RequiredField("LoadBalancerName")
	}

	tf := &cloudformationLoadBalancer{
		LoadBalancerName: e.LoadBalancerName,
		Scheme:           e.Scheme,
	}

	for _, subnet := range e.Subnets {
		tf.Subnets = append(tf.Subnets, subnet.CloudformationLink())
	}

	for _, sg := range e.SecurityGroups {
		tf.SecurityGroups = append(tf.SecurityGroups, sg.CloudformationLink())
	}

	for loadBalancerPort, listener := range e.Listeners {
		loadBalancerPortInt, err := strconv.ParseInt(loadBalancerPort, 10, 64)
		if err != nil {
			return fmt.Errorf("error parsing load balancer listener port: %q", loadBalancerPort)
		}

		tf.Listener = append(tf.Listener, &cloudformationLoadBalancerListener{
			InstanceProtocol:     "TCP",
			InstancePort:         listener.InstancePort,
			LoadBalancerPort:     loadBalancerPortInt,
			LoadBalancerProtocol: "TCP",
		})
	}

	if e.HealthCheck != nil {
		tf.HealthCheck = &cloudformationLoadBalancerHealthCheck{
			Target:             e.HealthCheck.Target,
			HealthyThreshold:   e.HealthCheck.HealthyThreshold,
			UnhealthyThreshold: e.HealthCheck.UnhealthyThreshold,
			Interval:           e.HealthCheck.Interval,
			Timeout:            e.HealthCheck.Timeout,
		}
	}

	if e.AccessLog != nil {
		tf.AccessLog = &cloudformationLoadBalancerAccessLog{
			EmitInterval:   e.AccessLog.EmitInterval,
			Enabled:        e.AccessLog.Enabled,
			S3BucketName:   e.AccessLog.S3BucketName,
			S3BucketPrefix: e.AccessLog.S3BucketPrefix,
		}
	}

	if e.ConnectionDraining != nil {
		tf.ConnectionDrainingPolicy = &cloudformationConnectionDrainingPolicy{
			Enabled: e.ConnectionDraining.Enabled,
			Timeout: e.ConnectionDraining.Timeout,
		}
	}

	if e.ConnectionSettings != nil {
		tf.ConnectionSettings = &cloudformationConnectionSettings{
			IdleTimeout: e.ConnectionSettings.IdleTimeout,
		}
	}

	if e.CrossZoneLoadBalancing != nil {
		tf.CrossZoneLoadBalancing = e.CrossZoneLoadBalancing.Enabled
	}

	var tags map[string]string = cloud.BuildTags(e.Name)
	for k, v := range e.Tags {
		tags[k] = v
	}

	tf.Tags = buildCloudformationTags(tags)

	return t.RenderResource("AWS::ElasticLoadBalancing::LoadBalancer", *e.Name, tf)
}

func (e *LoadBalancer) CloudformationLink() *cloudformation.Literal {
	return cloudformation.Ref("AWS::ElasticLoadBalancing::LoadBalancer", *e.Name)
}

func (e *LoadBalancer) CloudformationAttrCanonicalHostedZoneNameID() *cloudformation.Literal {
	return cloudformation.GetAtt("AWS::ElasticLoadBalancing::LoadBalancer", *e.Name, "CanonicalHostedZoneNameID")
}

func (e *LoadBalancer) CloudformationAttrDNSName() *cloudformation.Literal {
	return cloudformation.GetAtt("AWS::ElasticLoadBalancing::LoadBalancer", *e.Name, "DNSName")
}
