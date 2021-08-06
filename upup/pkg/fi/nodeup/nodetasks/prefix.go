/*
Copyright 2021 The Kubernetes Authors.

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

package nodetasks

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/nodeup/local"
)

type Prefix struct {
	Name string
}

var _ fi.HasName = &Prefix{}

func (f *Prefix) GetName() *string {
	return &f.Name
}

// String returns a string representation, implementing the Stringer interface
func (p *Prefix) String() string {
	return fmt.Sprintf("Prefix: %s", p.Name)
}

func (e *Prefix) Find(c *fi.Context) (*Prefix, error) {
	return nil, nil
}
func (e *Prefix) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *Prefix) CheckChanges(a, e, changes *Prefix) error {
	return nil
}

func (_ *Prefix) RenderLocal(t *local.LocalTarget, a, e, changes *Prefix) error {

	awsCloud := t.Cloud.(awsup.AWSCloud)

	netifs, err := awsCloud.EC2().DescribeNetworkInterfaces(&ec2.DescribeNetworkInterfacesInput{
		Filters: []*ec2.Filter{
			{
				Name: fi.String("attachment.instance-id"),
				Values: []*string{
					&t.InstanceID,
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to get interface: %w", err)
	}

	netif := netifs.NetworkInterfaces[0]

	_, err = awsCloud.EC2().AssignIpv6Addresses(&ec2.AssignIpv6AddressesInput{
		Ipv6PrefixCount:    fi.Int64(1),
		NetworkInterfaceId: netif.NetworkInterfaceId,
	})
	if err != nil {
		return fmt.Errorf("failed to assign ip address: %w", err)
	}

	return nil
}
