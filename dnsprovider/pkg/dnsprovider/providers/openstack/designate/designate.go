package designate

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"

	"github.com/golang/glog"
	"github.com/gophercloud/gophercloud/openstack"
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

	glog.V(2).Info("authenticating to keystone")

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

	}
	return New(sc), nil
}
