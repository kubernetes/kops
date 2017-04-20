package proxy

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/docker/docker/pkg/homedir"
)

type TLSConfig struct {
	Enabled, Verify   bool
	Cert, Key, CACert string
	*tls.Config
}

// IsEnabled returns true if TLS is enable, according to the config.
func (c *TLSConfig) IsEnabled() bool {
	if c == nil {
		return false
	}
	return c.Enabled || c.Verify
}

// LoadCerts loads the certificates into c.Config, if TLS is enabled.
func (c *TLSConfig) LoadCerts() error {
	if !c.IsEnabled() {
		return nil
	}

	dockerCertPath := os.Getenv("DOCKER_CERT_PATH")
	if dockerCertPath == "" {
		dockerCertPath = filepath.Join(homedir.Get(), ".docker")
	}

	if c.CACert == "" {
		c.CACert = filepath.Join(dockerCertPath, defaultCaFile)
	}
	if c.Cert == "" {
		c.Cert = filepath.Join(dockerCertPath, defaultCertFile)
	}
	if c.Key == "" {
		c.Key = filepath.Join(dockerCertPath, defaultKeyFile)
	}

	tlsConfig := &tls.Config{
		NextProtos: []string{"http/1.1"},
		// Avoid fallback on insecure SSL protocols
		MinVersion: tls.VersionTLS10,
	}

	if c.Verify {
		certPool := x509.NewCertPool()
		file, err := ioutil.ReadFile(c.CACert)
		if err != nil {
			return fmt.Errorf("Couldn't read CA certificate: %v", err)
		}
		certPool.AppendCertsFromPEM(file)
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		tlsConfig.ClientCAs = certPool
	}

	_, errCert := os.Stat(c.Cert)
	_, errKey := os.Stat(c.Key)
	if errCert == nil && errKey == nil {
		cert, err := tls.LoadX509KeyPair(c.Cert, c.Key)
		if err != nil {
			return fmt.Errorf("Couldn't load X509 key pair: %q. Make sure the key is encrypted", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	c.Config = tlsConfig
	return nil
}
