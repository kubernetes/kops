package kutil

import (
	"fmt"
	"k8s.io/kops/upup/pkg/fi"
)

type CreateKubecfg struct {
	ContextName  string
	KeyStore     fi.Keystore
	SecretStore  fi.SecretStore
	KubeMasterIP string
}

func (c *CreateKubecfg) WriteKubecfg() error {
	b, err := c.ExtractKubeconfig()
	if err != nil {
		return err
	}

	if err := b.WriteKubecfg(); err != nil {
		return err
	}

	return nil
}

func (c *CreateKubecfg) ExtractKubeconfig() (*KubeconfigBuilder, error) {
	b := NewKubeconfigBuilder()

	b.Context = c.ContextName

	{
		cert, _, err := c.KeyStore.FindKeypair(fi.CertificateId_CA)
		if err != nil {
			return nil, fmt.Errorf("error fetching CA keypair: %v", err)
		}
		if cert != nil {
			b.CACert, err = cert.AsBytes()
			if err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("cannot find CA certificate")
		}
	}

	{
		cert, key, err := c.KeyStore.FindKeypair("kubecfg")
		if err != nil {
			return nil, fmt.Errorf("error fetching kubecfg keypair: %v", err)
		}
		if cert != nil {
			b.ClientCert, err = cert.AsBytes()
			if err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("cannot find kubecfg certificate")
		}
		if key != nil {
			b.ClientKey, err = key.AsBytes()
			if err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("cannot find kubecfg key")
		}
	}

	b.KubeMasterIP = c.KubeMasterIP

	if c.SecretStore != nil {
		secret, err := c.SecretStore.FindSecret("kube")
		if err != nil {
			return nil, err
		}
		if secret != nil {
			b.KubeUser = "admin"
			b.KubePassword = string(secret.Data)
		}
	}

	return b, nil
}
