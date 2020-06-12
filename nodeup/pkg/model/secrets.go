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

	"k8s.io/kops/util/pkg/vfs"

	"k8s.io/kops/pkg/apis/kops"

	"k8s.io/kops/pkg/model/components"
	"k8s.io/kops/pkg/tokens"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

// SecretBuilder writes secrets
type SecretBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &SecretBuilder{}

const (
	adminUser  = "admin"
	adminGroup = "system:masters"
)

// Build is responsible for pulling down the secrets
func (b *SecretBuilder) Build(c *fi.ModelBuilderContext) error {
	if b.KeyStore == nil {
		return fmt.Errorf("KeyStore not set")
	}

	// @step: retrieve the platform ca
	if err := b.BuildCertificateTask(c, fi.CertificateIDCA, "ca.crt", nil); err != nil {
		return err
	}

	// Write out docker auth secret, if exists
	if b.SecretStore != nil {
		key := "dockerconfig"
		dockercfg, _ := b.SecretStore.Secret(key)
		if dockercfg != nil {
			contents := string(dockercfg.Data)
			c.AddTask(&nodetasks.File{
				Path:     filepath.Join("root", ".docker", "config.json"),
				Contents: fi.NewStringResource(contents),
				Type:     nodetasks.FileType_File,
				Mode:     s("0600"),
			})
		}
	}

	// if we are not a master we can stop here
	if !b.IsMaster {
		return nil
	}

	{
		// A few names used from inside the cluster, which all resolve the same based on our default suffixes
		alternateNames := []string{
			"kubernetes",
			"kubernetes.default",
			"kubernetes.default.svc",
			"kubernetes.default.svc." + b.Cluster.Spec.ClusterDNSDomain,
		}

		// Names specified in the cluster spec
		alternateNames = append(alternateNames, b.Cluster.Spec.MasterPublicName)
		alternateNames = append(alternateNames, b.Cluster.Spec.MasterInternalName)
		alternateNames = append(alternateNames, b.Cluster.Spec.AdditionalSANs...)

		// Load balancer IPs passed in through NodeupConfig
		alternateNames = append(alternateNames, b.NodeupConfig.ApiserverAdditionalIPs...)

		// Referencing it by internal IP should work also
		{
			ip, err := components.WellKnownServiceIP(&b.Cluster.Spec, 1)
			if err != nil {
				return err
			}
			alternateNames = append(alternateNames, ip.String())
		}

		// We also want to be able to reference it locally via https://127.0.0.1
		alternateNames = append(alternateNames, "127.0.0.1")

		if b.Cluster.Spec.CloudProvider == "openstack" {
			if b.Cluster.Spec.Topology != nil && b.Cluster.Spec.Topology.Masters == kops.TopologyPrivate {
				instanceAddress, err := getInstanceAddress()
				if err != nil {
					return err
				}
				alternateNames = append(alternateNames, instanceAddress)
			}
		}

		issueCert := &nodetasks.IssueCert{
			Name:           "master",
			Signer:         fi.CertificateIDCA,
			Type:           "server",
			Subject:        nodetasks.PKIXName{CommonName: "kubernetes-master"},
			AlternateNames: alternateNames,
		}

		// Including the CA certificate is more correct, and is needed for e.g. AWS WebIdentity federation
		issueCert.IncludeRootCertificate = true

		c.AddTask(issueCert)
		err := issueCert.AddFileTasks(c, b.PathSrvKubernetes(), "server", "", nil)
		if err != nil {
			return err
		}
	}

	{
		issueCert := &nodetasks.IssueCert{
			Name:   "apiserver-aggregator",
			Signer: "apiserver-aggregator-ca",
			Type:   "client",
			// Must match RequestheaderAllowedNames
			Subject: nodetasks.PKIXName{CommonName: "aggregator"},
		}
		c.AddTask(issueCert)
		err := issueCert.AddFileTasks(c, b.PathSrvKubernetes(), "apiserver-aggregator", "apiserver-aggregator-ca", nil)
		if err != nil {
			return err
		}
	}

	if err := b.BuildPrivateKeyTask(c, "master", "service-account.key", nil); err != nil {
		return err
	}

	// Support for basic auth was deprecated 1.16 and removed in 1.19
	// https://github.com/kubernetes/kubernetes/pull/89069
	if b.IsKubernetesLT("1.19") && b.SecretStore != nil {
		key := "kube"
		token, err := b.SecretStore.FindSecret(key)
		if err != nil {
			return err
		}
		if token == nil {
			return fmt.Errorf("token not found: %q", key)
		}
		csv := string(token.Data) + "," + adminUser + "," + adminUser + "," + adminGroup

		t := &nodetasks.File{
			Path:     filepath.Join(b.PathSrvKubernetes(), "basic_auth.csv"),
			Contents: fi.NewStringResource(csv),
			Type:     nodetasks.FileType_File,
			Mode:     s("0600"),
		}
		c.AddTask(t)
	}

	if b.SecretStore != nil {
		allTokens, err := b.allAuthTokens()
		if err != nil {
			return err
		}

		var lines []string
		for id, token := range allTokens {
			if id == adminUser {
				lines = append(lines, token+","+id+","+id+","+adminGroup)
			} else {
				lines = append(lines, token+","+id+","+id)
			}
		}
		csv := strings.Join(lines, "\n")

		c.AddTask(&nodetasks.File{
			Path:     filepath.Join(b.PathSrvKubernetes(), "known_tokens.csv"),
			Contents: fi.NewStringResource(csv),
			Type:     nodetasks.FileType_File,
			Mode:     s("0600"),
		})
	}

	return nil
}

// allTokens returns a map of all auth tokens that are present
func (b *SecretBuilder) allAuthTokens() (map[string]string, error) {
	possibleTokens := tokens.GetKubernetesAuthTokens_Deprecated()

	tokens := make(map[string]string)
	for _, id := range possibleTokens {
		token, err := b.SecretStore.FindSecret(id)
		if err != nil {
			return nil, err
		}
		if token != nil {
			tokens[id] = string(token.Data)
		}
	}
	return tokens, nil
}

func getInstanceAddress() (string, error) {

	addrBytes, err := vfs.Context.ReadFile("metadata://openstack/local-ipv4")
	if err != nil {
		return "", nil
	}
	return string(addrBytes), nil

}
