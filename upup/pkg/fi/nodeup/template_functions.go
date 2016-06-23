package nodeup

import (
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kube-deploy/upup/pkg/fi"
	"text/template"
	"k8s.io/kube-deploy/upup/pkg/fi/vfs"
)

type templateFunctions struct {
	config *NodeConfig
	// keyStore is populated with a KeyStore, if KeyStore is set
	keyStore fi.CAStore
	// secretStore is populated with a SecretStore, if SecretStore is set
	secretStore fi.SecretStore
}

func buildTemplateFunctions(config *NodeConfig, dest template.FuncMap) error {
	t := &templateFunctions{
		config: config,
	}

	if config.SecretStore != "" {
		glog.Infof("Building SecretStore at %q", config.SecretStore)
		p, err := vfs.Context.BuildVfsPath(config.SecretStore)
		if err != nil {
			return fmt.Errorf("error building secret store path: %v", err)
		}

		secretStore, err := fi.NewVFSSecretStore(p)
		if err != nil {
			return fmt.Errorf("error building secret store: %v", err)
		}

		t.secretStore = secretStore
	}

	if config.KeyStore != "" {
		glog.Infof("Building KeyStore at %q", config.KeyStore)
		p, err := vfs.Context.BuildVfsPath(config.KeyStore)
		if err != nil {
			return fmt.Errorf("error building key store path: %v", err)
		}

		keyStore, err := fi.NewVFSCAStore(p, false)
		if err != nil {
			return fmt.Errorf("error building key store: %v", err)
		}
		t.keyStore = keyStore
	}

	dest["CACertificatePool"] = t.CACertificatePool
	dest["CACertificate"] = t.CACertificate
	dest["PrivateKey"] = t.PrivateKey
	dest["Certificate"] = t.Certificate
	dest["AllTokens"] = t.AllTokens
	dest["GetToken"] = t.GetToken

	return nil
}

// CACertificatePool returns the set of valid CA certificates for the cluster
func (c *templateFunctions) CACertificatePool() (*fi.CertificatePool, error) {
	if c.keyStore != nil {
		return c.keyStore.CertificatePool(fi.CertificateId_CA)
	}

	// Fallback to direct properties
	glog.Infof("Falling back to direct configuration for keystore")
	cert, err := c.CACertificate()
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
func (c *templateFunctions) CACertificate() (*fi.Certificate, error) {
	if c.keyStore != nil {
		return c.keyStore.Cert(fi.CertificateId_CA)
	}

	// Fallback to direct properties
	return c.Certificate(fi.CertificateId_CA)
}

// PrivateKey returns the specified private key
func (c *templateFunctions) PrivateKey(id string) (*fi.PrivateKey, error) {
	if c.keyStore != nil {
		return c.keyStore.PrivateKey(id)
	}

	// Fallback to direct properties
	glog.Infof("Falling back to direct configuration for keystore")
	k := c.config.PrivateKeys[id]
	if k == nil {
		return nil, fmt.Errorf("private key not found: %q (with fallback)", id)
	}
	return k, nil
}

// Certificate returns the specified private key
func (c *templateFunctions) Certificate(id string) (*fi.Certificate, error) {
	if c.keyStore != nil {
		return c.keyStore.Cert(id)
	}

	// Fallback to direct properties
	glog.Infof("Falling back to direct configuration for keystore")
	cert := c.config.Certificates[id]
	if cert == nil {
		return nil, fmt.Errorf("certificate not found: %q (with fallback)", id)
	}
	return cert, nil
}

// AllTokens returns a map of all tokens
func (n *templateFunctions) AllTokens() (map[string]string, error) {
	if n.secretStore != nil {
		tokens := make(map[string]string)
		ids, err := n.secretStore.ListSecrets()
		if err != nil {
			return nil, err
		}
		for _, id := range ids {
			token, err := n.secretStore.FindSecret(id)
			if err != nil {
				return nil, err
			}
			tokens[id] = string(token.Data)
		}
		return tokens, nil
	}

	// Fallback to direct configuration
	glog.Infof("Falling back to direct configuration for secrets")
	return n.config.Tokens, nil
}

// GetToken returns the specified token
func (n *templateFunctions) GetToken(key string) (string, error) {
	if n.secretStore != nil {
		token, err := n.secretStore.FindSecret(key)
		if err != nil {
			return "", err
		}
		if token == nil {
			return "", fmt.Errorf("token not found: %q", key)
		}
		return string(token.Data), nil
	}

	// Fallback to direct configuration
	glog.Infof("Falling back to direct configuration for secrets")
	token := n.config.Tokens[key]
	if token == "" {
		return "", fmt.Errorf("token not found: %q", key)
	}
	return token, nil
}
