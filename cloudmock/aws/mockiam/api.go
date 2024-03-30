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

package mockiam

import (
	"fmt"
	"math/rand"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"k8s.io/kops/util/pkg/awsinterfaces"
)

type MockIAM struct {
	// Mock out interface
	awsinterfaces.IAMAPI

	mutex            sync.Mutex
	InstanceProfiles map[string]*iamtypes.InstanceProfile
	Roles            map[string]*iamtypes.Role
	OIDCProviders    map[string]*iam.GetOpenIDConnectProviderOutput
	RolePolicies     []*rolePolicy
	AttachedPolicies map[string][]iamtypes.AttachedPolicy
}

var _ awsinterfaces.IAMAPI = &MockIAM{}

func (m *MockIAM) createID() string {
	return "AID" + fmt.Sprintf("%x", rand.Int63())
}
