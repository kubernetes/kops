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

	"k8s.io/kops/upup/pkg/fi"
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

	kopsControllerClient := b.KopsControllerClient
	if kopsControllerClient == nil {
		return fmt.Errorf("KopsControllerClient not set")
	}

	bootstrapClientTask := &nodetasks.BootstrapClientTask{
		Client:     kopsControllerClient,
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
