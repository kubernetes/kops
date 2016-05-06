package fi

import (
	crypto_rand "crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
)

type SecretStore interface {
	Secret(id string) (*Secret, error)
	FindSecret(id string) (*Secret, error)
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
