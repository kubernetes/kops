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
	nodeidentitygce "k8s.io/kops/pkg/nodeidentity/gce"
	"k8s.io/kops/pkg/wellknownports"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

// BootstrapClientBuilder calls kops-controller to bootstrap the node.
type BootstrapClientBuilder struct {
	*NodeupModelContext
}

func buildAuthenticator(cluster *kops.Cluster) (fi.Authenticator, error) {
	switch kops.CloudProviderID(cluster.Spec.CloudProvider) {
	case kops.CloudProviderAWS:
		region, err := awsup.FindRegion(cluster)
		if err != nil {
			return nil, fmt.Errorf("error querying AWS region: %w", err)
		}
		return awsup.NewAWSAuthenticator(region)

	case kops.CloudProviderGCE:
		// Doesn't have to match DNS name
		audience := "kops-controller." + cluster.Name
		return nodeidentitygce.NewAuthenticator(audience)

	default:
		return nil, fmt.Errorf("unsupported cloud provider %s", cluster.Spec.CloudProvider)
	}
}

func (b BootstrapClientBuilder) Build(c *fi.ModelBuilderContext) error {
	if b.IsMaster || !b.UseKopsControllerForNodeBootstrap() {
		return nil
	}

	authenticator, err := buildAuthenticator(b.Cluster)
	if err != nil {
		return err
	}

	cert, err := b.GetCert(fi.CertificateIDCA)
	if err != nil {
		return err
	}

	baseURL := url.URL{
		Scheme: "https",
		Host:   net.JoinHostPort("kops-controller.internal."+b.Cluster.ObjectMeta.Name, strconv.Itoa(wellknownports.KopsControllerPort)),
		Path:   "/",
	}

	bootstrapClient := &nodetasks.KopsBootstrapClient{
		Authenticator: authenticator,
		CA:            cert,
		BaseURL:       baseURL,
	}

	bootstrapClientTask := &nodetasks.BootstrapClientTask{
		Client: bootstrapClient,
		Certs:  b.bootstrapCerts,
	}

	for _, cert := range b.bootstrapCerts {
		cert.Cert.Task = bootstrapClientTask
		cert.Key.Task = bootstrapClientTask
	}

	c.AddTask(bootstrapClientTask)
	return nil
}

var _ fi.ModelBuilder = &BootstrapClientBuilder{}
