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
	"github.com/blang/semver"
	"github.com/golang/glog"
	"k8s.io/kops"
	"k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/pkg/flagbuilder"
	"k8s.io/kops/pkg/systemd"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
	"strings"
)

// ProtokubeBuilder configures protokube
type ProtokubeBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &ProtokubeBuilder{}

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

	// TODO: Should we run _protokube on the nodes?
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
		b.ProtokubeImageName(),
		"/usr/bin/protokube",
	}
	protokubeCommand := strings.Join(dockerArgs, " ") + " " + protokubeFlagsArgs

	manifest := &systemd.Manifest{}
	manifest.Set("Unit", "Description", "Kubernetes Protokube Service")
	manifest.Set("Unit", "Documentation", "https://github.com/kubernetes/kops")

	//manifest.Set("Service", "EnvironmentFile", "/etc/sysconfig/protokube")
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
func (t *ProtokubeBuilder) ProtokubeImageName() string {
	name := ""
	if t.NodeupConfig.ProtokubeImage != nil && t.NodeupConfig.ProtokubeImage.Name != "" {
		name = t.NodeupConfig.ProtokubeImage.Name
	}
	if name == "" {
		// use current default corresponding to this version of nodeup
		name = kops.DefaultProtokubeImageName()
	}
	return name
}

// ProtokubeImagePullCommand returns the command to pull the image
func (t *ProtokubeBuilder) ProtokubeImagePullCommand() string {
	source := ""
	if t.NodeupConfig.ProtokubeImage != nil {
		source = t.NodeupConfig.ProtokubeImage.Source
	}
	if source == "" {
		// Nothing to pull; return dummy value
		return "/bin/true"
	}
	if strings.HasPrefix(source, "http:") || strings.HasPrefix(source, "https:") || strings.HasPrefix(source, "s3:") {
		// We preloaded the image; return a dummy value
		return "/bin/true"
	}
	return "/usr/bin/docker pull " + t.NodeupConfig.ProtokubeImage.Source
}

type ProtokubeFlags struct {
	Master        *bool  `json:"master,omitempty" flag:"master"`
	Containerized *bool  `json:"containerized,omitempty" flag:"containerized"`
	LogLevel      *int32 `json:"logLevel,omitempty" flag:"v"`

	InitializeRBAC *bool `json:"initializeRBAC,omitempty" flag:"initialize-rbac"`

	DNSProvider *string `json:"dnsProvider,omitempty" flag:"dns"`

	Zone []string `json:"zone,omitempty" flag:"zone"`

	Channels []string `json:"channels,omitempty" flag:"channels"`

	DNSInternalSuffix *string `json:"dnsInternalSuffix,omitempty" flag:"dns-internal-suffix"`
	Cloud             *string `json:"cloud,omitempty" flag:"cloud"`

	ApplyTaints *bool `json:"applyTaints,omitempty" flag:"apply-taints"`
}

// ProtokubeFlags returns the flags object for protokube
func (t *ProtokubeBuilder) ProtokubeFlags(k8sVersion semver.Version) *ProtokubeFlags {
	f := &ProtokubeFlags{}

	master := t.IsMaster

	f.Master = fi.Bool(master)
	if master {
		f.Channels = t.NodeupConfig.Channels
	}

	if k8sVersion.Major == 1 && k8sVersion.Minor >= 6 {
		if master {
			f.InitializeRBAC = fi.Bool(true)
		}
	}

	f.LogLevel = fi.Int32(4)
	f.Containerized = fi.Bool(true)

	zone := t.Cluster.Spec.DNSZone
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

	if t.Cluster.Spec.CloudProvider != "" {
		f.Cloud = fi.String(t.Cluster.Spec.CloudProvider)

		switch fi.CloudProviderID(t.Cluster.Spec.CloudProvider) {
		case fi.CloudProviderAWS:
			f.DNSProvider = fi.String("aws-route53")
		case fi.CloudProviderGCE:
			f.DNSProvider = fi.String("google-clouddns")
		default:
			glog.Warningf("Unknown cloudprovider %q; won't set DNS provider")
		}
	}

	f.DNSInternalSuffix = fi.String(".internal." + t.Cluster.ObjectMeta.Name)

	if k8sVersion.Major == 1 && k8sVersion.Minor <= 5 {
		f.ApplyTaints = fi.Bool(true)
	}

	return f
}
