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
	"bytes"
	"fmt"
	"os"
	"strings"

	kopsbase "k8s.io/kops"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/pkg/dns"
	"k8s.io/kops/pkg/flagbuilder"
	"k8s.io/kops/pkg/systemd"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"

	"github.com/blang/semver"
	"github.com/golang/glog"
)

// ProtokubeBuilder configures protokube
type ProtokubeBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &ProtokubeBuilder{}

// Build is responsible for generating the protokube service
func (b *ProtokubeBuilder) Build(c *fi.ModelBuilderContext) error {
	if b.IsMaster {
		kubeconfig, err := b.buildPKIKubeconfig("kops")
		if err != nil {
			return err
		}

		c.AddTask(&nodetasks.File{
			Path:     "/var/lib/kops/kubeconfig",
			Contents: fi.NewStringResource(kubeconfig),
			Type:     nodetasks.FileType_File,
			Mode:     s("0400"),
		})
	}

	// @check if protokube; we have decided to disable this by default (https://github.com/kubernetes/kops/pull/3091)
	// unless the gossip dns is switched on
	if !b.IsMaster && !dns.IsGossipHostname(b.Cluster.Spec.MasterInternalName) {
		glog.V(2).Infof("skipping the provisioning of protokube on the node, as gossip dns is disabled")
		return nil
	}

	service, err := b.buildSystemdService()
	if err != nil {
		return err
	}
	c.AddTask(service)

	return nil
}

func (b *ProtokubeBuilder) buildSystemdService() (*nodetasks.Service, error) {
	k8sVersion, err := util.ParseKubernetesVersion(b.Cluster.Spec.KubernetesVersion)
	if err != nil || k8sVersion == nil {
		return nil, fmt.Errorf("unable to parse KubernetesVersion %q", b.Cluster.Spec.KubernetesVersion)
	}

	protokubeFlags := b.ProtokubeFlags(*k8sVersion)
	protokubeFlagsArgs, err := flagbuilder.BuildFlags(protokubeFlags)
	if err != nil {
		return nil, err
	}

	dockerArgs := []string{
		"/usr/bin/docker",
		"run",
		"-v", "/:/rootfs/",
		"-v", "/var/run/dbus:/var/run/dbus",
		"-v", "/run/systemd:/run/systemd",
		"--net=host",
		"--privileged",
		"--env", "KUBECONFIG=/rootfs/var/lib/kops/kubeconfig",
		b.ProtokubeEnvironmentVariables(),
		b.ProtokubeImageName(),
		"/usr/bin/protokube",
	}
	protokubeCommand := strings.Join(dockerArgs, " ") + " " + protokubeFlagsArgs

	manifest := &systemd.Manifest{}
	manifest.Set("Unit", "Description", "Kubernetes Protokube Service")
	manifest.Set("Unit", "Documentation", "https://github.com/kubernetes/kops")
	manifest.Set("Service", "ExecStartPre", b.ProtokubeImagePullCommand())
	manifest.Set("Service", "ExecStart", protokubeCommand)
	manifest.Set("Service", "Restart", "always")
	manifest.Set("Service", "RestartSec", "2s")
	manifest.Set("Service", "StartLimitInterval", "0")

	manifest.Set("Install", "WantedBy", "multi-user.target")

	manifestString := manifest.Render()
	glog.V(8).Infof("Built service manifest %q\n%s", "protokube", manifestString)

	service := &nodetasks.Service{
		Name:       "protokube.service",
		Definition: s(manifestString),
	}

	service.InitDefaults()

	return service, nil
}

// ProtokubeImageName returns the docker image for protokube
func (b *ProtokubeBuilder) ProtokubeImageName() string {
	name := ""
	if b.NodeupConfig.ProtokubeImage != nil && b.NodeupConfig.ProtokubeImage.Name != "" {
		name = b.NodeupConfig.ProtokubeImage.Name
	}
	if name == "" {
		// use current default corresponding to this version of nodeup
		name = kopsbase.DefaultProtokubeImageName()
	}
	return name
}

// ProtokubeImagePullCommand returns the command to pull the image
func (b *ProtokubeBuilder) ProtokubeImagePullCommand() string {
	source := ""
	if b.NodeupConfig.ProtokubeImage != nil {
		source = b.NodeupConfig.ProtokubeImage.Source
	}
	if source == "" {
		// Nothing to pull; return dummy value
		return "/bin/true"
	}
	if strings.HasPrefix(source, "http:") || strings.HasPrefix(source, "https:") || strings.HasPrefix(source, "s3:") {
		// We preloaded the image; return a dummy value
		return "/bin/true"
	}
	return "/usr/bin/docker pull " + b.NodeupConfig.ProtokubeImage.Source
}

// ProtokubeFlags is the options passed to the service
type ProtokubeFlags struct {
	ApplyTaints       *bool    `json:"applyTaints,omitempty" flag:"apply-taints"`
	Channels          []string `json:"channels,omitempty" flag:"channels"`
	Cloud             *string  `json:"cloud,omitempty" flag:"cloud"`
	ClusterID         *string  `json:"cluster-id,omitempty" flag:"cluster-id"`
	Containerized     *bool    `json:"containerized,omitempty" flag:"containerized"`
	DNSInternalSuffix *string  `json:"dnsInternalSuffix,omitempty" flag:"dns-internal-suffix"`
	DNSProvider       *string  `json:"dnsProvider,omitempty" flag:"dns"`
	DNSServer         *string  `json:"dns-server,omitempty" flag:"dns-server"`
	InitializeRBAC    *bool    `json:"initializeRBAC,omitempty" flag:"initialize-rbac"`
	LogLevel          *int32   `json:"logLevel,omitempty" flag:"v"`
	Master            *bool    `json:"master,omitempty" flag:"master"`
	Zone              []string `json:"zone,omitempty" flag:"zone"`
}

// ProtokubeFlags returns the flags object for protokube
func (b *ProtokubeBuilder) ProtokubeFlags(k8sVersion semver.Version) *ProtokubeFlags {
	f := &ProtokubeFlags{}

	master := b.IsMaster

	f.Master = fi.Bool(master)
	if master {
		f.Channels = b.NodeupConfig.Channels
	}

	if k8sVersion.Major == 1 && k8sVersion.Minor >= 6 {
		if master {
			f.InitializeRBAC = fi.Bool(true)
		}
	}

	f.LogLevel = fi.Int32(4)
	f.Containerized = fi.Bool(true)

	zone := b.Cluster.Spec.DNSZone
	if zone != "" {
		if strings.Contains(zone, ".") {
			// match by name
			f.Zone = append(f.Zone, zone)
		} else {
			// match by id
			f.Zone = append(f.Zone, "*/"+zone)
		}
	} else {
		glog.Warningf("DNSZone not specified; protokube won't be able to update DNS")
		// TODO: Should we permit wildcard updates if zone is not specified?
		//argv = append(argv, "--zone=*/*")
	}

	if dns.IsGossipHostname(b.Cluster.Spec.MasterInternalName) {
		glog.Warningf("MasterInternalName %q implies gossip DNS", b.Cluster.Spec.MasterInternalName)
		f.DNSProvider = fi.String("gossip")

		/// TODO: This is hacky, but we want it so that we can have a different internal & external name
		internalSuffix := b.Cluster.Spec.MasterInternalName
		internalSuffix = strings.TrimPrefix(internalSuffix, "api.")
		f.DNSInternalSuffix = fi.String(internalSuffix)
	}

	if b.Cluster.Spec.CloudProvider != "" {
		f.Cloud = fi.String(b.Cluster.Spec.CloudProvider)

		if f.DNSProvider == nil {
			switch kops.CloudProviderID(b.Cluster.Spec.CloudProvider) {
			case kops.CloudProviderAWS:
				f.DNSProvider = fi.String("aws-route53")
			case kops.CloudProviderGCE:
				f.DNSProvider = fi.String("google-clouddns")
			case kops.CloudProviderVSphere:
				f.DNSProvider = fi.String("coredns")
				f.ClusterID = fi.String(b.Cluster.ObjectMeta.Name)
				f.DNSServer = fi.String(*b.Cluster.Spec.CloudConfig.VSphereCoreDNSServer)
			default:
				glog.Warningf("Unknown cloudprovider %q; won't set DNS provider", b.Cluster.Spec.CloudProvider)
			}
		}
	}

	if f.DNSInternalSuffix == nil {
		f.DNSInternalSuffix = fi.String(".internal." + b.Cluster.ObjectMeta.Name)
	}

	if k8sVersion.Major == 1 && k8sVersion.Minor <= 5 {
		f.ApplyTaints = fi.Bool(true)
	}

	return f
}

// ProtokubeEnvironmentVariables generates the environment variable for protokube
func (b *ProtokubeBuilder) ProtokubeEnvironmentVariables() string {
	var buffer bytes.Buffer

	if os.Getenv("AWS_REGION") != "" {
		buffer.WriteString(" ")
		buffer.WriteString("-e 'AWS_REGION=")
		buffer.WriteString(os.Getenv("AWS_REGION"))
		buffer.WriteString("'")
		buffer.WriteString(" ")
	}

	// Pass in required credentials when using user-defined s3 endpoint
	if os.Getenv("S3_ENDPOINT") != "" {
		buffer.WriteString(" ")
		buffer.WriteString("-e S3_ENDPOINT=")
		buffer.WriteString("'")
		buffer.WriteString(os.Getenv("S3_ENDPOINT"))
		buffer.WriteString("'")
		buffer.WriteString(" -e S3_REGION=")
		buffer.WriteString("'")
		buffer.WriteString(os.Getenv("S3_REGION"))
		buffer.WriteString("'")
		buffer.WriteString(" -e S3_ACCESS_KEY_ID=")
		buffer.WriteString("'")
		buffer.WriteString(os.Getenv("S3_ACCESS_KEY_ID"))
		buffer.WriteString("'")
		buffer.WriteString(" -e S3_SECRET_ACCESS_KEY=")
		buffer.WriteString("'")
		buffer.WriteString(os.Getenv("S3_SECRET_ACCESS_KEY"))
		buffer.WriteString("'")
		buffer.WriteString(" ")
	}

	return buffer.String()
}
