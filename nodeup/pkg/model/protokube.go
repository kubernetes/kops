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

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/pkg/k8scodecs"
	"k8s.io/kops/pkg/kubemanifest"

	kopsbase "k8s.io/kops"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/dns"
	"k8s.io/kops/pkg/flagbuilder"
	"k8s.io/kops/pkg/rbac"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
	"k8s.io/kops/util/pkg/proxy"

	"github.com/blang/semver/v4"
	"k8s.io/klog/v2"
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

	if protokubeImage := t.NodeupConfig.ProtokubeImage[t.Architecture]; protokubeImage != nil {
		c.AddTask(&nodetasks.LoadImageTask{
			Name:    "protokube",
			Sources: protokubeImage.Sources,
			Hash:    protokubeImage.Hash,
			Runtime: t.Cluster.Spec.ContainerRuntime,
		})
	}

	if t.IsMaster {
		name := nodetasks.PKIXName{
			CommonName:   "kops",
			Organization: []string{rbac.SystemPrivilegedGroup},
		}
		kubeconfig := t.BuildIssuedKubeconfig("kops", name, c)

		c.AddTask(&nodetasks.File{
			Path:     "/var/lib/kops/kubeconfig",
			Contents: kubeconfig,
			Type:     nodetasks.FileType_File,
			Mode:     s("0400"),
		})

		// retrieve the etcd peer certificates and private keys from the keystore
		if !t.UseEtcdManager() && t.UseEtcdTLS() {
			for _, x := range []string{"etcd", "etcd-peer", "etcd-client"} {
				if err := t.BuildCertificateTask(c, x, fmt.Sprintf("%s.pem", x), nil); err != nil {
					return err
				}
			}
			for _, x := range []string{"etcd", "etcd-peer", "etcd-client"} {
				if err := t.BuildPrivateKeyTask(c, x, fmt.Sprintf("%s-key.pem", x), nil); err != nil {
					return err
				}
			}
		}

		// rather than a systemd service, lets build a pod
		pod, err := t.buildPod()
		if err != nil {
			return fmt.Errorf("Error in building the protokube pod")
		}

		manifest, err := k8scodecs.ToVersionedYaml(pod)
		if err != nil {
			return fmt.Errorf("Error in get versioned yaml for protokube pod")
		}

		c.AddTask(&nodetasks.File{
			Path:     "/etc/kubernetes/manifests/protokube.manifest",
			Contents: fi.NewBytesResource(manifest),
			Type:     nodetasks.FileType_File,
		})
	}

	return nil
}

func (t *ProtokubeBuilder) buildPod() (*v1.Pod, error) {
	pod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "protokube",
			Namespace: "kube-system",
			Labels: map[string]string{
				"app": "protokube",
			},
		},
		Spec: v1.PodSpec{
			HostNetwork: true,
			HostPID:     true,
		},
	}

	var image string
	image = t.ProtokubeImageName()
	if t.Cluster.Spec.ContainerRuntime == "containerd" {
		image = "docker.io/library/" + t.ProtokubeImageName()
	}

	container := &v1.Container{
		Name:  "protokube",
		Image: image,
		Env: []v1.EnvVar{
			{
				Name:  "KUBECONFIG",
				Value: "/rootfs/var/lib/kops/kubeconfig",
			},
			{
				Name:  "PATH",
				Value: "/opt/kops/bin:/usr/bin:/sbin:/bin",
			},
		},
		Resources: v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU: resource.MustParse("100m"),
			},
		},
	}

	addHostPathMappingWithMount(pod, container, "rootfs", "/", "/rootfs")
	addHostPathMappingWithMount(pod, container, "kubectlpath", t.KubectlPath(), "/opt/kops/bin")
	addHostPathMapping(pod, container, "bin", "/bin")
	addHostPathMapping(pod, container, "lib", "/lib")
	addHostPathMapping(pod, container, "sbin", "/sbin")
	addHostPathMapping(pod, container, "usrbin", "/usr/bin")
	addHostPathMapping(pod, container, "dbus", "/var/run/dbus").ReadOnly = false
	addHostPathMapping(pod, container, "systemd", "/run/systemd").ReadOnly = false
	addHostPathMapping(pod, container, "lib64", "/lib64")
	addHostPathMapping(pod, container, "sslcerts", "/etc/ssl/certs").ReadOnly = false

	k8sVersion, err := util.ParseKubernetesVersion(t.Cluster.Spec.KubernetesVersion)
	if err != nil || k8sVersion == nil {
		return nil, fmt.Errorf("unable to parse KubernetesVersion %q", t.Cluster.Spec.KubernetesVersion)
	}

	protokubeFlags, err := t.ProtokubeFlags(*k8sVersion)
	if err != nil {
		return nil, err
	}
	protokubeRunArgs, err := flagbuilder.BuildFlagsList(protokubeFlags)
	if err != nil {
		return nil, err
	}

	container.Command = []string{"/protokube"}
	container.Args = protokubeRunArgs
	container.SecurityContext = &v1.SecurityContext{Privileged: b(true)}

	pod.Spec.Containers = append(pod.Spec.Containers, *container)

	kubemanifest.MarkPodAsClusterCritical(pod)
	kubemanifest.MarkPodAsCritical(pod)

	return pod, nil
}

// ProtokubeImageName returns the docker image for protokube
func (t *ProtokubeBuilder) ProtokubeImageName() string {
	name := ""
	if t.NodeupConfig.ProtokubeImage[t.Architecture] != nil && t.NodeupConfig.ProtokubeImage[t.Architecture].Name != "" {
		name = t.NodeupConfig.ProtokubeImage[t.Architecture].Name
	}
	if name == "" {
		// use current default corresponding to this version of nodeup
		name = kopsbase.DefaultProtokubeImageName()
	}
	return name
}

// ProtokubeFlags are the flags for protokube
type ProtokubeFlags struct {
	ApplyTaints               *bool    `json:"applyTaints,omitempty" flag:"apply-taints"`
	Channels                  []string `json:"channels,omitempty" flag:"channels"`
	Cloud                     *string  `json:"cloud,omitempty" flag:"cloud"`
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

	GossipProtocolSecondary *string `json:"gossip-protocol-secondary" flag:"gossip-protocol-secondary" flag-include-empty:"true"`
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
			"OS_APPLICATION_CREDENTIAL_ID",
			"OS_APPLICATION_CREDENTIAL_SECRET",
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

	if os.Getenv("AZURE_STORAGE_ACCOUNT") != "" {
		buffer.WriteString(" ")
		buffer.WriteString("--env 'AZURE_STORAGE_ACCOUNT=")
		buffer.WriteString(os.Getenv("AZURE_STORAGE_ACCOUNT"))
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
