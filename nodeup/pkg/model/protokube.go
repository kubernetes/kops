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
	"os"
	"path/filepath"
	"strings"

	kopsbase "k8s.io/kops"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/dns"
	"k8s.io/kops/pkg/flagbuilder"
	"k8s.io/kops/pkg/systemd"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
	"k8s.io/kops/util/pkg/proxy"

	"github.com/blang/semver"
	"k8s.io/klog"
)

// ProtokubeBuilder configures protokube
type ProtokubeBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &ProtokubeBuilder{}

// Build is responsible for generating the options for protokube
func (t *ProtokubeBuilder) Build(c *fi.ModelBuilderContext) error {
	useGossip := dns.IsGossipHostname(t.Cluster.Spec.MasterInternalName)

	// check is not a master and we are not using gossip (https://github.com/kubernetes/kops/pull/3091)
	if !t.IsMaster && !useGossip {
		klog.V(2).Infof("skipping the provisioning of protokube on the nodes")
		return nil
	}

	if protokubeImage := t.NodeupConfig.ProtokubeImage; protokubeImage != nil {
		c.AddTask(&nodetasks.LoadImageTask{
			Name:    "protokube",
			Sources: protokubeImage.Sources,
			Hash:    protokubeImage.Hash,
			Runtime: t.Cluster.Spec.ContainerRuntime,
		})
	}

	if t.IsMaster {
		kubeconfig, err := t.BuildPKIKubeconfig("kops")
		if err != nil {
			return err
		}

		c.AddTask(&nodetasks.File{
			Path:     "/var/lib/kops/kubeconfig",
			Contents: fi.NewStringResource(kubeconfig),
			Type:     nodetasks.FileType_File,
			Mode:     s("0400"),
		})

		// retrieve the etcd peer certificates and private keys from the keystore
		if !t.UseEtcdManager() && t.UseEtcdTLS() {
			for _, x := range []string{"etcd", "etcd-peer", "etcd-client"} {
				if err := t.BuildCertificateTask(c, x, fmt.Sprintf("%s.pem", x)); err != nil {
					return err
				}
			}
			for _, x := range []string{"etcd", "etcd-peer", "etcd-client"} {
				if err := t.BuildPrivateKeyTask(c, x, fmt.Sprintf("%s-key.pem", x)); err != nil {
					return err
				}
			}
		}
	}

	service, err := t.buildSystemdService()
	if err != nil {
		return err
	}
	c.AddTask(service)

	return nil
}

// buildSystemdService generates the manifest for the protokube service
func (t *ProtokubeBuilder) buildSystemdService() (*nodetasks.Service, error) {
	k8sVersion, err := util.ParseKubernetesVersion(t.Cluster.Spec.KubernetesVersion)
	if err != nil || k8sVersion == nil {
		return nil, fmt.Errorf("unable to parse KubernetesVersion %q", t.Cluster.Spec.KubernetesVersion)
	}

	protokubeFlags, err := t.ProtokubeFlags(*k8sVersion)
	if err != nil {
		return nil, err
	}
	protokubeRunArgs, err := flagbuilder.BuildFlags(protokubeFlags)
	if err != nil {
		return nil, err
	}

	protokubeImagePullCommand, err := t.ProtokubeImagePullCommand()
	if err != nil {
		return nil, err
	}
	protokubeContainerStopCommand, err := t.ProtokubeContainerStopCommand()
	if err != nil {
		return nil, err
	}
	protokubeContainerRemoveCommand, err := t.ProtokubeContainerRemoveCommand()
	if err != nil {
		return nil, err
	}
	protokubeContainerRunCommand, err := t.ProtokubeContainerRunCommand()
	if err != nil {
		return nil, err
	}

	manifest := &systemd.Manifest{}
	manifest.Set("Unit", "Description", "Kubernetes Protokube Service")
	manifest.Set("Unit", "Documentation", "https://github.com/kubernetes/kops")

	// @step: let need a dependency for any volumes to be mounted first
	manifest.Set("Service", "ExecStartPre", protokubeContainerStopCommand)
	manifest.Set("Service", "ExecStartPre", protokubeContainerRemoveCommand)
	manifest.Set("Service", "ExecStartPre", protokubeImagePullCommand)
	manifest.Set("Service", "ExecStart", protokubeContainerRunCommand+" "+protokubeRunArgs)
	manifest.Set("Service", "Restart", "always")
	manifest.Set("Service", "RestartSec", "2s")
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

// ProtokubeImageName returns the docker image for protokube
func (t *ProtokubeBuilder) ProtokubeImageName() string {
	name := ""
	if t.NodeupConfig.ProtokubeImage != nil && t.NodeupConfig.ProtokubeImage.Name != "" {
		name = t.NodeupConfig.ProtokubeImage.Name
	}
	if name == "" {
		// use current default corresponding to this version of nodeup
		name = kopsbase.DefaultProtokubeImageName()
	}
	return name
}

// ProtokubeImagePullCommand returns the command to pull the image
func (t *ProtokubeBuilder) ProtokubeImagePullCommand() (string, error) {
	var sources []string
	if t.NodeupConfig.ProtokubeImage != nil {
		sources = t.NodeupConfig.ProtokubeImage.Sources
	}
	if len(sources) == 0 {
		// Nothing to pull; return dummy value
		return "/bin/true", nil
	}
	if strings.HasPrefix(sources[0], "http:") || strings.HasPrefix(sources[0], "https:") || strings.HasPrefix(sources[0], "s3:") {
		// We preloaded the image; return a dummy value
		return "/bin/true", nil
	}

	var protokubeImagePullCommand string
	if t.Cluster.Spec.ContainerRuntime == "docker" {
		protokubeImagePullCommand = "-/usr/bin/docker pull " + sources[0]
	} else if t.Cluster.Spec.ContainerRuntime == "containerd" {
		protokubeImagePullCommand = "-/usr/bin/ctr images pull docker.io/" + sources[0]
	} else {
		return "", fmt.Errorf("unable to create protokube image pull command for unsupported runtime %q", t.Cluster.Spec.ContainerRuntime)
	}
	return protokubeImagePullCommand, nil
}

// ProtokubeContainerStopCommand returns the command that stops the Protokube container, before being removed
func (t *ProtokubeBuilder) ProtokubeContainerStopCommand() (string, error) {
	var containerStopCommand string
	if t.Cluster.Spec.ContainerRuntime == "docker" {
		containerStopCommand = "-/usr/bin/docker stop protokube"
	} else if t.Cluster.Spec.ContainerRuntime == "containerd" {
		containerStopCommand = "/bin/true"
	} else {
		return "", fmt.Errorf("unable to create protokube stop command for unsupported runtime %q", t.Cluster.Spec.ContainerRuntime)
	}
	return containerStopCommand, nil
}

// ProtokubeContainerRemoveCommand returns the command that removes the Protokube container
func (t *ProtokubeBuilder) ProtokubeContainerRemoveCommand() (string, error) {
	var containerRemoveCommand string
	if t.Cluster.Spec.ContainerRuntime == "docker" {
		containerRemoveCommand = "-/usr/bin/docker rm protokube"
	} else if t.Cluster.Spec.ContainerRuntime == "containerd" {
		containerRemoveCommand = "-/usr/bin/ctr container rm protokube"
	} else {
		return "", fmt.Errorf("unable to create protokube remove command for unsupported runtime %q", t.Cluster.Spec.ContainerRuntime)
	}
	return containerRemoveCommand, nil
}

// ProtokubeContainerRunCommand returns the command that runs the Protokube container
func (t *ProtokubeBuilder) ProtokubeContainerRunCommand() (string, error) {
	var containerRunArgs []string
	if t.Cluster.Spec.ContainerRuntime == "docker" {
		containerRunArgs = append(containerRunArgs, []string{
			"/usr/bin/docker run",
			"--net=host",
			"--pid=host",   // Needed for mounting in a container (when using systemd mounting?)
			"--privileged", // We execute in the host namespace
			"--volume /:/rootfs/",
			"--volume /var/run/dbus:/var/run/dbus",
			"--volume /run/systemd:/run/systemd",
			"--env KUBECONFIG=/rootfs/var/lib/kops/kubeconfig",
		}...)

		if fi.BoolValue(t.Cluster.Spec.UseHostCertificates) {
			containerRunArgs = append(containerRunArgs, []string{
				"--volume /etc/ssl/certs:/etc/ssl/certs",
			}...)
		}

		// add kubectl only if a master
		// path changes depending on distro, and always mount it on /opt/kops/bin
		// kubectl is downloaded and installed by other tasks
		if t.IsMaster {
			containerRunArgs = append(containerRunArgs, []string{
				"--volume " + t.KubectlPath() + ":/opt/kops/bin:ro",
				"--env PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/opt/kops/bin",
			}...)
		}

		protokubeEnvVars := t.ProtokubeEnvironmentVariables()
		if protokubeEnvVars != "" {
			containerRunArgs = append(containerRunArgs, []string{
				protokubeEnvVars,
			}...)
		}

		containerRunArgs = append(containerRunArgs, []string{
			"--name", "protokube",
			t.ProtokubeImageName(),
			"/usr/bin/protokube",
		}...)

	} else if t.Cluster.Spec.ContainerRuntime == "containerd" {
		containerRunArgs = append(containerRunArgs, []string{
			"/usr/bin/ctr run",
			"--net-host",
			"--with-ns pid:/proc/1/ns/pid",
			"--privileged",
			"--mount type=bind,src=/,dst=/rootfs,options=rbind:rslave",
			"--mount type=bind,src=/var/run/dbus,dst=/var/run/dbus,options=rbind:rprivate",
			"--mount type=bind,src=/run/systemd,dst=/run/systemd,options=rbind:rprivate",
			"--env KUBECONFIG=/rootfs/var/lib/kops/kubeconfig",
		}...)

		if fi.BoolValue(t.Cluster.Spec.UseHostCertificates) {
			containerRunArgs = append(containerRunArgs, []string{
				"--mount type=bind,src=/etc/ssl/certs,dst=/etc/ssl/certs,options=rbind:ro:rprivate",
			}...)
		}

		if t.IsMaster {
			containerRunArgs = append(containerRunArgs, []string{
				"--mount type=bind,src=" + t.KubectlPath() + ",dst=/opt/kops/bin,options=rbind:ro:rprivate",
				"--env PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/opt/kops/bin",
			}...)
		}

		protokubeEnvVars := t.ProtokubeEnvironmentVariables()
		if protokubeEnvVars != "" {
			containerRunArgs = append(containerRunArgs, []string{
				protokubeEnvVars,
			}...)
		}

		containerRunArgs = append(containerRunArgs, []string{
			"docker.io/library/" + t.ProtokubeImageName(),
			"protokube",
			"/usr/bin/protokube",
		}...)
	} else {
		return "", fmt.Errorf("unable to create protokube run command for unsupported runtime %q", t.Cluster.Spec.ContainerRuntime)
	}

	return strings.Join(containerRunArgs, " "), nil
}

// ProtokubeFlags are the flags for protokube
type ProtokubeFlags struct {
	ApplyTaints *bool    `json:"applyTaints,omitempty" flag:"apply-taints"`
	Channels    []string `json:"channels,omitempty" flag:"channels"`
	Cloud       *string  `json:"cloud,omitempty" flag:"cloud"`
	// ClusterID flag is required only for vSphere cloud type, to pass cluster id information to protokube. AWS and GCE workflows ignore this flag.
	ClusterID                 *string  `json:"cluster-id,omitempty" flag:"cluster-id"`
	Containerized             *bool    `json:"containerized,omitempty" flag:"containerized"`
	DNSInternalSuffix         *string  `json:"dnsInternalSuffix,omitempty" flag:"dns-internal-suffix"`
	DNSProvider               *string  `json:"dnsProvider,omitempty" flag:"dns"`
	DNSServer                 *string  `json:"dns-server,omitempty" flag:"dns-server"`
	EtcdBackupImage           string   `json:"etcd-backup-image,omitempty" flag:"etcd-backup-image"`
	EtcdBackupStore           string   `json:"etcd-backup-store,omitempty" flag:"etcd-backup-store"`
	EtcdImage                 *string  `json:"etcd-image,omitempty" flag:"etcd-image"`
	EtcdLeaderElectionTimeout *string  `json:"etcd-election-timeout,omitempty" flag:"etcd-election-timeout"`
	EtcdHearbeatInterval      *string  `json:"etcd-heartbeat-interval,omitempty" flag:"etcd-heartbeat-interval"`
	InitializeRBAC            *bool    `json:"initializeRBAC,omitempty" flag:"initialize-rbac"`
	LogLevel                  *int32   `json:"logLevel,omitempty" flag:"v"`
	Master                    *bool    `json:"master,omitempty" flag:"master"`
	PeerTLSCaFile             *string  `json:"peer-ca,omitempty" flag:"peer-ca"`
	PeerTLSCertFile           *string  `json:"peer-cert,omitempty" flag:"peer-cert"`
	PeerTLSKeyFile            *string  `json:"peer-key,omitempty" flag:"peer-key"`
	TLSAuth                   *bool    `json:"tls-auth,omitempty" flag:"tls-auth"`
	TLSCAFile                 *string  `json:"tls-ca,omitempty" flag:"tls-ca"`
	TLSCertFile               *string  `json:"tls-cert,omitempty" flag:"tls-cert"`
	TLSKeyFile                *string  `json:"tls-key,omitempty" flag:"tls-key"`
	Zone                      []string `json:"zone,omitempty" flag:"zone"`

	// ManageEtcd is true if protokube should manage etcd; being replaced by etcd-manager
	ManageEtcd bool `json:"manageEtcd,omitempty" flag:"manage-etcd"`

	// RemoveDNSNames allows us to remove dns records, so that they can be managed elsewhere
	// We use it e.g. for the switch to etcd-manager
	RemoveDNSNames string `json:"removeDNSNames,omitempty" flag:"remove-dns-names"`

	// BootstrapMasterNodeLabels applies the critical node-role labels to our node,
	// which lets us bring up the controllers that can only run on masters, which are then
	// responsible for node labels.  The node is specified by NodeName
	BootstrapMasterNodeLabels bool `json:"bootstrapMasterNodeLabels,omitempty" flag:"bootstrap-master-node-labels"`

	// NodeName is the name of the node as will be created in kubernetes.  Primarily used by BootstrapMasterNodeLabels.
	NodeName string `json:"nodeName,omitempty" flag:"node-name"`

	GossipProtocol *string `json:"gossip-protocol" flag:"gossip-protocol"`
	GossipListen   *string `json:"gossip-listen" flag:"gossip-listen"`
	GossipSecret   *string `json:"gossip-secret" flag:"gossip-secret"`

	GossipProtocolSecondary *string `json:"gossip-protocol-secondary" flag:"gossip-protocol-secondary"`
	GossipListenSecondary   *string `json:"gossip-listen-secondary" flag:"gossip-listen-secondary"`
	GossipSecretSecondary   *string `json:"gossip-secret-secondary" flag:"gossip-secret-secondary"`
}

// ProtokubeFlags is responsible for building the command line flags for protokube
func (t *ProtokubeBuilder) ProtokubeFlags(k8sVersion semver.Version) (*ProtokubeFlags, error) {
	imageVersion := t.Cluster.Spec.EtcdClusters[0].Version
	// overrides imageVersion if set
	etcdContainerImage := t.Cluster.Spec.EtcdClusters[0].Image

	var leaderElectionTimeout string
	var heartbeatInterval string

	if v := t.Cluster.Spec.EtcdClusters[0].LeaderElectionTimeout; v != nil {
		leaderElectionTimeout = convEtcdSettingsToMs(v)
	}

	if v := t.Cluster.Spec.EtcdClusters[0].HeartbeatInterval; v != nil {
		heartbeatInterval = convEtcdSettingsToMs(v)
	}

	f := &ProtokubeFlags{
		Channels:                  t.NodeupConfig.Channels,
		Containerized:             fi.Bool(true),
		EtcdLeaderElectionTimeout: s(leaderElectionTimeout),
		EtcdHearbeatInterval:      s(heartbeatInterval),
		LogLevel:                  fi.Int32(4),
		Master:                    b(t.IsMaster),
	}

	f.ManageEtcd = false
	if len(t.NodeupConfig.EtcdManifests) == 0 {
		klog.V(4).Infof("no EtcdManifests; protokube will manage etcd")
		f.ManageEtcd = true
	}

	if f.ManageEtcd {
		for _, e := range t.Cluster.Spec.EtcdClusters {
			// Because we can only specify a single EtcdBackupStore at the moment, we only backup main, not events
			if e.Name != "main" {
				continue
			}

			if e.Backups != nil {
				if f.EtcdBackupImage == "" {
					f.EtcdBackupImage = e.Backups.Image
				}

				if f.EtcdBackupStore == "" {
					f.EtcdBackupStore = e.Backups.BackupStore
				}
			}
		}

		// TODO this is duplicate code with etcd model
		image := fmt.Sprintf("k8s.gcr.io/etcd:%s", imageVersion)
		// override image if set as API value
		if etcdContainerImage != "" {
			image = etcdContainerImage
		}
		assets := assets.NewAssetBuilder(t.Cluster, "")
		remapped, err := assets.RemapImage(image)
		if err != nil {
			return nil, fmt.Errorf("unable to remap container %q: %v", image, err)
		}

		image = remapped
		f.EtcdImage = s(image)

		// check if we are using tls and add the options to protokube
		if t.UseEtcdTLS() {
			f.PeerTLSCaFile = s(filepath.Join(t.PathSrvKubernetes(), "ca.crt"))
			f.PeerTLSCertFile = s(filepath.Join(t.PathSrvKubernetes(), "etcd-peer.pem"))
			f.PeerTLSKeyFile = s(filepath.Join(t.PathSrvKubernetes(), "etcd-peer-key.pem"))
			f.TLSCAFile = s(filepath.Join(t.PathSrvKubernetes(), "ca.crt"))
			f.TLSCertFile = s(filepath.Join(t.PathSrvKubernetes(), "etcd.pem"))
			f.TLSKeyFile = s(filepath.Join(t.PathSrvKubernetes(), "etcd-key.pem"))
		}
		if t.UseEtcdTLSAuth() {
			enableAuth := true
			f.TLSAuth = b(enableAuth)
		}
	}

	f.InitializeRBAC = fi.Bool(true)

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
		klog.Warningf("DNSZone not specified; protokube won't be able to update DNS")
		// @TODO: Should we permit wildcard updates if zone is not specified?
		//argv = append(argv, "--zone=*/*")
	}

	if dns.IsGossipHostname(t.Cluster.Spec.MasterInternalName) {
		klog.Warningf("MasterInternalName %q implies gossip DNS", t.Cluster.Spec.MasterInternalName)
		f.DNSProvider = fi.String("gossip")
		if t.Cluster.Spec.GossipConfig != nil {
			f.GossipProtocol = t.Cluster.Spec.GossipConfig.Protocol
			f.GossipListen = t.Cluster.Spec.GossipConfig.Listen
			f.GossipSecret = t.Cluster.Spec.GossipConfig.Secret

			if t.Cluster.Spec.GossipConfig.Secondary != nil {
				f.GossipProtocolSecondary = t.Cluster.Spec.GossipConfig.Secondary.Protocol
				f.GossipListenSecondary = t.Cluster.Spec.GossipConfig.Secondary.Listen
				f.GossipSecretSecondary = t.Cluster.Spec.GossipConfig.Secondary.Secret
			}
		}

		// @TODO: This is hacky, but we want it so that we can have a different internal & external name
		internalSuffix := t.Cluster.Spec.MasterInternalName
		internalSuffix = strings.TrimPrefix(internalSuffix, "api.")
		f.DNSInternalSuffix = fi.String(internalSuffix)
	}

	if t.Cluster.Spec.CloudProvider != "" {
		f.Cloud = fi.String(t.Cluster.Spec.CloudProvider)

		if f.DNSProvider == nil {
			switch kops.CloudProviderID(t.Cluster.Spec.CloudProvider) {
			case kops.CloudProviderAWS:
				f.DNSProvider = fi.String("aws-route53")
			case kops.CloudProviderDO:
				f.DNSProvider = fi.String("digitalocean")
			case kops.CloudProviderGCE:
				f.DNSProvider = fi.String("google-clouddns")
			case kops.CloudProviderVSphere:
				f.DNSProvider = fi.String("coredns")
				f.ClusterID = fi.String(t.Cluster.ObjectMeta.Name)
				f.DNSServer = fi.String(*t.Cluster.Spec.CloudConfig.VSphereCoreDNSServer)
			default:
				klog.Warningf("Unknown cloudprovider %q; won't set DNS provider", t.Cluster.Spec.CloudProvider)
			}
		}
	}

	if f.DNSInternalSuffix == nil {
		f.DNSInternalSuffix = fi.String(".internal." + t.Cluster.ObjectMeta.Name)
	}

	if k8sVersion.Major == 1 && k8sVersion.Minor >= 16 {
		f.BootstrapMasterNodeLabels = true

		nodeName, err := t.NodeName()
		if err != nil {
			return nil, fmt.Errorf("error getting NodeName: %v", err)
		}
		f.NodeName = nodeName
	}

	// Remove DNS names if we're using etcd-manager
	if !f.ManageEtcd {
		var names []string

		// Mirroring the logic used to construct DNS names in protokube/pkg/protokube/etcd_cluster.go
		suffix := fi.StringValue(f.DNSInternalSuffix)
		if !strings.HasPrefix(suffix, ".") {
			suffix = "." + suffix
		}

		for _, c := range t.Cluster.Spec.EtcdClusters {
			clusterName := "etcd-" + c.Name
			if clusterName == "etcd-main" {
				clusterName = "etcd"
			}
			for _, m := range c.Members {
				name := clusterName + "-" + m.Name + suffix
				names = append(names, name)
			}
		}

		f.RemoveDNSNames = strings.Join(names, ",")
	}

	return f, nil
}

// ProtokubeEnvironmentVariables generates the environments variables for docker
func (t *ProtokubeBuilder) ProtokubeEnvironmentVariables() string {
	var buffer bytes.Buffer

	// TODO write out an environments file for this.  This is getting a tad long.

	// Pass in gossip dns connection limit
	if os.Getenv("GOSSIP_DNS_CONN_LIMIT") != "" {
		buffer.WriteString(" ")
		buffer.WriteString("--env 'GOSSIP_DNS_CONN_LIMIT=")
		buffer.WriteString(os.Getenv("GOSSIP_DNS_CONN_LIMIT"))
		buffer.WriteString("'")
		buffer.WriteString(" ")
	}

	// Pass in required credentials when using user-defined s3 endpoint
	if os.Getenv("AWS_REGION") != "" {
		buffer.WriteString(" ")
		buffer.WriteString("--env 'AWS_REGION=")
		buffer.WriteString(os.Getenv("AWS_REGION"))
		buffer.WriteString("'")
		buffer.WriteString(" ")
	}

	if os.Getenv("S3_ENDPOINT") != "" {
		buffer.WriteString(" ")
		buffer.WriteString("--env S3_ENDPOINT=")
		buffer.WriteString("'")
		buffer.WriteString(os.Getenv("S3_ENDPOINT"))
		buffer.WriteString("'")
		buffer.WriteString(" --env S3_REGION=")
		buffer.WriteString("'")
		buffer.WriteString(os.Getenv("S3_REGION"))
		buffer.WriteString("'")
		buffer.WriteString(" --env S3_ACCESS_KEY_ID=")
		buffer.WriteString("'")
		buffer.WriteString(os.Getenv("S3_ACCESS_KEY_ID"))
		buffer.WriteString("'")
		buffer.WriteString(" --env S3_SECRET_ACCESS_KEY=")
		buffer.WriteString("'")
		buffer.WriteString(os.Getenv("S3_SECRET_ACCESS_KEY"))
		buffer.WriteString("'")
		buffer.WriteString(" ")
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
		} {
			buffer.WriteString(" --env '")
			buffer.WriteString(envVar)
			buffer.WriteString("=")
			buffer.WriteString(os.Getenv(envVar))
			buffer.WriteString("'")
		}
	}

	if kops.CloudProviderID(t.Cluster.Spec.CloudProvider) == kops.CloudProviderDO && os.Getenv("DIGITALOCEAN_ACCESS_TOKEN") != "" {
		buffer.WriteString(" ")
		buffer.WriteString("--env 'DIGITALOCEAN_ACCESS_TOKEN=")
		buffer.WriteString(os.Getenv("DIGITALOCEAN_ACCESS_TOKEN"))
		buffer.WriteString("'")
		buffer.WriteString(" ")
	}

	if os.Getenv("OSS_REGION") != "" {
		buffer.WriteString(" ")
		buffer.WriteString("--env 'OSS_REGION=")
		buffer.WriteString(os.Getenv("OSS_REGION"))
		buffer.WriteString("'")
		buffer.WriteString(" ")
	}

	if os.Getenv("ALIYUN_ACCESS_KEY_ID") != "" {
		buffer.WriteString(" ")
		buffer.WriteString("--env 'ALIYUN_ACCESS_KEY_ID=")
		buffer.WriteString(os.Getenv("ALIYUN_ACCESS_KEY_ID"))
		buffer.WriteString("'")
		buffer.WriteString(" --env 'ALIYUN_ACCESS_KEY_SECRET=")
		buffer.WriteString(os.Getenv("ALIYUN_ACCESS_KEY_SECRET"))
		buffer.WriteString("'")
		buffer.WriteString(" ")
	}

	t.writeProxyEnvVars(&buffer)

	return buffer.String()
}

func (t *ProtokubeBuilder) writeProxyEnvVars(buffer *bytes.Buffer) {
	for _, envVar := range proxy.GetProxyEnvVars(t.Cluster.Spec.EgressProxy) {
		buffer.WriteString(" --env ")
		buffer.WriteString(envVar.Name)
		buffer.WriteString("=")
		buffer.WriteString(envVar.Value)
		buffer.WriteString(" ")
	}
}
