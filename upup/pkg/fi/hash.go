package fi

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"github.com/golang/glog"
	"hash"
)

type HashAlgorithm string

const (
	HashAlgorithmSHA256 = "sha256"
	HashAlgorithmSHA1   = "sha1"
	HashAlgorithmMD5    = "md5"
)

func NewHasher(hashAlgorithm HashAlgorithm) hash.Hash {
	switch hashAlgorithm {
	case HashAlgorithmMD5:
		return md5.New()

	case HashAlgorithmSHA1:
		return sha1.New()

	case HashAlgorithmSHA256:
		return sha256.New()
	}

	glog.Exitf("Unknown hash algorithm: %v", hashAlgorithm)
	return nil
}

func determineHashAlgorithm(hash string) (HashAlgorithm, error) {
	if len(hash) == 32 {
		return HashAlgorithmMD5, nil
	} else if len(hash) == 40 {
		return HashAlgorithmSHA1, nil
	} else if len(hash) == 64 {
		return HashAlgorithmSHA256, nil
	} else {
		return "", fmt.Errorf("Unrecognized hash format: %q", hash)
	}
}
