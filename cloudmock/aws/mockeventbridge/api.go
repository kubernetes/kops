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

package mockeventbridge

import (
	"context"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	eventbridgetypes "github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"k8s.io/kops/util/pkg/awsinterfaces"
)

type MockEventBridge struct {
	awsinterfaces.EventBridgeAPI
	mutex sync.Mutex

	Rules         map[string]*eventbridgetypes.Rule
	TagsByArn     map[string][]eventbridgetypes.Tag
	TargetsByRule map[string][]eventbridgetypes.Target
}

var _ awsinterfaces.EventBridgeAPI = &MockEventBridge{}

func (m *MockEventBridge) PutRule(ctx context.Context, input *eventbridge.PutRuleInput, optFns ...func(*eventbridge.Options)) (*eventbridge.PutRuleOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	name := *input.Name
	arn := "arn:aws-test:events:us-east-1:012345678901:rule/" + name

	rule := &eventbridgetypes.Rule{
		Arn:          &arn,
		EventPattern: input.EventPattern,
	}
	if m.Rules == nil {
		m.Rules = make(map[string]*eventbridgetypes.Rule)
	}
	if m.TagsByArn == nil {
		m.TagsByArn = make(map[string][]eventbridgetypes.Tag)
	}
	m.Rules[name] = rule
	m.TagsByArn[arn] = input.Tags

	response := &eventbridge.PutRuleOutput{
		RuleArn: &arn,
	}
	return response, nil
}

func (m *MockEventBridge) ListRules(ctx context.Context, input *eventbridge.ListRulesInput, optFns ...func(*eventbridge.Options)) (*eventbridge.ListRulesOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	response := &eventbridge.ListRulesOutput{}

	rule := m.Rules[*input.NamePrefix]
	if rule == nil {
		return response, nil
	}
	response.Rules = []eventbridgetypes.Rule{*rule}
	return response, nil
}

func (m *MockEventBridge) DeleteRule(ctx context.Context, input *eventbridge.DeleteRuleInput, optFns ...func(*eventbridge.Options)) (*eventbridge.DeleteRuleOutput, error) {
	panic("Not implemented")
}

func (m *MockEventBridge) ListTagsForResource(ctx context.Context, input *eventbridge.ListTagsForResourceInput, optFns ...func(*eventbridge.Options)) (*eventbridge.ListTagsForResourceOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	response := &eventbridge.ListTagsForResourceOutput{
		Tags: m.TagsByArn[*input.ResourceARN],
	}
	return response, nil
}

func (m *MockEventBridge) PutTargets(ctx context.Context, input *eventbridge.PutTargetsInput, optFns ...func(*eventbridge.Options)) (*eventbridge.PutTargetsOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.TargetsByRule == nil {
		m.TargetsByRule = make(map[string][]eventbridgetypes.Target)
	}
	m.TargetsByRule[*input.Rule] = input.Targets

	return &eventbridge.PutTargetsOutput{}, nil
}

func (m *MockEventBridge) ListTargetsByRule(ctx context.Context, input *eventbridge.ListTargetsByRuleInput, optFns ...func(*eventbridge.Options)) (*eventbridge.ListTargetsByRuleOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	response := &eventbridge.ListTargetsByRuleOutput{
		Targets: m.TargetsByRule[*input.Rule],
	}
	return response, nil
}

func (m *MockEventBridge) RemoveTargets(ctx context.Context, input *eventbridge.RemoveTargetsInput, optFns ...func(*eventbridge.Options)) (*eventbridge.RemoveTargetsOutput, error) {
	panic("Not implemented")
}
