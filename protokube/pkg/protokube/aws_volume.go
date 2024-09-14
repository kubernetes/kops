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
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"k8s.io/kops/protokube/pkg/gossip"
	gossipaws "k8s.io/kops/protokube/pkg/gossip/aws"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/util/pkg/awslog"
)

// AWSCloudProvider defines the AWS cloud provider implementation
type AWSCloudProvider struct {
	mutex sync.Mutex

	clusterTag string
	deviceMap  map[string]string
	ec2        ec2.DescribeInstancesAPIClient
	instanceId string
	imdsClient *imds.Client
	zone       string
}

var _ CloudProvider = &AWSCloudProvider{}

// NewAWSCloudProvider returns a new aws volume provider
func NewAWSCloudProvider() (*AWSCloudProvider, error) {
	ctx := context.TODO()
	a := &AWSCloudProvider{
		deviceMap: make(map[string]string),
	}

	config, err := awsconfig.LoadDefaultConfig(ctx, awslog.WithAWSLogger())
	if err != nil {
		return nil, fmt.Errorf("error loading AWS config: %w", err)
	}
	a.imdsClient = imds.NewFromConfig(config)

	regionResp, err := a.imdsClient.GetRegion(ctx, &imds.GetRegionInput{})
	if err != nil {
		return nil, fmt.Errorf("error querying ec2 metadata service (for az/region): %w", err)
	}

	zoneResp, err := a.imdsClient.GetMetadata(ctx, &imds.GetMetadataInput{Path: "placement/availability-zone"})
	if err != nil {
		return nil, fmt.Errorf("error querying ec2 metadata service (for az): %w", err)
	}
	zone, err := io.ReadAll(zoneResp.Content)
	if err != nil {
		return nil, fmt.Errorf("error reading ec2 metadata service response (for az): %w", err)
	}
	a.zone = string(zone)

	instanceIdResp, err := a.imdsClient.GetMetadata(ctx, &imds.GetMetadataInput{Path: "instance-id"})
	if err != nil {
		return nil, fmt.Errorf("error querying ec2 metadata service (for instance-id): %w", err)
	}
	instanceId, err := io.ReadAll(instanceIdResp.Content)
	if err != nil {
		return nil, fmt.Errorf("error reading ec2 metadata service response (for az): %w", err)
	}
	a.instanceId = string(instanceId)

	config.Region = regionResp.Region
	a.ec2 = ec2.NewFromConfig(config)

	err = a.discoverTags(ctx)
	if err != nil {
		return nil, err
	}

	return a, nil
}

func (a *AWSCloudProvider) discoverTags(ctx context.Context) error {
	instance, err := a.describeInstance(ctx)
	if err != nil {
		return err
	}

	tagMap := make(map[string]string)
	for _, tag := range instance.Tags {
		tagMap[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
	}

	clusterID := tagMap[awsup.TagClusterName]
	if clusterID == "" {
		return fmt.Errorf("Cluster tag %q not found on this instance (%q)", awsup.TagClusterName, a.instanceId)
	}

	a.clusterTag = clusterID

	return nil
}

func (a *AWSCloudProvider) describeInstance(ctx context.Context) (*ec2types.Instance, error) {
	request := &ec2.DescribeInstancesInput{}
	request.InstanceIds = []string{a.instanceId}

	var instances []ec2types.Instance
	paginator := ec2.NewDescribeInstancesPaginator(a.ec2, request)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("error querying for EC2 instance %q: %v", a.instanceId, err)
		}
		for _, r := range page.Reservations {
			instances = append(instances, r.Instances...)
		}
	}

	if len(instances) != 1 {
		return nil, fmt.Errorf("unexpected number of instances found with id %q: %d", a.instanceId, len(instances))
	}

	return &instances[0], nil
}

func (a *AWSCloudProvider) GossipSeeds() (gossip.SeedProvider, error) {
	tags := make(map[string]string)
	tags[awsup.TagClusterName] = a.clusterTag

	return gossipaws.NewSeedProvider(a.ec2, tags)
}

func (a *AWSCloudProvider) InstanceID() string {
	return a.instanceId
}
