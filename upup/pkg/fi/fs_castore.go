package fi

import (
	"bytes"
	"crypto"
	crypto_rand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"os"
	"path"
)

type FilesystemCAStore struct {
	basedir       string
	caCertificate *Certificate
	caPrivateKey  *PrivateKey
}

var _ CAStore = &FilesystemCAStore{}

func NewFilesystemCAStore(basedir string) (CAStore, error) {
	c := &FilesystemCAStore{
		basedir: basedir,
	}
	err := os.MkdirAll(path.Join(basedir, "private"), 0700)
	if err != nil {
		return nil, fmt.Errorf("error creating directory: %v", err)
	}
	err = os.MkdirAll(path.Join(basedir, "issued"), 0700)
	if err != nil {
		return nil, fmt.Errorf("error creating directory: %v", err)
	}
	caCertificate, err := c.loadCertificate(path.Join(basedir, "ca.crt"))
	if err != nil {
		return nil, err
	}

	if caCertificate != nil {
		privateKeyPath := path.Join(basedir, "private", "ca.key")
		caPrivateKey, err := c.loadPrivateKey(privateKeyPath)
		if err != nil {
			return nil, err
		}
		if caPrivateKey == nil {
			glog.Warningf("CA private key was not found %q", privateKeyPath)
			//return nil, fmt.Errorf("error loading CA private key - key not found")
		}
		c.caCertificate = caCertificate
		c.caPrivateKey = caPrivateKey
	} else {
		err := c.generateCACertificate()
		if err != nil {
			return nil, err
		}
	}
	return c, nil
}

func (c *FilesystemCAStore) generateCACertificate() error {
	subject := &pkix.Name{
		CommonName: "kubernetes",
	}
	template := &x509.Certificate{
		Subject:               *subject,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{},
		BasicConstraintsValid: true,
		IsCA: true,
	}

	caPrivateKey, err := rsa.GenerateKey(crypto_rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("error generating RSA private key: %v", err)
	}

	caCertificate, err := SignNewCertificate(&PrivateKey{Key: caPrivateKey}, template, nil, nil)
	if err != nil {
		return err
	}

	keyPath := path.Join(c.basedir, "private", "ca.key")
	err = c.storePrivateKey(caPrivateKey, keyPath)
	if err != nil {
		return err
	}

	certPath := path.Join(c.basedir, "ca.crt")
	err = c.storeCertificate(caCertificate, certPath)
	if err != nil {
		return err
	}

	// Make double-sure it round-trips
	caCertificate, err = c.loadCertificate(certPath)
	if err != nil {
		return err
	}

	c.caPrivateKey = &PrivateKey{Key: caPrivateKey}
	c.caCertificate = caCertificate
	return nil
}

func (c *FilesystemCAStore) getSubjectKey(subject *pkix.Name) string {
	seq := subject.ToRDNSequence()
	var s bytes.Buffer
	for _, rdnSet := range seq {
		for _, rdn := range rdnSet {
			if s.Len() != 0 {
				s.WriteString(",")
			}
			key := ""
			t := rdn.Type
			if len(t) == 4 && t[0] == 2 && t[1] == 5 && t[2] == 4 {
				switch t[3] {
				case 3:
					key = "cn"
				case 5:
					key = "serial"
				case 6:
					key = "c"
				case 7:
					key = "l"
				case 10:
					key = "o"
				case 11:
					key = "ou"
				}
			}
			if key == "" {
				key = t.String()
			}
			s.WriteString(fmt.Sprintf("%v=%v", key, rdn.Value))
		}
	}
	return s.String()
}

func (c *FilesystemCAStore) buildCertificatePath(id string) string {
	return path.Join(c.basedir, "issued", id+".crt")
}

func (c *FilesystemCAStore) buildPrivateKeyPath(id string) string {
	return path.Join(c.basedir, "private", id+".key")
}

func (c *FilesystemCAStore) loadCertificate(p string) (*Certificate, error) {
	data, err := ioutil.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
	}
	cert, err := LoadPEMCertificate(data)
	if err != nil {
		return nil, err
	}
	if cert == nil {
		return nil, nil
	}
	return cert, nil
}

func (c *FilesystemCAStore) Cert(id string) (*Certificate, error) {
	cert, err := c.FindCert(id)
	if err == nil && cert == nil {
		return nil, fmt.Errorf("cannot find cert %q", id)
	}
	return cert, err

}

func (c *FilesystemCAStore) FindCert(id string) (*Certificate, error) {
	var cert *Certificate
	if id == "ca" {
		cert = c.caCertificate
	} else {
		var err error
		p := c.buildCertificatePath(id)
		cert, err = c.loadCertificate(p)
		if err != nil {
			return nil, err
		}
	}
	return cert, nil
}

func (c *FilesystemCAStore) IssueCert(id string, privateKey *PrivateKey, template *x509.Certificate) (*Certificate, error) {
	p := c.buildCertificatePath(id)

	if c.caPrivateKey == nil {
		return nil, fmt.Errorf("ca.key was not found; cannot issue certificates")
	}
	cert, err := SignNewCertificate(privateKey, template, c.caCertificate.Certificate, c.caPrivateKey)
	if err != nil {
		return nil, err
	}

	err = c.storeCertificate(cert, p)
	if err != nil {
		return nil, err
	}

	// Make double-sure it round-trips
	return c.loadCertificate(p)
}

func (c *FilesystemCAStore) loadPrivateKey(p string) (*PrivateKey, error) {
	data, err := ioutil.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
	}
	k, err := parsePEMPrivateKey(data)
	if err != nil {
		return nil, fmt.Errorf("error parsing private key from %q: %v", p, err)
	}
	if k == nil {
		return nil, nil
	}
	return &PrivateKey{Key: k}, nil
}

func (c *FilesystemCAStore) FindPrivateKey(id string) (*PrivateKey, error) {
	var key *PrivateKey
	if id == "ca" {
		key = c.caPrivateKey
	} else {
		var err error
		p := c.buildPrivateKeyPath(id)
		key, err = c.loadPrivateKey(p)
		if err != nil {
			return nil, err
		}
	}
	return key, nil
}

func (c *FilesystemCAStore) PrivateKey(id string) (*PrivateKey, error) {
	key, err := c.FindPrivateKey(id)
	if err == nil && key == nil {
		return nil, fmt.Errorf("cannot find SSL key %q", id)
	}
	return key, err

}

func (c *FilesystemCAStore) CreatePrivateKey(id string) (*PrivateKey, error) {
	p := c.buildPrivateKeyPath(id)

	privateKey, err := rsa.GenerateKey(crypto_rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("error generating RSA private key: %v", err)
	}

	err = c.storePrivateKey(privateKey, p)
	if err != nil {
		return nil, err
	}

	return &PrivateKey{Key: privateKey}, nil
}

func (c *FilesystemCAStore) storePrivateKey(privateKey crypto.PrivateKey, p string) error {
	var data bytes.Buffer
	err := WritePrivateKey(privateKey, &data)
	if err != nil {
		return err
	}

	return c.writeFile(data.Bytes(), p)
}

func (c *FilesystemCAStore) storeCertificate(cert *Certificate, p string) error {
	var data bytes.Buffer
	err := cert.WriteCertificate(&data)
	if err != nil {
		return err
	}

	return c.writeFile(data.Bytes(), p)
}

func (c *FilesystemCAStore) writeFile(data []byte, p string) error {
	// TODO: concurrency?
	err := ioutil.WriteFile(p, data, 0600)
	if err != nil {
		// TODO: Delete file on disk?  Write a temp file and move it atomically?
		return fmt.Errorf("error writing certificate/key data to path %q: %v", p, err)
	}
	return nil
}
