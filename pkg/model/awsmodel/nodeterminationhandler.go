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
	"slices"

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

// ruleName returns the name of the EventBridge rule that delivers the event of the cluster.
func (e event) ruleName(clusterName string) string {
	return awsup.GetClusterName40(clusterName) + "-" + e.name
}

var (
	_ fi.CloudupModelBuilder = &NodeTerminationHandlerBuilder{}
	_ fi.HasDeletions        = &NodeTerminationHandlerBuilder{}

	asgLifecycleEvent = event{
		name:    "ASGLifecycle",
		pattern: `{"source":["aws.autoscaling"],"detail-type":["EC2 Instance-terminate Lifecycle Action"]}`,
	}

	fixedEvents = []event{
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

func (b *NodeTerminationHandlerBuilder) rebalanceDrainingEnabled() bool {
	nth := b.Cluster.Spec.CloudProvider.AWS.NodeTerminationHandler
	return nth != nil && fi.ValueOf(nth.EnableRebalanceDraining)
}

// asgLifecycleEnabled returns true when the ASG lifecycle hook and its event are wanted. Karpenter
// never completes an ASG lifecycle action, so a hook would only hold terminating instances until
// its heartbeat timeout. A disabled hook task removes the hook that NTH left on the ASG.
func (b *NodeTerminationHandlerBuilder) asgLifecycleEnabled() bool {
	return !b.Cluster.Spec.Karpenter.IsEnabled()
}

// rebalanceEnabled returns true when rebalance recommendations are delivered to the queue.
// Karpenter always subscribes to them, as its reference setup does.
func (b *NodeTerminationHandlerBuilder) rebalanceEnabled() bool {
	return b.Cluster.Spec.Karpenter.IsEnabled() || b.rebalanceDrainingEnabled()
}

// events returns the EventBridge events delivered to the interruption queue.
func (b *NodeTerminationHandlerBuilder) events() []event {
	var events []event
	if b.asgLifecycleEnabled() {
		events = append(events, asgLifecycleEvent)
	}
	events = append(events, fixedEvents...)
	if b.rebalanceEnabled() {
		events = append(events, rebalanceEvent)
	}
	return events
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
		Enabled:             aws.Bool(b.asgLifecycleEnabled()),
	}

	c.AddTask(lifecyleTask)

	return nil
}

func (b *NodeTerminationHandlerBuilder) build(c *fi.CloudupModelBuilderContext) error {
	queueName := model.InterruptionQueueName(b.ClusterName())

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

	for _, event := range b.events() {
		// build rule
		ruleName := aws.String(event.ruleName(clusterName))
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

// FindDeletions deletes the rules of the events that the cluster no longer wants, left behind by a
// configuration that wanted them.
func (b *NodeTerminationHandlerBuilder) FindDeletions(c *fi.CloudupModelBuilderContext, cloud fi.Cloud) error {
	wanted := b.events()
	for _, e := range slices.Concat([]event{asgLifecycleEvent}, fixedEvents, []event{rebalanceEvent}) {
		if slices.Contains(wanted, e) {
			continue
		}
		if err := b.findRuleDeletion(c, cloud, e); err != nil {
			return err
		}
	}
	return nil
}

// findRuleDeletion adds a task that deletes the rule of the event, if the rule exists and is owned
// by the cluster.
func (b *NodeTerminationHandlerBuilder) findRuleDeletion(c *fi.CloudupModelBuilderContext, cloud fi.Cloud, e event) error {
	ruleName := aws.String(e.ruleName(b.ClusterName()))

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
