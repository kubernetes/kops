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
	"bufio"
	"fmt"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
	"os"
	"strings"
)

const CloudConfigFilePath = "/etc/kubernetes/cloud.config"

// Required for vSphere CloudProvider
const MinimumVersionForVMUUID = "1.5.3"

// VM UUID is set by cloud-init
const VM_UUID_FILE_PATH = "/etc/vmware/vm_uuid"

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
		if cloudConfig.ElbSecurityGroup != nil {
			lines = append(lines, "ElbSecurityGroup = "+*cloudConfig.ElbSecurityGroup)
		}
	case "vsphere":
		vm_uuid, err := getVMUUID(b.Cluster.Spec.KubernetesVersion)
		if err != nil {
			return err
		}
		// Note: Segregate configuration for different sections as below
		// Global Config for vSphere CloudProvider
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
		if vm_uuid != "" {
			lines = append(lines, "vm-uuid = "+strings.Trim(vm_uuid, "\n"))
		}
		// Disk Config for vSphere CloudProvider
		// We need this to support Kubernetes vSphere CloudProvider < v1.5.3
		lines = append(lines, "[disk]")
		lines = append(lines, "scsicontrollertype = pvscsi")
	}

	config := "[global]\n" + strings.Join(lines, "\n") + "\n"

	t := &nodetasks.File{
		Path:     CloudConfigFilePath,
		Contents: fi.NewStringResource(config),
		Type:     nodetasks.FileType_File,
	}
	c.AddTask(t)

	return nil
}

// We need this for vSphere CloudProvider
// getVMUUID gets instance uuid of the VM from the file written by cloud-init
func getVMUUID(kubernetesVersion string) (string, error) {

	actualKubernetesVersion, err := util.ParseKubernetesVersion(kubernetesVersion)
	if err != nil {
		return "", err
	}
	minimumVersionForUUID, err := util.ParseKubernetesVersion(MinimumVersionForVMUUID)
	if err != nil {
		return "", err
	}

	// VM UUID is required only for Kubernetes version greater than 1.5.3
	if actualKubernetesVersion.GTE(*minimumVersionForUUID) {
		file, err := os.Open(VM_UUID_FILE_PATH)
		defer file.Close()
		if err != nil {
			return "", err
		}
		vm_uuid, err := bufio.NewReader(file).ReadString('\n')
		if err != nil {
			return "", err
		}
		return vm_uuid, err
	}

	return "", err
}
