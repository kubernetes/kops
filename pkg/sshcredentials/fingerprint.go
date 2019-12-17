/*
Copyright 2017 The Kubernetes Authors.

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

package sshcredentials

import (
	"bytes"
	"crypto/md5"
	"fmt"

	"golang.org/x/crypto/ssh"
	"k8s.io/klog"
)

func Fingerprint(pubkey string) (string, error) {
	sshPublicKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(pubkey))
	if err != nil {
		return "", fmt.Errorf("error parsing SSH public key: %v", err)
	}

	// compute fingerprint to serve as id
	h := md5.New()
	_, err = h.Write(sshPublicKey.Marshal())
	if err != nil {
		return "", fmt.Errorf("error fingerprinting SSH public key: %v", err)
	}
	id := formatFingerprint(h.Sum(nil))
	return id, nil
}

func formatFingerprint(data []byte) string {
	var buf bytes.Buffer

	for i, b := range data {
		s := fmt.Sprintf("%0.2x", b)
		if i != 0 {
			buf.WriteString(":")
		}
		buf.WriteString(s)
	}
	return buf.String()
}

func insertFingerprintColons(id string) string {
	remaining := id

	var buf bytes.Buffer
	for {
		if remaining == "" {
			break
		}
		if buf.Len() != 0 {
			buf.WriteString(":")
		}
		if len(remaining) < 2 {
			klog.Warningf("unexpected format for SSH public key id: %q", id)
			buf.WriteString(remaining)
			break
		} else {
			buf.WriteString(remaining[0:2])
			remaining = remaining[2:]
		}
	}
	return buf.String()
}
