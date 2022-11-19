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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
)

// +kops:fitask
type EventBridgeRule struct {
	ID        *string
	Name      *string
	Lifecycle fi.Lifecycle

	EventPattern *string
	SQSQueue     *SQS

	Tags map[string]string
}

var _ fi.CompareWithID = &EventBridgeRule{}

func (eb *EventBridgeRule) CompareWithID() *string {
	return eb.Name
}

func (eb *EventBridgeRule) Find(c *fi.Context) (*EventBridgeRule, error) {
	cloud := c.Cloud.(awsup.AWSCloud)

	if eb.Name == nil {
		return nil, nil
	}

	request := &eventbridge.ListRulesInput{
		NamePrefix: eb.Name,
	}
	response, err := cloud.EventBridge().ListRules(request)
	if err != nil {
		return nil, fmt.Errorf("error listing EventBridge rules: %v", err)
	}
	if response == nil || len(response.Rules) == 0 {
		return nil, nil
	}
	if len(response.Rules) > 1 {
		return nil, fmt.Errorf("found multiple EventBridge rules with the same name")
	}

	rule := response.Rules[0]

	tagResponse, err := cloud.EventBridge().ListTagsForResource(&eventbridge.ListTagsForResourceInput{ResourceARN: rule.Arn})
	if err != nil {
		return nil, fmt.Errorf("error listing tags for EventBridge rule: %v", err)
	}

	actual := &EventBridgeRule{
		ID:           eb.ID,
		Name:         eb.Name,
		Lifecycle:    eb.Lifecycle,
		EventPattern: rule.EventPattern,
		SQSQueue:     eb.SQSQueue,
		Tags:         mapEventBridgeTagsToMap(tagResponse.Tags),
	}
	return actual, nil
}

func (eb *EventBridgeRule) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(eb, c)
}

func (_ *EventBridgeRule) CheckChanges(a, e, changes *EventBridgeRule) error {
	if a == nil {
		if e.Name == nil {
			return field.Required(field.NewPath("Name"), "")
		}
	}

	return nil
}

func (eb *EventBridgeRule) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *EventBridgeRule) error {
	if a == nil {
		var tags []*eventbridge.Tag
		for k, v := range eb.Tags {
			tags = append(tags, &eventbridge.Tag{
				Key:   aws.String(k),
				Value: aws.String(v),
			})
		}

		request := &eventbridge.PutRuleInput{
			Name:         eb.Name,
			EventPattern: e.EventPattern,
			Tags:         tags,
		}

		_, err := t.Cloud.EventBridge().PutRule(request)
		if err != nil {
			return fmt.Errorf("error creating EventBridge rule: %v", err)
		}
	}

	return nil
}

type terraformEventBridgeRule struct {
	Name         *string                  `cty:"name"`
	EventPattern *terraformWriter.Literal `cty:"event_pattern"`
	Tags         map[string]string        `cty:"tags"`
}

func (_ *EventBridgeRule) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *EventBridgeRule) error {
	m, err := t.AddFileBytes("aws_cloudwatch_event_rule", *e.Name, "event_pattern", []byte(*e.EventPattern), false)
	if err != nil {
		return err
	}

	tf := &terraformEventBridgeRule{
		Name:         e.Name,
		EventPattern: m,
		Tags:         e.Tags,
	}

	return t.RenderResource("aws_cloudwatch_event_rule", *e.Name, tf)
}

func (eb *EventBridgeRule) TerraformLink() *terraformWriter.Literal {
	return terraformWriter.LiteralProperty("aws_cloudwatch_event_rule", fi.ValueOf(eb.Name), "id")
}
