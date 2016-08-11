package fi

import (
	"bytes"
	"crypto"
	crypto_rand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/golang/glog"
	"io"
	"k8s.io/kops/upup/pkg/fi/vfs"
	"math/big"
	"time"
)

const CertificateId_CA = "ca"

type Certificate struct {
	Subject pkix.Name
	IsCA    bool

	Certificate *x509.Certificate
	PublicKey   crypto.PublicKey
}

const (
	SecretTypeSSHPublicKey = "SSHPublicKey"
	SecretTypeKeypair      = "Keypair"
	SecretTypeSecret       = "Secret"

	// Name for the primary SSH key
	SecretNameSSHPrimary = "admin"
)

type KeystoreItem struct {
	Type string
	Name string
	Id   string
	Data []byte
}

func (c *Certificate) UnmarshalJSON(b []byte) error {
	s := ""
	if err := json.Unmarshal(b, &s); err == nil {
		r, err := LoadPEMCertificate([]byte(s))
		if err != nil {
			// Alternative form: Check if base64 encoded
			// TODO: Do we need this?  I think we need this only on nodeup, but maybe we could just not base64-it?
			d, err2 := base64.StdEncoding.DecodeString(s)
			if err2 == nil {
				r2, err2 := LoadPEMCertificate(d)
				if err2 == nil {
					glog.Warningf("used base64 decode of certificate")
					r = r2
					err = nil
				}
			}

			if err != nil {
				glog.Infof("Invalid certificate data: %q", string(b))
				return fmt.Errorf("error parsing certificate: %v", err)
			}
		}
		*c = *r
		return nil
	}
	return fmt.Errorf("unknown format for Certificate: %q", string(b))
}

func (c *Certificate) MarshalJSON() ([]byte, error) {
	var data bytes.Buffer
	_, err := c.WriteTo(&data)
	if err != nil {
		return nil, fmt.Errorf("error writing SSL certificate: %v", err)
	}
	return json.Marshal(data.String())
}

type CAStore interface {
	// Cert returns the primary specified certificate
	Cert(name string) (*Certificate, error)
	// CertificatePool returns all active certificates with the specified id
	CertificatePool(name string) (*CertificatePool, error)
	PrivateKey(name string) (*PrivateKey, error)

	FindCert(name string) (*Certificate, error)
	FindPrivateKey(name string) (*PrivateKey, error)

	//IssueCert(id string, privateKey *PrivateKey, template *x509.Certificate) (*Certificate, error)
	//CreatePrivateKey(id string) (*PrivateKey, error)

	CreateKeypair(name string, template *x509.Certificate) (*Certificate, *PrivateKey, error)

	// List will list all the items, but will not fetch the data
	List() ([]*KeystoreItem, error)

	// VFSPath returns the path where the CAStore is stored
	VFSPath() vfs.Path

	// AddCert adds an alternative certificate to the pool (primarily useful for CAs)
	AddCert(name string, cert *Certificate) error

	// AddSSHPublicKey adds an SSH public key
	AddSSHPublicKey(name string, data []byte) error

	// FindSSHPublicKeys retrieves the SSH public keys with the specific name
	FindSSHPublicKeys(name string) ([]*KeystoreItem, error)

	// DeleteSecret will delete the specified item
	DeleteSecret(item *KeystoreItem) error
}

func (c *Certificate) AsString() (string, error) {
	// Nicer behaviour because this is called from templates
	if c == nil {
		return "", fmt.Errorf("AsString called on nil Certificate")
	}

	var data bytes.Buffer
	_, err := c.WriteTo(&data)
	if err != nil {
		return "", fmt.Errorf("error writing SSL certificate: %v", err)
	}
	return data.String(), nil
}

type PrivateKey struct {
	Key crypto.PrivateKey
}

func (c *PrivateKey) AsString() (string, error) {
	// Nicer behaviour because this is called from templates
	if c == nil {
		return "", fmt.Errorf("AsString called on nil Certificate")
	}

	var data bytes.Buffer
	_, err := c.WriteTo(&data)
	if err != nil {
		return "", fmt.Errorf("error writing SSL private key: %v", err)
	}
	return data.String(), nil
}

func (k *PrivateKey) UnmarshalJSON(b []byte) (err error) {
	s := ""
	if err := json.Unmarshal(b, &s); err == nil {
		r, err := parsePEMPrivateKey([]byte(s))
		if err != nil {
			// Alternative form: Check if base64 encoded
			// TODO: Do we need this?  I think we need this only on nodeup, but maybe we could just not base64-it?
			d, err2 := base64.StdEncoding.DecodeString(s)
			if err2 == nil {
				r2, err2 := parsePEMPrivateKey(d)
				if err2 == nil {
					glog.Warningf("used base64 decode of PrivateKey")
					r = r2
					err = nil
				}
			}

			if err != nil {
				return fmt.Errorf("error parsing private key: %v", err)
			}
		}
		k.Key = r
		return nil
	}

	return fmt.Errorf("unknown format for private key: %q", string(b))
}

func (k *PrivateKey) MarshalJSON() ([]byte, error) {
	var data bytes.Buffer
	_, err := k.WriteTo(&data)
	if err != nil {
		return nil, fmt.Errorf("error writing SSL private key: %v", err)
	}
	return json.Marshal(data.String())
}

var _ io.WriterTo = &PrivateKey{}

func (k *PrivateKey) WriteTo(w io.Writer) (int64, error) {
	if k.Key == nil {
		// For the dry-run case
		return 0, nil
	}

	var data bytes.Buffer
	var err error

	switch pk := k.Key.(type) {
	case *rsa.PrivateKey:
		err = pem.Encode(w, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(pk)})
	default:
		return 0, fmt.Errorf("unknown private key type: %T", k.Key)
	}

	if err != nil {
		return 0, fmt.Errorf("error writing SSL private key: %v", err)
	}

	return data.WriteTo(w)
}

func LoadPEMCertificate(pemData []byte) (*Certificate, error) {
	cert, err := parsePEMCertificate(pemData)
	if err != nil {
		return nil, err
	}

	c := &Certificate{
		Subject:     cert.Subject,
		Certificate: cert,
		PublicKey:   cert.PublicKey,
		IsCA:        cert.IsCA,
	}
	return c, nil
}

func SignNewCertificate(privateKey *PrivateKey, template *x509.Certificate, signer *x509.Certificate, signerPrivateKey *PrivateKey) (*Certificate, error) {
	if template.PublicKey == nil {
		rsaPrivateKey, ok := privateKey.Key.(*rsa.PrivateKey)
		if ok {
			template.PublicKey = rsaPrivateKey.Public()
		}
	}

	if template.PublicKey == nil {
		return nil, fmt.Errorf("PublicKey not set, and cannot be determined from %T", privateKey)
	}

	now := time.Now()
	if template.NotBefore.IsZero() {
		template.NotBefore = now.Add(time.Hour * -48)
	}

	if template.NotAfter.IsZero() {
		template.NotAfter = now.Add(time.Hour * 10 * 365 * 24)
	}

	if template.SerialNumber == nil {
		serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
		serialNumber, err := crypto_rand.Int(crypto_rand.Reader, serialNumberLimit)
		if err != nil {
			return nil, fmt.Errorf("error generating certificate serial number: %s", err)
		}
		template.SerialNumber = serialNumber
	}
	var parent *x509.Certificate
	if signer != nil {
		parent = signer
	} else {
		parent = template
		signerPrivateKey = privateKey
	}

	if template.KeyUsage == 0 {
		template.KeyUsage = x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment
	}

	if template.ExtKeyUsage == nil {
		template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
	}
	//c.SignatureAlgorithm  = do we want to overrride?

	certificateData, err := x509.CreateCertificate(crypto_rand.Reader, template, parent, template.PublicKey, signerPrivateKey.Key)
	if err != nil {
		return nil, fmt.Errorf("error creating certificate: %v", err)
	}

	c := &Certificate{}
	c.PublicKey = template.PublicKey

	cert, err := x509.ParseCertificate(certificateData)
	if err != nil {
		return nil, fmt.Errorf("error parsing certificate: %v", err)
	}
	c.Certificate = cert

	return c, nil
}

var _ io.WriterTo = &Certificate{}

func (c *Certificate) WriteTo(w io.Writer) (int64, error) {
	// For the dry-run case
	if c.Certificate == nil {
		return 0, nil
	}

	var b bytes.Buffer
	err := pem.Encode(&b, &pem.Block{Type: "CERTIFICATE", Bytes: c.Certificate.Raw})
	if err != nil {
		return 0, err
	}
	return b.WriteTo(w)
}

func parsePEMCertificate(pemData []byte) (*x509.Certificate, error) {
	for {
		block, rest := pem.Decode(pemData)
		if block == nil {
			return nil, fmt.Errorf("could not parse certificate")
		}

		if block.Type == "CERTIFICATE" {
			glog.V(8).Infof("Parsing pem block: %q", block.Type)
			return x509.ParseCertificate(block.Bytes)
		} else {
			glog.Infof("Ignoring unexpected PEM block: %q", block.Type)
		}

		pemData = rest
	}
}

func parsePEMPrivateKey(pemData []byte) (crypto.PrivateKey, error) {
	for {
		block, rest := pem.Decode(pemData)
		if block == nil {
			return nil, fmt.Errorf("could not parse private key")
		}

		if block.Type == "RSA PRIVATE KEY" {
			glog.V(8).Infof("Parsing pem block: %q", block.Type)
			return x509.ParsePKCS1PrivateKey(block.Bytes)
		} else if block.Type == "PRIVATE KEY" {
			glog.V(8).Infof("Parsing pem block: %q", block.Type)
			k, err := x509.ParsePKCS8PrivateKey(block.Bytes)
			if err != nil {
				return nil, err
			}
			return k.(crypto.PrivateKey), nil
		} else {
			glog.Infof("Ignoring unexpected PEM block: %q", block.Type)
		}

		pemData = rest
	}
}

type CertificatePool struct {
	Secondary []*Certificate
	Primary   *Certificate
}

func (c *CertificatePool) AsString() (string, error) {
	// Nicer behaviour because this is called from templates
	if c == nil {
		return "", fmt.Errorf("AsString called on nil CertificatePool")
	}

	var data bytes.Buffer
	if c.Primary != nil {
		_, err := c.Primary.WriteTo(&data)
		if err != nil {
			return "", fmt.Errorf("error writing SSL certificate: %v", err)
		}
	}
	for _, cert := range c.Secondary {
		_, err := cert.WriteTo(&data)
		if err != nil {
			return "", fmt.Errorf("error writing SSL certificate: %v", err)
		}
	}
	return data.String(), nil
}
