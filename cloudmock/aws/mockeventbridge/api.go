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
	Rules []*eventbridge.Rule
}

var _ eventbridgeiface.EventBridgeAPI = &MockEventBridge{}

func (c *MockEventBridge) ListTargetsByRule(*eventbridge.ListTargetsByRuleInput) (*eventbridge.ListTargetsByRuleOutput, error) {
	panic("Not implemented")
}

func (c *MockEventBridge) RemoveTargets(*eventbridge.RemoveTargetsInput) (*eventbridge.RemoveTargetsOutput, error) {
	panic("Not implemented")
}

func (c *MockEventBridge) DeleteRule(*eventbridge.DeleteRuleInput) (*eventbridge.DeleteRuleOutput, error) {
	panic("Not implemented")
}

func (c *MockEventBridge) ListRules(*eventbridge.ListRulesInput) (*eventbridge.ListRulesOutput, error) {
	response := &eventbridge.ListRulesOutput{
		Rules: c.Rules,
	}

	return response, nil
}

func (c *MockEventBridge) PutRule(*eventbridge.PutRuleInput) (*eventbridge.PutRuleOutput, error) {
	panic("Not implemented")
}

func (c *MockEventBridge) PutTargets(*eventbridge.PutTargetsInput) (*eventbridge.PutTargetsOutput, error) {
	panic("Not implemented")
}
