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
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/iam"
	"k8s.io/kops/pkg/util/stringorset"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

const (
	DefaultMessageRetentionPeriod = 300
)

type event struct {
	name    string
	pattern string
}

var (
	_ fi.CloudupModelBuilder = &NodeTerminationHandlerBuilder{}
	_ fi.HasDeletions        = &NodeTerminationHandlerBuilder{}

	fixedEvents = []event{
		{
			name:    "ASGLifecycle",
			pattern: `{"source":["aws.autoscaling"],"detail-type":["EC2 Instance-terminate Lifecycle Action"]}`,
		},
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
	}

	rebalanceEvent = event{
		name:    "RebalanceRecommendation",
		pattern: `{"source": ["aws.ec2"],"detail-type": ["EC2 Instance Rebalance Recommendation"]}`,
	}
)

type NodeTerminationHandlerBuilder struct {
	*AWSModelContext

	Lifecycle fi.Lifecycle
}

func (b *NodeTerminationHandlerBuilder) Build(c *fi.CloudupModelBuilderContext) error {
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

func (b *NodeTerminationHandlerBuilder) configureASG(c *fi.CloudupModelBuilderContext, ig *kops.InstanceGroup) error {
	name := ig.Name + "-NTHLifecycleHook"

	lifecyleTask := &awstasks.AutoscalingLifecycleHook{
		ID:                  aws.String(name),
		Name:                aws.String(name),
		Lifecycle:           b.Lifecycle,
		AutoscalingGroup:    b.LinkToAutoscalingGroup(ig),
		DefaultResult:       aws.String("CONTINUE"),
		HeartbeatTimeout:    aws.Int32(DefaultMessageRetentionPeriod),
		LifecycleTransition: aws.String("autoscaling:EC2_INSTANCE_TERMINATING"),
		Enabled:             aws.Bool(true),
	}

	c.AddTask(lifecyleTask)

	return nil
}

func (b *NodeTerminationHandlerBuilder) build(c *fi.CloudupModelBuilderContext) error {
	queueName := model.QueueNamePrefix(b.ClusterName()) + "-nth"

	policy := iam.NewPolicy(b.ClusterName(), b.AWSPartition)
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
			Service: fi.PtrTo(stringorset.Of("events.amazonaws.com", "sqs.amazonaws.com")),
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

	events := append([]event(nil), fixedEvents...)
	if b.Cluster.Spec.CloudProvider.AWS.NodeTerminationHandler != nil && fi.ValueOf(b.Cluster.Spec.CloudProvider.AWS.NodeTerminationHandler.EnableRebalanceDraining) {
		events = append(events, rebalanceEvent)
	}

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

func (b *NodeTerminationHandlerBuilder) FindDeletions(c *fi.CloudupModelBuilderContext, cloud fi.Cloud) error {
	if b.Cluster.Spec.CloudProvider.AWS.NodeTerminationHandler != nil && fi.ValueOf(b.Cluster.Spec.CloudProvider.AWS.NodeTerminationHandler.EnableRebalanceDraining) {
		return nil
	}

	clusterName := b.ClusterName()
	clusterNamePrefix := awsup.GetClusterName40(clusterName)
	ruleName := aws.String(clusterNamePrefix + "-" + rebalanceEvent.name)

	eventBridge := cloud.(awsup.AWSCloud).EventBridge()
	request := &eventbridge.ListRulesInput{
		NamePrefix: ruleName,
	}
	response, err := eventBridge.ListRules(c.Context(), request)
	if err != nil {
		return fmt.Errorf("listing EventBridge rules: %w", err)
	}
	if response == nil || len(response.Rules) == 0 {
		return nil
	}
	if len(response.Rules) > 1 {
		return fmt.Errorf("found multiple EventBridge rules with the same name %s", *ruleName)
	}

	rule := response.Rules[0]

	tagResponse, err := eventBridge.ListTagsForResource(c.Context(), &eventbridge.ListTagsForResourceInput{ResourceARN: rule.Arn})
	if err != nil {
		return fmt.Errorf("listing tags for EventBridge rule: %w", err)
	}

	owned := false
	ownershipTag := "kubernetes.io/cluster/" + b.Cluster.ObjectMeta.Name
	for _, tag := range tagResponse.Tags {
		if fi.ValueOf(tag.Key) == ownershipTag && fi.ValueOf(tag.Value) == "owned" {
			owned = true
			break
		}
	}
	if !owned {
		return nil
	}

	ruleTask := &awstasks.EventBridgeRule{
		Name:      ruleName,
		Lifecycle: b.Lifecycle,
	}
	c.AddTask(ruleTask)

	return nil
}
