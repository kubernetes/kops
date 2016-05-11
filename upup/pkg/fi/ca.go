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
	"math/big"
	"time"
)

type Certificate struct {
	Subject pkix.Name
	IsCA    bool

	Certificate *x509.Certificate
	PublicKey   crypto.PublicKey
}

func (c *Certificate) UnmarshalJSON(b []byte) error {
	s := ""
	if err := json.Unmarshal(b, &s); err == nil {
		d, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			return fmt.Errorf("error decoding certificate base64 data: %q", string(b))
		}
		r, err := LoadPEMCertificate(d)
		if err != nil {
			glog.Infof("Invalid certificate data: %q", string(b))
			return fmt.Errorf("error parsing certificate: %v", err)
		}
		*c = *r
		return nil
	}
	return fmt.Errorf("unknown format for Certificate: %q", string(b))
}

func (c *Certificate) MarshalJSON() ([]byte, error) {
	var data bytes.Buffer
	err := c.WriteCertificate(&data)
	if err != nil {
		return nil, fmt.Errorf("error writing SSL certificate: %v", err)
	}
	return json.Marshal(data.String())
}

type CAStore interface {
	Cert(id string) (*Certificate, error)
	PrivateKey(id string) (*PrivateKey, error)

	FindCert(id string) (*Certificate, error)
	FindPrivateKey(id string) (*PrivateKey, error)

	IssueCert(id string, privateKey *PrivateKey, template *x509.Certificate) (*Certificate, error)
	CreatePrivateKey(id string) (*PrivateKey, error)
}

func (c *Certificate) AsString() (string, error) {
	// Nicer behaviour because this is called from templates
	if c == nil {
		return "", fmt.Errorf("AsString called on nil Certificate")
	}

	var data bytes.Buffer
	err := c.WriteCertificate(&data)
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
	err := WritePrivateKey(c.Key, &data)
	if err != nil {
		return "", fmt.Errorf("error writing SSL private key: %v", err)
	}
	return data.String(), nil
}

func (k *PrivateKey) UnmarshalJSON(b []byte) (err error) {
	s := ""
	if err = json.Unmarshal(b, &s); err == nil {
		d, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			return fmt.Errorf("error decoding private key base64 data: %q", string(b))
		}
		key, err := parsePEMPrivateKey(d)
		if err != nil {
			return fmt.Errorf("error parsing private key: %v", err)
		}
		k.Key = key
		return nil
	}
	return fmt.Errorf("unknown format for private key: %q", string(b))
}

func (k *PrivateKey) MarshalJSON() ([]byte, error) {
	var data bytes.Buffer
	err := WritePrivateKey(k.Key, &data)
	if err != nil {
		return nil, fmt.Errorf("error writing SSL private key: %v", err)
	}
	return json.Marshal(data.String())
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

func (c *Certificate) WriteCertificate(w io.Writer) error {
	return pem.Encode(w, &pem.Block{Type: "CERTIFICATE", Bytes: c.Certificate.Raw})
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

func WritePrivateKey(privateKey crypto.PrivateKey, w io.Writer) error {
	rsaPrivateKey, ok := privateKey.(*rsa.PrivateKey)
	if ok {
		return pem.Encode(w, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(rsaPrivateKey)})
	}

	return fmt.Errorf("unknown private key type: %T", privateKey)
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
