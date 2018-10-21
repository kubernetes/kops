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
	"fmt"
	"path"
	"path/filepath"
	"strings"
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

// Build is responsible for handling the node authorization client
func (b *NodeAuthorizationBuilder) Build(c *fi.ModelBuilderContext) error {
	// @check if we are a master and download the certificates for the node-authozier
	if b.UseBootstrapTokens() && b.IsMaster {
		d := b.ConfigDirForComponent("node-authorizer")

		// creates /etc/srv/kubernetes/node-authorizer/tls{.key,.crt}
		if err := b.BuildCertificatePairTask(c, "node-authorizer", d, "tls"); err != nil {
			return err
		}
		// creates /etc/srv/kubernetes/node-authorizer/ca.crt
		if err := b.BuildCertificateTask(c, fi.CertificateId_CA, filepath.Join(d, "ca.crt")); err != nil {
			return err
		}
	}

	// @check if bootstrap tokens are enabled and download client certificates for nodes
	if b.UseBootstrapTokens() && !b.IsMaster {
		d := b.ConfigDirForComponent("node-authorizer-client")

		if err := b.BuildCertificatePairTask(c, "node-authorizer-client", d, "tls"); err != nil {
			return err
		}
		if err := b.BuildCertificateTask(c, fi.CertificateId_CA, filepath.Join(d, "ca.crt")); err != nil {
			return err
		}
	}

	glog.V(3).Infof("bootstrap: %t, node authorization: %t, node authorizer: %t", b.UseBootstrapTokens(),
		b.UseNodeAuthorization(), b.UseNodeAuthorizer())

	// @check if the NodeAuthorizer provision the client service for nodes
	if b.UseNodeAuthorizer() && !b.IsMaster {
		d := b.ConfigDirForComponent("node-authorizer-client")

		na := b.Cluster.Spec.NodeAuthorization.NodeAuthorizer

		glog.V(3).Infof("node authorization service is enabled, authorizer: %s", na.Authorizer)
		glog.V(3).Infof("node authorization url: %s", na.NodeURL)

		// @step: create the systemd unit to run the node authorization client
		man := &systemd.Manifest{}
		man.Set("Unit", "Description", "Node Authorization Client")
		man.Set("Unit", "Documentation", "https://github.com/kubernetes/kops")
		man.Set("Unit", "After", "docker.service")
		man.Set("Unit", "Before", "kubelet.service")

		clientCert := filepath.Join(d, "tls.crt")
		man.Set("Service", "Type", "oneshot")
		man.Set("Service", "RemainAfterExit", "yes")
		man.Set("Service", "EnvironmentFile", "/etc/environment")
		man.Set("Service", "ExecStartPre", "/bin/mkdir -p /var/lib/kubelet")
		man.Set("Service", "ExecStartPre", "/usr/bin/docker pull "+na.Image)
		man.Set("Service", "ExecStartPre", "/bin/bash -c 'while [ ! -f "+clientCert+" ]; do sleep 5; done; sleep 5'")

		interval := 10 * time.Second
		timeout := 5 * time.Minute

		// @node: using a string array just to make it easier to read
		dockerCmd := []string{
			"/usr/bin/docker",
			"run",
			"--rm",
			"--net=host",
			"--volume=" + path.Dir(b.KubeletBootstrapKubeconfig()) + ":/var/lib/kubelet",
			"--volume=" + d + ":/config:ro",
			na.Image,
			"client",
			"--authorizer=" + na.Authorizer,
			"--interval=" + interval.String(),
			"--kubeapi-url=" + fmt.Sprintf("https://%s", b.Cluster.Spec.MasterInternalName),
			"--kubeconfig=" + b.KubeletBootstrapKubeconfig(),
			"--node-url=" + na.NodeURL,
			"--timeout=" + timeout.String(),
			"--tls-client-ca=/config/ca.crt",
			"--tls-cert=/config/tls.crt",
			"--tls-private-key=/config/tls.key",
		}
		man.Set("Service", "ExecStart", strings.Join(dockerCmd, " "))

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
