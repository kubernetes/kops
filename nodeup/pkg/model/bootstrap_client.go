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
	"k8s.io/kops/pkg/kopscontrollerclient"
	"k8s.io/kops/pkg/wellknownports"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce/tpm/gcetpmsigner"
	"k8s.io/kops/upup/pkg/fi/cloudup/hetzner"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

// BootstrapClientBuilder calls kops-controller to bootstrap the node.
type BootstrapClientBuilder struct {
	*NodeupModelContext
}

func (b BootstrapClientBuilder) Build(c *fi.NodeupModelBuilderContext) error {
	if b.IsMaster || !b.UseKopsControllerForNodeBootstrap() {
		return nil
	}

	var authenticator bootstrap.Authenticator
	var err error
	switch b.BootConfig.CloudProvider {
	case kops.CloudProviderAWS:
		authenticator, err = awsup.NewAWSAuthenticator(b.Cloud.Region())
	case kops.CloudProviderGCE:
		authenticator, err = gcetpmsigner.NewTPMAuthenticator()
		// We don't use the custom resolver here in gossip mode (though we could);
		// instead we use this as a check that protokube has now started.
	case kops.CloudProviderHetzner:
		authenticator, err = hetzner.NewHetznerAuthenticator()

	default:
		return fmt.Errorf("unsupported cloud provider for authenticator %q", b.BootConfig.CloudProvider)
	}

	if err != nil {
		return err
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

	for _, cert := range b.bootstrapCerts {
		cert.Cert.Task = bootstrapClientTask
		cert.Key.Task = bootstrapClientTask
	}

	c.AddTask(bootstrapClientTask)
	return nil
}

var _ fi.NodeupModelBuilder = &BootstrapClientBuilder{}
