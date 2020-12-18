package gce

import (
	"fmt"
	"crypto/sha256"
	"encoding/base64"
	"net/url"
	
	"k8s.io/kops/upup/pkg/fi"
	"cloud.google.com/go/compute/metadata"
)

type gceAuthenticator struct {
	audience string
}

var _ fi.Authenticator = &gceAuthenticator{}

func NewAuthenticator(audience string) (fi.Authenticator, error) {
	return &gceAuthenticator{
		audience: audience,
	}, nil
}

func (a *gceAuthenticator) CreateToken(body []byte) (string, error) {
	sha := sha256.Sum256(body)

	// Ensure the signature is only valid for this particular body content.
	audience := a.audience + "//" + base64.URLEncoding.EncodeToString(sha[:])

	suffix := "instance/service-accounts/default/identity?format=full&audience=" + url.QueryEscape(audience)

token, err := metadata.Get(suffix)
	if err != nil {
		return "", fmt.Errorf("unable to get token from GCE metadata service: %w", err)
	}

	return token, nil
}
