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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/iam"
	"k8s.io/klog/v2"
)

func (m *MockIAM) ListOpenIDConnectProviders(request *iam.ListOpenIDConnectProvidersInput) (*iam.ListOpenIDConnectProvidersOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	providers := make([]*iam.OpenIDConnectProviderListEntry, 0)
	for arn := range m.OIDCProviders {
		providers = append(providers, &iam.OpenIDConnectProviderListEntry{
			Arn: &arn,
		})
	}
	response := &iam.ListOpenIDConnectProvidersOutput{
		OpenIDConnectProviderList: providers,
	}
	return response, nil
}

func (m *MockIAM) ListOpenIDConnectProvidersWithContext(aws.Context, *iam.ListOpenIDConnectProvidersInput, ...request.Option) (*iam.ListOpenIDConnectProvidersOutput, error) {
	panic("Not implemented")
}
func (m *MockIAM) ListOpenIDConnectProvidersRequest(*iam.ListOpenIDConnectProvidersInput) (*request.Request, *iam.ListOpenIDConnectProvidersOutput) {
	panic("Not implemented")
}

func (m *MockIAM) GetOpenIDConnectProviderWithContext(ctx aws.Context, request *iam.GetOpenIDConnectProviderInput, options ...request.Option) (*iam.GetOpenIDConnectProviderOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	arn := aws.StringValue(request.OpenIDConnectProviderArn)

	provider := m.OIDCProviders[arn]
	if provider == nil {
		return nil, fmt.Errorf("OpenIDConnectProvider with arn=%q not found", arn)
	}

	response := &iam.GetOpenIDConnectProviderOutput{
		ClientIDList:   provider.ClientIDList,
		CreateDate:     provider.CreateDate,
		ThumbprintList: provider.ThumbprintList,
		Url:            provider.Url,
	}
	return response, nil
}

func (m *MockIAM) GetOpenIDConnectProvider(request *iam.GetOpenIDConnectProviderInput) (*iam.GetOpenIDConnectProviderOutput, error) {
	return m.GetOpenIDConnectProviderWithContext(context.Background(), request)
}

func (m *MockIAM) GetOpenIDConnectProviderRequest(*iam.GetOpenIDConnectProviderInput) (*request.Request, *iam.GetOpenIDConnectProviderOutput) {
	panic("Not implemented")
}

func (m *MockIAM) CreateOpenIDConnectProvider(request *iam.CreateOpenIDConnectProviderInput) (*iam.CreateOpenIDConnectProviderOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("CreateOpenIDConnectProvider: %v", request)

	arn := fmt.Sprintf("arn:aws:iam::0000000000:oidc-provider/%s", *request.Url)

	p := &iam.GetOpenIDConnectProviderOutput{
		ClientIDList:   request.ClientIDList,
		ThumbprintList: request.ThumbprintList,
		Url:            request.Url,
	}

	if m.OIDCProviders == nil {
		m.OIDCProviders = make(map[string]*iam.GetOpenIDConnectProviderOutput)
	}
	m.OIDCProviders[arn] = p

	return &iam.CreateOpenIDConnectProviderOutput{OpenIDConnectProviderArn: &arn}, nil
}

func (m *MockIAM) CreateOpenIDConnectProviderWithContext(aws.Context, *iam.CreateOpenIDConnectProviderInput, ...request.Option) (*iam.CreateOpenIDConnectProviderOutput, error) {
	panic("Not implemented")
}
func (m *MockIAM) CreateOpenIDConnectProviderRequest(*iam.CreateOpenIDConnectProviderInput) (*request.Request, *iam.CreateOpenIDConnectProviderOutput) {
	panic("Not implemented")
}

func (m *MockIAM) DeleteOpenIDConnectProvider(request *iam.DeleteOpenIDConnectProviderInput) (*iam.DeleteOpenIDConnectProviderOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DeleteOpenIDConnectProvider: %v", request)

	arn := aws.StringValue(request.OpenIDConnectProviderArn)
	o := m.OIDCProviders[arn]
	if o == nil {
		return nil, fmt.Errorf("OIDCProvider %q not found", arn)
	}
	delete(m.OIDCProviders, arn)

	return &iam.DeleteOpenIDConnectProviderOutput{}, nil
}

func (m *MockIAM) DeleteOpenIDConnectProviderWithContext(aws.Context, *iam.DeleteOpenIDConnectProviderInput, ...request.Option) (*iam.DeleteOpenIDConnectProviderOutput, error) {
	panic("Not implemented")
}

func (m *MockIAM) DeleteOpenIDConnectProviderRequest(*iam.DeleteOpenIDConnectProviderInput) (*request.Request, *iam.DeleteOpenIDConnectProviderOutput) {
	panic("Not implemented")
}

func (m *MockIAM) UpdateOpenIDConnectProviderThumbprint(*iam.UpdateOpenIDConnectProviderThumbprintInput) (*iam.UpdateOpenIDConnectProviderThumbprintOutput, error) {
	panic("Not implemented")
}
func (m *MockIAM) UpdateOpenIDConnectProviderThumbprintWithContext(aws.Context, *iam.UpdateOpenIDConnectProviderThumbprintInput, ...request.Option) (*iam.UpdateOpenIDConnectProviderThumbprintOutput, error) {
	panic("Not implemented")
}
func (m *MockIAM) UpdateOpenIDConnectProviderThumbprintRequest(*iam.UpdateOpenIDConnectProviderThumbprintInput) (*request.Request, *iam.UpdateOpenIDConnectProviderThumbprintOutput) {
	panic("Not implemented")
}
