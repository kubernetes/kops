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
	"runtime"
	"text/template"

	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/util/sets"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/pkg/flagbuilder"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/secrets"
	"k8s.io/kops/util/pkg/vfs"
)

const TagMaster = "_kubernetes_master"

// templateFunctions is a simple helper-class for the functions accessible to templates
type templateFunctions struct {
	nodeupConfig *nodeup.Config

	// cluster is populated with the current cluster
	cluster *api.Cluster
	// instanceGroup is populated with this node's instance group
	instanceGroup *api.InstanceGroup

	// keyStore is populated with a KeyStore, if KeyStore is set
	keyStore fi.CAStore
	// secretStore is populated with a SecretStore, if SecretStore is set
	secretStore fi.SecretStore

	tags sets.String
}

// newTemplateFunctions is the constructor for templateFunctions
func newTemplateFunctions(nodeupConfig *nodeup.Config, cluster *api.Cluster, instanceGroup *api.InstanceGroup, tags sets.String) (*templateFunctions, error) {
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

	return t, nil
}

func (t *templateFunctions) populate(dest template.FuncMap) {
	dest["Arch"] = func() string {
		return runtime.GOARCH
	}

	dest["CACertificate"] = t.CACertificate
	dest["PrivateKey"] = t.PrivateKey
	dest["Certificate"] = t.Certificate
	dest["GetToken"] = t.GetToken

	dest["BuildFlags"] = flagbuilder.BuildFlags
	dest["Base64Encode"] = func(s string) string {
		return base64.StdEncoding.EncodeToString([]byte(s))
	}

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

	dest["ClusterName"] = func() string {
		return t.cluster.ObjectMeta.Name
	}
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

// IsMaster returns true if we are tagged as a master
func (t *templateFunctions) isMaster() bool {
	return t.hasTag(TagMaster)
}

// Tag returns true if we are tagged with the specified tag
func (t *templateFunctions) hasTag(tag string) bool {
	_, found := t.tags[tag]
	return found
}
