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

package openstack

import (
	"fmt"

	"k8s.io/kops/pkg/bootstrap"
)

const OpenstackAuthenticationTokenPrefix = "x-openstack-id "

type openstackAuthenticator struct {
}

var _ bootstrap.Authenticator = &openstackAuthenticator{}

func NewOpenstackAuthenticator() (bootstrap.Authenticator, error) {
	return &openstackAuthenticator{}, nil
}

func (o openstackAuthenticator) CreateToken(body []byte, jwtToken string) (string, error) {
	metadata, err := GetLocalMetadata()
	if err != nil {
		return "", fmt.Errorf("unable to fetch metadata: %w", err)
	}
	return fmt.Sprintf("%s%s:%s", OpenstackAuthenticationTokenPrefix, metadata.ServerID, jwtToken), nil
}
