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

	"github.com/blang/semver"
	"github.com/golang/glog"
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
		glog.V(2).Infof("skipping the provisioning of protokube on the nodes")
		return nil
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
		if t.UseEtcdTLS() {
			for _, x := range []string{"etcd", "etcd-client"} {
				if err := t.BuildCertificateTask(c, x, fmt.Sprintf("%s.pem", x)); err != nil {
					return err
				}
			}
			for _, x := range []string{"etcd", "etcd-client"} {
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
	protokubeFlagsArgs, err := flagbuilder.BuildFlags(protokubeFlags)
	if err != nil {
		return nil, err
	}

	dockerArgs := []string{
		"/usr/bin/docker", "run",
		"-v", "/:/rootfs/",
		"-v", "/var/run/dbus:/var/run/dbus",
		"-v", "/run/systemd:/run/systemd",
	}

	// add kubectl only if a master
	// path changes depending on distro, and always mount it on /opt/kops/bin
	// kubectl is downloaded and installed by other tasks
	if t.IsMaster {
		dockerArgs = append(dockerArgs, []string{
			"-v", t.KubectlPath() + ":/opt/kops/bin:ro",
			"--env", "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/opt/kops/bin",
		}...)
	}

	dockerArgs = append(dockerArgs, []string{
		"--net=host",
		"--pid=host",   // Needed for mounting in a container (when using systemd mounting?)
		"--privileged", // We execute in the host namespace
		"--env", "KUBECONFIG=/rootfs/var/lib/kops/kubeconfig",
		t.ProtokubeEnvironmentVariables(),
		t.ProtokubeImageName(),
		"/usr/bin/protokube",
	}...)

	protokubeCommand := strings.Join(dockerArgs, " ") + " " + protokubeFlagsArgs

	manifest := &systemd.Manifest{}
	manifest.Set("Unit", "Description", "Kubernetes Protokube Service")
	manifest.Set("Unit", "Documentation", "https://github.com/kubernetes/kops")
	manifest.Set("Service", "ExecStartPre", t.ProtokubeImagePullCommand())
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
		name = kopsbase.DefaultProtokubeImageName()
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
		glog.V(4).Infof("no EtcdManifests; protokube will manage etcd")
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
			f.PeerTLSCertFile = s(filepath.Join(t.PathSrvKubernetes(), "etcd.pem"))
			f.PeerTLSKeyFile = s(filepath.Join(t.PathSrvKubernetes(), "etcd-key.pem"))
			f.TLSCAFile = s(filepath.Join(t.PathSrvKubernetes(), "ca.crt"))
			f.TLSCertFile = s(filepath.Join(t.PathSrvKubernetes(), "etcd.pem"))
			f.TLSKeyFile = s(filepath.Join(t.PathSrvKubernetes(), "etcd-key.pem"))
		}
		if t.UseEtcdTLSAuth() {
			enableAuth := true
			f.TLSAuth = b(enableAuth)
		}
	}

	// initialize rbac on Kubernetes >= 1.6 and master
	if k8sVersion.Major == 1 && k8sVersion.Minor >= 6 {
		f.InitializeRBAC = fi.Bool(true)
	}

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
		// @TODO: Should we permit wildcard updates if zone is not specified?
		//argv = append(argv, "--zone=*/*")
	}

	if dns.IsGossipHostname(t.Cluster.Spec.MasterInternalName) {
		glog.Warningf("MasterInternalName %q implies gossip DNS", t.Cluster.Spec.MasterInternalName)
		f.DNSProvider = fi.String("gossip")

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
				f.ClusterID = fi.String(t.Cluster.Name)
			case kops.CloudProviderGCE:
				f.DNSProvider = fi.String("google-clouddns")
			case kops.CloudProviderVSphere:
				f.DNSProvider = fi.String("coredns")
				f.ClusterID = fi.String(t.Cluster.ObjectMeta.Name)
				f.DNSServer = fi.String(*t.Cluster.Spec.CloudConfig.VSphereCoreDNSServer)
			default:
				glog.Warningf("Unknown cloudprovider %q; won't set DNS provider", t.Cluster.Spec.CloudProvider)
			}
		}
	}

	if f.DNSInternalSuffix == nil {
		f.DNSInternalSuffix = fi.String(".internal." + t.Cluster.ObjectMeta.Name)
	}

	if k8sVersion.Major == 1 && k8sVersion.Minor <= 5 {
		f.ApplyTaints = fi.Bool(true)
	}

	return f, nil
}

// ProtokubeEnvironmentVariables generates the environments variables for docker
func (t *ProtokubeBuilder) ProtokubeEnvironmentVariables() string {
	var buffer bytes.Buffer

	// TODO write out an environments file for this.  This is getting a tad long.

	// Passin gossip dns connection limit
	if os.Getenv("GOSSIP_DNS_CONN_LIMIT") != "" {
		buffer.WriteString(" ")
		buffer.WriteString("-e 'GOSSIP_DNS_CONN_LIMIT=")
		buffer.WriteString(os.Getenv("GOSSIP_DNS_CONN_LIMIT"))
		buffer.WriteString("'")
		buffer.WriteString(" ")
	}

	// Pass in required credentials when using user-defined s3 endpoint
	if os.Getenv("AWS_REGION") != "" {
		buffer.WriteString(" ")
		buffer.WriteString("-e 'AWS_REGION=")
		buffer.WriteString(os.Getenv("AWS_REGION"))
		buffer.WriteString("'")
		buffer.WriteString(" ")
	}

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

	if kops.CloudProviderID(t.Cluster.Spec.CloudProvider) == kops.CloudProviderDO && os.Getenv("DIGITALOCEAN_ACCESS_TOKEN") != "" {
		buffer.WriteString(" ")
		buffer.WriteString("-e 'DIGITALOCEAN_ACCESS_TOKEN=")
		buffer.WriteString(os.Getenv("DIGITALOCEAN_ACCESS_TOKEN"))
		buffer.WriteString("'")
		buffer.WriteString(" ")
	}

	t.writeProxyEnvVars(&buffer)

	return buffer.String()
}

func (t *ProtokubeBuilder) writeProxyEnvVars(buffer *bytes.Buffer) {
	for _, envVar := range getProxyEnvVars(t.Cluster.Spec.EgressProxy) {
		buffer.WriteString(" -e ")
		buffer.WriteString(envVar.Name)
		buffer.WriteString("=")
		buffer.WriteString(envVar.Value)
		buffer.WriteString(" ")
	}
}
