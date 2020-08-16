/*
Copyright 2020 The Kubernetes Authors.

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

package networking

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"

	"k8s.io/klog/v2"
	"k8s.io/kops/nodeup/pkg/model"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
)

type LyftVPCBuilder struct {
	*model.NodeupModelContext
}

var _ fi.ModelBuilder = &LyftVPCBuilder{}

// Build is responsible for configuring the network cni
func (b *LyftVPCBuilder) Build(c *fi.ModelBuilderContext) error {
	networking := b.Cluster.Spec.Networking

	if networking.LyftVPC == nil {
		return nil
	}

	assets := []string{
		"cni-ipvlan-vpc-k8s-ipam",
		"cni-ipvlan-vpc-k8s-ipvlan",
		"cni-ipvlan-vpc-k8s-tool",
		"cni-ipvlan-vpc-k8s-unnumbered-ptp",
	}

	if err := b.AddCNIBinAssets(c, assets); err != nil {
		return err
	}

	securityGroups, err := evaluateSecurityGroups(b.Cluster.Spec.NetworkID)
	if err != nil {
		return err
	}

	conflist := map[string]interface{}{
		"cniVersion": "0.3.1",
		"name":       "cni-ipvlan-vpc-k8s",
		"plugins": []map[string]interface{}{
			{
				"cniVersion":       "0.3.1",
				"type":             "cni-ipvlan-vpc-k8s-ipam",
				"interfaceIndex":   1,
				"skipDeallocation": true,
				"subnetTags":       getSubnetTags(b.Cluster),
				"secGroupIds":      securityGroups,
			},
			{
				"cniVersion": "0.3.1",
				"type":       "cni-ipvlan-vpc-k8s-ipvlan",
				"mode":       "l2",
			},
			{
				"cniVersion":         "0.3.1",
				"type":               "cni-ipvlan-vpc-k8s-unnumbered-ptp",
				"hostInterface":      "eth0",
				"containerInterface": "veth0",
				"ipMasq":             true,
			},
		},
	}

	bytes, err := json.Marshal(conflist)
	if err != nil {
		return err
	}

	c.AddTask(&nodetasks.File{
		Contents: fi.NewBytesResource(bytes),
		Path:     "/etc/cni/net.d/10-cni-ipvlan-vpc-k8s.conflist",
		Type:     nodetasks.FileType_File,
	})

	return nil
}

func getSubnetTags(cluster *api.Cluster) interface{} {
	var tags map[string]string
	if cluster.IsKubernetesGTE("1.18") {
		tags = map[string]string{
			"KubernetesCluster": cluster.Name,
		}
	} else {
		tags = map[string]string{
			"Type": "pod",
		}
	}
	if len(cluster.Spec.Networking.LyftVPC.SubnetTags) > 0 {
		tags = cluster.Spec.Networking.LyftVPC.SubnetTags
	}
	return tags
}

func evaluateSecurityGroups(vpcId string) ([]string, error) {
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

	metadata := ec2metadata.New(s, config)

	region, err := metadata.Region()
	if err != nil {
		return nil, fmt.Errorf("error querying ec2 metadata service (for az/region): %v", err)
	}

	sgNames, err := metadata.GetMetadata("security-groups")
	if err != nil {
		return nil, fmt.Errorf("error querying ec2 metadata service (for security-groups): %v", err)
	}
	svc := ec2.New(s, config.WithRegion(region))

	result, err := svc.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("group-name"),
				Values: aws.StringSlice(strings.Fields(sgNames)),
			},
			{
				Name:   aws.String("vpc-id"),
				Values: []*string{aws.String(vpcId)},
			},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("error looking up instance security group ids: %v", err)
	}
	var sgIds []string
	for _, group := range result.SecurityGroups {
		sgIds = append(sgIds, *group.GroupId)
	}

	return sgIds, nil

}
