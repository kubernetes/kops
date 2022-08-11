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
	"strings"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/flagbuilder"
	"k8s.io/kops/pkg/k8scodecs"
	"k8s.io/kops/pkg/kubemanifest"
	"k8s.io/kops/pkg/model/components"
	"k8s.io/kops/pkg/rbac"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
	"k8s.io/kops/util/pkg/architectures"
	"k8s.io/kops/util/pkg/distributions"
	"k8s.io/kops/util/pkg/proxy"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// KubeControllerManagerBuilder install kube-controller-manager (just the manifest at the moment)
type KubeControllerManagerBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &KubeControllerManagerBuilder{}

// Build is responsible for configuring the kube-controller-manager
func (b *KubeControllerManagerBuilder) Build(c *fi.ModelBuilderContext) error {
	if !b.IsMaster {
		return nil
	}

	pathSrvKCM := filepath.Join(b.PathSrvKubernetes(), "kube-controller-manager")

	kcm := *b.Cluster.Spec.KubeControllerManager
	kcm.RootCAFile = filepath.Join(b.PathSrvKubernetes(), "ca.crt")

	// Include the CA Key
	// @TODO: use a per-machine key?  use KMS?
	if err := b.BuildCertificatePairTask(c, fi.CertificateIDCA, pathSrvKCM, "ca", nil, nil); err != nil {
		return err
	}

	if err := b.BuildPrivateKeyTask(c, "service-account", pathSrvKCM, "service-account", nil, nil); err != nil {
		return err
	}
	kcm.ServiceAccountPrivateKeyFile = filepath.Join(pathSrvKCM, "service-account.key")

	if err := b.writeServerCertificate(c, &kcm); err != nil {
		return err
	}

	{
		pod, err := b.buildPod(&kcm)
		if err != nil {
			return fmt.Errorf("error building kube-controller-manager pod: %v", err)
		}

		manifest, err := k8scodecs.ToVersionedYaml(pod)
		if err != nil {
			return fmt.Errorf("error marshaling pod to yaml: %v", err)
		}

		c.AddTask(&nodetasks.File{
			Path:     "/etc/kubernetes/manifests/kube-controller-manager.manifest",
			Contents: fi.NewBytesResource(manifest),
			Type:     nodetasks.FileType_File,
		})
	}

	{
		c.AddTask(&nodetasks.File{
			Path:        "/var/log/kube-controller-manager.log",
			Contents:    fi.NewStringResource(""),
			Type:        nodetasks.FileType_File,
			Mode:        s("0400"),
			IfNotExists: true,
		})
	}

	// Add kubeconfig
	{
		kubeconfig := b.BuildIssuedKubeconfig("kube-controller-manager", nodetasks.PKIXName{CommonName: rbac.KubeControllerManager}, c)
		c.AddTask(&nodetasks.File{
			Path:     "/var/lib/kube-controller-manager/kubeconfig",
			Contents: kubeconfig,
			Type:     nodetasks.FileType_File,
			Mode:     s("0400"),
		})
	}

	return nil
}

func (b *KubeControllerManagerBuilder) writeServerCertificate(c *fi.ModelBuilderContext, kcm *kops.KubeControllerManagerConfig) error {
	pathSrvKCM := filepath.Join(b.PathSrvKubernetes(), "kube-controller-manager")

	if kcm.TLSCertFile == nil {
		alternateNames := []string{
			"kube-controller-manager.kube-system.svc." + b.Cluster.Spec.ClusterDNSDomain,
		}

		issueCert := &nodetasks.IssueCert{
			Name:           "kube-controller-manager-server",
			Signer:         fi.CertificateIDCA,
			KeypairID:      b.NodeupConfig.KeypairIDs[fi.CertificateIDCA],
			Type:           "server",
			Subject:        nodetasks.PKIXName{CommonName: "kube-controller-manager"},
			AlternateNames: alternateNames,
		}

		c.AddTask(issueCert)
		err := issueCert.AddFileTasks(c, pathSrvKCM, "server", "", nil)
		if err != nil {
			return err
		}

		kcm.TLSCertFile = fi.String(filepath.Join(pathSrvKCM, "server.crt"))
		kcm.TLSPrivateKeyFile = filepath.Join(pathSrvKCM, "server.key")
	}

	return nil
}

// buildPod is responsible for building the kubernetes manifest for the controller-manager
func (b *KubeControllerManagerBuilder) buildPod(kcm *kops.KubeControllerManagerConfig) (*v1.Pod, error) {
	pathSrvKCM := filepath.Join(b.PathSrvKubernetes(), "kube-controller-manager")

	flags, err := flagbuilder.BuildFlagsList(kcm)
	if err != nil {
		return nil, fmt.Errorf("error building kube-controller-manager flags: %v", err)
	}

	// Add cloud config file if needed
	if b.Cluster.Spec.CloudConfig != nil {
		flags = append(flags, "--cloud-config="+InTreeCloudConfigFilePath)
	}

	// Add kubeconfig flags
	for _, flag := range []string{"", "authentication-", "authorization-"} {
		flags = append(flags, "--"+flag+"kubeconfig="+"/var/lib/kube-controller-manager/kubeconfig")
	}

	// Configure CA certificate to be used to sign keys
	flags = append(flags, []string{
		"--cluster-signing-cert-file=" + filepath.Join(pathSrvKCM, "ca.crt"),
		"--cluster-signing-key-file=" + filepath.Join(pathSrvKCM, "ca.key"),
	}...)

	pod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kube-controller-manager",
			Namespace: "kube-system",
			Labels: map[string]string{
				"k8s-app": "kube-controller-manager",
			},
		},
		Spec: v1.PodSpec{
			HostNetwork: true,
		},
	}

	volumePluginDir := b.Cluster.Spec.Kubelet.VolumePluginDirectory

	// Ensure the Volume Plugin dir is mounted on the same path as the host machine so DaemonSet deployment is possible
	if volumePluginDir == "" {
		switch b.Distribution {
		case distributions.DistributionContainerOS:
			// Default is different on ContainerOS, see https://github.com/kubernetes/kubernetes/pull/58171
			volumePluginDir = "/home/kubernetes/flexvolume/"

		case distributions.DistributionFlatcar:
			// The /usr directory is read-only for Flatcar
			volumePluginDir = "/var/lib/kubelet/volumeplugins/"

		default:
			volumePluginDir = "/usr/libexec/kubernetes/kubelet-plugins/volume/exec/"
		}
	}

	// Add the volumePluginDir flag if provided in the kubelet spec, or set above based on the OS
	flags = append(flags, "--flex-volume-plugin-dir="+volumePluginDir)

	image := kcm.Image
	if components.IsBaseURL(b.Cluster.Spec.KubernetesVersion) && b.IsKubernetesLT("1.25") {
		image = strings.Replace(image, "registry.k8s.io", "k8s.gcr.io", 1)
	}
	if b.Architecture != architectures.ArchitectureAmd64 {
		image = strings.Replace(image, "-amd64", "-"+string(b.Architecture), 1)
	}

	container := &v1.Container{
		Name:  "kube-controller-manager",
		Image: image,
		Env:   proxy.GetProxyEnvVars(b.Cluster.Spec.EgressProxy),
		LivenessProbe: &v1.Probe{
			ProbeHandler: v1.ProbeHandler{
				HTTPGet: &v1.HTTPGetAction{
					Host:   "127.0.0.1",
					Path:   "/healthz",
					Port:   intstr.FromInt(10257),
					Scheme: "HTTPS",
				},
			},
			InitialDelaySeconds: 15,
			TimeoutSeconds:      15,
		},
		Resources: v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU: resource.MustParse("100m"),
			},
		},
	}

	// Log both to docker and to the logfile
	kubemanifest.AddHostPathMapping(pod, container, "logfile", "/var/log/kube-controller-manager.log").WithReadWrite()
	// We use lighter containers that don't include shells
	// But they have richer logging support via klog
	if b.IsKubernetesGTE("1.23") {
		container.Command = []string{"/go-runner"}
		container.Args = []string{
			"--log-file=/var/log/kube-controller-manager.log",
			"--also-stdout",
			"/usr/local/bin/kube-controller-manager",
		}
		container.Args = append(container.Args, sortedStrings(flags)...)
	} else {
		container.Command = []string{"/usr/local/bin/kube-controller-manager"}
		if kcm.LogFormat != "" && kcm.LogFormat != "text" {
			// When logging-format is not text, some flags are not accepted.
			// https://github.com/kubernetes/kops/issues/14100
			container.Args = sortedStrings(flags)
		} else {
			container.Args = append(
				sortedStrings(flags),
				"--logtostderr=false", // https://github.com/kubernetes/klog/issues/60
				"--alsologtostderr",
				"--log-file=/var/log/kube-controller-manager.log")
		}
	}
	for _, path := range b.SSLHostPaths() {
		name := strings.Replace(path, "/", "", -1)
		kubemanifest.AddHostPathMapping(pod, container, name, path)
	}

	// Add cloud config file if needed
	if b.Cluster.Spec.CloudConfig != nil {
		kubemanifest.AddHostPathMapping(pod, container, "cloudconfig", InTreeCloudConfigFilePath)
	}

	kubemanifest.AddHostPathMapping(pod, container, "cabundle", filepath.Join(b.PathSrvKubernetes(), "ca.crt"))

	kubemanifest.AddHostPathMapping(pod, container, "srvkcm", pathSrvKCM)

	kubemanifest.AddHostPathMapping(pod, container, "varlibkcm", "/var/lib/kube-controller-manager")

	kubemanifest.AddHostPathMapping(pod, container, "volplugins", volumePluginDir).WithReadWrite()

	pod.Spec.Containers = append(pod.Spec.Containers, *container)

	kubemanifest.MarkPodAsCritical(pod)
	kubemanifest.MarkPodAsClusterCritical(pod)

	return pod, nil
}
