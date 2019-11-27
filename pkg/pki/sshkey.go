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
	"crypto/md5"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"reflect"
	"strings"

	"golang.org/x/crypto/ssh"
)

// parseSSHPublicKey parses the SSH public key string
func parseSSHPublicKey(publicKey string) (ssh.PublicKey, error) {
	tokens := strings.Fields(publicKey)
	if len(tokens) < 2 {
		return nil, fmt.Errorf("error parsing SSH public key: %q", publicKey)
	}

	sshPublicKeyBytes, err := base64.StdEncoding.DecodeString(tokens[1])
	if err != nil {
		return nil, fmt.Errorf("error decoding SSH public key: %q err: %s", publicKey, err)
	}
	if len(tokens) < 2 {
		return nil, fmt.Errorf("error decoding SSH public key: %q", publicKey)
	}

	sshPublicKey, err := ssh.ParsePublicKey(sshPublicKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("error parsing SSH public key: %v", err)
	}
	return sshPublicKey, nil
}

// colonSeparatedHex formats the byte slice SSH-fingerprint style: hex bytes separated by colons
func colonSeparatedHex(data []byte) string {
	sshKeyFingerprint := fmt.Sprintf("%x", data)
	var colonSeparated bytes.Buffer
	for i := 0; i < len(sshKeyFingerprint); i++ {
		if (i%2) == 0 && i != 0 {
			colonSeparated.WriteByte(':')
		}
		colonSeparated.WriteByte(sshKeyFingerprint[i])
	}

	return colonSeparated.String()
}

// ComputeAWSKeyFingerprint computes the AWS-specific fingerprint of the SSH public key
func ComputeAWSKeyFingerprint(publicKey string) (string, error) {
	sshPublicKey, err := parseSSHPublicKey(publicKey)
	if err != nil {
		return "", err
	}

	der, err := toDER(sshPublicKey)
	if err != nil {
		return "", fmt.Errorf("error computing fingerprint for SSH public key: %v", err)
	}
	h := md5.Sum(der)

	return colonSeparatedHex(h[:]), nil
}

// ComputeOpenSSHKeyFingerprint computes the OpenSSH fingerprint of the SSH public key
func ComputeOpenSSHKeyFingerprint(publicKey string) (string, error) {
	sshPublicKey, err := parseSSHPublicKey(publicKey)
	if err != nil {
		return "", err
	}

	h := md5.Sum(sshPublicKey.Marshal())
	return colonSeparatedHex(h[:]), nil
}

// toDER gets the DER encoding of the SSH public key
// Annoyingly, the ssh code wraps the actual crypto keys, so we have to use reflection tricks
func toDER(pubkey ssh.PublicKey) ([]byte, error) {
	pubkeyValue := reflect.ValueOf(pubkey)
	typeName := fmt.Sprintf("%T", pubkey)

	var cryptoKey crypto.PublicKey
	switch typeName {
	case "*ssh.rsaPublicKey":
		var rsaPublicKey *rsa.PublicKey
		targetType := reflect.ValueOf(rsaPublicKey).Type()
		rsaPublicKey = pubkeyValue.Convert(targetType).Interface().(*rsa.PublicKey)
		cryptoKey = rsaPublicKey

	//case "*dsaPublicKey":
	//	var dsaPublicKey *dsa.PublicKey
	//	targetType := reflect.ValueOf(dsaPublicKey).Type()
	//	dsaPublicKey = pubkeyValue.Convert(targetType).Interface().(*dsa.PublicKey)
	//	cryptoKey = dsaPublicKey

	default:
		return nil, fmt.Errorf("Unexpected type of SSH key (%q); AWS can only import RSA keys", typeName)
	}

	der, err := x509.MarshalPKIXPublicKey(cryptoKey)
	if err != nil {
		return nil, fmt.Errorf("error marshaling SSH public key: %v", err)
	}
	return der, nil
}
