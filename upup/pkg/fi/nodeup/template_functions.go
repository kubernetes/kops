package nodeup

import (
	"encoding/base64"
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/api"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/vfs"
	"text/template"
)

const TagMaster = "_kubernetes_master"

// templateFunctions is a simple helper-class for the functions accessible to templates
type templateFunctions struct {
	nodeupConfig *NodeUpConfig
	cluster      *api.Cluster
	// keyStore is populated with a KeyStore, if KeyStore is set
	keyStore fi.CAStore
	// secretStore is populated with a SecretStore, if SecretStore is set
	secretStore fi.SecretStore

	tags map[string]struct{}
}

// newTemplateFunctions is the constructor for templateFunctions
func newTemplateFunctions(nodeupConfig *NodeUpConfig, cluster *api.Cluster, tags map[string]struct{}) (*templateFunctions, error) {
	t := &templateFunctions{
		nodeupConfig: nodeupConfig,
		cluster:      cluster,
		tags:         tags,
	}

	if cluster.Spec.SecretStore != "" {
		glog.Infof("Building SecretStore at %q", cluster.Spec.SecretStore)
		p, err := vfs.Context.BuildVfsPath(cluster.Spec.SecretStore)
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

	if cluster.Spec.KeyStore != "" {
		glog.Infof("Building KeyStore at %q", cluster.Spec.KeyStore)
		p, err := vfs.Context.BuildVfsPath(cluster.Spec.KeyStore)
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
	dest["KubeProxy"] = func() *api.KubeProxyConfig {
		return t.cluster.Spec.KubeProxy
	}
	dest["Kubelet"] = func() *api.KubeletConfig {
		if t.IsMaster() {
			return t.cluster.Spec.MasterKubelet
		} else {
			return t.cluster.Spec.Kubelet
		}
	}
	dest["ClusterName"] = func() string { return t.cluster.Name }
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
