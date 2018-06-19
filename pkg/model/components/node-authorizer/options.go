/*
Copyright 2018 The Kubernetes Authors.

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

package nodeauthorizer

import (
	"errors"
	"fmt"
	"os"
	"time"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model/components"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/loader"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// OptionsBuilder fills in the default options for the node-authorizer
type OptionsBuilder struct {
	Context *components.OptionsContext
}

var _ loader.OptionsBuilder = &OptionsBuilder{}

var (
	// DefaultPort is the default port to listen on
	DefaultPort = 10443
	// DefaultTimeout is the max time we are willing to wait before erroring
	DefaultTimeout = &metav1.Duration{Duration: 20 * time.Second}
	// DefaultTokenTTL is the default expiration on a bootstrap token
	DefaultTokenTTL = &metav1.Duration{Duration: 5 * time.Minute}
)

// BuildOptions generates the configurations used to create node authorizer
func (b *OptionsBuilder) BuildOptions(o interface{}) error {
	cs, ok := o.(*kops.ClusterSpec)
	if !ok {
		return errors.New("expected a ClusterSpec object")
	}

	if cs.NodeAuthorization != nil {
		na := cs.NodeAuthorization
		// NodeAuthorizerSpec
		if na.NodeAuthorizer != nil {
			if na.NodeAuthorizer.Authorizer == "" {
				switch kops.CloudProviderID(cs.CloudProvider) {
				case kops.CloudProviderAWS:
					na.NodeAuthorizer.Authorizer = "aws"
				default:
					na.NodeAuthorizer.Authorizer = "alwaysallow"
				}
			}
			if na.NodeAuthorizer.Image == "" {
				na.NodeAuthorizer.Image = GetNodeAuthorizerImage()
			}
			if na.NodeAuthorizer.Port == 0 {
				na.NodeAuthorizer.Port = DefaultPort
			}
			if na.NodeAuthorizer.Timeout == nil {
				na.NodeAuthorizer.Timeout = DefaultTimeout
			}
			if na.NodeAuthorizer.TokenTTL == nil {
				na.NodeAuthorizer.TokenTTL = DefaultTokenTTL
			}
			if na.NodeAuthorizer.EnableAddressCheck == nil {
				na.NodeAuthorizer.EnableAddressCheck = fi.Bool(true)
			}
			if na.NodeAuthorizer.EnableRegistrationCheck == nil {
				na.NodeAuthorizer.EnableRegistrationCheck = fi.Bool(true)
			}
			if na.NodeAuthorizer.NodeURL == "" {
				na.NodeAuthorizer.NodeURL = fmt.Sprintf("https://node-authorizer-internal.%s:%d", b.Context.ClusterName, na.NodeAuthorizer.Port)
			}
		}
	}

	return nil
}

// GetNodeAuthorizerImage returns the image to use for the node-authorizer
func GetNodeAuthorizerImage() string {
	if v := os.Getenv("NODE_AUTHORIZER_IMAGE"); v != "" {
		return v
	}

	return "quay.io/gambol99/node-authorizer:v0.0.1@sha256:b3ac87042a61ad62f3b95236654b85016343eac18a2a6cc9020465bd095a31e1"
}
