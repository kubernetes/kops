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

package model

import (
	"fmt"
	"strings"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack/openstackcloudconfig"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

const (
	CloudConfigFilePath = "/etc/kubernetes/cloud.config"
)

// CloudConfigBuilder creates the cloud configuration file
type CloudConfigBuilder struct {
	*NodeupModelContext
}

var _ fi.NodeupModelBuilder = &CloudConfigBuilder{}

func (b *CloudConfigBuilder) Build(c *fi.NodeupModelBuilderContext) error {
	// Azure reads its cloud config from the azure-cloud-provider Secret, so
	// nodeup does not write a cloud config file for Azure nodes.
	if b.CloudProvider() == kops.CloudProviderAzure {
		return nil
	}

	if !b.HasAPIServer && b.NodeupConfig.KubeletConfig.CloudProvider == "external" {
		return nil
	}

	return b.build(c)
}

func (b *CloudConfigBuilder) build(c *fi.NodeupModelBuilderContext) error {
	// Add cloud config file if needed
	var lines []string

	cloudProvider := b.CloudProvider()

	switch cloudProvider {
	case kops.CloudProviderGCE:
		if b.NodeupConfig.NodeTags != nil {
			lines = append(lines, "node-tags = "+*b.NodeupConfig.NodeTags)
		}
		if b.NodeupConfig.NodeInstancePrefix != nil {
			lines = append(lines, "node-instance-prefix = "+*b.NodeupConfig.NodeInstancePrefix)
		}
		if b.NodeupConfig.Multizone != nil {
			lines = append(lines, fmt.Sprintf("multizone = %t", *b.NodeupConfig.Multizone))
		}
	case kops.CloudProviderAWS:
		if b.NodeupConfig.DisableSecurityGroupIngress != nil {
			lines = append(lines, fmt.Sprintf("DisableSecurityGroupIngress = %t", *b.NodeupConfig.DisableSecurityGroupIngress))
		}
		if b.NodeupConfig.ElbSecurityGroup != nil {
			lines = append(lines, "ElbSecurityGroup = "+*b.NodeupConfig.ElbSecurityGroup)
		}
		if b.NodeupConfig.NLBSecurityGroupMode != nil {
			lines = append(lines, "NLBSecurityGroupMode = "+*b.NodeupConfig.NLBSecurityGroupMode)
		}
		for _, family := range b.NodeupConfig.NodeIPFamilies {
			lines = append(lines, "NodeIPFamilies = "+family)
		}
	case kops.CloudProviderOpenstack:
		lines = append(lines, openstackcloudconfig.MakeCloudConfig(b.NodeupConfig.Openstack)...)
	}

	config := "[global]\n" + strings.Join(lines, "\n") + "\n"
	path := CloudConfigFilePath
	t := &nodetasks.File{
		Path:     path,
		Contents: fi.NewStringResource(config),
		Type:     nodetasks.FileType_File,
	}
	c.AddTask(t)

	return nil
}
