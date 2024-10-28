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
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/gophercloud/gophercloud/v2"
	gos "github.com/gophercloud/gophercloud/v2/openstack"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/servers"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/bootstrap"
	"k8s.io/kops/pkg/wellknownports"
)

type OpenStackVerifierOptions struct {
}

type openstackVerifier struct {
	novaClient *gophercloud.ServiceClient
	kubeClient *kubernetes.Clientset
}

var _ bootstrap.Verifier = &openstackVerifier{}

func NewOpenstackVerifier(opt *OpenStackVerifierOptions) (bootstrap.Verifier, error) {
	env, err := gos.AuthOptionsFromEnv()
	if err != nil {
		return nil, err
	}

	region := os.Getenv("OS_REGION_NAME")
	if region == "" {
		return nil, fmt.Errorf("unable to find region")
	}

	provider, err := gos.NewClient(env.IdentityEndpoint)
	if err != nil {
		return nil, err
	}
	ua := gophercloud.UserAgent{}
	ua.Prepend("kops/kopscontrollerverifier")
	provider.UserAgent = ua
	klog.V(4).Infof("Using user-agent %s", ua.Join())

	// node-controller should be able to renew it tokens against OpenStack API
	env.AllowReauth = true

	err = gos.Authenticate(context.TODO(), provider, env)
	if err != nil {
		return nil, err
	}

	novaClient, err := gos.NewComputeV2(provider, gophercloud.EndpointOpts{
		Type:   "compute",
		Region: region,
	})
	if err != nil {
		return nil, fmt.Errorf("error building nova client: %v", err)
	}

	kubeClient, err := newClientSet()
	if err != nil {
		return nil, fmt.Errorf("error building kubernetes client: %w", err)
	}

	return &openstackVerifier{
		novaClient: novaClient,
		kubeClient: kubeClient,
	}, nil
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func newClientSet() (*kubernetes.Clientset, error) {
	config, err := readKubeConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

// readKubeConfig ...
func readKubeConfig() (*restclient.Config, error) {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{}).ClientConfig()
}

func (o openstackVerifier) VerifyToken(ctx context.Context, rawRequest *http.Request, token string, body []byte) (*bootstrap.VerifyResult, error) {
	if !strings.HasPrefix(token, OpenstackAuthenticationTokenPrefix) {
		return nil, bootstrap.ErrNotThisVerifier
	}
	serverID := strings.TrimPrefix(token, OpenstackAuthenticationTokenPrefix)

	instance, err := servers.Get(ctx, o.novaClient, serverID).Extract()
	if err != nil {
		return nil, fmt.Errorf("failed to get info for server %q: %w", token, err)
	}

	var addrs []string

	var addresses map[string][]Address
	err = mapstructure.Decode(instance.Addresses, &addresses)
	if err != nil {
		return nil, fmt.Errorf("unable to decode addresses: %w", err)
	}

	for _, addrList := range addresses {
		for _, props := range addrList {
			addrs = append(addrs, props.Addr)
		}
	}
	// ensure that request is coming from same machine
	requestAddr, _, err := net.SplitHostPort(rawRequest.RemoteAddr)
	if err != nil {
		return nil, fmt.Errorf("invalid remote address %q: %v", rawRequest.RemoteAddr, err)
	}
	if !stringInSlice(requestAddr, addrs) {
		return nil, fmt.Errorf("authentication request address %q does not match server addresses %v", requestAddr, addrs)
	}

	// We will call back onto this address, now that we have verified it is an instance IP
	challengeEndpoint := net.JoinHostPort(requestAddr, strconv.Itoa(wellknownports.NodeupChallenge))

	// check from kubernetes API does the instance already exist
	_, err = o.kubeClient.CoreV1().Nodes().Get(ctx, instance.Name, v1.GetOptions{})
	if err == nil {
		return nil, bootstrap.ErrAlreadyExists
	}
	if err != nil && !errors.IsNotFound(err) {
		return nil, fmt.Errorf("got error while querying kubernetes api: %w", err)
	}

	result := &bootstrap.VerifyResult{
		NodeName:          instance.Name,
		CertificateNames:  addrs,
		ChallengeEndpoint: challengeEndpoint,
	}
	value, ok := instance.Metadata[TagKopsInstanceGroup]
	if ok {
		result.InstanceGroupName = value
	}
	return result, nil
}
