/*
Copyright 2018 The Kubernetes Authors.

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

package gce

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type gceIdentityData struct {
	Issuer          string                `json:"iss"`
	IssuedAt        int64                 `json:"iat"`
	Expiry          int64                 `json:"exp"`
	Audience        string                `json:"aud"`
	Subject         string                `json:"sub"`
	AuthorizedParty string                `json:"azp"`
	Email           string                `json:"email"`
	EmailVerified   bool                  `json:"email_verified"`
	Google          gceIdentityDataGoogle `json:"google"`
}

type gceIdentityDataGoogle struct {
	ComputeEngine gceIdentityDataComputeEngine `json:"compute_engine"`
}

type gceIdentityDataComputeEngine struct {
	ProjectID                 string `json:"project_id"`
	ProjectNumber             int64  `json:"project_number"`
	Zone                      string `json:"zone"`
	InstanceID                string `json:"instance_id"`
	InstanceName              string `json:"instance_name"`
	InstanceCreationTimestamp int64  `json:"instance_creation_timestamp"`
}

const AudienceNodeBootstrap = "node-bootstrap.kubernetes.io"

// getLocalInstanceIdentityClaim reads and returns the signed GCE instance identity description document
func getLocalInstanceIdentityClaim(ctx context.Context, audience string) (string, error) {
	client := &http.Client{}

	v := url.Values{}

	v.Set("audience", audience)
	format := "full"
	v.Set("format", format)

	u := "http://metadata/computeMetadata/v1/instance/service-accounts/default/identity?" + v.Encode()
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return "", fmt.Errorf("error building GCE identity document request: %v", err)
	}

	req = req.WithContext(ctx)

	req.Header.Add("Metadata-Flavor", "Google")
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error retrieving GCE identity document from metadata: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status %d reading GCE identity document from metadata", resp.StatusCode)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading GCE identity document from metadata: %v", err)
	}
	return string(b), nil
}

type jwtHeader struct {
	Algorithm string `json:"alg"`
	KeyID     string `json:"kid"`
}

type cachedKeys struct {
	raw       []byte
	timestamp time.Time

	mutex sync.Mutex
	certs map[string]*x509.Certificate
}

func (c *cachedKeys) buildKeys() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.certs != nil {
		return nil
	}

	certMap := make(map[string]string)
	if err := json.Unmarshal(c.raw, &certMap); err != nil {
		return fmt.Errorf("error parsing certificates: %v", err)
	}

	certs := make(map[string]*x509.Certificate)
	for k, v := range certMap {
		block, _ := pem.Decode([]byte(v))
		if block == nil {
			return fmt.Errorf("failed to parse certificate PEM")
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return fmt.Errorf("failed to parse x509 certificate: %v", err)
		}

		certs[k] = cert
	}

	c.certs = certs
	return nil
}

type Validator struct {
	certURL string

	mutex sync.Mutex
	keys  *cachedKeys
}

func NewValidator(ctx context.Context) (*Validator, error) {
	url := "https://www.googleapis.com/oauth2/v1/certs"
	v := &Validator{certURL: url}
	cachedKeys, err := v.fetchKeys(ctx)
	if err != nil {
		return nil, err
	}

	if err := cachedKeys.buildKeys(); err != nil {
		return nil, err
	}

	v.keys = cachedKeys
	return v, nil
}

func (v *Validator) fetchKeys(ctx context.Context) (*cachedKeys, error) {
	timestamp := time.Now()

	client := &http.Client{}

	req, err := http.NewRequest("GET", v.certURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error building query for %s: %v", v.certURL, err)
	}

	req = req.WithContext(ctx)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error querying certificates from %s: %v", v.certURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d from %s", resp.StatusCode, v.certURL)
	}

	raw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading certificates from %s: %v", v.certURL, err)
	}

	return &cachedKeys{
		raw:       raw,
		timestamp: timestamp,
	}, nil
}

func (v *Validator) findKey(ctx context.Context, id string) (*x509.Certificate, error) {
	v.mutex.Lock()
	keys := v.keys
	v.mutex.Unlock()

	var cert *x509.Certificate
	if keys != nil {
		cert = keys.certs[id]
	}

	if cert != nil {
		return cert, nil
	}

	v.mutex.Lock()
	defer v.mutex.Unlock()

	// TODO: How do we avoid DoS-ing the certificate endpoint, but still rotate keys?

	// Check again, in case of concurrent creation
	keys = v.keys
	if keys != nil {
		cert = keys.certs[id]
	}
	if cert != nil {
		return cert, nil
	}

	newKeys, err := v.fetchKeys(ctx)
	if err != nil {
		return nil, err
	}

	if newKeys != nil && keys != nil && bytes.Equal(newKeys.raw, keys.raw) {
		return nil, nil
	}

	if err := newKeys.buildKeys(); err != nil {
		return nil, err
	}

	v.keys = newKeys

	cert = newKeys.certs[id]
	return cert, nil
}

func (v *Validator) verifySignature(ctx context.Context, header *jwtHeader, payload []byte, signature []byte) error {
	if header.Algorithm != "RS256" {
		return fmt.Errorf("unexpected signature algorithm %q", header.Algorithm)
	}

	if header.KeyID == "" {
		return fmt.Errorf("token did not have KeyID")
	}

	key, err := v.findKey(ctx, header.KeyID)
	if err != nil {
		return err
	}

	if key == nil {
		// The key is user-provided, and we haven't verified the signature yet (we don't have the key!), so this could be faked
		// But it's not a good sign if untrusted parties are able to reach us - we should be behind a firewall
		return fmt.Errorf("token specified unknown key %q", header.KeyID)
	}

	rsaPublicKey, ok := key.PublicKey.(*rsa.PublicKey)
	if !ok {
		return fmt.Errorf("key was not an RSA public key")
	}

	hasher := crypto.SHA256.New()
	if _, err := hasher.Write(payload); err != nil {
		return fmt.Errorf("error hashing payload")
	}
	hashed := hasher.Sum(nil)

	if err := rsa.VerifyPKCS1v15(rsaPublicKey, crypto.SHA256, hashed, signature); err != nil {
		return fmt.Errorf("signature was not valid")
	}

	return nil
}

// decodeFromJWT does base64 decoding, but does not require the input to be padded
func decodeFromJWT(s string) ([]byte, error) {
	for (len(s) % 4) != 0 {
		s += "="
	}

	return base64.URLEncoding.DecodeString(s)
}

// ValidateDocument parses and validates the document, returning an error if the document is invalid
func (v *Validator) ParseAndValidateClaim(ctx context.Context, doc string, audience string) (*gceIdentityData, error) {
	tokens := strings.Split(doc, ".")
	if len(tokens) != 3 {
		return nil, fmt.Errorf("identity document did not have expected 3 parts")
	}

	// @step: decode the signed document
	header := &jwtHeader{}
	{
		decoded, err := decodeFromJWT(tokens[0])
		if err != nil {
			return nil, fmt.Errorf("malformed token header")
		}
		if err := json.Unmarshal(decoded, header); err != nil {
			return nil, fmt.Errorf("error parsing token header")
		}
	}

	signature, err := decodeFromJWT(tokens[2])
	if err != nil {
		return nil, fmt.Errorf("malformed token signature")
	}

	if err := v.verifySignature(ctx, header, []byte(tokens[0]+"."+tokens[1]), signature); err != nil {
		return nil, fmt.Errorf("signature was not valid: %v", err)
	}

	data := &gceIdentityData{}
	{
		decoded, err := decodeFromJWT(tokens[1])
		if err != nil {
			return nil, fmt.Errorf("malformed token payload")
		}
		if err := json.Unmarshal(decoded, data); err != nil {
			return nil, fmt.Errorf("error parsing token payload")
		}
	}

	now := time.Now().UTC().Unix()
	if data.Expiry < now {
		return nil, fmt.Errorf("token has expired")
	}

	if data.Audience != audience {
		return nil, fmt.Errorf("invalid audience")
	}

	return data, nil
}
