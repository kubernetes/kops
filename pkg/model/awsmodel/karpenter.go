/*
Copyright 2026 The Kubernetes Authors.

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

package awsmodel

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/iam"
	"k8s.io/kops/pkg/util/stringorset"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

type karpenterEvent struct {
	name    string
	pattern string
}

var (
	_ fi.CloudupModelBuilder = &KarpenterBuilder{}

	karpenterEvents = []karpenterEvent{
		{
			name:    "SpotInterruption",
			pattern: `{"source": ["aws.ec2"],"detail-type": ["EC2 Spot Instance Interruption Warning"]}`,
		},
		{
			name:    "InstanceStateChange",
			pattern: `{"source": ["aws.ec2"],"detail-type": ["EC2 Instance State-change Notification"]}`,
		},
		{
			name:    "InstanceScheduledChange",
			pattern: `{"source": ["aws.health"],"detail-type": ["AWS Health Event"],"detail": {"service": ["EC2"],"eventTypeCategory": ["scheduledChange"]}}`,
		},
		{
			name:    "RebalanceRecommendation",
			pattern: `{"source": ["aws.ec2"],"detail-type": ["EC2 Instance Rebalance Recommendation"]}`,
		},
	}
)

type KarpenterBuilder struct {
	*AWSModelContext

	Lifecycle fi.Lifecycle
}

func (b *KarpenterBuilder) Build(c *fi.CloudupModelBuilderContext) error {
	if b.Cluster.Spec.Karpenter == nil || !b.Cluster.Spec.Karpenter.Enabled {
		return nil
	}

	queueName := model.QueueNamePrefix(b.ClusterName()) + "-karpenter"

	policy := iam.NewPolicy(b.ClusterName(), b.AWSPartition, b.Region)
	arn := arn.ARN{
		Partition: b.AWSPartition,
		Service:   "sqs",
		Region:    b.Region,
		AccountID: b.AWSAccountID,
		Resource:  queueName,
	}

	policy.Statement = append(policy.Statement, &iam.Statement{
		Effect: iam.StatementEffectAllow,
		Principal: iam.Principal{
			Service: new(stringorset.Of("events.amazonaws.com", "sqs.amazonaws.com")),
		},
		Action:   stringorset.Of("sqs:SendMessage"),
		Resource: stringorset.String(arn.String()),
	})
	policyJSON, err := policy.AsJSON()
	if err != nil {
		return fmt.Errorf("rendering policy as json: %w", err)
	}

	queue := &awstasks.SQS{
		Name:                   aws.String(queueName),
		Lifecycle:              b.Lifecycle,
		Policy:                 fi.NewStringResource(policyJSON),
		MessageRetentionPeriod: DefaultMessageRetentionPeriod,
		Tags:                   b.CloudTags(queueName, false),
	}

	c.AddTask(queue)

	clusterName := b.ClusterName()
	clusterNamePrefix := awsup.GetClusterName40(clusterName)

	for _, event := range karpenterEvents {
		// build rule
		ruleName := aws.String(clusterNamePrefix + "-Karpenter-" + event.name)
		pattern := event.pattern

		ruleTask := &awstasks.EventBridgeRule{
			Name:      ruleName,
			Lifecycle: b.Lifecycle,
			Tags:      b.CloudTags(*ruleName, false),

			EventPattern: &pattern,
			SQSQueue:     queue,
		}

		c.AddTask(ruleTask)

		// build target
		targetTask := &awstasks.EventBridgeTarget{
			Name:      aws.String(*ruleName + "-Target"),
			Lifecycle: b.Lifecycle,

			Rule:     ruleTask,
			SQSQueue: queue,
		}

		c.AddTask(targetTask)
	}

	return nil
}
