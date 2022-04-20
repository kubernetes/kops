/*
Copyright 2022 The Kubernetes Authors.

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

package util

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh"
)

// CreateSSHKeyPair creates a key pair in a temp directory
// and returns the paths to the private and public keys respectively.
// The file paths are deterministic from the clusterName.
func CreateSSHKeyPair(clusterName string) (string, string, error) {
	_, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		return "", "", err
	}
	publicKey, err := ssh.NewPublicKey(privateKey.Public())
	if err != nil {
		return "", "", err
	}

	publicKeyContents := ssh.MarshalAuthorizedKey(publicKey)

	user := os.Getenv("USER")
	if user == "" {
		user = "user"
	}
	comment := fmt.Sprintf(" %v\n", user)

	// AWS requires a comment on the SSH public key but MarshalAuthorizedKey doesn't create one
	publicKeyContents = publicKeyContents[:len(publicKeyContents)-1]
	publicKeyContents = append(publicKeyContents, []byte(comment)...)

	privateKeyContents, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return "", "", err
	}

	tmp := os.TempDir()
	dir := filepath.Join(tmp, "kops", clusterName)
	err = os.MkdirAll(dir, 0700)
	if err != nil {
		return "", "", err
	}

	publicKeyPath := filepath.Join(dir, "id_ed25519.pub")
	privateKeyPath := filepath.Join(dir, "id_ed25519")

	if _, err := os.Stat(privateKeyPath); errors.Is(err, os.ErrNotExist) {
		if err := os.WriteFile(publicKeyPath, publicKeyContents, 0644); err != nil {
			return "", "", err
		}
		f, err := os.OpenFile(privateKeyPath, os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			return "", "", err
		}
		defer f.Close()

		err = pem.Encode(f, &pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: privateKeyContents,
		})
		if err != nil {
			return "", "", err
		}
	}

	return publicKeyPath, privateKeyPath, nil
}
