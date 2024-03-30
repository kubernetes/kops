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

package mockiam

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"k8s.io/klog/v2"
)

func (m *MockIAM) ListOpenIDConnectProviders(ctx context.Context, params *iam.ListOpenIDConnectProvidersInput, optFns ...func(*iam.Options)) (*iam.ListOpenIDConnectProvidersOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	providers := make([]iamtypes.OpenIDConnectProviderListEntry, 0)
	for arn := range m.OIDCProviders {
		providers = append(providers, iamtypes.OpenIDConnectProviderListEntry{
			Arn: &arn,
		})
	}
	response := &iam.ListOpenIDConnectProvidersOutput{
		OpenIDConnectProviderList: providers,
	}
	return response, nil
}

func (m *MockIAM) GetOpenIDConnectProvider(ctx context.Context, request *iam.GetOpenIDConnectProviderInput, optFns ...func(*iam.Options)) (*iam.GetOpenIDConnectProviderOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	arn := aws.ToString(request.OpenIDConnectProviderArn)

	provider, ok := m.OIDCProviders[arn]
	if !ok {
		return nil, fmt.Errorf("OpenIDConnectProvider with arn=%q not found", arn)
	}

	response := &iam.GetOpenIDConnectProviderOutput{
		ClientIDList:   provider.ClientIDList,
		CreateDate:     provider.CreateDate,
		Tags:           provider.Tags,
		ThumbprintList: provider.ThumbprintList,
		Url:            provider.Url,
	}
	return response, nil
}

func (m *MockIAM) CreateOpenIDConnectProvider(ctx context.Context, request *iam.CreateOpenIDConnectProviderInput, optFns ...func(*iam.Options)) (*iam.CreateOpenIDConnectProviderOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("CreateOpenIDConnectProvider: %v", request)

	arn := fmt.Sprintf("arn:aws-test:iam::0000000000:oidc-provider/%s", *request.Url)

	p := iam.GetOpenIDConnectProviderOutput{
		ClientIDList:   request.ClientIDList,
		Tags:           request.Tags,
		ThumbprintList: request.ThumbprintList,
		Url:            request.Url,
	}

	if m.OIDCProviders == nil {
		m.OIDCProviders = make(map[string]*iam.GetOpenIDConnectProviderOutput)
	}
	m.OIDCProviders[arn] = &p

	return &iam.CreateOpenIDConnectProviderOutput{OpenIDConnectProviderArn: &arn}, nil
}

func (m *MockIAM) DeleteOpenIDConnectProvider(ctx context.Context, request *iam.DeleteOpenIDConnectProviderInput, optFns ...func(*iam.Options)) (*iam.DeleteOpenIDConnectProviderOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DeleteOpenIDConnectProvider: %v", request)

	arn := aws.ToString(request.OpenIDConnectProviderArn)
	_, ok := m.OIDCProviders[arn]
	if !ok {
		return nil, fmt.Errorf("OIDCProvider %q not found", arn)
	}
	delete(m.OIDCProviders, arn)

	return &iam.DeleteOpenIDConnectProviderOutput{}, nil
}

func (m *MockIAM) UpdateOpenIDConnectProviderThumbprint(ctx context.Context, params *iam.UpdateOpenIDConnectProviderThumbprintInput, optFns ...func(*iam.Options)) (*iam.UpdateOpenIDConnectProviderThumbprintOutput, error) {
	panic("Not implemented")
}
