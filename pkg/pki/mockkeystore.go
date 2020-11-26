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
	"fmt"
)

type MockKeystore struct {
	Signer string
	cert   *Certificate
	key    *PrivateKey

	invoked bool
}

func (m *MockKeystore) FindKeypair(name string) (*Certificate, *PrivateKey, bool, error) {
	m.invoked = true
	return m.cert, m.key, false, nil
}

func NewMockKeystore() (*MockKeystore, error) {

	caCertificate, err := ParsePEMCertificate([]byte("-----BEGIN CERTIFICATE-----\nMIIBRjCB8aADAgECAhAzhRMOcwfggPtgZNIOFU19MA0GCSqGSIb3DQEBCwUAMBIx\nEDAOBgNVBAMTB1Rlc3QgQ0EwHhcNMjAwNTE1MDIzNjI0WhcNMzAwNTE1MDIzNjI0\nWjASMRAwDgYDVQQDEwdUZXN0IENBMFwwDQYJKoZIhvcNAQEBBQADSwAwSAJBAM/S\ncagGaiDA3jJWBXUr8rM19TWLA65jK/iA05FCsmQbyvETs5gbJdBfnhQp8wkKFlkt\nKxZ34k3wQUzoB1lv8/kCAwEAAaMjMCEwDgYDVR0PAQH/BAQDAgEGMA8GA1UdEwEB\n/wQFMAMBAf8wDQYJKoZIhvcNAQELBQADQQCDOxvs58AVAWgWLtD3Obvy7XXsKx6d\nMzg9epbiQchLE4G/jlbgVu7vwh8l5XFNfQooG6stCU7pmLFXkXzkJQxr\n-----END CERTIFICATE-----\n"))
	if err != nil {
		return nil, fmt.Errorf("reading certificate: %v", err)
	}
	caPrivateKey, err := ParsePEMPrivateKey([]byte("-----BEGIN RSA PRIVATE KEY-----\nMIIBPAIBAAJBAM/ScagGaiDA3jJWBXUr8rM19TWLA65jK/iA05FCsmQbyvETs5gb\nJdBfnhQp8wkKFlktKxZ34k3wQUzoB1lv8/kCAwEAAQJBAJzXQZeBX87gP9DVQsEv\nLbc6XZjPFTQi/ChLcWALaf5J7drFJHUcWbKIHzOmM3fm3lQlb/1IcwOBU5cTY0e9\nBVECIQD73kxOWWAIzKqMOvFZ9s79Et7G1HUMnVAVKJ1NS1uvYwIhANM7LULdi0YD\nbcHvDl3+Msj4cPH7CXAJFyPWaQZPlXPzAiEAhDg6jpbUl0n57guzT6sFFk2lrXMy\nzyB2PeVITp9UzkkCIEpcF7flQ+U2ycmuvVELbpdfFmupIw5ktNex4DEPjR5PAiEA\n68vR1L1Kaja/GzU76qAQaYA/V1Ag4sPmOQdEaVZKu78=\n-----END RSA PRIVATE KEY-----\n"))
	if err != nil {
		return nil, fmt.Errorf("parsing key: %v", err)
	}

	keystore := &MockKeystore{
		cert: caCertificate,
		key:  caPrivateKey,
	}

	return keystore, nil
}
