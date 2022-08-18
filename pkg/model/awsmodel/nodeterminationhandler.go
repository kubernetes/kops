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

package awsmodel

import (
	"strings"

	"k8s.io/kops/pkg/model"

	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"

	"github.com/aws/aws-sdk-go/aws"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
)

const (
	NTHTemplate = `{
		"Version": "2012-10-17",
		"Statement": [{                     
			"Effect": "Allow",
			"Principal": {
				"Service": ["events.amazonaws.com", "sqs.amazonaws.com"]
			},
			"Action": "sqs:SendMessage",
			"Resource": "arn:{{ AWS_PARTITION }}:sqs:{{ AWS_REGION }}:{{ ACCOUNT_ID }}:{{ SQS_QUEUE_NAME }}"
		}]
	}`
	DefaultMessageRetentionPeriod = 300
)

type event struct {
	name    string
	pattern string
}

var (
	_ fi.ModelBuilder = &NodeTerminationHandlerBuilder{}

	events = []event{
		{
			name:    "ASGLifecycle",
			pattern: `{"source":["aws.autoscaling"],"detail-type":["EC2 Instance-terminate Lifecycle Action"]}`,
		},
		{
			name:    "SpotInterruption",
			pattern: `{"source": ["aws.ec2"],"detail-type": ["EC2 Spot Instance Interruption Warning"]}`,
		},
		{
			name:    "RebalanceRecommendation",
			pattern: `{"source": ["aws.ec2"],"detail-type": ["EC2 Instance Rebalance Recommendation"]}`,
		},
		{
			name:    "InstanceStateChange",
			pattern: `{"source": ["aws.ec2"],"detail-type": ["EC2 Instance State-change Notification"]}`,
		},
		{
			name:    "InstanceScheduledChange",
			pattern: `{"source": ["aws.health"],"detail-type": ["AWS Health Event"],"detail": {"service": ["EC2"],"eventTypeCategory": ["scheduledChange"]}}`,
		},
	}
)

type NodeTerminationHandlerBuilder struct {
	*AWSModelContext

	Lifecycle fi.Lifecycle
}

func (b *NodeTerminationHandlerBuilder) Build(c *fi.ModelBuilderContext) error {
	for _, ig := range b.InstanceGroups {
		if ig.Spec.Manager == kops.InstanceManagerCloudGroup {
			err := b.configureASG(c, ig)
			if err != nil {
				return err
			}
		}
	}

	err := b.build(c)
	if err != nil {
		return err
	}

	return nil
}

func (b *NodeTerminationHandlerBuilder) configureASG(c *fi.ModelBuilderContext, ig *kops.InstanceGroup) error {
	name := ig.Name + "-NTHLifecycleHook"

	lifecyleTask := &awstasks.AutoscalingLifecycleHook{
		ID:                  aws.String(name),
		Name:                aws.String(name),
		Lifecycle:           b.Lifecycle,
		AutoscalingGroup:    b.LinkToAutoscalingGroup(ig),
		DefaultResult:       aws.String("CONTINUE"),
		HeartbeatTimeout:    aws.Int64(DefaultMessageRetentionPeriod),
		LifecycleTransition: aws.String("autoscaling:EC2_INSTANCE_TERMINATING"),
		Enabled:             aws.Bool(true),
	}

	c.AddTask(lifecyleTask)

	return nil
}

func (b *NodeTerminationHandlerBuilder) build(c *fi.ModelBuilderContext) error {
	queueName := model.QueueNamePrefix(b.ClusterName()) + "-nth"
	policy := strings.ReplaceAll(NTHTemplate, "{{ AWS_REGION }}", b.Region)
	policy = strings.ReplaceAll(policy, "{{ AWS_PARTITION }}", b.AWSPartition)
	policy = strings.ReplaceAll(policy, "{{ ACCOUNT_ID }}", b.AWSAccountID)
	policy = strings.ReplaceAll(policy, "{{ SQS_QUEUE_NAME }}", queueName)

	queue := &awstasks.SQS{
		Name:                   aws.String(queueName),
		Lifecycle:              b.Lifecycle,
		Policy:                 fi.NewStringResource(policy),
		MessageRetentionPeriod: DefaultMessageRetentionPeriod,
		Tags:                   b.CloudTags(queueName, false),
	}

	c.AddTask(queue)

	clusterName := b.ClusterName()

	clusterNamePrefix := awsup.GetClusterName40(clusterName)
	for _, event := range events {
		// build rule
		ruleName := aws.String(clusterNamePrefix + "-" + event.name)
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
