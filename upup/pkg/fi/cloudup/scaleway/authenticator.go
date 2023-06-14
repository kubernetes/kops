/*
Copyright 2023 The Kubernetes Authors.

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

package scaleway

import (
	"fmt"

	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"k8s.io/kops/pkg/bootstrap"
)

const ScalewayAuthenticationTokenPrefix = "x-scaleway-instance-server-id "

type scalewayAuthenticator struct{}

var _ bootstrap.Authenticator = &scalewayAuthenticator{}

func NewScalewayAuthenticator() (bootstrap.Authenticator, error) {
	return &scalewayAuthenticator{}, nil
}

func (a *scalewayAuthenticator) CreateToken(body []byte) (string, error) {
	metadataAPI := instance.NewMetadataAPI()
	metadata, err := metadataAPI.GetMetadata()
	if err != nil {
		return "", fmt.Errorf("failed to retrieve server metadata: %w", err)
	}
	return ScalewayAuthenticationTokenPrefix + metadata.ID, nil
}
