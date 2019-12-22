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
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"

	v1 "k8s.io/api/core/v1"
	"k8s.io/klog"
	"k8s.io/kops/nodeup/pkg/distros"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/flagbuilder"
	"k8s.io/kops/pkg/nodelabels"
	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/pkg/rbac"
	"k8s.io/kops/pkg/systemd"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
	"k8s.io/kops/util/pkg/reflectutils"
)

const (
	// containerizedMounterHome is the path where we install the containerized mounter (on ContainerOS)
	containerizedMounterHome = "/home/kubernetes/containerized_mounter"
)

// KubeletBuilder installs kubelet
type KubeletBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &KubeletBuilder{}

// Build is responsible for building the kubelet configuration
func (b *KubeletBuilder) Build(c *fi.ModelBuilderContext) error {
	kubeletConfig, err := b.buildKubeletConfig()
	if err != nil {
		return fmt.Errorf("error building kubelet config: %v", err)
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
			c.EnsureTask(t)
		}
	}
	{
		// We always create the directory, avoids circular dependency on a bind-mount
		c.AddTask(&nodetasks.File{
			Path: filepath.Dir(b.KubeletKubeConfig()),
			Type: nodetasks.FileType_Directory,
			Mode: s("0755"),
		})

		// @check if bootstrap tokens are enabled and create the appropreiate certificates
		if b.UseBootstrapTokens() {
			// @check if a master and if so, we bypass the token strapping and instead generate our own kubeconfig
			if b.IsMaster {
				klog.V(3).Info("kubelet bootstrap tokens are enabled and running on a master")

				task, err := b.buildMasterKubeletKubeconfig()
				if err != nil {
					return err
				}
				c.AddTask(task)
			}
		} else {
			kubeconfig, err := b.BuildPKIKubeconfig("kubelet")
			if err != nil {
				return err
			}

			c.AddTask(&nodetasks.File{
				Path:     b.KubeletKubeConfig(),
				Contents: fi.NewStringResource(kubeconfig),
				Type:     nodetasks.FileType_File,
				Mode:     s("0400"),
			})
		}
	}

	if b.UsesCNI() {
		c.AddTask(&nodetasks.File{
			Path: b.CNIConfDir(),
			Type: nodetasks.FileType_Directory,
		})
	}

	if err := b.addStaticUtils(c); err != nil {
		return err
	}

	if err := b.addContainerizedMounter(c); err != nil {
		return err
	}

	c.AddTask(b.buildSystemdService())

	return nil
}

// kubeletPath returns the path of the kubelet based on distro
func (b *KubeletBuilder) kubeletPath() string {
	kubeletCommand := "/usr/local/bin/kubelet"
	if b.Distribution == distros.DistributionCoreOS {
		kubeletCommand = "/opt/kubernetes/bin/kubelet"
	}
	if b.Distribution == distros.DistributionFlatcar {
		kubeletCommand = "/opt/kubernetes/bin/kubelet"
	}
	if b.Distribution == distros.DistributionContainerOS {
		kubeletCommand = "/home/kubernetes/bin/kubelet"
	}
	return kubeletCommand
}

// buildManifestDirectory creates the directory where kubelet expects static manifests to reside
func (b *KubeletBuilder) buildManifestDirectory(kubeletConfig *kops.KubeletConfigSpec) (*nodetasks.File, error) {
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

	if kubeletConfig.ExperimentalAllowedUnsafeSysctls != nil {
		// The ExperimentalAllowedUnsafeSysctls flag was renamed in k/k #63717
		if b.IsKubernetesGTE("1.11") {
			klog.V(1).Info("ExperimentalAllowedUnsafeSysctls was renamed in k8s 1.11+, please use AllowedUnsafeSysctls instead.")
			kubeletConfig.AllowedUnsafeSysctls = append(kubeletConfig.ExperimentalAllowedUnsafeSysctls, kubeletConfig.AllowedUnsafeSysctls...)
			kubeletConfig.ExperimentalAllowedUnsafeSysctls = nil
		}
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
		flags += " --cloud-config=" + CloudConfigFilePath
	}

	if b.UsesCNI() {
		flags += " --cni-bin-dir=" + b.CNIBinDir()
		flags += " --cni-conf-dir=" + b.CNIConfDir()
	}

	if b.UsesSecondaryIP() {
		sess := session.Must(session.NewSession())
		metadata := ec2metadata.New(sess)
		localIpv4, err := metadata.GetMetadata("local-ipv4")
		if err != nil {
			return nil, fmt.Errorf("error fetching the local-ipv4 address from the ec2 meta-data: %v", err)
		}
		flags += " --node-ip=" + localIpv4
	}

	if b.Cluster.Spec.Networking != nil && b.Cluster.Spec.Networking.Kubenet != nil {
		// Kubenet is neither CNI nor not-CNI, so we need to pass it `--cni-bin-dir` also
		if b.IsKubernetesGTE("1.9") {
			// Flag renamed in #53564
			flags += " --cni-bin-dir=" + b.CNIBinDir()
		} else {
			flags += " --network-plugin-dir=" + b.CNIBinDir()
		}
	}

	if b.usesContainerizedMounter() {
		// We don't want to expose this in the model while it is experimental, but it is needed on COS
		flags += " --experimental-mounter-path=" + path.Join(containerizedMounterHome, "mounter")
	}

	// Add container runtime flags
	if b.Cluster.Spec.ContainerRuntime == "containerd" {
		flags += " --container-runtime=remote"
		flags += " --runtime-request-timeout=15m"
		if b.Cluster.Spec.Containerd == nil || b.Cluster.Spec.Containerd.Address == nil {
			flags += " --container-runtime-endpoint=unix:///run/containerd/containerd.sock"
		} else {
			flags += " --container-runtime-endpoint=unix://" + fi.StringValue(b.Cluster.Spec.Containerd.Address)
		}
	}

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
	manifest.Set("Unit", "After", "containerd.service")

	if b.Distribution == distros.DistributionCoreOS {
		// We add /opt/kubernetes/bin for our utilities (socat, conntrack)
		manifest.Set("Service", "Environment", "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/opt/kubernetes/bin")
	}
	if b.Distribution == distros.DistributionFlatcar {
		// We add /opt/kubernetes/bin for our utilities (conntrack)
		manifest.Set("Service", "Environment", "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/opt/kubernetes/bin")
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
	manifestString := manifest.Render()

	klog.V(8).Infof("Built service manifest %q\n%s", "kubelet", manifestString)

	service := &nodetasks.Service{
		Name:       "kubelet.service",
		Definition: s(manifestString),
	}

	// @check if we are a master allow protokube to start kubelet
	if b.IsMaster {
		service.Running = fi.Bool(false)
	}

	service.InitDefaults()

	return service
}

// buildKubeletConfig is responsible for creating the kubelet configuration
func (b *KubeletBuilder) buildKubeletConfig() (*kops.KubeletConfigSpec, error) {
	if b.InstanceGroup == nil {
		klog.Fatalf("InstanceGroup was not set")
	}

	kubeletConfigSpec, err := b.buildKubeletConfigSpec()
	if err != nil {
		return nil, fmt.Errorf("error building kubelet config: %v", err)
	}

	// TODO: Memoize if we reuse this
	return kubeletConfigSpec, nil
}

func (b *KubeletBuilder) addStaticUtils(c *fi.ModelBuilderContext) error {
	if b.Distribution == distros.DistributionCoreOS {
		// CoreOS does not ship with socat or conntrack.  Install our own (statically linked) version
		// TODO: Extract to common function?
		for _, binary := range []string{"socat", "conntrack"} {
			assetName := binary
			assetPath := ""
			asset, err := b.Assets.Find(assetName, assetPath)
			if err != nil {
				return fmt.Errorf("error trying to locate asset %q: %v", assetName, err)
			}
			if asset == nil {
				return fmt.Errorf("unable to locate asset %q", assetName)
			}

			t := &nodetasks.File{
				Path:     "/opt/kubernetes/bin/" + binary,
				Contents: asset,
				Type:     nodetasks.FileType_File,
				Mode:     s("0755"),
			}
			c.AddTask(t)
		}
	}

	if b.Distribution == distros.DistributionFlatcar {
		// Flatcar does not ship with conntrack.  Install our own (statically linked) version
		// TODO: Extract to common function?
		for _, binary := range []string{"conntrack"} {
			assetName := binary
			assetPath := ""
			asset, err := b.Assets.Find(assetName, assetPath)
			if err != nil {
				return fmt.Errorf("error trying to locate asset %q: %v", assetName, err)
			}
			if asset == nil {
				return fmt.Errorf("unable to locate asset %q", assetName)
			}

			t := &nodetasks.File{
				Path:     "/opt/kubernetes/bin/" + binary,
				Contents: asset,
				Type:     nodetasks.FileType_File,
				Mode:     s("0755"),
			}
			c.AddTask(t)
		}
	}

	return nil
}

// usesContainerizedMounter returns true if we use the containerized mounter
func (b *KubeletBuilder) usesContainerizedMounter() bool {
	switch b.Distribution {
	case distros.DistributionContainerOS:
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
		if !b.IsKubernetesGTE("1.9") {
			// legacy name (and stored in kubernetes-manifests.tar.gz)
			assetName = "gci-mounter"
		}
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
		Hash:      "8003b798cf33c7f91320cd6ee5cec4fa22244571",
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

	// Merge KubeletConfig for NodeLabels
	c := &kops.KubeletConfigSpec{}
	if isMaster {
		reflectutils.JsonMergeStruct(c, b.Cluster.Spec.MasterKubelet)
	} else {
		reflectutils.JsonMergeStruct(c, b.Cluster.Spec.Kubelet)
	}

	// check if we are using secure kubelet <-> api settings
	if b.UseSecureKubelet() {
		c.ClientCAFile = filepath.Join(b.PathSrvKubernetes(), "ca.crt")
	}

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
			instanceTypeName = strings.Split(b.InstanceGroup.Spec.MachineType, ",")[0]
		}

		// Get the instance type's detailed information.
		instanceType, err := awsup.GetMachineTypeInfo(instanceTypeName)
		if err != nil {
			return nil, err
		}

		// Default maximum pods per node defined by KubeletConfiguration, but
		// respect any value the user sets explicitly.
		maxPods := int32(110)
		if c.MaxPods != nil {
			maxPods = *c.MaxPods
		}

		// AWS VPC CNI plugin-specific maximum pod calculation based on:
		// https://github.com/aws/amazon-vpc-cni-k8s/blob/f52ad45/README.md
		//
		// Treat the calculated value as a hard max, since networking with the CNI
		// plugin won't work correctly once we exceed that maximum.
		enis := instanceType.InstanceENIs
		ips := instanceType.InstanceIPsPerENI
		if enis > 0 && ips > 0 {
			instanceMaxPods := enis*(ips-1) + 2
			if int32(instanceMaxPods) < maxPods {
				maxPods = int32(instanceMaxPods)
			}
		}

		// Write back values that could have changed
		c.MaxPods = &maxPods
		if b.InstanceGroup.Spec.Kubelet != nil {
			if b.InstanceGroup.Spec.Kubelet.MaxPods == nil {
				b.InstanceGroup.Spec.Kubelet.MaxPods = &maxPods
			}
		}
	}

	if b.InstanceGroup.Spec.Kubelet != nil {
		reflectutils.JsonMergeStruct(c, b.InstanceGroup.Spec.Kubelet)
	}

	// Use --register-with-taints for k8s 1.6 and on
	if b.Cluster.IsKubernetesGTE("1.6") {
		c.Taints = append(c.Taints, b.InstanceGroup.Spec.Taints...)

		if len(c.Taints) == 0 && isMaster {
			// (Even though the value is empty, we still expect <Key>=<Value>:<Effect>)
			c.Taints = append(c.Taints, nodelabels.RoleLabelMaster16+"=:"+string(v1.TaintEffectNoSchedule))
		}

		// Enable scheduling since it can be controlled via taints.
		// For pre-1.6.0 clusters, this is handled by tainter.go
		c.RegisterSchedulable = fi.Bool(true)
	}

	if c.VolumePluginDirectory == "" {
		switch b.Distribution {
		case distros.DistributionContainerOS:
			// Default is different on ContainerOS, see https://github.com/kubernetes/kubernetes/pull/58171
			c.VolumePluginDirectory = "/home/kubernetes/flexvolume/"

		case distros.DistributionCoreOS:
			// The /usr directory is read-only for CoreOS
			c.VolumePluginDirectory = "/var/lib/kubelet/volumeplugins/"

		case distros.DistributionFlatcar:
			// The /usr directory is read-only for Flatcar
			c.VolumePluginDirectory = "/var/lib/kubelet/volumeplugins/"

		default:
			c.VolumePluginDirectory = "/usr/libexec/kubernetes/kubelet-plugins/volume/exec/"
		}
	}

	// As of 1.16 we can no longer set critical labels.
	// kops-controller will set these labels.
	// For bootstrapping reasons, protokube sets the critical labels for kops-controller to run.
	if b.Cluster.IsKubernetesGTE("1.16") {
		c.NodeLabels = nil
	} else {
		nodeLabels, err := nodelabels.BuildNodeLabels(b.Cluster, b.InstanceGroup)
		if err != nil {
			return nil, err
		}
		c.NodeLabels = nodeLabels
	}

	return c, nil
}

// buildMasterKubeletKubeconfig builds a kubeconfig for the master kubelet, self-signing the kubelet cert
func (b *KubeletBuilder) buildMasterKubeletKubeconfig() (*nodetasks.File, error) {
	nodeName, err := b.NodeName()
	if err != nil {
		return nil, fmt.Errorf("error getting NodeName: %v", err)
	}

	caCert, err := b.KeyStore.FindCert(fi.CertificateId_CA)
	if err != nil {
		return nil, fmt.Errorf("error fetching CA certificate from keystore: %v", err)
	}
	if caCert == nil {
		return nil, fmt.Errorf("unable to find CA certificate %q in keystore", fi.CertificateId_CA)
	}

	caKey, err := b.KeyStore.FindPrivateKey(fi.CertificateId_CA)
	if err != nil {
		return nil, fmt.Errorf("error fetching CA certificate from keystore: %v", err)
	}
	if caKey == nil {
		return nil, fmt.Errorf("unable to find CA key %q in keystore", fi.CertificateId_CA)
	}

	privateKey, err := pki.GeneratePrivateKey()
	if err != nil {
		return nil, err
	}

	template := &x509.Certificate{
		BasicConstraintsValid: true,
		IsCA:                  false,
	}

	template.Subject = pkix.Name{
		CommonName:   fmt.Sprintf("system:node:%s", nodeName),
		Organization: []string{rbac.NodesGroup},
	}

	// https://tools.ietf.org/html/rfc5280#section-4.2.1.3
	//
	// Digital signature allows the certificate to be used to verify
	// digital signatures used during TLS negotiation.
	template.KeyUsage = template.KeyUsage | x509.KeyUsageDigitalSignature
	// KeyEncipherment allows the cert/key pair to be used to encrypt
	// keys, including the symmetric keys negotiated during TLS setup
	// and used for data transfer.
	template.KeyUsage = template.KeyUsage | x509.KeyUsageKeyEncipherment
	// ClientAuth allows the cert to be used by a TLS client to
	// authenticate itself to the TLS server.
	template.ExtKeyUsage = append(template.ExtKeyUsage, x509.ExtKeyUsageClientAuth)

	t := time.Now().UnixNano()
	template.SerialNumber = pki.BuildPKISerial(t)

	certificate, err := pki.SignNewCertificate(privateKey, template, caCert.Certificate, caKey)
	if err != nil {
		return nil, fmt.Errorf("error signing certificate for master kubelet: %v", err)
	}

	caBytes, err := caCert.AsBytes()
	if err != nil {
		return nil, fmt.Errorf("failed to get certificate authority data: %s", err)
	}
	certBytes, err := certificate.AsBytes()
	if err != nil {
		return nil, fmt.Errorf("failed to get certificate data: %s", err)
	}
	keyBytes, err := privateKey.AsBytes()
	if err != nil {
		return nil, fmt.Errorf("failed to get private key data: %s", err)
	}

	content, err := b.BuildKubeConfig("kubelet", caBytes, certBytes, keyBytes)
	if err != nil {
		return nil, err
	}

	return &nodetasks.File{
		Path:     b.KubeletKubeConfig(),
		Contents: fi.NewStringResource(content),
		Type:     nodetasks.FileType_File,
		Mode:     s("600"),
	}, nil
}
