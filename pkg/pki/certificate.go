/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pki

import (
	"bytes"
	"crypto"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"os"

	"k8s.io/klog"
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
		r, err := ParsePEMCertificate([]byte(s))
		if err != nil {
			// Alternative form: Check if base64 encoded
			// TODO: Do we need this?  I think we need this only on nodeup, but maybe we could just not base64-it?
			d, err2 := base64.StdEncoding.DecodeString(s)
			if err2 == nil {
				r2, err2 := ParsePEMCertificate(d)
				if err2 == nil {
					klog.Warningf("used base64 decode of certificate")
					r = r2
					err = nil
				}
			}

			if err != nil {
				klog.Infof("Invalid certificate data: %q", string(b))
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

func ParsePEMCertificate(pemData []byte) (*Certificate, error) {
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

var _ io.WriterTo = &Certificate{}

func parsePEMCertificate(pemData []byte) (*x509.Certificate, error) {
	for {
		block, rest := pem.Decode(pemData)
		if block == nil {
			return nil, fmt.Errorf("could not parse certificate")
		}

		if block.Type == "CERTIFICATE" {
			klog.V(10).Infof("Parsing pem block: %q", block.Type)
			return x509.ParseCertificate(block.Bytes)
		}
		klog.Infof("Ignoring unexpected PEM block: %q", block.Type)

		pemData = rest
	}
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

func (c *Certificate) AsBytes() ([]byte, error) {
	// Nicer behaviour because this is called from templates
	if c == nil {
		return nil, fmt.Errorf("AsBytes called on nil Certificate")
	}

	var data bytes.Buffer
	_, err := c.WriteTo(&data)
	if err != nil {
		return nil, fmt.Errorf("error writing SSL certificate: %v", err)
	}
	return data.Bytes(), nil
}

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

func (c *Certificate) WriteToFile(filename string, perm os.FileMode) error {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	_, err = c.WriteTo(f)
	if err1 := f.Close(); err == nil {
		err = err1
	}
	return err
}
