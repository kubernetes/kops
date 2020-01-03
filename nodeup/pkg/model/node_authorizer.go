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
	"path"
	"path/filepath"
	"strings"
	"time"

	"k8s.io/kops/pkg/systemd"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"

	"k8s.io/klog"
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

	klog.V(3).Infof("bootstrap: %t, node authorization: %t, node authorizer: %t", b.UseBootstrapTokens(),
		b.UseNodeAuthorization(), b.UseNodeAuthorizer())

	// @check if the NodeAuthorizer provision the client service for nodes
	if b.UseNodeAuthorizer() && !b.IsMaster {
		na := b.Cluster.Spec.NodeAuthorization.NodeAuthorizer

		klog.V(3).Infof("node authorization service is enabled, authorizer: %s", na.Authorizer)
		klog.V(3).Infof("node authorization url: %s", na.NodeURL)

		// @step: create the systemd unit to run the node authorization client
		man := &systemd.Manifest{}
		man.Set("Unit", "Description", "Node Authorization Client")
		man.Set("Unit", "Documentation", "https://github.com/kubernetes/kops")
		man.Set("Unit", "Before", "kubelet.service")
		switch b.Cluster.Spec.ContainerRuntime {
		case "docker":
			man.Set("Unit", "After", "docker.service")
		case "containerd":
			man.Set("Unit", "After", "containerd.service")
		default:
			klog.Warningf("unknown container runtime %q", b.Cluster.Spec.ContainerRuntime)
		}

		clientCert := filepath.Join(b.PathSrvKubernetes(), authorizerDir, "tls.pem")
		man.Set("Service", "Type", "oneshot")
		man.Set("Service", "RemainAfterExit", "yes")
		man.Set("Service", "EnvironmentFile", "/etc/environment")
		man.Set("Service", "ExecStartPre", "/bin/mkdir -p /var/lib/kubelet")
		man.Set("Service", "ExecStartPre", "/usr/bin/docker pull "+na.Image)
		man.Set("Service", "ExecStartPre", "/bin/bash -c 'while [ ! -f "+clientCert+" ]; do sleep 5; done; sleep 5'")

		interval := 10 * time.Second
		if na.Interval != nil {
			interval = na.Interval.Duration
		}
		timeout := 5 * time.Minute
		if na.Timeout != nil {
			timeout = na.Timeout.Duration
		}

		// @node: using a string array just to make it easier to read
		dockerCmd := []string{
			"/usr/bin/docker",
			"run",
			"--rm",
			"--net=host",
			"--volume=" + path.Dir(b.KubeletBootstrapKubeconfig()) + ":/var/lib/kubelet",
			"--volume=" + filepath.Join(b.PathSrvKubernetes(), authorizerDir) + ":/config:ro",
			na.Image,
			"client",
			"--authorizer=" + na.Authorizer,
			"--interval=" + interval.String(),
			"--kubeapi-url=" + fmt.Sprintf("https://%s", b.Cluster.Spec.MasterInternalName),
			"--kubeconfig=" + b.KubeletBootstrapKubeconfig(),
			"--node-url=" + na.NodeURL,
			"--timeout=" + timeout.String(),
			"--tls-client-ca=/config/ca.pem",
			"--tls-cert=/config/tls.pem",
			"--tls-private-key=/config/tls-key.pem",
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
