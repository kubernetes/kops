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

package hetzner

import (
	"fmt"
	"strconv"

	"github.com/hetznercloud/hcloud-go/hcloud/metadata"
	"k8s.io/kops/pkg/bootstrap"
)

const HetznerAuthenticationTokenPrefix = "x-hetzner-id "

type hetznerAuthenticator struct {
}

var _ bootstrap.Authenticator = &hetznerAuthenticator{}

func NewHetznerAuthenticator() (bootstrap.Authenticator, error) {
	return &hetznerAuthenticator{}, nil
}

func (h hetznerAuthenticator) CreateToken(body []byte, jwt string) (string, error) {
	serverID, err := metadata.NewClient().InstanceID()
	if err != nil {
		return "", fmt.Errorf("failed to retrieve server ID: %w", err)
	}

	return HetznerAuthenticationTokenPrefix + strconv.Itoa(serverID), nil
}
