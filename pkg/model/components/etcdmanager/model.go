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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
	scheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/dns"
	"k8s.io/kops/pkg/flagbuilder"
	"k8s.io/kops/pkg/k8scodecs"
	"k8s.io/kops/pkg/kubemanifest"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/wellknownports"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/do"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
	"k8s.io/kops/upup/pkg/fi/fitasks"
	"k8s.io/kops/util/pkg/env"
	"k8s.io/kops/util/pkg/exec"
)

// EtcdManagerBuilder builds the manifest for the etcd-manager
type EtcdManagerBuilder struct {
	*model.KopsModelContext
	Lifecycle    *fi.Lifecycle
	AssetBuilder *assets.AssetBuilder
}

var _ fi.ModelBuilder = &EtcdManagerBuilder{}

// Build creates the tasks
func (b *EtcdManagerBuilder) Build(c *fi.ModelBuilderContext) error {
	for _, etcdCluster := range b.Cluster.Spec.EtcdClusters {
		if etcdCluster.Provider != kops.EtcdProviderTypeManager {
			continue
		}

		name := etcdCluster.Name
		version := etcdCluster.Version

		backupStore := ""
		if etcdCluster.Backups != nil {
			backupStore = etcdCluster.Backups.BackupStore
		}
		if backupStore == "" {
			return fmt.Errorf("backupStore must be set for use with etcd-manager")
		}

		manifest, err := b.buildManifest(etcdCluster)
		if err != nil {
			return err
		}

		manifestYAML, err := k8scodecs.ToVersionedYaml(manifest)
		if err != nil {
			return fmt.Errorf("error marshaling manifest to yaml: %v", err)
		}

		c.AddTask(&fitasks.ManagedFile{
			Contents:  fi.WrapResource(fi.NewBytesResource(manifestYAML)),
			Lifecycle: b.Lifecycle,
			Location:  fi.String("manifests/etcd/" + name + ".yaml"),
			Name:      fi.String("manifests-etcdmanager-" + name),
		})

		info := &etcdClusterSpec{
			EtcdVersion: version,
			MemberCount: int32(len(etcdCluster.Members)),
		}

		d, err := json.MarshalIndent(info, "", "  ")
		if err != nil {
			return err
		}

		format := string(fi.KeysetFormatV1Alpha2)

		c.AddTask(&fitasks.ManagedFile{
			Contents:  fi.WrapResource(fi.NewBytesResource(d)),
			Lifecycle: b.Lifecycle,
			// TODO: We need this to match the backup base (currently)
			Location: fi.String("backups/etcd/" + etcdCluster.Name + "/control/etcd-cluster-spec"),
			Name:     fi.String("etcd-cluster-spec-" + name),
		})

		// We create a CA keypair to enable secure communication
		c.AddTask(&fitasks.Keypair{
			Name:    fi.String("etcd-manager-ca-" + etcdCluster.Name),
			Subject: "cn=etcd-manager-ca-" + etcdCluster.Name,
			Type:    "ca",
			Format:  format,
		})

		// We create a CA for etcd peers and a separate one for clients
		c.AddTask(&fitasks.Keypair{
			Name:    fi.String("etcd-peers-ca-" + etcdCluster.Name),
			Subject: "cn=etcd-peers-ca-" + etcdCluster.Name,
			Type:    "ca",
			Format:  format,
		})

		// Because API server can only have a single client-cert, we need to share a client CA
		if err := c.EnsureTask(&fitasks.Keypair{
			Name:    fi.String("etcd-clients-ca"),
			Subject: "cn=etcd-clients-ca",
			Type:    "ca",
			Format:  format,
		}); err != nil {
			return err
		}
	}

	return nil
}

type etcdClusterSpec struct {
	MemberCount int32  `json:"memberCount,omitempty"`
	EtcdVersion string `json:"etcdVersion,omitempty"`
}

func (b *EtcdManagerBuilder) buildManifest(etcdCluster *kops.EtcdClusterSpec) (*v1.Pod, error) {
	return b.buildPod(etcdCluster)
}

// parseManifest parses a set of objects from a []byte
func parseManifest(data []byte) ([]runtime.Object, error) {
	decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(data), 4096)
	deser := scheme.Codecs.UniversalDeserializer()

	var objects []runtime.Object

	for {
		ext := runtime.RawExtension{}
		if err := decoder.Decode(&ext); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "%s", string(data))
			klog.Infof("manifest: %s", string(data))
			return nil, fmt.Errorf("error parsing manifest: %v", err)
		}

		obj, _, err := deser.Decode([]byte(ext.Raw), nil, nil)
		if err != nil {
			return nil, fmt.Errorf("error parsing object in manifest: %v", err)
		}

		objects = append(objects, obj)
	}

	return objects, nil
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
  - image: kopeio/etcd-manager:3.0.20200116
    name: etcd-manager
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
    - mountPath: /etc/kubernetes/pki/etcd-manager
      name: pki
  hostNetwork: true
  hostPID: true # helps with mounting volumes from inside a container
  volumes:
  - hostPath:
      path: /
      type: Directory
    name: rootfs
  - hostPath:
      path: /etc/kubernetes/pki/etcd-manager
      type: DirectoryOrCreate
    name: pki
`

// buildPod creates the pod spec, based on the EtcdClusterSpec
func (b *EtcdManagerBuilder) buildPod(etcdCluster *kops.EtcdClusterSpec) (*v1.Pod, error) {
	var pod *v1.Pod
	var container *v1.Container

	var manifest []byte

	// TODO: pull from bundle
	bundle := "(embedded etcd manifest)"
	manifest = []byte(defaultManifest)

	{
		objects, err := parseManifest(manifest)
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

		if len(pod.Spec.Containers) != 1 {
			return nil, fmt.Errorf("expected exactly one container in etcd-manager Pod, found %d", len(pod.Spec.Containers))
		}
		container = &pod.Spec.Containers[0]

		if etcdCluster.Manager != nil && etcdCluster.Manager.Image != "" {
			klog.Warningf("overloading image in manifest %s with images %s", bundle, etcdCluster.Manager.Image)
			container.Image = etcdCluster.Manager.Image
		}
	}

	// With etcd-manager the hosts changes are self-contained, so
	// we don't need to share /etc/hosts.  By not sharing we avoid
	// (1) the temptation to address etcd directly and (2)
	// problems of concurrent updates to /etc/hosts being hard
	// from within a container (because locking is very difficult
	// across bind mounts).
	//
	// Introduced with 1.17 to avoid changing existing versions.
	if b.IsKubernetesLT("1.17") {
		kubemanifest.MapEtcHosts(pod, container, false)
	}

	// Remap image via AssetBuilder
	{
		remapped, err := b.AssetBuilder.RemapImage(container.Image)
		if err != nil {
			return nil, fmt.Errorf("unable to remap container image %q: %v", container.Image, err)
		}
		container.Image = remapped
	}

	etcdInsecure := !b.UseEtcdTLS()

	clientPort := 4001

	clusterName := "etcd-" + etcdCluster.Name
	peerPort := 2380
	backupStore := ""
	if etcdCluster.Backups != nil {
		backupStore = etcdCluster.Backups.BackupStore
	}

	pod.Name = "etcd-manager-" + etcdCluster.Name
	if pod.Labels == nil {
		pod.Labels = make(map[string]string)
	}
	pod.Labels["k8s-app"] = pod.Name

	// TODO: Use a socket file for the quarantine port
	quarantinedClientPort := wellknownports.EtcdMainQuarantinedClientPort

	grpcPort := wellknownports.EtcdMainGRPC

	// The dns suffix logic mirrors the existing logic, so we should be compatible with existing clusters
	// (etcd makes it difficult to change peer urls, treating it as a cluster event, for reasons unknown)
	dnsInternalSuffix := ""
	if dns.IsGossipHostname(b.Cluster.Spec.MasterInternalName) {
		// @TODO: This is hacky, but we want it so that we can have a different internal & external name
		dnsInternalSuffix = b.Cluster.Spec.MasterInternalName
		dnsInternalSuffix = strings.TrimPrefix(dnsInternalSuffix, "api.")
	}

	if dnsInternalSuffix == "" {
		dnsInternalSuffix = ".internal." + b.Cluster.ObjectMeta.Name
	}

	switch etcdCluster.Name {
	case "main":
		clusterName = "etcd"

	case "events":
		clientPort = 4002
		peerPort = 2381
		grpcPort = wellknownports.EtcdEventsGRPC
		quarantinedClientPort = wellknownports.EtcdEventsQuarantinedClientPort

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
		GrpcPort:      grpcPort,
		DNSSuffix:     dnsInternalSuffix,
		EtcdInsecure:  etcdInsecure,
	}

	config.LogVerbosity = 6

	{
		scheme := "https"

		config.PeerUrls = fmt.Sprintf("%s://__name__:%d", scheme, peerPort)
		config.ClientUrls = fmt.Sprintf("%s://__name__:%d", scheme, clientPort)
		config.QuarantineClientUrls = fmt.Sprintf("%s://__name__:%d", scheme, quarantinedClientPort)

		// TODO: We need to wire these into the etcd-manager spec
		// // add timeout/heartbeat settings
		if etcdCluster.LeaderElectionTimeout != nil {
			// 	envs = append(envs, v1.EnvVar{Name: "ETCD_ELECTION_TIMEOUT", Value: convEtcdSettingsToMs(etcdClusterSpec.LeaderElectionTimeout)})
			return nil, fmt.Errorf("LeaderElectionTimeout not supported by etcd-manager")
		}
		if etcdCluster.HeartbeatInterval != nil {
			// 	envs = append(envs, v1.EnvVar{Name: "ETCD_HEARTBEAT_INTERVAL", Value: convEtcdSettingsToMs(etcdClusterSpec.HeartbeatInterval)})
			return nil, fmt.Errorf("HeartbeatInterval not supported by etcd-manager")
		}
	}

	{
		switch kops.CloudProviderID(b.Cluster.Spec.CloudProvider) {
		case kops.CloudProviderAWS:
			config.VolumeProvider = "aws"

			config.VolumeTag = []string{
				fmt.Sprintf("kubernetes.io/cluster/%s=owned", b.Cluster.Name),
				awsup.TagNameEtcdClusterPrefix + etcdCluster.Name,
				awsup.TagNameRolePrefix + "master=1",
			}
			config.VolumeNameTag = awsup.TagNameEtcdClusterPrefix + etcdCluster.Name

		case kops.CloudProviderGCE:
			config.VolumeProvider = "gce"

			config.VolumeTag = []string{
				gce.GceLabelNameKubernetesCluster + "=" + gce.SafeClusterName(b.Cluster.Name),
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

		case kops.CloudProviderOpenstack:
			config.VolumeProvider = "openstack"

			config.VolumeTag = []string{
				openstack.TagNameEtcdClusterPrefix + etcdCluster.Name,
				openstack.TagNameRolePrefix + "master=1",
				fmt.Sprintf("%s=%s", openstack.TagClusterName, b.Cluster.Name),
			}
			config.VolumeNameTag = openstack.TagNameEtcdClusterPrefix + etcdCluster.Name

		default:
			return nil, fmt.Errorf("CloudProvider %q not supported with etcd-manager", b.Cluster.Spec.CloudProvider)
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

		// TODO: Use helper function here
		container.VolumeMounts = append(container.VolumeMounts, v1.VolumeMount{
			Name:      "varlogetcd",
			MountPath: "/var/log/etcd.log",
			ReadOnly:  false,
		})
		hostPathFileOrCreate := v1.HostPathFileOrCreate
		pod.Spec.Volumes = append(pod.Spec.Volumes, v1.Volume{
			Name: "varlogetcd",
			VolumeSource: v1.VolumeSource{
				HostPath: &v1.HostPathVolumeSource{
					Path: logFile,
					Type: &hostPathFileOrCreate,
				},
			},
		})

		if fi.BoolValue(b.Cluster.Spec.UseHostCertificates) {
			container.VolumeMounts = append(container.VolumeMounts, v1.VolumeMount{
				Name:      "etc-ssl-certs",
				MountPath: "/etc/ssl/certs",
				ReadOnly:  true,
			})
			hostPathDirectoryOrCreate := v1.HostPathDirectoryOrCreate
			pod.Spec.Volumes = append(pod.Spec.Volumes, v1.Volume{
				Name: "etc-ssl-certs",
				VolumeSource: v1.VolumeSource{
					HostPath: &v1.HostPathVolumeSource{
						Path: "/etc/ssl/certs",
						Type: &hostPathDirectoryOrCreate,
					},
				},
			})
		}
	}

	envMap := env.BuildSystemComponentEnvVars(&b.Cluster.Spec)

	container.Env = envMap.ToEnvVars()

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
	// LogVerbosity sets the log verbosity level
	LogVerbosity int `flag:"v"`

	// Containerized is set if etcd-manager is running in a container
	Containerized bool `flag:"containerized"`

	// PKIDir is set to the directory for PKI keys, used to secure commucations between etcd-manager peers
	PKIDir string `flag:"pki-dir"`

	// Insecure can be used to turn off tls for etcd-manager (compare with EtcdInsecure)
	Insecure bool `flag:"insecure"`

	// EtcdInsecure can be used to turn off tls for etcd itself (compare with Insecure)
	EtcdInsecure bool `flag:"etcd-insecure"`

	Address              string   `flag:"address"`
	PeerUrls             string   `flag:"peer-urls"`
	GrpcPort             int      `flag:"grpc-port"`
	ClientUrls           string   `flag:"client-urls"`
	QuarantineClientUrls string   `flag:"quarantine-client-urls"`
	ClusterName          string   `flag:"cluster-name"`
	BackupStore          string   `flag:"backup-store"`
	DataDir              string   `flag:"data-dir"`
	VolumeProvider       string   `flag:"volume-provider"`
	VolumeTag            []string `flag:"volume-tag,repeat"`
	VolumeNameTag        string   `flag:"volume-name-tag"`
	DNSSuffix            string   `flag:"dns-suffix"`
}
