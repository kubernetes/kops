/*
Copyright 2021 The Kubernetes Authors.

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
	"strings"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

// DiscoveryService registers with the discovery service.
type DiscoveryService struct {
	*NodeupModelContext
}

var _ fi.NodeupModelBuilder = &DiscoveryService{}

// Build is responsible for configuring the discovery service registration tasks.
func (b *DiscoveryService) Build(c *fi.NodeupModelBuilderContext) error {
	if !b.IsMaster {
		return nil
	}
	discoveryServiceOptions := b.DiscoveryServiceOptions()
	if discoveryServiceOptions == nil {
		return nil
	}

	nodeName, err := b.NodeName()
	if err != nil {
		return err
	}

	certificateName := nodeName + "." + b.ClusterName()

	namespace := strings.ReplaceAll(b.ClusterName(), ".", "-")
	id := types.NamespacedName{
		Namespace: namespace,
		Name:      nodeName,
	}

	issueCert := &nodetasks.IssueCert{
		Name:      "discovery-service-client",
		Signer:    fi.DiscoveryCAID,
		KeypairID: b.NodeupConfig.KeypairIDs[fi.DiscoveryCAID],
		Type:      "client",
		Subject: nodetasks.PKIXName{
			CommonName: certificateName,
		},
		AlternateNames: []string{certificateName},
	}
	c.AddTask(issueCert)

	certResource, keyResource, caResource := issueCert.GetResources()

	registerTask := &nodetasks.DiscoveryServiceRegisterTask{
		Name:              "register",
		DiscoveryService:  discoveryServiceOptions.URL,
		ClientCert:        certResource,
		ClientKey:         keyResource,
		ClientCA:          caResource,
		RegisterName:      id.Name,
		RegisterNamespace: id.Namespace,
	}
	c.AddTask(registerTask)

	return nil
}
