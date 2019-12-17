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
	"crypto/tls"
	"fmt"
	"io"
	"net/http"

	"github.com/gophercloud/gophercloud/openstack"
	"k8s.io/klog"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	"k8s.io/kops/util/pkg/vfs"
)

const (
	// ProviderName is the name of this DNS provider
	ProviderName = "openstack-designate"
)

func init() {
	dnsprovider.RegisterDnsProvider(ProviderName, func(config io.Reader) (dnsprovider.Interface, error) {
		return newDesignate(config)
	})
}

func newDesignate(_ io.Reader) (*Interface, error) {
	oc := vfs.OpenstackConfig{}
	ao, err := oc.GetCredential()
	if err != nil {
		return nil, err
	}

	/*
		pc, err := openstack.AuthenticatedClient(ao)
		if err != nil {
			return nil, fmt.Errorf("error building openstack authenticated client: %v", err)
		}*/

	provider, err := openstack.NewClient(ao.IdentityEndpoint)
	if err != nil {
		return nil, fmt.Errorf("error building openstack provider client: %v", err)
	}

	tlsconfig := &tls.Config{}
	tlsconfig.InsecureSkipVerify = true
	transport := &http.Transport{TLSClientConfig: tlsconfig}
	provider.HTTPClient = http.Client{
		Transport: transport,
	}

	klog.V(2).Info("authenticating to keystone")

	err = openstack.Authenticate(provider, ao)
	if err != nil {
		return nil, fmt.Errorf("error building openstack authenticated client: %v", err)
	}

	endpointOpt, err := oc.GetServiceConfig("Designate")
	if err != nil {
		return nil, err
	}
	sc, err := openstack.NewDNSV2(provider, endpointOpt)
	if err != nil {
		return nil, fmt.Errorf("error creating a ServiceClient: %v", err)
	}
	return New(sc), nil
}
