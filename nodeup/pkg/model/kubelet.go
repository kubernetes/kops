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
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/golang/glog"
	"k8s.io/kops/nodeup/pkg/distros"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/flagbuilder"
	"k8s.io/kops/pkg/kubeconfig"
	"k8s.io/kops/pkg/systemd"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
	"k8s.io/kops/util/pkg/hashing"
)

// KubeletBuilder install kubelet
type KubeletBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &DockerBuilder{}

const socatStaticBinaryURL = "https://github.com/aledbf/socat-static-binary/releases/download/v0.0.1/socat-linux-amd64"
const socatStaticBinarySHA256 = "https://github.com/aledbf/socat-static-binary/releases/download/v0.0.1/socat-linux-amd64.sha256"

func (b *KubeletBuilder) Build(c *fi.ModelBuilderContext) error {
	kubeletConfig, err := b.buildKubeletConfig()
	if err != nil {
		return fmt.Errorf("error building kubelet config: %v", err)
	}

	// Add sysconfig file
	{
		// TODO: Dump this - just complexity!
		flags, err := flagbuilder.BuildFlags(kubeletConfig)
		if err != nil {
			return fmt.Errorf("error building kubelet flags: %v", err)
		}
		sysconfig := "DAEMON_ARGS=\"" + flags + "\"\n"

		t := &nodetasks.File{
			Path:     "/etc/sysconfig/kubelet",
			Contents: fi.NewStringResource(sysconfig),
			Type:     nodetasks.FileType_File,
		}
		c.AddTask(t)
	}

	// Add kubelet file itself (as an asset)
	{
		// TODO: Extract to common function?
		assetName := "kubelet"
		assetPath := ""
		asset, err := b.Assets.Find(assetName, assetPath)
		if err != nil {
			return fmt.Errorf("error trying to locate asset %q: %v", assetName, err)
		}
		if asset == nil {
			return fmt.Errorf("unable to locate asset %q", assetName)
		}

		t := &nodetasks.File{
			Path:     b.kubeletPath(),
			Contents: asset,
			Type:     nodetasks.FileType_File,
			Mode:     s("0755"),
		}
		c.AddTask(t)
	}

	// Add kubeconfig
	{
		kubeconfig, err := b.buildKubeconfig()
		if err != nil {
			return err
		}
		t := &nodetasks.File{
			Path:     "/var/lib/kubelet/kubeconfig",
			Contents: fi.NewStringResource(kubeconfig),
			Type:     nodetasks.FileType_File,
			Mode:     s("0400"),
		}
		c.AddTask(t)
	}

	if b.UsesCNI {
		t := &nodetasks.File{
			Path: "/etc/cni/net.d/",
			Type: nodetasks.FileType_Directory,
		}
		c.AddTask(t)
	}

	c.AddTask(b.buildSystemdService())

	if b.Distribution == distros.DistributionCoreOS {
		sha256URL := os.Getenv("SOCAT_SHA256_URL")
		if sha256URL == "" {
			sha256URL = socatStaticBinarySHA256
		}
		resp, err := http.Get(sha256URL)
		if err != nil {
			return fmt.Errorf("unexpected error downloading socat SHA256 from %v: %v", sha256URL, err)
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("unexpected error downloading socat SHA256 from %v: %v", sha256URL, err)
		}

		socatHash := strings.TrimSpace(string(body))
		hash, err := hashing.FromString(socatHash)
		if err != nil {
			return fmt.Errorf("unexpected error reading socat SHA: %s", err)
		}

		socatURL := os.Getenv("SOCAT_URL")
		if socatURL == "" {
			socatURL = socatStaticBinaryURL
		}

		_, err = fi.DownloadURL(socatURL, "/opt/kubernetes/bin/socat", hash)
		if err != nil {
			return fmt.Errorf("unexpected error downloading socat from %v: %v", socatURL, err)
		}
		err = os.Chmod("/opt/kubernetes/bin/socat", 0755)
		if err != nil {
			return fmt.Errorf("failed setting executable on socat binary: %s", err)
		}
	}

	return nil
}

func (b *KubeletBuilder) kubeletPath() string {
	kubeletCommand := "/usr/local/bin/kubelet"
	if b.Distribution == distros.DistributionCoreOS {
		kubeletCommand = "/opt/kubernetes/bin/kubelet"
	}
	return kubeletCommand
}

func (b *KubeletBuilder) buildSystemdService() *nodetasks.Service {
	kubeletCommand := b.kubeletPath()

	manifest := &systemd.Manifest{}
	manifest.Set("Unit", "Description", "Kubernetes Kubelet Server")
	manifest.Set("Unit", "Documentation", "https://github.com/kubernetes/kubernetes")
	manifest.Set("Unit", "After", "docker.service")

	if b.Distribution == distros.DistributionCoreOS {
		manifest.Set("Service", "Environment", "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/opt/kubernetes/bin")
	}

	manifest.Set("Service", "EnvironmentFile", "/etc/sysconfig/kubelet")
	manifest.Set("Service", "ExecStart", kubeletCommand+" \"$DAEMON_ARGS\"")
	manifest.Set("Service", "Restart", "always")
	manifest.Set("Service", "RestartSec", "2s")
	manifest.Set("Service", "StartLimitInterval", "0")
	manifest.Set("Service", "KillMode", "process")

	manifestString := manifest.Render()
	glog.V(8).Infof("Built service manifest %q\n%s", "docker", manifestString)

	service := &nodetasks.Service{
		Name:       "kubelet.service",
		Definition: s(manifestString),
	}

	// To avoid going in to backoff, we wait for protokube to start us
	service.Running = fi.Bool(false)

	service.InitDefaults()

	return service
}

func (b *KubeletBuilder) buildKubeconfig() (string, error) {
	caCertificate, err := b.KeyStore.Cert(fi.CertificateId_CA)
	if err != nil {
		return "", fmt.Errorf("error fetching CA certificate from keystore: %v", err)
	}

	kubeletCertificate, err := b.KeyStore.Cert("kubelet")
	if err != nil {
		return "", fmt.Errorf("error fetching kubelet certificate from keystore: %v", err)
	}
	kubeletPrivateKey, err := b.KeyStore.PrivateKey("kubelet")
	if err != nil {
		return "", fmt.Errorf("error fetching kubelet private key from keystore: %v", err)
	}

	user := kubeconfig.KubectlUser{}
	user.ClientCertificateData, err = kubeletCertificate.AsBytes()
	if err != nil {
		return "", fmt.Errorf("error encoding kubelet certificate: %v", err)
	}
	user.ClientKeyData, err = kubeletPrivateKey.AsBytes()
	if err != nil {
		return "", fmt.Errorf("error encoding kubelet private key: %v", err)
	}
	cluster := kubeconfig.KubectlCluster{}
	cluster.CertificateAuthorityData, err = caCertificate.AsBytes()
	if err != nil {
		return "", fmt.Errorf("error encoding CA certificate: %v", err)
	}

	config := &kubeconfig.KubectlConfig{
		ApiVersion: "v1",
		Kind:       "Config",
		Users: []*kubeconfig.KubectlUserWithName{
			{
				Name: "kubelet",
				User: user,
			},
		},
		Clusters: []*kubeconfig.KubectlClusterWithName{
			{
				Name:    "local",
				Cluster: cluster,
			},
		},
		Contexts: []*kubeconfig.KubectlContextWithName{
			{
				Name: "service-account-context",
				Context: kubeconfig.KubectlContext{
					Cluster: "local",
					User:    "kubelet",
				},
			},
		},
		CurrentContext: "service-account-context",
	}

	yaml, err := kops.ToRawYaml(config)
	if err != nil {
		return "", fmt.Errorf("error marshalling kubeconfig to yaml: %v", err)
	}

	return string(yaml), nil
}

func (b *KubeletBuilder) buildKubeletConfig() (*kops.KubeletConfigSpec, error) {
	instanceGroup := b.InstanceGroup
	if instanceGroup == nil {
		// Old clusters might not have exported instance groups
		// in that case we build a synthetic instance group with the information that BuildKubeletConfigSpec needs
		// TODO: Remove this once we have a stable release
		glog.Warningf("Building a synthetic instance group")
		instanceGroup = &kops.InstanceGroup{}
		instanceGroup.ObjectMeta.Name = "synthetic"
		if b.IsMaster {
			instanceGroup.Spec.Role = kops.InstanceGroupRoleMaster
		} else {
			instanceGroup.Spec.Role = kops.InstanceGroupRoleNode
		}
		//b.InstanceGroup = instanceGroup
	}
	kubeletConfigSpec, err := kops.BuildKubeletConfigSpec(b.Cluster, instanceGroup)
	if err != nil {
		return nil, fmt.Errorf("error building kubelet config: %v", err)
	}
	// TODO: Memoize if we reuse this
	return kubeletConfigSpec, nil

}
