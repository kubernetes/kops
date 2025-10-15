/*
Copyright 2025 The Kubernetes Authors.

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

package elemento

import (
	// "fmt"
	// "strconv"

	// "github.com/Elemento-Modular-Cloud/ecloud-go/ecloud/metadata"
	"k8s.io/kops/pkg/bootstrap"
)

const ElementoAuthenticationTokenPrefix = "x-elemento-id "

type elementoAuthenticator struct {
}

var _ bootstrap.Authenticator = &elementoAuthenticator{}

func NewElementoAuthenticator() (bootstrap.Authenticator, error) {
	return &elementoAuthenticator{}, nil
}

func (h *elementoAuthenticator) CreateToken(body []byte) (string, error) {
	// DISABLED: Comment out metadata check for testing
	/*
		serverID, err := metadata.NewClient().InstanceID()
		if err != nil {
			return "", fmt.Errorf("failed to retrieve server ID: %w", err)
		}
		return ElementoAuthenticationTokenPrefix + strconv.Itoa(serverID), nil
	*/

	// DISABLED: Return a dummy token
	return ElementoAuthenticationTokenPrefix + "test-server-123", nil
}
