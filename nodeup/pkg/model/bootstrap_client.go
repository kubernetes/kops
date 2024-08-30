/*
Copyright 2020 The Kubernetes Authors.

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

package model

import (
	"fmt"
	"net"
	"net/url"
	"strconv"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/bootstrap"
	"k8s.io/kops/pkg/bootstrap/pkibootstrap"
	"k8s.io/kops/pkg/kopscontrollerclient"
	"k8s.io/kops/pkg/wellknownports"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"
	"k8s.io/kops/upup/pkg/fi/cloudup/do"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce/tpm/gcetpmsigner"
	"k8s.io/kops/upup/pkg/fi/cloudup/hetzner"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
	"k8s.io/kops/upup/pkg/fi/cloudup/scaleway"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

// BootstrapClientBuilder calls kops-controller to bootstrap the node.
type BootstrapClientBuilder struct {
	*NodeupModelContext
}

func (b BootstrapClientBuilder) Build(c *fi.NodeupModelBuilderContext) error {
	if b.IsMaster {
		return nil
	}

	var authenticator bootstrap.Authenticator

	switch b.CloudProvider() {
	case kops.CloudProviderAWS:
		a, err := awsup.NewAWSAuthenticator(c.Context(), b.Cloud.Region())
		if err != nil {
			return err
		}
		authenticator = a
	case kops.CloudProviderGCE:
		a, err := gcetpmsigner.NewTPMAuthenticator()
		if err != nil {
			return err
		}
		authenticator = a
	case kops.CloudProviderHetzner:
		a, err := hetzner.NewHetznerAuthenticator()
		if err != nil {
			return err
		}
		authenticator = a
	case kops.CloudProviderOpenstack:
		a, err := openstack.NewOpenstackAuthenticator()
		if err != nil {
			return err
		}
		authenticator = a
	case kops.CloudProviderDO:
		a, err := do.NewAuthenticator()
		if err != nil {
			return err
		}
		authenticator = a
	case kops.CloudProviderScaleway:
		a, err := scaleway.NewScalewayAuthenticator()
		if err != nil {
			return err
		}
		authenticator = a
	case kops.CloudProviderAzure:
		a, err := azure.NewAzureAuthenticator()
		if err != nil {
			return err
		}
		authenticator = a

	case kops.CloudProviderMetal:
		a, err := pkibootstrap.NewAuthenticatorFromFile("/etc/kubernetes/kops/pki/machine/private.pem")
		if err != nil {
			return err
		}
		authenticator = a

	default:
		return fmt.Errorf("unsupported cloud provider for authenticator %q", b.CloudProvider())
	}

	baseURL := url.URL{
		Scheme: "https",
		Host:   net.JoinHostPort("kops-controller.internal."+b.NodeupConfig.ClusterName, strconv.Itoa(wellknownports.KopsControllerPort)),
		Path:   "/",
	}

	bootstrapClient := &kopscontrollerclient.Client{
		Authenticator: authenticator,
		CAs:           []byte(b.NodeupConfig.CAs[fi.CertificateIDCA]),
		BaseURL:       baseURL,
	}

	bootstrapClientTask := &nodetasks.BootstrapClientTask{
		Client:     bootstrapClient,
		Certs:      b.bootstrapCerts,
		KeypairIDs: b.bootstrapKeypairIDs,
	}
	bootstrapClientTask.UseChallengeCallback = b.UseChallengeCallback(b.CloudProvider())
	bootstrapClientTask.ClusterName = b.NodeupConfig.ClusterName

	for _, cert := range b.bootstrapCerts {
		cert.Cert.Task = bootstrapClientTask
		cert.Key.Task = bootstrapClientTask
	}

	c.AddTask(bootstrapClientTask)
	return nil
}

var _ fi.NodeupModelBuilder = &BootstrapClientBuilder{}
