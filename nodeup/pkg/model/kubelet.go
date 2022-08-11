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
	"bytes"
	"fmt"
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/flagbuilder"
	"k8s.io/kops/pkg/model/components"
	"k8s.io/kops/pkg/nodelabels"
	"k8s.io/kops/pkg/rbac"
	"k8s.io/kops/pkg/systemd"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
	"k8s.io/kops/util/pkg/distributions"
	kubelet "k8s.io/kubelet/config/v1beta1"
)

const (
	// containerizedMounterHome is the path where we install the containerized mounter (on ContainerOS)
	containerizedMounterHome = "/home/kubernetes/containerized_mounter"

	// kubeletService is the name of the kubelet service
	kubeletService = "kubelet.service"

	kubeletConfigFilePath = "/var/lib/kubelet/kubelet.conf"
)

// KubeletBuilder installs kubelet
type KubeletBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &KubeletBuilder{}

// Build is responsible for building the kubelet configuration
func (b *KubeletBuilder) Build(c *fi.ModelBuilderContext) error {
	err := b.buildKubeletServingCertificate(c)
	if err != nil {
		return fmt.Errorf("error building kubelet server cert: %v", err)
	}

	kubeletConfig, err := b.buildKubeletConfigSpec()
	if err != nil {
		return fmt.Errorf("error building kubelet config: %v", err)
	}

	{
		t, err := buildKubeletComponentConfig(kubeletConfig)
		if err != nil {
			return err
		}

		c.AddTask(t)
	}

	{
		t, err := b.buildSystemdEnvironmentFile(kubeletConfig)
		if err != nil {
			return err
		}
		c.AddTask(t)
	}

	{
		// @TODO Extract to common function?
		assetName := "kubelet"
		assetPath := ""
		// @TODO make Find call to an interface, we cannot mock out this function because it finds a file on disk
		asset, err := b.Assets.Find(assetName, assetPath)
		if err != nil {
			return fmt.Errorf("error trying to locate asset %q: %v", assetName, err)
		}
		if asset == nil {
			return fmt.Errorf("unable to locate asset %q", assetName)
		}

		c.AddTask(&nodetasks.File{
			Path:     b.kubeletPath(),
			Contents: asset,
			Type:     nodetasks.FileType_File,
			Mode:     s("0755"),
		})
	}
	{
		if kubeletConfig.PodManifestPath != "" {
			t, err := b.buildManifestDirectory(kubeletConfig)
			if err != nil {
				return err
			}
			err = c.EnsureTask(t)
			if err != nil {
				return err
			}
		}
	}
	{
		// We always create the directory, avoids circular dependency on a bind-mount
		c.EnsureTask(&nodetasks.File{
			Path: filepath.Dir(b.KubeletKubeConfig()), // e.g. "/var/lib/kubelet"
			Type: nodetasks.FileType_Directory,
			Mode: s("0755"),
		})

		if b.HasAPIServer || !b.UseBootstrapTokens() {
			var kubeconfig fi.Resource
			if b.HasAPIServer {
				kubeconfig, err = b.buildMasterKubeletKubeconfig(c)
			} else {
				kubeconfig, err = b.BuildBootstrapKubeconfig("kubelet", c)
			}
			if err != nil {
				return err
			}

			c.AddTask(&nodetasks.File{
				Path:           b.KubeletKubeConfig(),
				Contents:       kubeconfig,
				Type:           nodetasks.FileType_File,
				Mode:           s("0400"),
				BeforeServices: []string{kubeletService},
			})
		}
	}

	if components.UsesCNI(b.Cluster.Spec.Networking) {
		c.AddTask(&nodetasks.File{
			Path: b.CNIConfDir(),
			Type: nodetasks.FileType_Directory,
		})
	}

	if err := b.addContainerizedMounter(c); err != nil {
		return err
	}

	if kubeletConfig.CgroupDriver == "systemd" && b.Cluster.Spec.ContainerRuntime == "containerd" {

		{
			cgroup := kubeletConfig.KubeletCgroups
			if cgroup != "" {
				c.EnsureTask(b.buildCgroupService(cgroup))
			}

		}
		{
			cgroup := kubeletConfig.RuntimeCgroups
			if cgroup != "" {
				c.EnsureTask(b.buildCgroupService(cgroup))
			}

		}
		/* Kubelet incorrectly interprets this value when CgroupDriver is systemd
		See https://github.com/kubernetes/kubernetes/issues/101189
		{
			cgroup := kubeletConfig.KubeReservedCgroup
			if cgroup != "" {
				c.EnsureTask(b.buildCgroupService(cgroup))
			}
		}
		*/

		{
			cgroup := kubeletConfig.SystemCgroups
			if cgroup != "" {
				c.EnsureTask(b.buildCgroupService(cgroup))
			}
		}

		/* This suffers from the same issue as KubeReservedCgroup
		{
			cgroup := kubeletConfig.SystemReservedCgroup
			if cgroup != "" {
				c.EnsureTask(b.buildCgroupService(cgroup))
			}
		}
		*/
	}

	c.AddTask(b.buildSystemdService())

	return nil
}

func buildKubeletComponentConfig(kubeletConfig *kops.KubeletConfigSpec) (*nodetasks.File, error) {
	componentConfig := kubelet.KubeletConfiguration{}
	if kubeletConfig.ShutdownGracePeriod != nil {
		componentConfig.ShutdownGracePeriod = *kubeletConfig.ShutdownGracePeriod
	}
	if kubeletConfig.ShutdownGracePeriodCriticalPods != nil {
		componentConfig.ShutdownGracePeriodCriticalPods = *kubeletConfig.ShutdownGracePeriodCriticalPods
	}

	s := runtime.NewScheme()
	if err := kubelet.AddToScheme(s); err != nil {
		return nil, err
	}

	gv := kubelet.SchemeGroupVersion
	codecFactory := serializer.NewCodecFactory(s)
	info, ok := runtime.SerializerInfoForMediaType(codecFactory.SupportedMediaTypes(), "application/yaml")
	if !ok {
		return nil, fmt.Errorf("failed to find serializer")
	}
	encoder := codecFactory.EncoderForVersion(info.Serializer, gv)
	var w bytes.Buffer
	if err := encoder.Encode(&componentConfig, &w); err != nil {
		return nil, err
	}

	t := &nodetasks.File{
		Path:           "/var/lib/kubelet/kubelet.conf",
		Contents:       fi.NewBytesResource(w.Bytes()),
		Type:           nodetasks.FileType_File,
		BeforeServices: []string{kubeletService},
	}

	return t, nil
}

// kubeletPath returns the path of the kubelet based on distro
func (b *KubeletBuilder) kubeletPath() string {
	kubeletCommand := "/usr/local/bin/kubelet"
	if b.Distribution == distributions.DistributionFlatcar {
		kubeletCommand = "/opt/kubernetes/bin/kubelet"
	}
	if b.Distribution == distributions.DistributionContainerOS {
		kubeletCommand = "/home/kubernetes/bin/kubelet"
	}
	return kubeletCommand
}

// buildManifestDirectory creates the directory where kubelet expects static manifests to reside
func (b *KubeletBuilder) buildManifestDirectory(kubeletConfig *kops.KubeletConfigSpec) (*nodetasks.File, error) {
	if kubeletConfig.PodManifestPath == "" {
		return nil, fmt.Errorf("failed to build manifest path. Path was empty")
	}
	directory := &nodetasks.File{
		Path: kubeletConfig.PodManifestPath,
		Type: nodetasks.FileType_Directory,
		Mode: s("0755"),
	}
	return directory, nil
}

// buildSystemdEnvironmentFile renders the environment file for the kubelet
func (b *KubeletBuilder) buildSystemdEnvironmentFile(kubeletConfig *kops.KubeletConfigSpec) (*nodetasks.File, error) {
	// @step: ensure the masters do not get a bootstrap configuration
	if b.UseBootstrapTokens() && b.IsMaster {
		kubeletConfig.BootstrapKubeconfig = ""
	}

	// TODO: Dump the separate file for flags - just complexity!
	flags, err := flagbuilder.BuildFlags(kubeletConfig)
	if err != nil {
		return nil, fmt.Errorf("error building kubelet flags: %v", err)
	}

	// Add cloud config file if needed
	// We build this flag differently because it depends on CloudConfig, and to expose it directly
	// would be a degree of freedom we don't have (we'd have to write the config to different files)
	// We can always add this later if it is needed.
	if b.Cluster.Spec.CloudConfig != nil {
		flags += " --cloud-config=" + InTreeCloudConfigFilePath
	}

	if b.UsesSecondaryIP() {
		localIP, err := b.GetMetadataLocalIP()
		if err != nil {
			return nil, err
		}
		if localIP != "" {
			flags += " --node-ip=" + localIP
		}
	}

	if b.usesContainerizedMounter() {
		// We don't want to expose this in the model while it is experimental, but it is needed on COS
		flags += " --experimental-mounter-path=" + path.Join(containerizedMounterHome, "mounter")
	}

	// Add container runtime spcific flags
	switch b.Cluster.Spec.ContainerRuntime {
	case "docker":
		if b.IsKubernetesLT("1.24") {
			flags += " --container-runtime=docker"
			flags += " --cni-bin-dir=" + b.CNIBinDir()
			flags += " --cni-conf-dir=" + b.CNIConfDir()
		}
	case "containerd":
		if b.IsKubernetesLT("1.24") {
			flags += " --container-runtime=remote"
		}
		flags += " --runtime-request-timeout=15m"
		if b.Cluster.Spec.Containerd == nil || b.Cluster.Spec.Containerd.Address == nil {
			flags += " --container-runtime-endpoint=unix:///run/containerd/containerd.sock"
		} else {
			flags += " --container-runtime-endpoint=unix://" + fi.StringValue(b.Cluster.Spec.Containerd.Address)
		}
	}

	if b.UseKopsControllerForNodeBootstrap() {
		flags += " --tls-cert-file=" + b.PathSrvKubernetes() + "/kubelet-server.crt"
		flags += " --tls-private-key-file=" + b.PathSrvKubernetes() + "/kubelet-server.key"
	}

	if b.Cluster.Spec.IsIPv6Only() {
		flags += " --node-ip=::"
	}

	flags += " --config=" + kubeletConfigFilePath

	sysconfig := "DAEMON_ARGS=\"" + flags + "\"\n"
	// Makes kubelet read /root/.docker/config.json properly
	sysconfig = sysconfig + "HOME=\"/root" + "\"\n"

	t := &nodetasks.File{
		Path:     "/etc/sysconfig/kubelet",
		Contents: fi.NewStringResource(sysconfig),
		Type:     nodetasks.FileType_File,
	}

	return t, nil
}

// buildSystemdService is responsible for generating the kubelet systemd unit
func (b *KubeletBuilder) buildSystemdService() *nodetasks.Service {
	kubeletCommand := b.kubeletPath()

	manifest := &systemd.Manifest{}
	manifest.Set("Unit", "Description", "Kubernetes Kubelet Server")
	manifest.Set("Unit", "Documentation", "https://github.com/kubernetes/kubernetes")
	switch b.Cluster.Spec.ContainerRuntime {
	case "docker":
		manifest.Set("Unit", "After", "docker.service")
	case "containerd":
		manifest.Set("Unit", "After", "containerd.service")
	default:
		klog.Warningf("unknown container runtime %q", b.Cluster.Spec.ContainerRuntime)
	}

	manifest.Set("Service", "EnvironmentFile", "/etc/sysconfig/kubelet")

	// @check if we are using bootstrap tokens and file checker
	if !b.IsMaster && b.UseBootstrapTokens() {
		manifest.Set("Service", "ExecStartPre",
			fmt.Sprintf("/bin/bash -c 'while [ ! -f %s ]; do sleep 5; done;'", b.KubeletBootstrapKubeconfig()))
	}

	manifest.Set("Service", "ExecStart", kubeletCommand+" \"$DAEMON_ARGS\"")
	manifest.Set("Service", "Restart", "always")
	manifest.Set("Service", "RestartSec", "2s")
	manifest.Set("Service", "StartLimitInterval", "0")
	manifest.Set("Service", "KillMode", "process")
	manifest.Set("Service", "User", "root")
	manifest.Set("Service", "CPUAccounting", "true")
	manifest.Set("Service", "MemoryAccounting", "true")

	manifest.Set("Install", "WantedBy", "multi-user.target")

	if b.Cluster.Spec.Kubelet.CgroupDriver == "systemd" && b.Cluster.Spec.ContainerRuntime == "containerd" {
		cgroup := b.Cluster.Spec.Kubelet.KubeletCgroups
		if cgroup != "" {
			manifest.Set("Service", "Slice", strings.Trim(cgroup, "/")+".slice")
		}
	}

	manifestString := manifest.Render()

	klog.V(8).Infof("Built service manifest %q\n%s", "kubelet", manifestString)

	service := &nodetasks.Service{
		Name:       kubeletService,
		Definition: s(manifestString),
	}

	service.InitDefaults()

	if b.ConfigurationMode == "Warming" {
		service.Running = fi.Bool(false)
	}

	return service
}

// usesContainerizedMounter returns true if we use the containerized mounter
func (b *KubeletBuilder) usesContainerizedMounter() bool {
	switch b.Distribution {
	case distributions.DistributionContainerOS:
		return true
	default:
		return false
	}
}

// addContainerizedMounter downloads and installs the containerized mounter, that we need on ContainerOS
func (b *KubeletBuilder) addContainerizedMounter(c *fi.ModelBuilderContext) error {
	if !b.usesContainerizedMounter() {
		return nil
	}

	// This is not a race because /etc is ephemeral on COS, and we start kubelet (also in /etc on COS)

	// So what we do here is we download a tarred container image, expand it to containerizedMounterHome, then
	// set up bind mounts so that the script is executable (most of containeros is noexec),
	// and set up some bind mounts of proc and dev so that mounting can take place inside that container
	// - it isn't a full docker container.

	{
		// @TODO Extract to common function?
		assetName := "mounter"
		assetPath := ""
		asset, err := b.Assets.Find(assetName, assetPath)
		if err != nil {
			return fmt.Errorf("error trying to locate asset %q: %v", assetName, err)
		}
		if asset == nil {
			return fmt.Errorf("unable to locate asset %q", assetName)
		}

		t := &nodetasks.File{
			Path:     path.Join(containerizedMounterHome, "mounter"),
			Contents: asset,
			Type:     nodetasks.FileType_File,
			Mode:     s("0755"),
		}
		c.AddTask(t)
	}

	c.AddTask(&nodetasks.File{
		Path: containerizedMounterHome,
		Type: nodetasks.FileType_Directory,
	})

	// TODO: leverage assets for this tar file (but we want to avoid expansion of the archive)
	c.AddTask(&nodetasks.Archive{
		Name:      "containerized_mounter",
		Source:    "https://storage.googleapis.com/kubernetes-release/gci-mounter/mounter.tar",
		Hash:      "6a9f5f52e0b066183e6b90a3820b8c2c660d30f6ac7aeafb5064355bf0a5b6dd",
		TargetDir: path.Join(containerizedMounterHome, "rootfs"),
	})

	c.AddTask(&nodetasks.File{
		Path: path.Join(containerizedMounterHome, "rootfs/var/lib/kubelet"),
		Type: nodetasks.FileType_Directory,
	})

	c.AddTask(&nodetasks.BindMount{
		Source:     containerizedMounterHome,
		Mountpoint: containerizedMounterHome,
		Options:    []string{"exec"},
	})

	c.AddTask(&nodetasks.BindMount{
		Source:     "/var/lib/kubelet/",
		Mountpoint: path.Join(containerizedMounterHome, "rootfs/var/lib/kubelet"),
		Options:    []string{"rshared"},
		Recursive:  true,
	})

	c.AddTask(&nodetasks.BindMount{
		Source:     "/proc",
		Mountpoint: path.Join(containerizedMounterHome, "rootfs/proc"),
		Options:    []string{"ro"},
	})

	c.AddTask(&nodetasks.BindMount{
		Source:     "/dev",
		Mountpoint: path.Join(containerizedMounterHome, "rootfs/dev"),
		Options:    []string{"ro"},
	})

	// kube-up does a file cp, but we probably want to make changes visible (e.g. for gossip DNS)
	c.AddTask(&nodetasks.BindMount{
		Source:     "/etc/resolv.conf",
		Mountpoint: path.Join(containerizedMounterHome, "rootfs/etc/resolv.conf"),
		Options:    []string{"ro"},
	})

	return nil
}

// NodeLabels are defined in the InstanceGroup, but set flags on the kubelet config.
// We have a conflict here: on the one hand we want an easy to use abstract specification
// for the cluster, on the other hand we don't want two fields that do the same thing.
// So we make the logic for combining a KubeletConfig part of our core logic.
// NodeLabels are set on the instanceGroup.  We might allow specification of them on the kubelet
// config as well, but for now the precedence is not fully specified.
// (Today, NodeLabels on the InstanceGroup are merged in to NodeLabels on the KubeletConfig in the Cluster).
// In future, we will likely deprecate KubeletConfig in the Cluster, and move it into componentconfig,
// once that is part of core k8s.

// buildKubeletConfigSpec returns the kubeletconfig for the specified instanceGroup
func (b *KubeletBuilder) buildKubeletConfigSpec() (*kops.KubeletConfigSpec, error) {
	isMaster := b.IsMaster
	isAPIServer := b.BootConfig.InstanceGroupRole == kops.InstanceGroupRoleAPIServer

	// Merge KubeletConfig for NodeLabels
	c := b.NodeupConfig.KubeletConfig

	c.ClientCAFile = filepath.Join(b.PathSrvKubernetes(), "ca.crt")

	if isMaster {
		c.BootstrapKubeconfig = ""
	}

	if b.Cluster.Spec.Networking != nil && b.Cluster.Spec.Networking.AmazonVPC != nil {
		sess := session.Must(session.NewSession())
		metadata := ec2metadata.New(sess)

		// Get the actual instance type by querying the EC2 instance metadata service.
		instanceTypeName, err := metadata.GetMetadata("instance-type")
		if err != nil {
			// Otherwise, fall back to the Instance Group spec.
			instanceTypeName = *b.NodeupConfig.DefaultMachineType
		}

		awsCloud := b.Cloud.(awsup.AWSCloud)
		// Get the instance type's detailed information.
		instanceType, err := awsup.GetMachineTypeInfo(awsCloud, instanceTypeName)
		if err != nil {
			return nil, err
		}

		// Respect any MaxPods value the user sets explicitly.
		if c.MaxPods == nil {
			// Default maximum pods per node defined by KubeletConfiguration
			maxPods := 110

			// AWS VPC CNI plugin-specific maximum pod calculation based on:
			// https://github.com/aws/amazon-vpc-cni-k8s/blob/v1.9.3/README.md#setup
			enis := instanceType.InstanceENIs
			ips := instanceType.InstanceIPsPerENI
			if enis > 0 && ips > 0 {
				instanceMaxPods := enis*(ips-1) + 2
				if instanceMaxPods < maxPods {
					maxPods = instanceMaxPods
				}
			}

			// Write back values that could have changed
			c.MaxPods = fi.Int32(int32(maxPods))
		}
	}

	// Use --register-with-taints
	{
		if len(c.Taints) == 0 && isMaster {
			// (Even though the value is empty, we still expect <Key>=<Value>:<Effect>)
			if b.IsKubernetesLT("1.24") {
				c.Taints = append(c.Taints, nodelabels.RoleLabelMaster16+"=:"+string(v1.TaintEffectNoSchedule))
			} else {
				c.Taints = append(c.Taints, nodelabels.RoleLabelControlPlane20+"=:"+string(v1.TaintEffectNoSchedule))
			}
		}
		if len(c.Taints) == 0 && isAPIServer {
			// (Even though the value is empty, we still expect <Key>=<Value>:<Effect>)
			c.Taints = append(c.Taints, nodelabels.RoleLabelAPIServer16+"=:"+string(v1.TaintEffectNoSchedule))
		}

		// Enable scheduling since it can be controlled via taints.
		c.RegisterSchedulable = fi.Bool(true)
	}

	if c.VolumePluginDirectory == "" {
		switch b.Distribution {
		case distributions.DistributionContainerOS:
			// Default is different on ContainerOS, see https://github.com/kubernetes/kubernetes/pull/58171
			c.VolumePluginDirectory = "/home/kubernetes/flexvolume/"

		case distributions.DistributionFlatcar:
			// The /usr directory is read-only for Flatcar
			c.VolumePluginDirectory = "/var/lib/kubelet/volumeplugins/"

		default:
			c.VolumePluginDirectory = "/usr/libexec/kubernetes/kubelet-plugins/volume/exec/"
		}
	}

	// In certain configurations systemd-resolved will put the loopback address 127.0.0.53 as a nameserver into /etc/resolv.conf
	// https://github.com/coredns/coredns/blob/master/plugin/loop/README.md#troubleshooting-loops-in-kubernetes-clusters
	if c.ResolverConfig == nil {
		if b.Distribution.HasLoopbackEtcResolvConf() {
			c.ResolverConfig = s("/run/systemd/resolve/resolv.conf")
		}
	}

	// As of 1.16 we can no longer set critical labels.
	// kops-controller will set these labels.
	// For bootstrapping reasons, protokube sets the critical labels for kops-controller to run.
	c.NodeLabels = nil

	if c.AuthorizationMode == "" {
		c.AuthorizationMode = "Webhook"
	}

	if c.AuthenticationTokenWebhook == nil {
		c.AuthenticationTokenWebhook = fi.Bool(true)
	}

	return &c, nil
}

// buildMasterKubeletKubeconfig builds a kubeconfig for the master kubelet, self-signing the kubelet cert
func (b *KubeletBuilder) buildMasterKubeletKubeconfig(c *fi.ModelBuilderContext) (fi.Resource, error) {
	nodeName, err := b.NodeName()
	if err != nil {
		return nil, fmt.Errorf("error getting NodeName: %v", err)
	}
	certName := nodetasks.PKIXName{
		CommonName:   fmt.Sprintf("system:node:%s", nodeName),
		Organization: []string{rbac.NodesGroup},
	}

	return b.BuildIssuedKubeconfig("kubelet", certName, c), nil
}

func (b *KubeletBuilder) buildKubeletServingCertificate(c *fi.ModelBuilderContext) error {
	if b.UseKopsControllerForNodeBootstrap() {
		name := "kubelet-server"
		dir := b.PathSrvKubernetes()

		names, err := b.kubeletNames()
		if err != nil {
			return err
		}

		if !b.HasAPIServer {
			cert, key, err := b.GetBootstrapCert(name, fi.CertificateIDCA)
			if err != nil {
				return err
			}

			c.AddTask(&nodetasks.File{
				Path:           filepath.Join(dir, name+".crt"),
				Contents:       cert,
				Type:           nodetasks.FileType_File,
				Mode:           fi.String("0644"),
				BeforeServices: []string{"kubelet.service"},
			})

			c.AddTask(&nodetasks.File{
				Path:           filepath.Join(dir, name+".key"),
				Contents:       key,
				Type:           nodetasks.FileType_File,
				Mode:           fi.String("0400"),
				BeforeServices: []string{"kubelet.service"},
			})

		} else {
			issueCert := &nodetasks.IssueCert{
				Name:      name,
				Signer:    fi.CertificateIDCA,
				KeypairID: b.NodeupConfig.KeypairIDs[fi.CertificateIDCA],
				Type:      "server",
				Subject: nodetasks.PKIXName{
					CommonName: names[0],
				},
				AlternateNames: names,
			}
			c.AddTask(issueCert)
			return issueCert.AddFileTasks(c, dir, name, "", nil)
		}
	}
	return nil
}

func (b *KubeletBuilder) kubeletNames() ([]string, error) {
	if b.CloudProvider != kops.CloudProviderAWS {
		name, err := os.Hostname()
		if err != nil {
			return nil, err
		}

		addrs, _ := net.LookupHost(name)

		return append(addrs, name), nil
	}

	cloud := b.Cloud.(awsup.AWSCloud)

	result, err := cloud.EC2().DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds: []*string{&b.InstanceID},
	})
	if err != nil {
		return nil, fmt.Errorf("error describing instances: %v", err)
	}

	useInstanceIDForNodeName := b.Cluster.Spec.ExternalCloudControllerManager != nil && b.IsKubernetesGTE("1.23")
	return awsup.GetInstanceCertificateNames(result, useInstanceIDForNodeName)
}

func (b *KubeletBuilder) buildCgroupService(name string) *nodetasks.Service {
	name = strings.Trim(name, "/")

	manifest := &systemd.Manifest{}
	manifest.Set("Unit", "Documentation", "man:systemd.special(7)")
	manifest.Set("Unit", "Before", "slices.target")
	manifest.Set("Unit", "DefaultDependencies", "no")

	manifestString := manifest.Render()

	service := &nodetasks.Service{
		Name:       name + ".slice",
		Definition: s(manifestString),
	}

	return service
}
