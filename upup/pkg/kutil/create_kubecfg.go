package kutil

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"

	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/fi"
)

type CreateKubecfg struct {
	ClusterName      string
	KeyStore         fi.CAStore
	SecretStore      fi.SecretStore
	MasterPublicName string

	tmpdir string
}

func (c *CreateKubecfg) WriteKubecfg() error {
	if c.tmpdir == "" {
		tmpdir, err := ioutil.TempDir("", "k8s")
		if err != nil {
			return fmt.Errorf("error creating temporary directory: %v", err)
		}
		c.tmpdir = tmpdir

	}

	b := &KubeconfigBuilder{}
	b.Init()

	b.Context = c.ClusterName

	var err error
	if b.CACert, err = c.copyCertificate(fi.CertificateId_CA); err != nil {
		return err
	}

	if b.KubecfgCert, err = c.copyCertificate("kubecfg"); err != nil {
		return err
	}

	if b.KubecfgKey, err = c.copyPrivateKey("kubecfg"); err != nil {
		return err
	}

	b.KubeMasterIP = c.MasterPublicName

	{
		secret, err := c.SecretStore.FindSecret("kube")
		if err != nil {
			return err
		}
		if secret != nil {
			b.KubeUser = "admin"
			b.KubePassword = string(secret.Data)
		}
	}

	err = b.WriteKubecfg()
	if err != nil {
		return err
	}

	return nil
}

func (c *CreateKubecfg) Close() {
	if c.tmpdir != "" {
		err := os.RemoveAll(c.tmpdir)
		if err != nil {
			glog.Warningf("error deleting tempdir %q: %v", c.tmpdir, err)
		} else {
			c.tmpdir = ""
		}
	}
}

func (c *CreateKubecfg) copyCertificate(id string) (string, error) {
	p := path.Join(c.tmpdir, id+".crt")
	cert, err := c.KeyStore.Cert(id)
	if err != nil {
		return "", fmt.Errorf("error fetching certificate %q: %v", id, err)
	}

	_, err = writeFile(p, cert)
	if err != nil {
		return "", fmt.Errorf("error writing certificate %q: %v", id, err)
	}

	return p, nil
}

func (c *CreateKubecfg) copyPrivateKey(id string) (string, error) {
	p := path.Join(c.tmpdir, id+".key")
	cert, err := c.KeyStore.PrivateKey(id)
	if err != nil {
		return "", fmt.Errorf("error fetching private key %q: %v", id, err)
	}

	_, err = writeFile(p, cert)
	if err != nil {
		return "", fmt.Errorf("error writing private key %q: %v", id, err)
	}

	return p, nil
}

func writeFile(dst string, src io.WriterTo) (int64, error) {
	f, err := os.Create(dst)
	if err != nil {
		return 0, fmt.Errorf("error creating file %q: %v", dst, err)
	}
	defer fi.SafeClose(f)
	return src.WriteTo(f)
}
