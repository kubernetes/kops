/*
Copyright 2022 The Kubernetes Authors.

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

package gce

import (
	"context"
	"fmt"

	"google.golang.org/api/iam/v1"
)

type IamClient interface {
	ServiceAccounts() ServiceAccountClient
}

type iamClientImpl struct {
	srv *iam.Service
}

var _ IamClient = &iamClientImpl{}

func newIamClientImpl(ctx context.Context) (*iamClientImpl, error) {
	srv, err := iam.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("error building iam API client: %v", err)
	}
	return &iamClientImpl{
		srv: srv,
	}, nil
}

func (i *iamClientImpl) ServiceAccounts() ServiceAccountClient {
	return &serviceAccountClientImpl{
		srv: i.srv.Projects.ServiceAccounts,
	}
}

type ServiceAccountClient interface {
	Get(ctx context.Context, fqn string) (*iam.ServiceAccount, error)
	Create(ctx context.Context, project string, req *iam.CreateServiceAccountRequest) (*iam.ServiceAccount, error)
	Update(ctx context.Context, fqn string, sa *iam.ServiceAccount) (*iam.ServiceAccount, error)
	Delete(saName string) (*iam.Empty, error)
	List(ctx context.Context, project string) ([]*iam.ServiceAccount, error)
}

type serviceAccountClientImpl struct {
	srv *iam.ProjectsServiceAccountsService
}

var _ ServiceAccountClient = &serviceAccountClientImpl{}

func (s *serviceAccountClientImpl) Create(ctx context.Context, project string, req *iam.CreateServiceAccountRequest) (*iam.ServiceAccount, error) {
	return s.srv.Create(project, req).Context(ctx).Do()
}

func (s *serviceAccountClientImpl) Update(ctx context.Context, fqn string, sa *iam.ServiceAccount) (*iam.ServiceAccount, error) {
	return s.srv.Update(fqn, sa).Context(ctx).Do()
}

func (s *serviceAccountClientImpl) Get(ctx context.Context, fqn string) (*iam.ServiceAccount, error) {
	return s.srv.Get(fqn).Context(ctx).Do()
}

func (s *serviceAccountClientImpl) List(ctx context.Context, project string) ([]*iam.ServiceAccount, error) {
	var sas []*iam.ServiceAccount
	if err := s.srv.List(project).Pages(ctx, func(p *iam.ListServiceAccountsResponse) error {
		sas = append(sas, p.Accounts...)
		return nil
	}); err != nil {
		return nil, err
	}
	return sas, nil
}

func (s *serviceAccountClientImpl) Delete(saName string) (*iam.Empty, error) {
	return s.srv.Delete(saName).Do()
}
