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

package etcdmanager

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/flagbuilder"
	"k8s.io/kops/pkg/k8scodecs"
	"k8s.io/kops/pkg/kubemanifest"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/wellknownports"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"
	"k8s.io/kops/upup/pkg/fi/cloudup/do"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/hetzner"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
	"k8s.io/kops/upup/pkg/fi/cloudup/scaleway"
	"k8s.io/kops/upup/pkg/fi/fitasks"
	"k8s.io/kops/util/pkg/env"
	"k8s.io/kops/util/pkg/exec"
)

// EtcdManagerBuilder builds the manifest for the etcd-manager
type EtcdManagerBuilder struct {
	*model.KopsModelContext
	Lifecycle    fi.Lifecycle
	AssetBuilder *assets.AssetBuilder
}

var _ fi.CloudupModelBuilder = &EtcdManagerBuilder{}

// Build creates the tasks
func (b *EtcdManagerBuilder) Build(c *fi.CloudupModelBuilderContext) error {
	for _, etcdCluster := range b.Cluster.Spec.EtcdClusters {
		backupStore := ""
		if etcdCluster.Backups != nil {
			backupStore = etcdCluster.Backups.BackupStore
		}
		if backupStore == "" {
			return fmt.Errorf("backupStore must be set for use with etcd-manager")
		}

		for _, member := range etcdCluster.Members {
			instanceGroupName := fi.ValueOf(member.InstanceGroup)
			manifest, err := b.buildManifest(etcdCluster, instanceGroupName)
			if err != nil {
				return err
			}

			manifestYAML, err := k8scodecs.ToVersionedYaml(manifest)
			if err != nil {
				return fmt.Errorf("error marshaling manifest to yaml: %v", err)
			}

			name := fmt.Sprintf("%s-%s", etcdCluster.Name, instanceGroupName)
			c.AddTask(&fitasks.ManagedFile{
				Contents:  fi.NewBytesResource(manifestYAML),
				Lifecycle: b.Lifecycle,
				Location:  fi.PtrTo("manifests/etcd/" + name + ".yaml"),
				Name:      fi.PtrTo("manifests-etcdmanager-" + name),
			})
		}

		info := &etcdClusterSpec{
			EtcdVersion: etcdCluster.Version,
			MemberCount: int32(len(etcdCluster.Members)),
		}

		d, err := json.MarshalIndent(info, "", "  ")
		if err != nil {
			return err
		}

		// Ensure a unique backup location for each etcd cluster
		// if a backupStore is not specified.
		var location string
		if backupStore == "" {
			location = "backups/etcd/" + etcdCluster.Name
		}

		c.AddTask(&fitasks.ManagedFile{
			Contents:  fi.NewBytesResource(d),
			Lifecycle: b.Lifecycle,
			Base:      fi.PtrTo(backupStore),
			// TODO: We need this to match the backup base (currently)
			Location: fi.PtrTo(location + "/control/etcd-cluster-spec"),
			Name:     fi.PtrTo("etcd-cluster-spec-" + etcdCluster.Name),
		})

		// We create a CA keypair to enable secure communication
		c.AddTask(&fitasks.Keypair{
			Name:      fi.PtrTo("etcd-manager-ca-" + etcdCluster.Name),
			Lifecycle: b.Lifecycle,
			Subject:   "cn=etcd-manager-ca-" + etcdCluster.Name,
			Type:      "ca",
		})

		// We create a CA for etcd peers and a separate one for clients
		c.AddTask(&fitasks.Keypair{
			Name:      fi.PtrTo("etcd-peers-ca-" + etcdCluster.Name),
			Lifecycle: b.Lifecycle,
			Subject:   "cn=etcd-peers-ca-" + etcdCluster.Name,
			Type:      "ca",
		})

		// Because API server can only have a single client-cert, we need to share a client CA
		c.EnsureTask(&fitasks.Keypair{
			Name:      fi.PtrTo("etcd-clients-ca"),
			Lifecycle: b.Lifecycle,
			Subject:   "cn=etcd-clients-ca",
			Type:      "ca",
		})

		if etcdCluster.Name == "cilium" {
			clientsCaCilium := &fitasks.Keypair{
				Name:      fi.PtrTo("etcd-clients-ca-cilium"),
				Lifecycle: b.Lifecycle,
				Subject:   "cn=etcd-clients-ca-cilium",
				Type:      "ca",
			}
			c.AddTask(clientsCaCilium)
		}
	}

	return nil
}

type etcdClusterSpec struct {
	MemberCount int32  `json:"memberCount,omitempty"`
	EtcdVersion string `json:"etcdVersion,omitempty"`
}

func (b *EtcdManagerBuilder) buildManifest(etcdCluster kops.EtcdClusterSpec, instanceGroupName string) (*v1.Pod, error) {
	return b.buildPod(etcdCluster, instanceGroupName)
}

// Until we introduce the bundle, we hard-code the manifest
var defaultManifest = `
apiVersion: v1
kind: Pod
metadata:
  name: etcd-manager
  namespace: kube-system
spec:
  containers:
  - name: etcd-manager
    image: us-central1-docker.pkg.dev/k8s-staging-images/etcd-manager/etcd-manager-slim:f1ea649
    resources:
      requests:
        cpu: 100m
        memory: 100Mi
    # TODO: Would be nice to reduce these permissions; needed for volume mounting
    securityContext:
      privileged: true
    volumeMounts:
    # TODO: Would be nice to scope this more tightly, but needed for volume mounting
    - mountPath: /rootfs
      name: rootfs
    - mountPath: /run
      name: run
    - mountPath: /etc/kubernetes/pki/etcd-manager
      name: pki
    - mountPath: /opt
      name: opt
  hostNetwork: true
  hostPID: true # helps with mounting volumes from inside a container
  volumes:
  - hostPath:
      path: /
      type: Directory
    name: rootfs
  - hostPath:
      path: /run
      type: DirectoryOrCreate
    name: run
  - hostPath:
      path: /etc/kubernetes/pki/etcd-manager
      type: DirectoryOrCreate
    name: pki
  - name: opt
    emptyDir: {}
`

const kopsUtilsImage = "registry.k8s.io/kops/kops-utils-cp:1.30.0-beta.1"

// buildPod creates the pod spec, based on the EtcdClusterSpec
func (b *EtcdManagerBuilder) buildPod(etcdCluster kops.EtcdClusterSpec, instanceGroupName string) (*v1.Pod, error) {
	var pod *v1.Pod
	var container *v1.Container

	var manifest []byte

	// TODO: pull from bundle
	bundle := "(embedded etcd manifest)"
	manifest = []byte(defaultManifest)

	{
		objects, err := model.ParseManifest(manifest)
		if err != nil {
			return nil, err
		}
		if len(objects) != 1 {
			return nil, fmt.Errorf("expected exactly one object in manifest %s, found %d", bundle, len(objects))
		}
		if podObject, ok := objects[0].(*v1.Pod); !ok {
			return nil, fmt.Errorf("expected v1.Pod object in manifest %s, found %T", bundle, objects[0])
		} else {
			pod = podObject
		}
	}

	{
		utilMounts := []v1.VolumeMount{
			{
				MountPath: "/opt",
				Name:      "opt",
			},
		}
		{
			initContainer := v1.Container{
				Name:    "kops-utils-cp",
				Image:   kopsUtilsImage,
				Command: []string{"/ko-app/kops-utils-cp"},
				Args: []string{
					"--target-dir=/opt/kops-utils/",
					"--src=/ko-app/kops-utils-cp",
				},
				VolumeMounts: utilMounts,
			}
			pod.Spec.InitContainers = append(pod.Spec.InitContainers, initContainer)
		}

		symlinkToVersions := sets.NewString()
		for _, etcdVersion := range etcdSupportedVersions() {
			if etcdVersion.SymlinkToVersion != "" {
				symlinkToVersions.Insert(etcdVersion.SymlinkToVersion)
				continue
			}

			initContainer := v1.Container{
				Name:         "init-etcd-" + strings.ReplaceAll(etcdVersion.Version, ".", "-"),
				Image:        etcdVersion.Image,
				Command:      []string{"/opt/kops-utils/kops-utils-cp"},
				VolumeMounts: utilMounts,
			}

			initContainer.Args = []string{
				"--target-dir=/opt/etcd-v" + etcdVersion.Version,
				"--src=/usr/local/bin/etcd",
				"--src=/usr/local/bin/etcdctl",
			}

			pod.Spec.InitContainers = append(pod.Spec.InitContainers, initContainer)
		}

		for _, symlinkToVersion := range symlinkToVersions.List() {
			targetVersions := sets.NewString()

			for _, etcdVersion := range etcdSupportedVersions() {
				if etcdVersion.SymlinkToVersion == symlinkToVersion {
					targetVersions.Insert(etcdVersion.Version)
				}
			}

			initContainer := v1.Container{
				Name:         "init-etcd-symlinks-" + strings.ReplaceAll(symlinkToVersion, ".", "-"),
				Image:        kopsUtilsImage,
				Command:      []string{"/opt/kops-utils/kops-utils-cp"},
				VolumeMounts: utilMounts,
			}

			initContainer.Args = []string{
				"--symlink",
			}
			for _, targetVersion := range targetVersions.List() {
				initContainer.Args = append(initContainer.Args, "--target-dir=/opt/etcd-v"+targetVersion)
			}
			// NOTE: Flags must come before positional arguments
			initContainer.Args = append(initContainer.Args,
				"--src=/opt/etcd-v"+symlinkToVersion+"/etcd",
				"--src=/opt/etcd-v"+symlinkToVersion+"/etcdctl",
			)

			pod.Spec.InitContainers = append(pod.Spec.InitContainers, initContainer)
		}

		// Remap image via AssetBuilder
		for i := range pod.Spec.InitContainers {
			initContainer := &pod.Spec.InitContainers[i]
			remapped, err := b.AssetBuilder.RemapImage(initContainer.Image)
			if err != nil {
				return nil, fmt.Errorf("unable to remap init container image %q: %w", container.Image, err)
			}
			initContainer.Image = remapped
		}
	}

	{
		if len(pod.Spec.Containers) != 1 {
			return nil, fmt.Errorf("expected exactly one container in etcd-manager Pod, found %d", len(pod.Spec.Containers))
		}
		container = &pod.Spec.Containers[0]

		if etcdCluster.Manager != nil && etcdCluster.Manager.Image != "" {
			klog.Warningf("overloading image in manifest %s with images %s", bundle, etcdCluster.Manager.Image)
			container.Image = etcdCluster.Manager.Image
		}

		// Remap image via AssetBuilder
		remapped, err := b.AssetBuilder.RemapImage(container.Image)
		if err != nil {
			return nil, fmt.Errorf("unable to remap container image %q: %w", container.Image, err)
		}
		container.Image = remapped
	}

	var clientHost string

	if featureflag.APIServerNodes.Enabled() {
		clientHost = etcdCluster.Name + ".etcd.internal." + b.ClusterName()
	} else {
		clientHost = "__name__"
	}

	clusterName := "etcd-" + etcdCluster.Name
	backupStore := ""
	if etcdCluster.Backups != nil {
		backupStore = etcdCluster.Backups.BackupStore
	}

	pod.Name = "etcd-manager-" + etcdCluster.Name

	if pod.Annotations == nil {
		pod.Annotations = make(map[string]string)
	}

	if featureflag.APIServerNodes.Enabled() {
		pod.Annotations["dns.alpha.kubernetes.io/internal"] = clientHost
	}

	if pod.Labels == nil {
		pod.Labels = make(map[string]string)
	}
	for k, v := range SelectorForCluster(etcdCluster) {
		pod.Labels[k] = v
	}

	// The dns suffix logic mirrors the existing logic, so we should be compatible with existing clusters
	// (etcd makes it difficult to change peer urls, treating it as a cluster event, for reasons unknown)
	dnsInternalSuffix := ".internal." + b.Cluster.Name

	ports, err := PortsForCluster(etcdCluster)
	if err != nil {
		return nil, err
	}

	switch etcdCluster.Name {
	case "main":
		clusterName = "etcd"

	case "events":
		// ok

	case "cilium":
		if !featureflag.APIServerNodes.Enabled() {
			clientHost = b.Cluster.APIInternalName()
		}
	default:
		return nil, fmt.Errorf("unknown etcd cluster key %q", etcdCluster.Name)
	}

	if backupStore == "" {
		return nil, fmt.Errorf("backupStore must be set for use with etcd-manager")
	}

	name := clusterName
	if !strings.HasPrefix(name, "etcd") {
		// For sanity, and to avoid collisions in directories / dns
		return nil, fmt.Errorf("unexpected name for etcd cluster (must start with etcd): %q", name)
	}
	logFile := "/var/log/" + name + ".log"

	config := &config{
		Containerized: true,
		ClusterName:   clusterName,
		BackupStore:   backupStore,
		GrpcPort:      ports.GRPCPort,
		DNSSuffix:     dnsInternalSuffix,
	}

	config.LogLevel = 6

	if etcdCluster.Manager != nil && etcdCluster.Manager.LogLevel != nil {
		klog.Warningf("overriding log level in manifest %s, new level is %d", bundle, int(*etcdCluster.Manager.LogLevel))
		config.LogLevel = int(*etcdCluster.Manager.LogLevel)
	}

	if etcdCluster.Manager != nil && etcdCluster.Manager.BackupInterval != nil {
		config.BackupInterval = fi.PtrTo(etcdCluster.Manager.BackupInterval.Duration.String())
	}

	if etcdCluster.Manager != nil && etcdCluster.Manager.DiscoveryPollInterval != nil {
		config.DiscoveryPollInterval = fi.PtrTo(etcdCluster.Manager.DiscoveryPollInterval.Duration.String())
	}

	{
		scheme := "https"

		config.PeerUrls = fmt.Sprintf("%s://__name__:%d", scheme, ports.PeerPort)
		config.ClientUrls = fmt.Sprintf("%s://%s:%d", scheme, clientHost, ports.ClientPort)
		config.QuarantineClientUrls = fmt.Sprintf("%s://__name__:%d", scheme, ports.QuarantinedGRPCPort)

		// TODO: We need to wire these into the etcd-manager spec
		// // add timeout/heartbeat settings
		if etcdCluster.LeaderElectionTimeout != nil {
			//      envs = append(envs, v1.EnvVar{Name: "ETCD_ELECTION_TIMEOUT", Value: convEtcdSettingsToMs(etcdClusterSpec.LeaderElectionTimeout)})
			return nil, fmt.Errorf("LeaderElectionTimeout not supported by etcd-manager")
		}
		if etcdCluster.HeartbeatInterval != nil {
			//      envs = append(envs, v1.EnvVar{Name: "ETCD_HEARTBEAT_INTERVAL", Value: convEtcdSettingsToMs(etcdClusterSpec.HeartbeatInterval)})
			return nil, fmt.Errorf("HeartbeatInterval not supported by etcd-manager")
		}
	}

	{
		switch b.Cluster.GetCloudProvider() {
		case kops.CloudProviderAWS:
			config.VolumeProvider = "aws"

			config.VolumeTag = []string{
				fmt.Sprintf("kubernetes.io/cluster/%s=owned", b.Cluster.Name),
				awsup.TagNameEtcdClusterPrefix + etcdCluster.Name,
				awsup.TagNameRolePrefix + "control-plane=1",
			}
			config.VolumeNameTag = awsup.TagNameEtcdClusterPrefix + etcdCluster.Name

		case kops.CloudProviderAzure:
			config.VolumeProvider = "azure"

			config.VolumeTag = []string{
				// Use dash (_) as a splitter. Other CSPs use slash (/), but slash is not
				// allowed as a tag key in Azure.
				fmt.Sprintf("kubernetes.io_cluster_%s=owned", b.Cluster.Name),
				azure.TagNameEtcdClusterPrefix + etcdCluster.Name,
				azure.TagNameRolePrefix + "control_plane=1",
			}
			config.VolumeNameTag = azure.TagNameEtcdClusterPrefix + etcdCluster.Name

		case kops.CloudProviderGCE:
			config.VolumeProvider = "gce"

			clusterLabel := gce.LabelForCluster(b.Cluster.Name)
			config.VolumeTag = []string{
				clusterLabel.Key + "=" + clusterLabel.Value,
				gce.GceLabelNameEtcdClusterPrefix + etcdCluster.Name,
				gce.GceLabelNameRolePrefix + "master=master",
			}
			config.VolumeNameTag = gce.GceLabelNameEtcdClusterPrefix + etcdCluster.Name

		case kops.CloudProviderDO:
			config.VolumeProvider = "do"

			// DO does not support . in tags / names
			safeClusterName := do.SafeClusterName(b.Cluster.Name)

			config.VolumeTag = []string{
				fmt.Sprintf("%s=%s", do.TagKubernetesClusterNamePrefix, safeClusterName),
				do.TagKubernetesClusterIndex,
			}
			config.VolumeNameTag = do.TagNameEtcdClusterPrefix + etcdCluster.Name

		case kops.CloudProviderHetzner:
			config.VolumeProvider = "hetzner"

			config.VolumeTag = []string{
				fmt.Sprintf("%s=%s", hetzner.TagKubernetesClusterName, b.Cluster.Name),
				fmt.Sprintf("%s=%s", hetzner.TagKubernetesVolumeRole, etcdCluster.Name),
			}
			config.VolumeNameTag = fmt.Sprintf("%s=%s", hetzner.TagKubernetesInstanceGroup, instanceGroupName)

		case kops.CloudProviderOpenstack:
			config.VolumeProvider = "openstack"

			config.VolumeTag = []string{
				openstack.TagNameEtcdClusterPrefix + etcdCluster.Name,
				openstack.TagNameRolePrefix + "control-plane=1",
				fmt.Sprintf("%s=%s", openstack.TagClusterName, b.Cluster.Name),
			}
			config.VolumeNameTag = openstack.TagNameEtcdClusterPrefix + etcdCluster.Name
			config.NetworkCIDR = fi.PtrTo(b.Cluster.Spec.Networking.NetworkCIDR)

		case kops.CloudProviderScaleway:
			config.VolumeProvider = "scaleway"

			config.VolumeTag = []string{
				fmt.Sprintf("%s=%s", scaleway.TagClusterName, b.Cluster.Name),
				fmt.Sprintf("%s=%s", scaleway.TagNameEtcdClusterPrefix, etcdCluster.Name),
				fmt.Sprintf("%s=%s", scaleway.TagNameRolePrefix, scaleway.TagRoleControlPlane),
			}
			config.VolumeNameTag = fmt.Sprintf("%s=%s", scaleway.TagInstanceGroup, instanceGroupName)

		case kops.CloudProviderMetal:
			config.VolumeProvider = "external"
			config.BackupStore = "file:///mnt/disks/backups"
			config.VolumeTag = []string{
				fmt.Sprintf("%s--%s--", b.Cluster.Name, etcdCluster.Name),
			}

			staticConfig := &StaticConfig{
				EtcdVersion: etcdCluster.Version,
			}
			staticConfig.Nodes = append(staticConfig.Nodes, StaticConfigNode{
				ID: fmt.Sprintf("%s--%s--%d", b.Cluster.Name, etcdCluster.Name, 0),
				// TODO: Support multiple control-plane nodes (will be interesting!)
				IP: []string{"node0" + "." + etcdCluster.Name + "." + b.Cluster.Name},
			})
			b, err := json.Marshal(staticConfig)
			if err != nil {
				return nil, fmt.Errorf("building static config: %w", err)
			}
			config.StaticConfig = string(b)

		default:
			return nil, fmt.Errorf("CloudProvider %q not supported with etcd-manager", b.Cluster.GetCloudProvider())
		}
	}

	args, err := flagbuilder.BuildFlagsList(config)
	if err != nil {
		return nil, err
	}

	{
		container.Command = exec.WithTee("/etcd-manager", args, "/var/log/etcd.log")

		cpuRequest := resource.MustParse("200m")
		if etcdCluster.CPURequest != nil {
			cpuRequest = *etcdCluster.CPURequest
		}
		memoryRequest := resource.MustParse("100Mi")
		if etcdCluster.MemoryRequest != nil {
			memoryRequest = *etcdCluster.MemoryRequest
		}

		container.Resources = v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU:    cpuRequest,
				v1.ResourceMemory: memoryRequest,
			},
		}

		kubemanifest.AddHostPathMapping(pod, container, "varlogetcd", "/var/log/etcd.log",
			kubemanifest.WithReadWrite(),
			kubemanifest.WithType(v1.HostPathFileOrCreate),
			kubemanifest.WithHostPath(logFile))

		if fi.ValueOf(b.Cluster.Spec.UseHostCertificates) {
			kubemanifest.AddHostPathMapping(pod, container, "etc-ssl-certs", "/etc/ssl/certs", kubemanifest.WithType(v1.HostPathDirectoryOrCreate))
		}
	}

	envMap := env.BuildSystemComponentEnvVars(&b.Cluster.Spec)

	container.Env = envMap.ToEnvVars()

	if etcdCluster.Manager != nil {
		if etcdCluster.Manager.BackupRetentionDays != nil {
			envVar := v1.EnvVar{
				Name:  "ETCD_MANAGER_DAILY_BACKUPS_RETENTION",
				Value: strconv.FormatUint(uint64(fi.ValueOf(etcdCluster.Manager.BackupRetentionDays)), 10) + "d",
			}

			container.Env = append(container.Env, envVar)
		}

		if len(etcdCluster.Manager.ListenMetricsURLs) > 0 {
			envVar := v1.EnvVar{
				Name:  "ETCD_LISTEN_METRICS_URLS",
				Value: strings.Join(etcdCluster.Manager.ListenMetricsURLs, ","),
			}

			container.Env = append(container.Env, envVar)
		}

		for _, envVar := range etcdCluster.Manager.Env {
			klog.V(2).Infof("overloading ENV var in manifest %s with %s=%s", bundle, envVar.Name, envVar.Value)
			configOverwrite := v1.EnvVar{
				Name:  envVar.Name,
				Value: envVar.Value,
			}

			container.Env = append(container.Env, configOverwrite)
		}
	}

	{
		foundPKI := false
		for i := range pod.Spec.Volumes {
			v := &pod.Spec.Volumes[i]
			if v.Name == "pki" {
				if v.HostPath == nil {
					return nil, fmt.Errorf("found PKI volume, but HostPath was nil")
				}
				dirname := "etcd-manager-" + etcdCluster.Name
				v.HostPath.Path = "/etc/kubernetes/pki/" + dirname
				foundPKI = true
			}
		}
		if !foundPKI {
			return nil, fmt.Errorf("did not find PKI volume")
		}
	}

	kubemanifest.MarkPodAsCritical(pod)
	kubemanifest.MarkPodAsClusterCritical(pod)

	return pod, nil
}

// config defines the flags for etcd-manager
type config struct {
	// LogLevel sets the log verbosity level
	LogLevel int `flag:"v"`

	// Containerized is set if etcd-manager is running in a container
	Containerized bool `flag:"containerized"`

	// PKIDir is set to the directory for PKI keys, used to secure commucations between etcd-manager peers
	PKIDir string `flag:"pki-dir"`

	Address               string   `flag:"address"`
	PeerUrls              string   `flag:"peer-urls"`
	GrpcPort              int      `flag:"grpc-port"`
	ClientUrls            string   `flag:"client-urls"`
	DiscoveryPollInterval *string  `flag:"discovery-poll-interval"`
	QuarantineClientUrls  string   `flag:"quarantine-client-urls"`
	ClusterName           string   `flag:"cluster-name"`
	BackupStore           string   `flag:"backup-store"`
	BackupInterval        *string  `flag:"backup-interval"`
	DataDir               string   `flag:"data-dir"`
	VolumeProvider        string   `flag:"volume-provider"`
	VolumeTag             []string `flag:"volume-tag,repeat"`
	VolumeNameTag         string   `flag:"volume-name-tag"`
	DNSSuffix             string   `flag:"dns-suffix"`
	NetworkCIDR           *string  `flag:"network-cidr"`

	// StaticConfig enables running with a fixed etcd cluster configuration.
	StaticConfig string `flag:"static-config"`
}

type StaticConfig struct {
	EtcdVersion string             `json:"etcdVersion,omitempty"`
	Nodes       []StaticConfigNode `json:"nodes,omitempty"`
}

type StaticConfigNode struct {
	ID string   `json:"id,omitempty"`
	IP []string `json:"ip,omitempty"`
}

// SelectorForCluster returns the selector that should be used to select our pods (from services)
func SelectorForCluster(etcdCluster kops.EtcdClusterSpec) map[string]string {
	return map[string]string{
		"k8s-app": "etcd-manager-" + etcdCluster.Name,
	}
}

type Ports struct {
	ClientPort          int
	PeerPort            int
	GRPCPort            int
	QuarantinedGRPCPort int
}

// PortsForCluster returns the ports that the cluster users.
func PortsForCluster(etcdCluster kops.EtcdClusterSpec) (Ports, error) {
	switch etcdCluster.Name {
	case "main":
		return Ports{
			GRPCPort: wellknownports.EtcdMainGRPC,
			// TODO: Use a socket file for the quarantine port
			QuarantinedGRPCPort: wellknownports.EtcdMainQuarantinedClientPort,
			ClientPort:          4001,
			PeerPort:            2380,
		}, nil

	case "events":
		return Ports{
			GRPCPort:            wellknownports.EtcdEventsGRPC,
			QuarantinedGRPCPort: wellknownports.EtcdEventsQuarantinedClientPort,
			ClientPort:          4002,
			PeerPort:            2381,
		}, nil
	case "cilium":
		return Ports{
			GRPCPort:            wellknownports.EtcdCiliumGRPC,
			QuarantinedGRPCPort: wellknownports.EtcdCiliumQuarantinedClientPort,
			ClientPort:          4003,
			PeerPort:            2382,
		}, nil

	default:
		return Ports{}, fmt.Errorf("unknown etcd cluster key %q", etcdCluster.Name)
	}
}
