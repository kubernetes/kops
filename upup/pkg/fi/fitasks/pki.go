package fitasks

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kube-deploy/upup/pkg/fi"
	"k8s.io/kube-deploy/upup/pkg/fi/utils"
	"net"
	"strings"
)

const (
	CertificateType_Client string = "client"
	CertificateType_Server string = "server"
)

type PKIKeyPairTask struct {
	Name           string
	Subject        *pkix.Name `json:"subject"`
	Type           string     `json:"type"`
	AlternateNames []string   `json:"alternateNames"`
}

func (t *PKIKeyPairTask) String() string {
	return fmt.Sprintf("PKI: %s", t.Name)
}

func NewPKIKeyPairTask(name string, contents string, meta string) (fi.Task, error) {
	t := &PKIKeyPairTask{Name: name}

	if contents != "" {
		err := utils.YamlUnmarshal([]byte(contents), t)
		if err != nil {
			return nil, fmt.Errorf("error parsing data for PKIKeyPairTask %q: %v", name, err)
		}
	}

	if meta != "" {
		return nil, fmt.Errorf("meta is not supported for PKIKeyPairTask")
	}

	return t, nil
}

func (t *PKIKeyPairTask) Run(c *fi.Context) error {
	castore := c.CAStore
	cert, err := castore.FindCert(t.Name)
	if err != nil {
		return err
	}
	if cert != nil {
		key, err := castore.FindPrivateKey(t.Name)
		if err != nil {
			return err
		}
		if key == nil {
			return fmt.Errorf("found cert in store, but did not find keypair: %q", t.Name)
		}
	}

	if cert == nil {
		glog.V(2).Infof("Creating PKI keypair %q", t.Name)

		template := &x509.Certificate{
			Subject:               *t.Subject,
			BasicConstraintsValid: true,
			IsCA: false,
		}

		if len(t.Subject.ToRDNSequence()) == 0 {
			return fmt.Errorf("Subject name was empty for SSL keypair %q", t.Name)
		}

		switch t.Type {
		case CertificateType_Client:
			template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}
			template.KeyUsage = x509.KeyUsageDigitalSignature
			break

		case CertificateType_Server:
			template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
			template.KeyUsage = x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment
			break

		default:
			return fmt.Errorf("unknown certificate type: %q", t.Type)
		}

		for _, san := range t.AlternateNames {
			san = strings.TrimSpace(san)
			if san == "" {
				continue
			}
			if ip := net.ParseIP(san); ip != nil {
				template.IPAddresses = append(template.IPAddresses, ip)
			} else {
				template.DNSNames = append(template.DNSNames, san)
			}
		}

		privateKey, err := castore.CreatePrivateKey(t.Name)
		if err != nil {
			return err
		}
		cert, err = castore.IssueCert(t.Name, privateKey, template)
		if err != nil {
			return err
		}
	}

	// TODO: Check correct subject / flags

	return nil
}
