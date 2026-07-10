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
	"os"
	"path/filepath"
	"regexp"

	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/flagbuilder"
	"k8s.io/kops/pkg/systemd"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/scaleway/scalewaymetadata"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
	"k8s.io/kops/util/pkg/distributions"
	"k8s.io/kops/util/pkg/env"
	"k8s.io/kops/util/pkg/vfs/openstackconfig"
)

// ProtokubeBuilder configures protokube
type ProtokubeBuilder struct {
	*NodeupModelContext
}

var _ fi.NodeupModelBuilder = &ProtokubeBuilder{}

// Build is responsible for generating the options for protokube
func (t *ProtokubeBuilder) Build(c *fi.NodeupModelBuilderContext) error {
	// Skip the cluster doesn't use gossip, or this is a worker that bootstraps via kops-controller.
	// Must match the asset-skip condition in pkg/nodemodel/nodeupconfigbuilder.go.
	if !t.UsesLegacyGossip() || (!t.IsMaster && len(t.BootConfig.APIServerIPs) > 0) {
		klog.V(2).Infof("skipping protokube provisioning")
		return nil
	}

	{
		name, res, err := t.Assets.FindMatch(regexp.MustCompile("protokube$"))
		if err != nil {
			return err
		}

		c.AddTask(&nodetasks.File{
			Path:     filepath.Join("/opt/kops/bin", name),
			Contents: res,
			Type:     nodetasks.FileType_File,
			Mode:     new("0755"),
		})
	}

	envFile, err := t.buildEnvFile()
	if err != nil {
		return err
	}
	c.AddTask(envFile)

	service, err := t.buildSystemdService()
	if err != nil {
		return err
	}
	c.AddTask(service)

	return nil
}

// buildSystemdService generates the manifest for the protokube service
func (t *ProtokubeBuilder) buildSystemdService() (*nodetasks.Service, error) {
	protokubeFlags, err := t.ProtokubeFlags()
	if err != nil {
		return nil, err
	}
	protokubeRunArgs, err := flagbuilder.BuildFlags(protokubeFlags)
	if err != nil {
		return nil, err
	}

	manifest := &systemd.Manifest{}
	manifest.Set("Unit", "Description", "Kubernetes Protokube Service")
	manifest.Set("Unit", "Documentation", "https://kops.sigs.k8s.io")

	manifest.Set("Service", "ExecStart", "/opt/kops/bin/protokube"+" "+protokubeRunArgs)
	manifest.Set("Service", "EnvironmentFile", "/etc/sysconfig/protokube")
	manifest.Set("Service", "Restart", "always")
	manifest.Set("Service", "RestartSec", "3s")
	manifest.Set("Service", "StartLimitInterval", "0")
	manifest.Set("Install", "WantedBy", "multi-user.target")

	manifestString := manifest.Render()
	klog.V(8).Infof("Built service manifest %q\n%s", "protokube", manifestString)

	service := &nodetasks.Service{
		Name:       "protokube.service",
		Definition: s(manifestString),
	}

	service.InitDefaults()

	return service, nil
}

// ProtokubeFlags are the flags for protokube
type ProtokubeFlags struct {
	Cloud         *string `json:"cloud,omitempty" flag:"cloud"`
	Containerized *bool   `json:"containerized,omitempty" flag:"containerized"`
	Gossip        *bool   `json:"gossip,omitempty" flag:"gossip"`
	LogLevel      *int32  `json:"logLevel,omitempty" flag:"v"`

	GossipProtocol *string `json:"gossip-protocol" flag:"gossip-protocol"`
	GossipListen   *string `json:"gossip-listen" flag:"gossip-listen"`
	GossipSecret   *string `json:"gossip-secret" flag:"gossip-secret"`

	GossipProtocolSecondary *string `json:"gossip-protocol-secondary" flag:"gossip-protocol-secondary" flag-include-empty:"true"`
	GossipListenSecondary   *string `json:"gossip-listen-secondary" flag:"gossip-listen-secondary"`
	GossipSecretSecondary   *string `json:"gossip-secret-secondary" flag:"gossip-secret-secondary"`
}

// ProtokubeFlags is responsible for building the command line flags for protokube
func (t *ProtokubeBuilder) ProtokubeFlags() (*ProtokubeFlags, error) {
	f := &ProtokubeFlags{
		Cloud:         new(string(t.CloudProvider())),
		Containerized: new(false),
		LogLevel:      new(int32(4)),
	}

	if t.UsesLegacyGossip() {
		klog.Warningf("using (legacy) gossip DNS")
		f.Gossip = new(true)
		if t.NodeupConfig.GossipConfig != nil {
			f.GossipProtocol = t.NodeupConfig.GossipConfig.Protocol
			f.GossipListen = t.NodeupConfig.GossipConfig.Listen
			f.GossipSecret = t.NodeupConfig.GossipConfig.Secret

			if t.NodeupConfig.GossipConfig.Secondary != nil {
				f.GossipProtocolSecondary = t.NodeupConfig.GossipConfig.Secondary.Protocol
				f.GossipListenSecondary = t.NodeupConfig.GossipConfig.Secondary.Listen
				f.GossipSecretSecondary = t.NodeupConfig.GossipConfig.Secondary.Secret
			}
		}
	}

	return f, nil
}

func (t *ProtokubeBuilder) buildEnvFile() (*nodetasks.File, error) {
	envVars := make(map[string]string)

	// Pass in gossip dns connection limit
	if os.Getenv("GOSSIP_DNS_CONN_LIMIT") != "" {
		envVars["GOSSIP_DNS_CONN_LIMIT"] = os.Getenv("GOSSIP_DNS_CONN_LIMIT")
	}

	// Pass in required credentials when using user-defined s3 endpoint
	if os.Getenv("AWS_REGION") != "" {
		envVars["AWS_REGION"] = os.Getenv("AWS_REGION")
	}

	if os.Getenv("S3_ENDPOINT") != "" {
		envVars["S3_ENDPOINT"] = os.Getenv("S3_ENDPOINT")
		envVars["S3_REGION"] = os.Getenv("S3_REGION")
		envVars["S3_ACCESS_KEY_ID"] = os.Getenv("S3_ACCESS_KEY_ID")
		envVars["S3_SECRET_ACCESS_KEY"] = os.Getenv("S3_SECRET_ACCESS_KEY")
	}

	if os.Getenv("OS_AUTH_URL") != "" {
		for _, envVar := range []string{
			"OS_TENANT_ID", "OS_TENANT_NAME", "OS_PROJECT_ID", "OS_PROJECT_NAME",
			"OS_PROJECT_DOMAIN_NAME", "OS_PROJECT_DOMAIN_ID",
			"OS_DOMAIN_NAME", "OS_DOMAIN_ID",
			"OS_USERNAME",
			"OS_PASSWORD",
			"OS_AUTH_URL",
			"OS_REGION_NAME",
			"OS_APPLICATION_CREDENTIAL_ID",
			"OS_APPLICATION_CREDENTIAL_SECRET",
			openstackconfig.EnvKeyOpenstackTLSInsecureSkipVerify,
		} {
			envVars[envVar] = os.Getenv(envVar)
		}
	}

	if t.CloudProvider() == kops.CloudProviderDO && os.Getenv("DIGITALOCEAN_ACCESS_TOKEN") != "" {
		envVars["DIGITALOCEAN_ACCESS_TOKEN"] = os.Getenv("DIGITALOCEAN_ACCESS_TOKEN")
	}

	if os.Getenv("HCLOUD_TOKEN") != "" {
		envVars["HCLOUD_TOKEN"] = os.Getenv("HCLOUD_TOKEN")
	}

	if os.Getenv("OSS_REGION") != "" {
		envVars["OSS_REGION"] = os.Getenv("OSS_REGION")
	}

	if t.CloudProvider() == kops.CloudProviderScaleway {
		if os.Getenv("SCW_PROFILE") != "" || os.Getenv("SCW_SECRET_KEY") != "" {
			profile, err := scalewaymetadata.CreateValidScalewayProfile()
			if err != nil {
				return nil, err
			}
			envVars["SCW_ACCESS_KEY"] = fi.ValueOf(profile.AccessKey)
			envVars["SCW_SECRET_KEY"] = fi.ValueOf(profile.SecretKey)
			envVars["SCW_DEFAULT_PROJECT_ID"] = fi.ValueOf(profile.DefaultProjectID)
		}
	}

	for _, envVar := range env.GetProxyEnvVars(t.NodeupConfig.Networking.EgressProxy) {
		envVars[envVar.Name] = envVar.Value
	}

	switch t.Distribution {
	case distributions.DistributionFlatcar:
		envVars["PATH"] = fmt.Sprintf("/opt/kops/bin:%v", os.Getenv("PATH"))
	}

	sysconfig := ""
	for key, value := range envVars {
		sysconfig += key + "=" + value + "\n"
	}

	task := &nodetasks.File{
		Path:     "/etc/sysconfig/protokube",
		Contents: fi.NewStringResource(sysconfig),
		Type:     nodetasks.FileType_File,
	}

	return task, nil
}
