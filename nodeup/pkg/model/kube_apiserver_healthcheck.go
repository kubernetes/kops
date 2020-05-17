/*
Copyright 2020 The Kubernetes Authors.

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
	"crypto/x509/pkix"
	"fmt"
	"path/filepath"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog"
	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/pkg/wellknownusers"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
	"sigs.k8s.io/yaml"
)

func (b *KubeAPIServerBuilder) findHealthcheckManifest() *nodeup.StaticManifest {
	if b.NodeupConfig == nil {
		return nil
	}
	for _, manifest := range b.NodeupConfig.StaticManifests {
		if manifest.Key == "kube-apiserver-healthcheck" {
			return manifest
		}
	}
	return nil
}

func (b *KubeAPIServerBuilder) addHealthcheckSidecar(pod *corev1.Pod) error {
	manifest := b.findHealthcheckManifest()
	if manifest == nil {
		return nil
	}

	p := b.ConfigBase.Join(manifest.Path)

	data, err := p.ReadFile()
	if err != nil {
		return fmt.Errorf("error reading kube-apiserver-healthcheck manifest %s: %v", manifest.Path, err)
	}

	sidecar := &corev1.Pod{}
	if err := yaml.Unmarshal(data, sidecar); err != nil {
		return fmt.Errorf("error parsing kube-apiserver-healthcheck manifest %s: %v", manifest.Path, err)
	}

	// Quick-and-dirty merge of the fields we care about
	pod.Spec.Containers = append(pod.Spec.Containers, sidecar.Spec.Containers...)
	pod.Spec.Volumes = append(pod.Spec.Volumes, sidecar.Spec.Volumes...)

	return nil
}

func (b *KubeAPIServerBuilder) addHealthcheckSidecarTasks(c *fi.ModelBuilderContext) error {
	id := "kube-apiserver-healthcheck"
	secretsDir := "/etc/kubernetes/" + id + "/secrets"
	userID := wellknownusers.KubeApiserverHealthcheckID
	userName := wellknownusers.KubeApiserverHealthcheckName

	// We create user a user and hardcode its UID to 10012 as
	// that is the ID used inside the container.
	{
		c.AddTask(&nodetasks.UserTask{
			Name:  userName,
			UID:   userID,
			Shell: "/sbin/nologin",
			Home:  secretsDir,
		})
	}

	req := &pki.IssueCertRequest{
		Signer: fi.CertificateId_CA,
		Type:   "client",
		Subject: pkix.Name{
			CommonName: id,
		},
		MinValidDays: 455,
	}

	klog.Infof("signing certificate for %q", id)
	clientCert, clientKey, _, err := pki.IssueCert(req, b.KeyStore)
	if err != nil {
		return err
	}

	c.AddTask(&nodetasks.File{
		Path: filepath.Join(secretsDir),
		Type: nodetasks.FileType_Directory,
		Mode: s("0755"),
	})

	clientCertBytes, err := clientCert.AsBytes()
	if err != nil {
		return err
	}
	c.AddTask(&nodetasks.File{
		Path:     filepath.Join(secretsDir, "client.crt"),
		Contents: fi.NewBytesResource(clientCertBytes),
		Type:     nodetasks.FileType_File,
		Mode:     s("0644"),
		Owner:    s(userName),
	})

	clientKeyBytes, err := clientKey.AsBytes()
	if err != nil {
		return err
	}
	c.AddTask(&nodetasks.File{
		Path:     filepath.Join(secretsDir, "client.key"),
		Contents: fi.NewBytesResource(clientKeyBytes),
		Type:     nodetasks.FileType_File,
		Mode:     s("0600"),
		Owner:    s(userName),
	})

	cert, err := b.GetCert(fi.CertificateId_CA)
	if err != nil {
		return err
	}

	c.AddTask(&nodetasks.File{
		Path:     filepath.Join(secretsDir, "ca.crt"),
		Contents: fi.NewBytesResource(cert),
		Type:     nodetasks.FileType_File,
		Mode:     s("0644"),
		Owner:    s(userName),
	})

	return nil
}
