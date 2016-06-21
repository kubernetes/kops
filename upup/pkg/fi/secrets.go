package fi

import (
	crypto_rand "crypto/rand"
	"encoding/base64"
	"fmt"
	"k8s.io/kube-deploy/upup/pkg/fi/vfs"
	"strings"
)

type SecretStore interface {
	// Get a secret.  Returns an error if not found
	Secret(id string) (*Secret, error)
	// Find a secret, if exists.  Returns nil,nil if not found
	FindSecret(id string) (*Secret, error)
	// Create or replace a secret
	GetOrCreateSecret(id string) (secret *Secret, created bool, err error)
	// Lists the ids of all known secrets
	ListSecrets() ([]string, error)

	// VFSPath returns the path where the SecretStore is stored
	VFSPath() vfs.Path
}

type Secret struct {
	Data []byte
}

func (s *Secret) AsString() (string, error) {
	// Nicer behaviour because this is called from templates
	if s == nil {
		return "", fmt.Errorf("AsString called on nil Secret")
	}

	return string(s.Data), nil
}

func CreateSecret() (*Secret, error) {
	data := make([]byte, 128)
	_, err := crypto_rand.Read(data)
	if err != nil {
		return nil, fmt.Errorf("error reading crypto_rand: %v", err)
	}

	s := base64.StdEncoding.EncodeToString(data)
	r := strings.NewReplacer("+", "", "=", "", "/", "")
	s = r.Replace(s)
	s = s[:32]

	return &Secret{
		Data: []byte(s),
	}, nil
}
