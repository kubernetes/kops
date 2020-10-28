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
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

func findNetworkLoadBalancerAttributes(cloud awsup.AWSCloud, LoadBalancerArn string) ([]*elbv2.LoadBalancerAttribute, error) {

	request := &elbv2.DescribeLoadBalancerAttributesInput{
		LoadBalancerArn: aws.String(LoadBalancerArn),
	}

	response, err := cloud.ELBV2().DescribeLoadBalancerAttributes(request)
	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, nil
	}

	//we get back an array of attributes

	/*
	   Key *string `type:"string"`
	   Value *string `type:"string"`
	*/

	return response.Attributes, nil
}

func (_ *NetworkLoadBalancer) modifyLoadBalancerAttributes(t *awsup.AWSAPITarget, a, e, changes *NetworkLoadBalancer, loadBalancerArn string) error {
	if changes.CrossZoneLoadBalancing == nil {
		klog.V(4).Infof("No LoadBalancerAttribute changes; skipping update")
		return nil
	}

	loadBalancerName := fi.StringValue(e.LoadBalancerName)

	request := &elbv2.ModifyLoadBalancerAttributesInput{
		LoadBalancerArn: aws.String(loadBalancerArn),
	}

	var attributes []*elbv2.LoadBalancerAttribute

	attribute := &elbv2.LoadBalancerAttribute{}
	attribute.Key = aws.String("load_balancing.cross_zone.enabled")
	if e.CrossZoneLoadBalancing == nil {
		attribute.Value = aws.String("false")
	} else {
		attribute.Value = aws.String(strconv.FormatBool(aws.BoolValue(e.CrossZoneLoadBalancing)))
	}
	attributes = append(attributes, attribute)

	request.Attributes = attributes

	klog.V(2).Infof("Configuring NLB attributes for NLB %q", loadBalancerName)

	response, err := t.Cloud.ELBV2().ModifyLoadBalancerAttributes(request)
	if err != nil {
		return fmt.Errorf("error configuring NLB attributes for NLB %q: %v", loadBalancerName, err)
	}

	klog.V(4).Infof("modified NLB attributes for NLB %q, response %+v", loadBalancerName, response)

	return nil
}
