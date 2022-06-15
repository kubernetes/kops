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

package protokube

import (
	"fmt"
	"net"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/klog/v2"
	"k8s.io/kops/protokube/pkg/gossip"
	gossipaws "k8s.io/kops/protokube/pkg/gossip/aws"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

// AWSCloudProvider defines the AWS cloud provider implementation
type AWSCloudProvider struct {
	mutex sync.Mutex

	clusterTag string
	deviceMap  map[string]string
	ec2        *ec2.EC2
	instanceId string
	internalIP net.IP
	metadata   *ec2metadata.EC2Metadata
	zone       string
}

var _ CloudProvider = &AWSCloudProvider{}

// NewAWSCloudProvider returns a new aws volume provider
func NewAWSCloudProvider() (*AWSCloudProvider, error) {
	a := &AWSCloudProvider{
		deviceMap: make(map[string]string),
	}

	config := aws.NewConfig()
	config = config.WithCredentialsChainVerboseErrors(true)

	s, err := session.NewSession(config)
	if err != nil {
		return nil, fmt.Errorf("error starting new AWS session: %v", err)
	}
	s.Handlers.Send.PushFront(func(r *request.Request) {
		// Log requests
		klog.V(4).Infof("AWS API Request: %s/%s", r.ClientInfo.ServiceName, r.Operation.Name)
	})

	a.metadata = ec2metadata.New(s, config)

	region, err := a.metadata.Region()
	if err != nil {
		return nil, fmt.Errorf("error querying ec2 metadata service (for az/region): %v", err)
	}

	a.zone, err = a.metadata.GetMetadata("placement/availability-zone")
	if err != nil {
		return nil, fmt.Errorf("error querying ec2 metadata service (for az): %v", err)
	}

	a.instanceId, err = a.metadata.GetMetadata("instance-id")
	if err != nil {
		return nil, fmt.Errorf("error querying ec2 metadata service (for instance-id): %v", err)
	}

	a.ec2 = ec2.New(s, config.WithRegion(region))

	err = a.discoverTags()
	if err != nil {
		return nil, err
	}

	return a, nil
}

func (a *AWSCloudProvider) InstanceInternalIP() net.IP {
	return a.internalIP
}

func (a *AWSCloudProvider) discoverTags() error {
	instance, err := a.describeInstance()
	if err != nil {
		return err
	}

	tagMap := make(map[string]string)
	for _, tag := range instance.Tags {
		tagMap[aws.StringValue(tag.Key)] = aws.StringValue(tag.Value)
	}

	clusterID := tagMap[awsup.TagClusterName]
	if clusterID == "" {
		return fmt.Errorf("Cluster tag %q not found on this instance (%q)", awsup.TagClusterName, a.instanceId)
	}

	a.clusterTag = clusterID

	a.internalIP = net.ParseIP(aws.StringValue(instance.Ipv6Address))
	if a.internalIP == nil {
		a.internalIP = net.ParseIP(aws.StringValue(instance.PrivateIpAddress))
	}
	if a.internalIP == nil {
		return fmt.Errorf("Internal IP not found on this instance (%q)", a.instanceId)
	}

	return nil
}

func (a *AWSCloudProvider) describeInstance() (*ec2.Instance, error) {
	request := &ec2.DescribeInstancesInput{}
	request.InstanceIds = []*string{&a.instanceId}

	var instances []*ec2.Instance
	err := a.ec2.DescribeInstancesPages(request, func(p *ec2.DescribeInstancesOutput, lastPage bool) (shouldContinue bool) {
		for _, r := range p.Reservations {
			instances = append(instances, r.Instances...)
		}
		return true
	})
	if err != nil {
		return nil, fmt.Errorf("error querying for EC2 instance %q: %v", a.instanceId, err)
	}

	if len(instances) != 1 {
		return nil, fmt.Errorf("unexpected number of instances found with id %q: %d", a.instanceId, len(instances))
	}

	return instances[0], nil
}

func (a *AWSCloudProvider) GossipSeeds() (gossip.SeedProvider, error) {
	tags := make(map[string]string)
	tags[awsup.TagClusterName] = a.clusterTag

	return gossipaws.NewSeedProvider(a.ec2, tags)
}

func (a *AWSCloudProvider) InstanceID() string {
	return a.instanceId
}
