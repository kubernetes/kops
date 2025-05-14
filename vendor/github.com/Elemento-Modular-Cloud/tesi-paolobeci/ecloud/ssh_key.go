package ecloud

// TODO: other methods like GetByName, List, Create, Update, Delete

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"

	// "encoding/json"
	"errors"
	"fmt"

	// "net/url"
	// "strconv"
	"time"
	"golang.org/x/crypto/ssh"
)

// SSHKey represents an SSH key in Elemento Cloud.
type SSHKey struct {
	ID          int
	Name        string
	Fingerprint string
	PublicKey   string
	Labels      map[string]string
	Created     time.Time
}

// SSHKeyClient is a client for the SSH keys API.
type SSHKeyClient struct {
	client *Client
}

// All returns all SSH keys.
func (c *SSHKeyClient) All(ctx context.Context) ([]*SSHKey, error) {
	return c.AllWithOpts(ctx, SSHKeyListOpts{})
}

// AllWithOpts returns all SSH keys with the given options.
func (c *SSHKeyClient) AllWithOpts(ctx context.Context, opts SSHKeyListOpts) ([]*SSHKey, error) {
	allSSHKeys := []*SSHKey{}

	return allSSHKeys, nil
}

// SSHKeyListOpts specifies options for listing SSH keys.
type SSHKeyListOpts struct {
	Name        string
	Fingerprint string
	Sort        []string
}

// SSHKeyCreateOpts specifies parameters for creating a SSH key.
type SSHKeyCreateOpts struct {
	Name      string
	PublicKey string
	Labels    map[string]string
}

// Validate checks if options are valid.
func (o SSHKeyCreateOpts) Validate() error {
	if o.Name == "" {
		return errors.New("missing name")
	}
	if o.PublicKey == "" {
		return errors.New("missing public key")
	}
	return nil
}

// Create creates a new SSH key with the given options.
func (c *SSHKeyClient) Create(ctx context.Context, opts SSHKeyCreateOpts) (*SSHKey, *Response, error) {
	if err := opts.Validate(); err != nil {
		return nil, nil, err
	}

	// 1. Generate SSH key pair
	privateKey, publicKey, err := generateSSHKeyPair()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate SSH key pair: %w", err)
	}

	// 2. Save keys to local .txt files
	basePath := fmt.Sprintf("./ssh_keys/%s", opts.Name)
	err = os.MkdirAll(filepath.Dir(basePath), 0700)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create directory: %w", err)
	}

	err = os.WriteFile(basePath+"_private.txt", privateKey, 0600)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to save private key: %w", err)
	}

	err = os.WriteFile(basePath+"_public.txt", publicKey, 0644)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to save public key: %w", err)
	}
	// 3. Construct and return SSHKey object
	sshKey := &SSHKey{
		Name:      opts.Name,
		PublicKey: string(publicKey),
		Labels:    opts.Labels,
	}
	return sshKey, &Response{}, nil
}

func generateSSHKeyPair() ([]byte, []byte, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	privatePEM := &bytes.Buffer{}
	err = pem.Encode(privatePEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	if err != nil {
		return nil, nil, err
	}

	pub, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, nil, err
	}
	publicKeyBytes := ssh.MarshalAuthorizedKey(pub)

	return privatePEM.Bytes(), publicKeyBytes, nil
}
