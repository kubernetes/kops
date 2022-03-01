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

package mockiam

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"google.golang.org/api/iam/v1"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
)

type serviceAccountClient struct {
	// serviceaccounts are keyed by name.
	serviceaccounts map[string]*iam.ServiceAccount
	project         string
	sync.Mutex
}

var _ gce.ServiceAccountClient = &serviceAccountClient{}

func newServiceAccounts(project string) *serviceAccountClient {
	return &serviceAccountClient{
		serviceaccounts: map[string]*iam.ServiceAccount{},
		project:         project,
	}
}

func (s *serviceAccountClient) Get(ctx context.Context, name string) (*iam.ServiceAccount, error) {
	s.Lock()
	defer s.Unlock()
	result, ok := s.serviceaccounts[name]
	if !ok {
		return nil, notFoundError()
	}
	return result, nil
}

func (s *serviceAccountClient) Update(ctx context.Context, name string, sa *iam.ServiceAccount) (*iam.ServiceAccount, error) {
	s.Lock()
	defer s.Unlock()
	s.serviceaccounts[name] = sa
	return s.serviceaccounts[name], nil
}

func (s *serviceAccountClient) Create(ctx context.Context, name string, req *iam.CreateServiceAccountRequest) (*iam.ServiceAccount, error) {
	s.Lock()
	defer s.Unlock()
	req.ServiceAccount.Email = fmt.Sprintf("%s@%s.iam.gserviceaccount.com", req.AccountId, s.project)
	fqn := fmt.Sprintf("%s/serviceAccounts/%s", name, req.ServiceAccount.Email)
	req.ServiceAccount.Name = fqn
	s.serviceaccounts[fqn] = req.ServiceAccount
	return s.serviceaccounts[fqn], nil
}

func (s *serviceAccountClient) Delete(name string) (*iam.Empty, error) {
	s.Lock()
	defer s.Unlock()
	fqn := "projects/" + s.project + "/serviceAccounts/" + name
	if _, ok := s.serviceaccounts[fqn]; !ok {
		return nil, nil
	}
	delete(s.serviceaccounts, fqn)
	return nil, nil
}

func (s *serviceAccountClient) List(ctx context.Context, project string) ([]*iam.ServiceAccount, error) {
	s.Lock()
	defer s.Unlock()
	var r []*iam.ServiceAccount
	for k, v := range s.serviceaccounts {
		if strings.Contains(k, project) {
			r = append(r, v)
		}
	}
	return r, nil
}
