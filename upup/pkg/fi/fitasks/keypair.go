/*
Copyright 2016 The Kubernetes Authors.

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

	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/fi"
)

var wellKnownCertificateTypes = map[string]string{
	"client": "ExtKeyUsageClientAuth,KeyUsageDigitalSignature",
	"server": "ExtKeyUsageServerAuth,KeyUsageDigitalSignature,KeyUsageKeyEncipherment",
}

//go:generate fitask -type=Keypair
type Keypair struct {
	Name               *string
	Lifecycle          *fi.Lifecycle
	Subject            string    `json:"subject"`
	Type               string    `json:"type"`
	AlternateNames     []string  `json:"alternateNames"`
	AlternateNameTasks []fi.Task `json:"alternateNameTasks"`
}

var _ fi.HasCheckExisting = &Keypair{}
var _ fi.HasName = &Keypair{}

// It's important always to check for the existing key, so we don't regenerate keys e.g. on terraform
func (e *Keypair) CheckExisting(c *fi.Context) bool {
	return true
}

func (e *Keypair) Find(c *fi.Context) (*Keypair, error) {
	name := fi.StringValue(e.Name)
	if name == "" {
		return nil, nil
	}

	cert, key, err := c.Keystore.FindKeypair(name)
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
		Subject:        pkixNameToString(&cert.Subject),
		AlternateNames: alternateNames,
		Type:           buildTypeDescription(cert.Certificate),
	}

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
				glog.Warningf("Task did not have an address: %v", task)
				continue
			}
			glog.V(8).Infof("Resolved alternateName %q for %q", *address, task)
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

func (s *Keypair) CheckChanges(a, e, changes *Keypair) error {
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

	createCertificate := false
	if a == nil {
		createCertificate = true
	} else if changes != nil {
		if changes.AlternateNames != nil {
			createCertificate = true
		} else if changes.Subject != "" {
			createCertificate = true
		} else {
			glog.Warningf("Ignoring changes in key: %v", fi.DebugAsJsonString(changes))
		}
	}

	if createCertificate {
		glog.V(2).Infof("Creating PKI keypair %q", name)

		cert, privateKey, err := c.Keystore.FindKeypair(name)
		if err != nil {
			return err
		}

		// We always reuse the private key if it exists,
		// if we change keys we often have to regenerate e.g. the service accounts
		// TODO: Eventually rotate keys / don't always reuse?
		if privateKey == nil {
			privateKey, err = fi.GeneratePrivateKey()
			if err != nil {
				return err
			}
		}

		cert, err = c.Keystore.CreateKeypair(name, template, privateKey)
		if err != nil {
			return err
		}

		glog.V(8).Infof("created certificate %v", cert)
	}

	// TODO: Check correct subject / flags

	return nil
}

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
		return nil, fmt.Errorf("Subject name was empty for SSL keypair %q", e.Name)
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
		IsCA: false,
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
		} else {
			return nil, fmt.Errorf("unrecognized certificate option: %v", t)
		}
	}

	return template, nil
}

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
