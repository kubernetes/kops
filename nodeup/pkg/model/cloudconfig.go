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
	"bufio"
	"fmt"
	"os"
	"strings"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/pkg/try"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

const (
	CloudConfigFilePath = "/etc/kubernetes/cloud.config"

	// Required for vSphere CloudProvider
	MinimumVersionForVMUUID = "1.5.3"

	// VM UUID is set by cloud-init
	VM_UUID_FILE_PATH = "/etc/vmware/vm_uuid"
)

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
		VMUUID, err := getVMUUID(b.Cluster.Spec.KubernetesVersion)
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
		if VMUUID != "" {
			lines = append(lines, "vm-uuid = "+strings.Trim(VMUUID, "\n"))
		}
		// Disk Config for vSphere CloudProvider
		// We need this to support Kubernetes vSphere CloudProvider < v1.5.3
		lines = append(lines, "[disk]")
		lines = append(lines, "scsicontrollertype = pvscsi")
	case "openstack":
		osc := cloudConfig.Openstack
		if osc == nil {
			break
		}
		//Support mapping of older keystone API
		tenantName := os.Getenv("OS_TENANT_NAME")
		if tenantName == "" {
			tenantName = os.Getenv("OS_PROJECT_NAME")
		}
		tenantID := os.Getenv("OS_TENANT_ID")
		if tenantID == "" {
			tenantID = os.Getenv("OS_PROJECT_ID")
		}
		lines = append(lines,
			fmt.Sprintf("auth-url=\"%s\"", os.Getenv("OS_AUTH_URL")),
			fmt.Sprintf("username=\"%s\"", os.Getenv("OS_USERNAME")),
			fmt.Sprintf("password=\"%s\"", os.Getenv("OS_PASSWORD")),
			fmt.Sprintf("region=\"%s\"", os.Getenv("OS_REGION_NAME")),
			fmt.Sprintf("tenant-id=\"%s\"", tenantID),
			fmt.Sprintf("tenant-name=\"%s\"", tenantName),
			fmt.Sprintf("domain-name=\"%s\"", os.Getenv("OS_DOMAIN_NAME")),
			fmt.Sprintf("domain-id=\"%s\"", os.Getenv("OS_DOMAIN_ID")),
			"",
		)

		if lb := osc.Loadbalancer; lb != nil {
			lines = append(lines,
				"[LoadBalancer]",
				fmt.Sprintf("floating-network-id=%s", fi.StringValue(lb.FloatingNetworkID)),
				fmt.Sprintf("lb-method=%s", fi.StringValue(lb.Method)),
				fmt.Sprintf("lb-provider=%s", fi.StringValue(lb.Provider)),
				fmt.Sprintf("use-octavia=%t", fi.BoolValue(lb.UseOctavia)),
				fmt.Sprintf("manage-security-groups=%t", fi.BoolValue(lb.ManageSecGroups)),
				"",
			)

			if monitor := osc.Monitor; monitor != nil {
				lines = append(lines,
					"create-monitor=yes",
					fmt.Sprintf("monitor-delay=%s", fi.StringValue(monitor.Delay)),
					fmt.Sprintf("monitor-timeout=%s", fi.StringValue(monitor.Timeout)),
					fmt.Sprintf("monitor-max-retries=%d", fi.IntValue(monitor.MaxRetries)),
					"",
				)
			}
		}

		if bs := osc.BlockStorage; bs != nil {
			//Block Storage Config
			lines = append(lines,
				"[BlockStorage]",
				fmt.Sprintf("bs-version=%s", fi.StringValue(bs.Version)),
				fmt.Sprintf("ignore-volume-az=%t", fi.BoolValue(bs.IgnoreAZ)),
				"")
		}
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
		if err != nil {
			return "", err
		}

		defer try.CloseFile(file)

		VMUUID, err := bufio.NewReader(file).ReadString('\n')
		if err != nil {
			return "", err
		}
		return VMUUID, err
	}

	return "", err
}
