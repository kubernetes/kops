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

package hashing

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"strings"

	"k8s.io/klog/v2"

	"k8s.io/kops/pkg/try"
)

type HashAlgorithm string

const (
	HashAlgorithmSHA256 HashAlgorithm = "sha256"
	HashAlgorithmSHA1   HashAlgorithm = "sha1"
	HashAlgorithmMD5    HashAlgorithm = "md5"
)

type Hash struct {
	Algorithm HashAlgorithm
	HashValue []byte
}

func (h *Hash) String() string {
	return string(h.Algorithm) + ":" + h.Hex()
}

func (h *Hash) Hex() string {
	return hex.EncodeToString(h.HashValue)
}

func (ha HashAlgorithm) NewHasher() hash.Hash {
	switch ha {
	case HashAlgorithmMD5:
		return md5.New()

	case HashAlgorithmSHA1:
		return sha1.New()

	case HashAlgorithmSHA256:
		return sha256.New()
	}

	klog.Exitf("Unknown hash algorithm: %v", ha)
	return nil
}

func (ha HashAlgorithm) FromString(s string) (*Hash, error) {
	l := -1
	switch ha {
	case HashAlgorithmMD5:
		l = 32
	case HashAlgorithmSHA1:
		l = 40
	case HashAlgorithmSHA256:
		l = 64
	default:
		return nil, fmt.Errorf("unknown hash algorithm: %q", ha)
	}

	if len(s) != l {
		return nil, fmt.Errorf("invalid %q hash - unexpected length %d", ha, len(s))
	}

	hashValue, err := hex.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("invalid hash %q - not hex", s)
	}
	return &Hash{Algorithm: ha, HashValue: hashValue}, nil
}

func FromString(s string) (*Hash, error) {
	for _, ha := range []HashAlgorithm{HashAlgorithmMD5, HashAlgorithmSHA1, HashAlgorithmSHA256} {
		prefix := fmt.Sprintf("%s:", ha)
		if strings.HasPrefix(s, prefix) {
			return ha.FromString(s[len(prefix):])
		}
	}

	var ha HashAlgorithm
	switch len(s) {
	case 32:
		ha = HashAlgorithmMD5
	case 40:
		ha = HashAlgorithmSHA1
	case 64:
		ha = HashAlgorithmSHA256
	default:
		return nil, fmt.Errorf("cannot determine algorithm for hash length: %d", len(s))
	}

	return ha.FromString(s)
}

func (ha HashAlgorithm) Hash(r io.Reader) (*Hash, error) {
	hasher := ha.NewHasher()
	_, err := copyToHasher(hasher, r)
	if err != nil {
		return nil, fmt.Errorf("error while hashing resource: %v", err)
	}
	return &Hash{Algorithm: ha, HashValue: hasher.Sum(nil)}, nil
}

func (ha HashAlgorithm) HashFile(p string) (*Hash, error) {
	f, err := os.OpenFile(p, os.O_RDONLY, 0)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, err
		}
		return nil, fmt.Errorf("error opening file %q: %v", p, err)
	}
	defer try.CloseFile(f)
	return ha.Hash(f)
}

func copyToHasher(dest io.Writer, src io.Reader) (int64, error) {
	n, err := io.Copy(dest, src)
	if err != nil {
		return n, fmt.Errorf("error hashing data: %v", err)
	}
	return n, nil
}

func (l *Hash) Equal(r *Hash) bool {
	return (l.Algorithm == r.Algorithm) && bytes.Equal(l.HashValue, r.HashValue)
}
