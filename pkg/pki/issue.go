/*
Copyright 2020 The Kubernetes Authors.

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
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"hash/fnv"
	"math/big"
	"net"
	"sort"
	"strings"
	"time"

	"k8s.io/klog"
)

var wellKnownCertificateTypes = map[string]string{
	"ca":           "CA,KeyUsageCRLSign,KeyUsageCertSign",
	"client":       "ExtKeyUsageClientAuth,KeyUsageDigitalSignature",
	"clientServer": "ExtKeyUsageClientAuth,ExtKeyUsageServerAuth,KeyUsageDigitalSignature,KeyUsageKeyEncipherment",
	"server":       "ExtKeyUsageServerAuth,KeyUsageDigitalSignature,KeyUsageKeyEncipherment",
}

type IssueCertRequest struct {
	// Signer is the keypair to use to sign. Ignored if Type is "CA", in which case the cert will be self-signed.
	Signer string
	// Type is the type of certificate i.e. CA, server, client etc.
	Type string
	// Subject is the certificate subject.
	Subject pkix.Name
	// AlternateNames is a list of alternative names for this certificate.
	AlternateNames []string

	// PrivateKey is the private key for this certificate. If nil, a new private key will be generated.
	PrivateKey *PrivateKey
	// MinValidDays is the lower bound on the certificate validity, in days. If specified, up to 30 days
	// will be added so that certificate generated at the same time on different hosts will be unlikely to
	// expire at the same time. The default is 10 years (without the up to 30 day skew).
	MinValidDays int

	// Serial is the certificate serial number. If nil, a random number will be generated.
	Serial *big.Int
}

type Keystore interface {
	// FindKeypair finds a cert & private key, returning nil where either is not found
	// (if the certificate is found but not keypair, that is not an error: only the cert will be returned).
	// This func returns a cert, private key and a bool.  The bool value is whether the keypair is stored
	// in a legacy format. This bool is used by a keypair
	// task to convert a Legacy Keypair to the new Keypair API format.
	FindKeypair(name string) (*Certificate, *PrivateKey, bool, error)
}

// IssueCert issues a certificate, either a self-signed CA or from a CA in a keystore.
func IssueCert(request *IssueCertRequest, keystore Keystore) (issuedCertificate *Certificate, issuedKey *PrivateKey, caCertificate *Certificate, err error) {
	certificateType := request.Type
	if expanded, found := wellKnownCertificateTypes[certificateType]; found {
		certificateType = expanded
	}

	template := &x509.Certificate{
		BasicConstraintsValid: true,
		IsCA:                  false,
		SerialNumber:          request.Serial,
	}

	tokens := strings.Split(certificateType, ",")
	for _, t := range tokens {
		if strings.HasPrefix(t, "KeyUsage") {
			ku, found := parseKeyUsage(t)
			if !found {
				return nil, nil, nil, fmt.Errorf("unrecognized certificate option: %v", t)
			}
			template.KeyUsage |= ku
		} else if strings.HasPrefix(t, "ExtKeyUsage") {
			ku, found := parseExtKeyUsage(t)
			if !found {
				return nil, nil, nil, fmt.Errorf("unrecognized certificate option: %v", t)
			}
			template.ExtKeyUsage = append(template.ExtKeyUsage, ku)
		} else if t == "CA" {
			template.IsCA = true
		} else {
			return nil, nil, nil, fmt.Errorf("unrecognized certificate option: %q", t)
		}
	}

	template.Subject = request.Subject

	var alternateNames []string
	alternateNames = append(alternateNames, request.AlternateNames...)

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

	var caPrivateKey *PrivateKey
	var signer *x509.Certificate
	if !template.IsCA {
		var err error
		caCertificate, caPrivateKey, _, err = keystore.FindKeypair(request.Signer)
		if err != nil {
			return nil, nil, nil, err
		}
		if caPrivateKey == nil {
			return nil, nil, nil, fmt.Errorf("ca key for %q was not found; cannot issue certificates", request.Signer)
		}
		if caCertificate == nil {
			return nil, nil, nil, fmt.Errorf("ca certificate for %q was not found; cannot issue certificates", request.Signer)
		}
		signer = caCertificate.Certificate
	}

	privateKey := request.PrivateKey
	if privateKey == nil {
		var err error
		privateKey, err = GeneratePrivateKey()
		if err != nil {
			return nil, nil, nil, err
		}
	}

	// Skew the certificate lifetime by up to 30 days based on information about the generating node.
	// This is so that different nodes created at the same time have the certificates they generated
	// expire at different times, but all certificates on a given node expire around the same time.
	if request.MinValidDays != 0 {
		hash := fnv.New32()
		addrs, err := net.InterfaceAddrs()
		sort.Slice(addrs, func(i, j int) bool {
			return addrs[i].String() < addrs[j].String()
		})
		if err == nil {
			for _, addr := range addrs {
				_, _ = hash.Write([]byte(addr.String()))
			}
		} else {
			klog.Warningf("cannot skew certificate lifetime: failed to get interface addresses: %v", err)
		}
		template.NotAfter = time.Now().Add(time.Hour * 24 * time.Duration(request.MinValidDays)).Add(time.Hour * time.Duration(hash.Sum32()%(30*24))).UTC()
	}

	certificate, err := signNewCertificate(privateKey, template, signer, caPrivateKey)
	if err != nil {
		return nil, nil, nil, err
	}

	if signer == nil {
		caCertificate = certificate
	}

	return certificate, privateKey, caCertificate, err
}
