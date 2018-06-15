/*
Copyright 2018 The Kubernetes Authors.

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
	"strings"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/dns"
	"k8s.io/kops/pkg/flagbuilder"
	"k8s.io/kops/pkg/k8scodecs"
	"k8s.io/kops/pkg/kubemanifest"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/fitasks"
	"k8s.io/kops/util/pkg/exec"
)

const metaFilename = "_etcd_backup.meta"

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
		if etcdCluster.Manager == nil {
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
			return fmt.Errorf("error marshalling manifest to yaml: %v", err)
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

		c.AddTask(&fitasks.ManagedFile{
			Contents:  fi.WrapResource(fi.NewBytesResource(d)),
			Lifecycle: b.Lifecycle,
			// TODO: We need this to match the backup base (currently)
			Location: fi.String("backups/etcd/" + etcdCluster.Name + "/control/etcd-cluster-spec"),
			Name:     fi.String("etcd-cluster-spec-" + name),
		})
	}

	return nil
}

type etcdClusterSpec struct {
	MemberCount int32  `json:"member_count,omitempty"`
	EtcdVersion string `json:"etcd_version,omitempty"`
}

func (b *EtcdManagerBuilder) buildManifest(etcdCluster *kops.EtcdClusterSpec) (*v1.Pod, error) {
	if etcdCluster.Manager == nil {
		return nil, fmt.Errorf("manager not set for EtcdCluster")
	}

	return b.buildPod(etcdCluster)
}

// BuildEtcdManifest creates the pod spec, based on the etcd cluster
func (b *EtcdManagerBuilder) buildPod(etcdCluster *kops.EtcdClusterSpec) (*v1.Pod, error) {
	image := etcdCluster.Manager.Image
	{
		remapped, err := b.AssetBuilder.RemapImage(image)
		if err != nil {
			return nil, fmt.Errorf("unable to remap container %q: %v", image, err)
		} else {
			image = remapped
		}
	}

	isTLS := etcdCluster.EnableEtcdTLS

	cpuRequest := resource.MustParse("100m")
	clientPort := 4001

	clusterName := "etcd-" + etcdCluster.Name
	peerPort := 2380
	backupStore := ""
	if etcdCluster.Backups != nil {
		backupStore = etcdCluster.Backups.BackupStore
	}

	podName := "etcd-manager-" + etcdCluster.Name

	// TODO: Use a socket file for the quarantine port
	quarantinedClientPort := 3994

	grpcPort := 3996

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
		cpuRequest = resource.MustParse("200m")

	case "events":
		clientPort = 4002
		peerPort = 2381
		grpcPort = 3997
		quarantinedClientPort = 3995

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

	config := &EtcdManagerConfig{
		Containerized: true,
		ClusterName:   clusterName,
		BackupStore:   backupStore,
		GrpcPort:      grpcPort,
		DNSSuffix:     dnsInternalSuffix,
	}

	config.LogVerbosity = 8

	var envs []v1.EnvVar

	{
		// @check if we are using TLS
		scheme := "http"
		if isTLS {
			scheme = "https"
		}

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

		if isTLS {
			return nil, fmt.Errorf("TLS not supported for etcd-manager")
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

		default:
			return nil, fmt.Errorf("CloudProvider %q not supported with etcd-manager", b.Cluster.Spec.CloudProvider)
		}
	}

	args, err := flagbuilder.BuildFlagsList(config)
	if err != nil {
		return nil, err
	}

	pod := &v1.Pod{}
	pod.APIVersion = "v1"
	pod.Kind = "Pod"
	pod.Name = podName
	pod.Namespace = "kube-system"
	pod.Labels = map[string]string{"k8s-app": podName}
	pod.Spec.HostNetwork = true

	{
		container := &v1.Container{
			Name:  "etcd-manager",
			Image: image,
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceCPU: cpuRequest,
				},
			},
			Command: exec.WithTee("/etcd-manager", args, "/var/log/etcd.log"),
			Env:     envs,
		}

		// TODO: Reduce these permissions (they are needed for volume mounting)
		container.SecurityContext = &v1.SecurityContext{
			Privileged: fi.Bool(true),
		}

		// TODO: Use helper function here
		container.VolumeMounts = append(container.VolumeMounts, v1.VolumeMount{
			Name:      "varlogetcd",
			MountPath: "/var/log/etcd.log",
			ReadOnly:  false,
		})

		// TODO: Would be nice to narrow this mount
		container.VolumeMounts = append(container.VolumeMounts, v1.VolumeMount{
			Name:      "rootfs",
			MountPath: "/rootfs",
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

		hostPathDirectory := v1.HostPathDirectory
		pod.Spec.Volumes = append(pod.Spec.Volumes, v1.Volume{
			Name: "rootfs",
			VolumeSource: v1.VolumeSource{
				HostPath: &v1.HostPathVolumeSource{
					Path: "/",
					Type: &hostPathDirectory,
				},
			},
		})

		kubemanifest.MapEtcHosts(pod, container, false)

		if isTLS {
			return nil, fmt.Errorf("TLS not supported for etcd-manager")
		}

		pod.Spec.Containers = append(pod.Spec.Containers, *container)
	}

	kubemanifest.MarkPodAsCritical(pod)

	return pod, nil
}

// EtcdManagerConfig are the flags for etcd-manager
type EtcdManagerConfig struct {
	// LogVerbosity sets the log verbosity level
	LogVerbosity int `flag:"v"`

	// Containerized is set if etcd-manager is running in a container
	Containerized bool `flag:"containerized"`

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
