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

package mockssm

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"k8s.io/kops/util/pkg/awsinterfaces"
)

type MockSSM struct {
	awsinterfaces.SSMAPI
	mutex sync.Mutex

	Parameters map[string]string
}

var _ awsinterfaces.SSMAPI = &MockSSM{}

func (m *MockSSM) GetParameter(ctx context.Context, input *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if input == nil || input.Name == nil {
		return nil, fmt.Errorf("invalid GetParameter input: missing parameter name")
	}

	val, ok := m.Parameters[*input.Name]
	if !ok {
		return nil, fmt.Errorf("SSM parameter %q not found", *input.Name)
	}

	return &ssm.GetParameterOutput{
		Parameter: &ssmtypes.Parameter{
			Name:  input.Name,
			Value: aws.String(val),
		},
	}, nil
}
