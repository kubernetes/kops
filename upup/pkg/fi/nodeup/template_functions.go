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

package nodeup

import (
	"encoding/base64"
	"fmt"
	"github.com/golang/glog"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/secrets"
	"k8s.io/kops/util/pkg/vfs"
	"text/template"
)

const TagMaster = "_kubernetes_master"

const DefaultProtokubeImage = "kope/protokube:1.4"

// templateFunctions is a simple helper-class for the functions accessible to templates
type templateFunctions struct {
	nodeupConfig *NodeUpConfig

	// cluster is populated with the current cluster
	cluster *api.Cluster
	// instanceGroup is populated with this node's instance group
	instanceGroup *api.InstanceGroup

	// keyStore is populated with a KeyStore, if KeyStore is set
	keyStore fi.CAStore
	// secretStore is populated with a SecretStore, if SecretStore is set
	secretStore fi.SecretStore

	tags map[string]struct{}

	// kubeletConfig is the kubelet config for the current node
	kubeletConfig *api.KubeletConfigSpec
}

// newTemplateFunctions is the constructor for templateFunctions
func newTemplateFunctions(nodeupConfig *NodeUpConfig, cluster *api.Cluster, instanceGroup *api.InstanceGroup, tags map[string]struct{}) (*templateFunctions, error) {
	t := &templateFunctions{
		nodeupConfig:  nodeupConfig,
		cluster:       cluster,
		instanceGroup: instanceGroup,
		tags:          tags,
	}

	if cluster.Spec.SecretStore != "" {
		glog.Infof("Building SecretStore at %q", cluster.Spec.SecretStore)
		p, err := vfs.Context.BuildVfsPath(cluster.Spec.SecretStore)
		if err != nil {
			return nil, fmt.Errorf("error building secret store path: %v", err)
		}

		t.secretStore = secrets.NewVFSSecretStore(p)
	} else {
		return nil, fmt.Errorf("SecretStore not set")
	}

	if cluster.Spec.KeyStore != "" {
		glog.Infof("Building KeyStore at %q", cluster.Spec.KeyStore)
		p, err := vfs.Context.BuildVfsPath(cluster.Spec.KeyStore)
		if err != nil {
			return nil, fmt.Errorf("error building key store path: %v", err)
		}

		t.keyStore = fi.NewVFSCAStore(p)
	} else {
		return nil, fmt.Errorf("KeyStore not set")
	}

	{
		instanceGroup := t.instanceGroup
		if instanceGroup == nil {
			// Old clusters might not have exported instance groups
			// in that case we build a synthetic instance group with the information that BuildKubeletConfigSpec needs
			// TODO: Remove this once we have a stable release
			glog.Warningf("Building a synthetic instance group")
			instanceGroup = &api.InstanceGroup{}
			instanceGroup.Name = "synthetic"
			if t.IsMaster() {
				instanceGroup.Spec.Role = api.InstanceGroupRoleMaster
			} else {
				instanceGroup.Spec.Role = api.InstanceGroupRoleNode
			}
			t.instanceGroup = instanceGroup
		}
		kubeletConfigSpec, err := api.BuildKubeletConfigSpec(cluster, instanceGroup)
		if err != nil {
			return nil, fmt.Errorf("error building kubelet config: %v", err)
		}
		t.kubeletConfig = kubeletConfigSpec
	}

	return t, nil
}

func (t *templateFunctions) populate(dest template.FuncMap) {

	dest["IsTopologyPublic"] = t.cluster.IsTopologyPublic
	dest["IsTopologyPrivate"] = t.cluster.IsTopologyPrivate

	dest["CACertificatePool"] = t.CACertificatePool
	dest["CACertificate"] = t.CACertificate
	dest["PrivateKey"] = t.PrivateKey
	dest["Certificate"] = t.Certificate
	dest["AllTokens"] = t.AllTokens
	dest["GetToken"] = t.GetToken

	dest["BuildFlags"] = buildFlags
	dest["Base64Encode"] = func(s string) string {
		return base64.StdEncoding.EncodeToString([]byte(s))
	}
	dest["HasTag"] = t.HasTag
	dest["IsMaster"] = t.IsMaster

	// TODO: We may want to move these to a nodeset / masterset specific thing
	dest["KubeDNS"] = func() *api.KubeDNSConfig {
		return t.cluster.Spec.KubeDNS
	}
	dest["KubeScheduler"] = func() *api.KubeSchedulerConfig {
		return t.cluster.Spec.KubeScheduler
	}
	dest["KubeAPIServer"] = func() *api.KubeAPIServerConfig {
		return t.cluster.Spec.KubeAPIServer
	}
	dest["KubeControllerManager"] = func() *api.KubeControllerManagerConfig {
		return t.cluster.Spec.KubeControllerManager
	}
	dest["KubeProxy"] = t.KubeProxyConfig
	dest["KubeletConfig"] = func() *api.KubeletConfigSpec {
		return t.kubeletConfig
	}

	dest["ClusterName"] = func() string {
		return t.cluster.Name
	}

	dest["ProtokubeImage"] = t.ProtokubeImage

	dest["ProtokubeFlags"] = t.ProtokubeFlags
}

// IsMaster returns true if we are tagged as a master
func (t *templateFunctions) IsMaster() bool {
	return t.HasTag(TagMaster)
}

// Tag returns true if we are tagged with the specified tag
func (t *templateFunctions) HasTag(tag string) bool {
	_, found := t.tags[tag]
	return found
}

// CACertificatePool returns the set of valid CA certificates for the cluster
func (t *templateFunctions) CACertificatePool() (*fi.CertificatePool, error) {
	if t.keyStore != nil {
		return t.keyStore.CertificatePool(fi.CertificateId_CA)
	}

	// Fallback to direct properties
	glog.Infof("Falling back to direct configuration for keystore")
	cert, err := t.CACertificate()
	if err != nil {
		return nil, err
	}
	if cert == nil {
		return nil, fmt.Errorf("CA certificate not found (with fallback)")
	}
	pool := &fi.CertificatePool{}
	pool.Primary = cert
	return pool, nil
}

// CACertificate returns the primary CA certificate for the cluster
func (t *templateFunctions) CACertificate() (*fi.Certificate, error) {
	return t.keyStore.Cert(fi.CertificateId_CA)
}

// PrivateKey returns the specified private key
func (t *templateFunctions) PrivateKey(id string) (*fi.PrivateKey, error) {
	return t.keyStore.PrivateKey(id)
}

// Certificate returns the specified private key
func (t *templateFunctions) Certificate(id string) (*fi.Certificate, error) {
	return t.keyStore.Cert(id)
}

// AllTokens returns a map of all tokens
func (t *templateFunctions) AllTokens() (map[string]string, error) {
	tokens := make(map[string]string)
	ids, err := t.secretStore.ListSecrets()
	if err != nil {
		return nil, err
	}
	for _, id := range ids {
		token, err := t.secretStore.FindSecret(id)
		if err != nil {
			return nil, err
		}
		tokens[id] = string(token.Data)
	}
	return tokens, nil
}

// GetToken returns the specified token
func (t *templateFunctions) GetToken(key string) (string, error) {
	token, err := t.secretStore.FindSecret(key)
	if err != nil {
		return "", err
	}
	if token == nil {
		return "", fmt.Errorf("token not found: %q", key)
	}
	return string(token.Data), nil
}

// ProtokubeImage returns the docker image for protokube
func (t *templateFunctions) ProtokubeImage() string {
	image := ""
	if t.nodeupConfig.ProtokubeImage != nil {
		image = t.nodeupConfig.ProtokubeImage.Source
	}
	if image == "" {
		// use current default corresponding to this version of nodeup
		image = DefaultProtokubeImage
	}
	return image
}

// ProtokubeFlags returns the flags object for protokube
func (t *templateFunctions) ProtokubeFlags() *ProtokubeFlags {
	f := &ProtokubeFlags{}

	master := t.IsMaster()

	f.Master = fi.Bool(master)
	if master {
		f.Channels = t.nodeupConfig.Channels
	}

	f.LogLevel = fi.Int(8)
	f.Containerized = fi.Bool(true)
	if t.cluster.Spec.DNSZone != "" {
		f.DNSZoneName = fi.String(t.cluster.Spec.DNSZone)
	}

	return f
}

// KubeProxyConfig builds the KubeProxyConfig configuration object
func (t *templateFunctions) KubeProxyConfig() *api.KubeProxyConfig {
	config := &api.KubeProxyConfig{}
	*config = *t.cluster.Spec.KubeProxy

	// As a special case, if this is the master, we point kube-proxy to the local IP
	// This prevents a circular dependency where kube-proxy can't come up until DNS comes up,
	// which would mean that DNS can't rely on API to come up
	if t.IsMaster() {
		glog.Infof("kube-proxy running on the master; setting API endpoint to localhost")
		config.Master = "http://127.0.0.1:8080"
	}

	return config
}
