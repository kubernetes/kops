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

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

// BootstrapClientBuilder calls kops-controller to bootstrap the node.
type BootstrapClientBuilder struct {
	*NodeupModelContext
}

func (b BootstrapClientBuilder) Build(c *fi.ModelBuilderContext) error {
	if b.IsMaster || !b.UseKopsControllerForNodeBootstrap() {
		return nil
	}

	var authenticator fi.Authenticator
	var err error
	switch kops.CloudProviderID(b.Cluster.Spec.CloudProvider) {
	case kops.CloudProviderAWS:
		authenticator, err = awsup.NewAWSAuthenticator()
	default:
		return fmt.Errorf("unsupported cloud provider %s", b.Cluster.Spec.CloudProvider)
	}
	if err != nil {
		return err
	}

	cert, err := b.GetCert(fi.CertificateIDCA)
	if err != nil {
		return err
	}

	bootstrapClient := &nodetasks.BootstrapClient{
		Authenticator: authenticator,
		CA:            cert,
		Certs:         b.bootstrapCerts,
	}

	for _, cert := range b.bootstrapCerts {
		cert.Cert.Task = bootstrapClient
		cert.Key.Task = bootstrapClient
	}

	c.AddTask(bootstrapClient)
	return nil
}

var _ fi.ModelBuilder = &BootstrapClientBuilder{}
