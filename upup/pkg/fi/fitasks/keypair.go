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

package fitasks

import (
	"crypto/sha1"
	"crypto/x509/pkix"
	"fmt"
	"sort"
	"strings"
	"time"

	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/upup/pkg/fi"
)

// +kops:fitask
type Keypair struct {
	// Name is the name of the keypair
	Name *string
	// AlternateNames a list of alternative names for this certificate
	AlternateNames []string `json:"alternateNames"`
	// Lifecycle is context for a task
	Lifecycle *fi.Lifecycle
	// Signer is the keypair to use to sign, for when we want to use an alternative CA
	Signer *Keypair
	// Subject is the certificate subject
	Subject string `json:"subject"`
	// Type the type of certificate i.e. CA, server, client etc
	Type string `json:"type"`
	// LegacyFormat is whether the keypair is stored in a legacy format.
	LegacyFormat bool `json:"oldFormat"`

	certificate                *fi.TaskDependentResource
	certificateSHA1Fingerprint *fi.TaskDependentResource
}

var _ fi.HasCheckExisting = &Keypair{}
var _ fi.HasName = &Keypair{}

// It's important always to check for the existing key, so we don't regenerate keys e.g. on terraform
func (e *Keypair) CheckExisting(c *fi.Context) bool {
	return true
}

var _ fi.CompareWithID = &Keypair{}

func (e *Keypair) CompareWithID() *string {
	return &e.Subject
}

func (e *Keypair) Find(c *fi.Context) (*Keypair, error) {
	name := fi.StringValue(e.Name)
	if name == "" {
		return nil, nil
	}

	cert, key, legacyFormat, err := c.Keystore.FindKeypair(name)
	if err != nil {
		return nil, err
	}
	if cert == nil {
		return nil, nil
	}
	if key == nil {
		return nil, fmt.Errorf("found cert in store, but did not find private key: %q", name)
	}

	var alternateNames []string
	alternateNames = append(alternateNames, cert.Certificate.DNSNames...)
	alternateNames = append(alternateNames, cert.Certificate.EmailAddresses...)
	for _, ip := range cert.Certificate.IPAddresses {
		alternateNames = append(alternateNames, ip.String())
	}
	sort.Strings(alternateNames)

	actual := &Keypair{
		Name:           &name,
		AlternateNames: alternateNames,
		Subject:        pki.PkixNameToString(&cert.Subject),
		Type:           pki.BuildTypeDescription(cert.Certificate),
		LegacyFormat:   legacyFormat,
	}

	actual.Signer = &Keypair{Subject: pki.PkixNameToString(&cert.Certificate.Issuer)}

	// Avoid spurious changes
	actual.Lifecycle = e.Lifecycle

	if err := e.setResources(cert); err != nil {
		return nil, fmt.Errorf("error setting resources: %v", err)
	}

	return actual, nil
}

func (e *Keypair) Run(c *fi.Context) error {
	err := e.normalize()
	if err != nil {
		return err
	}
	return fi.DefaultDeltaRunMethod(e, c)
}

func (e *Keypair) normalize() error {
	var alternateNames []string

	for _, s := range e.AlternateNames {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		alternateNames = append(alternateNames, s)
	}

	sort.Strings(alternateNames)
	e.AlternateNames = alternateNames

	return nil
}

func (_ *Keypair) CheckChanges(a, e, changes *Keypair) error {
	if a != nil {
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
	}
	return nil
}

func (_ *Keypair) Render(c *fi.Context, a, e, changes *Keypair) error {
	name := fi.StringValue(e.Name)
	if name == "" {
		return fi.RequiredField("Name")
	}

	changeStoredFormat := false
	createCertificate := false
	if a == nil {
		createCertificate = true
		klog.V(8).Infof("creating brand new certificate")
	} else if changes != nil {
		klog.V(8).Infof("creating certificate as changes are not nil")
		if changes.AlternateNames != nil {
			createCertificate = true
			klog.V(8).Infof("creating certificate new AlternateNames")
		} else if changes.Subject != "" {
			createCertificate = true
			klog.V(8).Infof("creating certificate new Subject")
		} else if changes.Type != "" {
			createCertificate = true
			klog.Infof("creating certificate %q as Type has changed (actual=%v, expected=%v)", name, a.Type, e.Type)
		} else if changes.LegacyFormat {
			changeStoredFormat = true
		} else {
			klog.Warningf("Ignoring changes in key: %v", fi.DebugAsJsonString(changes))
		}
	}

	if createCertificate {
		klog.V(2).Infof("Creating PKI keypair %q", name)

		_, privateKey, _, err := c.Keystore.FindKeypair(name)
		if err != nil {
			return err
		}

		// We always reuse the private key if it exists,
		// if we change keys we often have to regenerate e.g. the service accounts
		// TODO: Eventually rotate keys / don't always reuse?
		if privateKey == nil {
			klog.V(2).Infof("Creating privateKey %q", name)
		}

		signer := fi.CertificateIDCA
		if e.Signer != nil {
			signer = fi.StringValue(e.Signer.Name)
		}

		klog.Infof("Issuing new certificate: %q", *e.Name)

		serial := pki.BuildPKISerial(time.Now().UnixNano())

		subjectPkix, err := parsePkixName(e.Subject)
		if err != nil {
			return fmt.Errorf("error parsing Subject: %v", err)
		}

		if len(subjectPkix.ToRDNSequence()) == 0 {
			return fmt.Errorf("subject name was empty for SSL keypair %q", *e.Name)
		}

		req := pki.IssueCertRequest{
			Signer:         signer,
			Type:           e.Type,
			Subject:        *subjectPkix,
			AlternateNames: e.AlternateNames,
			PrivateKey:     privateKey,
			Serial:         serial,
		}
		cert, privateKey, _, err := pki.IssueCert(&req, c.Keystore)
		if err != nil {
			return err
		}
		err = c.Keystore.StoreKeypair(name, cert, privateKey)
		if err != nil {
			return err
		}

		if err := e.setResources(cert); err != nil {
			return fmt.Errorf("error setting resources: %v", err)
		}

		// Make double-sure it round-trips
		_, _, _, err = c.Keystore.FindKeypair(name)
		if err != nil {
			return err
		}

		klog.V(8).Infof("created certificate with cn=%s", cert.Subject.CommonName)
	}

	// TODO: Check correct subject / flags

	if changeStoredFormat {
		// We fetch and reinsert the same keypair, forcing an update to our preferred format
		// TODO: We're assuming that we want to save in the preferred format
		cert, privateKey, _, err := c.Keystore.FindKeypair(name)
		if err != nil {
			return err
		}
		err = c.Keystore.StoreKeypair(name, cert, privateKey)
		if err != nil {
			return err
		}

		klog.Infof("updated Keypair %q to new format", name)
	}

	return nil
}

func parsePkixName(s string) (*pkix.Name, error) {
	name := new(pkix.Name)

	tokens := strings.Split(s, ",")
	for _, token := range tokens {
		token = strings.TrimSpace(token)
		kv := strings.SplitN(token, "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("unrecognized token (expected k=v): %q", token)
		}
		k := strings.ToLower(kv[0])
		v := kv[1]

		switch k {
		case "cn":
			name.CommonName = v
		case "o":
			name.Organization = append(name.Organization, v)
		default:
			return nil, fmt.Errorf("unrecognized key %q in token %q", k, token)
		}
	}

	return name, nil
}

func (e *Keypair) ensureResources() {
	if e.certificate == nil {
		e.certificate = &fi.TaskDependentResource{Task: e}
	}
	if e.certificateSHA1Fingerprint == nil {
		e.certificateSHA1Fingerprint = &fi.TaskDependentResource{Task: e}
	}
}

func (e *Keypair) setResources(cert *pki.Certificate) error {
	e.ensureResources()

	s, err := cert.AsString()
	if err != nil {
		return err
	}
	e.certificate.Resource = fi.NewStringResource(s)

	fingerprint := sha1.Sum(cert.Certificate.Raw)
	hex := fmt.Sprintf("%x", fingerprint)
	e.certificateSHA1Fingerprint.Resource = fi.NewStringResource(hex)

	return nil
}

func (e *Keypair) CertificateSHA1Fingerprint() fi.Resource {
	e.ensureResources()
	return e.certificateSHA1Fingerprint
}
