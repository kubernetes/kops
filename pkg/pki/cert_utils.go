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
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"sort"
	"strings"

	"k8s.io/klog/v2"
)

func PkixNameToString(name *pkix.Name) string {
	seq := name.ToRDNSequence()
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

var keyUsageStrings = map[x509.KeyUsage]string{
	x509.KeyUsageDigitalSignature:  "KeyUsageDigitalSignature",
	x509.KeyUsageContentCommitment: "KeyUsageContentCommitment",
	x509.KeyUsageKeyEncipherment:   "KeyUsageKeyEncipherment",
	x509.KeyUsageDataEncipherment:  "KeyUsageDataEncipherment",
	x509.KeyUsageKeyAgreement:      "KeyUsageKeyAgreement",
	x509.KeyUsageCertSign:          "KeyUsageCertSign",
	x509.KeyUsageCRLSign:           "KeyUsageCRLSign",
	x509.KeyUsageEncipherOnly:      "KeyUsageEncipherOnly",
	x509.KeyUsageDecipherOnly:      "KeyUsageDecipherOnly",
}

func keyUsageToString(u x509.KeyUsage) []string {
	var usages []string

	for k, v := range keyUsageStrings {
		if (u & k) != 0 {
			usages = append(usages, v)
		}
	}

	// TODO: Detect if there are other flags set?
	return usages
}

func parseKeyUsage(s string) (x509.KeyUsage, bool) {
	for k, v := range keyUsageStrings {
		if v == s {
			return k, true
		}
	}
	return 0, false
}

var extKeyUsageStrings = map[x509.ExtKeyUsage]string{
	x509.ExtKeyUsageAny:                        "ExtKeyUsageAny",
	x509.ExtKeyUsageServerAuth:                 "ExtKeyUsageServerAuth",
	x509.ExtKeyUsageClientAuth:                 "ExtKeyUsageClientAuth",
	x509.ExtKeyUsageCodeSigning:                "ExtKeyUsageCodeSigning",
	x509.ExtKeyUsageEmailProtection:            "ExtKeyUsageEmailProtection",
	x509.ExtKeyUsageIPSECEndSystem:             "ExtKeyUsageIPSECEndSystem",
	x509.ExtKeyUsageIPSECTunnel:                "ExtKeyUsageIPSECTunnel",
	x509.ExtKeyUsageIPSECUser:                  "ExtKeyUsageIPSECUser",
	x509.ExtKeyUsageTimeStamping:               "ExtKeyUsageTimeStamping",
	x509.ExtKeyUsageOCSPSigning:                "ExtKeyUsageOCSPSigning",
	x509.ExtKeyUsageMicrosoftServerGatedCrypto: "ExtKeyUsageMicrosoftServerGatedCrypto",
	x509.ExtKeyUsageNetscapeServerGatedCrypto:  "ExtKeyUsageNetscapeServerGatedCrypto",
}

func extKeyUsageToString(u x509.ExtKeyUsage) string {
	s := extKeyUsageStrings[u]
	if s == "" {
		klog.Warningf("Unhandled ExtKeyUsage: %v", u)
		s = fmt.Sprintf("ExtKeyUsage:%v", u)
	}
	return s
}

func parseExtKeyUsage(s string) (x509.ExtKeyUsage, bool) {
	for k, v := range extKeyUsageStrings {
		if v == s {
			return k, true
		}
	}
	return 0, false
}

// BuildTypeDescription extracts the type based on the certificate extensions
func BuildTypeDescription(cert *x509.Certificate) string {
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
