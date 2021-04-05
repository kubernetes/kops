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
	"sync"

	"github.com/aws/aws-sdk-go/service/eventbridge"
	"github.com/aws/aws-sdk-go/service/eventbridge/eventbridgeiface"
)

type MockEventBridge struct {
	eventbridgeiface.EventBridgeAPI
	mutex sync.Mutex

	Rules         map[string]*eventbridge.Rule
	TagsByArn     map[string][]*eventbridge.Tag
	TargetsByRule map[string][]*eventbridge.Target
}

var _ eventbridgeiface.EventBridgeAPI = &MockEventBridge{}

func (m *MockEventBridge) PutRule(input *eventbridge.PutRuleInput) (*eventbridge.PutRuleOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	name := *input.Name
	arn := "arn:aws:events:us-east-1:012345678901:rule/" + name

	rule := &eventbridge.Rule{
		Arn:          &arn,
		EventPattern: input.EventPattern,
	}
	if m.Rules == nil {
		m.Rules = make(map[string]*eventbridge.Rule)
	}
	if m.TagsByArn == nil {
		m.TagsByArn = make(map[string][]*eventbridge.Tag)
	}
	m.Rules[name] = rule
	m.TagsByArn[arn] = input.Tags

	response := &eventbridge.PutRuleOutput{
		RuleArn: &arn,
	}
	return response, nil
}

func (m *MockEventBridge) ListRules(input *eventbridge.ListRulesInput) (*eventbridge.ListRulesOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	response := &eventbridge.ListRulesOutput{}

	rule := m.Rules[*input.NamePrefix]
	if rule == nil {
		return response, nil
	}
	response.Rules = []*eventbridge.Rule{rule}
	return response, nil
}

func (m *MockEventBridge) DeleteRule(*eventbridge.DeleteRuleInput) (*eventbridge.DeleteRuleOutput, error) {
	panic("Not implemented")
}

func (m *MockEventBridge) ListTagsForResource(input *eventbridge.ListTagsForResourceInput) (*eventbridge.ListTagsForResourceOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	response := &eventbridge.ListTagsForResourceOutput{
		Tags: m.TagsByArn[*input.ResourceARN],
	}
	return response, nil
}

func (m *MockEventBridge) PutTargets(input *eventbridge.PutTargetsInput) (*eventbridge.PutTargetsOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.TargetsByRule == nil {
		m.TargetsByRule = make(map[string][]*eventbridge.Target)
	}
	m.TargetsByRule[*input.Rule] = input.Targets

	return &eventbridge.PutTargetsOutput{}, nil
}

func (m *MockEventBridge) ListTargetsByRule(input *eventbridge.ListTargetsByRuleInput) (*eventbridge.ListTargetsByRuleOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	response := &eventbridge.ListTargetsByRuleOutput{
		Targets: m.TargetsByRule[*input.Rule],
	}
	return response, nil
}

func (m *MockEventBridge) RemoveTargets(*eventbridge.RemoveTargetsInput) (*eventbridge.RemoveTargetsOutput, error) {
	panic("Not implemented")
}
