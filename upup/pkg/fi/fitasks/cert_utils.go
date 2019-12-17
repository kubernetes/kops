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
	"bytes"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"strings"

	"k8s.io/klog"
)

func pkixNameToString(name *pkix.Name) string {
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
