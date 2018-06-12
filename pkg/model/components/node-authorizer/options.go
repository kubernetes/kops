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
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model/components"
	"k8s.io/kops/upup/pkg/fi/loader"
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
			// @check the authorizer method is set
			if na.NodeAuthorizer.Authorizer == "" {
				switch kops.CloudProviderID(cs.CloudProvider) {
				case kops.CloudProviderAWS:
					na.NodeAuthorizer.Authorizer = "aws"
				default:
					na.NodeAuthorizer.Authorizer = "alwaysallow"
				}
			}

			// @check the image is set
			if na.NodeAuthorizer.Image == "" {
				// @TODO !!!!NEEDS TO BE CHANGED!!!!
				na.NodeAuthorizer.Image = "quay.io/gambol99/node-authorizer:latest"
			}

			// @check the port, authorization time and tokenttl
			if na.NodeAuthorizer.Port == 0 {
				na.NodeAuthorizer.Port = DefaultPort
			}
			if na.NodeAuthorizer.Timeout == nil {
				na.NodeAuthorizer.Timeout = DefaultTimeout
			}
			if na.NodeAuthorizer.TokenTTL == nil {
				na.NodeAuthorizer.TokenTTL = DefaultTokenTTL
			}

			// @check if the node url is set, @TODO this should probably be set in a global somewhere
			if na.NodeAuthorizer.NodeURL == "" {
				na.NodeAuthorizer.NodeURL = fmt.Sprintf("https://node-authorizer-internal.%s:%d", b.Context.ClusterName, na.NodeAuthorizer.Port)
			}
		}
	}

	return nil
}
