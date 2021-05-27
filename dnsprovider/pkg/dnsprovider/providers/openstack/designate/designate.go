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

package designate

import (
	"fmt"
	"io"

	"github.com/gophercloud/gophercloud/openstack"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	osauth "k8s.io/kops/upup/pkg/fi/cloudup/openstackauth"
)

const (
	// ProviderName is the name of this DNS provider
	ProviderName = "openstack-designate"
)

func init() {
	dnsprovider.RegisterDNSProvider(ProviderName, func(config io.Reader) (dnsprovider.Interface, error) {
		return NewDesignateClient(config)
	})
}

func NewDesignateClient(_ io.Reader) (*Interface, error) {
	InsecureSkipVerify := true
	provider, err := osauth.NewOpenStackProvider(InsecureSkipVerify)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate to OpenStack: %v", err)
	}

	client, err := openstack.NewDNSV2(provider, osauth.GetOpenStackEndpointOpts())
	if err != nil {
		return nil, fmt.Errorf("error building dns client: %v", err)
	}

	return New(client), nil
}
