package hashing

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/golang/glog"
	"hash"
	"io"
	"os"
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
	return fmt.Sprintf("%s:%s", h.Algorithm, hex.EncodeToString(h.HashValue))
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

	glog.Exitf("Unknown hash algorithm: %v", ha)
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
		return nil, fmt.Errorf("invalid %q hash - unexpected length %s", ha, len(s))
	}

	hashValue, err := hex.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("invalid hash %q - not hex", s)
	}
	return &Hash{Algorithm: ha, HashValue: hashValue}, nil
}

func FromString(s string) (*Hash, error) {
	var ha HashAlgorithm
	switch len(s) {
	case 32:
		ha = HashAlgorithmMD5
	case 40:
		ha = HashAlgorithmSHA1
	case 64:
		ha = HashAlgorithmSHA256
	default:
		return nil, fmt.Errorf("cannot determine algorithm for hash: %d", len(s))
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
	defer f.Close()
	return ha.Hash(f)
}

func HashesForResource(r io.Reader, hashAlgorithms []HashAlgorithm) ([]*Hash, error) {
	var hashers []hash.Hash
	var writers []io.Writer
	for _, hashAlgorithm := range hashAlgorithms {
		hasher := hashAlgorithm.NewHasher()
		hashers = append(hashers, hasher)
		writers = append(writers, hasher)
	}

	w := io.MultiWriter(writers...)

	_, err := copyToHasher(w, r)
	if err != nil {
		return nil, fmt.Errorf("error while hashing resource: %v", err)
	}

	var hashes []*Hash
	for i, hasher := range hashers {
		hashes = append(hashes, &Hash{Algorithm: hashAlgorithms[i], HashValue: hasher.Sum(nil)})
	}

	return hashes, nil
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
