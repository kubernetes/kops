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

	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

type InstanceElasticIPAttachment struct {
	Instance  *Instance
	ElasticIP *ElasticIP
}

func (e *InstanceElasticIPAttachment) String() string {
	return fi.TaskAsString(e)
}

func (e *InstanceElasticIPAttachment) Find(c *fi.Context) (*InstanceElasticIPAttachment, error) {
	cloud := c.Cloud.(awsup.AWSCloud)

	instanceID := e.Instance.ID
	eipID := e.ElasticIP.ID

	if instanceID == nil || eipID == nil {
		return nil, nil
	}

	request := &ec2.DescribeAddressesInput{
		AllocationIds: []*string{eipID},
	}

	response, err := cloud.EC2().DescribeAddresses(request)
	if err != nil {
		return nil, fmt.Errorf("error listing ElasticIPs: %v", err)
	}
	if response == nil || len(response.Addresses) == 0 {
		return nil, nil
	}

	if len(response.Addresses) != 1 {
		klog.Fatalf("found multiple ElasticIPs for public IP")
	}

	a := response.Addresses[0]
	actual := &InstanceElasticIPAttachment{}
	if a.InstanceId != nil {
		actual.Instance = &Instance{ID: a.InstanceId}
	}
	actual.ElasticIP = &ElasticIP{ID: a.AllocationId}
	return actual, nil
}

func (e *InstanceElasticIPAttachment) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *InstanceElasticIPAttachment) CheckChanges(a, e, changes *InstanceElasticIPAttachment) error {
	return nil
}

func (_ *InstanceElasticIPAttachment) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *InstanceElasticIPAttachment) error {
	if changes.Instance != nil {
		err := t.WaitForInstanceRunning(*e.Instance.ID)
		if err != nil {
			return err
		}

		request := &ec2.AssociateAddressInput{}
		request.InstanceId = e.Instance.ID
		request.AllocationId = a.ElasticIP.ID

		_, err = t.Cloud.EC2().AssociateAddress(request)
		if err != nil {
			return fmt.Errorf("error creating InstanceElasticIPAttachment: %v", err)
		}
	}

	return nil // no tags
}
