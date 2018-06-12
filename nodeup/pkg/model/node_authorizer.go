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

package model

import (
	"bytes"
	"fmt"
	"path"
	"path/filepath"
	"text/template"
	"time"

	"k8s.io/kops/pkg/systemd"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"

	"github.com/golang/glog"
)

// NodeAuthorizationBuilder is responsible for node authorization
type NodeAuthorizationBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &NodeAuthorizationBuilder{}

// nodeAuthorizationServiceTemplate is the template used by the node client
var nodeAuthorizationServiceTemplate = `
Type=oneshot
EnvironmentFile=/etc/environment
{{- if .client_file }}
ExecStartPre=/usr/bin/bash -c 'while [ ! -f {{ .client_cert }} ]; do sleep 5; done; sleep 5'
{{- end }}
ExecStartPre=/usr/bin/mkdir -p /var/lib/kubelet
ExecStart=/usr/bin/docker run --rm --net=host \
	--volume={{ .kube_config_dir }}:/var/lib/kubelet \
	--volume={{ .tls_path }}:/config:ro \
	{{ .image }} \
	client \
	--authorizer={{ .authorizer }} \
	--interval={{ .interval }} \
	--kubeapi-url={{ .kubeapi_url }} \
	--kubeconfig={{ .kube_config }} \
	--node-url={{ .node_url }} \
	--timeout={{ .timeout }} \
	--tls-ca=/config/ca.pem \
	--tls-cert=/config/tls.pem \
	--tls-private-key=/config/tls-key.pem
`

// Build is responsible for handling the node authorization client
func (b *NodeAuthorizationBuilder) Build(c *fi.ModelBuilderContext) error {
	// @check if we are a master and download the certificates for the node-authozier
	if b.UseNodeAuthorizer() && b.IsMaster {
		name := "node-authorizer"
		// creates /src/kubernetes/node-authorizer/{tls,tls-key}.pem
		if err := b.BuildCertificatePairTask(c, name, name, "tls"); err != nil {
			return err
		}
		// creates /src/kubernetes/node-authorizer/ca.pem
		if err := b.BuildCertificateTask(c, fi.CertificateId_CA, filepath.Join(name, "ca.pem")); err != nil {
			return err
		}
	}

	authorizerDir := "node-authorizer"

	// @check if bootstrap tokens are enabled and download client certificates for nodes
	if b.UseBootstrapTokens() && !b.IsMaster {
		if err := b.BuildCertificatePairTask(c, "node-authorizer-client", authorizerDir, "tls"); err != nil {
			return err
		}
		if err := b.BuildCertificateTask(c, fi.CertificateId_CA, authorizerDir+"/ca.pem"); err != nil {
			return err
		}
	}

	glog.V(3).Infof("bootstrap: %t, node authorization: %t, node authorizer: %t", b.UseBootstrapTokens(),
		b.UseNodeAuthorization(), b.UseNodeAuthorizer())

	// @check if the NodeAuhtorizer provision the client service for nodes
	if b.UseNodeAuthorizer() && !b.IsMaster {
		na := b.Cluster.Spec.NodeAuthorization.NodeAuthorizer

		glog.V(3).Infof("node authorization service is enabled, authorizer: %s", na.Authorizer)
		glog.V(3).Infof("node authorization url: %s", na.NodeURL)

		// @step: create the systemd unit to run the node authorization client
		man := &systemd.Manifest{}
		man.Set("Unit", "Description", "Node Authorization Client")
		man.Set("Unit", "Documentation", "https://github.com/kubernetes/kops")
		man.Set("Unit", "After", "docker.service")
		man.Set("Unit", "Before", "kubelet.service")

		tp, err := template.New("service").Parse(nodeAuthorizationServiceTemplate)
		if err != nil {
			return fmt.Errorf("failed to create node authorization client template: %s", err)
		}

		interval := 10 * time.Second
		timeout := 5 * time.Minute

		model := map[string]string{
			"authorizer":      na.Authorizer,
			"client_cert":     filepath.Join(b.PathSrvKubernetes(), authorizerDir, "tls.pem"),
			"image":           na.Image,
			"interval":        interval.String(),
			"kube_config":     b.KubeletBootstrapKubeconfig(),
			"kube_config_dir": path.Dir(b.KubeletBootstrapKubeconfig()),
			"kubeapi_url":     fmt.Sprintf("https://%s", b.Cluster.Spec.MasterInternalName),
			"node_url":        na.NodeURL,
			"timeout":         timeout.String(),
			"tls_path":        filepath.Join(b.PathSrvKubernetes(), authorizerDir),
		}

		content := &bytes.Buffer{}
		if err := tp.ExecuteTemplate(content, "service", model); err != nil {
			return fmt.Errorf("failed to render node authorization client template: %s", err)
		}
		man.SetSection("Service", content.String())

		// @step: add the service task
		c.AddTask(&nodetasks.Service{
			Name: "node-authorizer.service",

			Definition:   s(man.Render()),
			Enabled:      fi.Bool(true),
			ManageState:  fi.Bool(true),
			Running:      fi.Bool(true),
			SmartRestart: fi.Bool(true),
		})
	}

	return nil
}
