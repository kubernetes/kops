package nodeup

import (
	"encoding/base64"
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kube-deploy/upup/pkg/fi"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup"
	"k8s.io/kube-deploy/upup/pkg/fi/vfs"
	"text/template"
)

const TagMaster = "_kubernetes_master"

// templateFunctions is a simple helper-class for the functions accessible to templates
type templateFunctions struct {
	nodeupConfig *NodeUpConfig
	cluster      *cloudup.ClusterConfig
	// keyStore is populated with a KeyStore, if KeyStore is set
	keyStore fi.CAStore
	// secretStore is populated with a SecretStore, if SecretStore is set
	secretStore fi.SecretStore

	tags map[string]struct{}
}

// newTemplateFunctions is the constructor for templateFunctions
func newTemplateFunctions(nodeupConfig *NodeUpConfig, cluster *cloudup.ClusterConfig, tags map[string]struct{}) (*templateFunctions, error) {
	t := &templateFunctions{
		nodeupConfig: nodeupConfig,
		cluster:      cluster,
		tags:         tags,
	}

	if cluster.SecretStore != "" {
		glog.Infof("Building SecretStore at %q", cluster.SecretStore)
		p, err := vfs.Context.BuildVfsPath(cluster.SecretStore)
		if err != nil {
			return nil, fmt.Errorf("error building secret store path: %v", err)
		}

		secretStore, err := fi.NewVFSSecretStore(p)
		if err != nil {
			return nil, fmt.Errorf("error building secret store: %v", err)
		}

		t.secretStore = secretStore
	} else {
		return nil, fmt.Errorf("SecretStore not set")
	}

	if cluster.KeyStore != "" {
		glog.Infof("Building KeyStore at %q", cluster.KeyStore)
		p, err := vfs.Context.BuildVfsPath(cluster.KeyStore)
		if err != nil {
			return nil, fmt.Errorf("error building key store path: %v", err)
		}

		keyStore, err := fi.NewVFSCAStore(p, false)
		if err != nil {
			return nil, fmt.Errorf("error building key store: %v", err)
		}
		t.keyStore = keyStore
	} else {
		return nil, fmt.Errorf("KeyStore not set")
	}

	return t, nil
}

func (t *templateFunctions) populate(dest template.FuncMap) {
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
	dest["KubeDNS"] = func() *cloudup.KubeDNSConfig {
		return t.cluster.KubeDNS
	}
	dest["KubeScheduler"] = func() *cloudup.KubeSchedulerConfig {
		return t.cluster.KubeScheduler
	}
	dest["APIServer"] = func() *cloudup.APIServerConfig {
		return t.cluster.APIServer
	}
	dest["KubeControllerManager"] = func() *cloudup.KubeControllerManagerConfig {
		return t.cluster.KubeControllerManager
	}
	dest["KubeProxy"] = func() *cloudup.KubeProxyConfig {
		return t.cluster.KubeProxy
	}
	dest["Kubelet"] = func() *cloudup.KubeletConfig {
		if t.IsMaster() {
			return t.cluster.MasterKubelet
		} else {
			return t.cluster.Kubelet
		}
	}
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
