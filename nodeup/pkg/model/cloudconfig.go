/*
Copyright 2016 The Kubernetes Authors.

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
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
	"strings"
)

const CloudConfigFilePath = "/etc/kubernetes/cloud.config"

// CloudConfigBuilder creates the cloud configuration file
type CloudConfigBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &CloudConfigBuilder{}

func (b *CloudConfigBuilder) Build(c *fi.ModelBuilderContext) error {
	// Add cloud config file if needed
	var lines []string

	cloudProvider := b.Cluster.Spec.CloudProvider
	cloudConfig := b.Cluster.Spec.CloudConfig

	if cloudConfig == nil {
		cloudConfig = &kops.CloudConfiguration{}
	}

	switch cloudProvider {
	case "gce":
		if cloudConfig.NodeTags != nil {
			lines = append(lines, "node-tags = "+*cloudConfig.NodeTags)
		}
		if cloudConfig.NodeInstancePrefix != nil {
			lines = append(lines, "node-instance-prefix = "+*cloudConfig.NodeInstancePrefix)
		}
		if cloudConfig.Multizone != nil {
			lines = append(lines, fmt.Sprintf("multizone = %t", *cloudConfig.Multizone))
		}
	case "aws":
		if cloudConfig.DisableSecurityGroupIngress != nil {
			lines = append(lines, fmt.Sprintf("DisableSecurityGroupIngress = %t", *cloudConfig.DisableSecurityGroupIngress))
		}
	}

	if cloudConfig.VSphereUsername != nil {
		lines = append(lines, "user = "+*cloudConfig.VSphereUsername)
	}

	if cloudConfig.VSpherePassword != nil {
		lines = append(lines, "password = "+*cloudConfig.VSpherePassword)
	}

	if cloudConfig.VSphereServer != nil {
		lines = append(lines, "server = "+*cloudConfig.VSphereServer)
		lines = append(lines, "port = 443")
		lines = append(lines, fmt.Sprintf("insecure-flag = %t", true))
	}

	if cloudConfig.VSphereDatacenter != nil {
		lines = append(lines, "datacenter = "+*cloudConfig.VSphereDatacenter)
	}

	if cloudConfig.VSphereDatastore != nil {
		lines = append(lines, "datastore = "+*cloudConfig.VSphereDatastore)
	}

	config := "[global]\n" + strings.Join(lines, "\n") + "\n"

	// We need this to support Kubernetes vSphere CloudProvider < v1.5.3
	if fi.CloudProviderID(b.Cluster.Spec.CloudProvider) == fi.CloudProviderVSphere {
		config += "[disk]\n" + "scsicontrollertype = pvscsi" + "\n"
	}


	t := &nodetasks.File{
		Path:     CloudConfigFilePath,
		Contents: fi.NewStringResource(config),
		Type:     nodetasks.FileType_File,
	}
	c.AddTask(t)

	return nil
}
