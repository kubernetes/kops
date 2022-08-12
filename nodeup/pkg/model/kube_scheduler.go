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
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/configbuilder"
	"k8s.io/kops/pkg/flagbuilder"
	"k8s.io/kops/pkg/k8scodecs"
	"k8s.io/kops/pkg/kubemanifest"
	"k8s.io/kops/pkg/model/components"
	"k8s.io/kops/pkg/model/components/kubescheduler"
	"k8s.io/kops/pkg/rbac"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
	"k8s.io/kops/util/pkg/architectures"
	"k8s.io/kops/util/pkg/proxy"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// ClientConnectionConfig is used by kube-scheduler to talk to the api server
type ClientConnectionConfig struct {
	Burst      int32    `json:"burst,omitempty"`
	Kubeconfig string   `json:"kubeconfig"`
	QPS        *float64 `json:"qps,omitempty"`
}

// SchedulerConfig is used to generate the config file
type SchedulerConfig struct {
	APIVersion       string                 `json:"apiVersion"`
	Kind             string                 `json:"kind"`
	ClientConnection ClientConnectionConfig `json:"clientConnection,omitempty"`
}

// KubeSchedulerBuilder install kube-scheduler
type KubeSchedulerBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &KubeSchedulerBuilder{}

// Build is responsible for building the manifest for the kube-scheduler
func (b *KubeSchedulerBuilder) Build(c *fi.ModelBuilderContext) error {
	if !b.IsMaster {
		return nil
	}

	kubeScheduler := *b.Cluster.Spec.KubeScheduler

	if err := b.writeServerCertificate(c, &kubeScheduler); err != nil {
		return err
	}

	{
		pod, err := b.buildPod(&kubeScheduler)
		if err != nil {
			return fmt.Errorf("error building kube-scheduler pod: %v", err)
		}

		manifest, err := k8scodecs.ToVersionedYaml(pod)
		if err != nil {
			return fmt.Errorf("error marshaling pod to yaml: %v", err)
		}

		c.AddTask(&nodetasks.File{
			Path:     "/etc/kubernetes/manifests/kube-scheduler.manifest",
			Contents: fi.NewBytesResource(manifest),
			Type:     nodetasks.FileType_File,
		})
	}

	{
		kubeconfig := b.BuildIssuedKubeconfig("kube-scheduler", nodetasks.PKIXName{CommonName: rbac.KubeScheduler}, c)

		c.AddTask(&nodetasks.File{
			Path:     kubescheduler.KubeConfigPath,
			Contents: kubeconfig,
			Type:     nodetasks.FileType_File,
			Mode:     s("0400"),
		})
	}

	// Load the kube-scheduler config object if one has been provided.
	kubeSchedulerConfigAsset := b.findFileAsset(kubescheduler.KubeSchedulerConfigPath)

	if kubeSchedulerConfigAsset != nil {
		klog.Infof("using kubescheduler configuration from file assets")
		// FileAssets are written automatically, we don't need to write it.
	} else {
		// We didn't get a kubescheduler configuration; warn as we're aiming to move this to generation in the kops CLI
		klog.Warningf("using embedded kubescheduler configuration")
		var config *SchedulerConfig
		if b.IsKubernetesGTE("1.22") {
			config = NewSchedulerConfig("kubescheduler.config.k8s.io/v1beta2")
		} else {
			config = NewSchedulerConfig("kubescheduler.config.k8s.io/v1beta1")
		}

		kubeSchedulerConfig, err := configbuilder.BuildConfigYaml(&kubeScheduler, config)
		if err != nil {
			return err
		}
		c.AddTask(&nodetasks.File{
			Path:     kubescheduler.KubeSchedulerConfigPath,
			Contents: fi.NewBytesResource(kubeSchedulerConfig),
			Type:     nodetasks.FileType_File,
			Mode:     s("0400"),
		})
	}

	{
		c.AddTask(&nodetasks.File{
			Path:        "/var/log/kube-scheduler.log",
			Contents:    fi.NewStringResource(""),
			Type:        nodetasks.FileType_File,
			Mode:        s("0400"),
			IfNotExists: true,
		})
	}

	return nil
}

// NewSchedulerConfig initializes a new kube-scheduler config file
func NewSchedulerConfig(apiVersion string) *SchedulerConfig {
	schedConfig := new(SchedulerConfig)
	schedConfig.APIVersion = apiVersion
	schedConfig.Kind = "KubeSchedulerConfiguration"
	schedConfig.ClientConnection = ClientConnectionConfig{}
	schedConfig.ClientConnection.Kubeconfig = kubescheduler.KubeConfigPath
	return schedConfig
}

func (b *KubeSchedulerBuilder) writeServerCertificate(c *fi.ModelBuilderContext, kubeScheduler *kops.KubeSchedulerConfig) error {
	pathSrvScheduler := filepath.Join(b.PathSrvKubernetes(), "kube-scheduler")

	if kubeScheduler.TLSCertFile == nil {
		alternateNames := []string{
			"kube-scheduler.kube-system.svc." + b.Cluster.Spec.ClusterDNSDomain,
		}

		issueCert := &nodetasks.IssueCert{
			Name:           "kube-scheduler-server",
			Signer:         fi.CertificateIDCA,
			KeypairID:      b.NodeupConfig.KeypairIDs[fi.CertificateIDCA],
			Type:           "server",
			Subject:        nodetasks.PKIXName{CommonName: "kube-scheduler"},
			AlternateNames: alternateNames,
		}

		c.AddTask(issueCert)
		err := issueCert.AddFileTasks(c, pathSrvScheduler, "server", "", nil)
		if err != nil {
			return err
		}

		kubeScheduler.TLSCertFile = fi.String(filepath.Join(pathSrvScheduler, "server.crt"))
		kubeScheduler.TLSPrivateKeyFile = filepath.Join(pathSrvScheduler, "server.key")
	}

	return nil
}

// buildPod is responsible for constructing the pod specification
func (b *KubeSchedulerBuilder) buildPod(kubeScheduler *kops.KubeSchedulerConfig) (*v1.Pod, error) {
	pathSrvScheduler := filepath.Join(b.PathSrvKubernetes(), "kube-scheduler")

	flags, err := flagbuilder.BuildFlagsList(kubeScheduler)
	if err != nil {
		return nil, fmt.Errorf("error building kube-scheduler flags: %v", err)
	}

	flags = append(flags, "--config="+"/var/lib/kube-scheduler/config.yaml")

	// Add kubeconfig flags
	for _, flag := range []string{"authentication-", "authorization-"} {
		flags = append(flags, "--"+flag+"kubeconfig="+kubescheduler.KubeConfigPath)
	}

	if fi.BoolValue(kubeScheduler.UsePolicyConfigMap) {
		flags = append(flags, "--policy-configmap=scheduler-policy", "--policy-configmap-namespace=kube-system")
	}

	pod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kube-scheduler",
			Namespace: "kube-system",
			Labels: map[string]string{
				"k8s-app": "kube-scheduler",
			},
		},
		Spec: v1.PodSpec{
			HostNetwork: true,
		},
	}

	image := kubeScheduler.Image
	if components.IsBaseURL(b.Cluster.Spec.KubernetesVersion) && b.IsKubernetesLT("1.25") {
		image = strings.Replace(image, "registry.k8s.io", "k8s.gcr.io", 1)
	}
	if b.Architecture != architectures.ArchitectureAmd64 {
		image = strings.Replace(image, "-amd64", "-"+string(b.Architecture), 1)
	}

	healthAction := &v1.HTTPGetAction{
		Host: "127.0.0.1",
		Path: "/healthz",
		Port: intstr.FromInt(10251),
	}
	if b.IsKubernetesGTE("1.23") {
		healthAction.Port = intstr.FromInt(10259)
		healthAction.Scheme = v1.URISchemeHTTPS
	}

	container := &v1.Container{
		Name:  "kube-scheduler",
		Image: image,
		Env:   proxy.GetProxyEnvVars(b.Cluster.Spec.EgressProxy),
		LivenessProbe: &v1.Probe{
			ProbeHandler:        v1.ProbeHandler{HTTPGet: healthAction},
			InitialDelaySeconds: 15,
			TimeoutSeconds:      15,
		},
		Resources: v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU: resource.MustParse("100m"),
			},
		},
	}
	kubemanifest.AddHostPathMapping(pod, container, "varlibkubescheduler", "/var/lib/kube-scheduler")
	kubemanifest.AddHostPathMapping(pod, container, "srvscheduler", pathSrvScheduler)

	// Log both to docker and to the logfile
	kubemanifest.AddHostPathMapping(pod, container, "logfile", "/var/log/kube-scheduler.log").WithReadWrite()
	// We use lighter containers that don't include shells
	// But they have richer logging support via klog
	if b.IsKubernetesGTE("1.23") {
		container.Command = []string{"/go-runner"}
		container.Args = []string{
			"--log-file=/var/log/kube-scheduler.log",
			"--also-stdout",
			"/usr/local/bin/kube-scheduler",
		}
		container.Args = append(container.Args, sortedStrings(flags)...)
	} else {
		container.Command = []string{"/usr/local/bin/kube-scheduler"}
		if kubeScheduler.LogFormat != "" && kubeScheduler.LogFormat != "text" {
			// When logging-format is not text, some flags are not accepted.
			// https://github.com/kubernetes/kops/issues/14100
			container.Args = sortedStrings(flags)
		} else {
			container.Args = append(
				sortedStrings(flags),
				"--logtostderr=false", // https://github.com/kubernetes/klog/issues/60
				"--alsologtostderr",
				"--log-file=/var/log/kube-scheduler.log")
		}
	}

	if kubeScheduler.MaxPersistentVolumes != nil {
		maxPDV := v1.EnvVar{
			Name:  "KUBE_MAX_PD_VOLS", // https://kubernetes.io/docs/concepts/storage/storage-limits/
			Value: strconv.Itoa(int(*kubeScheduler.MaxPersistentVolumes)),
		}
		container.Env = append(container.Env, maxPDV)
	}

	pod.Spec.Containers = append(pod.Spec.Containers, *container)

	kubemanifest.MarkPodAsCritical(pod)
	kubemanifest.MarkPodAsClusterCritical(pod)

	return pod, nil
}
