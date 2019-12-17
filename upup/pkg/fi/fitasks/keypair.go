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
	"crypto/x509"
	"fmt"
	"net"
	"sort"
	"strings"

	"k8s.io/klog"
	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/upup/pkg/fi"
)

var wellKnownCertificateTypes = map[string]string{
	"ca":           "CA,KeyUsageCRLSign,KeyUsageCertSign",
	"client":       "ExtKeyUsageClientAuth,KeyUsageDigitalSignature",
	"clientServer": "ExtKeyUsageClientAuth,ExtKeyUsageServerAuth,KeyUsageDigitalSignature,KeyUsageKeyEncipherment",
	"server":       "ExtKeyUsageServerAuth,KeyUsageDigitalSignature,KeyUsageKeyEncipherment",
}

//go:generate fitask -type=Keypair
type Keypair struct {
	// Name is the name of the keypair
	Name *string
	// AlternateNames a list of alternative names for this certificate
	AlternateNames []string `json:"alternateNames"`
	// AlternateNameTasks is a collection of subtask
	AlternateNameTasks []fi.Task `json:"alternateNameTasks"`
	// Lifecycle is context for a task
	Lifecycle *fi.Lifecycle
	// Signer is the keypair to use to sign, for when we want to use an alternative CA
	Signer *Keypair
	// Subject is the certificate subject
	Subject string `json:"subject"`
	// Type the type of certificate i.e. CA, server, client etc
	Type string `json:"type"`
	// Format stores the api version of kops.Keyset.  We are using this info in order to determine if kops
	// is accessing legacy secrets that do not use keyset.yaml.
	Format string `json:"format"`
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

	cert, key, format, err := c.Keystore.FindKeypair(name)
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
		Subject:        pkixNameToString(&cert.Subject),
		Type:           buildTypeDescription(cert.Certificate),
		Format:         string(format),
	}

	actual.Signer = &Keypair{Subject: pkixNameToString(&cert.Certificate.Issuer)}

	// Avoid spurious changes
	actual.Lifecycle = e.Lifecycle

	return actual, nil
}

func (e *Keypair) Run(c *fi.Context) error {
	err := e.normalize(c)
	if err != nil {
		return err
	}
	return fi.DefaultDeltaRunMethod(e, c)
}

func (e *Keypair) normalize(c *fi.Context) error {
	var alternateNames []string

	for _, s := range e.AlternateNames {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		alternateNames = append(alternateNames, s)
	}

	for _, task := range e.AlternateNameTasks {
		if hasAddress, ok := task.(fi.HasAddress); ok {
			address, err := hasAddress.FindIPAddress(c)
			if err != nil {
				return fmt.Errorf("error finding address for %v: %v", task, err)
			}
			if address == nil {
				klog.Warningf("Task did not have an address: %v", task)
				continue
			}
			klog.V(8).Infof("Resolved alternateName %q for %q", *address, task)
			alternateNames = append(alternateNames, *address)
		} else {
			return fmt.Errorf("Unsupported type for AlternateNameDependencies: %v", task)
		}
	}

	sort.Strings(alternateNames)
	e.AlternateNames = alternateNames
	e.AlternateNameTasks = nil

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

	template, err := e.BuildCertificateTemplate()
	if err != nil {
		return err
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
			klog.V(8).Infof("creating certificate new Type")
		} else if changes.Format != "" {
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

			privateKey, err = pki.GeneratePrivateKey()
			if err != nil {
				return err
			}
		}

		signer := fi.CertificateId_CA
		if e.Signer != nil {
			signer = fi.StringValue(e.Signer.Name)
		}

		cert, err := c.Keystore.CreateKeypair(signer, name, template, privateKey)
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

		klog.Infof("updated Keypair %q to API format %q", name, e.Format)
	}

	return nil
}

// BuildCertificateTemplate is responsible for constructing a certificate template
func (e *Keypair) BuildCertificateTemplate() (*x509.Certificate, error) {
	template, err := buildCertificateTemplateForType(e.Type)
	if err != nil {
		return nil, err
	}

	subjectPkix, err := parsePkixName(e.Subject)
	if err != nil {
		return nil, fmt.Errorf("error parsing Subject: %v", err)
	}

	if len(subjectPkix.ToRDNSequence()) == 0 {
		return nil, fmt.Errorf("Subject name was empty for SSL keypair %q", *e.Name)
	}

	template.Subject = *subjectPkix

	var alternateNames []string
	alternateNames = append(alternateNames, e.AlternateNames...)

	for _, san := range alternateNames {
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

	return template, nil
}

func buildCertificateTemplateForType(certificateType string) (*x509.Certificate, error) {
	if expanded, found := wellKnownCertificateTypes[certificateType]; found {
		certificateType = expanded
	}

	template := &x509.Certificate{
		BasicConstraintsValid: true,
		IsCA:                  false,
	}

	tokens := strings.Split(certificateType, ",")
	for _, t := range tokens {
		if strings.HasPrefix(t, "KeyUsage") {
			ku, found := parseKeyUsage(t)
			if !found {
				return nil, fmt.Errorf("unrecognized certificate option: %v", t)
			}
			template.KeyUsage |= ku
		} else if strings.HasPrefix(t, "ExtKeyUsage") {
			ku, found := parseExtKeyUsage(t)
			if !found {
				return nil, fmt.Errorf("unrecognized certificate option: %v", t)
			}
			template.ExtKeyUsage = append(template.ExtKeyUsage, ku)
		} else if t == "CA" {
			template.IsCA = true
		} else {
			return nil, fmt.Errorf("unrecognized certificate option: %q", t)
		}
	}

	return template, nil
}

// buildTypeDescription extracts the type based on the certificate extensions
func buildTypeDescription(cert *x509.Certificate) string {
	var options []string

	if cert.IsCA {
		options = append(options, "CA")
	}

	options = append(options, keyUsageToString(cert.KeyUsage)...)

	for _, extKeyUsage := range cert.ExtKeyUsage {
		options = append(options, extKeyUsageToString(extKeyUsage))
	}

	sort.Strings(options)
	s := strings.Join(options, ",")

	for k, v := range wellKnownCertificateTypes {
		if v == s {
			s = k
		}
	}

	return s
}
